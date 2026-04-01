package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services/ai_gateway/llm"
	"cloud-platform-api/app/Services/ai_robot/internal/llmutil"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/common/parse"
)

func ResolveProjectIDsFromProjects(ctx context.Context, c llm.ChatCompletionClient, cfg *Config.AiGatewayConfig, userQuestion string, projectsJSON []byte) ([]int, string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(projectsJSON, &root); err != nil {
		return nil, "", fmt.Errorf("解析项目列表失败: %w", err)
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return nil, "", errors.New("未查询到可用项目")
	}
	candidates := cloudclient.ExtractNumberCandidates(userQuestion)
	if len(candidates) > 0 {
		for _, cand := range candidates {
			ids := cloudclient.CollectIDsByProjectNumber(data, cand)
			if len(ids) > 0 {
				return ids, cand, nil
			}
		}
		if cfg != nil && cfg.StrictMode {
			return nil, "", errors.New(parse.ProjectMatchGuidanceForQuestion(userQuestion))
		}
	} else if cfg != nil && cfg.StrictMode {
		return nil, "", errors.New(parse.ProjectMatchGuidanceForQuestion(userQuestion))
	}
	if len(data) == 1 {
		if m, ok := data[0].(map[string]interface{}); ok {
			if id, ok := cloudclient.ToInt(m["id"]); ok && id > 0 {
				return []int{id}, "", nil
			}
		}
	}
	prompt := fmt.Sprintf(`你是项目匹配器。根据用户问题，从给定项目列表中选出最可能的 project_id。
用户问题：%s
项目列表JSON：%s

只输出 JSON：
{"project_id":123,"reason":"不超过60字"}
若无法判断，project_id 输出 0。`, userQuestion, string(projectsJSON))
	text, err := c.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, "", err
	}
	raw := llmutil.ExtractJSONFromLLM(text)
	var parsed struct {
		ProjectID int    `json:"project_id"`
		Reason    string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, "", fmt.Errorf("解析 project_id 失败: %w", err)
	}
	if parsed.ProjectID <= 0 {
		return nil, "", errors.New(parse.ProjectMatchGuidanceForQuestion(userQuestion))
	}
	return []int{parsed.ProjectID}, "", nil
}
