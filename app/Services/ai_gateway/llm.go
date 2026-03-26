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

func (s *Service) SummarizeWithData(ctx context.Context, userQuestion string, apiJSON []byte) (string, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return "", err
	}
	prompt := fmt.Sprintf(`你是云平台客服助手。用户问题如下：
%s

下面是云平台接口返回的 JSON（可能包含 code、showMsg、content 等字段）：
%s

请用简洁、准确、礼貌的中文回答用户，适当使用列表。若列表为空或 code 不为 1，请说明暂无数据或提示用户登录/权限。不要编造接口中不存在的数据。`, userQuestion, string(apiJSON))
	return s.chatCompletion(ctx, prompt)
}

func (s *Service) SummarizeUnsupported(ctx context.Context, userQuestion string, reason string) (string, error) {
	if err := s.validateGatewayForLLM(); err != nil {
		return "", err
	}
	prompt := fmt.Sprintf(`用户问题：%s
系统判断：%s

请用一两句礼貌、专业的中文回复：说明该能力尚在开发中或暂不支持，并引导用户稍后再试或联系人工支持。不要承诺具体时间。`, userQuestion, reason)
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
