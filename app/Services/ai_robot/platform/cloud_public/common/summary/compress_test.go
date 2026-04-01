package summary

import (
	"strings"
	"testing"
)

func TestCompressAPIJSONForPrompt_CompressesLargeArray(t *testing.T) {
	raw := []byte(`{
		"code":1,
		"showMsg":"操作成功",
		"content":{
			"data":[
				{"id":1,"number":"P-001","name":"项目1"},
				{"id":2,"number":"P-002","name":"项目2"},
				{"id":3,"number":"P-003","name":"项目3"},
				{"id":4,"number":"P-004","name":"项目4"},
				{"id":5,"number":"P-005","name":"项目5"},
				{"id":6,"number":"P-006","name":"项目6"}
			],
			"total":6
		}
	}`)
	got := compressAPIJSONForPrompt(raw)
	if !strings.Contains(got, `"count":6`) {
		t.Fatalf("expected count in compressed json: %s", got)
	}
	if !strings.Contains(got, `"omitted_count":1`) {
		t.Fatalf("expected omitted_count in compressed json: %s", got)
	}
}

func TestCompressAPIJSONForPrompt_KeepsPriorityKeys(t *testing.T) {
	raw := []byte(`{"code":2,"showMsg":"账号登录已失效","content":{"data":[]}}`)
	got := compressAPIJSONForPrompt(raw)
	if !strings.Contains(got, `"code":2`) || !strings.Contains(got, `账号登录已失效`) {
		t.Fatalf("priority keys lost: %s", got)
	}
}
