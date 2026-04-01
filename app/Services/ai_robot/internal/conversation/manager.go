package conversation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
)

var (
	defaultStore     Store
	defaultStoreOnce sync.Once
)

func DefaultStore() Store {
	defaultStoreOnce.Do(func() {
		defaultStore = newDefaultStore()
	})
	return defaultStore
}

func newDefaultStore() Store {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil {
		return NewMemoryStore()
	}
	if strings.EqualFold(strings.TrimSpace(cfg.ConversationStoreDriver), "memory") {
		return NewMemoryStore()
	}
	store, err := NewMongoStore(cfg)
	if err != nil {
		log.Printf("ai_robot conversation store fallback to memory: %v", err)
		return NewMemoryStore()
	}
	return store
}

func BuildStoreKey(userID, platform, conversationID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		userID = "anonymous"
	}
	// 对话上下文按用户维度隔离，避免相同 conversation_id 在不同用户间串话。
	return userID + ":" + strings.TrimSpace(platform) + ":" + strings.TrimSpace(conversationID)
}

func LoadState(ctx context.Context, store Store, req *Requests.AiRobotChatRequest) (*State, error) {
	if req == nil || !req.EnableContext || strings.TrimSpace(req.ConversationID) == "" {
		return nil, nil
	}
	key := BuildStoreKey(req.UserID, req.Platform, req.ConversationID)
	state, err := store.Get(ctx, key)
	if err != nil {
		if errorsIsNotFound(err) {
			return &State{
				UserID:         req.UserID,
				ConversationID: req.ConversationID,
				Platform:       req.Platform,
				QuestionType:   req.QuestionType,
			}, nil
		}
		return nil, err
	}
	if state != nil && strings.TrimSpace(state.UserID) != "" && strings.TrimSpace(req.UserID) != "" && strings.TrimSpace(state.UserID) != strings.TrimSpace(req.UserID) {
		return nil, fmt.Errorf("conversation user mismatch")
	}
	return state, nil
}

func SaveState(ctx context.Context, store Store, req *Requests.AiRobotChatRequest, state *State) error {
	if req == nil || !req.EnableContext || strings.TrimSpace(req.ConversationID) == "" || state == nil {
		return nil
	}
	state.UserID = strings.TrimSpace(req.UserID)
	state.ConversationID = req.ConversationID
	state.Platform = req.Platform
	state.QuestionType = req.QuestionType
	state.RecordTurn(strings.TrimSpace(req.MessageID))
	state.RefreshSummaryWithLimits(summaryMaxTurns())
	state.Touch()
	return store.Save(ctx, BuildStoreKey(req.UserID, req.Platform, req.ConversationID), state)
}

func EnhanceQuestion(req *Requests.AiRobotChatRequest, state *State) string {
	if req == nil {
		return ""
	}
	question := strings.TrimSpace(req.Question)
	if question == "" || !req.EnableContext || state == nil {
		return question
	}
	if isLikelyTopicShift(req, state, question) {
		return question
	}
	if !isAutoFillFresh(state, autoFillMaxAge(), autoFillMaxTurns()) {
		return question
	}

	question = applyResultContinuation(req, question, state)
	question = applyIntentContinuation(question, state)
	lower := strings.ToLower(question)
	switch req.QuestionType {
	case Requests.AiQuestionTypeProject:
		if projectFallbackLevel(lower, state) != fallbackNone && state.LastProjectNumber != "" && !strings.Contains(question, state.LastProjectNumber) {
			return question + " 项目编号 " + state.LastProjectNumber
		}
		if contractFallbackLevel(lower, state) == fallbackStrong && state.LastContractNumber != "" && !strings.Contains(question, state.LastContractNumber) {
			return question + " 合同编号 " + state.LastContractNumber
		}
	case Requests.AiQuestionTypeTask:
		if taskFallbackLevel(lower, state) != fallbackNone && len(state.LastTaskUUIDs) > 0 {
			if !strings.Contains(question, state.LastTaskUUIDs[0]) {
				return question + " 任务编号 " + state.LastTaskUUIDs[0]
			}
		}
	case Requests.AiQuestionTypeProjectArticle:
		if articleFallbackLevel(lower, state) != fallbackNone {
			switch {
			case state.LastArticleNameCN != "" && !strings.Contains(question, state.LastArticleNameCN):
				return question + " 中文名称 " + state.LastArticleNameCN
			case state.LastJournalName != "" && !strings.Contains(question, state.LastJournalName):
				return question + " 期刊名称 " + state.LastJournalName
			}
		}
	}
	return question
}

