package intent

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

func TestPlanTaskIntent_IncludesContextSummaryInPrompt(t *testing.T) {
	mock := &mockChatCompletionClient{
		reply: `{"intent":"unsupported","reason":"test"}`,
	}

	_, err := PlanTaskIntent(context.Background(), mock, "请结合上文帮我判断下一步", "最近一轮意图=task_status 任务=abc-123")
	if err != nil {
		t.Fatalf("plan task intent failed: %v", err)
	}
	if !strings.Contains(mock.lastPrompt, "最近一轮意图=task_status 任务=abc-123") {
		t.Fatalf("expected context summary in prompt: %s", mock.lastPrompt)
	}
}
