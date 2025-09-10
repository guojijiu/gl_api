package Config

import (
	"fmt"
	"github.com/spf13/viper"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

// SetDefaults 设置默认值
func (e *EmailConfig) SetDefaults() {
	viper.SetDefault("email.host", "smtp.gmail.com")
	viper.SetDefault("email.port", 587)
	viper.SetDefault("email.username", "")
	viper.SetDefault("email.password", "")
	viper.SetDefault("email.from", "noreply@example.com")
	viper.SetDefault("email.use_tls", true)
}

// BindEnvs 绑定环境变量
func (e *EmailConfig) BindEnvs() {
	viper.BindEnv("email.host", "EMAIL_HOST")
	viper.BindEnv("email.port", "EMAIL_PORT")
	viper.BindEnv("email.username", "EMAIL_USERNAME")
	viper.BindEnv("email.password", "EMAIL_PASSWORD")
	viper.BindEnv("email.from", "EMAIL_FROM")
	viper.BindEnv("email.use_tls", "EMAIL_USE_TLS")
}

// Validate 验证配置
func (e *EmailConfig) Validate() error {
	if e.Host == "" {
		return fmt.Errorf("邮件服务器地址不能为空")
	}
	
	if e.Port <= 0 || e.Port > 65535 {
		return fmt.Errorf("邮件服务器端口无效: %d", e.Port)
	}
	
	if e.Username == "" {
		return fmt.Errorf("邮件用户名不能为空")
	}
	
	if e.Password == "" {
		return fmt.Errorf("邮件密码不能为空")
	}
	
	if e.From == "" {
		return fmt.Errorf("发件人地址不能为空")
	}
	
	return nil
}

// GetDSN 获取邮件服务器DSN
func (e *EmailConfig) GetDSN() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

// IsConfigured 检查是否已配置
func (e *EmailConfig) IsConfigured() bool {
	return e.Host != "" && e.Username != "" && e.Password != ""
}