func applyResultContinuation(req *Requests.AiRobotChatRequest, question string, state *State) string {
	if req == nil || state == nil {
		return question
	}
	lower := strings.ToLower(strings.TrimSpace(question))
	switch req.QuestionType {
	case Requests.AiQuestionTypeTask:
		if mentionsAny(lower, "为什么失败", "失败原因", "报错原因", "为啥失败") && len(state.LastTaskUUIDs) > 0 && !strings.Contains(question, "任务编号") {
			return question + " 任务编号 " + state.LastTaskUUIDs[0]
		}
		if mentionsAny(lower, "哪个链接", "哪个下载", "有下载吗", "有链接吗") && state.LastDownloadKind != "" && len(state.LastTaskUUIDs) > 0 && !strings.Contains(question, "任务编号") {
			return question + " 任务编号 " + state.LastTaskUUIDs[0] + " 下载类型 " + state.LastDownloadKind
		}
		if mentionsAny(lower, "还有别的结果吗", "还有其他结果吗", "还有吗") && len(state.LastTaskUUIDs) > 0 && !strings.Contains(question, "任务编号") {
			return question + " 任务编号 " + state.LastTaskUUIDs[0]
		}
	case Requests.AiQuestionTypeProject:
		if mentionsAny(lower, "哪个链接", "哪个下载", "有下载吗", "有链接吗") && state.LastDownloadKind != "" && state.LastProjectNumber != "" && !strings.Contains(question, "项目编号") {
			return question + " 项目编号 " + state.LastProjectNumber + " 下载类型 " + state.LastDownloadKind
		}
		if mentionsAny(lower, "还有别的结果吗", "还有其他结果吗", "还有吗") && state.LastProjectNumber != "" && !strings.Contains(question, "项目编号") {
			return question + " 项目编号 " + state.LastProjectNumber
		}
	case Requests.AiQuestionTypeProjectArticle:
		if mentionsAny(lower, "还有别的结果吗", "还有其他结果吗", "还有吗") {
			switch {
			case state.LastArticleNameCN != "" && !strings.Contains(question, "中文名称"):
				return question + " 中文名称 " + state.LastArticleNameCN
			case state.LastJournalName != "" && !strings.Contains(question, "期刊名称"):
				return question + " 期刊名称 " + state.LastJournalName
			}
		}
	}
	return question
}

type fallbackLevel int

const (
	fallbackNone fallbackLevel = iota
	fallbackWeak
	fallbackStrong
)

func isLikelyTopicShift(req *Requests.AiRobotChatRequest, state *State, question string) bool {
	if req == nil || state == nil {
		return false
	}
	if state.QuestionType > 0 && req.QuestionType > 0 && state.QuestionType != req.QuestionType {
		return true
	}
	domain := detectQuestionDomain(question)
	if domain == 0 {
		return false
	}
	return req.QuestionType > 0 && domain != req.QuestionType
}

func detectQuestionDomain(question string) int {
	q := strings.ToLower(strings.TrimSpace(question))
	switch {
	case mentionsAny(q, "文章", "文献", "期刊", "中文名称", "英文名称"):
		return Requests.AiQuestionTypeProjectArticle
	case mentionsAny(q, "任务", "工单", "任务编号", "uuid", "状态", "进度", "失败原因", "结果呢", "跑完没", "做完没"):
		return Requests.AiQuestionTypeTask
	case mentionsAny(q, "项目", "合同", "项目编号", "合同编号", "结题报告", "原始数据", "质控报告"):
		return Requests.AiQuestionTypeProject
	default:
		return 0
	}
}

