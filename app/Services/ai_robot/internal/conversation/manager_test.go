package conversation

import (
	"context"
	"strings"
	"testing"
	"time"

	"cloud-platform-api/app/Http/Requests"
)

func TestEnhanceQuestion_ProjectFallback(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "这个项目的结题报告下载链接",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-1",
		EnableContext:  true,
	}
	state := &State{LastProjectNumber: "MWXS-25-10500-a"}

	got := EnhanceQuestion(req, state)
	if got != "这个项目的结题报告下载链接 项目编号 MWXS-25-10500-a" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_TaskFallback(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "这个工单跑完没",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeTask,
		ConversationID: "conv-2",
		EnableContext:  true,
	}
	state := &State{LastTaskUUIDs: []string{"86ce0f8b-ee05-47af-9495-6a48addd8a39"}}

	got := EnhanceQuestion(req, state)
	if got != "这个工单跑完没 任务编号 86ce0f8b-ee05-47af-9495-6a48addd8a39" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_ShortFollowUpUsesLastProject(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "下载链接呢",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-3",
		EnableContext:  true,
	}
	state := &State{LastProjectNumber: "MWXS-25-10500-a"}

	got := EnhanceQuestion(req, state)
	if got != "下载链接呢 项目编号 MWXS-25-10500-a" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_AmbiguousShortQuestionDoesNotAutoFill(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "帮我看看",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-3b",
		EnableContext:  true,
	}
	state := &State{LastProjectNumber: "MWXS-25-10500-a"}

	got := EnhanceQuestion(req, state)
	if got != "帮我看看" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestSaveState_RecordsRecentTurn(t *testing.T) {
	store := NewMemoryStore()
	req := &Requests.AiRobotChatRequest{
		Question:         "这个项目的结题报告下载链接",
		ResolvedQuestion: "这个项目的结题报告下载链接 项目编号 MWXS-25-10500-a",
		Platform:         Requests.AiPlatformCloudPublic,
		QuestionType:     Requests.AiQuestionTypeProject,
		UserID:           "user-1",
		ConversationID:   "conv-4",
		MessageID:        "msg-1",
		EnableContext:    true,
	}
	state := &State{
		LastIntent:        "download_final_report",
		LastQuestion:      req.Question,
		LastResolved:      req.ResolvedQuestion,
		LastAnswer:        "这是本轮回答",
		LastProjectNumber: "MWXS-25-10500-a",
	}

	if err := SaveState(context.Background(), store, req, state); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	got, err := store.Get(context.Background(), BuildStoreKey(req.UserID, req.Platform, req.ConversationID))
	if err != nil {
		t.Fatalf("load state failed: %v", err)
	}
	if len(got.RecentTurns) != 1 {
		t.Fatalf("unexpected recent turns count: %d", len(got.RecentTurns))
	}
	if got.RecentTurns[0].ProjectNumber != "MWXS-25-10500-a" {
		t.Fatalf("unexpected project number: %+v", got.RecentTurns[0])
	}
	if got.RecentTurns[0].Answer != "这是本轮回答" {
		t.Fatalf("unexpected answer: %+v", got.RecentTurns[0])
	}
	if got.LastAnswer != "这是本轮回答" {
		t.Fatalf("unexpected last answer: %+v", got)
	}
	if got.ContextSummary == "" {
		t.Fatalf("expected context summary to be generated")
	}
}

func TestRefreshSummary_UsesRecentThreeTurns(t *testing.T) {
	state := &State{
		LastIntent:        "download_original_data",
		LastProjectNumber: "MWXS-25-10500-a",
		LastDownloadKind:  "original_data",
		LastHasLink:       true,
		QuestionType:      Requests.AiQuestionTypeProject,
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeTask, Intent: "task_status", Question: "这个工单跑完没", TaskUUID: "u-1"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_final_report", Question: "下载这个项目的结题报告", ProjectNumber: "P-002"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_original_data", Question: "原始数据呢", ProjectNumber: "P-002"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_original_data", Question: "授权码呢", ProjectNumber: "MWXS-25-10500-a"},
		},
	}

	state.RefreshSummary()
	if strings.Contains(state.ContextSummary, "u-1") {
		t.Fatalf("summary should ignore other domains: %s", state.ContextSummary)
	}
	if !strings.Contains(state.ContextSummary, "MWXS-25-10500-a") {
		t.Fatalf("summary should include current focus: %s", state.ContextSummary)
	}
	if !strings.Contains(state.ContextSummary, "最近第1轮") {
		t.Fatalf("summary should include turn labels: %s", state.ContextSummary)
	}
	if !strings.Contains(state.ContextSummary, "下载=original_data") || !strings.Contains(state.ContextSummary, "已命中链接=true") {
		t.Fatalf("summary should include download result markers: %s", state.ContextSummary)
	}
}

