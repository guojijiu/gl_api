package flow

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/backends"
	"cloud-platform-api/app/Services/ai_robot/internal/conversation"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/internal/timeutil"
	intranetclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/client"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"

	"github.com/gin-gonic/gin"
)

type Processor struct{}

type timedLLMClient struct {
	inner llm.ChatCompletionClient
	deps  *deps.Deps
}

// Processor is the ai_robot orchestration entry.
func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) respondFrontFormat(ctx *gin.Context, code int, showMsg string, debugMsg string, content interface{}) {
	if content == nil {
		content = gin.H{}
	}
	captureAiRobotResponseMeta(ctx, code, showMsg, debugMsg, content)
	content = enrichAiRobotFlowContent(ctx, content)
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
	captureAiRobotAnswer(ctx, data)
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

// FrontFailed and FrontSuccess implement deps.Responder.
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
	store := conversation.DefaultStore()
	state, err := conversation.LoadState(actx, store, req)
	if err != nil {
		p.frontFailed(ginCtx, "上下文初始化失败", err)
		return
	}
	req.ResolvedQuestion = conversation.EnhanceQuestion(req, state)

	d := &deps.Deps{
		Ctx:               actx,
		Gin:               ginCtx,
		Req:               req,
		Token:             token,
		Cfg:               cfg,
		LLM:               nil,
		Backend:           b,
		ConversationStore: store,
		Conversation:      state,
	}
	d.LLM = &timedLLMClient{inner: llmClient, deps: d}
	if d.Conversation != nil {
		d.Conversation.UserID = d.CurrentUserID()
		d.Conversation.LastAnswer = strings.TrimSpace(ginCtx.GetString("ai_robot_answer_text"))
	}
	p.dispatchByRegistry(d)
	if d.Conversation != nil {
		d.Conversation.LastAnswer = strings.TrimSpace(ginCtx.GetString("ai_robot_answer_text"))
	}
	if err = conversation.SaveState(actx, store, req, d.Conversation); err != nil {
		log.Printf(
			"ai_robot: save conversation state failed, platform=%s user_id=%s conversation_id=%s message_id=%s err=%v",
			strings.TrimSpace(req.Platform),
			strings.TrimSpace(req.UserID),
			strings.TrimSpace(req.ConversationID),
			strings.TrimSpace(req.MessageID),
			err,
		)
	}
	if err = persistConversationMessage(actx, cfg, req, d); err != nil {
		log.Printf(
			"ai_robot: persist conversation message failed, platform=%s user_id=%s conversation_id=%s message_id=%s err=%v",
			strings.TrimSpace(req.Platform),
			strings.TrimSpace(req.UserID),
			strings.TrimSpace(req.ConversationID),
			strings.TrimSpace(req.MessageID),
			err,
		)
	}
}

func enrichAiRobotFlowContent(ctx *gin.Context, content interface{}) gin.H {
	base := gin.H{}
	if existing, ok := content.(gin.H); ok {
		for k, v := range existing {
			base[k] = v
		}
	} else {
		base["data"] = content
	}
	if ctx != nil {
		if conversationID := ctx.GetString("ai_robot_conversation_id"); conversationID != "" {
			base["conversation_id"] = conversationID
		}
		if messageID := ctx.GetString("ai_robot_message_id"); messageID != "" {
			base["message_id"] = messageID
		}
		// 当结果较多时，给出统一简洁提示，避免客户误以为“总共就这些数据”。
		if _, exists := base["notice"]; !exists {
			if fallback := buildChatNoticeFromContext(ctx); fallback != "" {
				base["notice"] = fallback
			}
		}
	}
	return base
}

func buildChatNoticeFromContext(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	code := ctx.GetInt("ai_robot_response_code")
	if code != 1 {
		return ""
	}
	kind := strings.TrimSpace(ctx.GetString("ai_robot_result_kind"))
	count := ctx.GetInt("ai_robot_result_count")
	if (kind == "list" || kind == "download" || kind == "detail") && count > 5 {
		return "当前默认仅展示前 5 条（共 " + intToString(count) + " 条）。如需查看更多，可直接回复“查看更多”或“下一页”，也可在会话消息列表页面分页查看完整数据。"
	}
	return ""
}

