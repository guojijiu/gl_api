package aigateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type IntentPlan struct {
	Intent string `json:"intent"`
	Reason string `json:"reason"`
}

func (s *Service) validateGatewayForLLM() error {
	if s.cfg.LLMAPIKey == "" {
		return errors.New("未配置 AI_GATEWAY_LLM_API_KEY（百炼 API Key），无法调用大模型")
	}
	return nil
}

func (s *Service) PlanProjectIntent(ctx context.Context, userQuestion string) (*IntentPlan, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return nil, err
	}
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
	text, err := s.chatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := extractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, truncate(text, 500))
	}
	switch plan.Intent {
	case IntentProjectList, IntentProjectExpiring, IntentContractList, IntentContractProjects, IntentDownloadFinal, IntentDownloadOriginal, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}

func (s *Service) PlanTaskIntent(ctx context.Context, userQuestion string) (*IntentPlan, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return nil, err
	}
	prompt := fmt.Sprintf(`你是云平台「任务」类接口的路由决策器。根据用户问题，只能输出以下 intent 之一：
- task_download_result：用户要下载某个任务的结果、导出结果、下载zip等（通常会带任务编号/uuid）
- task_status：用户问任务有没有完成、成功还是失败、失败原因是什么、为什么失败
- unsupported：与任务下载/状态无关，或能力尚未开放

用户问题：%s

只输出一个 JSON 对象，不要 markdown，不要其它文字，格式：
{"intent":%s,"reason":"不超过80字的中文原因"}`, userQuestion, BuildIntentEnumForPrompt(TaskIntentWhitelist))
	text, err := s.chatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := extractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, truncate(text, 500))
	}
	switch plan.Intent {
	case IntentTaskDownloadResult, IntentTaskStatus, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}

func (s *Service) PlanProjectArticleIntent(ctx context.Context, userQuestion string) (*IntentPlan, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return nil, err
	}
	prompt := fmt.Sprintf(`你是云平台「项目文章」类接口的路由决策器。根据用户问题，只能输出以下 intent 之一：
- project_article_list：用户想查看项目文章数据

用户问题：%s

只输出一个 JSON 对象，不要 markdown，不要其它文字，格式：
{"intent":%s,"reason":"不超过80字的中文原因"}`, userQuestion, BuildIntentEnumForPrompt(TaskIntentWhitelist))
	text, err := s.chatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	raw := extractJSONFromLLM(text)
	var plan IntentPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析意图 JSON 失败: %w, 原文: %s", err, truncate(text, 500))
	}
	switch plan.Intent {
	case IntentProjectArticleList, IntentUnsupported:
		return &plan, nil
	default:
		return &IntentPlan{Intent: IntentUnsupported, Reason: "模型返回未知 intent"}, nil
	}
}

func (s *Service) SummarizeWithData(ctx context.Context, userQuestion string, apiJSON []byte) (string, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return "", err
	}
	prompt := fmt.Sprintf(`你是云平台客服助手。请严格按以下规则回复，且只输出最终给客户的答复，不要任何额外信息。

用户问题：
%s

云平台接口返回 JSON（可能包含 code、showMsg、content 等字段）：
%s

回复规则（按优先级）：
1）若 code=2（账号登录问题）：
   - 若 showMsg 表示“账号已在另一台设备登录/异地登录/其他设备登录”，只回复“您的账号已在另一台设备登录，请重新登录。”。
   - 其他登录问题：根据 showMsg 翻译成客户易识别中文，直接告知客户下一步处理（如重新登录），不要解释系统细节。
2）若接口结果正常且存在有效数据：用简短、清晰的中文直接回答用户问题。
3）若接口结果正常但没有有效数据（如列表为空、content 为空等）：只回复“未查询到相关数据”。
4）若接口错误（如 code 不为 1 或存在明确错误信息）：根据 showMsg 翻译成客户易识别的中文；若 showMsg 缺失，则只回复“未查询到相关数据”。

格式规则：
- 输出 Markdown 正文，不要代码块。
- 仅输出客户可见内容，不要标题“结果/说明”等包装词。
- 有多条结果时可用 Markdown 列表；单条结果时用一行文本。

禁止项：
- 不要礼貌寒暄
- 不要解释规则
- 不要补充建议
- 不要编造任何接口中不存在的信息
- 不要出现“根据接口返回/接口显示/code/showMsg/系统层面/客户无需关注细节”等描述
`, userQuestion, string(apiJSON))
	return s.chatCompletion(ctx, prompt)
}

func (s *Service) SummarizeUnsupported(ctx context.Context, userQuestion string, reason string) (string, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return "", err
	}
	prompt := fmt.Sprintf(`你是云平台客服助手。请严格按以下规则回复，且只输出最终给客户的答复，不要任何额外信息。

用户问题：%s
系统判断信息：%s

回复规则（按优先级）：
1）若系统判断信息表明 code=2（账号登录问题）：
   - 若信息中体现“账号已在另一台设备登录/异地登录/其他设备登录”，只回复“您的账号已在另一台设备登录，请重新登录。”。
   - 其他登录问题：根据信息翻译成客户易识别中文，直接告知客户下一步处理（如重新登录），不要解释系统细节。
2）若系统判断信息中有明确可用结果：用简短、清晰的中文直接回答。
3）若无结果：只回复“未查询到相关数据”。
4）若属于错误或不支持：根据系统判断信息翻译成客户易识别的中文。

格式规则：
- 输出 Markdown 正文，不要代码块。
- 仅输出客户可见内容，不要标题“结果/说明”等包装词。
- 有多条结果时可用 Markdown 列表；单条结果时用一行文本。

禁止项：
- 不要礼貌寒暄
- 不要解释规则
- 不要补充建议
- 不要承诺具体时间
- 不要添加任何额外信息
- 不要出现“根据接口返回/接口显示/code/showMsg/系统层面/客户无需关注细节”等描述
`, userQuestion, reason)
	return s.chatCompletion(ctx, prompt)
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type openAIChatReq struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	Top_k       int             `json:"top_k"`
}
type openAIChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *Service) chatCompletion(ctx context.Context, userPrompt string) (string, error) {
	base := strings.TrimRight(s.cfg.LLMBaseURL, "/")
	url := base + "/chat/completions"
	body := openAIChatReq{Model: s.cfg.LLMModel, Messages: []openAIMessage{{Role: "user", Content: userPrompt}}, Temperature: 0.0, Top_k: 1}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.LLMAPIKey)
	resp, err := s.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("大模型 HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 800))
	}
	var parsed openAIChatResp
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析大模型响应失败: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", errors.New("大模型返回空内容")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func extractJSONFromLLM(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.SplitN(s, "\n", 2)
		if len(lines) > 1 {
			s = lines[1]
		}
		if idx := strings.LastIndex(s, "```"); idx > 0 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
