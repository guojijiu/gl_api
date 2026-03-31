package project

import "cloud-platform-api/app/Services/ai_robot/internal/deps"

// Dispatch 内网云平台 project 能力占位：当前仅保留架构入口，不承载业务逻辑。
func Dispatch(r deps.Responder, d *deps.Deps) {
	r.FrontFailed(d.Gin, "内网云平台项目能力暂未对接", nil)
}
