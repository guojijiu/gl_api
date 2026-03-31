package cloudpublic

import (
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/project"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/project_article"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/task"
)

// DispatchPublic 公网云平台：按 question_type 分发到领域处理。
func DispatchPublic(r deps.Responder, d *deps.Deps) {
	if d.Cloud() == nil {
		r.FrontFailed(d.Gin, "内部错误：公网路由需要云平台后端", nil)
		return
	}
	switch d.Req.QuestionType {
	case Requests.AiQuestionTypeProject:
		project.Dispatch(r, d)
	case Requests.AiQuestionTypeTask:
		task.Dispatch(r, d)
	case Requests.AiQuestionTypeProjectArticle:
		projectarticle.Dispatch(r, d)
	default:
		r.FrontFailed(d.Gin, "公网云平台下不支持的提问类型", nil)
	}
}
