package projectarticle

import (
	"fmt"

	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/intent"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/summary"
	"cloud-platform-api/app/Services/ai_robot/policy"

	"github.com/gin-gonic/gin"
)

func failIfProjectArticleCloudCallFailed(r deps.Responder, d *deps.Deps, errMsg string, status int, err error) bool {
	if err != nil {
		r.FrontFailed(d.Gin, errMsg, err)
		return true
	}
	if status >= 400 {
		r.FrontFailed(d.Gin, fmt.Sprintf("云平台返回 HTTP %d", status), nil)
		return true
	}
	return false
}

func summarizeProjectArticleWithData(r deps.Responder, d *deps.Deps, intentKey string, raw []byte) bool {
	answer, err := summary.SummarizeWithData(d.Ctx, d.LLM, d.Req.Question, raw)
	if err != nil {
		r.FrontFailed(d.Gin, "生成自然语言回复失败", err)
		return false
	}
	r.FrontSuccess(d.Gin, "操作成功", gin.H{
		"answer":         answer,
		"intent":         intentKey,
		"cloud_called":   true,
		"strict_mode":    d.Cfg.StrictMode,
		"raw_cloud_json": cloudclient.JsonRaw(raw),
	})
	return true
}

// Dispatch 处理 question_type 为「项目文章」的请求。
func Dispatch(r deps.Responder, d *deps.Deps) {
	plan, err := intent.PlanProjectArticleIntent(d.Ctx, d.LLM, d.Req.Question)
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
		answer, e := summary.SummarizeUnsupported(d.Ctx, d.LLM, d.Req.Question, plan.Reason)
		if e != nil {
			r.FrontFailed(d.Gin, "生成回复失败", e)
			return
		}
		r.FrontSuccess(d.Gin, "操作成功", gin.H{
			"answer":         answer,
			"intent":         plan.Intent,
			"cloud_called":   false,
			"strict_mode":    d.Cfg.StrictMode,
			"raw_cloud_json": nil,
		})
	case intent.IntentProjectArticleList:
		handleProjectArticleList(r, d, plan.Intent)
	default:
		r.FrontFailed(d.Gin, "未知意图分支", nil)
	}
}

func handleProjectArticleList(r deps.Responder, d *deps.Deps, intentKey string) {
	raw, status, err := d.Cloud().ProjectArticle().List(d.Ctx, d.PlatformID(), d.Token)
	if failIfProjectArticleCloudCallFailed(r, d, "请求云平台失败", status, err) {
		return
	}
	_ = summarizeProjectArticleWithData(r, d, intentKey, raw)
}
