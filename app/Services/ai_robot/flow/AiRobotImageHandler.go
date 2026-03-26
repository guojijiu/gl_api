package flow

import (
	AiGateway "cloud-platform-api/app/Services/ai_gateway"

	"github.com/gin-gonic/gin"
)

func (p *Processor) handleImageDomain(d *aiRobotDeps) {
	answer, err := d.svc.SummarizeUnsupported(d.ctx, d.req.Question, "图片对比平台能力尚未接入")
	if err != nil {
		p.frontFailed(d.gin, "生成回复失败", err)
		return
	}
	p.frontSuccess(d.gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         AiGateway.IntentUnsupported,
		"cloud_called":   false,
		"strict_mode":    d.cfg.StrictMode,
		"raw_cloud_json": nil,
	})
}
