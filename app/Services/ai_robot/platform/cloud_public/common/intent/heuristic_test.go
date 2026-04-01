package intent

import "testing"

func TestHeuristicProjectIntent(t *testing.T) {
	cases := []struct {
		q      string
		intent string
	}{
		{"帮我下载这个项目的报告包", IntentDownloadFinal},
		{"这个合同下面有哪些项目", IntentContractProjects},
		{"我的项目快过期了吗", IntentProjectExpiring},
	}
	for _, tc := range cases {
		got := heuristicProjectIntent(tc.q)
		if got == nil || got.Intent != tc.intent {
			t.Fatalf("question=%q got=%v want=%s", tc.q, got, tc.intent)
		}
	}
}

func TestHeuristicTaskIntent(t *testing.T) {
	cases := []struct {
		q      string
		intent string
	}{
		{"这个工单跑完没", IntentTaskStatus},
		{"帮我导出这个工单的结果包", IntentTaskDownloadResult},
	}
	for _, tc := range cases {
		got := heuristicTaskIntent(tc.q)
		if got == nil || got.Intent != tc.intent {
			t.Fatalf("question=%q got=%v want=%s", tc.q, got, tc.intent)
		}
	}
}

func TestHeuristicProjectArticleIntent(t *testing.T) {
	got := heuristicProjectArticleIntent("帮我查一下植物代谢组相关文献")
	if got == nil || got.Intent != IntentProjectArticleList {
		t.Fatalf("unexpected plan: %+v", got)
	}
}
