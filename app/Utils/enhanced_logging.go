package Utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogContext 日志上下文
type LogContext struct {
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Method    string                 `json:"method,omitempty"`
	URL       string                 `json:"url,omitempty"`
	Status    int                    `json:"status,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Context     *LogContext            `json:"context,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Source      string                 `json:"source,omitempty"`
	GoroutineID string                 `json:"goroutine_id,omitempty"`
}

// EnhancedLogger 增强的日志记录器
type EnhancedLogger struct {
	level       LogLevel
	context     *LogContext
	fields      map[string]interface{}
	output      *log.Logger
	enableJSON  bool
	enableColor bool
}

// LoggerConfig 日志器配置
type LoggerConfig struct {
	Level       LogLevel `json:"level"`
	EnableJSON  bool     `json:"enable_json"`
	EnableColor bool     `json:"enable_color"`
	Output      string   `json:"output"` // stdout, stderr, file
	FilePath    string   `json:"file_path,omitempty"`
}

// NewEnhancedLogger 创建增强日志记录器
func NewEnhancedLogger(config *LoggerConfig) *EnhancedLogger {
	if config == nil {
		config = &LoggerConfig{
			Level:       LogLevelInfo,
			EnableJSON:  true,
			EnableColor: false,
			Output:      "stdout",
		}
	}

	var output *log.Logger
	switch config.Output {
	case "stderr":
		output = log.New(os.Stderr, "", 0)
	case "file":
		if config.FilePath != "" {
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Printf("Failed to open log file: %v", err)
				output = log.New(os.Stdout, "", 0)
			} else {
				output = log.New(file, "", 0)
			}
		} else {
			output = log.New(os.Stdout, "", 0)
		}
	default:
		output = log.New(os.Stdout, "", 0)
	}

	return &EnhancedLogger{
		level:       config.Level,
		context:     &LogContext{},
		fields:      make(map[string]interface{}),
		output:      output,
		enableJSON:  config.EnableJSON,
		enableColor: config.EnableColor,
	}
}

// WithContext 设置上下文
func (l *EnhancedLogger) WithContext(ctx *LogContext) *EnhancedLogger {
	newLogger := *l
	newLogger.context = ctx
	return &newLogger
}

// WithField 添加字段
func (l *EnhancedLogger) WithField(key string, value interface{}) *EnhancedLogger {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return &newLogger
}

// WithFields 添加多个字段
func (l *EnhancedLogger) WithFields(fields map[string]interface{}) *EnhancedLogger {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return &newLogger
}

// WithUserID 设置用户ID
func (l *EnhancedLogger) WithUserID(userID string) *EnhancedLogger {
	newLogger := *l
	if newLogger.context == nil {
		newLogger.context = &LogContext{}
	}
	newLogger.context.UserID = userID
	return &newLogger
}

// WithRequestID 设置请求ID
func (l *EnhancedLogger) WithRequestID(requestID string) *EnhancedLogger {
	newLogger := *l
	if newLogger.context == nil {
		newLogger.context = &LogContext{}
	}
	newLogger.context.RequestID = requestID
	return &newLogger
}

// WithContextValue 从context.Context中提取值
func (l *EnhancedLogger) WithContextValue(ctx context.Context) *EnhancedLogger {
	newLogger := *l
	if newLogger.context == nil {
		newLogger.context = &LogContext{}
	}

	// 提取常用值
	if userID, ok := ctx.Value("user_id").(string); ok {
		newLogger.context.UserID = userID
	}
	if requestID, ok := ctx.Value("request_id").(string); ok {
		newLogger.context.RequestID = requestID
	}
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		newLogger.context.SessionID = sessionID
	}
	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		newLogger.context.IPAddress = ipAddress
	}
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		newLogger.context.UserAgent = userAgent
	}

	return &newLogger
}

// Debug 记录调试日志
func (l *EnhancedLogger) Debug(message string) {
	l.log(LogLevelDebug, message, nil)
}

// Debugf 记录格式化调试日志
func (l *EnhancedLogger) Debugf(format string, args ...interface{}) {
	l.log(LogLevelDebug, fmt.Sprintf(format, args...), nil)
}

// Info 记录信息日志
func (l *EnhancedLogger) Info(message string) {
	l.log(LogLevelInfo, message, nil)
}

// Infof 记录格式化信息日志
func (l *EnhancedLogger) Infof(format string, args ...interface{}) {
	l.log(LogLevelInfo, fmt.Sprintf(format, args...), nil)
}

// Warn 记录警告日志
func (l *EnhancedLogger) Warn(message string) {
	l.log(LogLevelWarn, message, nil)
}

// Warnf 记录格式化警告日志
func (l *EnhancedLogger) Warnf(format string, args ...interface{}) {
	l.log(LogLevelWarn, fmt.Sprintf(format, args...), nil)
}

// Error 记录错误日志
func (l *EnhancedLogger) Error(message string) {
	l.log(LogLevelError, message, nil)
}

// Errorf 记录格式化错误日志
func (l *EnhancedLogger) Errorf(format string, args ...interface{}) {
	l.log(LogLevelError, fmt.Sprintf(format, args...), nil)
}

// Fatal 记录致命错误日志并退出
func (l *EnhancedLogger) Fatal(message string) {
	l.log(LogLevelFatal, message, nil)
	os.Exit(1)
}

