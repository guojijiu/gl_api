package flow

import (
	"context"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	AiGateway "cloud-platform-api/app/Services/ai_gateway"

	"github.com/gin-gonic/gin"
)

type Processor struct{}

// Processor 是 ai_robot 的业务编排入口：
// - 接收 Controller 已校验过的请求
// - 初始化本次请求的依赖集合（deps）
// - 分发到平台/业务/意图处理链
func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) respondFrontFormat(ctx *gin.Context, code int, showMsg string, debugMsg string, content interface{}) {
	if content == nil {
		content = gin.H{}
	}
	ctx.JSON(200, gin.H{
		"code":     code,
		"showMsg":  showMsg,
		"debugMsg": debugMsg,
		"content":  content,
	})
}

func (p *Processor) frontSuccess(ctx *gin.Context, showMsg string, data interface{}) {
	if showMsg == "" {
		showMsg = "操作成功"
	}
	p.respondFrontFormat(ctx, 1, showMsg, "", gin.H{"data": data})
}

func (p *Processor) frontFailed(ctx *gin.Context, showMsg string, err error) {
	debugMsg := ""
	if err != nil {
		debugMsg = err.Error()
	}
	if showMsg == "" {
		showMsg = "操作失败"
	}
	p.respondFrontFormat(ctx, 0, showMsg, debugMsg, gin.H{})
}

func (p *Processor) ProcessChat(actx context.Context, ginCtx *gin.Context, req *Requests.AiRobotChatRequest, token string, cfg *Config.AiGatewayConfig) {
	// 仅初始化基础服务；具体的 project/contract/task client 采用懒初始化。
	svc := AiGateway.NewService(cfg)
	deps := &aiRobotDeps{
		ctx:   actx,
		gin:   ginCtx,
		req:   req,
		token: token,
		cfg:   cfg,
		svc:   svc,
	}
	p.dispatchByRegistry(deps)
}
