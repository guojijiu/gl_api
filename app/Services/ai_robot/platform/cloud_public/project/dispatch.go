package project

// 公网云「项目/合同」领域：内网 cloud_intranet 复制本目录逻辑时同步更新 import。

import (
	"encoding/json"
	"fmt"
	"sync"

	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"
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

func failIfCloudCallFailed(r deps.Responder, d *deps.Deps, errMsg string, status int, err error) bool {
	if err != nil {
		r.FrontFailed(d.Gin, errMsg, err)
		return true
	}
	if status >= 400 {
		r.FrontFailed(d.Gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return true
	}
	return false
}

func summarizeAndReply(r deps.Responder, d *deps.Deps, intentKey string, raw []byte) bool {
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), raw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return false
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intentKey,
		"cloud_called":   true,
		"strict_mode":    d.Cfg.StrictMode,
		"raw_cloud_json": cloudclient.JsonRaw(raw),
	})
	return true
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
		answer, e := summary.SummarizeUnsupported(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), plan.Reason)
		if e != nil {
			r.FrontFailed(d.Gin, "生成回复失败", e)
			return
		}
		r.FrontSuccess(d.Gin, "操作成功", gin.H{
			"answer":         answer,
			"intent":         plan.Intent,
			"cloud_called":   false,
			"strict_mode":    d.Cfg.StrictMode,
			"raw_cloud_json": nil,
		})
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
	if failIfCloudCallFailed(r, d, "请求云平台失败", status, err) {
		return
	}
	if intentKey == intent.IntentProjectExpiring {
		raw, _ = FilterExpiringProjects(d.Cfg, raw)
	}
	_ = summarizeAndReply(r, d, intentKey, raw)
}

func handleContractList(r deps.Responder, d *deps.Deps, intentKey string) {
	raw, status, err := deps.TrackNamedCloudCall("contract.list", d, func() ([]byte, int, error) {
		return d.Cloud().Contract().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractContractFiltersFromQuestion(d.Question()))
	})
	if failIfCloudCallFailed(r, d, "查询合同列表失败", status, err) {
		return
	}
	_ = summarizeAndReply(r, d, intentKey, raw)
}

func handleContractProjects(r deps.Responder, d *deps.Deps, intentKey string) {
	contractFilters := parse.ExtractContractFiltersFromQuestion(d.Question())
	projectFilters := parse.ExtractProjectFiltersFromQuestion(d.Question())
	contractsRaw, status, err := deps.TrackNamedCloudCall("contract.list", d, func() ([]byte, int, error) {
		return d.Cloud().Contract().List(d.Ctx, d.PlatformID(), d.Token, contractFilters)
	})
	if failIfCloudCallFailed(r, d, "查询合同列表失败", status, err) {
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

		sem := make(chan struct{}, maxConcurrentContractProjects)
		var wg sync.WaitGroup
		wg.Add(len(contractIDs))

		for i, cid := range contractIDs {
			go func(idx int, contractID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

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
			}(i, cid)
		}

		wg.Wait()

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
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), summaryRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":                  answer,
		"intent":                  intentKey,
		"cloud_called":            true,
		"strict_mode":             d.Cfg.StrictMode,
		"matched_contract_ids":    contractIDs,
		"matched_contract_number": matchedContractNumber,
		"contract_projects":       grouped,
	})
}

