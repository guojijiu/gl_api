package task

// 公网云「任务」领域：与内网 cloud_intranet 独立维护。

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"
	"cloud-platform-api/app/Services/ai_robot/policy"

	"github.com/gin-gonic/gin"
)

const (
	maxConcurrentUUIDs     = 4
	maxConcurrentModuleIDs = 4
)

func failIfTaskCloudCallFailed(r deps.Responder, d *deps.Deps, errMsg string, status int, err error) bool {
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

func summarizeTaskWithData(r deps.Responder, d *deps.Deps, intentKey string, raw []byte, extra gin.H) bool {
	d.RememberIntent(intentKey)
	answer, err := summary.SummarizeTaskData(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), raw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return false
	}
	resp := gin.H{
		"answer":       answer,
		"intent":       intentKey,
		"cloud_called": true,
		"strict_mode":  d.Cfg.StrictMode,
	}
	for k, v := range extra {
		resp[k] = v
	}
	r.FrontSuccess(d.Gin, "操作成功", resp)
	return true
}

// Dispatch 处理 question_type 为「任务」的请求。
func Dispatch(r deps.Responder, d *deps.Deps) {
	plan, err := intent.PlanTaskIntent(d.Ctx, d.LLM, d.Question(), d.ContextSummary())
	if err != nil {
		r.FrontFailed(d.Gin, "意图解析失败", err)
		return
	}
	uuids := parse.ExtractTaskUUIDsFromQuestion(d.Question())
	if len(uuids) == 0 {
		r.FrontFailed(d.Gin, parse.TaskUUIDGuidanceForQuestion(d.Question()), nil)
		return
	}
	d.RememberTaskUUIDs(uuids)
	uuidsCSV := strings.Join(uuids, ",")

	if !policy.IsIntentAllowed(d.PlatformID(), d.Req.QuestionType, plan.Intent) {
		r.FrontFailed(d.Gin, "当前平台下不支持该任务能力", nil)
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
	case intent.IntentTaskStatus:
		handleTaskStatus(r, d, plan.Intent, uuids, uuidsCSV)
	case intent.IntentTaskDownloadResult:
		handleTaskDownload(r, d, plan.Intent, uuids)
	default:
		r.FrontFailed(d.Gin, "未知意图分支", nil)
	}
}

func handleTaskStatus(r deps.Responder, d *deps.Deps, intentKey string, uuids []string, uuidsCSV string) {
	statusRaw, statusCode, err := d.Cloud().Task().StatusByUUIDs(d.Ctx, d.PlatformID(), d.Token, uuidsCSV)
	if failIfTaskCloudCallFailed(r, d, "查询任务状态失败", statusCode, err) {
		return
	}
	d.RememberTaskStatusResult(statusRaw)

	// 并行处理每个 uuid 的下游调用，降低整体耗时。

	type uuidDetail struct {
		h gin.H
	}
	detailsByIndex := make([]uuidDetail, len(uuids))

	sem := make(chan struct{}, maxConcurrentUUIDs)
	var wg sync.WaitGroup
	wg.Add(len(uuids))

	for i, u := range uuids {
		go func(idx int, uuid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			taskFilters := parse.ExtractTaskFiltersFromQuestion(d.Question(), uuid)
			listRaw, st, e := d.Cloud().Task().ListToolByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
			if e != nil || st >= 400 {
				listRaw = nil
			}
			workflowRaw, wst, we := d.Cloud().Task().ListWorkflowByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
			if we != nil || wst >= 400 {
				workflowRaw = nil
			}

			taskID := cloudclient.ExtractFirstIDFromFrontList(listRaw)
			if taskID <= 0 {
				taskID = cloudclient.ExtractFirstIDFromFrontList(workflowRaw)
			}

			var result interface{}
			if taskID > 0 {
				rRaw, rst, re := d.Cloud().Task().Result(d.Ctx, d.PlatformID(), d.Token, taskID)
				if re == nil && rst < 400 {
					result = cloudclient.JsonRaw(rRaw)
				}
			}
			detailsByIndex[idx] = uuidDetail{
				h: gin.H{
					"uuid":          uuid,
					"task_id":       taskID,
					"tool_task":     cloudclient.JsonRaw(listRaw),
					"workflow_task": cloudclient.JsonRaw(workflowRaw),
					"result":        result,
				},
			}
		}(i, u)
	}
	wg.Wait()

	var details []gin.H
	details = make([]gin.H, 0, len(uuids))
	for i := range detailsByIndex {
		details = append(details, detailsByIndex[i].h)
	}

	moduleIDs := parse.ExtractNumericIDsFromQuestion(d.Question())
	if len(moduleIDs) > 0 {
		// 并行处理 module 相关下游调用，降低整体耗时。
		// 保持输出顺序：按 moduleIDs 的原始顺序追加 details。

		type moduleDetailResult struct {
			ok bool
			h  gin.H
		}
		resultsByIndex := make([]moduleDetailResult, len(moduleIDs))

		sem := make(chan struct{}, maxConcurrentModuleIDs)
		var wg sync.WaitGroup
		wg.Add(len(moduleIDs))

		for i, mid := range moduleIDs {
			go func(idx int, moduleID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				mRaw, mst, me := d.Cloud().Task().ModuleDetail(d.Ctx, d.PlatformID(), d.Token, moduleID)
				if me != nil || mst >= 400 {
					return
				}
				resultsByIndex[idx] = moduleDetailResult{
					ok: true,
					h: gin.H{
						"module_task_id": moduleID,
						"module_detail":  cloudclient.JsonRaw(mRaw),
					},
				}
			}(i, mid)
		}

		wg.Wait()

		for i := range resultsByIndex {
			if resultsByIndex[i].ok {
				details = append(details, resultsByIndex[i].h)
			}
		}
	}

	summaryRaw, _ := json.Marshal(gin.H{"uuids": uuids, "status_batch": cloudclient.JsonRaw(statusRaw), "details": details})
	_ = summarizeTaskWithData(r, d, intentKey, summaryRaw, gin.H{
		"uuids": uuids,
		"raw_cloud_json": gin.H{
			"status_by_uuids": cloudclient.JsonRaw(statusRaw),
			"details":         details,
		},
	})
}