// Fatalf 记录格式化致命错误日志并退出
func (l *EnhancedLogger) Fatalf(format string, args ...interface{}) {
	l.log(LogLevelFatal, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

// LogError 记录增强错误
func (l *EnhancedLogger) LogError(err *EnhancedError) {
	if err == nil {
		return
	}

	fields := map[string]interface{}{
		"error_code":     err.Code,
		"error_status":   err.Status,
		"error_category": err.Category,
		"error_severity": err.Severity,
		"recoverable":    err.Recoverable,
		"retryable":      err.Retryable,
	}

	if err.Details != "" {
		fields["error_details"] = err.Details
	}

	if err.UserID != "" {
		fields["user_id"] = err.UserID
	}

	if err.RequestID != "" {
		fields["request_id"] = err.RequestID
	}

	if err.Context != nil {
		for k, v := range err.Context {
			fields["error_context_"+k] = v
		}
	}

	if err.StackTrace != "" {
		fields["stack_trace"] = err.StackTrace
	}

	if err.Source != "" {
		fields["error_source"] = err.Source
	}

	// 根据错误严重程度确定日志级别
	var logLevel LogLevel
	switch err.Severity {
	case SeverityCritical, SeverityHigh:
		logLevel = LogLevelError
	case SeverityMedium:
		logLevel = LogLevelWarn
	case SeverityLow:
		logLevel = LogLevelInfo
	default:
		logLevel = LogLevelInfo
	}

	l.log(logLevel, err.Message, fields)
}

// log 内部日志记录方法
func (l *EnhancedLogger) log(level LogLevel, message string, fields map[string]interface{}) {
	// 检查日志级别
	if !l.shouldLog(level) {
		return
	}

	// 创建日志条目
	entry := &LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		Context:     l.context,
		Fields:      make(map[string]interface{}),
		GoroutineID: getGoroutineID(),
	}

	// 合并字段
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	for k, v := range fields {
		entry.Fields[k] = v
	}

	// 添加调用者信息
	if pc, file, line, ok := runtime.Caller(3); ok {
		funcName := runtime.FuncForPC(pc).Name()
		entry.Source = fmt.Sprintf("%s:%d %s", file, line, funcName)
	}

	// 对于错误级别，添加堆栈跟踪
	if level == LogLevelError || level == LogLevelFatal {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		entry.StackTrace = string(buf[:n])
	}

	// 输出日志
	l.outputLog(entry)
}

// shouldLog 检查是否应该记录日志
func (l *EnhancedLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levels[level] >= levels[l.level]
}

// outputLog 输出日志
func (l *EnhancedLogger) outputLog(entry *LogEntry) {
	if l.enableJSON {
		l.outputJSON(entry)
	} else {
		l.outputText(entry)
	}
}

// outputJSON 输出JSON格式日志
func (l *EnhancedLogger) outputJSON(entry *LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		l.output.Printf("Failed to marshal log entry: %v", err)
		return
	}
	l.output.Println(string(data))
}

// outputText 输出文本格式日志
func (l *EnhancedLogger) outputText(entry *LogEntry) {
	var color string
	if l.enableColor {
		color = l.getColor(entry.Level)
	}

	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(string(entry.Level))

	// 基础日志格式
	format := "%s [%s] %s"
	args := []interface{}{timestamp, level, entry.Message}

	// 添加颜色
	if color != "" {
		format = "%s" + format + "\033[0m"
		args = append([]interface{}{color}, args...)
	}

	// 添加上下文信息
	if entry.Context != nil {
		if entry.Context.UserID != "" {
			format += " user_id=%s"
			args = append(args, entry.Context.UserID)
		}
		if entry.Context.RequestID != "" {
			format += " request_id=%s"
			args = append(args, entry.Context.RequestID)
		}
	}

	// 添加字段
	for k, v := range entry.Fields {
		format += " %s=%v"
		args = append(args, k, v)
	}

	// 添加来源信息
	if entry.Source != "" {
		format += " source=%s"
		args = append(args, entry.Source)
	}

	l.output.Printf(format, args...)
}

// getColor 获取颜色代码
func (l *EnhancedLogger) getColor(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "\033[36m" // 青色
	case LogLevelInfo:
		return "\033[32m" // 绿色
	case LogLevelWarn:
		return "\033[33m" // 黄色
	case LogLevelError:
		return "\033[31m" // 红色
	case LogLevelFatal:
		return "\033[35m" // 紫色
	default:
		return ""
	}
}

// 全局日志记录器
var (
	DefaultLogger *EnhancedLogger
)

// 初始化默认日志记录器
func init() {
	DefaultLogger = NewEnhancedLogger(&LoggerConfig{
		Level:       LogLevelInfo,
		EnableJSON:  true,
		EnableColor: false,
		Output:      "stdout",
	})
}

// 全局日志函数
func Debug(message string) {
	DefaultLogger.Debug(message)
}

func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

func Info(message string) {
	DefaultLogger.Info(message)
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

func Warn(message string) {
	DefaultLogger.Warn(message)
}

func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

func Error(message string) {
	DefaultLogger.Error(message)
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

func Fatal(message string) {
	DefaultLogger.Fatal(message)
}

func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

// LogError 记录增强错误
func LogError(err *EnhancedError) {
	DefaultLogger.LogError(err)
}

// WithContext 创建带上下文的日志记录器
func WithContext(ctx context.Context) *EnhancedLogger {
	return DefaultLogger.WithContextValue(ctx)
}

// WithField 创建带字段的日志记录器
func WithField(key string, value interface{}) *EnhancedLogger {
	return DefaultLogger.WithField(key, value)
}

// WithFields 创建带多个字段的日志记录器
func WithFields(fields map[string]interface{}) *EnhancedLogger {
	return DefaultLogger.WithFields(fields)
}
