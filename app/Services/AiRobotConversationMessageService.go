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

type AiRobotConversationMessageListFilter struct {
	UserID         string
	Platform       string
	ConversationID string
	MessageID      string
	QuestionType   int
	StartAt        time.Time
	EndAt          time.Time
	Page           int64
	Limit          int64
}

type AiRobotConversationMessageService struct{}

type AiRobotConversationMessageTrendPoint struct {
	Date                        string `json:"date"`
	Total                       int64  `json:"total"`
	FailedCount                 int64  `json:"failed_count"`
	EmptyCount                  int64  `json:"empty_count"`
	DownloadCount               int64  `json:"download_count"`
	SuccessWithoutDownloadCount int64  `json:"success_without_download_count"`
}

type AiRobotConversationMessageStats struct {
	Total              int64                                  `json:"total"`
	SuccessCount       int64                                  `json:"success_count"`
	FailedCount        int64                                  `json:"failed_count"`
	ResultKindCounts   map[string]int64                       `json:"result_kind_counts"`
	QuestionTypeCounts map[string]int64                       `json:"question_type_counts"`
	DailyTrend         []AiRobotConversationMessageTrendPoint `json:"daily_trend"`
}

func NewAiRobotConversationMessageService() *AiRobotConversationMessageService {
	return &AiRobotConversationMessageService{}
}

func (s *AiRobotConversationMessageService) BuildDocumentID(userID, platform, conversationID, messageID string) string {
	return strings.TrimSpace(userID) + ":" + strings.TrimSpace(platform) + ":" + strings.TrimSpace(conversationID) + ":" + strings.TrimSpace(messageID)
}

func (s *AiRobotConversationMessageService) Upsert(ctx context.Context, databaseName, collectionName string, doc *Models.AiRobotConversationMessage) error {
	if doc == nil {
		return fmt.Errorf("ai robot conversation message doc is nil")
	}
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return err
	}
	_, err = collection.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true))
	return err
}

func (s *AiRobotConversationMessageService) FindOne(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) (*Models.AiRobotConversationMessage, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}
	query := s.buildListQuery(filter)
	var doc Models.AiRobotConversationMessage
	if err := collection.FindOne(ctx, query).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *AiRobotConversationMessageService) FindLatest(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) (*Models.AiRobotConversationMessage, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}
	query := s.buildListQuery(filter)
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var doc Models.AiRobotConversationMessage
	if err := collection.FindOne(ctx, query, opts).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *AiRobotConversationMessageService) List(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) ([]Models.AiRobotConversationMessage, int64, error) {
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
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetSkip((filter.Page - 1) * filter.Limit).
		SetLimit(filter.Limit)
	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var docs []Models.AiRobotConversationMessage
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

// ListNoTotal 仅返回消息列表，不执行 total 统计，适合会话详情等轻量场景。
func (s *AiRobotConversationMessageService) ListNoTotal(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) ([]Models.AiRobotConversationMessage, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	query := s.buildListQuery(filter)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetSkip((filter.Page - 1) * filter.Limit).
		SetLimit(filter.Limit)
	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []Models.AiRobotConversationMessage
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *AiRobotConversationMessageService) DeleteByConversation(ctx context.Context, databaseName, collectionName, userID, platform, conversationID string) (int64, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return 0, err
	}
	query := bson.M{
		"user_id":         strings.TrimSpace(userID),
		"platform":        strings.TrimSpace(platform),
		"conversation_id": strings.TrimSpace(conversationID),
	}
	result, err := collection.DeleteMany(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (s *AiRobotConversationMessageService) DeleteByFilter(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) (int64, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return 0, err
	}
	result, err := collection.DeleteMany(ctx, s.buildListQuery(filter))
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (s *AiRobotConversationMessageService) Stats(ctx context.Context, databaseName, collectionName string, filter AiRobotConversationMessageListFilter) (*AiRobotConversationMessageStats, error) {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return nil, err
	}
	query := s.buildListQuery(filter)
	total, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, err
	}
	successCount, err := collection.CountDocuments(ctx, mergeMessageStatsQuery(query, bson.M{"success": true}))
	if err != nil {
		return nil, err
	}
	failedCount, err := collection.CountDocuments(ctx, mergeMessageStatsQuery(query, bson.M{"success": false}))
	if err != nil {
		return nil, err
	}
	resultKindCounts, err := s.groupCount(ctx, collection, query, "result_kind")
	if err != nil {
		return nil, err
	}
	questionTypeCounts, err := s.groupCount(ctx, collection, query, "question_type")
	if err != nil {
		return nil, err
	}
	dailyTrend, err := s.groupDailyTrend(ctx, collection, query)
	if err != nil {
		return nil, err
	}
	return &AiRobotConversationMessageStats{
		Total:              total,
		SuccessCount:       successCount,
		FailedCount:        failedCount,
		ResultKindCounts:   resultKindCounts,
		QuestionTypeCounts: questionTypeCounts,
		DailyTrend:         dailyTrend,
	}, nil
}