func handleTaskDownload(r deps.Responder, d *deps.Deps, intentKey string, uuids []string) {
	var links []gin.H
	var rawList []interface{}

	// 并行处理每个 uuid 的下游调用，降低整体耗时。

	type uuidDownloadResult struct {
		raw   []interface{}
		links []gin.H
	}
	resultsByIndex := make([]uuidDownloadResult, len(uuids))

	sem := make(chan struct{}, maxConcurrentUUIDs)
	var wg sync.WaitGroup
	wg.Add(len(uuids))

	for i, u := range uuids {
		go func(idx int, uuid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			taskFilters := parse.ExtractTaskFiltersFromQuestion(d.Question(), uuid)
			listRaw, st, e := d.Cloud().Task().ListToolByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
			if e != nil || st >= 400 {
				listRaw = nil
			}
			workflowRaw, wst, we := d.Cloud().Task().ListWorkflowByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
			if we != nil || wst >= 400 {
				workflowRaw = nil
			}

			taskID := cloudclient.ExtractFirstIDFromFrontList(listRaw)
			taskKind := "tool"
			if taskID <= 0 {
				taskID = cloudclient.ExtractFirstIDFromFrontList(workflowRaw)
				taskKind = "workflow"
			}
			if taskID <= 0 {
				return
			}

			dRaw, dst, de := d.Cloud().Task().DownloadResult(d.Ctx, d.PlatformID(), d.Token, taskID)
			if de != nil || dst >= 400 {
				return
			}

			res := uuidDownloadResult{}
			res.raw = append(res.raw, cloudclient.JsonRaw(dRaw))
			if fp := cloudclient.ExtractFrontFilePath(dRaw); fp != "" {
				res.links = append(res.links, gin.H{"uuid": uuid, "task_id": taskID, "task_kind": taskKind, "file_path": fp})
			}
			resultsByIndex[idx] = res
		}(i, u)
	}
	wg.Wait()

	for i := range resultsByIndex {
		rawList = append(rawList, resultsByIndex[i].raw...)
		links = append(links, resultsByIndex[i].links...)
	}

	moduleIDs := parse.ExtractNumericIDsFromQuestion(d.Question())
	if len(moduleIDs) > 0 {
		// 并行处理 module 相关下游调用，降低整体耗时。
		// 保持输出顺序：按 moduleIDs 的原始顺序追加 rawList/links。

		type moduleDownloadResult struct {
			ok      bool
			raw     interface{}
			link    gin.H
			hasLink bool
		}
		resultsByIndex := make([]moduleDownloadResult, len(moduleIDs))

		sem := make(chan struct{}, maxConcurrentModuleIDs)
		var wg sync.WaitGroup
		wg.Add(len(moduleIDs))

		for i, mid := range moduleIDs {
			go func(idx int, moduleID int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				mRaw, mst, me := d.Cloud().Task().ModuleResultURL(d.Ctx, d.PlatformID(), d.Token, moduleID)
				if me != nil || mst >= 400 {
					return
				}

				rawItem := cloudclient.JsonRaw(mRaw)
				fp := cloudclient.ExtractAnyURL(mRaw)
				var link gin.H
				hasLink := false
				if fp != "" {
					link = gin.H{"task_id": moduleID, "task_kind": "module", "file_path": fp}
					hasLink = true
				}

				resultsByIndex[idx] = moduleDownloadResult{
					ok:      true,
					raw:     rawItem,
					link:    link,
					hasLink: hasLink,
				}
			}(i, mid)
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
		r.FrontFailed(d.Gin, parse.TaskUUIDGuidanceForQuestion(d.Question()), nil)
		return
	}
	d.RememberDownloadResult("task_result", len(links) > 0)
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	_ = summarizeTaskWithData(r, d, intentKey, summaryRaw, gin.H{
		"uuids":          uuids,
		"links":          links,
		"raw_cloud_json": rawList,
	})
}
