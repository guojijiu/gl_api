package project

import (
	"context"
	"testing"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services/ai_gateway/llm"
)

func TestResolveProjectIDs_StrictModeRequiresNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	c := llm.NewClient(cfg, nil)
	projectsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 11, "number": "P-001"},
				{"id": 12, "number": "P-002"}
			]
		}
	}`)

	_, _, err := ResolveProjectIDsFromProjects(context.Background(), c, cfg, "下载结题报告", projectsJSON)
	if err == nil {
		t.Fatalf("strict_mode=true 时，未提供项目编号应返回错误")
	}
}

func TestResolveProjectIDs_MultiMatchByNumber(t *testing.T) {
	cfg := &Config.AiGatewayConfig{StrictMode: true}
	c := llm.NewClient(cfg, nil)
	projectsJSON := []byte(`{
		"content": {
			"data": [
				{"id": 101, "number": "P-001"},
				{"id": 102, "number": "P-001"},
				{"id": 103, "number": "P-003"}
			]
		}
	}`)

	ids, matched, err := ResolveProjectIDsFromProjects(context.Background(), c, cfg, "下载 P-001 的原始数据", projectsJSON)
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
