package Config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    string `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	BaseURL string `mapstructure:"base_url"`
}

// SetDefaults 设置服务器配置默认值
func (s *ServerConfig) SetDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.base_url", "http://localhost:8080")
}

// BindEnvs 绑定服务器环境变量
func (s *ServerConfig) BindEnvs() {
	viper.BindEnv("server.port", "PORT")        // 支持 PORT 环境变量
	viper.BindEnv("server.port", "SERVER_PORT") // 也支持 SERVER_PORT 环境变量
	viper.BindEnv("server.mode", "SERVER_MODE")
	viper.BindEnv("server.base_url", "SERVER_BASE_URL")
}

// Validate 验证服务器配置
func (s *ServerConfig) Validate() error {
	if s.Port == "" {
		return fmt.Errorf("服务器端口未配置")
	}
	return nil
}

// GetServerConfig 获取服务器配置
func GetServerConfig() *ServerConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.Server
}

// IsDebugMode 检查是否为调试模式
func (s *ServerConfig) IsDebugMode() bool {
	return s.Mode == "debug"
}

// IsProductionMode 检查是否为生产模式
func (s *ServerConfig) IsProductionMode() bool {
	return s.Mode == "production"
}

// GetFullURL 获取完整的URL
func (s *ServerConfig) GetFullURL(path string) string {
	return s.BaseURL + path
}
