package Config

import (
	"strings"
	"testing"
)

func TestConversationMongoConnString(t *testing.T) {
	cfg := &AiGatewayConfig{
		ConversationMongoHost:       "127.0.0.1",
		ConversationMongoPort:       27117,
		ConversationMongoDatabase:   "cloud_platform_v2",
		ConversationMongoUsername:   "root",
		ConversationMongoPassword:   `}rR.92Gc=?[(`,
		ConversationMongoAuthSource: "admin",
	}

	got := cfg.ConversationMongoConnString()
	if !strings.Contains(got, "mongodb://root:") {
		t.Fatalf("unexpected conn string: %s", got)
	}
	if !strings.Contains(got, "@127.0.0.1:27117/cloud_platform_v2") {
		t.Fatalf("unexpected conn string: %s", got)
	}
	if !strings.Contains(got, "authSource=admin") {
		t.Fatalf("unexpected conn string: %s", got)
	}
}

func TestConversationAutoFillDefaults(t *testing.T) {
	cfg := &AiGatewayConfig{}
	cfg.SetDefaults()
	if cfg.ConversationAutoFillMaxAgeMinutes != 0 || cfg.ConversationAutoFillMaxTurns != 0 {
		t.Fatalf("set defaults should rely on viper defaults, struct should remain zero before unmarshal")
	}
}

func TestEffectiveConversationMongoConfigUsesSharedMongoConfig(t *testing.T) {
	shared := &MongoDBConfig{
		Host:       "121.43.37.234",
		Port:       27117,
		Database:   "cloud_platform",
		Username:   "root",
		Password:   `}rR.92Gc=?[(`,
		AuthSource: "admin",
		TimeoutSec: 5,
	}
	cfg := &AiGatewayConfig{}

	got := cfg.EffectiveConversationMongoConfig(shared)
	if got.Host != shared.Host || got.Port != shared.Port || got.Database != shared.Database {
		t.Fatalf("unexpected effective mongo config: %+v", got)
	}
	if !strings.Contains(got.ConnString(), "@121.43.37.234:27117/cloud_platform") {
		t.Fatalf("unexpected conn string: %s", got.ConnString())
	}
}
