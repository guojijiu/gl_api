package project

// 公网云「项目/合同」领域：内网 cloud_intranet 复制本目录逻辑时同步更新 import。

import (
	"encoding/json"

	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/cloudfail"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/concurrent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/limitcap"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/reply"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"
	"cloud-platform-api/app/Services/ai_robot/policy"
	AiRobotUtils "cloud-platform-api/app/Services/ai_robot/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxConcurrentContractProjects    = 4
	maxConcurrentProjectZips         = 4
	maxConcurrentProjectOriginalData = 4
)

type projectDownloadResult struct {
	ok      bool
	raw     interface{}
	link    gin.H
	hasLink bool
}

// Dispatch 处理 question_type 为「项目」的请求。
func Dispatch(r deps.Responder, d *deps.Deps) {
	plan, err := intent.PlanProjectIntent(d.Ctx, d.LLM, d.Question(), d.ContextSummary())
	if err != nil {
		r.FrontFailed(d.Gin, "意图解析失败", err)
		return
	}
	if !policy.IsIntentAllowed(d.PlatformID(), d.Req.QuestionType, plan.Intent) {
		r.FrontFailed(d.Gin, "当前平台下不支持该项目能力", nil)
		return
	}
	switch plan.Intent {
	case intent.IntentUnsupported:
		_ = reply.RespondUnsupported(r, d, plan.Intent, plan.Reason)
	case intent.IntentProjectList, intent.IntentProjectExpiring:
		handleProjectListOrExpiring(r, d, plan.Intent)
	case intent.IntentContractList:
		handleContractList(r, d, plan.Intent)
	case intent.IntentContractProjects:
		handleContractProjects(r, d, plan.Intent)
	case intent.IntentDownloadFinal:
		handleDownloadFinalReport(r, d, plan.Intent)
	case intent.IntentDownloadOriginal:
		handleDownloadOriginalData(r, d, plan.Intent)
	default:
		r.FrontFailed(d.Gin, "未知意图分支", nil)
	}
}

func handleProjectListOrExpiring(r deps.Responder, d *deps.Deps, intentKey string) {
	raw, status, err := deps.TrackNamedCloudCall("project.list", d, func() ([]byte, int, error) {
		return d.Cloud().Project().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractProjectFiltersFromQuestion(d.Question()))
	})
	if cloudfail.HandleCloudCallFailure(r, d, "请求云平台失败", status, err) {
		return
	}
	if intentKey == intent.IntentProjectExpiring {
		raw, _ = FilterExpiringProjects(d.Cfg, raw)
	}
	_ = reply.RespondCloudListSummary(r, d, intentKey, raw, summary.SummarizeProjectData)
}

func handleContractList(r deps.Responder, d *deps.Deps, intentKey string) {
	raw, status, err := deps.TrackNamedCloudCall("contract.list", d, func() ([]byte, int, error) {
		return d.Cloud().Contract().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractContractFiltersFromQuestion(d.Question()))
	})
	if cloudfail.HandleCloudCallFailure(r, d, "查询合同列表失败", status, err) {
		return
	}
	_ = reply.RespondCloudListSummary(r, d, intentKey, raw, summary.SummarizeProjectData)
}

