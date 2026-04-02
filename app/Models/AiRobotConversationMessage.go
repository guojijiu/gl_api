package Models

import "time"

type AiRobotConversationMessage struct {
	ID string `json:"id" bson:"_id"`

	UserID                string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ConversationID        string    `json:"conversation_id" bson:"conversation_id"`
	Platform              string    `json:"platform" bson:"platform"`
	QuestionType          int       `json:"question_type" bson:"question_type"`
	MessageID             string    `json:"message_id" bson:"message_id"`
	Question              string    `json:"question" bson:"question"`
	Resolved              string    `json:"resolved_question,omitempty" bson:"resolved_question,omitempty"`
	NormalizedQuestion    string    `json:"normalized_question,omitempty" bson:"normalized_question,omitempty"`
	HitContext            bool      `json:"hit_context" bson:"hit_context"`
	ContextSource         string    `json:"context_source,omitempty" bson:"context_source,omitempty"`
	ClarificationNeeded   bool      `json:"clarification_needed" bson:"clarification_needed"`
	ClarificationReason   string    `json:"clarification_reason,omitempty" bson:"clarification_reason,omitempty"`
	Answer                string    `json:"answer,omitempty" bson:"answer,omitempty"`
	Intent                string    `json:"intent,omitempty" bson:"intent,omitempty"`
	Success               bool      `json:"success" bson:"success"`
	ShowMsg               string    `json:"show_msg,omitempty" bson:"show_msg,omitempty"`
	DebugMsg              string    `json:"debug_msg,omitempty" bson:"debug_msg,omitempty"`
	Stage                 string    `json:"stage,omitempty" bson:"stage,omitempty"`
	ErrorType             string    `json:"error_type,omitempty" bson:"error_type,omitempty"`
	CloudAPI              string    `json:"cloud_api,omitempty" bson:"cloud_api,omitempty"`
	CloudStatusCode       int       `json:"cloud_status_code,omitempty" bson:"cloud_status_code,omitempty"`
	LLMModel              string    `json:"llm_model,omitempty" bson:"llm_model,omitempty"`
	RequestPayloadSummary string    `json:"request_payload_summary,omitempty" bson:"request_payload_summary,omitempty"`
	ResultKind            string    `json:"result_kind,omitempty" bson:"result_kind,omitempty"`
	ResultCount           int       `json:"result_count,omitempty" bson:"result_count,omitempty"`
	ResultBrief           string    `json:"result_brief,omitempty" bson:"result_brief,omitempty"`
	ResponseSummary       string    `json:"response_summary,omitempty" bson:"response_summary,omitempty"`
	ResponseSize          int64     `json:"response_size,omitempty" bson:"response_size,omitempty"`
	CloudCalled           bool      `json:"cloud_called" bson:"cloud_called"`
	DurationMS            int64     `json:"duration_ms,omitempty" bson:"duration_ms,omitempty"`
	LLMDurationMS         int64     `json:"llm_duration_ms,omitempty" bson:"llm_duration_ms,omitempty"`
	CloudDurationMS       int64     `json:"cloud_duration_ms,omitempty" bson:"cloud_duration_ms,omitempty"`
	CreatedAt             time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" bson:"updated_at"`
}

func (AiRobotConversationMessage) CollectionName() string {
	return "ai_robot_conversation_messages"
}
