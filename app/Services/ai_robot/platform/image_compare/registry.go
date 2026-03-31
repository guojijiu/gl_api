package imagecompare

import (
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/image_compare/project"
)

// Dispatch 图片对比平台入口；业务仅在本目录内扩展，勿写在 flow。
func Dispatch(r deps.Responder, d *deps.Deps) {
	switch d.Req.QuestionType {
	case Requests.AiQuestionTypeProject:
		project.Dispatch(r, d)
	default:
		r.FrontFailed(d.Gin, "图片对比平台下不支持的提问类型", nil)
	}
}
