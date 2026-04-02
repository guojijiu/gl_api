package flow

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"

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
	if ctx != nil {
		ctx.Set("ai_robot_stage", "completed")
		ctx.Set("ai_robot_error_type", "")
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
	if ctx != nil {
		if _, ok := ctx.Get("ai_robot_stage"); !ok {
			ctx.Set("ai_robot_stage", inferStageFromFailure(showMsg, debugMsg))
		}
		ctx.Set("ai_robot_error_type", inferErrorType(ctx, showMsg, debugMsg))
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
	startedAt := time.Now()
	if ginCtx != nil {
		ginCtx.Set("ai_robot_started_at", startedAt)
		ginCtx.Set("ai_robot_stage", "init")
		ginCtx.Set("ai_robot_question_type", req.QuestionType)
		ginCtx.Set("ai_robot_request_payload_summary", buildRequestPayloadSummary(req))
	}
	if cfg == nil {
		cfg = &Config.AiGatewayConfig{}
		cfg.SetDefaults()
	}
	hc, publicAPI, intranetAPI := p.initPlatformClients(cfg, req.Platform)
	llmClient := llm.NewClient(cfg, hc)
	if ginCtx != nil {
		ginCtx.Set("ai_robot_stage", "backend_init")
	}
	b, err := backends.NewForPlatform(req.Platform, publicAPI, intranetAPI)
	if err != nil {
		p.frontFailed(ginCtx, "平台后端初始化失败", err)
		return
	}
	store := conversation.DefaultStore()
	if ginCtx != nil {
		ginCtx.Set("ai_robot_stage", "context_load")
	}
	state, err := conversation.LoadState(actx, store, req)
	if err != nil {
		p.frontFailed(ginCtx, "上下文初始化失败", err)
		return
	}
	req.ResolvedQuestion = conversation.EnhanceQuestion(req, state)
	if ginCtx != nil {
		hitContext, contextSource := detectContextUsage(req, state)
		clarificationNeeded, clarificationReason := detectClarificationNeed(req, state)
		ginCtx.Set("ai_robot_hit_context", hitContext)
		ginCtx.Set("ai_robot_context_source", contextSource)
		ginCtx.Set("ai_robot_clarification_needed", clarificationNeeded)
		ginCtx.Set("ai_robot_clarification_reason", clarificationReason)
	}
	if ginCtx != nil {
		ginCtx.Set("ai_robot_stage", "business_dispatch")
	}

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
	_ = conversation.SaveState(actx, store, req, d.Conversation)
	_ = persistConversationMessage(actx, cfg, req, d)
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
		if durationMS := aiRobotDurationMSFromContext(ctx); durationMS > 0 {
			base["duration_ms"] = durationMS
		}
		if llmDurationMS := aiRobotLLMDurationMSFromContext(ctx); llmDurationMS > 0 {
			base["llm_duration_ms"] = llmDurationMS
		}
		if cloudDurationMS := aiRobotCloudDurationMSFromContext(ctx); cloudDurationMS > 0 {
			base["cloud_duration_ms"] = cloudDurationMS
		}
		if cloudAPI := strings.TrimSpace(ctx.GetString("ai_robot_cloud_api")); cloudAPI != "" {
			base["cloud_api"] = cloudAPI
		}
		if cloudStatusCode := ctx.GetInt("ai_robot_cloud_status_code"); cloudStatusCode > 0 {
			base["cloud_status_code"] = cloudStatusCode
		}
		if llmModel := strings.TrimSpace(ctx.GetString("ai_robot_llm_model")); llmModel != "" {
			base["llm_model"] = llmModel
		}
		if requestPayloadSummary := strings.TrimSpace(ctx.GetString("ai_robot_request_payload_summary")); requestPayloadSummary != "" {
			base["request_payload_summary"] = requestPayloadSummary
		}
		if responseSummary := strings.TrimSpace(ctx.GetString("ai_robot_response_summary")); responseSummary != "" {
			base["response_summary"] = responseSummary
		}
		if responseSize := ctx.GetInt64("ai_robot_response_size"); responseSize > 0 {
			base["response_size"] = responseSize
		}
		// 当结果较多/响应体较大时，回答通常只会展示摘要或部分示例；
		// 这里给出显式提示，避免客户误以为“总共就这些数据”。
		if _, exists := base["notice"]; !exists {
			if notice := buildChatNoticeFromContext(ctx); notice != "" {
				base["notice"] = notice
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
	size := ctx.GetInt64("ai_robot_response_size")

	// 经验阈值：summary 中通常只会采样少量条目，超过该阈值就提醒“还有更多”。
	if (kind == "list" || kind == "download") && count > 3 {
		return "结果较多，本次回答仅展示摘要/部分示例；如需更多数据可继续翻问补充筛选条件，或查看原始数据（raw_cloud_json）"
	}
	if kind == "detail" && count > 1 {
		return "结果较多，本次回答仅展示摘要/部分示例；如需更多数据可继续翻问补充筛选条件，或查看原始数据（raw_cloud_json）"
	}
	// 响应体很大时，即使统计不明显，也提醒可能存在信息被摘要压缩。
	if size >= 64*1024 {
		return "返回内容较多，本次回答为摘要/压缩展示；如需完整信息可查看原始数据（raw_cloud_json）或继续追问"
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
	ctx.Set("ai_robot_cloud_called", extractAiRobotCloudCalled(content))
	resultMeta := analyzeResultMeta(code, content)
	ctx.Set("ai_robot_result_kind", resultMeta.Kind)
	ctx.Set("ai_robot_result_count", resultMeta.Count)
	ctx.Set("ai_robot_result_brief", buildResultBrief(resultMeta, showMsg, content))
	ctx.Set("ai_robot_response_summary", buildResponseSummary(ctx, content))
	ctx.Set("ai_robot_response_size", estimateResponseSize(content))
	if llmModel := strings.TrimSpace(ctx.GetString("ai_robot_llm_model")); llmModel == "" {
		if model := extractAiRobotLLMModel(content); model != "" {
			ctx.Set("ai_robot_llm_model", model)
		}
	}
	if code == 1 {
		ctx.Set("ai_robot_error_type", "")
	}
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

func extractAiRobotCloudCalled(data interface{}) bool {
	switch v := data.(type) {
	case gin.H:
		if called, ok := v["cloud_called"].(bool); ok {
			return called
		}
		if nested, ok := v["data"].(gin.H); ok {
			if called, ok := nested["cloud_called"].(bool); ok {
				return called
			}
		}
	case map[string]interface{}:
		if called, ok := v["cloud_called"].(bool); ok {
			return called
		}
		if nested, ok := v["data"].(map[string]interface{}); ok {
			if called, ok := nested["cloud_called"].(bool); ok {
				return called
			}
		}
	}
	return false
}

func extractAiRobotLLMModel(_ interface{}) string {
	return ""
}

func estimateResponseSize(content interface{}) int64 {
	if content == nil {
		return 0
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return 0
	}
	return int64(len(raw))
}

type aiRobotResultMeta struct {
	Kind  string
	Count int
}

func buildResultBrief(meta aiRobotResultMeta, showMsg string, content interface{}) string {
	switch meta.Kind {
	case "error":
		if msg := strings.TrimSpace(showMsg); msg != "" {
			return truncateAuditText(msg, 80)
		}
		return "接口处理失败"
	case "empty":
		return "未查询到相关数据"
	case "download":
		if meta.Count > 0 {
			return "下载链接 " + intToString(meta.Count) + " 个"
		}
		return "下载结果"
	case "list":
		if meta.Count > 0 {
			return "列表结果 " + intToString(meta.Count) + " 条"
		}
		return "列表结果"
	case "detail":
		if answer := extractAiRobotAnswerText(content); strings.TrimSpace(answer) != "" {
			return truncateAuditText(answer, 80)
		}
		if meta.Count > 1 {
			return "详情结果 " + intToString(meta.Count) + " 条"
		}
		return "详情结果"
	default:
		return ""
	}
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
	if items, ok := toSlice(root["links"]); ok {
		if len(items) == 0 {
			return aiRobotResultMeta{Kind: "empty", Count: 0}
		}
		return aiRobotResultMeta{Kind: "download", Count: len(items)}
	}
	if items, ok := toSlice(root["details"]); ok {
		if len(items) == 0 {
			return aiRobotResultMeta{Kind: "empty", Count: 0}
		}
		return aiRobotResultMeta{Kind: "detail", Count: len(items)}
	}
	if items, ok := toSlice(root["contract_projects"]); ok {
		if len(items) == 0 {
			return aiRobotResultMeta{Kind: "empty", Count: 0}
		}
		return aiRobotResultMeta{Kind: "list", Count: len(items)}
	}
	if contentRoot := normalizeContentMap(root["content"]); len(contentRoot) > 0 {
		if items, ok := toSlice(contentRoot["data"]); ok {
			if len(items) == 0 {
				return aiRobotResultMeta{Kind: "empty", Count: 0}
			}
			if len(items) == 1 {
				return aiRobotResultMeta{Kind: "detail", Count: 1}
			}
			return aiRobotResultMeta{Kind: "list", Count: len(items)}
		}
	}
	if answer := strings.TrimSpace(stringifyAiRobotAnswerValue(root["answer"])); answer != "" {
		return aiRobotResultMeta{Kind: "detail", Count: 1}
	}
	if rawCloudJSON, exists := root["raw_cloud_json"]; exists && estimateResponseSize(rawCloudJSON) > 0 {
		return aiRobotResultMeta{Kind: "detail", Count: 1}
	}
	return aiRobotResultMeta{Kind: "empty", Count: 0}
}

func buildResponseSummary(ctx *gin.Context, content interface{}) string {
	if content == nil {
		return ""
	}
	summary := map[string]interface{}{}
	switch v := content.(type) {
	case gin.H:
		fillResponseSummaryMap(ctx, summary, v)
	case map[string]interface{}:
		fillResponseSummaryMap(ctx, summary, v)
	default:
		summary["type"] = truncateAuditText(typeNameOf(content), 60)
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		return ""
	}
	return string(raw)
}

func fillResponseSummaryMap(ctx *gin.Context, summary map[string]interface{}, data map[string]interface{}) {
	if summary == nil || data == nil {
		return
	}
	if answer, ok := data["answer"]; ok {
		summary["answer"] = truncateAuditText(stringifyAiRobotAnswerValue(answer), 160)
	}
	if intent, ok := data["intent"].(string); ok {
		summary["intent"] = truncateAuditText(intent, 80)
	}
	if uuids, ok := data["uuids"]; ok {
		summary["uuids"] = truncateCollectionPreview(uuids, 5)
	}
	if links, ok := data["links"]; ok {
		summary["links"] = truncateCollectionPreview(links, 3)
	}
	if cloudCalled, ok := data["cloud_called"].(bool); ok {
		summary["cloud_called"] = cloudCalled
	}
	if rawCloudJSON, ok := data["raw_cloud_json"]; ok {
		summary["raw_cloud_json"] = map[string]interface{}{
			"type": typeNameOf(rawCloudJSON),
			"size": estimateResponseSize(rawCloudJSON),
		}
		if domainSummary := summarizeDomainResponse(ctx, rawCloudJSON); len(domainSummary) > 0 {
			summary["domain_summary"] = domainSummary
		}
	}
	if nested, ok := data["data"].(map[string]interface{}); ok {
		fillResponseSummaryMap(ctx, summary, nested)
	}
}

func summarizeDomainResponse(ctx *gin.Context, raw interface{}) map[string]interface{} {
	rawJSON := marshalSummaryJSON(raw)
	if len(rawJSON) == 0 {
		return nil
	}
	var root map[string]interface{}
	if err := json.Unmarshal(rawJSON, &root); err != nil {
		return nil
	}
	questionType := 0
	if ctx != nil {
		questionType = ctx.GetInt("ai_robot_question_type")
	}
	switch questionType {
	case Requests.AiQuestionTypeProject:
		return summarizeProjectResponse(root)
	case Requests.AiQuestionTypeTask:
		return summarizeTaskResponse(root)
	case Requests.AiQuestionTypeProjectArticle:
		return summarizeProjectArticleResponse(root)
	default:
		return summarizeCommonListResponse(root)
	}
}

func summarizeProjectResponse(root map[string]interface{}) map[string]interface{} {
	if len(root) == 0 {
		return nil
	}
	if links, ok := toSlice(root["links"]); ok {
		return map[string]interface{}{
			"domain":       "project_download",
			"link_count":   len(links),
			"links_sample": sampleMaps(links, 3, []string{"name", "url", "type", "file_name"}),
		}
	}
	if cps, ok := toSlice(root["contract_projects"]); ok {
		return map[string]interface{}{
			"domain":                   "project_contract_projects",
			"contract_project_count":   len(cps),
			"contract_projects_sample": sampleMaps(cps, 3, []string{"id", "number", "name", "workflow_name_cn", "effective_end_at", "contract_id"}),
		}
	}
	result := summarizeCommonListResponse(root)
	if len(result) == 0 {
		return nil
	}
	result["domain"] = "project"
	if items, ok := extractContentDataSlice(root); ok {
		result["items_sample"] = sampleMaps(items, 3, []string{"id", "number", "name", "workflow_name_cn", "effective_end_at", "contract_id"})
	}
	return result
}

func summarizeTaskResponse(root map[string]interface{}) map[string]interface{} {
	if len(root) == 0 {
		return nil
	}
	if details, ok := toSlice(root["details"]); ok {
		summary := map[string]interface{}{
			"domain":         "task_status",
			"detail_count":   len(details),
			"details_sample": sampleMaps(details, 3, []string{"uuid", "name", "status", "status_value", "workflow_name_cn", "project_number"}),
		}
		if uuids, ok := stringSlice(root["uuids"]); ok && len(uuids) > 0 {
			summary["uuids"] = truncateStringSlice(uuids, 5)
		}
		return summary
	}
	if links, ok := toSlice(root["links"]); ok {
		summary := map[string]interface{}{
			"domain":       "task_download",
			"link_count":   len(links),
			"links_sample": sampleMaps(links, 3, []string{"name", "url", "type", "file_name"}),
		}
		if uuids, ok := stringSlice(root["uuids"]); ok && len(uuids) > 0 {
			summary["uuids"] = truncateStringSlice(uuids, 5)
		}
		return summary
	}
	result := summarizeCommonListResponse(root)
	if len(result) == 0 {
		return nil
	}
	result["domain"] = "task"
	if items, ok := extractContentDataSlice(root); ok {
		result["items_sample"] = sampleMaps(items, 3, []string{"id", "uuid", "name", "status", "status_value", "workflow_name_cn", "project_id"})
	}
	return result
}

func summarizeProjectArticleResponse(root map[string]interface{}) map[string]interface{} {
	result := summarizeCommonListResponse(root)
	if len(result) == 0 {
		return nil
	}
	result["domain"] = "project_article"
	if items, ok := extractContentDataSlice(root); ok {
		result["items_sample"] = sampleMaps(items, 3, []string{"name_cn", "name_en", "journal_name", "publish_date", "product_category", "region"})
	}
	return result
}

func summarizeCommonListResponse(root map[string]interface{}) map[string]interface{} {
	if len(root) == 0 {
		return nil
	}
	summary := map[string]interface{}{}
	if code, ok := root["code"]; ok {
		summary["code"] = code
	}
	if showMsg, ok := root["showMsg"].(string); ok && strings.TrimSpace(showMsg) != "" {
		summary["show_msg"] = truncateAuditText(showMsg, 80)
	}
	if items, ok := extractContentDataSlice(root); ok {
		summary["total"] = firstNonZeroInt(toInt(root["total"]), toInt(nestedContentValue(root, "total")), len(items))
		summary["items_sample"] = sampleMaps(items, 3, nil)
		summary["omitted_count"] = maxInt(0, len(items)-3)
		return summary
	}
	return summary
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

func extractContentDataSlice(root map[string]interface{}) ([]interface{}, bool) {
	content, _ := root["content"].(map[string]interface{})
	if content == nil {
		return nil, false
	}
	items, ok := toSlice(content["data"])
	return items, ok
}

func nestedContentValue(root map[string]interface{}, key string) interface{} {
	content, _ := root["content"].(map[string]interface{})
	if content == nil {
		return nil
	}
	return content[key]
}

func sampleMaps(items []interface{}, maxItems int, fields []string) []map[string]interface{} {
	limit := minInt(len(items), maxItems)
	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		item, ok := items[i].(map[string]interface{})
		if !ok || item == nil {
			continue
		}
		row := map[string]interface{}{}
		if len(fields) == 0 {
			for key, value := range item {
				row[key] = compressSummaryValue(value)
				if len(row) >= 6 {
					break
				}
			}
		} else {
			for _, field := range fields {
				if value, exists := item[field]; exists {
					row[field] = compressSummaryValue(value)
				}
			}
		}
		if len(row) > 0 {
			result = append(result, row)
		}
	}
	return result
}

func compressSummaryValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return truncateAuditText(v, 120)
	case []interface{}:
		return truncateCollectionPreview(v, 3)
	case []string:
		return truncateStringSlice(v, 3)
	default:
		return value
	}
}

func marshalSummaryJSON(value interface{}) []byte {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		return v
	case string:
		return []byte(v)
	case json.RawMessage:
		return []byte(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return raw
	}
}

func toSlice(value interface{}) ([]interface{}, bool) {
	items, ok := value.([]interface{})
	return items, ok
}

func stringSlice(value interface{}) ([]string, bool) {
	switch v := value.(type) {
	case []string:
		return v, true
	case []interface{}:
		items := make([]string, 0, len(v))
		for _, item := range v {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				items = append(items, strings.TrimSpace(text))
			}
		}
		return items, len(items) > 0
	default:
		return nil, false
	}
}

func truncateStringSlice(items []string, maxItems int) []string {
	if len(items) <= maxItems {
		return items
	}
	return items[:maxItems]
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateCollectionPreview(value interface{}, maxItems int) interface{} {
	if maxItems <= 0 {
		maxItems = 3
	}
	switch v := value.(type) {
	case []string:
		if len(v) > maxItems {
			return v[:maxItems]
		}
		return v
	case []interface{}:
		if len(v) > maxItems {
			return v[:maxItems]
		}
		return v
	default:
		return value
	}
}

func typeNameOf(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return jsonTypeName(value)
}
func jsonTypeName(value interface{}) string {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 {
		return "unknown"
	}
	switch raw[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	default:
		return "scalar"
	}
}

func buildRequestPayloadSummary(req *Requests.AiRobotChatRequest) string {
	if req == nil {
		return ""
	}
	summary := map[string]interface{}{
		"question":        truncateAuditText(req.Question, 120),
		"platform":        strings.TrimSpace(req.Platform),
		"question_type":   req.QuestionType,
		"enable_context":  req.EnableContext,
		"user_id":         truncateAuditText(req.UserID, 40),
		"conversation_id": truncateAuditText(req.ConversationID, 60),
		"message_id":      truncateAuditText(req.MessageID, 60),
	}
	if resolved := strings.TrimSpace(req.ResolvedQuestion); resolved != "" {
		summary["resolved_question"] = truncateAuditText(resolved, 160)
	}
	switch req.QuestionType {
	case Requests.AiQuestionTypeProject:
		projectFilters := parse.ExtractProjectFiltersFromQuestion(req.Question)
		contractFilters := parse.ExtractContractFiltersFromQuestion(req.Question)
		summary["project_filters"] = map[string]interface{}{
			"number":           truncateAuditText(projectFilters.Number, 80),
			"name":             truncateAuditText(projectFilters.Name, 80),
			"workflow_name_cn": truncateAuditText(projectFilters.WorkflowNameCN, 80),
		}
		summary["contract_filters"] = map[string]interface{}{
			"contract_number": truncateAuditText(contractFilters.ContractNumber, 80),
			"name":            truncateAuditText(contractFilters.Name, 80),
		}
	case Requests.AiQuestionTypeTask:
		taskUUIDs := parse.ExtractTaskUUIDsFromQuestion(req.Question)
		taskFilters := parse.ExtractTaskFiltersFromQuestion(req.Question, firstTaskUUID(taskUUIDs))
		summary["task_uuids"] = taskUUIDs
		summary["task_filters"] = map[string]interface{}{
			"uuid":             truncateAuditText(taskFilters.UUID, 80),
			"name":             truncateAuditText(taskFilters.Name, 80),
			"tool_name":        truncateAuditText(taskFilters.ToolName, 80),
			"project_number":   truncateAuditText(taskFilters.ProjectNumber, 80),
			"project_name":     truncateAuditText(taskFilters.ProjectName, 80),
			"workflow_name_cn": truncateAuditText(taskFilters.WorkflowNameCN, 80),
			"status_value":     truncateAuditText(taskFilters.StatusValue, 40),
			"created_at_start": truncateAuditText(taskFilters.CreatedAtStart, 40),
			"created_at_end":   truncateAuditText(taskFilters.CreatedAtEnd, 40),
		}
	case Requests.AiQuestionTypeProjectArticle:
		articleFilters := parse.ExtractProjectArticleFiltersFromQuestion(req.Question)
		items := make([]map[string]string, 0, len(articleFilters.SearchFilter))
		for _, item := range articleFilters.SearchFilter {
			items = append(items, map[string]string{
				"column":   truncateAuditText(item.Column, 40),
				"operator": truncateAuditText(item.Operator, 40),
				"value":    truncateAuditText(item.Value, 80),
			})
		}
		summary["article_filters"] = items
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		return ""
	}
	return string(raw)
}

func firstTaskUUID(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[0])
}

func truncateAuditText(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if maxLen <= 0 {
		return value
	}
	rs := []rune(value)
	if len(rs) <= maxLen {
		return value
	}
	return string(rs[:maxLen]) + "..."
}

func normalizedQuestion(req *Requests.AiRobotChatRequest) string {
	if req == nil {
		return ""
	}
	return firstNonEmptyQuestion(req.ResolvedQuestion, req.Question)
}

func firstNonEmptyQuestion(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func detectContextUsage(req *Requests.AiRobotChatRequest, state *conversation.State) (bool, string) {
	if req == nil {
		return false, "none"
	}
	question := strings.TrimSpace(req.Question)
	resolved := strings.TrimSpace(req.ResolvedQuestion)
	if resolved == "" || resolved == question {
		return false, "none"
	}
	return true, inferContextSource(question, state)
}

func inferContextSource(question string, state *conversation.State) string {
	question = strings.TrimSpace(strings.ToLower(question))
	if question == "" {
		return "none"
	}
	if isRecentTurnStyleFollowUp(question) {
		return "recent_turn"
	}
	if state != nil && strings.TrimSpace(state.ContextSummary) != "" {
		return "summary"
	}
	if state != nil && len(state.RecentTurns) > 0 {
		return "recent_turn"
	}
	return "none"
}

func isRecentTurnStyleFollowUp(question string) bool {
	return isShortFollowUpQuestion(question) || containsAnyText(question,
		"这个项目", "那个项目", "该项目", "上个项目",
		"这个任务", "那个任务", "该任务", "上个任务",
		"这个工单", "那个工单", "该工单",
		"这篇文章", "那个文章", "这篇文献", "那个文献",
		"结果呢", "状态呢", "链接呢", "下载呢", "报告呢", "原始数据呢",
		"为什么失败", "失败原因", "哪个链接", "还有吗", "继续查", "再查下", "详细点",
	)
}

func isShortFollowUpQuestion(question string) bool {
	return len([]rune(strings.TrimSpace(question))) > 0 && len([]rune(strings.TrimSpace(question))) <= 12
}

func containsAnyText(question string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(question, strings.ToLower(strings.TrimSpace(keyword))) {
			return true
		}
	}
	return false
}

func detectClarificationNeed(req *Requests.AiRobotChatRequest, state *conversation.State) (bool, string) {
	if req == nil {
		return false, ""
	}
	question := strings.TrimSpace(req.Question)
	resolved := strings.TrimSpace(req.ResolvedQuestion)
	if question == "" {
		return true, "empty_question"
	}
	if len([]rune(question)) <= 2 {
		return true, "question_too_short"
	}
	if isRecentTurnStyleFollowUp(strings.ToLower(question)) && !hasUsableConversationContext(state) && resolved == question {
		return true, "follow_up_without_context"
	}
	if requiresDomainAnchor(req.QuestionType, question) && resolved == question && !containsDomainAnchor(req.QuestionType, question) {
		return true, "missing_core_identifier"
	}
	return false, ""
}

func hasUsableConversationContext(state *conversation.State) bool {
	if state == nil {
		return false
	}
	if len(state.RecentTurns) > 0 || strings.TrimSpace(state.ContextSummary) != "" {
		return true
	}
	switch {
	case strings.TrimSpace(state.LastProjectNumber) != "":
		return true
	case strings.TrimSpace(state.LastContractNumber) != "":
		return true
	case len(state.LastTaskUUIDs) > 0:
		return true
	case strings.TrimSpace(state.LastArticleNameCN) != "":
		return true
	case strings.TrimSpace(state.LastJournalName) != "":
		return true
	default:
		return false
	}
}

func requiresDomainAnchor(questionType int, question string) bool {
	question = strings.TrimSpace(strings.ToLower(question))
	switch questionType {
	case Requests.AiQuestionTypeProject:
		return containsAnyText(question, "下载", "报告", "原始数据", "质控", "合同", "项目")
	case Requests.AiQuestionTypeTask:
		return containsAnyText(question, "任务", "工单", "状态", "结果", "下载", "失败")
	case Requests.AiQuestionTypeProjectArticle:
		return containsAnyText(question, "文章", "文献", "期刊", "中文名称", "英文名称")
	default:
		return false
	}
}

func containsDomainAnchor(questionType int, question string) bool {
	question = strings.TrimSpace(strings.ToLower(question))
	switch questionType {
	case Requests.AiQuestionTypeProject:
		return containsAnyText(question, "项目编号", "合同编号")
	case Requests.AiQuestionTypeTask:
		return containsAnyText(question, "任务编号", "uuid")
	case Requests.AiQuestionTypeProjectArticle:
		return containsAnyText(question, "中文名称", "英文名称", "期刊名称")
	default:
		return false
	}
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
		ID:                    service.BuildDocumentID(req.UserID, req.Platform, req.ConversationID, req.MessageID),
		UserID:                strings.TrimSpace(req.UserID),
		ConversationID:        strings.TrimSpace(req.ConversationID),
		Platform:              strings.TrimSpace(req.Platform),
		QuestionType:          req.QuestionType,
		MessageID:             strings.TrimSpace(req.MessageID),
		Question:              strings.TrimSpace(req.Question),
		Resolved:              strings.TrimSpace(req.ResolvedQuestion),
		NormalizedQuestion:    normalizedQuestion(req),
		HitContext:            ginHitContext(d),
		ContextSource:         ginContextSource(d),
		ClarificationNeeded:   ginClarificationNeeded(d),
		ClarificationReason:   ginClarificationReason(d),
		Answer:                strings.TrimSpace(ginAnswerText(d)),
		Success:               ginResponseCode(d) == 1,
		ShowMsg:               ginShowMsg(d),
		DebugMsg:              ginDebugMsg(d),
		Stage:                 ginStage(d),
		ErrorType:             ginErrorType(d),
		CloudAPI:              ginCloudAPI(d),
		CloudStatusCode:       ginCloudStatusCode(d),
		LLMModel:              ginLLMModel(d),
		RequestPayloadSummary: ginRequestPayloadSummary(d),
		ResultKind:            ginResultKind(d),
		ResultCount:           ginResultCount(d),
		ResultBrief:           ginResultBrief(d),
		ResponseSummary:       ginResponseSummary(d),
		ResponseSize:          ginResponseSize(d),
		CloudCalled:           ginCloudCalled(d),
		DurationMS:            ginDurationMS(d),
		LLMDurationMS:         llmDurationMS(d),
		CloudDurationMS:       cloudDurationMS(d),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if d != nil && d.Conversation != nil {
		doc.Intent = strings.TrimSpace(d.Conversation.LastIntent)
	}
	return service.Upsert(actx, effectiveMessageDatabase(cfg), "", doc)
}

func ginAnswerText(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return d.Gin.GetString("ai_robot_answer_text")
}

func ginHitContext(d *deps.Deps) bool {
	if d == nil || d.Gin == nil {
		return false
	}
	value, ok := d.Gin.Get("ai_robot_hit_context")
	if !ok {
		return false
	}
	hitContext, _ := value.(bool)
	return hitContext
}

func ginContextSource(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return "none"
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_context_source"))
}

func ginClarificationNeeded(d *deps.Deps) bool {
	if d == nil || d.Gin == nil {
		return false
	}
	value, ok := d.Gin.Get("ai_robot_clarification_needed")
	if !ok {
		return false
	}
	needed, _ := value.(bool)
	return needed
}

func ginClarificationReason(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_clarification_reason"))
}

func ginResponseCode(d *deps.Deps) int {
	if d == nil || d.Gin == nil {
		return 0
	}
	return d.Gin.GetInt("ai_robot_response_code")
}

func ginShowMsg(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_show_msg"))
}

func ginDebugMsg(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_debug_msg"))
}

func ginCloudCalled(d *deps.Deps) bool {
	if d == nil || d.Gin == nil {
		return false
	}
	value, ok := d.Gin.Get("ai_robot_cloud_called")
	if !ok {
		return false
	}
	called, _ := value.(bool)
	return called
}

func ginCloudAPI(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_cloud_api"))
}

func ginCloudStatusCode(d *deps.Deps) int {
	if d == nil || d.Gin == nil {
		return 0
	}
	return d.Gin.GetInt("ai_robot_cloud_status_code")
}

func ginLLMModel(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_llm_model"))
}

func ginRequestPayloadSummary(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_request_payload_summary"))
}

func ginResponseSize(d *deps.Deps) int64 {
	if d == nil || d.Gin == nil {
		return 0
	}
	return d.Gin.GetInt64("ai_robot_response_size")
}

func ginResponseSummary(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_response_summary"))
}

func ginResultKind(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_result_kind"))
}

func ginResultCount(d *deps.Deps) int {
	if d == nil || d.Gin == nil {
		return 0
	}
	return d.Gin.GetInt("ai_robot_result_count")
}

func ginResultBrief(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_result_brief"))
}

func intToString(value int) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(raw)
}

func ginStage(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_stage"))
}

func ginErrorType(d *deps.Deps) string {
	if d == nil || d.Gin == nil {
		return ""
	}
	return strings.TrimSpace(d.Gin.GetString("ai_robot_error_type"))
}

func ginDurationMS(d *deps.Deps) int64 {
	if d == nil || d.Gin == nil {
		return 0
	}
	return aiRobotDurationMSFromContext(d.Gin)
}

func aiRobotDurationMSFromContext(ctx *gin.Context) int64 {
	if ctx == nil {
		return 0
	}
	value, ok := ctx.Get("ai_robot_started_at")
	if !ok {
		return 0
	}
	startedAt, ok := value.(time.Time)
	if !ok || startedAt.IsZero() {
		return 0
	}
	durationMS := time.Since(startedAt).Milliseconds()
	if durationMS < 0 {
		return 0
	}
	return durationMS
}

func llmDurationMS(d *deps.Deps) int64 {
	if d == nil {
		return 0
	}
	if d.Gin != nil {
		d.Gin.Set("ai_robot_llm_duration_ms", d.LLMDurationMS())
	}
	return d.LLMDurationMS()
}

func cloudDurationMS(d *deps.Deps) int64 {
	if d == nil {
		return 0
	}
	if d.Gin != nil {
		d.Gin.Set("ai_robot_cloud_duration_ms", d.CloudDurationMS())
	}
	return d.CloudDurationMS()
}

func aiRobotLLMDurationMSFromContext(ctx *gin.Context) int64 {
	if ctx == nil {
		return 0
	}
	return ctx.GetInt64("ai_robot_llm_duration_ms")
}

func aiRobotCloudDurationMSFromContext(ctx *gin.Context) int64 {
	if ctx == nil {
		return 0
	}
	return ctx.GetInt64("ai_robot_cloud_duration_ms")
}

func inferStageFromFailure(showMsg string, debugMsg string) string {
	combined := strings.ToLower(strings.TrimSpace(showMsg + " " + debugMsg))
	switch {
	case strings.Contains(combined, "参数"):
		return "request"
	case strings.Contains(combined, "上下文"):
		return "context_load"
	case strings.Contains(combined, "平台后端初始化"):
		return "backend_init"
	case strings.Contains(combined, "大模型"):
		return "llm"
	case strings.Contains(combined, "云平台") || strings.Contains(combined, "http 4") || strings.Contains(combined, "http 5"):
		return "cloud_call"
	default:
		return "business"
	}
}

func inferErrorType(ctx *gin.Context, showMsg string, debugMsg string) string {
	stage := ""
	if ctx != nil {
		stage = strings.TrimSpace(ctx.GetString("ai_robot_stage"))
	}
	if stage == "" {
		stage = inferStageFromFailure(showMsg, debugMsg)
	}
	switch stage {
	case "request":
		return "request_error"
	case "context_load":
		return "context_error"
	case "backend_init":
		return "backend_init_error"
	case "cloud_call":
		return "cloud_error"
	case "llm":
		return "llm_error"
	case "completed":
		return ""
	default:
		return "business_error"
	}
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
	if c.deps != nil && c.deps.Gin != nil {
		c.deps.Gin.Set("ai_robot_stage", "llm")
		if c.deps.Cfg != nil {
			c.deps.Gin.Set("ai_robot_llm_model", strings.TrimSpace(c.deps.Cfg.LLMModel))
		}
	}
	startedAt := time.Now()
	result, err := c.inner.ChatCompletion(ctx, userPrompt)
	if c.deps != nil {
		c.deps.AddLLMDuration(time.Since(startedAt))
	}
	return result, err
}