func TestRefreshSummaryWithLimits_CanUseLongerWindow(t *testing.T) {
	state := &State{
		QuestionType: Requests.AiQuestionTypeProject,
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "project_list", Question: "我的项目有哪些", ProjectNumber: "P-001"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_final_report", Question: "下载报告", ProjectNumber: "P-002"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_original_data", Question: "原始数据呢", ProjectNumber: "P-003"},
			{QuestionType: Requests.AiQuestionTypeProject, Intent: "download_original_data", Question: "授权码呢", ProjectNumber: "P-004"},
		},
		LastProjectNumber: "P-004",
	}

	state.RefreshSummaryWithLimits(4)
	if !strings.Contains(state.ContextSummary, "P-001") {
		t.Fatalf("expected longer summary window to include older turn: %s", state.ContextSummary)
	}
}

func TestBuildStoreKey_IncludesUserID(t *testing.T) {
	got := BuildStoreKey("user-string-id", "cloud_public", "conv-001")
	if got != "user-string-id:cloud_public:conv-001" {
		t.Fatalf("unexpected store key: %s", got)
	}
}

func TestRecordTurn_TrimsRecentTurns(t *testing.T) {
	state := &State{}
	for i := 0; i < maxRecentTurns+2; i++ {
		state.LastQuestion = "q"
		state.LastResolved = "r"
		state.RecordTurn("")
	}
	if len(state.RecentTurns) != maxRecentTurns {
		t.Fatalf("unexpected recent turns len: %d", len(state.RecentTurns))
	}
}

func TestEnhanceQuestion_ProjectIntentContinuation(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "原始数据呢",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-5",
		EnableContext:  true,
	}
	state := &State{
		LastIntent:        "download_final_report",
		LastProjectNumber: "MWXS-25-10500-a",
	}

	got := EnhanceQuestion(req, state)
	if got != "原始数据呢 项目编号 MWXS-25-10500-a" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_TaskIntentContinuation(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "结果呢",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeTask,
		ConversationID: "conv-6",
		EnableContext:  true,
	}
	state := &State{
		LastIntent:    "task_status",
		LastTaskUUIDs: []string{"86ce0f8b-ee05-47af-9495-6a48addd8a39"},
	}

	got := EnhanceQuestion(req, state)
	if got != "结果呢 任务编号 86ce0f8b-ee05-47af-9495-6a48addd8a39" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_WeakArticleFollowUpUsesContext(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "详细点",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProjectArticle,
		ConversationID: "conv-6b",
		EnableContext:  true,
	}
	state := &State{
		LastArticleNameCN: "水稻转录组",
	}

	got := EnhanceQuestion(req, state)
	if got != "详细点 中文名称 水稻转录组" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestRefreshSummary_IncludesTaskStatusAndFailureReason(t *testing.T) {
	state := &State{
		QuestionType:      Requests.AiQuestionTypeTask,
		LastIntent:        "task_status",
		LastTaskUUIDs:     []string{"u-1"},
		LastTaskStatus:    "failed",
		LastFailureReason: "参数校验失败",
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeTask, Intent: "task_status", Question: "这个工单为什么失败", TaskUUID: "u-1"},
		},
	}

	state.RefreshSummary()
	if !strings.Contains(state.ContextSummary, "状态=failed") || !strings.Contains(state.ContextSummary, "失败原因=参数校验失败") {
		t.Fatalf("summary should include task status markers: %s", state.ContextSummary)
	}
}

