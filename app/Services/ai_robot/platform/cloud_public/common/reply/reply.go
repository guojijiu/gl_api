package reply

import (
	"context"

	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/limitcap"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"

	"github.com/gin-gonic/gin"
)

func RespondUnsupported(r deps.Responder, d *deps.Deps, intent string, reason string) bool {
	answer, err := summary.SummarizeUnsupported(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), reason)
	if err != nil {
		r.FrontFailed(d.Gin, "生成回复失败", err)
		return false
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intent,
		"cloud_called":   false,
		"strict_mode":    d.Cfg.StrictMode,
		"raw_cloud_json": nil,
	})
	return true
}

func BuildCloudSuccessPayload(d *deps.Deps, intent string, answer string, extra gin.H) gin.H {
	resp := gin.H{
		"answer":       answer,
		"intent":       intent,
		"cloud_called": true,
	}
	if d != nil && d.Cfg != nil {
		resp["strict_mode"] = d.Cfg.StrictMode
	}
	for k, v := range extra {
		resp[k] = v
	}
	return resp
}

type JSONSummarizer func(context.Context, llm.ChatCompletionClient, string, string, []byte) (string, error)

func RespondCloudSummary(r deps.Responder, d *deps.Deps, intent string, raw []byte, summarize JSONSummarizer, extra gin.H) bool {
	if summarize == nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", nil)
		return false
	}
	d.RememberIntent(intent)
	answer, err := summarize(d.Ctx, d.LLM, d.Question(), d.ContextSummary(), raw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return false
	}
	r.FrontSuccess(d.Gin, "操作成功", BuildCloudSuccessPayload(d, intent, answer, extra))
	return true
}

func RespondCloudListSummary(r deps.Responder, d *deps.Deps, intent string, raw []byte, summarize JSONSummarizer) bool {
	displayRaw, listTotal := limitcap.CapCloudListResponseJSON(raw, limitcap.DefaultVisibleItems)
	extra := gin.H{
		"raw_cloud_json": cloudclient.JsonRaw(displayRaw),
	}
	if listTotal > limitcap.DefaultVisibleItems {
		extra["content_list_total"] = listTotal
		extra["content_list_truncated"] = true
	}
	return RespondCloudSummary(r, d, intent, raw, summarize, extra)
}
