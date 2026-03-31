package deps

import (
	"context"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/backends"

	"github.com/gin-gonic/gin"
)

// Responder 为平台业务处理与 Processor 之间的响应契约，避免 platform 包依赖 flow 产生循环引用。
type Responder interface {
	FrontFailed(*gin.Context, string, error)
	FrontSuccess(*gin.Context, string, interface{})
}

// Deps 单次 AI 对话请求的上下文：HTTP、鉴权、大模型客户端、按平台注入的 Backend。
type Deps struct {
	Ctx   context.Context
	Gin   *gin.Context
	Req   *Requests.AiRobotChatRequest
	Token string
	Cfg   *Config.AiGatewayConfig
	LLM   llm.ChatCompletionClient

	Backend backends.Backend
}

// PlatformID 来自 Backend，与请求 platform 一致；策略/路由统一走此值，避免与 Req 漂移。
func (d *Deps) PlatformID() string {
	if d.Backend == nil {
		return ""
	}
	return d.Backend.PlatformID()
}

// Cloud 返回绑定到当前请求 platform 的云平台网关（含独立 client 组）；非云平台路由返回 nil。
func (d *Deps) Cloud() *backends.CloudGateway {
	g, _ := d.Backend.(*backends.CloudGateway)
	return g
}
