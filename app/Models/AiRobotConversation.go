package Models

import "time"

type AiRobotConversationTurn struct {
	MessageID      string    `json:"message_id,omitempty" bson:"message_id,omitempty"`
	QuestionType   int       `json:"question_type" bson:"question_type"`
	Question       string    `json:"question" bson:"question"`
	Resolved       string    `json:"resolved_question" bson:"resolved_question"`
	Answer         string    `json:"answer,omitempty" bson:"answer,omitempty"`
	Intent         string    `json:"intent,omitempty" bson:"intent,omitempty"`
	ProjectNumber  string    `json:"project_number,omitempty" bson:"project_number,omitempty"`
	ContractNumber string    `json:"contract_number,omitempty" bson:"contract_number,omitempty"`
	TaskUUID       string    `json:"task_uuid,omitempty" bson:"task_uuid,omitempty"`
	ArticleNameCN  string    `json:"article_name_cn,omitempty" bson:"article_name_cn,omitempty"`
	JournalName    string    `json:"journal_name,omitempty" bson:"journal_name,omitempty"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
}

type AiRobotConversation struct {
	ID string `json:"id" bson:"_id"`

	UserID         string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ConversationID string `json:"conversation_id" bson:"conversation_id"`
	Platform       string `json:"platform" bson:"platform"`
	QuestionType   int    `json:"question_type" bson:"question_type"`
	LastMessageID  string `json:"last_message_id,omitempty" bson:"last_message_id,omitempty"`
	LastIntent     string `json:"last_intent" bson:"last_intent"`
	LastQuestion   string `json:"last_question" bson:"last_question"`
	LastResolved   string `json:"last_resolved_question" bson:"last_resolved_question"`
	LastAnswer     string `json:"last_answer,omitempty" bson:"last_answer,omitempty"`
	ContextSummary string `json:"context_summary,omitempty" bson:"context_summary,omitempty"`

	LastProjectIDs     []int    `json:"last_project_ids,omitempty" bson:"last_project_ids,omitempty"`
	LastProjectNumber  string   `json:"last_project_number,omitempty" bson:"last_project_number,omitempty"`
	LastContractIDs    []int    `json:"last_contract_ids,omitempty" bson:"last_contract_ids,omitempty"`
	LastContractNumber string   `json:"last_contract_number,omitempty" bson:"last_contract_number,omitempty"`
	LastTaskUUIDs      []string `json:"last_task_uuids,omitempty" bson:"last_task_uuids,omitempty"`

	LastArticleNameCN string   `json:"last_article_name_cn,omitempty" bson:"last_article_name_cn,omitempty"`
	LastArticleNameEN string   `json:"last_article_name_en,omitempty" bson:"last_article_name_en,omitempty"`
	LastArticleLabels []string `json:"last_article_labels,omitempty" bson:"last_article_labels,omitempty"`
	LastJournalName   string   `json:"last_journal_name,omitempty" bson:"last_journal_name,omitempty"`
	LastTaskStatus    string   `json:"last_task_status,omitempty" bson:"last_task_status,omitempty"`
	LastFailureReason string   `json:"last_failure_reason,omitempty" bson:"last_failure_reason,omitempty"`
	LastDownloadKind  string   `json:"last_download_kind,omitempty" bson:"last_download_kind,omitempty"`
	LastHasLink       bool     `json:"last_has_link,omitempty" bson:"last_has_link,omitempty"`

	RecentTurns []AiRobotConversationTurn `json:"recent_turns,omitempty" bson:"recent_turns,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at" bson:"updated_at"`
}

func (AiRobotConversation) CollectionName() string {
	return "ai_robot_conversations"
}
