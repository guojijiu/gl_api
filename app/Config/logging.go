package Config

import (
	"fmt"
	"time"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// LogFormat 日志格式
type LogFormat string

const (
	LogFormatJSON   LogFormat = "json"
	LogFormatText   LogFormat = "text"
	LogFormatCustom LogFormat = "custom"
)

// LogRotation 日志轮转配置
type LogRotation struct {
	MaxSize    int           `mapstructure:"max_size"`    // 单个日志文件最大大小(MB)
	MaxAge     time.Duration `mapstructure:"max_age"`     // 日志文件保留时间
	MaxBackups int           `mapstructure:"max_backups"` // 保留的日志文件数量
	Compress   bool          `mapstructure:"compress"`    // 是否压缩旧日志文件
}

// LogConfig 日志配置结构
type LogConfig struct {
	// 全局配置
	Level      LogLevel  `mapstructure:"level"`      // 全局日志级别
	Format     LogFormat `mapstructure:"format"`     // 日志格式
	Output     string    `mapstructure:"output"`     // 输出方式: file, console, both
	Timestamp  bool      `mapstructure:"timestamp"`  // 是否包含时间戳
	Caller     bool      `mapstructure:"caller"`     // 是否包含调用者信息
	Stacktrace bool      `mapstructure:"stacktrace"` // 是否包含堆栈跟踪

	// 文件配置
	BasePath string      `mapstructure:"base_path"` // 日志基础路径
	Rotation LogRotation `mapstructure:"rotation"`  // 日志轮转配置

	// 各类型日志配置
	RequestLog  RequestLogConfig  `mapstructure:"request_log"`  // 请求日志配置
	SQLLog      SQLLogConfig      `mapstructure:"sql_log"`      // SQL日志配置
	ErrorLog    ErrorLogConfig    `mapstructure:"error_log"`    // 错误日志配置
	AuditLog    AuditLogConfig    `mapstructure:"audit_log"`    // 审计日志配置
	SecurityLog SecurityLogConfig `mapstructure:"security_log"` // 安全日志配置
	BusinessLog BusinessLogConfig `mapstructure:"business_log"` // 业务日志配置
	AccessLog   AccessLogConfig   `mapstructure:"access_log"`   // 访问日志配置
}

// RequestLogConfig 请求日志配置
type RequestLogConfig struct {
	Enabled     bool      `mapstructure:"enabled"`       // 是否启用
	Level       LogLevel  `mapstructure:"level"`         // 日志级别
	Path        string    `mapstructure:"path"`          // 存储路径(相对于base_path)
	Format      LogFormat `mapstructure:"format"`        // 日志格式
	IncludeBody bool      `mapstructure:"include_body"`  // 是否包含请求/响应体
	MaxBodySize int       `mapstructure:"max_body_size"` // 最大记录体大小(KB)
	FilterPaths []string  `mapstructure:"filter_paths"`  // 过滤的路径(不记录)
	MaskFields  []string  `mapstructure:"mask_fields"`   // 需要脱敏的字段
}

// SQLLogConfig SQL日志配置
type SQLLogConfig struct {
	Enabled       bool          `mapstructure:"enabled"`        // 是否启用
	Level         LogLevel      `mapstructure:"level"`          // 日志级别
	Path          string        `mapstructure:"path"`           // 存储路径
	Format        LogFormat     `mapstructure:"format"`         // 日志格式
	SlowThreshold time.Duration `mapstructure:"slow_threshold"` // 慢查询阈值
	IncludeParams bool          `mapstructure:"include_params"` // 是否包含SQL参数
	IncludeStack  bool          `mapstructure:"include_stack"`  // 是否包含调用栈
	MaxQuerySize  int           `mapstructure:"max_query_size"` // 最大SQL记录大小(KB)
}

// ErrorLogConfig 错误日志配置
type ErrorLogConfig struct {
	Enabled      bool      `mapstructure:"enabled"`       // 是否启用
	Level        LogLevel  `mapstructure:"level"`         // 日志级别
	Path         string    `mapstructure:"path"`          // 存储路径
	Format       LogFormat `mapstructure:"format"`        // 日志格式
	IncludeStack bool      `mapstructure:"include_stack"` // 是否包含堆栈跟踪
	NotifyEmail  string    `mapstructure:"notify_email"`  // 错误通知邮箱
	MaxErrors    int       `mapstructure:"max_errors"`    // 最大错误记录数
}

// AuditLogConfig 审计日志配置
type AuditLogConfig struct {
	Enabled   bool          `mapstructure:"enabled"`   // 是否启用
	Level     LogLevel      `mapstructure:"level"`     // 日志级别
	Path      string        `mapstructure:"path"`      // 存储路径
	Format    LogFormat     `mapstructure:"format"`    // 日志格式
	Retention time.Duration `mapstructure:"retention"` // 保留时间
	Encrypt   bool          `mapstructure:"encrypt"`   // 是否加密存储
	Compress  bool          `mapstructure:"compress"`  // 是否压缩存储
}

// SecurityLogConfig 安全日志配置
type SecurityLogConfig struct {
	Enabled     bool      `mapstructure:"enabled"`      // 是否启用
	Level       LogLevel  `mapstructure:"level"`        // 日志级别
	Path        string    `mapstructure:"path"`         // 存储路径
	Format      LogFormat `mapstructure:"format"`       // 日志格式
	IncludeIP   bool      `mapstructure:"include_ip"`   // 是否包含IP地址
	IncludeUser bool      `mapstructure:"include_user"` // 是否包含用户信息
	AlertLevel  LogLevel  `mapstructure:"alert_level"`  // 告警级别
	RealTime    bool      `mapstructure:"real_time"`    // 是否实时告警
}

// BusinessLogConfig 业务日志配置
type BusinessLogConfig struct {
	Enabled bool      `mapstructure:"enabled"` // 是否启用
	Level   LogLevel  `mapstructure:"level"`   // 日志级别
	Path    string    `mapstructure:"path"`    // 存储路径
	Format  LogFormat `mapstructure:"format"`  // 日志格式
	Modules []string  `mapstructure:"modules"` // 启用的业务模块
}

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Enabled     bool      `mapstructure:"enabled"`      // 是否启用
	Level       LogLevel  `mapstructure:"level"`        // 日志级别
	Path        string    `mapstructure:"path"`         // 存储路径
	Format      LogFormat `mapstructure:"format"`       // 日志格式
	IncludeUser bool      `mapstructure:"include_user"` // 是否包含用户信息
	IncludeIP   bool      `mapstructure:"include_ip"`   // 是否包含IP地址
	IncludeUA   bool      `mapstructure:"include_ua"`   // 是否包含User-Agent
}

