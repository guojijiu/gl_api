package conversation

import (
	"strings"
	"time"

	"cloud-platform-api/app/Services/ai_robot/internal/timeutil"
)

const maxRecentTurns = 6
const defaultSummaryMaxTurns = 3

type Turn struct {
	MessageID      string    `json:"message_id,omitempty"`
	QuestionType   int       `json:"question_type"`
	Question       string    `json:"question"`
	Resolved       string    `json:"resolved_question"`
	Answer         string    `json:"answer,omitempty"`
	Intent         string    `json:"intent,omitempty"`
	ProjectNumber  string    `json:"project_number,omitempty"`
	ContractNumber string    `json:"contract_number,omitempty"`
	TaskUUID       string    `json:"task_uuid,omitempty"`
	ArticleNameCN  string    `json:"article_name_cn,omitempty"`
	JournalName    string    `json:"journal_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type State struct {
	UserID         string `json:"user_id,omitempty"`
	ConversationID string `json:"conversation_id"`
	Platform       string `json:"platform"`
	QuestionType   int    `json:"question_type"`
	LastMessageID  string `json:"last_message_id,omitempty"`
	LastIntent     string `json:"last_intent"`
	LastQuestion   string `json:"last_question"`
	LastResolved   string `json:"last_resolved_question"`
	LastAnswer     string `json:"last_answer,omitempty"`
	ContextSummary string `json:"context_summary,omitempty"`

	LastProjectIDs     []int    `json:"last_project_ids,omitempty"`
	LastProjectNumber  string   `json:"last_project_number,omitempty"`
	LastContractIDs    []int    `json:"last_contract_ids,omitempty"`
	LastContractNumber string   `json:"last_contract_number,omitempty"`
	LastTaskUUIDs      []string `json:"last_task_uuids,omitempty"`

	LastArticleNameCN string   `json:"last_article_name_cn,omitempty"`
	LastArticleNameEN string   `json:"last_article_name_en,omitempty"`
	LastArticleLabels []string `json:"last_article_labels,omitempty"`
	LastJournalName   string   `json:"last_journal_name,omitempty"`
	LastTaskStatus    string   `json:"last_task_status,omitempty"`
	LastFailureReason string   `json:"last_failure_reason,omitempty"`
	LastDownloadKind  string   `json:"last_download_kind,omitempty"`
	LastHasLink       bool     `json:"last_has_link,omitempty"`

	RecentTurns []Turn    `json:"recent_turns,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *State) Touch() {
	s.UpdatedAt = timeutil.NowInShanghai()
}

func (s *State) RecordTurn(messageID string) {
	if s == nil {
		return
	}
	if s.LastQuestion == "" && s.LastResolved == "" {
		return
	}
	if messageID != "" && s.LastMessageID == messageID && len(s.RecentTurns) > 0 {
		return
	}

	turn := Turn{
		MessageID:      messageID,
		QuestionType:   s.QuestionType,
		Question:       s.LastQuestion,
		Resolved:       s.LastResolved,
		Answer:         s.LastAnswer,
		Intent:         s.LastIntent,
		ProjectNumber:  s.LastProjectNumber,
		ContractNumber: s.LastContractNumber,
		ArticleNameCN:  s.LastArticleNameCN,
		JournalName:    s.LastJournalName,
		CreatedAt:      timeutil.NowInShanghai(),
	}
	if len(s.LastTaskUUIDs) > 0 {
		turn.TaskUUID = s.LastTaskUUIDs[0]
	}

	s.RecentTurns = append(s.RecentTurns, turn)
	if len(s.RecentTurns) > maxRecentTurns {
		s.RecentTurns = append([]Turn(nil), s.RecentTurns[len(s.RecentTurns)-maxRecentTurns:]...)
	}
	s.LastMessageID = messageID
}

func (s *State) RefreshSummary() {
	s.RefreshSummaryWithLimits(defaultSummaryMaxTurns)
}

func (s *State) RefreshSummaryWithLimits(maxTurns int) {
	if s == nil {
		return
	}
	if len(s.RecentTurns) == 0 {
		s.ContextSummary = ""
		return
	}
	filtered := recentTurnsForCurrentDomain(s.RecentTurns, s.QuestionType)
	if maxTurns <= 0 {
		maxTurns = defaultSummaryMaxTurns
	}
	start := len(filtered) - maxTurns
	if start < 0 {
		start = 0
	}
	turns := filtered[start:]
	parts := make([]string, 0, len(turns)+1)
	for i, turn := range turns {
		parts = append(parts, buildTurnSummary(len(turns)-i, turn))
	}
	if current := buildCurrentFocusSummary(s); current != "" {
		parts = append(parts, current)
	}
	s.ContextSummary = strings.Join(parts, " | ")
}

