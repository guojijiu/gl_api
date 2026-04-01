package summary

import (
	"strings"
	"testing"
)

func TestCompressProjectJSONForPrompt_List(t *testing.T) {
	apiJSON := []byte(`{
		"code":1,
		"showMsg":"ok",
		"content":{
			"total":2,
			"data":[
				{"id":1,"number":"P-001","name":"水稻项目","workflow_name_cn":"蛋白组"},
				{"id":2,"number":"P-002","name":"玉米项目","workflow_name_cn":"代谢组"}
			]
		}
	}`)

	got := compressProjectJSONForPrompt(apiJSON)
	if !strings.Contains(got, `"domain":"project"`) {
		t.Fatalf("unexpected compressed project json: %s", got)
	}
	if !strings.Contains(got, `"items_sample"`) {
		t.Fatalf("expected items_sample in compressed project json: %s", got)
	}
}

func TestCompressTaskJSONForPrompt_Status(t *testing.T) {
	apiJSON := []byte(`{
		"uuids":["u-1"],
		"status_batch":{"code":1,"content":{"data":[{"uuid":"u-1","status_value":"success"}]}},
		"details":[{"uuid":"u-1","task_id":11,"task_kind":"tool","result":{"file_path":"/tmp/a.txt"}}]
	}`)

	got := compressTaskJSONForPrompt(apiJSON)
	if !strings.Contains(got, `"domain":"task_status"`) {
		t.Fatalf("unexpected compressed task json: %s", got)
	}
	if !strings.Contains(got, `"status_batch"`) {
		t.Fatalf("expected status_batch in compressed task json: %s", got)
	}
}
