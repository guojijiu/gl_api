package cloudintranet

import (
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/project"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/task"
)

// DispatchIntranet 内网云平台：与公网逻辑独立维护（初版由公网复制，可各自演进）。
func DispatchIntranet(r deps.Responder, d *deps.Deps) {
	if d.Cloud() == nil {
		r.FrontFailed(d.Gin, "内部错误：内网路由需要云平台后端", nil)
		return
	}
	switch d.Req.QuestionType {
	case Requests.AiQuestionTypeProject:
		project.Dispatch(r, d)
	case Requests.AiQuestionTypeTask:
		task.Dispatch(r, d)
	default:
		r.FrontFailed(d.Gin, "内网云平台下不支持的提问类型", nil)
	}
}
