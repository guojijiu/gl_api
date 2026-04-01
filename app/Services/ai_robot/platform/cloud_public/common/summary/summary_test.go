package summary

import (
	"context"
	"strings"
	"testing"
)

type mockChatCompletionClient struct {
	lastPrompt string
	reply      string
}

func (m *mockChatCompletionClient) ChatCompletion(_ context.Context, userPrompt string) (string, error) {
	m.lastPrompt = userPrompt
	return m.reply, nil
}

func TestSummarizeProjectData_IncludesContextSummaryInPrompt(t *testing.T) {
	mock := &mockChatCompletionClient{
		reply: "ok",
	}

	_, err := SummarizeProjectData(context.Background(), mock, "原始数据呢", "最近一轮意图=download_final_report 项目=MWXS-25-10500-a", []byte(`{"code":1,"content":{"data":[]}}`))
	if err != nil {
		t.Fatalf("summarize project data failed: %v", err)
	}
	if !strings.Contains(mock.lastPrompt, "最近一轮意图=download_final_report 项目=MWXS-25-10500-a") {
		t.Fatalf("expected context summary in prompt: %s", mock.lastPrompt)
	}
}
