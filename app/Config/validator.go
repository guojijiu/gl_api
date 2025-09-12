package Config

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules map[string][]ValidationRule
}

// ValidationRule 验证规则接口
type ValidationRule interface {
	Validate(value interface{}) error
	GetMessage() string
}

// RequiredRule 必填规则
type RequiredRule struct {
	message string
}

func (r *RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf(r.message)
	}

	// 检查字符串是否为空
	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		return fmt.Errorf(r.message)
	}

	// 检查切片是否为空
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		slice := reflect.ValueOf(value)
		if slice.Len() == 0 {
			return fmt.Errorf(r.message)
		}
	}

	return nil
}

func (r *RequiredRule) GetMessage() string {
	return r.message
}

// MinLengthRule 最小长度规则
type MinLengthRule struct {
	minLength int
	message   string
}

func (r *MinLengthRule) Validate(value interface{}) error {
	if str, ok := value.(string); ok {
		if len(str) < r.minLength {
			return fmt.Errorf(r.message)
		}
	}
	return nil
}

func (r *MinLengthRule) GetMessage() string {
	return r.message
}

// MaxLengthRule 最大长度规则
type MaxLengthRule struct {
	maxLength int
	message   string
}

func (r *MaxLengthRule) Validate(value interface{}) error {
	if str, ok := value.(string); ok {
		if len(str) > r.maxLength {
			return fmt.Errorf(r.message)
		}
	}
	return nil
}

func (r *MaxLengthRule) GetMessage() string {
	return r.message
}

// RangeRule 范围规则
type RangeRule struct {
	min, max int
	message  string
}

func (r *RangeRule) Validate(value interface{}) error {
	switch v := value.(type) {
	case int:
		if v < r.min || v > r.max {
			return fmt.Errorf(r.message)
		}
	case int64:
		if v < int64(r.min) || v > int64(r.max) {
			return fmt.Errorf(r.message)
		}
	case float64:
		if v < float64(r.min) || v > float64(r.max) {
			return fmt.Errorf(r.message)
		}
	}
	return nil
}

func (r *RangeRule) GetMessage() string {
	return r.message
}

// DurationRule 持续时间规则
type DurationRule struct {
	min, max time.Duration
	message  string
}

func (r *DurationRule) Validate(value interface{}) error {
	if duration, ok := value.(time.Duration); ok {
		if duration < r.min || duration > r.max {
			return fmt.Errorf(r.message)
		}
	}
	return nil
}

func (r *DurationRule) GetMessage() string {
	return r.message
}

// EnumRule 枚举规则
type EnumRule struct {
	allowedValues []interface{}
	message       string
}

func (r *EnumRule) Validate(value interface{}) error {
	for _, allowed := range r.allowedValues {
		if reflect.DeepEqual(value, allowed) {
			return nil
		}
	}
	return fmt.Errorf(r.message)
}

func (r *EnumRule) GetMessage() string {
	return r.message
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make(map[string][]ValidationRule),
	}

	// 添加默认验证规则
	validator.addDefaultRules()

	return validator
}

// addDefaultRules 添加默认验证规则
func (validator *ConfigValidator) addDefaultRules() {
	// 服务器配置验证规则
	validator.AddRule("server.port", &RequiredRule{message: "服务器端口不能为空"})
	validator.AddRule("server.port", &MinLengthRule{minLength: 1, message: "服务器端口长度不能小于1"})

	// 数据库配置验证规则
	validator.AddRule("database.driver", &RequiredRule{message: "数据库驱动不能为空"})
	validator.AddRule("database.driver", &EnumRule{
		allowedValues: []interface{}{"mysql", "postgres", "sqlite"},
		message:       "数据库驱动必须是 mysql、postgres 或 sqlite",
	})
	validator.AddRule("database.max_open_conns", &RangeRule{min: 1, max: 1000, message: "最大打开连接数必须在1-1000之间"})
	validator.AddRule("database.max_idle_conns", &RangeRule{min: 0, max: 100, message: "最大空闲连接数必须在0-100之间"})

	// JWT配置验证规则
	validator.AddRule("jwt.secret", &RequiredRule{message: "JWT密钥不能为空"})
	validator.AddRule("jwt.secret", &MinLengthRule{minLength: 32, message: "JWT密钥长度不能小于32个字符"})

	// Redis配置验证规则
	validator.AddRule("redis.port", &RangeRule{min: 1, max: 65535, message: "Redis端口必须在1-65535之间"})
	validator.AddRule("redis.database", &RangeRule{min: 0, max: 15, message: "Redis数据库编号必须在0-15之间"})

	// 存储配置验证规则
	validator.AddRule("storage.max_file_size", &RangeRule{min: 1, max: 1000, message: "最大文件大小必须在1-1000MB之间"})

	// 监控配置验证规则
	validator.AddRule("monitoring.base.check_interval", &DurationRule{
		min:     time.Second,
		max:     24 * time.Hour,
		message: "检查间隔必须在1秒到24小时之间",
	})
}

