package Controllers

import (
	"context"
	"errors"
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

var aiRobotShanghaiLocation = loadAiRobotShanghaiLocation()

type AiRobotController struct {
	Controller
	conversationService        *Services.AiRobotConversationService
	conversationMessageService *Services.AiRobotConversationMessageService
}

type aiRobotConversationListItem struct {
	ID                 string                          `json:"id"`
	UserID             string                          `json:"user_id,omitempty"`
	ConversationID     string                          `json:"conversation_id"`
	Platform           string                          `json:"platform"`
	QuestionType       int                             `json:"question_type"`
	LastMessageID      string                          `json:"last_message_id,omitempty"`
	LastIntent         string                          `json:"last_intent,omitempty"`
	LastQuestion       string                          `json:"last_question,omitempty"`
	LastResolved       string                          `json:"last_resolved_question,omitempty"`
	ContextSummary     string                          `json:"context_summary,omitempty"`
	LastProjectNumber  string                          `json:"last_project_number,omitempty"`
	LastContractNumber string                          `json:"last_contract_number,omitempty"`
	LastTaskUUID       string                          `json:"last_task_uuid,omitempty"`
	LastArticleNameCN  string                          `json:"last_article_name_cn,omitempty"`
	LastJournalName    string                          `json:"last_journal_name,omitempty"`
	LastTaskStatus     string                          `json:"last_task_status,omitempty"`
	LastDownloadKind   string                          `json:"last_download_kind,omitempty"`
	LastHasLink        bool                            `json:"last_has_link"`
	LatestMessage      *aiRobotConversationMessageItem `json:"latest_message,omitempty"`
	UpdatedAt          string                          `json:"updated_at,omitempty"`
}

type aiRobotConversationMessageItem struct {
	MessageID             string `json:"message_id,omitempty"`
	Question              string `json:"question"`
	Resolved              string `json:"resolved_question,omitempty"`
	NormalizedQuestion    string `json:"normalized_question,omitempty"`
	HitContext            bool   `json:"hit_context"`
	ContextSource         string `json:"context_source,omitempty"`
	ClarificationNeeded   bool   `json:"clarification_needed"`
	ClarificationReason   string `json:"clarification_reason,omitempty"`
	Answer                string `json:"answer,omitempty"`
	Intent                string `json:"intent,omitempty"`
	Success               bool   `json:"success"`
	ShowMsg               string `json:"show_msg,omitempty"`
	DebugMsg              string `json:"debug_msg,omitempty"`
	Stage                 string `json:"stage,omitempty"`
	ErrorType             string `json:"error_type,omitempty"`
	CloudAPI              string `json:"cloud_api,omitempty"`
	CloudStatusCode       int    `json:"cloud_status_code,omitempty"`
	LLMModel              string `json:"llm_model,omitempty"`
	RequestPayloadSummary string `json:"request_payload_summary,omitempty"`
	ResultKind            string `json:"result_kind,omitempty"`
	ResultCount           int    `json:"result_count,omitempty"`
	ResultBrief           string `json:"result_brief,omitempty"`
	ResponseSummary       string `json:"response_summary,omitempty"`
	ResponseSize          int64  `json:"response_size,omitempty"`
	CloudCalled           bool   `json:"cloud_called"`
	DurationMS            int64  `json:"duration_ms,omitempty"`
	LLMDurationMS         int64  `json:"llm_duration_ms,omitempty"`
	CloudDurationMS       int64  `json:"cloud_duration_ms,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
}

// NewAiRobotController 仅负责创建控制器实例。
// 入口在 Services/ai_robot/flow；各平台业务在 Services/ai_robot/platform/<平台>/。
func NewAiRobotController() *AiRobotController {
	return &AiRobotController{
		conversationService:        Services.NewAiRobotConversationService(),
		conversationMessageService: Services.NewAiRobotConversationMessageService(),
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

func (c *AiRobotController) respondMessageCleanupPartialSuccess(ctx *gin.Context, showMsg string, err error, data gin.H) {
	debugMsg := ""
	if err != nil {
		debugMsg = err.Error()
	}
	if data == nil {
		data = gin.H{}
	}
	data["message_cleanup_failed"] = true
	if debugMsg != "" {
		data["message_cleanup_error"] = debugMsg
	}
	c.respondFrontFormat(ctx, 1, showMsg, debugMsg, gin.H{"data": data})
}

func buildMessageCleanupSuccessData(data gin.H, deletedMessageCount int64) gin.H {
	if data == nil {
		data = gin.H{}
	}
	data["deleted_message_count"] = deletedMessageCount
	data["message_cleanup_failed"] = false
	return data
}

// getAiGatewayConfigOrFail 统一读取 AI 网关配置。
// 控制器内多个接口都依赖同一份配置，集中到一个 helper 后可减少重复分支。
func (c *AiRobotController) getAiGatewayConfigOrFail(ctx *gin.Context) *Config.AiGatewayConfig {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		c.frontFailed(ctx, "配置未初始化", nil)
		return nil
	}
	return cfg
}

// queryPageLimit 统一解析分页参数，并复用请求层的默认值语义。
// 这里不直接做边界收敛，最终仍交由各 Request.Validate() 负责。
func queryPageLimit(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	return page, limit
}

func buildPagination(page int, limit int, total int64) gin.H {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	hasMore := int64(page)*int64(limit) < total
	return gin.H{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"pages":    (total + int64(limit) - 1) / int64(limit),
		"has_more": hasMore,
	}
}

// buildPartialDataNotice 给分页接口返回一个“仅本页数据”的显式提示，
// 避免客户误以为接口只返回这些数据。
func buildPartialDataNotice(pagination gin.H) string {
	if pagination == nil {
		return ""
	}
	hasMore, _ := pagination["has_more"].(bool)
	if !hasMore {
		return ""
	}
	page, _ := pagination["page"].(int)
	limit, _ := pagination["limit"].(int)
	total, _ := pagination["total"].(int64)
	pages, _ := pagination["pages"].(int64)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if pages <= 0 && limit > 0 {
		pages = (total + int64(limit) - 1) / int64(limit)
	}
	if total <= 0 {
		return "当前为分页返回，仅展示本页数据"
	}
	return "当前为分页返回，仅展示本页数据；可继续翻页获取更多（第 " +
		strconv.Itoa(page) + " 页 / 共 " + strconv.FormatInt(pages, 10) + " 页，总计 " +
		strconv.FormatInt(total, 10) + " 条）"
}

func readConversationManageRequest(ctx *gin.Context) Requests.AiRobotConversationManageRequest {
	return Requests.AiRobotConversationManageRequest{
		UserID:         ctx.Query("user_id"),
		Platform:       ctx.Query("platform"),
		ConversationID: ctx.Query("conversation_id"),
		MessageID:      ctx.Query("message_id"),
	}
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
// 说明：
// 1. configured_capability_matrix 反映配置层声明，便于排查环境配置；
// 2. exposed_chat_capability_matrix 反映当前接口层实际对外开放的聊天能力；
// 3. 二者分开展示，避免“配置里写了，但代码尚未正式开放”造成误解。
func (c *AiRobotController) Capabilities(ctx *gin.Context) {
	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"strict_mode":                    cfg.StrictMode,
		"configured_capability_matrix":   cfg.CapabilityMatrixDebugView(),
		"exposed_chat_capability_matrix": buildExposedChatCapabilityView(),
		"llm_model":                      cfg.LLMModel,
		"cloud_public_base":              cfg.CloudFrontAPIBase(),
		"cloud_intranet_base":            cfg.CloudIntranetAPIBase(),
	})
}

func buildExposedChatCapabilityView() gin.H {
	return gin.H{
		Requests.AiPlatformCloudPublic: gin.H{
			"enabled":        true,
			"question_types": []int{Requests.AiQuestionTypeProject, Requests.AiQuestionTypeTask, Requests.AiQuestionTypeProjectArticle},
			"status":         "ready",
			"status_message": "公网云平台聊天能力已开放",
			"implementation": "independent",
		},
		Requests.AiPlatformCloudIntranet: gin.H{
			"enabled":        false,
			"question_types": []int{},
			"status":         "reserved",
			"status_message": "内网云平台入口保留，聊天能力尚未开放",
			"implementation": "independent",
		},
		Requests.AiPlatformImageCompare: gin.H{
			"enabled":        false,
			"question_types": []int{},
			"status":         "reserved",
			"status_message": "图片对比平台仍为占位能力",
			"implementation": "independent",
		},
	}
}

// ListConversations 返回 ai_robot 会话列表，便于排障和后台管理。
func (c *AiRobotController) ListConversations(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
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

	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
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

	pagination := buildPagination(req.Page, req.Limit, total)
	c.frontSuccess(ctx, "操作成功", gin.H{
		"items":      c.buildConversationListItems(ctx, cfg, items),
		"pagination": pagination,
		"notice":     buildPartialDataNotice(pagination),
	})
}

// GetConversation 返回单条 ai_robot 会话，用于排障和后续复用。
func (c *AiRobotController) GetConversation(ctx *gin.Context) {
	req := readConversationManageRequest(ctx)
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
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
	messages, messageTotal := c.listConversationMessages(ctx, cfg, req.UserID, req.Platform, req.ConversationID)
	detail := buildConversationDetail(doc, req.MessageID, messages)
	if messageTotal > int64(len(messages)) && len(messages) > 0 {
		detail["messages_notice"] = "当前会话消息较多，仅返回最近 " + strconv.Itoa(len(messages)) + " 条；可使用 /messages 接口分页获取完整历史"
		detail["messages_truncated"] = true
		detail["messages_total"] = messageTotal
	} else {
		detail["messages_truncated"] = false
		if messageTotal > 0 {
			detail["messages_total"] = messageTotal
		}
	}
	c.frontSuccess(ctx, "操作成功", detail)
}

// GetMessage 返回单条 ai_robot 消息详情，便于前端按 message_id 精确读取。
func (c *AiRobotController) GetMessage(ctx *gin.Context) {
	req := readConversationManageRequest(ctx)
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}
	if strings.TrimSpace(req.MessageID) == "" {
		c.frontFailed(ctx, "请求参数错误", errors.New("message_id 不能为空"))
		return
	}

	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
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

	message, err := c.findSingleMessage(ctx, cfg, req.UserID, req.Platform, req.ConversationID, req.MessageID)
	if err != nil {
		c.frontFailed(ctx, "读取消息失败", err)
		return
	}
	if message == nil {
		c.frontFailed(ctx, "消息不存在", mongo.ErrNoDocuments)
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"conversation_id": doc.ConversationID,
		"platform":        doc.Platform,
		"user_id":         doc.UserID,
		"message":         message,
		"updated_at":      formatAiRobotTime(doc.UpdatedAt),
	})
}

// ListMessages 返回单个会话的完整消息列表，支持分页。
func (c *AiRobotController) ListMessages(ctx *gin.Context) {
	page, limit := queryPageLimit(ctx)
	req := Requests.AiRobotConversationMessageListRequest{
		UserID:         ctx.Query("user_id"),
		Platform:       ctx.Query("platform"),
		ConversationID: ctx.Query("conversation_id"),
		Page:           page,
		Limit:          limit,
	}
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}
	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
		return
	}
	items, total, err := c.conversationMessageService.List(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		Services.AiRobotConversationMessageListFilter{
			UserID:         req.UserID,
			Platform:       req.Platform,
			ConversationID: req.ConversationID,
			Page:           int64(req.Page),
			Limit:          int64(req.Limit),
		},
	)
	if err != nil {
		c.frontFailed(ctx, "读取消息列表失败", err)
		return
	}
	pagination := buildPagination(req.Page, req.Limit, total)
	c.frontSuccess(ctx, "操作成功", gin.H{
		"items":      buildConversationMessageItems(items),
		"pagination": pagination,
		"notice":     buildPartialDataNotice(pagination),
	})
}

// MessageStats 返回消息审计统计，便于后台快速查看 ai_robot 使用效果。
func (c *AiRobotController) MessageStats(ctx *gin.Context) {
	req := Requests.AiRobotConversationMessageStatsRequest{
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
	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
		return
	}
	stats, err := c.conversationMessageService.Stats(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		Services.AiRobotConversationMessageListFilter{
			UserID:       req.UserID,
			Platform:     req.Platform,
			QuestionType: req.QuestionType,
			StartAt:      mustParseConversationListTime(req.StartTime),
			EndAt:        mustParseConversationListTime(req.EndTime),
		},
	)
	if err != nil {
		c.frontFailed(ctx, "读取消息统计失败", err)
		return
	}
	c.frontSuccess(ctx, "操作成功", gin.H{
		"stats": stats,
		"filters": gin.H{
			"user_id":       req.UserID,
			"platform":      req.Platform,
			"question_type": req.QuestionType,
			"start_time":    req.StartTime,
			"end_time":      req.EndTime,
		},
	})
}

// DeleteConversation 删除单条 ai_robot 会话。
func (c *AiRobotController) DeleteConversation(ctx *gin.Context) {
	req := readConversationManageRequest(ctx)
	if err := req.Validate(); err != nil {
		c.frontFailed(ctx, "请求参数错误", err)
		return
	}

	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
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
	deletedMessages, messageErr := c.conversationMessageService.DeleteByConversation(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		req.UserID,
		req.Platform,
		req.ConversationID,
	)
	if messageErr != nil {
		c.respondMessageCleanupPartialSuccess(ctx, "会话删除成功，但消息清理失败", messageErr, gin.H{
			"platform":        strings.TrimSpace(req.Platform),
			"conversation_id": strings.TrimSpace(req.ConversationID),
		})
		return
	}
	c.frontSuccess(ctx, "操作成功", buildMessageCleanupSuccessData(gin.H{
		"platform":        strings.TrimSpace(req.Platform),
		"conversation_id": strings.TrimSpace(req.ConversationID),
	}, deletedMessages))
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

	cfg := c.getAiGatewayConfigOrFail(ctx)
	if cfg == nil {
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
	deletedMessages, messageErr := c.conversationMessageService.DeleteByFilter(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		Services.AiRobotConversationMessageListFilter{
			UserID:       req.UserID,
			Platform:     req.Platform,
			QuestionType: req.QuestionType,
			StartAt:      mustParseConversationListTime(req.StartTime),
			EndAt:        mustParseConversationListTime(req.EndTime),
		},
	)
	if messageErr != nil {
		c.respondMessageCleanupPartialSuccess(ctx, "会话批量删除成功，但消息清理失败", messageErr, gin.H{
			"deleted_count": deletedCount,
			"filters": gin.H{
				"platform":      req.Platform,
				"question_type": req.QuestionType,
				"start_time":    req.StartTime,
				"end_time":      req.EndTime,
			},
		})
		return
	}
	c.frontSuccess(ctx, "操作成功", buildMessageCleanupSuccessData(gin.H{
		"deleted_count": deletedCount,
		"filters": gin.H{
			"platform":      req.Platform,
			"question_type": req.QuestionType,
			"start_time":    req.StartTime,
			"end_time":      req.EndTime,
		},
	}, deletedMessages))
}

func effectiveConversationDatabase(cfg *Config.AiGatewayConfig) string {
	if cfg == nil {
		return ""
	}
	effective := cfg.EffectiveConversationMongoConfig(Config.GetMongoDBConfig())
	return effective.Database
}

func (c *AiRobotController) buildConversationListItems(ctx *gin.Context, cfg *Config.AiGatewayConfig, items []Models.AiRobotConversation) []aiRobotConversationListItem {
	out := make([]aiRobotConversationListItem, 0, len(items))
	for _, item := range items {
		lastTaskUUID := ""
		if len(item.LastTaskUUIDs) > 0 {
			lastTaskUUID = item.LastTaskUUIDs[0]
		}
		latestMessage := c.latestConversationMessageForList(ctx, cfg, &item)
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
			LatestMessage:      latestMessage,
			UpdatedAt:          formatAiRobotTime(item.UpdatedAt),
		})
	}
	return out
}

func (c *AiRobotController) latestConversationMessageForList(ctx *gin.Context, cfg *Config.AiGatewayConfig, doc *Models.AiRobotConversation) *aiRobotConversationMessageItem {
	if c != nil && c.conversationMessageService != nil && ctx != nil && cfg != nil && doc != nil {
		message, err := c.conversationMessageService.FindLatest(
			ctx.Request.Context(),
			effectiveConversationDatabase(cfg),
			"",
			Services.AiRobotConversationMessageListFilter{
				UserID:         doc.UserID,
				Platform:       doc.Platform,
				ConversationID: doc.ConversationID,
			},
		)
		if err == nil && message != nil {
			items := buildConversationMessageItems([]Models.AiRobotConversationMessage{*message})
			if len(items) > 0 {
				return &items[0]
			}
		}
	}
	if doc == nil {
		return nil
	}
	return latestConversationMessage(doc.RecentTurns)
}

func buildConversationDetail(doc *Models.AiRobotConversation, messageID string, messages []aiRobotConversationMessageItem) gin.H {
	if doc == nil {
		return gin.H{
			"messages":        []aiRobotConversationMessageItem{},
			"current_message": nil,
		}
	}
	messageID = strings.TrimSpace(messageID)
	if len(messages) == 0 {
		messages = buildConversationMessageItemsFromTurns(doc.RecentTurns)
	}
	var currentMessage interface{}
	for _, item := range messages {
		if messageID != "" && strings.TrimSpace(item.MessageID) == messageID {
			currentMessage = item
		}
	}
	if currentMessage == nil && len(messages) > 0 {
		currentMessage = messages[len(messages)-1]
	}
	return gin.H{
		"id":                     doc.ID,
		"user_id":                doc.UserID,
		"conversation_id":        doc.ConversationID,
		"platform":               doc.Platform,
		"question_type":          doc.QuestionType,
		"last_message_id":        doc.LastMessageID,
		"last_intent":            doc.LastIntent,
		"last_question":          doc.LastQuestion,
		"last_resolved_question": doc.LastResolved,
		"last_answer":            doc.LastAnswer,
		"context_summary":        doc.ContextSummary,
		"last_project_ids":       doc.LastProjectIDs,
		"last_project_number":    doc.LastProjectNumber,
		"last_contract_ids":      doc.LastContractIDs,
		"last_contract_number":   doc.LastContractNumber,
		"last_task_uuids":        doc.LastTaskUUIDs,
		"last_article_name_cn":   doc.LastArticleNameCN,
		"last_article_name_en":   doc.LastArticleNameEN,
		"last_article_labels":    doc.LastArticleLabels,
		"last_journal_name":      doc.LastJournalName,
		"last_task_status":       doc.LastTaskStatus,
		"last_failure_reason":    doc.LastFailureReason,
		"last_download_kind":     doc.LastDownloadKind,
		"last_has_link":          doc.LastHasLink,
		"recent_turns":           buildConversationTurnViews(doc.RecentTurns),
		"messages":               messages,
		"current_message":        currentMessage,
		"updated_at":             formatAiRobotTime(doc.UpdatedAt),
	}
}

func findConversationMessage(doc *Models.AiRobotConversation, messageID string) *aiRobotConversationMessageItem {
	if doc == nil {
		return nil
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return nil
	}
	for _, turn := range doc.RecentTurns {
		if strings.TrimSpace(turn.MessageID) != messageID {
			continue
		}
		item := aiRobotConversationMessageItem{
			MessageID:           turn.MessageID,
			Question:            turn.Question,
			Resolved:            turn.Resolved,
			NormalizedQuestion:  firstNonEmptyText(turn.Resolved, turn.Question),
			HitContext:          strings.TrimSpace(turn.Resolved) != "" && strings.TrimSpace(turn.Resolved) != strings.TrimSpace(turn.Question),
			ContextSource:       fallbackContextSource(turn.Question, turn.Resolved),
			ClarificationNeeded: false,
			Answer:              turn.Answer,
			Intent:              turn.Intent,
			CreatedAt:           formatAiRobotTime(turn.CreatedAt),
		}
		return &item
	}
	return nil
}

func latestConversationMessage(turns []Models.AiRobotConversationTurn) *aiRobotConversationMessageItem {
	if len(turns) == 0 {
		return nil
	}
	last := turns[len(turns)-1]
	return &aiRobotConversationMessageItem{
		MessageID:           last.MessageID,
		Question:            last.Question,
		Resolved:            last.Resolved,
		NormalizedQuestion:  firstNonEmptyText(last.Resolved, last.Question),
		HitContext:          strings.TrimSpace(last.Resolved) != "" && strings.TrimSpace(last.Resolved) != strings.TrimSpace(last.Question),
		ContextSource:       fallbackContextSource(last.Question, last.Resolved),
		ClarificationNeeded: false,
		Answer:              last.Answer,
		Intent:              last.Intent,
		CreatedAt:           formatAiRobotTime(last.CreatedAt),
	}
}

func buildConversationMessageItems(items []Models.AiRobotConversationMessage) []aiRobotConversationMessageItem {
	out := make([]aiRobotConversationMessageItem, 0, len(items))
	for _, item := range items {
		out = append(out, aiRobotConversationMessageItem{
			MessageID:             item.MessageID,
			Question:              item.Question,
			Resolved:              item.Resolved,
			NormalizedQuestion:    firstNonEmptyText(item.NormalizedQuestion, item.Resolved, item.Question),
			HitContext:            item.HitContext,
			ContextSource:         item.ContextSource,
			ClarificationNeeded:   item.ClarificationNeeded,
			ClarificationReason:   item.ClarificationReason,
			Answer:                item.Answer,
			Intent:                item.Intent,
			Success:               item.Success,
			ShowMsg:               item.ShowMsg,
			DebugMsg:              item.DebugMsg,
			Stage:                 item.Stage,
			ErrorType:             item.ErrorType,
			CloudAPI:              item.CloudAPI,
			CloudStatusCode:       item.CloudStatusCode,
			LLMModel:              item.LLMModel,
			RequestPayloadSummary: item.RequestPayloadSummary,
			ResultKind:            item.ResultKind,
			ResultCount:           item.ResultCount,
			ResultBrief:           item.ResultBrief,
			ResponseSummary:       item.ResponseSummary,
			ResponseSize:          item.ResponseSize,
			CloudCalled:           item.CloudCalled,
			DurationMS:            item.DurationMS,
			LLMDurationMS:         item.LLMDurationMS,
			CloudDurationMS:       item.CloudDurationMS,
			CreatedAt:             formatAiRobotTime(item.CreatedAt),
		})
	}
	return out
}

func buildConversationMessageItemsFromTurns(items []Models.AiRobotConversationTurn) []aiRobotConversationMessageItem {
	out := make([]aiRobotConversationMessageItem, 0, len(items))
	for _, item := range items {
		out = append(out, aiRobotConversationMessageItem{
			MessageID:           item.MessageID,
			Question:            item.Question,
			Resolved:            item.Resolved,
			NormalizedQuestion:  firstNonEmptyText(item.Resolved, item.Question),
			HitContext:          strings.TrimSpace(item.Resolved) != "" && strings.TrimSpace(item.Resolved) != strings.TrimSpace(item.Question),
			ContextSource:       fallbackContextSource(item.Question, item.Resolved),
			ClarificationNeeded: false,
			Answer:              item.Answer,
			Intent:              item.Intent,
			CreatedAt:           formatAiRobotTime(item.CreatedAt),
		})
	}
	return out
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func fallbackContextSource(question string, resolved string) string {
	question = strings.TrimSpace(question)
	resolved = strings.TrimSpace(resolved)
	if resolved != "" && resolved != question {
		return "recent_turn"
	}
	return "none"
}

func formatAiRobotTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.In(aiRobotShanghaiLocation).Format(time.RFC3339)
}

func loadAiRobotShanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil && loc != nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}

func buildConversationTurnViews(items []Models.AiRobotConversationTurn) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, gin.H{
			"message_id":        item.MessageID,
			"question_type":     item.QuestionType,
			"question":          item.Question,
			"resolved_question": item.Resolved,
			"answer":            item.Answer,
			"intent":            item.Intent,
			"project_number":    item.ProjectNumber,
			"contract_number":   item.ContractNumber,
			"task_uuid":         item.TaskUUID,
			"article_name_cn":   item.ArticleNameCN,
			"journal_name":      item.JournalName,
			"created_at":        formatAiRobotTime(item.CreatedAt),
		})
	}
	return out
}

func (c *AiRobotController) listConversationMessages(ctx *gin.Context, cfg *Config.AiGatewayConfig, userID, platform, conversationID string) ([]aiRobotConversationMessageItem, int64) {
	if c == nil || c.conversationMessageService == nil || cfg == nil {
		return nil, 0
	}
	items, total, err := c.conversationMessageService.List(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		Services.AiRobotConversationMessageListFilter{
			UserID:         userID,
			Platform:       platform,
			ConversationID: conversationID,
			Page:           1,
			Limit:          200,
		},
	)
	if err != nil {
		return nil, 0
	}
	return buildConversationMessageItems(items), total
}

func (c *AiRobotController) findSingleMessage(ctx *gin.Context, cfg *Config.AiGatewayConfig, userID, platform, conversationID, messageID string) (*aiRobotConversationMessageItem, error) {
	if c == nil || c.conversationMessageService == nil || cfg == nil {
		return nil, nil
	}
	doc, err := c.conversationMessageService.FindOne(
		ctx.Request.Context(),
		effectiveConversationDatabase(cfg),
		"",
		Services.AiRobotConversationMessageListFilter{
			UserID:         userID,
			Platform:       platform,
			ConversationID: conversationID,
			MessageID:      messageID,
		},
	)
	if err == nil && doc != nil {
		items := buildConversationMessageItems([]Models.AiRobotConversationMessage{*doc})
		if len(items) > 0 {
			return &items[0], nil
		}
		return nil, nil
	}
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	return nil, nil
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
