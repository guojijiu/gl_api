// Package backends 按请求 platform 注入平台专属 Backend。
//
// 新增 AI 对话平台 checklist（少一步会导致「初始化失败 / 分流不到 / 能力校验失败」）：
//
//  1. Http/Requests/AiRobotRequests.go：定义平台常量（如 AiPlatformXxx = "xxx"），并写入 aiPlatformCapabilities（各 question_type）；
//     若生产启用 Config.GetAiGatewayConfig()，还需在配置侧允许该平台与 question_type（与 IsSupportedPlatform / IsSupportedQuestionType 一致）。
//
//  2. internal/backends：为本包 NewForPlatform 增加 case，返回对应 Backend（走云 REST 用 NewCloudGateway；仅占位/无下游 HTTP 用 NewMinimal；其它类型可新增 Backend 实现）。
//
//  3. Services/ai_robot/platform/<平台目录>/：实现 Dispatch… 及按 question_type、领域的处理逻辑；包内勿依赖 flow。
//
//  4. flow/AiRobotRegistry.go：platformRegistry 增加「平台常量 → handle…Platform」，handle 内仅调用上一步的 Dispatch。
//
//  5. （可选）AiRobotChatRequest.Validate()：若需对 platform 做额外校验，在此收紧；与 1 保持一致。
//
// flow 侧 platformRegistry 注释指向本 checklist。
package backends

import (
	"fmt"

	"cloud-platform-api/app/Http/Requests"
	intranetclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/client"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
)

func errPlatformAPIUninitialized(platform string) error {
	return fmt.Errorf("ai_robot: %s 平台 API 未初始化", platform)
}

// NewForPlatform 按 platform 构造 Backend；未在 switch 中注册则返回错误。
func NewForPlatform(platform string, publicAPI *cloudclient.API, intranetAPI *intranetclient.API) (Backend, error) {
	switch platform {
	case Requests.AiPlatformCloudPublic:
		if publicAPI == nil {
			return nil, errPlatformAPIUninitialized(platform)
		}
		return NewPublicCloudGateway(platform, publicAPI), nil
	case Requests.AiPlatformCloudIntranet:
		if intranetAPI == nil {
			return nil, errPlatformAPIUninitialized(platform)
		}
		return NewIntranetCloudGateway(platform, intranetAPI), nil
	case Requests.AiPlatformImageCompare:
		return NewMinimal(platform), nil
	default:
		return nil, fmt.Errorf("ai_robot: 未注册平台后端: %s", platform)
	}
}
