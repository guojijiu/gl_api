package project

import (
	"context"
	"testing"

	"cloud-platform-api/app/Config"
)

func TestResolveContractIDs_StrictModeRequiresContractNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	contractsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 7, "contract_number": "C-001"}
			]
		}
	}`)

	_, _, err := ResolveContractIDsFromContracts(context.Background(), cfg, "我有哪些合同下的项目", contractsJSON)
	if err == nil {
		t.Fatalf("strict_mode=true 时，未提供合同编号应返回错误")
	}
}

func TestResolveContractIDs_NonStrictSingleContractFallback(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: false}
	contractsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 9, "contract_number": "C-ONLY-ONE"}
			]
		}
	}`)

	ids, matched, err := ResolveContractIDsFromContracts(context.Background(), cfg, "这个合同下有什么项目", contractsJSON)
	if err != nil {
		t.Fatalf("非严格模式单条应回退成功: %v", err)
	}
	if matched != "" {
		t.Fatalf("单条回退不应有 matched contract number, got=%s", matched)
	}
	if len(ids) != 1 || ids[0] != 9 {
		t.Fatalf("单条回退 id 不正确，got=%v", ids)
	}
}
