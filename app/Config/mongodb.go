package Config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// MongoDBConfig 公共 MongoDB 配置。
// 供 ai_robot 以及后续其他需要 MongoDB CRUD 的模块复用。
type MongoDBConfig struct {
	URI        string `mapstructure:"uri"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Database   string `mapstructure:"database"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	AuthSource string `mapstructure:"auth_source"`
	TimeoutSec int    `mapstructure:"timeout_sec"`
}

func (c *MongoDBConfig) SetDefaults() {
	viper.SetDefault("mongodb.host", "127.0.0.1")
	viper.SetDefault("mongodb.port", 27017)
	viper.SetDefault("mongodb.database", "cloud_platform")
	viper.SetDefault("mongodb.auth_source", "admin")
	viper.SetDefault("mongodb.timeout_sec", 5)
}

func (c *MongoDBConfig) BindEnvs() {
	viper.BindEnv("mongodb.uri", "MONGODB_URI")
	viper.BindEnv("mongodb.host", "MONGODB_HOST")
	viper.BindEnv("mongodb.port", "MONGODB_PORT")
	viper.BindEnv("mongodb.database", "MONGODB_DATABASE")
	viper.BindEnv("mongodb.username", "MONGODB_USERNAME")
	viper.BindEnv("mongodb.password", "MONGODB_PASSWORD")
	viper.BindEnv("mongodb.auth_source", "MONGODB_AUTH_SOURCE")
	viper.BindEnv("mongodb.timeout_sec", "MONGODB_TIMEOUT_SEC")
}

func GetMongoDBConfig() *MongoDBConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.MongoDB
}

func (c *MongoDBConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("mongodb config is nil")
	}
	if strings.TrimSpace(c.URI) != "" {
		return nil
	}
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("mongodb host 未配置")
	}
	if c.Port <= 0 {
		return fmt.Errorf("mongodb port 未配置")
	}
	if strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("mongodb database 未配置")
	}
	return nil
}

func (c *MongoDBConfig) ConnString() string {
	if c == nil {
		return ""
	}
	if raw := strings.TrimSpace(c.URI); raw != "" {
		return raw
	}

	host := strings.TrimSpace(c.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := c.Port
	if port <= 0 {
		port = 27017
	}
	database := strings.TrimSpace(c.Database)
	if database == "" {
		database = "cloud_platform"
	}
	authSource := strings.TrimSpace(c.AuthSource)
	if authSource == "" {
		authSource = "admin"
	}

	credentials := ""
	if user := strings.TrimSpace(c.Username); user != "" {
		credentials = url.QueryEscape(user)
		if pwd := c.Password; pwd != "" {
			credentials += ":" + url.QueryEscape(pwd)
		}
		credentials += "@"
	}

	params := url.Values{}
	if credentials != "" {
		params.Set("authSource", authSource)
	}

	query := params.Encode()
	if query != "" {
		query = "?" + query
	}

	return "mongodb://" + credentials + host + ":" + strconv.Itoa(port) + "/" + database + query
}
