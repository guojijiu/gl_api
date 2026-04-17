package Models

import "time"

type AiRobotConversationMessage struct {
	ID string `json:"id" bson:"_id"`

	UserID         string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ConversationID string    `json:"conversation_id" bson:"conversation_id"`
	Platform       string    `json:"platform" bson:"platform"`
	QuestionType   int       `json:"question_type" bson:"question_type"`
	MessageID      string    `json:"message_id" bson:"message_id"`
	Question       string    `json:"question" bson:"question"`
	Resolved       string    `json:"resolved_question,omitempty" bson:"resolved_question,omitempty"`
	Answer         string    `json:"answer,omitempty" bson:"answer,omitempty"`
	Intent         string    `json:"intent,omitempty" bson:"intent,omitempty"`
	Success        bool      `json:"success" bson:"success"`
	ShowMsg        string    `json:"show_msg,omitempty" bson:"show_msg,omitempty"`
	DebugMsg       string    `json:"debug_msg,omitempty" bson:"debug_msg,omitempty"`
	ResultKind     string    `json:"result_kind,omitempty" bson:"result_kind,omitempty"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
}

func (AiRobotConversationMessage) CollectionName() string {
	return "ai_robot_conversation_messages"
}
