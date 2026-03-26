package flow

import (
	"encoding/json"
	"fmt"

	AiGateway "cloud-platform-api/app/Services/ai_gateway"

	"github.com/gin-gonic/gin"
)

func (p *Processor) handleProjectDomain(d *aiRobotDeps) {
	// 先让大模型判断用户意图，再落到具体处理分支。
	plan, err := d.svc.PlanProjectIntent(d.ctx, d.req.Question)
	if err != nil {
		p.frontFailed(d.gin, "意图解析失败", err)
		return
	}
	handlers := p.projectIntentRegistry()
	handler := handlers[plan.Intent]
	if handler == nil {
		p.frontFailed(d.gin, "未知意图分支", nil)
		return
	}
	if !isIntentAllowed(d.req.Platform, d.req.QuestionType, plan.Intent) {
		p.frontFailed(d.gin, "当前平台下不支持该项目能力", nil)
		return
	}
	handler(d, plan)
}

func (p *Processor) projectIntentRegistry() map[string]func(*aiRobotDeps, *AiGateway.IntentPlan) {
	return map[string]func(*aiRobotDeps, *AiGateway.IntentPlan){
		AiGateway.IntentUnsupported: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
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
		AiGateway.IntentProjectList: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleProjectListOrExpiring(d, plan.Intent)
		},
		AiGateway.IntentProjectExpiring: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleProjectListOrExpiring(d, plan.Intent)
		},
		AiGateway.IntentContractList: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleContractList(d, plan.Intent)
		},
		AiGateway.IntentContractProjects: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleContractProjects(d, plan.Intent)
		},
		AiGateway.IntentDownloadFinal: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleDownloadFinalReport(d, plan.Intent)
		},
		AiGateway.IntentDownloadOriginal: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleDownloadOriginalData(d, plan.Intent)
		},
	}
}

func (p *Processor) handleProjectListOrExpiring(d *aiRobotDeps, intent string) {
	raw, status, err := d.project().List(d.ctx, d.req.Platform, d.token)
	if err != nil {
		p.frontFailed(d.gin, "请求云平台失败", err)
		return
	}
	if status >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return
	}
	if intent == AiGateway.IntentProjectExpiring {
		raw, _ = d.svc.FilterExpiringProjects(raw)
	}
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, raw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{"answer": answer, "intent": intent, "cloud_called": true, "strict_mode": d.cfg.StrictMode, "raw_cloud_json": AiGateway.JsonRaw(raw)})
}

func (p *Processor) handleContractList(d *aiRobotDeps, intent string) {
	raw, status, err := d.contract().List(d.ctx, d.req.Platform, d.token, "")
	if err != nil {
		p.frontFailed(d.gin, "查询合同列表失败", err)
		return
	}
	if status >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return
	}
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, raw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{"answer": answer, "intent": intent, "cloud_called": true, "strict_mode": d.cfg.StrictMode, "raw_cloud_json": AiGateway.JsonRaw(raw)})
}

func (p *Processor) handleContractProjects(d *aiRobotDeps, intent string) {
	// 两阶段：先找合同，再按合同拉项目列表。
	contractsRaw, status, err := d.contract().List(d.ctx, d.req.Platform, d.token, "")
	if err != nil {
		p.frontFailed(d.gin, "查询合同列表失败", err)
		return
	}
	if status >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return
	}
	contractIDs, matchedContractNumber, err := d.svc.ResolveContractIDsFromContracts(d.ctx, d.req.Question, contractsRaw)
	if err != nil {
		p.frontFailed(d.gin, "自动匹配合同失败", err)
		return
	}
	var grouped []gin.H
	for _, cid := range contractIDs {
		raw, st, e := d.project().ListByContract(d.ctx, d.req.Platform, d.token, cid)
		if e != nil || st >= 400 {
			continue
		}
		grouped = append(grouped, gin.H{"contract_id": cid, "projects_result": AiGateway.JsonRaw(raw)})
	}
	if len(grouped) == 0 {
		p.frontFailed(d.gin, "查询合同下项目失败", nil)
		return
	}
	summaryRaw, _ := json.Marshal(gin.H{"contract_projects": grouped})
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, summaryRaw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":                  answer,
		"intent":                  intent,
		"cloud_called":            true,
		"strict_mode":             d.cfg.StrictMode,
		"matched_contract_ids":    contractIDs,
		"matched_contract_number": matchedContractNumber,
		"contract_projects":       grouped,
	})
}

