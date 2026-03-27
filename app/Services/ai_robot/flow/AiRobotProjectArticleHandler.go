package flow

import (
	AiGateway "cloud-platform-api/app/Services/ai_gateway"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (p *Processor) handleProjectArticleDomain(d *aiRobotDeps) {
	plan, err := d.svc.PlanProjectArticleIntent(d.ctx, d.req.Question)
	if err != nil {
		p.frontFailed(d.gin, "意图解析失败", err)
		return
	}

	handlers := p.projectArticleIntentRegistry()
	handler := handlers[plan.Intent]
	if handler == nil {
		p.frontFailed(d.gin, "未知意图分支", nil)
		return
	}
	if !isIntentAllowed(d.req.Platform, d.req.QuestionType, plan.Intent) {
		p.frontFailed(d.gin, "当前平台下不支持该任务能力", nil)
		return
	}
	handler(d, plan)
}

func (p *Processor) projectArticleIntentRegistry() map[string]func(*aiRobotDeps, *AiGateway.IntentPlan) {
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
		AiGateway.IntentProjectArticleList: func(d *aiRobotDeps, plan *AiGateway.IntentPlan) {
			p.handleProjectArticleList(d, plan.Intent)
		},
	}
}

func (p *Processor) handleProjectArticleList(d *aiRobotDeps, intent string) {
	raw, status, err := d.projectArticle().List(d.ctx, d.req.Platform, d.token)
	if err != nil {
		p.frontFailed(d.gin, "请求云平台失败", err)
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
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intent,
		"cloud_called":   true,
		"strict_mode":    d.cfg.StrictMode,
		"raw_cloud_json": AiGateway.JsonRaw(raw),
	})

}
