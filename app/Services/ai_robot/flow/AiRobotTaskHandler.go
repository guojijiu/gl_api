package flow

import (
	"encoding/json"
	"fmt"
	"strings"

	AiGateway "cloud-platform-api/app/Services/ai_gateway"

	"github.com/gin-gonic/gin"
)

func (p *Processor) handleTaskDomain(d *aiRobotDeps) {
	// 任务域要求问题里带 uuid；否则无法定位任务。
	plan, err := d.svc.PlanTaskIntent(d.ctx, d.req.Question)
	if err != nil {
		p.frontFailed(d.gin, "意图解析失败", err)
		return
	}
	uuids := d.svc.ExtractTaskUUIDsFromQuestion(d.req.Question)
	if len(uuids) == 0 {
		p.frontFailed(d.gin, "请在问题中提供任务编号（uuid），例如：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", nil)
		return
	}
	uuidsCSV := strings.Join(uuids, ",")

	handlers := p.taskIntentRegistry()
	handler := handlers[plan.Intent]
	if handler == nil {
		p.frontFailed(d.gin, "未知意图分支", nil)
		return
	}
	if !isIntentAllowed(d.req.Platform, d.req.QuestionType, plan.Intent) {
		p.frontFailed(d.gin, "当前平台下不支持该任务能力", nil)
		return
	}
	handler(d, plan, uuids, uuidsCSV)
}

func (p *Processor) taskIntentRegistry() map[string]func(*aiRobotDeps, *AiGateway.IntentPlan, []string, string) {
	return map[string]func(*aiRobotDeps, *AiGateway.IntentPlan, []string, string){
		AiGateway.IntentUnsupported: func(d *aiRobotDeps, plan *AiGateway.IntentPlan, _ []string, _ string) {
			answer, e := d.svc.SummarizeUnsupported(d.ctx, d.req.Question, plan.Reason)
			if e != nil {
				p.frontFailed(d.gin, "生成回复失败", e)
				return
			}
			p.frontSuccess(d.gin, "操作成功", gin.H{
				"answer":         answer,
				"intent":         plan.Intent,
				"cloud_called":   false,
				"strict_mode":    d.cfg.StrictMode,
				"raw_cloud_json": nil,
			})
		},
		AiGateway.IntentTaskStatus: func(d *aiRobotDeps, plan *AiGateway.IntentPlan, uuids []string, uuidsCSV string) {
			p.handleTaskStatus(d, plan.Intent, uuids, uuidsCSV)
		},
		AiGateway.IntentTaskDownloadResult: func(d *aiRobotDeps, plan *AiGateway.IntentPlan, uuids []string, _ string) {
			p.handleTaskDownload(d, plan.Intent, uuids)
		},
	}
}

func (p *Processor) handleTaskStatus(d *aiRobotDeps, intent string, uuids []string, uuidsCSV string) {
	// 总体状态（批量）+ 任务详情（逐条）两层信息合并后再给大模型总结。
	statusRaw, statusCode, err := d.task().StatusByUUIDs(d.ctx, d.req.Platform, d.token, uuidsCSV)
	if err != nil {
		p.frontFailed(d.gin, "查询任务状态失败", err)
		return
	}
	if statusCode >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", statusCode), nil)
		return
	}

	var details []gin.H
	for _, u := range uuids {
		listRaw, st, e := d.task().ListToolByUUID(d.ctx, d.req.Platform, d.token, u)
		if e != nil || st >= 400 {
			listRaw = nil
		}
		workflowRaw, wst, we := d.task().ListWorkflowByUUID(d.ctx, d.req.Platform, d.token, u)
		if we != nil || wst >= 400 {
			workflowRaw = nil
		}
		taskID := AiGateway.ExtractFirstIDFromFrontList(listRaw)
		if taskID <= 0 {
			taskID = AiGateway.ExtractFirstIDFromFrontList(workflowRaw)
		}

		var result interface{}
		if taskID > 0 {
			rRaw, rst, re := d.task().Result(d.ctx, d.req.Platform, d.token, taskID)
			if re == nil && rst < 400 {
				result = AiGateway.JsonRaw(rRaw)
			}
		}
		details = append(details, gin.H{
			"uuid":          u,
			"task_id":       taskID,
			"tool_task":     AiGateway.JsonRaw(listRaw),
			"workflow_task": AiGateway.JsonRaw(workflowRaw),
			"result":        result,
		})
	}

	moduleIDs := d.svc.ExtractNumericIDsFromQuestion(d.req.Question)
	for _, mid := range moduleIDs {
		mRaw, mst, me := d.task().ModuleDetail(d.ctx, d.req.Platform, d.token, mid)
		if me != nil || mst >= 400 {
			continue
		}
		details = append(details, gin.H{"module_task_id": mid, "module_detail": AiGateway.JsonRaw(mRaw)})
	}

	summaryRaw, _ := json.Marshal(gin.H{"uuids": uuids, "status_batch": AiGateway.JsonRaw(statusRaw), "details": details})
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, summaryRaw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":       answer,
		"intent":       intent,
		"cloud_called": true,
		"strict_mode":  d.cfg.StrictMode,
		"uuids":        uuids,
		"raw_cloud_json": gin.H{
			"status_by_uuids": AiGateway.JsonRaw(statusRaw),
			"details":         details,
		},
	})
}

func (p *Processor) handleTaskDownload(d *aiRobotDeps, intent string, uuids []string) {
	// 对每个 uuid 尝试解析工具任务/流程任务并生成下载链接。
	var links []gin.H
	var rawList []interface{}
	for _, u := range uuids {
		listRaw, st, e := d.task().ListToolByUUID(d.ctx, d.req.Platform, d.token, u)
		if e != nil || st >= 400 {
			listRaw = nil
		}
		workflowRaw, wst, we := d.task().ListWorkflowByUUID(d.ctx, d.req.Platform, d.token, u)
		if we != nil || wst >= 400 {
			workflowRaw = nil
		}
		taskID := AiGateway.ExtractFirstIDFromFrontList(listRaw)
		taskKind := "tool"
		if taskID <= 0 {
			taskID = AiGateway.ExtractFirstIDFromFrontList(workflowRaw)
			taskKind = "workflow"
		}
		if taskID <= 0 {
			continue
		}
		dRaw, dst, de := d.task().DownloadResult(d.ctx, d.req.Platform, d.token, taskID)
		if de != nil || dst >= 400 {
			continue
		}
		rawList = append(rawList, AiGateway.JsonRaw(dRaw))
		if fp := AiGateway.ExtractFrontFilePath(dRaw); fp != "" {
			links = append(links, gin.H{"uuid": u, "task_id": taskID, "task_kind": taskKind, "file_path": fp})
		}
	}
	moduleIDs := d.svc.ExtractNumericIDsFromQuestion(d.req.Question)
	for _, mid := range moduleIDs {
		mRaw, mst, me := d.task().ModuleResultURL(d.ctx, d.req.Platform, d.token, mid)
		if me != nil || mst >= 400 {
			continue
		}
		rawList = append(rawList, AiGateway.JsonRaw(mRaw))
		if fp := AiGateway.ExtractAnyURL(mRaw); fp != "" {
			links = append(links, gin.H{"task_id": mid, "task_kind": "module", "file_path": fp})
		}
	}
	if len(rawList) == 0 {
		p.frontFailed(d.gin, "下载链接生成失败，请确认任务编号/权限", nil)
		return
	}
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, summaryRaw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intent,
		"cloud_called":   true,
		"strict_mode":    d.cfg.StrictMode,
		"uuids":          uuids,
		"links":          links,
		"raw_cloud_json": rawList,
	})
}
