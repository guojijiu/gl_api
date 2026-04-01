package deps

import (
	"context"
	"encoding/json"
	"strings"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/backends"
	"cloud-platform-api/app/Services/ai_robot/internal/conversation"
	"cloud-platform-api/app/Services/ai_robot/internal/filters"

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

	ConversationStore conversation.Store
	Conversation      *conversation.State
}

func (d *Deps) CurrentUserID() string {
	if d == nil {
		return ""
	}
	if d.Req != nil && strings.TrimSpace(d.Req.UserID) != "" {
		return strings.TrimSpace(d.Req.UserID)
	}
	if d.Gin != nil {
		return strings.TrimSpace(d.Gin.GetString("user_id"))
	}
	return ""
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

func (d *Deps) Question() string {
	if d == nil || d.Req == nil {
		return ""
	}
	if q := strings.TrimSpace(d.Req.ResolvedQuestion); q != "" {
		return q
	}
	return strings.TrimSpace(d.Req.Question)
}

func (d *Deps) RememberIntent(intent string) {
	if d == nil || d.Conversation == nil {
		return
	}
	d.Conversation.LastIntent = intent
	d.Conversation.LastQuestion = strings.TrimSpace(d.Req.Question)
	d.Conversation.LastResolved = d.Question()
}

func (d *Deps) RememberProjectMatch(ids []int, number string) {
	if d == nil || d.Conversation == nil {
		return
	}
	d.Conversation.LastProjectIDs = append([]int(nil), ids...)
	d.Conversation.LastProjectNumber = strings.TrimSpace(number)
}

func (d *Deps) RememberContractMatch(ids []int, number string) {
	if d == nil || d.Conversation == nil {
		return
	}
	d.Conversation.LastContractIDs = append([]int(nil), ids...)
	d.Conversation.LastContractNumber = strings.TrimSpace(number)
}

func (d *Deps) RememberTaskUUIDs(uuids []string) {
	if d == nil || d.Conversation == nil {
		return
	}
	d.Conversation.LastTaskUUIDs = append([]string(nil), uuids...)
}

func (d *Deps) RememberProjectArticleFilters(opts filters.ProjectArticleFilters) {
	if d == nil || d.Conversation == nil {
		return
	}
	for _, item := range opts.SearchFilter {
		switch item.Column {
		case "name_cn":
			d.Conversation.LastArticleNameCN = strings.TrimSpace(item.Value)
		case "name_en":
			d.Conversation.LastArticleNameEN = strings.TrimSpace(item.Value)
		case "journal_name":
			d.Conversation.LastJournalName = strings.TrimSpace(item.Value)
		case "product_label":
			if v := strings.TrimSpace(item.Value); v != "" {
				d.Conversation.LastArticleLabels = append(d.Conversation.LastArticleLabels, v)
			}
		}
	}
}

func (d *Deps) ContextSummary() string {
	if d == nil || d.Conversation == nil {
		return ""
	}
	return strings.TrimSpace(d.Conversation.ContextSummary)
}

func (d *Deps) RememberTaskStatusResult(statusRaw []byte) {
	if d == nil || d.Conversation == nil || len(statusRaw) == 0 {
		return
	}
	var root map[string]interface{}
	if err := json.Unmarshal(statusRaw, &root); err != nil {
		return
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return
	}
	first, _ := data[0].(map[string]interface{})
	if first == nil {
		return
	}
	if v, ok := first["status_value"].(string); ok {
		d.Conversation.LastTaskStatus = strings.TrimSpace(v)
	}
	for _, key := range []string{"fail_reason", "failure_reason", "reason", "error_msg", "message"} {
		if v, ok := first[key].(string); ok && strings.TrimSpace(v) != "" {
			d.Conversation.LastFailureReason = strings.TrimSpace(v)
			break
		}
	}
}

func (d *Deps) RememberDownloadResult(kind string, hasLink bool) {
	if d == nil || d.Conversation == nil {
		return
	}
	d.Conversation.LastDownloadKind = strings.TrimSpace(kind)
	d.Conversation.LastHasLink = hasLink
}