// AddRule 添加验证规则
func (validator *ConfigValidator) AddRule(field string, rule ValidationRule) {
	validator.rules[field] = append(validator.rules[field], rule)
}

// Validate 验证配置
func (validator *ConfigValidator) Validate(config *Config) error {
	var errors []string

	// 验证各个配置字段
	errors = append(errors, validator.validateServerConfig(config.Server)...)
	errors = append(errors, validator.validateDatabaseConfig(config.Database)...)
	errors = append(errors, validator.validateJWTConfig(config.JWT)...)
	errors = append(errors, validator.validateRedisConfig(config.Redis)...)
	errors = append(errors, validator.validateStorageConfig(config.Storage)...)
	errors = append(errors, validator.validateEmailConfig(config.Email)...)
	errors = append(errors, validator.validateLogConfig(config.Log)...)
	errors = append(errors, validator.validateMonitoringConfig(config.Monitoring)...)

	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateServerConfig 验证服务器配置
func (validator *ConfigValidator) validateServerConfig(config ServerConfig) []string {
	var errors []string

	if err := validator.validateField("server.port", config.Port); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateDatabaseConfig 验证数据库配置
func (validator *ConfigValidator) validateDatabaseConfig(config DatabaseConfig) []string {
	var errors []string

	if err := validator.validateField("database.driver", config.Driver); err != nil {
		errors = append(errors, err.Error())
	}

	if err := validator.validateField("database.max_open_conns", config.MaxOpenConns); err != nil {
		errors = append(errors, err.Error())
	}

	if err := validator.validateField("database.max_idle_conns", config.MaxIdleConns); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateJWTConfig 验证JWT配置
func (validator *ConfigValidator) validateJWTConfig(config JWTConfig) []string {
	var errors []string

	if err := validator.validateField("jwt.secret", config.Secret); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateRedisConfig 验证Redis配置
func (validator *ConfigValidator) validateRedisConfig(config RedisConfig) []string {
	var errors []string

	if err := validator.validateField("redis.port", config.Port); err != nil {
		errors = append(errors, err.Error())
	}

	if err := validator.validateField("redis.database", config.Database); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateStorageConfig 验证存储配置
func (validator *ConfigValidator) validateStorageConfig(config StorageConfig) []string {
	var errors []string

	if err := validator.validateField("storage.max_file_size", config.MaxFileSize); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateEmailConfig 验证邮件配置
func (validator *ConfigValidator) validateEmailConfig(config EmailConfig) []string {
	var errors []string

	// 如果邮件配置已启用，则验证必填字段
	if config.IsConfigured() {
		if err := validator.validateField("email.host", config.Host); err != nil {
			errors = append(errors, err.Error())
		}

		if err := validator.validateField("email.port", config.Port); err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

// validateLogConfig 验证日志配置
func (validator *ConfigValidator) validateLogConfig(config LogConfig) []string {
	var errors []string

	// 验证日志级别
	validLevels := []interface{}{"debug", "info", "warning", "error", "fatal"}
	levelRule := &EnumRule{
		allowedValues: validLevels,
		message:       "日志级别必须是 debug、info、warning、error 或 fatal",
	}

	if err := levelRule.Validate(config.Level); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateMonitoringConfig 验证监控配置
func (validator *ConfigValidator) validateMonitoringConfig(config MonitoringConfig) []string {
	var errors []string

	if err := validator.validateField("monitoring.base.check_interval", config.BaseConfig.CheckInterval); err != nil {
		errors = append(errors, err.Error())
	}

	return errors
}

// validateField 验证字段
func (validator *ConfigValidator) validateField(field string, value interface{}) error {
	rules, exists := validator.rules[field]
	if !exists {
		return nil
	}

	for _, rule := range rules {
		if err := rule.Validate(value); err != nil {
			return err
		}
	}

	return nil
}

// 全局配置验证器
var globalConfigValidator *ConfigValidator

// GetConfigValidator 获取全局配置验证器
func GetConfigValidator() *ConfigValidator {
	if globalConfigValidator == nil {
		globalConfigValidator = NewConfigValidator()
	}
	return globalConfigValidator
}

// ValidateConfigWithValidator 使用验证器验证配置
func ValidateConfigWithValidator(config *Config) error {
	validator := GetConfigValidator()
	return validator.Validate(config)
}
