package Config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// AiGatewayConfig AI 网关：连接「阿里云百炼大模型」与「自有 Laravel 云平台」。
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

// ProjectGetUserAllProjectURL Laravel 路由：routes/front.php 中 jwt 组内 project/get_user_all_project，
// 全局前缀一般为 api/front（以你 RouteServiceProvider 为准）。
func (c *AiGatewayConfig) ProjectGetUserAllProjectURL(platform string) string {
	return c.CloudAPIBaseByPlatform(platform) + "/api/front/project/get_user_all_project"
}

// ProjectGetZipURL 结题报告下载链接接口（front.php: project/zip_url）。
// 需要 query 参数：id（项目ID）。
func (c *AiGatewayConfig) ProjectGetZipURL(platform string, projectID int) string {
	return c.CloudAPIBaseByPlatform(platform) + "/api/front/project/zip_url?id=" + fmt.Sprintf("%d", projectID)
}

// ProjectGetOriginalDataURL 原始数据下载链接接口（front.php: project/original_data_url）。
// 需要 query 参数：id（原始数据ID）+ access_code（授权码）。
func (c *AiGatewayConfig) ProjectGetOriginalDataURL(platform string, dataID int, accessCode string) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/project/original_data_url?id=" + fmt.Sprintf("%d", dataID) +
		"&access_code=" + url.QueryEscape(accessCode)
}

// ContractListOfProjectURL 合同列表接口（front.php: contract/list_of_project）。
// 说明：该接口支持 page/size，可附带 contract_number 进行筛选。
func (c *AiGatewayConfig) ContractListOfProjectURL(platform string, page int, size int, contractNumber string) string {
	base := c.CloudAPIBaseByPlatform(platform) +
		"/api/front/contract/list_of_project?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(contractNumber) != "" {
		base += "&contract_number=" + url.QueryEscape(contractNumber)
	}
	return base
}

// ProjectListByContractURL 项目列表接口（front.php: project/list）。
func (c *AiGatewayConfig) ProjectListByContractURL(platform string, contractID int, page int, size int) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/project/list?contract_id=" + fmt.Sprintf("%d", contractID) +
		"&page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
}

// TaskListURL 任务列表接口（front.php: task/list），支持按 uuid 筛选。
func (c *AiGatewayConfig) TaskListURL(platform string, page int, size int, uuid string) string {
	base := c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/list?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		base += "&uuid=" + url.QueryEscape(uuid)
	}
	return base
}

// TaskStatusByUUIDsURL 获取指定 uuid 集合的状态（front.php: task/status_by_uuids）。
// uuids 以英文逗号拼接。
func (c *AiGatewayConfig) TaskStatusByUUIDsURL(platform string, uuids string) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/status_by_uuids?uuids=" + url.QueryEscape(uuids)
}

// TaskGetResultURL 获取任务结果详情（front.php: task/result?id=...），失败原因在返回内容里。
func (c *AiGatewayConfig) TaskGetResultURL(platform string, taskID int) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/result?id=" + fmt.Sprintf("%d", taskID)
}

// TaskDownloadResultURL 下载任务结果（front.php: task/download_result），POST body: {"id": <task_id>}
func (c *AiGatewayConfig) TaskDownloadResultURL(platform string) string {
	return c.CloudAPIBaseByPlatform(platform) + "/api/front/task/download_result"
}

// TaskListOfWorkflowURL 流程任务列表（front.php: task/list_of_workflow）。
func (c *AiGatewayConfig) TaskListOfWorkflowURL(platform string, page int, size int, uuid string) string {
	base := c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/list_of_workflow?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		base += "&uuid=" + url.QueryEscape(uuid)
	}
	return base
}

// TaskDetailOfModuleToolURL 模块化任务详情（front.php: task/detail_of_module_tool?id=...）。
func (c *AiGatewayConfig) TaskDetailOfModuleToolURL(platform string, id int) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/detail_of_module_tool?id=" + fmt.Sprintf("%d", id)
}

// TaskPdfURLOfModuleToolURL 模块化任务结果链接（front.php: task/pdf_url_of_module_tool）。
// type=2 表示 source_type=task。
func (c *AiGatewayConfig) TaskPdfURLOfModuleToolURL(platform string, taskID int) string {
	return c.CloudAPIBaseByPlatform(platform) +
		"/api/front/task/pdf_url_of_module_tool?type=2&source_id=" + fmt.Sprintf("%d", taskID)
}

// GetAiGatewayConfig 从全局 Config 取 AI 网关配置指针（main 里 LoadConfig 之后才有值）。
func GetAiGatewayConfig() *AiGatewayConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.AiGateway
}