// SetDefaults 设置默认值
func (c *LogConfig) SetDefaults() {
	c.Level = LogLevelInfo
	c.Format = LogFormatJSON
	c.Output = "file" // 只输出到文件，去掉控制台输出
	c.Timestamp = true
	c.Caller = true
	c.Stacktrace = false
	c.BasePath = "./storage/logs"

	// 设置轮转默认值 - 按天记录
	c.Rotation.MaxSize = 0             // 0表示不按大小轮转，只按时间轮转
	c.Rotation.MaxAge = 24 * time.Hour // 1天
	c.Rotation.MaxBackups = 30         // 保留30天的日志文件
	c.Rotation.Compress = true

	// 设置各类型日志默认值
	c.RequestLog.SetDefaults()
	c.SQLLog.SetDefaults()
	c.ErrorLog.SetDefaults()
	c.AuditLog.SetDefaults()
	c.SecurityLog.SetDefaults()
	c.BusinessLog.SetDefaults()
	c.AccessLog.SetDefaults()
}

// BindEnvs 绑定环境变量
func (c *LogConfig) BindEnvs() {
	// 全局配置
	bindEnv("LOG_LEVEL", &c.Level)
	bindEnv("LOG_FORMAT", &c.Format)
	bindEnv("LOG_OUTPUT", &c.Output)
	bindEnv("LOG_TIMESTAMP", &c.Timestamp)
	bindEnv("LOG_CALLER", &c.Caller)
	bindEnv("LOG_STACKTRACE", &c.Stacktrace)
	bindEnv("LOG_BASE_PATH", &c.BasePath)

	// 轮转配置
	bindEnv("LOG_MAX_SIZE", &c.Rotation.MaxSize)
	bindEnv("LOG_MAX_AGE", &c.Rotation.MaxAge)
	bindEnv("LOG_MAX_BACKUPS", &c.Rotation.MaxBackups)
	bindEnv("LOG_COMPRESS", &c.Rotation.Compress)

	// 各类型日志配置
	c.RequestLog.BindEnvs("REQUEST_LOG")
	c.SQLLog.BindEnvs("SQL_LOG")
	c.ErrorLog.BindEnvs("ERROR_LOG")
	c.AuditLog.BindEnvs("AUDIT_LOG")
	c.SecurityLog.BindEnvs("SECURITY_LOG")
	c.BusinessLog.BindEnvs("BUSINESS_LOG")
	c.AccessLog.BindEnvs("ACCESS_LOG")
}