func TestEnhanceQuestion_WhyFailedUsesRecentTask(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "为什么失败",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeTask,
		ConversationID: "conv-9",
		EnableContext:  true,
	}
	state := &State{
		LastTaskUUIDs:     []string{"86ce0f8b-ee05-47af-9495-6a48addd8a39"},
		LastTaskStatus:    "failed",
		LastFailureReason: "参数校验失败",
	}

	got := EnhanceQuestion(req, state)
	if got != "为什么失败 任务编号 86ce0f8b-ee05-47af-9495-6a48addd8a39" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_WhichLinkUsesRecentDownloadKind(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "哪个链接",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-10",
		EnableContext:  true,
	}
	state := &State{
		LastProjectNumber: "MWXS-25-10500-a",
		LastDownloadKind:  "original_data",
		LastHasLink:       true,
	}

	got := EnhanceQuestion(req, state)
	if got != "哪个链接 项目编号 MWXS-25-10500-a 下载类型 original_data" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_MoreResultsUsesRecentArticle(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "还有吗",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProjectArticle,
		ConversationID: "conv-11",
		EnableContext:  true,
	}
	state := &State{
		LastArticleNameCN: "水稻转录组",
	}

	got := EnhanceQuestion(req, state)
	if got != "还有吗 中文名称 水稻转录组" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_DoesNotReuseTaskContextForProject(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "下载链接呢",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-7",
		EnableContext:  true,
	}
	state := &State{
		QuestionType:      Requests.AiQuestionTypeTask,
		LastTaskUUIDs:     []string{"86ce0f8b-ee05-47af-9495-6a48addd8a39"},
		LastProjectNumber: "MWXS-25-10500-a",
	}

	got := EnhanceQuestion(req, state)
	if got != "下载链接呢" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_DoesNotReuseProjectContextForArticle(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "查文章",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProjectArticle,
		ConversationID: "conv-8",
		EnableContext:  true,
	}
	state := &State{
		QuestionType:      Requests.AiQuestionTypeProject,
		LastProjectNumber: "MWXS-25-10500-a",
		LastIntent:        "download_final_report",
	}

	got := EnhanceQuestion(req, state)
	if got != "查文章" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_FreshContextStillAutofills(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "哪个链接",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-12",
		EnableContext:  true,
	}
	state := &State{
		QuestionType:      Requests.AiQuestionTypeProject,
		LastProjectNumber: "MWXS-25-10500-a",
		LastDownloadKind:  "original_data",
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeProject, ProjectNumber: "MWXS-25-10500-a", CreatedAt: time.Now().Add(-2 * time.Minute)},
		},
	}

	got := EnhanceQuestion(req, state)
	if got != "哪个链接 项目编号 MWXS-25-10500-a 下载类型 original_data" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestEnhanceQuestion_StaleContextDoesNotAutofill(t *testing.T) {
	req := &Requests.AiRobotChatRequest{
		Question:       "哪个链接",
		Platform:       Requests.AiPlatformCloudPublic,
		QuestionType:   Requests.AiQuestionTypeProject,
		ConversationID: "conv-13",
		EnableContext:  true,
	}
	state := &State{
		QuestionType:      Requests.AiQuestionTypeProject,
		LastProjectNumber: "MWXS-25-10500-a",
		LastDownloadKind:  "original_data",
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeProject, ProjectNumber: "MWXS-25-10500-a", CreatedAt: time.Now().Add(-30 * time.Minute)},
		},
	}

	got := EnhanceQuestion(req, state)
	if got != "哪个链接" {
		t.Fatalf("unexpected enhanced question: %q", got)
	}
}

func TestIsAutoFillFresh_RespectsAge(t *testing.T) {
	state := &State{
		QuestionType: Requests.AiQuestionTypeProject,
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeProject, CreatedAt: time.Now().Add(-11 * time.Minute)},
		},
	}
	if isAutoFillFresh(state, 10*time.Minute, 3) {
		t.Fatalf("expected stale context to be rejected")
	}
}

func TestIsAutoFillFresh_RespectsTurnWindow(t *testing.T) {
	now := time.Now()
	state := &State{
		QuestionType: Requests.AiQuestionTypeProject,
		RecentTurns: []Turn{
			{QuestionType: Requests.AiQuestionTypeProject, CreatedAt: now.Add(-30 * time.Minute)},
			{QuestionType: Requests.AiQuestionTypeProject, CreatedAt: now.Add(-20 * time.Minute)},
			{QuestionType: Requests.AiQuestionTypeProject, CreatedAt: now.Add(-2 * time.Minute)},
		},
	}
	if !isAutoFillFresh(state, 10*time.Minute, 1) {
		t.Fatalf("expected recent turn inside turn window to keep context fresh")
	}
}