func applyIntentContinuation(question string, state *State) string {
	if state == nil {
		return question
	}
	lower := strings.ToLower(strings.TrimSpace(question))
	switch state.LastIntent {
	case "download_final_report":
		if mentionsAny(lower, "原始数据呢", "原始数据", "授权码呢", "提取码呢") && !strings.Contains(question, "项目编号") && state.LastProjectNumber != "" {
			return question + " 项目编号 " + state.LastProjectNumber
		}
	case "download_original_data":
		if mentionsAny(lower, "结题报告呢", "报告呢", "质控报告呢", "报告下载") && !strings.Contains(question, "项目编号") && state.LastProjectNumber != "" {
			return question + " 项目编号 " + state.LastProjectNumber
		}
	case "task_status":
		if mentionsAny(lower, "结果呢", "下载结果", "导出结果", "下载呢", "下载结果呢") && !strings.Contains(question, "任务编号") && len(state.LastTaskUUIDs) > 0 {
			return question + " 任务编号 " + state.LastTaskUUIDs[0]
		}
	case "task_download_result":
		if mentionsAny(lower, "状态呢", "跑完没", "失败了吗", "成功了吗") && !strings.Contains(question, "任务编号") && len(state.LastTaskUUIDs) > 0 {
			return question + " 任务编号 " + state.LastTaskUUIDs[0]
		}
	case "project_article_list":
		if mentionsAny(lower, "还有别的吗", "再查下", "继续查", "详细点") {
			switch {
			case state.LastArticleNameCN != "" && !strings.Contains(question, "中文名称"):
				return question + " 中文名称 " + state.LastArticleNameCN
			case state.LastJournalName != "" && !strings.Contains(question, "期刊名称"):
				return question + " 期刊名称 " + state.LastJournalName
			}
		}
	}
	return question
}

func mentionsAny(question string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(question, keyword) {
			return true
		}
	}
	return false
}

func isShortFollowUp(question string) bool {
	question = strings.TrimSpace(question)
	return len([]rune(question)) > 0 && len([]rune(question)) <= 12
}

func isWeakFollowUp(question string) bool {
	return isShortFollowUp(question) && mentionsAny(question,
		"下载链接", "链接呢", "原始数据呢", "报告呢", "结果呢", "状态呢",
		"授权码呢", "提取码呢", "详细点", "再查下", "继续查",
	)
}

func projectFallbackLevel(question string, state *State) fallbackLevel {
	if mentionsAny(question, "这个项目", "那个项目", "上个项目", "该项目", "还是这个项目", "继续查这个项目", "刚才这个项目") {
		return fallbackStrong
	}
	if isWeakFollowUp(question) && state != nil && state.LastProjectNumber != "" {
		return fallbackWeak
	}
	return fallbackNone
}

func contractFallbackLevel(question string, state *State) fallbackLevel {
	if mentionsAny(question, "这个合同", "那个合同", "上个合同", "该合同", "还是这个合同", "继续查这个合同") && state != nil && state.LastContractNumber != "" {
		return fallbackStrong
	}
	return fallbackNone
}

func taskFallbackLevel(question string, state *State) fallbackLevel {
	if mentionsAny(question, "这个工单", "那个工单", "这个任务", "那个任务", "上个任务", "该任务", "继续查这个工单", "还是这个工单", "它跑完没", "它失败了吗") {
		return fallbackStrong
	}
	if isWeakFollowUp(question) && state != nil && len(state.LastTaskUUIDs) > 0 {
		return fallbackWeak
	}
	return fallbackNone
}

func articleFallbackLevel(question string, state *State) fallbackLevel {
	if mentionsAny(question, "这篇文章", "那个文章", "这篇文献", "那个文献", "刚才那篇文章", "还是这篇文章") {
		return fallbackStrong
	}
	if isWeakFollowUp(question) && state != nil && (state.LastArticleNameCN != "" || state.LastJournalName != "") {
		return fallbackWeak
	}
	return fallbackNone
}

func errorsIsNotFound(err error) bool {
	return err == ErrConversationNotFound
}

func autoFillMaxAge() time.Duration {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil || cfg.ConversationAutoFillMaxAgeMinutes <= 0 {
		return 10 * time.Minute
	}
	return time.Duration(cfg.ConversationAutoFillMaxAgeMinutes) * time.Minute
}

func autoFillMaxTurns() int {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil || cfg.ConversationAutoFillMaxTurns <= 0 {
		return 3
	}
	return cfg.ConversationAutoFillMaxTurns
}

func summaryMaxTurns() int {
	cfg := Config.GetAiGatewayConfig()
	if cfg == nil || cfg.ConversationSummaryMaxTurns <= 0 {
		return 5
	}
	return cfg.ConversationSummaryMaxTurns
}

func DebugString(state *State) string {
	if state == nil {
		return ""
	}
	return fmt.Sprintf("intent=%s project=%s contract=%s task=%v turns=%d", state.LastIntent, state.LastProjectNumber, state.LastContractNumber, state.LastTaskUUIDs, len(state.RecentTurns))
}