// Validate 验证配置
func (c *LogConfig) Validate() error {
	// 验证日志级别
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}

	// 验证日志格式
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}

	// 验证输出方式
	if c.Output != "file" && c.Output != "console" && c.Output != "both" {
		return fmt.Errorf("无效的输出方式: %s", c.Output)
	}

	// 验证各类型日志配置
	if err := c.RequestLog.Validate(); err != nil {
		return fmt.Errorf("请求日志配置错误: %v", err)
	}
	if err := c.SQLLog.Validate(); err != nil {
		return fmt.Errorf("SQL日志配置错误: %v", err)
	}
	if err := c.ErrorLog.Validate(); err != nil {
		return fmt.Errorf("错误日志配置错误: %v", err)
	}
	if err := c.AuditLog.Validate(); err != nil {
		return fmt.Errorf("审计日志配置错误: %v", err)
	}
	if err := c.SecurityLog.Validate(); err != nil {
		return fmt.Errorf("安全日志配置错误: %v", err)
	}
	if err := c.BusinessLog.Validate(); err != nil {
		return fmt.Errorf("业务日志配置错误: %v", err)
	}
	if err := c.AccessLog.Validate(); err != nil {
		return fmt.Errorf("访问日志配置错误: %v", err)
	}

	return nil
}

// SetDefaults 设置请求日志默认值
func (c *RequestLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelInfo
	c.Path = "requests" // 日志文件将自动添加日期后缀，如 requests-2025-01-20.log
	c.Format = LogFormatJSON
	c.IncludeBody = false
	c.MaxBodySize = 1024 // 1KB
	c.FilterPaths = []string{"/health", "/metrics"}
	c.MaskFields = []string{"password", "token", "secret"}
}

// BindEnvs 绑定请求日志环境变量
func (c *RequestLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_INCLUDE_BODY", &c.IncludeBody)
	bindEnv(prefix+"_MAX_BODY_SIZE", &c.MaxBodySize)
	bindEnv(prefix+"_FILTER_PATHS", &c.FilterPaths)
	bindEnv(prefix+"_MASK_FIELDS", &c.MaskFields)
}

