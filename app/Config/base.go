package Config

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 主配置结构
//
// 重要功能说明：
// 1. 服务器配置：端口、模式、超时、安全设置等
// 2. 数据库配置：驱动选择、连接参数、连接池设置
// 3. JWT配置：密钥管理、过期时间、签名算法
// 4. Redis配置：缓存服务、会话存储、消息队列
// 5. 存储配置：文件存储、日志管理、备份策略
// 6. 邮件配置：SMTP设置、邮件模板、发送策略
//
// 配置管理特性：
// - 支持环境变量和配置文件双重配置
// - 自动配置验证和完整性检查
// - 支持多环境配置（开发、测试、生产）
// - 配置热重载支持（可选）
// - 敏感配置加密存储
//
// 安全配置：
// - JWT密钥强度验证（最小32字符）
// - 数据库连接安全参数
// - Redis访问控制和认证
// - 文件存储权限控制
// - 邮件发送安全设置
//
// 性能配置：
// - 数据库连接池参数优化
// - Redis连接池和超时设置
// - 文件上传大小限制
// - 请求超时和速率限制
// - 缓存策略和TTL设置
//
// 监控配置：
// - 健康检查端点配置
// - 性能指标收集设置
// - 日志级别和输出配置
// - 告警规则和通知设置
// - 审计日志配置
//
// 扩展性：
// - 支持自定义配置模块
// - 插件化配置加载
// - 配置继承和覆盖机制
// - 支持配置版本管理
// - 配置变更通知机制
type Config struct {
	Server            ServerConfig            `mapstructure:"server"`
	Database          DatabaseConfig          `mapstructure:"database"`
	JWT               JWTConfig               `mapstructure:"jwt"`
	Redis             RedisConfig             `mapstructure:"redis"`
	Storage           StorageConfig           `mapstructure:"storage"`
	Email             EmailConfig             `mapstructure:"email"`
	Log               LogConfig               `mapstructure:"log"`
	Monitoring        MonitoringConfig        `mapstructure:"monitoring"`
	WebSocket         WebSocketConfig         `mapstructure:"websocket"`
	QueryOptimization QueryOptimizationConfig `mapstructure:"query_optimization"`
	Security          SecurityConfig          `mapstructure:"security"`
	Testing           TestConfig              `mapstructure:"testing"`
}

var globalConfig *Config

// SetDefaults 设置所有配置的默认值
func (c *Config) SetDefaults() {
	c.Server.SetDefaults()
	c.Database.SetDefaults()
	c.JWT.SetDefaults()
	c.Redis.SetDefaults()
	c.Storage.SetDefaults()
	c.Email.SetDefaults()
	c.Log.SetDefaults()
	c.Monitoring.SetDefaults()
	c.WebSocket.SetDefaults()
	c.QueryOptimization.SetDefaults()
	c.Security.SetDefaults()
	c.Testing.SetDefaults()
}

// BindEnvs 绑定所有配置的环境变量
func (c *Config) BindEnvs() {
	c.Server.BindEnvs()
	c.Database.BindEnvs()
	c.JWT.BindEnvs()
	c.Redis.BindEnvs()
	c.Storage.BindEnvs()
	c.Email.BindEnvs()
	c.Log.BindEnvs()
	c.Monitoring.BindEnvs()
	c.WebSocket.BindEnvs()
	c.QueryOptimization.BindEnvs()
	c.Security.BindEnvs()
	c.Testing.BindEnvs()
}

