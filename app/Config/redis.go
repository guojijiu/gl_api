package Config

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// SetDefaults 设置Redis配置默认值
func (r *RedisConfig) SetDefaults() {
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.database", 0)
}

// BindEnvs 绑定Redis环境变量
func (r *RedisConfig) BindEnvs() {
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.database", "REDIS_DATABASE")
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
	if r.Password == "" {
		return fmt.Sprintf("redis://%s:%s/%d", r.Host, r.Port, r.Database)
	}
	return fmt.Sprintf("redis://:%s@%s:%s/%d", r.Password, r.Host, r.Port, r.Database)
}

// Validate 验证Redis配置
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("Redis主机未配置")
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("Redis端口配置无效: %d", r.Port)
	}

	if r.Database < 0 || r.Database > 15 {
		return fmt.Errorf("Redis数据库编号无效，应在0-15之间")
	}

	return nil
}

// IsPasswordSet 检查是否设置了密码
func (r *RedisConfig) IsPasswordSet() bool {
	return r.Password != ""
}
