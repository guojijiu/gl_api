package flow

import AiRobotConfig "cloud-platform-api/app/Services/ai_robot/config"

func isIntentAllowed(platform string, questionType int, intent string) bool {
	return AiRobotConfig.IsIntentAllowed(platform, questionType, intent)
}