func recentTurnsForCurrentDomain(turns []Turn, questionType int) []Turn {
	if len(turns) == 0 {
		return nil
	}
	if questionType == 0 {
		return append([]Turn(nil), turns...)
	}
	var out []Turn
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].QuestionType != questionType {
			if len(out) > 0 {
				break
			}
			continue
		}
		out = append(out, turns[i])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func latestTurnTime(turns []Turn) time.Time {
	for i := len(turns) - 1; i >= 0; i-- {
		if !turns[i].CreatedAt.IsZero() {
			return turns[i].CreatedAt
		}
	}
	return time.Time{}
}

func isAutoFillFresh(s *State, maxAge time.Duration, maxTurns int) bool {
	if s == nil {
		return false
	}
	if len(s.RecentTurns) == 0 {
		return true
	}
	turns := recentTurnsForCurrentDomain(s.RecentTurns, s.QuestionType)
	if maxTurns > 0 && len(turns) > maxTurns {
		turns = turns[len(turns)-maxTurns:]
	}
	lastAt := latestTurnTime(turns)
	if lastAt.IsZero() {
		return true
	}
	if maxAge <= 0 {
		maxAge = 10 * time.Minute
	}
	return time.Since(lastAt) <= maxAge
}

func buildTurnSummary(order int, turn Turn) string {
	label := "最近第" + itoa(order) + "轮"
	fields := []string{label}
	if turn.Intent != "" {
		fields = append(fields, "意图="+turn.Intent)
	}
	switch {
	case turn.ProjectNumber != "":
		fields = append(fields, "项目="+turn.ProjectNumber)
	case turn.ContractNumber != "":
		fields = append(fields, "合同="+turn.ContractNumber)
	case turn.TaskUUID != "":
		fields = append(fields, "任务="+turn.TaskUUID)
	case turn.ArticleNameCN != "":
		fields = append(fields, "文章="+turn.ArticleNameCN)
	case turn.JournalName != "":
		fields = append(fields, "期刊="+turn.JournalName)
	}
	if turn.Question != "" {
		fields = append(fields, "问题="+truncateTurnText(turn.Question, 20))
	}
	if turn.Answer != "" {
		fields = append(fields, "回答="+truncateTurnText(turn.Answer, 24))
	}
	return strings.Join(fields, " ")
}

func buildCurrentFocusSummary(s *State) string {
	fields := []string{"当前关注"}
	switch {
	case s.LastProjectNumber != "":
		fields = append(fields, "项目="+s.LastProjectNumber)
	case s.LastContractNumber != "":
		fields = append(fields, "合同="+s.LastContractNumber)
	case len(s.LastTaskUUIDs) > 0:
		fields = append(fields, "任务="+s.LastTaskUUIDs[0])
	case s.LastArticleNameCN != "":
		fields = append(fields, "文章="+s.LastArticleNameCN)
	case s.LastJournalName != "":
		fields = append(fields, "期刊="+s.LastJournalName)
	default:
		return ""
	}
	if s.LastIntent != "" {
		fields = append(fields, "意图="+s.LastIntent)
	}
	if s.LastTaskStatus != "" {
		fields = append(fields, "状态="+s.LastTaskStatus)
	}
	if s.LastFailureReason != "" {
		fields = append(fields, "失败原因="+truncateTurnText(s.LastFailureReason, 18))
	}
	if s.LastDownloadKind != "" {
		fields = append(fields, "下载="+s.LastDownloadKind)
	}
	if s.LastHasLink {
		fields = append(fields, "已命中链接=true")
	}
	return strings.Join(fields, " ")
}

func truncateTurnText(text string, maxRunes int) string {
	rs := []rune(strings.TrimSpace(text))
	if len(rs) <= maxRunes {
		return string(rs)
	}
	return string(rs[:maxRunes]) + "..."
}

func itoa(v int) string {
	if v == 1 {
		return "1"
	}
	if v == 2 {
		return "2"
	}
	if v == 3 {
		return "3"
	}
	return "0"
}