func handleContractProjects(r deps.Responder, d *deps.Deps, intentKey string) {
	contractFilters := parse.ExtractContractFiltersFromQuestion(d.Question())
	projectFilters := parse.ExtractProjectFiltersFromQuestion(d.Question())
	contractsRaw, status, err := deps.TrackNamedCloudCall("contract.list", d, func() ([]byte, int, error) {
		return d.Cloud().Contract().List(d.Ctx, d.PlatformID(), d.Token, contractFilters)
	})
	if cloudfail.HandleCloudCallFailure(r, d, "查询合同列表失败", status, err) {
		return
	}
	contractIDs, matchedContractNumber, err := ResolveContractIDsFromContracts(d.Ctx, d.Cfg, d.Question(), contractsRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "自动匹配合同失败", err)
		return
	}

	d.RememberContractMatch(contractIDs, matchedContractNumber)
	var grouped []gin.H
	if len(contractIDs) > 0 {
		// 并行拉取每个 contract_id 下的项目列表，降低 overall 延迟。
		// 保持输出顺序：按 contractIDs 的原始顺序追加 grouped。
		type contractProjectsResult struct {
			ok bool
			h  gin.H
		}
		resultsByIndex := make([]contractProjectsResult, len(contractIDs))
		concurrent.RunIndexed(len(contractIDs), maxConcurrentContractProjects, func(idx int) {
			contractID := contractIDs[idx]
			raw, st, e := deps.TrackNamedCloudCall("project.list_by_contract", d, func() ([]byte, int, error) {
				return d.Cloud().Project().ListByContract(d.Ctx, d.PlatformID(), d.Token, contractID, projectFilters)
			})
			if e != nil || st >= 400 {
				return
			}

			resultsByIndex[idx] = contractProjectsResult{
				ok: true,
				h: gin.H{
					"contract_id":     contractID,
					"projects_result": cloudclient.JsonRaw(raw),
				},
			}
		})

		grouped = make([]gin.H, 0, len(contractIDs))
		for i := range resultsByIndex {
			if resultsByIndex[i].ok {
				grouped = append(grouped, resultsByIndex[i].h)
			}
		}
	}
	if len(grouped) == 0 {
		r.FrontFailed(d.Gin, "查询合同下项目失败", nil)
		return
	}
	summaryRaw, _ := json.Marshal(gin.H{"contract_projects": grouped})
	groupedOut, groupedTotal := limitcap.CapGinHSlice(grouped, limitcap.DefaultVisibleItems)
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), summaryRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	r.FrontSuccess(d.Gin, "操作成功", reply.BuildCloudSuccessPayload(d, intentKey, answer, gin.H{
		"matched_contract_ids":        contractIDs,
		"matched_contract_number":     matchedContractNumber,
		"contract_projects":           groupedOut,
		"contract_projects_total":     groupedTotal,
		"contract_projects_truncated": groupedTotal > len(groupedOut),
	}))
}

func handleDownloadFinalReport(r deps.Responder, d *deps.Deps, intentKey string) {
	projectsRaw, status, err := deps.TrackNamedCloudCall("project.list", d, func() ([]byte, int, error) {
		return d.Cloud().Project().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractProjectFiltersFromQuestion(d.Question()))
	})
	if cloudfail.HandleCloudCallFailure(r, d, "查询项目列表失败", status, err) {
		return
	}
	projectIDs, matchedNumber, err := ResolveProjectIDsFromProjects(d.Ctx, d.LLM, d.Cfg, d.Question(), projectsRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "自动匹配项目失败", err)
		return
	}
	d.RememberProjectMatch(projectIDs, matchedNumber)
	links, rawList := collectProjectDownloads(projectIDs, maxConcurrentProjectZips, func(projectID int) projectDownloadResult {
		raw, st, e := deps.TrackNamedCloudCall("project.zip_url", d, func() ([]byte, int, error) {
			return d.Cloud().Project().ZipURL(d.Ctx, d.PlatformID(), d.Token, projectID)
		})
		if e != nil || st >= 400 {
			return projectDownloadResult{}
		}
		result := projectDownloadResult{
			ok:  true,
			raw: cloudclient.JsonRaw(raw),
		}
		if u := cloudclient.ExtractFrontURL(raw); u != "" {
			result.link = gin.H{"project_id": projectID, "url": u}
			result.hasLink = true
		}
		return result
	})
	if len(rawList) == 0 {
		r.FrontFailed(d.Gin, "请求云平台失败", nil)
		return
	}
	answer, payload, err := buildProjectDownloadPayload(d, intentKey, "final_report", links, rawList)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	payload["matched_project_ids"] = projectIDs
	payload["matched_number"] = matchedNumber
	payload["answer"] = answer
	r.FrontSuccess(d.Gin, "操作成功", payload)
}