func (s *AiRobotConversationMessageService) EnsureIndexes(ctx context.Context, databaseName, collectionName string, ttlHours int) error {
	collection, err := s.getCollection(databaseName, collectionName)
	if err != nil {
		return err
	}
	if ttlHours <= 0 {
		ttlHours = 24 * 30
	}
	_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "platform", Value: 1}, {Key: "conversation_id", Value: 1}, {Key: "message_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "platform", Value: 1}, {Key: "conversation_id", Value: 1}, {Key: "created_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "platform", Value: 1}, {Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "platform", Value: 1}, {Key: "question_type", Value: 1}, {Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "platform", Value: 1}, {Key: "result_kind", Value: 1}, {Key: "updated_at", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "updated_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(ttlHours * 3600)),
		},
	})
	return err
}

func (s *AiRobotConversationMessageService) buildListQuery(filter AiRobotConversationMessageListFilter) bson.M {
	query := bson.M{}
	if userID := strings.TrimSpace(filter.UserID); userID != "" {
		query["user_id"] = userID
	}
	if platform := strings.TrimSpace(filter.Platform); platform != "" {
		query["platform"] = platform
	}
	if conversationID := strings.TrimSpace(filter.ConversationID); conversationID != "" {
		query["conversation_id"] = conversationID
	}
	if messageID := strings.TrimSpace(filter.MessageID); messageID != "" {
		query["message_id"] = messageID
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

func (s *AiRobotConversationMessageService) getCollection(databaseName, collectionName string) (*mongo.Collection, error) {
	name := strings.TrimSpace(collectionName)
	if name == "" {
		name = Models.AiRobotConversationMessage{}.CollectionName()
	}
	return Database.GetMongoCollection(databaseName, name)
}

func (s *AiRobotConversationMessageService) groupCount(ctx context.Context, collection *mongo.Collection, query bson.M, field string) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: query}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$" + field,
			"count": bson.M{"$sum": 1},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(fmt.Sprint(row["_id"]))
		if key == "" || key == "<nil>" {
			key = "unknown"
		}
		result[key] = toInt64(row["count"])
	}
	return result, nil
}

func mergeMessageStatsQuery(base bson.M, extra bson.M) bson.M {
	merged := bson.M{}
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return merged
}

func toInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func (s *AiRobotConversationMessageService) groupDailyTrend(ctx context.Context, collection *mongo.Collection, query bson.M) ([]AiRobotConversationMessageTrendPoint, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: query}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$updated_at",
				},
			},
			"total": bson.M{"$sum": 1},
			"failed_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$success", false}},
					1,
					0,
				},
			}},
			"empty_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$result_kind", "empty"}},
					1,
					0,
				},
			}},
			"download_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$result_kind", "download"}},
					1,
					0,
				},
			}},
			"success_without_download_count": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$success", true}},
						bson.M{"$ne": bson.A{"$result_kind", "download"}},
					}},
					1,
					0,
				},
			}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	points := make([]AiRobotConversationMessageTrendPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, AiRobotConversationMessageTrendPoint{
			Date:                        strings.TrimSpace(fmt.Sprint(row["_id"])),
			Total:                       toInt64(row["total"]),
			FailedCount:                 toInt64(row["failed_count"]),
			EmptyCount:                  toInt64(row["empty_count"]),
			DownloadCount:               toInt64(row["download_count"]),
			SuccessWithoutDownloadCount: toInt64(row["success_without_download_count"]),
		})
	}
	return points, nil
}
