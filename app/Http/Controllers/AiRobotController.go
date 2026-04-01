package Controllers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	AiRobotFlow "cloud-platform-api/app/Services/ai_robot/flow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AiRobotController struct {
	Controller
	conversationService *Services.AiRobotConversationService
}

type aiRobotConversationListItem struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id,omitempty"`
	ConversationID     string    `json:"conversation_id"`
	Platform           string    `json:"platform"`
	QuestionType       int       `json:"question_type"`
	LastMessageID      string    `json:"last_message_id,omitempty"`
	LastIntent         string    `json:"last_intent,omitempty"`
	LastQuestion       string    `json:"last_question,omitempty"`
	LastResolved       string    `json:"last_resolved_question,omitempty"`
	ContextSummary     string    `json:"context_summary,omitempty"`
	LastProjectNumber  string    `json:"last_project_number,omitempty"`
	LastContractNumber string    `json:"last_contract_number,omitempty"`
	LastTaskUUID       string    `json:"last_task_uuid,omitempty"`
	LastArticleNameCN  string    `json:"last_article_name_cn,omitempty"`
	LastJournalName    string    `json:"last_journal_name,omitempty"`
	LastTaskStatus     string    `json:"last_task_status,omitempty"`
	LastDownloadKind   string    `json:"last_download_kind,omitempty"`
	LastHasLink        bool      `json:"last_has_link"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// NewAiRobotController 仅负责创建控制器实例。
// 入口在 Services/ai_robot/flow；各平台业务在 Services/ai_robot/platform/<平台>/。
func NewAiRobotController() *AiRobotController {
	return &AiRobotController{
		conversationService: Services.NewAiRobotConversationService(),
	}
}

// respondFrontFormat 统一返回前端约定结构。
// 约定：HTTP 状态固定 200，业务成功失败由 code 表示。
func (c *AiRobotController) respondFrontFormat(ctx *gin.Context, code int, showMsg string, debugMsg string, content interface{}) {
	if content == nil {
		content = gin.H{}
	}
	content = enrichAiRobotContent(ctx, content)
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
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID != "" {
		req.EnableContext = true
	}
	if strings.TrimSpace(req.ConversationID) == "" {
		req.ConversationID = uuid.NewString()
	}
	if strings.TrimSpace(req.MessageID) == "" {
		req.MessageID = uuid.NewString()
	}
	ctx.Set("ai_robot_conversation_id", req.ConversationID)
	ctx.Set("ai_robot_message_id", req.MessageID)
	if req.EnableContext && req.UserID == "" {
		c.frontFailed(ctx, "启用上下文时 user_id 不能为空", nil)
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
		// ai_robot 单次请求内通常包含多次云平台 HTTP 调用 + 1 次 LLM；
		// 若仅用 LLMTimeoutSec 作为“请求总时长”，在多 uuid/多分支场景下容易提前超时。
		// 这里给总时长留出缓冲，但仍以 LLMTimeoutSec 为基准控制上限。
		totalTimeoutSec := cfg.LLMTimeoutSec * 3
		if totalTimeoutSec < cfg.LLMTimeoutSec {
			// 防止整型溢出（理论上不会发生，但保持健壮性）
			totalTimeoutSec = cfg.LLMTimeoutSec
		}
		actx, cancel = context.WithTimeout(ctx.Request.Context(), time.Duration(totalTimeoutSec)*time.Second)
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

// ListConversations 返回 ai_robot 会话列表，便于排障和后台管理。
func (c *AiRobotController) ListConversations(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	req := Requests.AiRobotConversationListRequest{
		UserID:         ctx.Query("user_id"),
		Platform:       ctx.Query("platform"),
		ConversationID: ctx.Query("conversation_id"),
		QuestionType:   parseIntOrZero(ctx.Query("question_type")),
		StartTime:      ctx.Query("start_time"),
		EndTime:        ctx.Query("end_time"),
		Page:           page,
		Limit:          limit,
	}
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}

	items, total, err := c.conversationService.List(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		cfg.ConversationMongoCollection,
		Services.AiRobotConversationListFilter{
			UserID:         req.UserID,
			Platform:       req.Platform,
			ConversationID: req.ConversationID,
			QuestionType:   req.QuestionType,
			StartAt:        mustParseConversationListTime(req.StartTime),
			EndAt:          mustParseConversationListTime(req.EndTime),
			Page:           int64(req.Page),
			Limit:          int64(req.Limit),
		},
	)
	if err != nil {
		c.frontFailed(ctx, "读取会话列表失败", err)
		return
	}

	c.frontSuccess(ctx, "操作成功", gin.H{
		"items": buildConversationListItems(items),
		"pagination": gin.H{
			"page":  req.Page,
			"limit": req.Limit,
			"total": total,
			"pages": (total + int64(req.Limit) - 1) / int64(req.Limit),
		},
	})
}

// GetConversation 返回单条 ai_robot 会话，用于排障和后续复用。
func (c *AiRobotController) GetConversation(ctx *gin.Context) {
	var req Requests.AiRobotConversationManageRequest
	req.UserID = ctx.Query("user_id")
	req.Platform = ctx.Query("platform")
	req.ConversationID = ctx.Query("conversation_id")
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}

	doc, err := c.conversationService.FindByConversation(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		cfg.ConversationMongoCollection,
		req.UserID,
		req.Platform,
		req.ConversationID,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.frontFailed(ctx, "会话不存在", err)
			return
		}
		c.frontFailed(ctx, "读取会话失败", err)
		return
	}
	c.frontSuccess(ctx, "操作成功", doc)
}

// DeleteConversation 删除单条 ai_robot 会话。
func (c *AiRobotController) DeleteConversation(ctx *gin.Context) {
	var req Requests.AiRobotConversationManageRequest
	req.UserID = ctx.Query("user_id")
	req.Platform = ctx.Query("platform")
	req.ConversationID = ctx.Query("conversation_id")
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}

	err := c.conversationService.DeleteByConversation(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		cfg.ConversationMongoCollection,
		req.UserID,
		req.Platform,
		req.ConversationID,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.frontFailed(ctx, "会话不存在", err)
			return
		}
		c.frontFailed(ctx, "删除会话失败", err)
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"platform":        strings.TrimSpace(req.Platform),
		"conversation_id": strings.TrimSpace(req.ConversationID),
	})
}

// DeleteConversations 按条件批量删除 ai_robot 会话。
func (c *AiRobotController) DeleteConversations(ctx *gin.Context) {
	req := Requests.AiRobotConversationCleanupRequest{
		UserID:       ctx.Query("user_id"),
		Platform:     ctx.Query("platform"),
		QuestionType: parseIntOrZero(ctx.Query("question_type")),
		StartTime:    ctx.Query("start_time"),
		EndTime:      ctx.Query("end_time"),
	}
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return
	}

	deletedCount, err := c.conversationService.DeleteByFilter(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		cfg.ConversationMongoCollection,
		Services.AiRobotConversationCleanupFilter{
			UserID:       req.UserID,
			Platform:     req.Platform,
			QuestionType: req.QuestionType,
			StartAt:      mustParseConversationListTime(req.StartTime),
			EndAt:        mustParseConversationListTime(req.EndTime),
		},
	)
	if err != nil {
		c.frontFailed(ctx, "批量删除会话失败", err)
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"deleted_count": deletedCount,
		"filters": gin.H{
			"platform":      req.Platform,
			"question_type": req.QuestionType,
			"start_time":    req.StartTime,
			"end_time":      req.EndTime,
		},
	})
}

func effectiveConversationDatabase(cfg *Config.AiGatewayConfig) string {
	if cfg == nil {
		return ""
	}
	effective := cfg.EffectiveConversationMongoConfig(Config.GetMongoDBConfig())
	return effective.Database
}

func buildConversationListItems(items []Models.AiRobotConversation) []aiRobotConversationListItem {
	out := make([]aiRobotConversationListItem, 0, len(items))
	for _, item := range items {
		lastTaskUUID := ""
		if len(item.LastTaskUUIDs) > 0 {
			lastTaskUUID = item.LastTaskUUIDs[0]
		}
		out = append(out, aiRobotConversationListItem{
			ID:                 item.ID,
			UserID:             item.UserID,
			ConversationID:     item.ConversationID,
			Platform:           item.Platform,
			QuestionType:       item.QuestionType,
			LastMessageID:      item.LastMessageID,
			LastIntent:         item.LastIntent,
			LastQuestion:       item.LastQuestion,
			LastResolved:       item.LastResolved,
			ContextSummary:     item.ContextSummary,
			LastProjectNumber:  item.LastProjectNumber,
			LastContractNumber: item.LastContractNumber,
			LastTaskUUID:       lastTaskUUID,
			LastArticleNameCN:  item.LastArticleNameCN,
			LastJournalName:    item.LastJournalName,
			LastTaskStatus:     item.LastTaskStatus,
			LastDownloadKind:   item.LastDownloadKind,
			LastHasLink:        item.LastHasLink,
			UpdatedAt:          item.UpdatedAt,
		})
	}
	return out
}

func enrichAiRobotContent(ctx *gin.Context, content interface{}) gin.H {
	base := gin.H{}
	if existing, ok := content.(gin.H); ok {
		for k, v := range existing {
			base[k] = v
		}
	} else {
		base["data"] = content
	}
	if ctx != nil {
		if conversationID := strings.TrimSpace(ctx.GetString("ai_robot_conversation_id")); conversationID != "" {
			base["conversation_id"] = conversationID
		}
		if messageID := strings.TrimSpace(ctx.GetString("ai_robot_message_id")); messageID != "" {
			base["message_id"] = messageID
		}
	}
	return base
}

func parseIntOrZero(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	n, _ := strconv.Atoi(value)
	return n
}

func mustParseConversationListTime(value string) time.Time {
	t, _ := Requests.ParseConversationListTimeForController(value)
	return t
}