func captureAiRobotAnswer(ctx *gin.Context, data interface{}) {
	if ctx == nil {
		return
	}
	answer := extractAiRobotAnswerText(data)
	if strings.TrimSpace(answer) != "" {
		ctx.Set("ai_robot_answer_text", strings.TrimSpace(answer))
	}
}

func captureAiRobotResponseMeta(ctx *gin.Context, code int, showMsg string, debugMsg string, content interface{}) {
	if ctx == nil {
		return
	}
	ctx.Set("ai_robot_response_code", code)
	ctx.Set("ai_robot_show_msg", strings.TrimSpace(showMsg))
	ctx.Set("ai_robot_debug_msg", strings.TrimSpace(debugMsg))
	resultMeta := analyzeResultMeta(code, content)
	ctx.Set("ai_robot_result_kind", resultMeta.Kind)
	ctx.Set("ai_robot_result_count", resultMeta.Count)
}

func extractAiRobotAnswerText(data interface{}) string {
	switch v := data.(type) {
	case gin.H:
		if answer := stringifyAiRobotAnswerValue(v["answer"]); answer != "" {
			return answer
		}
		if nested, ok := v["data"].(gin.H); ok {
			if answer := stringifyAiRobotAnswerValue(nested["answer"]); answer != "" {
				return answer
			}
		}
	case map[string]interface{}:
		if answer := stringifyAiRobotAnswerValue(v["answer"]); answer != "" {
			return answer
		}
		if nested, ok := v["data"].(map[string]interface{}); ok {
			if answer := stringifyAiRobotAnswerValue(nested["answer"]); answer != "" {
				return answer
			}
		}
	case string:
		return strings.TrimSpace(v)
	}
	return ""
}

func stringifyAiRobotAnswerValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(raw))
	}
}

type aiRobotResultMeta struct {
	Kind  string
	Count int
}

func analyzeResultMeta(code int, content interface{}) aiRobotResultMeta {
	if code != 1 {
		return aiRobotResultMeta{Kind: "error", Count: 0}
	}
	root := normalizeContentMap(content)
	if len(root) == 0 {
		return aiRobotResultMeta{Kind: "empty", Count: 0}
	}
	if data, ok := nestedMap(root, "data"); ok && len(data) > 0 {
		root = data
	}
	if meta, ok := metaByRootSlice(root, "links", "download"); ok {
		return meta
	}
	if meta, ok := metaByRootSlice(root, "details", "detail"); ok {
		return meta
	}
	if meta, ok := metaByRootSlice(root, "contract_projects", "list"); ok {
		return meta
	}
	if contentRoot := normalizeContentMap(root["content"]); len(contentRoot) > 0 {
		if items, ok := toSlice(contentRoot["data"]); ok {
			return metaByContentItems(items)
		}
	}
	if answer := strings.TrimSpace(stringifyAiRobotAnswerValue(root["answer"])); answer != "" {
		return aiRobotResultMeta{Kind: "detail", Count: 1}
	}
	if rawCloudJSON, exists := root["raw_cloud_json"]; exists && rawCloudJSON != nil {
		return aiRobotResultMeta{Kind: "detail", Count: 1}
	}
	return aiRobotResultMeta{Kind: "empty", Count: 0}
}

func metaByRootSlice(root map[string]interface{}, key string, kind string) (aiRobotResultMeta, bool) {
	if len(root) == 0 {
		return aiRobotResultMeta{}, false
	}
	items, ok := toSlice(root[key])
	if !ok {
		return aiRobotResultMeta{}, false
	}
	return metaByFixedKind(items, kind), true
}

func metaByFixedKind(items []interface{}, kind string) aiRobotResultMeta {
	if len(items) == 0 {
		return aiRobotResultMeta{Kind: "empty", Count: 0}
	}
	return aiRobotResultMeta{Kind: kind, Count: len(items)}
}