// Validate 验证请求日志配置
func (c *RequestLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// SetDefaults 设置SQL日志默认值
func (c *SQLLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelInfo
	c.Path = "sql" // 日志文件将自动添加日期后缀，如 sql-2025-01-20.log
	c.Format = LogFormatJSON
	c.SlowThreshold = 1 * time.Second
	c.IncludeParams = true
	c.IncludeStack = false
	c.MaxQuerySize = 2048 // 2KB
}

// BindEnvs 绑定SQL日志环境变量
func (c *SQLLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_SLOW_THRESHOLD", &c.SlowThreshold)
	bindEnv(prefix+"_INCLUDE_PARAMS", &c.IncludeParams)
	bindEnv(prefix+"_INCLUDE_STACK", &c.IncludeStack)
	bindEnv(prefix+"_MAX_QUERY_SIZE", &c.MaxQuerySize)
}

// Validate 验证SQL日志配置
func (c *SQLLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// SetDefaults 设置错误日志默认值
func (c *ErrorLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelError
	c.Path = "errors" // 日志文件将自动添加日期后缀，如 errors-2025-01-20.log
	c.Format = LogFormatJSON
	c.IncludeStack = true
	c.NotifyEmail = ""
	c.MaxErrors = 10000
}

// BindEnvs 绑定错误日志环境变量
func (c *ErrorLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_INCLUDE_STACK", &c.IncludeStack)
	bindEnv(prefix+"_NOTIFY_EMAIL", &c.NotifyEmail)
	bindEnv(prefix+"_MAX_ERRORS", &c.MaxErrors)
}

// Validate 验证错误日志配置
func (c *ErrorLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// SetDefaults 设置审计日志默认值
func (c *AuditLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelInfo
	c.Path = "audit" // 日志文件将自动添加日期后缀，如 audit-2025-01-20.log
	c.Format = LogFormatJSON
	c.Retention = 365 * 24 * time.Hour // 1年
	c.Encrypt = false
	c.Compress = true
}

// BindEnvs 绑定审计日志环境变量
func (c *AuditLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_RETENTION", &c.Retention)
	bindEnv(prefix+"_ENCRYPT", &c.Encrypt)
	bindEnv(prefix+"_COMPRESS", &c.Compress)
}

// Validate 验证审计日志配置
func (c *AuditLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// SetDefaults 设置安全日志默认值
func (c *SecurityLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelWarning
	c.Path = "security" // 日志文件将自动添加日期后缀，如 security-2025-01-20.log
	c.Format = LogFormatJSON
	c.IncludeIP = true
	c.IncludeUser = true
	c.AlertLevel = LogLevelError
	c.RealTime = true
}

// BindEnvs 绑定安全日志环境变量
func (c *SecurityLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_INCLUDE_IP", &c.IncludeIP)
	bindEnv(prefix+"_INCLUDE_USER", &c.IncludeUser)
	bindEnv(prefix+"_ALERT_LEVEL", &c.AlertLevel)
	bindEnv(prefix+"_REAL_TIME", &c.RealTime)
}

// Validate 验证安全日志配置
func (c *SecurityLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	if !isValidLogLevel(c.AlertLevel) {
		return fmt.Errorf("无效的告警级别: %s", c.AlertLevel)
	}
	return nil
}

// SetDefaults 设置业务日志默认值
func (c *BusinessLogConfig) SetDefaults() {
	c.Enabled = true // 启用业务日志记录
	c.Level = LogLevelInfo
	c.Path = "business" // 日志文件将自动添加日期后缀，如 business-2025-01-20.log
	c.Format = LogFormatJSON
	c.Modules = []string{"user", "post", "category", "tag"}
}

// BindEnvs 绑定业务日志环境变量
func (c *BusinessLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
}

// Validate 验证业务日志配置
func (c *BusinessLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// SetDefaults 设置访问日志默认值
func (c *AccessLogConfig) SetDefaults() {
	c.Enabled = true
	c.Level = LogLevelInfo
	c.Path = "access" // 日志文件将自动添加日期后缀，如 access-2025-01-20.log
	c.Format = LogFormatJSON
	c.IncludeUser = true
	c.IncludeIP = true
	c.IncludeUA = true
}

// BindEnvs 绑定访问日志环境变量
func (c *AccessLogConfig) BindEnvs(prefix string) {
	bindEnv(prefix+"_ENABLED", &c.Enabled)
	bindEnv(prefix+"_LEVEL", &c.Level)
	bindEnv(prefix+"_PATH", &c.Path)
	bindEnv(prefix+"_FORMAT", &c.Format)
	bindEnv(prefix+"_INCLUDE_USER", &c.IncludeUser)
	bindEnv(prefix+"_INCLUDE_IP", &c.IncludeIP)
	bindEnv(prefix+"_INCLUDE_UA", &c.IncludeUA)
}

// Validate 验证访问日志配置
func (c *AccessLogConfig) Validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("无效的日志格式: %s", c.Format)
	}
	return nil
}

// 辅助函数
func isValidLogLevel(level LogLevel) bool {
	switch level {
	case LogLevelDebug, LogLevelInfo, LogLevelWarning, LogLevelError, LogLevelFatal:
		return true
	default:
		return false
	}
}

func isValidLogFormat(format LogFormat) bool {
	switch format {
	case LogFormatJSON, LogFormatText, LogFormatCustom:
		return true
	default:
		return false
	}
}
