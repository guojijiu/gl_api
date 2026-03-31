package project

// 图片对比平台「项目」域（占位）：与公网/内网云目录并列，后续在此独立实现。

import (
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"

	"github.com/gin-gonic/gin"
)

// Dispatch 图片对比平台 project question_type 处理入口（占位：目前能力未接入）。
func Dispatch(r deps.Responder, d *deps.Deps) {
	answer, err := summary.SummarizeUnsupported(d.Ctx, d.LLM, d.Req.Question, "图片对比平台能力尚未接入")
	if err != nil {
		r.FrontFailed(d.Gin, "生成回复失败", err)
		return
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intent.IntentUnsupported,
		"cloud_called":   false,
		"strict_mode":    d.Cfg.StrictMode,
		"raw_cloud_json": nil,
	})
}
