package projectarticle

import (
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/cloudfail"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/reply"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"
	"cloud-platform-api/app/Services/ai_robot/policy"
)

// Dispatch 处理 question_type 为「项目文章」的请求。
func Dispatch(r deps.Responder, d *deps.Deps) {
	plan, err := intent.PlanProjectArticleIntent(d.Ctx, d.LLM, d.Question(), d.ContextSummary())
	if err != nil {
		r.FrontFailed(d.Gin, "意图解析失败", err)
		return
	}

	if !policy.IsIntentAllowed(d.PlatformID(), d.Req.QuestionType, plan.Intent) {
		r.FrontFailed(d.Gin, "当前平台下不支持该项目文章能力", nil)
		return
	}
	switch plan.Intent {
	case intent.IntentUnsupported:
		_ = reply.RespondUnsupported(r, d, plan.Intent, plan.Reason)
	case intent.IntentProjectArticleList:
		handleProjectArticleList(r, d, plan.Intent)
	default:
		r.FrontFailed(d.Gin, "未知意图分支", nil)
	}
}

func handleProjectArticleList(r deps.Responder, d *deps.Deps, intentKey string) {
	articleFilters := parse.ExtractProjectArticleFiltersFromQuestion(d.Question())
	d.RememberProjectArticleFilters(articleFilters)
	raw, status, err := deps.TrackNamedCloudCall("project_article.list", d, func() ([]byte, int, error) {
		return d.Cloud().ProjectArticle().List(d.Ctx, d.PlatformID(), d.Token, articleFilters)
	})
	if cloudfail.HandleCloudCallFailure(r, d, "请求云平台失败", status, err) {
		return
	}
	_ = reply.RespondCloudListSummary(r, d, intentKey, raw, summary.SummarizeProjectArticleData)
}
