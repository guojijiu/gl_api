package Database

import (
	"os"
	"testing"

	"cloud-platform-api/app/Config"
)

func TestGetMongoDatabaseWithSharedConfig(t *testing.T) {
	if os.Getenv("MONGODB_HOST") == "" {
		t.Skip("skip mongo integration test without MONGODB_HOST")
	}

	Config.LoadConfig()

	db, err := GetMongoDatabase("")
	if err != nil {
		t.Fatalf("get mongo database failed: %v", err)
	}
	if db == nil {
		t.Fatal("mongo database is nil")
	}

	cfg := Config.GetMongoDBConfig()
	if cfg == nil {
		t.Fatal("mongo config is nil")
	}
	if db.Name() != cfg.Database {
		t.Fatalf("unexpected mongo database name: got=%s want=%s", db.Name(), cfg.Database)
	}
}
