package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cloud-platform-api/app/Config"
)

// Client 百炼等大模型 HTTP 调用（不含云平台业务编排）。
type Client struct {
	cfg *Config.AiGatewayConfig
	hc  *http.Client
}

// ChatCompletionClient 只暴露 ai_robot 编排所需的最小能力。
// ai_robot 依赖接口而不是依赖具体 *Client 实现，以降低耦合并方便后续扩展模型供应商。
type ChatCompletionClient interface {
	ChatCompletion(ctx context.Context, userPrompt string) (string, error)
}

func NewClient(cfg *Config.AiGatewayConfig, hc *http.Client) *Client {
	return &Client{cfg: cfg, hc: hc}
}

func (c *Client) ValidateForLLM() error {
	if c.cfg.LLMAPIKey == "" {
		return errors.New("未配置 AI_GATEWAY_LLM_API_KEY（百炼 API Key），无法调用大模型")
	}
	return nil
}

func (c *Client) ChatCompletion(ctx context.Context, userPrompt string) (string, error) {
	if err := c.ValidateForLLM(); err != nil {
		return "", err
	}
	base := strings.TrimRight(c.cfg.LLMBaseURL, "/")
	url := base + "/chat/completions"
	body := openAIChatReq{Model: c.cfg.LLMModel, Messages: []openAIMessage{{Role: "user", Content: userPrompt}}, Temperature: 0.0, Top_k: 1}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.LLMAPIKey)
	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("大模型 HTTP %d: %s", resp.StatusCode, Truncate(string(respBody), 800))
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

// ExtractJSONFromLLM 从模型返回中剥离 markdown 代码块等，得到纯 JSON 文本。
func ExtractJSONFromLLM(s string) string {
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

// Truncate 截断长文本用于日志/错误信息。
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
