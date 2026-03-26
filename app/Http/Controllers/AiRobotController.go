package Controllers

import (
	"context"
	"strings"
	"time"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	AiRobotFlow "cloud-platform-api/app/Services/ai_robot/flow"

	"github.com/gin-gonic/gin"
)

type AiRobotController struct {
	Controller
}

// NewAiRobotController 仅负责创建控制器实例。
// 业务编排（平台分流/意图分流/云平台调用）都在 Services/ai_robot/flow。
func NewAiRobotController() *AiRobotController {
	return &AiRobotController{}
}

// respondFrontFormat 统一返回前端约定结构。
// 约定：HTTP 状态固定 200，业务成功失败由 code 表示。
func (c *AiRobotController) respondFrontFormat(ctx *gin.Context, code int, showMsg string, debugMsg string, content interface{}) {
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

func (c *AiRobotController) frontSuccess(ctx *gin.Context, showMsg string, data interface{}) {
	if showMsg == "" {
		showMsg = "操作成功"
	}
	c.respondFrontFormat(ctx, 1, showMsg, "", gin.H{"data": data})
}

func (c *AiRobotController) frontFailed(ctx *gin.Context, showMsg string, err error) {
	debugMsg := ""
	if err != nil {
		debugMsg = err.Error()
	}
	if showMsg == "" {
		showMsg = "操作失败"
	}
	c.respondFrontFormat(ctx, 0, showMsg, debugMsg, gin.H{})
}

func (c *AiRobotController) Chat(ctx *gin.Context) {
	// 第1步：请求体校验（字段存在性 + 业务可支持性）
	var req Requests.AiRobotChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}
	token := strings.TrimSpace(ctx.GetHeader("Token"))
	if token == "" {
		c.frontFailed(ctx, "请在 Header 中携带云平台用户 token", nil)
		return
	}

	// 第2步：读取 AI 网关配置（包含云平台地址、大模型配置、能力矩阵等）
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}
	actx := ctx.Request.Context()
	if cfg.LLMTimeoutSec > 0 {
		var cancel context.CancelFunc
		actx, cancel = context.WithTimeout(ctx.Request.Context(), time.Duration(cfg.LLMTimeoutSec)*time.Second)
		defer cancel()
	}
	// 第3步：进入服务层编排（控制器不承载业务细节）
	AiRobotFlow.NewProcessor().ProcessChat(actx, ctx, &req, token, cfg)
}

// Capabilities 返回当前能力矩阵与关键配置，用于联调/排障。
func (c *AiRobotController) Capabilities(ctx *gin.Context) {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"strict_mode":         cfg.StrictMode,
		"capability_matrix":   cfg.CapabilityMatrixDebugView(),
		"llm_model":           cfg.LLMModel,
		"cloud_public_base":   cfg.CloudFrontAPIBase(),
		"cloud_intranet_base": cfg.CloudIntranetAPIBase(),
	})
}
