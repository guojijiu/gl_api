package Services

import "testing"

func TestAiRobotConversationService_BuildDocumentID(t *testing.T) {
	service := NewAiRobotConversationService()
	got := service.BuildDocumentID(" user-string-id ", " cloud_public ", " conv-001 ")
	if got != "user-string-id:cloud_public:conv-001" {
		t.Fatalf("unexpected document id: %s", got)
	}
}

func TestAiRobotConversationListFilter_Defaults(t *testing.T) {
	filter := AiRobotConversationListFilter{}
	if filter.Page != 0 || filter.Limit != 0 {
		t.Fatalf("unexpected initial filter: %+v", filter)
	}
}

func TestAiRobotConversationService_BuildListQueryIncludesUserID(t *testing.T) {
	service := NewAiRobotConversationService()
	query := service.buildListQuery(AiRobotConversationListFilter{
		UserID:       "user-string-id",
		Platform:     "cloud_public",
		QuestionType: 1,
	})
	if query["user_id"] != "user-string-id" {
		t.Fatalf("expected user_id in query, got: %+v", query)
	}
}
