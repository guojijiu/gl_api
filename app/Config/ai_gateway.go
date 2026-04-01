package Config

import (
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// AiGatewayConfig AI 网关：大模型连接信息、云平台站点根地址、以及行为/能力矩阵。
// Laravel 各接口 path 与完整 URL 拼装见 cloud_public/cloudapi。
//
// 说明（给不熟悉 Go 的同事）：
//   - 本结构体里的字段会通过 viper 从环境变量读入（见 BindEnvs）。
//   - mapstructure 标签是 viper 把配置树映射到字段时用的「路径」，一般不用改。
//   - 大模型侧采用百炼的 **OpenAI 兼容模式**：请求格式与 OpenAI Chat Completions 相同，
//     只需把 Base URL 换成百炼地址、API Key 换成百炼控制台申请的 Key、模型名换成 qwen-* 即可。
//
// 百炼文档（OpenAI 兼容）：https://help.aliyun.com/zh/model-studio/compatibility-of-openai-with-dashscope
type AiGatewayConfig struct {
	// CloudPlatformBaseURL 你的 Laravel 云平台站点根地址，**不要**带末尾斜杠。
	// 例：https://www.example.com
	// 程序会拼出：{CloudPlatformBaseURL}/api/front/project/get_user_all_project
	CloudPlatformBaseURL string `mapstructure:"cloud_platform_base_url"`

	// CloudPlatformIntranetBaseURL 内网云平台根地址（可选）。
	// 未来可用于把相同的“项目”能力转发到内网环境，调用方式与公网一致。
	// 例：http://10.0.0.12 或 http://intranet-cloud.example.local
	CloudPlatformIntranetBaseURL string `mapstructure:"cloud_platform_intranet_base_url"`

	// LLMBaseURL 百炼 OpenAI 兼容端点（到 /v1 为止，不要写 /chat/completions）。
	// 国内（北京）常用默认值见 SetDefaults；新加坡、美东等地域域名不同，以控制台说明为准。
	LLMBaseURL string `mapstructure:"llm_base_url"`

	// LLMAPIKey 百炼 API Key（控制台创建）。请求头：Authorization: Bearer <key>
	LLMAPIKey string `mapstructure:"llm_api_key"`

	// LLMModel 模型名，如 qwen-plus、qwen-turbo、qwen-max 等，以百炼当前可用列表为准。
	LLMModel string `mapstructure:"llm_model"`

	// LLMTimeoutSec 单次 HTTP 调用大模型/云平台的超时（秒）。大模型推理较慢，建议 ≥ 60。
	LLMTimeoutSec int `mapstructure:"llm_timeout_sec"`

	// ExpiringWithinDays 「即将到期」筛选：只保留有效期结束日在 [今天, 今天+N 天] 内的项目。
	ExpiringWithinDays int `mapstructure:"expiring_within_days"`

	// StrictMode 严格模式：
	// true 时下载场景必须按“项目编号”命中，不再回退到大模型猜测。
	StrictMode bool `mapstructure:"strict_mode"`

	// CapabilityMatrixJSON 能力矩阵 JSON（可选，用于配置化平台/业务/intent 支持关系）。
	// 结构示例：
	// {
	//   "cloud_public":{"1":["project_list","unsupported"],"2":["task_status"]},
	//   "cloud_intranet":{"1":["project_list"],"2":["task_download_result"]},
	//   "image_compare":{"1":["unsupported"]}
	// }
	CapabilityMatrixJSON string `mapstructure:"capability_matrix_json"`

	// ConversationStoreDriver 对话上下文存储驱动：mongodb / memory。
	ConversationStoreDriver string `mapstructure:"conversation_store_driver"`
	// ConversationMongoURI Mongo 连接串；如已提供则优先使用。
	ConversationMongoURI string `mapstructure:"conversation_mongo_uri"`
	// ConversationMongoHost Mongo 主机。
	ConversationMongoHost string `mapstructure:"conversation_mongo_host"`
	// ConversationMongoPort Mongo 端口。
	ConversationMongoPort int `mapstructure:"conversation_mongo_port"`
	// ConversationMongoDatabase Mongo 数据库名。
	ConversationMongoDatabase string `mapstructure:"conversation_mongo_database"`
	// ConversationMongoCollection Mongo 集合名。
	ConversationMongoCollection string `mapstructure:"conversation_mongo_collection"`
	// ConversationMongoUsername Mongo 用户名。
	ConversationMongoUsername string `mapstructure:"conversation_mongo_username"`
	// ConversationMongoPassword Mongo 密码。
	ConversationMongoPassword string `mapstructure:"conversation_mongo_password"`
	// ConversationMongoAuthSource Mongo 认证库。
	ConversationMongoAuthSource string `mapstructure:"conversation_mongo_auth_source"`
	// ConversationMongoTimeoutSec Mongo 连接超时秒数。
	ConversationMongoTimeoutSec int `mapstructure:"conversation_mongo_timeout_sec"`
	// ConversationTTLHours 对话上下文保留时长，单位小时；用于 TTL 清理。
	ConversationTTLHours int `mapstructure:"conversation_ttl_hours"`
	// ConversationAutoFillMaxAgeMinutes 自动补参上下文最大有效分钟数；超时后只保留摘要，不自动补参。
	ConversationAutoFillMaxAgeMinutes int `mapstructure:"conversation_auto_fill_max_age_minutes"`
	// ConversationAutoFillMaxTurns 自动补参最多参考最近连续几轮；更早轮次只保留给模型参考。
	ConversationAutoFillMaxTurns int `mapstructure:"conversation_auto_fill_max_turns"`
	// ConversationSummaryMaxTurns 上下文摘要最多参考最近连续几轮；可大于自动补参窗口。
	ConversationSummaryMaxTurns int `mapstructure:"conversation_summary_max_turns"`
}

// SetDefaults 设置 viper 默认值（环境变量未配置时使用）。
// 默认按 **阿里云百炼 · 北京地域 · OpenAI 兼容模式** 填写；若你使用新加坡等国际域名，请用环境变量覆盖 LLMBaseURL。
func (c *AiGatewayConfig) SetDefaults() {
	// 百炼 OpenAI 兼容 Base URL（北京）。完整请求示例：POST {LLMBaseURL}/chat/completions
	viper.SetDefault("ai_gateway.llm_base_url", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	// 通义千问常用模型，可按成本/效果在控制台换为 qwen-max、qwen-turbo 等
	viper.SetDefault("ai_gateway.llm_model", "qwen-plus")
	viper.SetDefault("ai_gateway.llm_timeout_sec", 120)
	viper.SetDefault("ai_gateway.expiring_within_days", 30)
	viper.SetDefault("ai_gateway.strict_mode", false)
	viper.SetDefault("ai_gateway.capability_matrix_json", "")
	viper.SetDefault("ai_gateway.conversation_store_driver", "mongodb")
	viper.SetDefault("ai_gateway.conversation_mongo_host", "127.0.0.1")
	viper.SetDefault("ai_gateway.conversation_mongo_port", 27017)
	viper.SetDefault("ai_gateway.conversation_mongo_database", "cloud_platform_v2")
	viper.SetDefault("ai_gateway.conversation_mongo_collection", "ai_robot_conversations")
	viper.SetDefault("ai_gateway.conversation_mongo_auth_source", "admin")
	viper.SetDefault("ai_gateway.conversation_mongo_timeout_sec", 5)
	viper.SetDefault("ai_gateway.conversation_ttl_hours", 24*30)
	viper.SetDefault("ai_gateway.conversation_auto_fill_max_age_minutes", 10)
	viper.SetDefault("ai_gateway.conversation_auto_fill_max_turns", 3)
	viper.SetDefault("ai_gateway.conversation_summary_max_turns", 5)
}

// BindEnvs 把环境变量绑定到 viper 的键上；LoadConfig() 时会 Unmarshal 进 AiGatewayConfig。
//
// 部署时最少需要：
//   - CLOUD_PLATFORM_BASE_URL
//   - AI_GATEWAY_LLM_API_KEY（百炼 API Key）
//
// 可选覆盖：AI_GATEWAY_LLM_BASE_URL、AI_GATEWAY_LLM_MODEL、超时与到期天数。
func (c *AiGatewayConfig) BindEnvs() {
	viper.BindEnv("ai_gateway.cloud_platform_base_url", "CLOUD_PLATFORM_BASE_URL")
	viper.BindEnv("ai_gateway.cloud_platform_intranet_base_url", "CLOUD_PLATFORM_INTRANET_BASE_URL")
	viper.BindEnv("ai_gateway.llm_base_url", "AI_GATEWAY_LLM_BASE_URL")
	viper.BindEnv("ai_gateway.llm_api_key", "AI_GATEWAY_LLM_API_KEY")
	viper.BindEnv("ai_gateway.llm_model", "AI_GATEWAY_LLM_MODEL")
	viper.BindEnv("ai_gateway.llm_timeout_sec", "AI_GATEWAY_LLM_TIMEOUT_SEC")
	viper.BindEnv("ai_gateway.expiring_within_days", "AI_GATEWAY_EXPIRING_DAYS")
	viper.BindEnv("ai_gateway.strict_mode", "AI_GATEWAY_STRICT_MODE")
	viper.BindEnv("ai_gateway.capability_matrix_json", "AI_GATEWAY_CAPABILITY_MATRIX_JSON")
	viper.BindEnv("ai_gateway.conversation_store_driver", "AI_ROBOT_CONVERSATION_STORE_DRIVER")
	viper.BindEnv("ai_gateway.conversation_mongo_uri", "AI_ROBOT_CONVERSATION_MONGO_URI")
	viper.BindEnv("ai_gateway.conversation_mongo_host", "AI_ROBOT_CONVERSATION_MONGO_HOST", "MONGODB_HOST")
	viper.BindEnv("ai_gateway.conversation_mongo_port", "AI_ROBOT_CONVERSATION_MONGO_PORT", "MONGODB_PORT")
	viper.BindEnv("ai_gateway.conversation_mongo_database", "AI_ROBOT_CONVERSATION_MONGO_DATABASE", "MONGODB_DATABASE")
	viper.BindEnv("ai_gateway.conversation_mongo_collection", "AI_ROBOT_CONVERSATION_MONGO_COLLECTION")
	viper.BindEnv("ai_gateway.conversation_mongo_username", "AI_ROBOT_CONVERSATION_MONGO_USERNAME", "MONGODB_USERNAME")
	viper.BindEnv("ai_gateway.conversation_mongo_password", "AI_ROBOT_CONVERSATION_MONGO_PASSWORD", "MONGODB_PASSWORD")
	viper.BindEnv("ai_gateway.conversation_mongo_auth_source", "AI_ROBOT_CONVERSATION_MONGO_AUTH_SOURCE", "MONGODB_AUTH_SOURCE")
	viper.BindEnv("ai_gateway.conversation_mongo_timeout_sec", "AI_ROBOT_CONVERSATION_MONGO_TIMEOUT_SEC", "MONGODB_TIMEOUT_SEC")
	viper.BindEnv("ai_gateway.conversation_ttl_hours", "AI_ROBOT_CONVERSATION_TTL_HOURS")
	viper.BindEnv("ai_gateway.conversation_auto_fill_max_age_minutes", "AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_AGE_MINUTES")
	viper.BindEnv("ai_gateway.conversation_auto_fill_max_turns", "AI_ROBOT_CONVERSATION_AUTO_FILL_MAX_TURNS")
	viper.BindEnv("ai_gateway.conversation_summary_max_turns", "AI_ROBOT_CONVERSATION_SUMMARY_MAX_TURNS")
}

func (c *AiGatewayConfig) parsedCapabilityMatrix() map[string]map[int]map[string]struct{} {
	raw := strings.TrimSpace(c.CapabilityMatrixJSON)
	if raw == "" {
		return nil
	}
	var m map[string]map[string][]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	out := map[string]map[int]map[string]struct{}{}
	for platform, byType := range m {
		out[platform] = map[int]map[string]struct{}{}
		for qTypeText, intents := range byType {
			qType, err := strconv.Atoi(strings.TrimSpace(qTypeText))
			if err != nil {
				continue
			}
			if _, ok := out[platform][qType]; !ok {
				out[platform][qType] = map[string]struct{}{}
			}
			for _, intent := range intents {
				intent = strings.TrimSpace(intent)
				if intent == "" {
					continue
				}
				out[platform][qType][intent] = struct{}{}
			}
		}
	}
	return out
}

func defaultCapabilityMatrix() map[string]map[int]map[string]struct{} {
	return map[string]map[int]map[string]struct{}{
		"cloud_public": {
			1: {"project_list": {}, "project_expiring": {}, "contract_list": {}, "contract_projects": {}, "download_final_report": {}, "download_original_data": {}, "unsupported": {}},
			2: {"task_status": {}, "task_download_result": {}, "unsupported": {}},
			3: {"unsupported": {}, "project_article_list": {}},
		},
		"cloud_intranet": {
			1: {"project_list": {}, "project_expiring": {}, "contract_list": {}, "contract_projects": {}, "download_final_report": {}, "download_original_data": {}, "unsupported": {}},
			2: {"task_status": {}, "task_download_result": {}, "unsupported": {}},
		},
		"image_compare": {
			1: {"unsupported": {}},
		},
	}
}

func (c *AiGatewayConfig) effectiveCapabilityMatrixRaw() map[string]map[int]map[string]struct{} {
	if parsed := c.parsedCapabilityMatrix(); parsed != nil {
		return parsed
	}
	return defaultCapabilityMatrix()
}

func (c *AiGatewayConfig) IsPlatformSupported(platform string) bool {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return false
	}
	_, ok := c.effectiveCapabilityMatrixRaw()[platform]
	return ok
}

func (c *AiGatewayConfig) IsQuestionTypeAllowed(platform string, questionType int) bool {
	platform = strings.TrimSpace(platform)
	byType, ok := c.effectiveCapabilityMatrixRaw()[platform]
	if !ok {
		return false
	}
	_, ok = byType[questionType]
	return ok
}

func (c *AiGatewayConfig) IsIntentAllowed(platform string, questionType int, intent string) bool {
	intent = strings.TrimSpace(intent)
	byType, ok := c.effectiveCapabilityMatrixRaw()[strings.TrimSpace(platform)]
	if !ok {
		return false
	}
	intents, ok := byType[questionType]
	if !ok {
		return false
	}
	_, ok = intents[intent]
	return ok
}

func (c *AiGatewayConfig) CapabilityMatrixDebugView() map[string]interface{} {
	source := "default"
	if c.parsedCapabilityMatrix() != nil {
		source = "env_json"
	}
	view := map[string]interface{}{
		"source":   source,
		"raw_json": c.CapabilityMatrixJSON,
		"matrix":   map[string]interface{}{},
	}
	matrixView := view["matrix"].(map[string]interface{})
	for platform, byType := range c.effectiveCapabilityMatrixRaw() {
		typeView := map[string]interface{}{}
		for qType, intents := range byType {
			var list []string
			for intent := range intents {
				list = append(list, intent)
			}
			sort.Strings(list)
			typeView[strconv.Itoa(qType)] = list
		}
		matrixView[platform] = typeView
	}
	return view
}

func (c *AiGatewayConfig) EffectiveConversationMongoConfig(shared *MongoDBConfig) MongoDBConfig {
	effective := MongoDBConfig{
		Host:       "127.0.0.1",
		Port:       27017,
		Database:   "cloud_platform_v2",
		AuthSource: "admin",
		TimeoutSec: 5,
	}
	if shared != nil {
		effective = *shared
	}
	if raw := strings.TrimSpace(c.ConversationMongoURI); raw != "" {
		effective.URI = raw
	}
	if host := strings.TrimSpace(c.ConversationMongoHost); host != "" {
		effective.Host = host
	}
	if c.ConversationMongoPort > 0 {
		effective.Port = c.ConversationMongoPort
	}
	if database := strings.TrimSpace(c.ConversationMongoDatabase); database != "" {
		effective.Database = database
	}
	if username := strings.TrimSpace(c.ConversationMongoUsername); username != "" {
		effective.Username = username
	}
	if c.ConversationMongoPassword != "" {
		effective.Password = c.ConversationMongoPassword
	}
	if authSource := strings.TrimSpace(c.ConversationMongoAuthSource); authSource != "" {
		effective.AuthSource = authSource
	}
	if c.ConversationMongoTimeoutSec > 0 {
		effective.TimeoutSec = c.ConversationMongoTimeoutSec
	}
	return effective
}

func (c *AiGatewayConfig) ConversationMongoConnString() string {
	effective := c.EffectiveConversationMongoConfig(GetMongoDBConfig())
	if raw := strings.TrimSpace(effective.URI); raw != "" {
		return raw
	}

	host := strings.TrimSpace(effective.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := effective.Port
	if port <= 0 {
		port = 27017
	}
	database := strings.TrimSpace(effective.Database)
	if database == "" {
		database = "cloud_platform_v2"
	}
	authSource := strings.TrimSpace(effective.AuthSource)
	if authSource == "" {
		authSource = "admin"
	}

	credentials := ""
	if user := strings.TrimSpace(effective.Username); user != "" {
		credentials = url.QueryEscape(user)
		if pwd := effective.Password; pwd != "" {
			credentials += ":" + url.QueryEscape(pwd)
		}
		credentials += "@"
	}

	params := url.Values{}
	if credentials != "" && authSource != "" {
		params.Set("authSource", authSource)
	}
	query := params.Encode()
	if query != "" {
		query = "?" + query
	}
	return "mongodb://" + credentials + host + ":" + strconv.Itoa(port) + "/" + database + query
}

// CloudFrontAPIBase 返回去掉右侧斜杠的云平台根 URL，避免拼接时出现 //。
func (c *AiGatewayConfig) CloudFrontAPIBase() string {
	return strings.TrimRight(c.CloudPlatformBaseURL, "/")
}

// CloudIntranetAPIBase 返回去掉右侧斜杠的内网云平台根 URL。
func (c *AiGatewayConfig) CloudIntranetAPIBase() string {
	return strings.TrimRight(c.CloudPlatformIntranetBaseURL, "/")
}

// CloudAPIBaseByPlatform 根据平台标识返回对应云平台根地址（不带末尾斜杠）。
//
// platform 的值由请求参数决定（见 Requests.AiRobotChatRequest.Platform）。
// 目前仅实现两个云平台：公网/内网。
func (c *AiGatewayConfig) CloudAPIBaseByPlatform(platform string) string {
	switch platform {
	case "cloud_intranet":
		return c.CloudIntranetAPIBase()
	default: // cloud_public 或空值都走公网
		return c.CloudFrontAPIBase()
	}
}

// GetAiGatewayConfig 从全局 Config 取 AI 网关配置指针（main 里 LoadConfig 之后才有值）。
func GetAiGatewayConfig() *AiGatewayConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.AiGateway
}
