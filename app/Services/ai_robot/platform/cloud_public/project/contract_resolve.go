package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
)

func ResolveContractIDsFromContracts(_ context.Context, cfg *Config.AiGatewayConfig, userQuestion string, contractsJSON []byte) ([]int, string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(contractsJSON, &root); err != nil {
		return nil, "", fmt.Errorf("解析合同列表失败: %w", err)
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return nil, "", errors.New("未查询到可用合同")
	}
	candidates := cloudclient.ExtractContractNumberCandidates(userQuestion)
	if len(candidates) > 0 {
		for _, c := range candidates {
			ids := cloudclient.CollectIDsByContractNumber(data, c)
			if len(ids) > 0 {
				return ids, c, nil
			}
		}
		if cfg != nil && cfg.StrictMode {
			return nil, "", errors.New("严格模式已开启：未按合同编号匹配到数据，请确认合同编号后重试")
		}
	} else if cfg != nil && cfg.StrictMode {
		return nil, "", errors.New("严格模式已开启：请在问题中明确提供合同编号")
	}
	if len(data) == 1 {
		if m, ok := data[0].(map[string]interface{}); ok {
			if id, ok := cloudclient.ToInt(m["id"]); ok && id > 0 {
				return []int{id}, "", nil
			}
		}
	}
	return nil, "", errors.New("无法从问题中匹配合同，请补充合同编号")
}
