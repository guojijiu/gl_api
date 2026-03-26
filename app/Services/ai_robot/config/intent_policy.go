package config

import "cloud-platform-api/app/Config"

// IsIntentAllowed 判断某个平台在某业务类型下是否允许某个意图。
// 该函数是 ai_robot 服务层的配置入口，避免业务层直接依赖全局配置细节。
func IsIntentAllowed(platform string, questionType int, intent string) bool {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		return true
	}
	return cfg.IsIntentAllowed(platform, questionType, intent)
}
