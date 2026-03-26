package aigateway

import (
	"context"
	"testing"

	"cloud-platform-api/app/Config"
)

func TestResolveProjectIDs_StrictModeRequiresNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	svc := NewService(cfg)
	projectsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 11, "number": "P-001"},
				{"id": 12, "number": "P-002"}
			]
		}
	}`)

	_, _, err := svc.ResolveProjectIDsFromProjects(context.Background(), "下载结题报告", projectsJSON)
	if err == nil {
		t.Fatalf("strict_mode=true 时，未提供项目编号应返回错误")
	}
}

func TestResolveProjectIDs_MultiMatchByNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	svc := NewService(cfg)
	projectsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 101, "number": "P-001"},
				{"id": 102, "number": "P-001"},
				{"id": 103, "number": "P-003"}
			]
		}
	}`)

	ids, matched, err := svc.ResolveProjectIDsFromProjects(context.Background(), "下载 P-001 的原始数据", projectsJSON)
	if err != nil {
		t.Fatalf("应匹配成功，但返回错误: %v", err)
	}
	if matched != "P-001" {
		t.Fatalf("matched number 错误，got=%s", matched)
	}
	if len(ids) != 2 || ids[0] != 101 || ids[1] != 102 {
		t.Fatalf("多条匹配不符合预期，got=%v", ids)
	}
}

func TestResolveContractIDs_StrictModeRequiresContractNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	svc := NewService(cfg)
	contractsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 7, "contract_number": "C-001"}
			]
		}
	}`)

	_, _, err := svc.ResolveContractIDsFromContracts(context.Background(), "我有哪些合同下的项目", contractsJSON)
	if err == nil {
		t.Fatalf("strict_mode=true 时，未提供合同编号应返回错误")
	}
}

func TestResolveContractIDs_NonStrictSingleContractFallback(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: false}
	svc := NewService(cfg)
	contractsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 9, "contract_number": "C-ONLY-ONE"}
			]
		}
	}`)

	ids, matched, err := svc.ResolveContractIDsFromContracts(context.Background(), "这个合同下有什么项目", contractsJSON)
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