func (p *Processor) handleDownloadFinalReport(d *aiRobotDeps, intent string) {
	projectsRaw, status, err := d.project().List(d.ctx, d.req.Platform, d.token)
	if err != nil {
		p.frontFailed(d.gin, "查询项目列表失败", err)
		return
	}
	if status >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return
	}
	projectIDs, matchedNumber, err := d.svc.ResolveProjectIDsFromProjects(d.ctx, d.req.Question, projectsRaw)
	if err != nil {
		p.frontFailed(d.gin, "自动匹配项目失败", err)
		return
	}
	var links []gin.H
	var rawList []interface{}
	for _, id := range projectIDs {
		raw, st, e := d.project().ZipURL(d.ctx, d.req.Platform, d.token, id)
		if e != nil || st >= 400 {
			continue
		}
		rawList = append(rawList, AiGateway.JsonRaw(raw))
		if u := AiGateway.ExtractFrontURL(raw); u != "" {
			links = append(links, gin.H{"project_id": id, "url": u})
		}
	}
	if len(rawList) == 0 {
		p.frontFailed(d.gin, "请求云平台失败", nil)
		return
	}
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, summaryRaw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":              answer,
		"intent":              intent,
		"cloud_called":        true,
		"strict_mode":         d.cfg.StrictMode,
		"matched_project_ids": projectIDs,
		"matched_number":      matchedNumber,
		"links":               links,
		"raw_cloud_json":      rawList,
	})
}

func (p *Processor) handleDownloadOriginalData(d *aiRobotDeps, intent string) {
	projectsRaw, status, err := d.project().List(d.ctx, d.req.Platform, d.token)
	if err != nil {
		p.frontFailed(d.gin, "查询项目列表失败", err)
		return
	}
	if status >= 400 {
		p.frontFailed(d.gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return
	}
	projectIDs, matchedNumber, err := d.svc.ResolveProjectIDsFromProjects(d.ctx, d.req.Question, projectsRaw)
	if err != nil {
		p.frontFailed(d.gin, "自动匹配项目失败", err)
		return
	}
	var links []gin.H
	var rawList []interface{}
	for _, id := range projectIDs {
		// 原始数据下载依赖 access_code，后端需要该参数生成临时下载地址。
		accessCode := generateAccessCode(6)
		raw, st, e := d.project().OriginalDataURL(d.ctx, d.req.Platform, d.token, id, accessCode)
		if e != nil || st >= 400 {
			continue
		}
		rawList = append(rawList, AiGateway.JsonRaw(raw))
		if u := AiGateway.ExtractFrontURL(raw); u != "" {
			links = append(links, gin.H{"project_id": id, "access_code": accessCode, "url": u})
		}
	}
	if len(rawList) == 0 {
		p.frontFailed(d.gin, "请求云平台失败", nil)
		return
	}
	summaryRaw, _ := json.Marshal(gin.H{"links": links, "raw": rawList})
	answer, err := d.svc.SummarizeWithData(d.ctx, d.req.Question, summaryRaw)
	if err != nil {
		p.frontFailed(d.gin, "生成自然语言回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":              answer,
		"intent":              intent,
		"cloud_called":        true,
		"strict_mode":         d.cfg.StrictMode,
		"matched_project_ids": projectIDs,
		"matched_number":      matchedNumber,
		"links":               links,
		"raw_cloud_json":      rawList,
	})
}