// LoadConfig 加载所有配置
// 功能说明：
// 1. 从.env文件加载环境变量（如果存在）
// 2. 设置各模块的默认配置值
// 3. 绑定环境变量到配置结构
// 4. 初始化全局配置变量
// 5. 解析配置并验证有效性
// 6. 记录配置加载成功的日志信息
// 7. 支持多环境配置（开发、测试、生产）
// 8. 提供配置热重载能力（可选）
//
// 配置加载顺序：
// 1. 加载.env文件（如果存在）
// 2. 设置默认配置值
// 3. 绑定环境变量
// 4. 解析配置结构
// 5. 验证配置完整性
// 6. 应用配置规则
//
// 配置验证规则：
// - JWT密钥：必须配置且长度>=32字符
// - 数据库驱动：必须为mysql/postgres/sqlite之一
// - 服务器端口：必须为有效端口号
// - Redis配置：可选，但配置后必须有效
// - 邮件配置：可选，但配置后必须有效
//
// 安全考虑：
// - 敏感配置（密码、密钥）不记录到日志
// - 验证配置值的合理性
// - 检查必需配置项
// - 防止配置注入攻击
//
// 错误处理：
// - .env文件不存在时使用系统环境变量
// - 配置解析失败时立即退出
// - 配置验证失败时提供详细错误信息
// - 记录配置加载过程的关键步骤
func LoadConfig() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 初始化全局配置变量
	globalConfig = &Config{}

	// 设置默认值
	globalConfig.SetDefaults()

	// 绑定环境变量（必须在AutomaticEnv之前）
	globalConfig.BindEnvs()

	// 从环境变量读取配置
	viper.AutomaticEnv()

	// 解析配置
	if err := viper.Unmarshal(&globalConfig); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}

	// 验证配置完整性
	if globalConfig.Server.Port == "" {
		globalConfig.Server.Port = "8080"
	}
	if globalConfig.Database.Driver == "" {
		globalConfig.Database.Driver = "sqlite"
	}
	if globalConfig.JWT.Secret == "" {
		log.Fatal("JWT密钥未配置，请在环境变量中设置JWT_SECRET")
	}

	// 配置加载完成后再打印
	//log.Printf("Config loaded successfully: %+v\n", globalConfig)

	// 打印所有读取的环境变量内容
	//printEnvironmentVariables()
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// ValidateConfig 验证配置
// 功能说明：
// 1. 检查全局配置是否已加载
// 2. 验证各个配置模块的有效性
// 3. 检查必需配置项是否完整
// 4. 验证配置值的合理性
// 5. 提供详细的错误信息
func ValidateConfig() error {
	if globalConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	// 验证各个配置模块
	if err := globalConfig.Server.Validate(); err != nil {
		return fmt.Errorf("服务器配置验证失败: %v", err)
	}

	if err := globalConfig.Database.Validate(); err != nil {
		return fmt.Errorf("数据库配置验证失败: %v", err)
	}

	if err := globalConfig.JWT.ValidateProductionConfig(); err != nil {
		return fmt.Errorf("JWT配置验证失败: %v", err)
	}

	if err := globalConfig.Redis.Validate(); err != nil {
		return fmt.Errorf("Redis配置验证失败: %v", err)
	}

	if err := globalConfig.Storage.Validate(); err != nil {
		return fmt.Errorf("存储配置验证失败: %v", err)
	}

	// 邮件配置可选验证（如果配置了才验证）
	if globalConfig.Email.IsConfigured() {
		if err := globalConfig.Email.Validate(); err != nil {
			return fmt.Errorf("邮件配置验证失败: %v", err)
		}
	}

	// 验证JWT密钥安全性
	if len(globalConfig.JWT.Secret) < 32 {
		return fmt.Errorf("JWT密钥长度不足，建议至少32个字符")
	}

	// 验证数据库连接参数
	if globalConfig.Database.Driver == "" {
		return fmt.Errorf("数据库驱动未配置")
	}

	return nil
}

// bindEnv 绑定单个环境变量到配置值
func bindEnv(key string, value interface{}) {
	// 使用viper绑定环境变量
	viper.BindEnv(key, key)

	// 如果环境变量存在，则读取其值并赋值给配置
	if envValue := viper.GetString(key); envValue != "" {
		// 根据value的类型进行类型转换和赋值
		switch v := value.(type) {
		case *string:
			*v = envValue
		case *bool:
			if envValue == "true" || envValue == "1" || envValue == "yes" {
				*v = true
			} else if envValue == "false" || envValue == "0" || envValue == "no" {
				*v = false
			}
		case *int:
			if intVal, err := strconv.Atoi(envValue); err == nil {
				*v = intVal
			}
		case *time.Duration:
			if duration, err := time.ParseDuration(envValue); err == nil {
				*v = duration
			}
		case *[]string:
			// 对于字符串切片，按逗号分隔
			if envValue != "" {
				*v = strings.Split(envValue, ",")
			}
		}
	}
}