func handleDownloadFinalReport(r deps.Responder, d *deps.Deps, intentKey string) {
	projectsRaw, status, err := deps.TrackNamedCloudCall("project.list", d, func() ([]byte, int, error) {
		return d.Cloud().Project().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractProjectFiltersFromQuestion(d.Question()))
	})
	if failIfCloudCallFailed(r, d, "查询项目列表失败", status, err) {
		return
	}
	projectIDs, matchedNumber, err := ResolveProjectIDsFromProjects(d.Ctx, d.LLM, d.Cfg, d.Question(), projectsRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "自动匹配项目失败", err)
		return
	}
	d.RememberProjectMatch(projectIDs, matchedNumber)
	var links []gin.H
	var rawList []interface{}
	if len(projectIDs) > 0 {
		// 并行拉取每个项目的 zip 下载链接，降低整体耗时。
		// 保持输出顺序：按 projectIDs 的原始顺序追加 rawList/links。
		type projectZipResult struct {
			ok      bool
			raw     interface{}
			link    gin.H
			hasLink bool
		}
		resultsByIndex := make([]projectZipResult, len(projectIDs))

		sem := make(chan struct{}, maxConcurrentProjectZips)
		var wg sync.WaitGroup
		wg.Add(len(projectIDs))

		for i, id := range projectIDs {
			go func(idx int, projectID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				raw, st, e := deps.TrackNamedCloudCall("project.zip_url", d, func() ([]byte, int, error) {
					return d.Cloud().Project().ZipURL(d.Ctx, d.PlatformID(), d.Token, projectID)
				})
				if e != nil || st >= 400 {
					return
				}

				rawItem := cloudclient.JsonRaw(raw)
				u := cloudclient.ExtractFrontURL(raw)
				var link gin.H
				hasLink := false
				if u != "" {
					link = gin.H{"project_id": projectID, "url": u}
					hasLink = true
				}

				resultsByIndex[idx] = projectZipResult{
					ok:      true,
					raw:     rawItem,
					link:    link,
					hasLink: hasLink,
				}
			}(i, id)
		}

		wg.Wait()

		for i := range resultsByIndex {
			if !resultsByIndex[i].ok {
				continue
			}
			rawList = append(rawList, resultsByIndex[i].raw)
			if resultsByIndex[i].hasLink {
				links = append(links, resultsByIndex[i].link)
			}
		}
	}
	if len(rawList) == 0 {
		r.FrontFailed(d.Gin, "请求云平台失败", nil)
		return
	}
	d.RememberDownloadResult("final_report", len(links) > 0)
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), summaryRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":              answer,
		"intent":              intentKey,
		"cloud_called":        true,
		"strict_mode":         d.Cfg.StrictMode,
		"matched_project_ids": projectIDs,
		"matched_number":      matchedNumber,
		"links":               links,
		"raw_cloud_json":      rawList,
	})
}

func handleDownloadOriginalData(r deps.Responder, d *deps.Deps, intentKey string) {
	projectsRaw, status, err := deps.TrackNamedCloudCall("project.list", d, func() ([]byte, int, error) {
		return d.Cloud().Project().List(d.Ctx, d.PlatformID(), d.Token, parse.ExtractProjectFiltersFromQuestion(d.Question()))
	})
	if failIfCloudCallFailed(r, d, "查询项目列表失败", status, err) {
		return
	}
	projectIDs, matchedNumber, err := ResolveProjectIDsFromProjects(d.Ctx, d.LLM, d.Cfg, d.Question(), projectsRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "自动匹配项目失败", err)
		return
	}
	d.RememberProjectMatch(projectIDs, matchedNumber)
	var links []gin.H
	var rawList []interface{}
	if len(projectIDs) > 0 {
		// 并行拉取每个项目的原始数据下载链接，降低 overall 延迟。
		// 保持输出顺序：按 projectIDs 的原始顺序追加 rawList/links。
		type originalDataResult struct {
			ok      bool
			raw     interface{}
			link    gin.H
			hasLink bool
		}
		resultsByIndex := make([]originalDataResult, len(projectIDs))

		sem := make(chan struct{}, maxConcurrentProjectOriginalData)
		var wg sync.WaitGroup
		wg.Add(len(projectIDs))

		for i, id := range projectIDs {
			go func(idx int, projectID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				accessCode := AiRobotUtils.GenerateAccessCode(6)
				raw, st, e := deps.TrackNamedCloudCall("project.original_data_url", d, func() ([]byte, int, error) {
					return d.Cloud().Project().OriginalDataURL(d.Ctx, d.PlatformID(), d.Token, projectID, accessCode)
				})
				if e != nil || st >= 400 {
					return
				}

				rawItem := cloudclient.JsonRaw(raw)
				u := cloudclient.ExtractFrontURL(raw)
				var link gin.H
				hasLink := false
				if u != "" {
					link = gin.H{"project_id": projectID, "access_code": accessCode, "url": u}
					hasLink = true
				}

				resultsByIndex[idx] = originalDataResult{
					ok:      true,
					raw:     rawItem,
					link:    link,
					hasLink: hasLink,
				}
			}(i, id)
		}

		wg.Wait()

		for i := range resultsByIndex {
			if !resultsByIndex[i].ok {
				continue
			}
			rawList = append(rawList, resultsByIndex[i].raw)
			if resultsByIndex[i].hasLink {
				links = append(links, resultsByIndex[i].link)
			}
		}
	}
	if len(rawList) == 0 {
		r.FrontFailed(d.Gin, "请求云平台失败", nil)
		return
	}
	d.RememberDownloadResult("original_data", len(links) > 0)
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeProjectData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), summaryRaw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":              answer,
		"intent":              intentKey,
		"cloud_called":        true,
		"strict_mode":         d.Cfg.StrictMode,
		"matched_project_ids": projectIDs,
		"matched_number":      matchedNumber,
		"links":               links,
		"raw_cloud_json":      rawList,
	})
}
