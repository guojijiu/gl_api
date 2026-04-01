package Requests

import "testing"

func TestValidate_UnsupportedPlatform(t *testing.T) {
	req := AiRobotChatRequest{
		Question:     "test",
		Platform:     "unknown",
		QuestionType: AiQuestionTypeProject,
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("未知平台应校验失败")
	}
}

func TestValidate_ImageCompareRejectTaskType(t *testing.T) {
	req := AiRobotChatRequest{
		Question:     "test",
		Platform:     AiPlatformImageCompare,
		QuestionType: AiQuestionTypeTask,
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("image_compare 不应支持 task 类型")
	}
}

func TestValidate_CloudPublicSupportsTaskType(t *testing.T) {
	req := AiRobotChatRequest{
		Question:     "test",
		UserID:       "user-1",
		Platform:     AiPlatformCloudPublic,
		QuestionType: AiQuestionTypeTask,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("cloud_public 应支持 task 类型, err=%v", err)
	}
}

func TestConversationManageValidate_Success(t *testing.T) {
	req := AiRobotConversationManageRequest{
		UserID:         "user-1",
		Platform:       AiPlatformCloudPublic,
		ConversationID: "conv-001",
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("会话管理请求应校验成功, err=%v", err)
	}
}

func TestConversationManageValidate_EmptyConversationID(t *testing.T) {
	req := AiRobotConversationManageRequest{
		UserID:         "user-1",
		Platform:       AiPlatformCloudPublic,
		ConversationID: "",
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("conversation_id 为空应校验失败")
	}
}

func TestConversationListValidate_Normalize(t *testing.T) {
	req := AiRobotConversationListRequest{
		UserID:         " user-1 ",
		Platform:       AiPlatformCloudPublic,
		ConversationID: " conv ",
		Page:           0,
		Limit:          999,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("列表请求应校验成功, err=%v", err)
	}
	if req.Page != 1 || req.Limit != 100 || req.ConversationID != "conv" || req.UserID != "user-1" {
		t.Fatalf("normalize failed: %+v", req)
	}
}

func TestConversationListValidate_UnsupportedPlatform(t *testing.T) {
	req := AiRobotConversationListRequest{
		Platform: "unknown",
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("未知平台应校验失败")
	}
}

func TestConversationListValidate_QuestionTypeRequiresPlatform(t *testing.T) {
	req := AiRobotConversationListRequest{
		QuestionType: AiQuestionTypeTask,
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("指定 question_type 但未传 platform 应校验失败")
	}
}

func TestConversationListValidate_TimeRange(t *testing.T) {
	req := AiRobotConversationListRequest{
		Platform:  AiPlatformCloudPublic,
		StartTime: "2026-04-02",
		EndTime:   "2026-04-01",
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("开始时间晚于结束时间应校验失败")
	}
}

func TestConversationCleanupValidate_RequiresFilter(t *testing.T) {
	req := AiRobotConversationCleanupRequest{}
	if err := req.Validate(); err == nil {
		t.Fatalf("批量删除未传筛选条件应校验失败")
	}
}

func TestConversationCleanupValidate_Success(t *testing.T) {
	req := AiRobotConversationCleanupRequest{
		UserID:    "user-1",
		Platform:  AiPlatformCloudPublic,
		StartTime: "2026-04-01",
		EndTime:   "2026-04-02",
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("批量删除请求应校验成功, err=%v", err)
	}
}

func TestChatValidate_ContextRequiresUserID(t *testing.T) {
	req := AiRobotChatRequest{
		Question:       "test",
		Platform:       AiPlatformCloudPublic,
		QuestionType:   AiQuestionTypeTask,
		ConversationID: "conv-001",
		EnableContext:  true,
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("enable_context=true 且缺少 user_id 应校验失败")
	}
}

func TestChatValidate_UserIDWithoutContextIsAllowed(t *testing.T) {
	req := AiRobotChatRequest{
		Question:     "test",
		UserID:       "user-1",
		Platform:     AiPlatformCloudPublic,
		QuestionType: AiQuestionTypeTask,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("传 user_id 但未显式开启上下文时，请求校验应通过, err=%v", err)
	}
}

func TestConversationManageValidate_EmptyUserID(t *testing.T) {
	req := AiRobotConversationManageRequest{
		UserID:         "",
		Platform:       AiPlatformCloudPublic,
		ConversationID: "conv-001",
	}
	if err := req.Validate(); err == nil {
		t.Fatalf("缺少 user_id 应校验失败")
	}
}
