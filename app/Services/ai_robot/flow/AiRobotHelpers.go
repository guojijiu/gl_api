package flow

import AiRobotUtils "cloud-platform-api/app/Services/ai_robot/utils"

func generateAccessCode(length int) string {
	return AiRobotUtils.GenerateAccessCode(length)
}
