package flow

import (
	"context"
	"net/http"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/backends"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	intranetclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/client"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"

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

// FrontFailed、FrontSuccess 实现 deps.Responder，供 platform 包调用。
func (p *Processor) FrontFailed(ctx *gin.Context, showMsg string, err error) {
	p.frontFailed(ctx, showMsg, err)
}

func (p *Processor) FrontSuccess(ctx *gin.Context, showMsg string, data interface{}) {
	p.frontSuccess(ctx, showMsg, data)
}

func (p *Processor) initPlatformClients(cfg *Config.AiGatewayConfig, platform string) (*http.Client, *cloudclient.API, *intranetclient.API) {
	switch platform {
	case Requests.AiPlatformCloudIntranet:
		hc := intranetclient.NewHTTPClient(cfg)
		return hc, nil, intranetclient.NewAPI(cfg, hc)
	case Requests.AiPlatformCloudPublic:
		hc := cloudclient.NewHTTPClient(cfg)
		return hc, cloudclient.NewAPI(cfg, hc), nil
	default:
		hc := cloudclient.NewHTTPClient(cfg)
		return hc, nil, nil
	}
}

func (p *Processor) ProcessChat(actx context.Context, ginCtx *gin.Context, req *Requests.AiRobotChatRequest, token string, cfg *Config.AiGatewayConfig) {
	if cfg == nil {
		cfg = &Config.AiGatewayConfig{}
		cfg.SetDefaults()
	}
	hc, publicAPI, intranetAPI := p.initPlatformClients(cfg, req.Platform)
	llmClient := llm.NewClient(cfg, hc)
	b, err := backends.NewForPlatform(req.Platform, publicAPI, intranetAPI)
	if err != nil {
		p.frontFailed(ginCtx, "平台后端初始化失败", err)
		return
	}
	d := &deps.Deps{
		Ctx:     actx,
		Gin:     ginCtx,
		Req:     req,
		Token:   token,
		Cfg:     cfg,
		LLM:     llmClient,
		Backend: b,
	}
	p.dispatchByRegistry(d)
}
