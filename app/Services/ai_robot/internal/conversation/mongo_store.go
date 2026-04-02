package conversation

import (
	"context"
	"fmt"
	"time"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoStore struct {
	service        *Services.AiRobotConversationService
	databaseName   string
	collectionName string
}

func NewMongoStore(cfg *Config.AiGatewayConfig) (*MongoStore, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ai gateway config is nil")
	}
	effective := cfg.EffectiveConversationMongoConfig(Config.GetMongoDBConfig())
	timeoutSec := cfg.ConversationMongoTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = effective.TimeoutSec
	}
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	store := &MongoStore{
		service:        Services.NewAiRobotConversationService(),
		databaseName:   effective.Database,
		collectionName: cfg.ConversationMongoCollection,
	}
	if err := store.ensureIndexes(ctx, cfg.ConversationTTLHours); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *MongoStore) Get(ctx context.Context, key string) (*State, error) {
	doc, err := s.service.FindByID(ctx, s.databaseName, s.collectionName, key)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}
	state := modelToState(doc)
	return &state, nil
}

func (s *MongoStore) Save(ctx context.Context, key string, state *State) error {
	if state == nil {
		return fmt.Errorf("conversation state is nil")
	}
	doc := stateToModel(key, state)
	return s.service.Upsert(ctx, s.databaseName, s.collectionName, &doc)
}

func (s *MongoStore) ensureIndexes(ctx context.Context, ttlHours int) error {
	return s.service.EnsureIndexes(ctx, s.databaseName, s.collectionName, ttlHours)
}

func stateToModel(key string, state *State) Models.AiRobotConversation {
	doc := Models.AiRobotConversation{
		ID:                 key,
		UserID:             state.UserID,
		ConversationID:     state.ConversationID,
		Platform:           state.Platform,
		QuestionType:       state.QuestionType,
		LastMessageID:      state.LastMessageID,
		LastIntent:         state.LastIntent,
		LastQuestion:       state.LastQuestion,
		LastResolved:       state.LastResolved,
		LastAnswer:         state.LastAnswer,
		ContextSummary:     state.ContextSummary,
		LastProjectIDs:     append([]int(nil), state.LastProjectIDs...),
		LastProjectNumber:  state.LastProjectNumber,
		LastContractIDs:    append([]int(nil), state.LastContractIDs...),
		LastContractNumber: state.LastContractNumber,
		LastTaskUUIDs:      append([]string(nil), state.LastTaskUUIDs...),
		LastArticleNameCN:  state.LastArticleNameCN,
		LastArticleNameEN:  state.LastArticleNameEN,
		LastArticleLabels:  append([]string(nil), state.LastArticleLabels...),
		LastJournalName:    state.LastJournalName,
		LastTaskStatus:     state.LastTaskStatus,
		LastFailureReason:  state.LastFailureReason,
		LastDownloadKind:   state.LastDownloadKind,
		LastHasLink:        state.LastHasLink,
		UpdatedAt:          state.UpdatedAt,
	}
	if len(state.RecentTurns) > 0 {
		doc.RecentTurns = make([]Models.AiRobotConversationTurn, 0, len(state.RecentTurns))
		for _, turn := range state.RecentTurns {
			doc.RecentTurns = append(doc.RecentTurns, Models.AiRobotConversationTurn{
				MessageID:      turn.MessageID,
				QuestionType:   turn.QuestionType,
				Question:       turn.Question,
				Resolved:       turn.Resolved,
				Answer:         turn.Answer,
				Intent:         turn.Intent,
				ProjectNumber:  turn.ProjectNumber,
				ContractNumber: turn.ContractNumber,
				TaskUUID:       turn.TaskUUID,
				ArticleNameCN:  turn.ArticleNameCN,
				JournalName:    turn.JournalName,
				CreatedAt:      turn.CreatedAt,
			})
		}
	}
	return doc
}

func modelToState(doc *Models.AiRobotConversation) State {
	if doc == nil {
		return State{}
	}
	state := State{
		UserID:             doc.UserID,
		ConversationID:     doc.ConversationID,
		Platform:           doc.Platform,
		QuestionType:       doc.QuestionType,
		LastMessageID:      doc.LastMessageID,
		LastIntent:         doc.LastIntent,
		LastQuestion:       doc.LastQuestion,
		LastResolved:       doc.LastResolved,
		LastAnswer:         doc.LastAnswer,
		ContextSummary:     doc.ContextSummary,
		LastProjectIDs:     append([]int(nil), doc.LastProjectIDs...),
		LastProjectNumber:  doc.LastProjectNumber,
		LastContractIDs:    append([]int(nil), doc.LastContractIDs...),
		LastContractNumber: doc.LastContractNumber,
		LastTaskUUIDs:      append([]string(nil), doc.LastTaskUUIDs...),
		LastArticleNameCN:  doc.LastArticleNameCN,
		LastArticleNameEN:  doc.LastArticleNameEN,
		LastArticleLabels:  append([]string(nil), doc.LastArticleLabels...),
		LastJournalName:    doc.LastJournalName,
		LastTaskStatus:     doc.LastTaskStatus,
		LastFailureReason:  doc.LastFailureReason,
		LastDownloadKind:   doc.LastDownloadKind,
		LastHasLink:        doc.LastHasLink,
		UpdatedAt:          doc.UpdatedAt,
	}
	if len(doc.RecentTurns) > 0 {
		state.RecentTurns = make([]Turn, 0, len(doc.RecentTurns))
		for _, turn := range doc.RecentTurns {
			state.RecentTurns = append(state.RecentTurns, Turn{
				MessageID:      turn.MessageID,
				QuestionType:   turn.QuestionType,
				Question:       turn.Question,
				Resolved:       turn.Resolved,
				Answer:         turn.Answer,
				Intent:         turn.Intent,
				ProjectNumber:  turn.ProjectNumber,
				ContractNumber: turn.ContractNumber,
				TaskUUID:       turn.TaskUUID,
				ArticleNameCN:  turn.ArticleNameCN,
				JournalName:    turn.JournalName,
				CreatedAt:      turn.CreatedAt,
			})
		}
	}
	return state
}
