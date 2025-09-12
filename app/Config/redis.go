package Config

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
	DB       int    `mapstructure:"db"` // 添加 DB 字段，与 Database 字段兼容
}

// SetDefaults 设置Redis配置默认值
func (r *RedisConfig) SetDefaults() {
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.db", 0)
}

// BindEnvs 绑定Redis环境变量
func (r *RedisConfig) BindEnvs() {
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.database", "REDIS_DATABASE")
	viper.BindEnv("redis.db", "REDIS_DB")
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() *RedisConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.Redis
}

// GetAddr 获取Redis地址
func (r *RedisConfig) GetAddr() string {
	return r.Host + ":" + strconv.Itoa(r.Port)
}

// GetConnectionString 获取Redis连接字符串
func (r *RedisConfig) GetConnectionString() string {
	// 优先使用 DB 字段，如果没有则使用 Database 字段
	db := r.DB
	if db == 0 && r.Database != 0 {
		db = r.Database
	}

	if r.Password == "" {
		return fmt.Sprintf("redis://%s:%d/%d", r.Host, r.Port, db)
	}
	return fmt.Sprintf("redis://:%s@%s:%d/%d", r.Password, r.Host, r.Port, db)
}

// Validate 验证Redis配置
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("Redis主机未配置")
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("Redis端口配置无效: %d", r.Port)
	}

	// 优先使用 DB 字段，如果没有则使用 Database 字段
	db := r.DB
	if db == 0 && r.Database != 0 {
		db = r.Database
	}

	if db < 0 || db > 15 {
		return fmt.Errorf("Redis数据库编号无效，应在0-15之间")
	}

	return nil
}

// IsPasswordSet 检查是否设置了密码
func (r *RedisConfig) IsPasswordSet() bool {
	return r.Password != ""
}