func metaByContentItems(items []interface{}) aiRobotResultMeta {
	switch n := len(items); {
	case n == 0:
		return aiRobotResultMeta{Kind: "empty", Count: 0}
	case n == 1:
		return aiRobotResultMeta{Kind: "detail", Count: 1}
	default:
		return aiRobotResultMeta{Kind: "list", Count: n}
	}
}

func normalizeContentMap(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case gin.H:
		return map[string]interface{}(v)
	case map[string]interface{}:
		return v
	default:
		return nil
	}
}

func nestedMap(root map[string]interface{}, key string) (map[string]interface{}, bool) {
	if len(root) == 0 {
		return nil, false
	}
	value, ok := root[key]
	if !ok {
		return nil, false
	}
	m := normalizeContentMap(value)
	return m, len(m) > 0
}

func toSlice(value interface{}) ([]interface{}, bool) {
	items, ok := value.([]interface{})
	return items, ok
}

func persistConversationMessage(actx context.Context, cfg *Config.AiGatewayConfig, req *Requests.AiRobotChatRequest, d *deps.Deps) error {
	if cfg == nil || req == nil || strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.ConversationID) == "" || strings.TrimSpace(req.MessageID) == "" {
		return nil
	}
	service := Services.NewAiRobotConversationMessageService()
	now := timeutil.NowInShanghai()
	if err := service.EnsureIndexes(actx, effectiveMessageDatabase(cfg), "", cfg.ConversationTTLHours); err != nil {
		return err
	}
	doc := &Models.AiRobotConversationMessage{
		ID:             service.BuildDocumentID(req.UserID, req.Platform, req.ConversationID, req.MessageID),
		UserID:         strings.TrimSpace(req.UserID),
		ConversationID: strings.TrimSpace(req.ConversationID),
		Platform:       strings.TrimSpace(req.Platform),
		QuestionType:   req.QuestionType,
		MessageID:      strings.TrimSpace(req.MessageID),
		Question:       strings.TrimSpace(req.Question),
		Resolved:       strings.TrimSpace(req.ResolvedQuestion),
		Answer:         strings.TrimSpace(ginAnswerText(d)),
		Success:        ginResponseCode(d) == 1,
		ShowMsg:        ginShowMsg(d),
		DebugMsg:       ginDebugMsg(d),
		ResultKind:     ginResultKind(d),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if d != nil && d.Conversation != nil {
		doc.Intent = strings.TrimSpace(d.Conversation.LastIntent)
	}
	return service.Upsert(actx, effectiveMessageDatabase(cfg), "", doc)
}

func ginAnswerText(d *deps.Deps) string {
	return ginStringValue(d, "ai_robot_answer_text")
}

func ginResponseCode(d *deps.Deps) int {
	return ginIntValue(d, "ai_robot_response_code")
}

func ginShowMsg(d *deps.Deps) string {
	return strings.TrimSpace(ginStringValue(d, "ai_robot_show_msg"))
}

func ginDebugMsg(d *deps.Deps) string {
	return strings.TrimSpace(ginStringValue(d, "ai_robot_debug_msg"))
}

func ginResultKind(d *deps.Deps) string {
	return strings.TrimSpace(ginStringValue(d, "ai_robot_result_kind"))
}

func ginStringValue(d *deps.Deps, key string) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return d.Gin.GetString(key)
}

func ginIntValue(d *deps.Deps, key string) int {
	if d == nil || d.Gin == nil {
		return 0
	}
	return d.Gin.GetInt(key)
}

func intToString(value int) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(raw)
}

func effectiveMessageDatabase(cfg *Config.AiGatewayConfig) string {
	if cfg == nil {
		return ""
	}
	effective := cfg.EffectiveConversationMongoConfig(Config.GetMongoDBConfig())
	return effective.Database
}

func (c *timedLLMClient) ChatCompletion(ctx context.Context, userPrompt string) (string, error) {
	if c == nil || c.inner == nil {
		return "", nil
	}
	return c.inner.ChatCompletion(ctx, userPrompt)
}