func handleDownloadOriginalData(r deps.Responder, d *deps.Deps, intentKey string) {
	projectsRaw, status, err := deps.TrackNamedCloudCall("project.list", d, func() ([]byte, int, error) {
		return d.Cloud().Project().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractProjectFiltersFromQuestion(d.Question()))
	})
	if cloudfail.HandleCloudCallFailure(r, d, "查询项目列表失败", status, err) {
		return
	}
	projectIDs, matchedNumber, err := ResolveProjectIDsFromProjects(d.Ctx, d.LLM, d.Cfg, d.Question(), projectsRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "自动匹配项目失败", err)
		return
	}
	d.RememberProjectMatch(projectIDs, matchedNumber)
	links, rawList := collectProjectDownloads(projectIDs, maxConcurrentProjectOriginalData, func(projectID int) projectDownloadResult {
		accessCode := AiRobotUtils.GenerateAccessCode(6)
		raw, st, e := deps.TrackNamedCloudCall("project.original_data_url", d, func() ([]byte, int, error) {
			return d.Cloud().Project().OriginalDataURL(d.Ctx, d.PlatformID(), d.Token, projectID, accessCode)
		})
		if e != nil || st >= 400 {
			return projectDownloadResult{}
		}
		result := projectDownloadResult{
			ok:  true,
			raw: cloudclient.JsonRaw(raw),
		}
		if u := cloudclient.ExtractFrontURL(raw); u != "" {
			result.link = gin.H{"project_id": projectID, "access_code": accessCode, "url": u}
			result.hasLink = true
		}
		return result
	})
	if len(rawList) == 0 {
		r.FrontFailed(d.Gin, "请求云平台失败", nil)
		return
	}
	answer, payload, err := buildProjectDownloadPayload(d, intentKey, "original_data", links, rawList)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	payload["matched_project_ids"] = projectIDs
	payload["matched_number"] = matchedNumber
	payload["answer"] = answer
	r.FrontSuccess(d.Gin, "操作成功", payload)
}

func collectProjectDownloads(projectIDs []int, maxConcurrent int, worker func(projectID int) projectDownloadResult) ([]gin.H, []interface{}) {
	if len(projectIDs) == 0 || worker == nil {
		return nil, nil
	}
	resultsByIndex := make([]projectDownloadResult, len(projectIDs))
	concurrent.RunIndexed(len(projectIDs), maxConcurrent, func(idx int) {
		projectID := projectIDs[idx]
		resultsByIndex[idx] = worker(projectID)
	})

	links := make([]gin.H, 0, len(projectIDs))
	rawList := make([]interface{}, 0, len(projectIDs))
	for i := range resultsByIndex {
		if !resultsByIndex[i].ok {
			continue
		}
		rawList = append(rawList, resultsByIndex[i].raw)
		if resultsByIndex[i].hasLink {
			links = append(links, resultsByIndex[i].link)
		}
	}
	return links, rawList
}

func buildProjectDownloadPayload(d *deps.Deps, intentKey string, downloadKind string, links []gin.H, rawList []interface{}) (string, gin.H, error) {
	d.RememberDownloadResult(downloadKind, len(links) > 0)
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	linksOut, linksTotal := limitcap.CapGinHSlice(links, limitcap.DefaultVisibleItems)
	rawOut, rawTotal := limitcap.CapAnySlice(rawList, limitcap.DefaultVisibleItems)
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), summaryRaw)
	if err != nil {
		return "", nil, err
	}
	return answer, reply.BuildCloudSuccessPayload(d, intentKey, "", gin.H{
		"links":              linksOut,
		"links_total":        linksTotal,
		"links_truncated":    linksTotal > len(linksOut),
		"raw_cloud_json":     rawOut,
		"raw_list_total":     rawTotal,
		"raw_list_truncated": rawTotal > len(rawOut),
	}), nil
}
