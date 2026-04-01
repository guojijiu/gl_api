package Database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud-platform-api/app/Config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoClient     *mongo.Client
	mongoClientErr  error
	mongoClientOnce sync.Once
)

func GetMongoClient() (*mongo.Client, error) {
	mongoClientOnce.Do(func() {
		cfg := Config.GetMongoDBConfig()
		if cfg == nil {
			mongoClientErr = fmt.Errorf("mongodb config is nil")
			return
		}
		if err := cfg.Validate(); err != nil {
			mongoClientErr = err
			return
		}

		timeoutSec := cfg.TimeoutSec
		if timeoutSec <= 0 {
			timeoutSec = 5
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
		defer cancel()

		client, err := mongo.Connect(options.Client().ApplyURI(cfg.ConnString()))
		if err != nil {
			mongoClientErr = err
			return
		}
		if err := client.Ping(ctx, nil); err != nil {
			_ = client.Disconnect(context.Background())
			mongoClientErr = err
			return
		}
		mongoClient = client
	})
	return mongoClient, mongoClientErr
}

func GetMongoDatabase(databaseName string) (*mongo.Database, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(databaseName)
	if name == "" {
		cfg := Config.GetMongoDBConfig()
		if cfg == nil {
			return nil, fmt.Errorf("mongodb config is nil")
		}
		name = strings.TrimSpace(cfg.Database)
	}
	if name == "" {
		return nil, fmt.Errorf("mongodb database is empty")
	}
	return client.Database(name), nil
}

func GetMongoCollection(databaseName, collectionName string) (*mongo.Collection, error) {
	db, err := GetMongoDatabase(databaseName)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(collectionName)
	if name == "" {
		return nil, fmt.Errorf("mongodb collection is empty")
	}
	return db.Collection(name), nil
}

func CloseMongoClient() error {
	if mongoClient == nil {
		return nil
	}
	return mongoClient.Disconnect(context.Background())
}
