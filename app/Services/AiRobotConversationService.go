package Services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AiRobotConversationService struct{}

type AiRobotConversationListFilter struct {
	UserID         string
	Platform       string
	ConversationID string
	QuestionType   int
	StartAt        time.Time
	EndAt          time.Time
	Page           int64
	Limit          int64
}

type AiRobotConversationCleanupFilter struct {
	UserID       string
	Platform     string
	QuestionType int
	StartAt      time.Time
	EndAt        time.Time
}

func NewAiRobotConversationService() *AiRobotConversationService {
	return &AiRobotConversationService{}
}

func (s *AiRobotConversationService) BuildDocumentID(userID, platform, conversationID string) string {
	return strings.TrimSpace(userID) + ":" + strings.TrimSpace(platform) + ":" + strings.TrimSpace(conversationID)
}

func (s *AiRobotConversationService) FindByID(ctx context.Context, databaseName, collectionName, id string) (*Models.AiRobotConversation, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}

	var doc Models.AiRobotConversation
	if err := collection.FindOne(ctx, bson.M{"_id": strings.TrimSpace(id)}).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *AiRobotConversationService) FindByConversation(ctx context.Context, databaseName, collectionName, userID, platform, conversationID string) (*Models.AiRobotConversation, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}
	// 单条会话读取优先走业务字段查询，而不是直接拼 _id。
	// 这样管理员显式传 user_id 时能精确定位，后续若 _id 规则再调整也更稳。
	query := bson.M{
		"platform":        strings.TrimSpace(platform),
		"conversation_id": strings.TrimSpace(conversationID),
	}
	if userID = strings.TrimSpace(userID); userID != "" {
		query["user_id"] = userID
	}
	var doc Models.AiRobotConversation
	if err := collection.FindOne(ctx, query).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *AiRobotConversationService) Upsert(ctx context.Context, databaseName, collectionName string, doc *Models.AiRobotConversation) error {
	if doc == nil {
		return fmt.Errorf("ai robot conversation doc is nil")
	}
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return err
	}
	_, err = collection.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true))
	return err
}

func (s *AiRobotConversationService) DeleteByID(ctx context.Context, databaseName, collectionName, id string) error {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return err
	}
	result, err := collection.DeleteOne(ctx, bson.M{"_id": strings.TrimSpace(id)})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (s *AiRobotConversationService) DeleteByConversation(ctx context.Context, databaseName, collectionName, userID, platform, conversationID string) error {
	doc, err := s.FindByConversation(ctx, databaseName, collectionName, userID, platform, conversationID)
	if err != nil {
		return err
	}
	return s.DeleteByID(ctx, databaseName, collectionName, doc.ID)
}

func (s *AiRobotConversationService) DeleteByFilter(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationCleanupFilter) (int64, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return 0, err
	}
	query := s.buildListQuery(AiRobotConversationListFilter{
		UserID:       filter.UserID,
		Platform:     filter.Platform,
		QuestionType: filter.QuestionType,
		StartAt:      filter.StartAt,
		EndAt:        filter.EndAt,
	})
	result, err := collection.DeleteMany(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (s *AiRobotConversationService) List(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationListFilter) ([]Models.AiRobotConversation, int64, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, 0, err
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	query := s.buildListQuery(filter)

	total, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetSkip((filter.Page - 1) * filter.Limit).
		SetLimit(filter.Limit)

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var docs []Models.AiRobotConversation
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

func (s *AiRobotConversationService) buildListQuery(filter AiRobotConversationListFilter) bson.M {
	query := bson.M{}
	if userID := strings.TrimSpace(filter.UserID); userID != "" {
		query["user_id"] = userID
	}
	if platform := strings.TrimSpace(filter.Platform); platform != "" {
		query["platform"] = platform
	}
	if conversationID := strings.TrimSpace(filter.ConversationID); conversationID != "" {
		query["conversation_id"] = bson.M{"$regex": conversationID, "$options": "i"}
	}
	if filter.QuestionType > 0 {
		query["question_type"] = filter.QuestionType
	}
	if !filter.StartAt.IsZero() || !filter.EndAt.IsZero() {
		updatedAt := bson.M{}
		if !filter.StartAt.IsZero() {
			updatedAt["$gte"] = filter.StartAt
		}
		if !filter.EndAt.IsZero() {
			updatedAt["$lte"] = filter.EndAt
		}
		query["updated_at"] = updatedAt
	}
	return query
}

func (s *AiRobotConversationService) EnsureIndexes(ctx context.Context, databaseName, collectionName string, ttlHours int) error {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return err
	}
	if ttlHours <= 0 {
		ttlHours = 24 * 30
	}
	_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "platform", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "updated_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(ttlHours * 3600)),
		},
	})
	return err
}

func (s *AiRobotConversationService) getCollection(databaseName, collectionName string) (*mongo.Collection, error) {
	name := strings.TrimSpace(collectionName)
	if name == "" {
		name = Models.AiRobotConversation{}.CollectionName()
	}
	return Database.GetMongoCollection(databaseName, name)
}
