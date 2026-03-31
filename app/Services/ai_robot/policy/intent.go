package policy

import "cloud-platform-api/app/Config"

// IsIntentAllowed 判断某平台在某 question_type 下是否允许该意图（能力矩阵）。
// 数据来自应用 Config，此处为服务层入口，避免业务直接依赖全局 Config 细节。
func IsIntentAllowed(platform string, questionType int, intent string) bool {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		return true
	}
	return cfg.IsIntentAllowed(platform, questionType, intent)
}
