package intent

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/llmutil"
)

func PlanProjectIntent(ctx context.Context, c llm.ChatCompletionClient, userQuestion string) (*IntentPlan, error) {
	prompt := fmt.Sprintf(`你是云平台「项目」类接口的路由决策器。根据用户问题，只能输出以下 intent 之一：
- project_list：用户想查看自己有哪些项目、项目列表、我的项目等
- project_expiring：用户关心即将到期、快过期、什么时候到期、剩余期限等
- contract_list：用户想查看自己有哪些合同、合同列表
- contract_projects：用户想看某个合同下有哪些项目
- download_final_report：用户要下载结题报告、质控报告、项目压缩包等
- download_original_data：用户要下载原始数据，通常会提到原始数据、授权码、提取码
- unsupported：与项目列表/到期无关，或需要的能力尚未开放

用户问题：%s

只输出一个 JSON 对象，不要 markdown，不要其它文字，格式：
{"intent":%s,"reason":"不超过80字的中文原因"}`, userQuestion, BuildIntentEnumForPrompt(ProjectIntentWhitelist))
	text, err := c.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := llmutil.ExtractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, llmutil.Truncate(text, 500))
	}
	switch plan.Intent {
	case IntentProjectList, IntentProjectExpiring, IntentContractList, IntentContractProjects, IntentDownloadFinal, IntentDownloadOriginal, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}

func PlanTaskIntent(ctx context.Context, c llm.ChatCompletionClient, userQuestion string) (*IntentPlan, error) {
	prompt := fmt.Sprintf(`你是云平台「任务」类接口的路由决策器。根据用户问题，只能输出以下 intent 之一：
- task_download_result：用户要下载某个任务的结果、导出结果、下载zip等（通常会带任务编号/uuid）
- task_status：用户问任务有没有完成、成功还是失败、失败原因是什么、为什么失败
- unsupported：与任务下载/状态无关，或能力尚未开放

用户问题：%s

只输出一个 JSON 对象，不要 markdown，不要其它文字，格式：
{"intent":%s,"reason":"不超过80字的中文原因"}`, userQuestion, BuildIntentEnumForPrompt(TaskIntentWhitelist))
	text, err := c.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := llmutil.ExtractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, llmutil.Truncate(text, 500))
	}
	switch plan.Intent {
	case IntentTaskDownloadResult, IntentTaskStatus, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}

func PlanProjectArticleIntent(ctx context.Context, c llm.ChatCompletionClient, userQuestion string) (*IntentPlan, error) {
	prompt := fmt.Sprintf(`你是云平台「项目文章」类接口的路由决策器。根据用户问题，只能输出以下 intent 之一：
- project_article_list：用户想查看项目文章数据

用户问题：%s

只输出一个 JSON 对象，不要 markdown，不要其它文字，格式：
{"intent":%s,"reason":"不超过80字的中文原因"}`, userQuestion, BuildIntentEnumForPrompt(ProjectArticleIntentWhitelist))
	text, err := c.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := llmutil.ExtractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, llmutil.Truncate(text, 500))
	}
	switch plan.Intent {
	case IntentProjectArticleList, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}
