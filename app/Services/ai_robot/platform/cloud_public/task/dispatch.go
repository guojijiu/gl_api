package task

// 公网云「任务」领域：与内网 cloud_intranet 独立维护。

import (
	"encoding/json"
	"strings"

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

	"github.com/gin-gonic/gin"
)

const (
	maxConcurrentUUIDs     = 4
	maxConcurrentModuleIDs = 4
)

type taskLookupResult struct {
	TaskID      int
	TaskKind    string
	ToolRaw     []byte
	WorkflowRaw []byte
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
		_ = reply.RespondUnsupported(r, d, plan.Intent, plan.Reason)
	case intent.IntentTaskStatus:
		handleTaskStatus(r, d, plan.Intent, uuids, uuidsCSV)
	case intent.IntentTaskDownloadResult:
		handleTaskDownload(r, d, plan.Intent, uuids)
	default:
		r.FrontFailed(d.Gin, "未知意图分支", nil)
	}
}

func handleTaskStatus(r deps.Responder, d *deps.Deps, intentKey string, uuids []string, uuidsCSV string) {
	statusRaw, statusCode, err := deps.TrackNamedCloudCall("task.status_by_uuids", d, func() ([]byte, int, error) {
		return d.Cloud().Task().StatusByUUIDs(d.Ctx, d.PlatformID(), d.Token, uuidsCSV)
	})
	if cloudfail.HandleCloudCallFailure(r, d, "查询任务状态失败", statusCode, err) {
		return
	}
	d.RememberTaskStatusResult(statusRaw)

	// 并行处理每个 uuid 的下游调用，降低整体耗时。

	type uuidDetail struct {
		h gin.H
	}
	detailsByIndex := make([]uuidDetail, len(uuids))
	concurrent.RunIndexed(len(uuids), maxConcurrentUUIDs, func(idx int) {
		uuid := uuids[idx]
		lookup := lookupTaskByUUID(d, uuid)
		taskID := lookup.TaskID

		var result interface{}
		if taskID > 0 {
			rRaw, rst, re := deps.TrackNamedCloudCall("task.result", d, func() ([]byte, int, error) {
				return d.Cloud().Task().Result(d.Ctx, d.PlatformID(), d.Token, taskID)
			})
			if re == nil && rst < 400 {
				result = cloudclient.JsonRaw(rRaw)
			}
		}
		detailsByIndex[idx] = uuidDetail{
			h: gin.H{
				"uuid":          uuid,
				"task_id":       taskID,
				"tool_task":     cloudclient.JsonRaw(lookup.ToolRaw),
				"workflow_task": cloudclient.JsonRaw(lookup.WorkflowRaw),
				"result":        result,
			},
		}
	})

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
		concurrent.RunIndexed(len(moduleIDs), maxConcurrentModuleIDs, func(idx int) {
			moduleID := moduleIDs[idx]
			mRaw, mst, me := deps.TrackNamedCloudCall("task.module_detail", d, func() ([]byte, int, error) {
				return d.Cloud().Task().ModuleDetail(d.Ctx, d.PlatformID(), d.Token, moduleID)
			})
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
		})

		for i := range resultsByIndex {
			if resultsByIndex[i].ok {
				details = append(details, resultsByIndex[i].h)
			}
		}
	}

	summaryRaw, _ := json.Marshal(gin.H{"uuids": uuids, "status_batch": cloudclient.JsonRaw(statusRaw), "details": details})
	detailsOut, detailsTotal := limitcap.CapGinHSlice(details, limitcap.DefaultVisibleItems)
	_ = reply.RespondCloudSummary(r, d, intentKey, summaryRaw, summary.SummarizeTaskData, gin.H{
		"uuids": uuids,
		"raw_cloud_json": gin.H{
			"status_by_uuids":   cloudclient.JsonRaw(statusRaw),
			"details":           detailsOut,
			"details_total":     detailsTotal,
			"details_truncated": detailsTotal > len(detailsOut),
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
	concurrent.RunIndexed(len(uuids), maxConcurrentUUIDs, func(idx int) {
		uuid := uuids[idx]
		lookup := lookupTaskByUUID(d, uuid)
		taskID := lookup.TaskID
		taskKind := lookup.TaskKind
		if taskID <= 0 {
			return
		}

		dRaw, dst, de := deps.TrackNamedCloudCall("task.download_result", d, func() ([]byte, int, error) {
			return d.Cloud().Task().DownloadResult(d.Ctx, d.PlatformID(), d.Token, taskID)
		})
		if de != nil || dst >= 400 {
			return
		}

		res := uuidDownloadResult{}
		res.raw = append(res.raw, cloudclient.JsonRaw(dRaw))
		if fp := cloudclient.ExtractFrontFilePath(dRaw); fp != "" {
			res.links = append(res.links, gin.H{"uuid": uuid, "task_id": taskID, "task_kind": taskKind, "file_path": fp})
		}
		resultsByIndex[idx] = res
	})

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
		concurrent.RunIndexed(len(moduleIDs), maxConcurrentModuleIDs, func(idx int) {
			moduleID := moduleIDs[idx]
			mRaw, mst, me := deps.TrackNamedCloudCall("task.module_result_url", d, func() ([]byte, int, error) {
				return d.Cloud().Task().ModuleResultURL(d.Ctx, d.PlatformID(), d.Token, moduleID)
			})
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
		})

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
	linksOut, linksTotal := limitcap.CapGinHSlice(links, limitcap.DefaultVisibleItems)
	rawOut, rawTotal := limitcap.CapAnySlice(rawList, limitcap.DefaultVisibleItems)
	_ = reply.RespondCloudSummary(r, d, intentKey, summaryRaw, summary.SummarizeTaskData, gin.H{
		"uuids":              uuids,
		"links":              linksOut,
		"links_total":        linksTotal,
		"links_truncated":    linksTotal > len(linksOut),
		"raw_cloud_json":     rawOut,
		"raw_list_total":     rawTotal,
		"raw_list_truncated": rawTotal > len(rawOut),
	})
}

func lookupTaskByUUID(d *deps.Deps, uuid string) taskLookupResult {
	result := taskLookupResult{TaskKind: "tool"}
	if d == nil || d.Cloud() == nil {
		return result
	}
	taskFilters := parse.ExtractTaskFiltersFromQuestion(d.Question(), uuid)
	listRaw, st, e := deps.TrackNamedCloudCall("task.list_tool_by_uuid", d, func() ([]byte, int, error) {
		return d.Cloud().Task().ListToolByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
	})
	if e == nil && st < 400 {
		result.ToolRaw = listRaw
		result.TaskID = cloudclient.ExtractFirstIDFromFrontList(listRaw)
	}
	workflowRaw, wst, we := deps.TrackNamedCloudCall("task.list_workflow_by_uuid", d, func() ([]byte, int, error) {
		return d.Cloud().Task().ListWorkflowByUUID(d.Ctx, d.PlatformID(), d.Token, taskFilters)
	})
	if we == nil && wst < 400 {
		result.WorkflowRaw = workflowRaw
		if result.TaskID <= 0 {
			result.TaskID = cloudclient.ExtractFirstIDFromFrontList(workflowRaw)
			if result.TaskID > 0 {
				result.TaskKind = "workflow"
			}
		}
	}
	return result
}
