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
		Platform:     AiPlatformCloudPublic,
		QuestionType: AiQuestionTypeTask,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("cloud_public 应支持 task 类型, err=%v", err)
	}
}
