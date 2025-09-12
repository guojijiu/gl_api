package Utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

// LogLevelString 日志级别字符串映射
var LogLevelString = map[LogLevel]string{
	LogLevelDebug:   "DEBUG",
	LogLevelInfo:    "INFO",
	LogLevelWarning: "WARNING",
	LogLevelError:   "ERROR",
	LogLevelFatal:   "FATAL",
}

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Module    string                 `json:"module,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Context   context.Context        `json:"-"`
	Caller    string                 `json:"caller,omitempty"`
	Stack     []string               `json:"stack,omitempty"`
}

// EnhancedLogger 增强日志记录器
type EnhancedLogger struct {
	level        LogLevel
	output       io.Writer
	enableCaller bool
	enableStack  bool
	formatter    LogFormatter
}

// LogFormatter 日志格式化器接口
type LogFormatter interface {
	Format(entry *LogEntry) ([]byte, error)
}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

// Format 格式化日志条目为JSON
func (f *JSONFormatter) Format(entry *LogEntry) ([]byte, error) {
	// 创建格式化结构
	formatted := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"level":     LogLevelString[entry.Level],
		"message":   entry.Message,
	}

	if entry.Module != "" {
		formatted["module"] = entry.Module
	}
	if entry.Action != "" {
		formatted["action"] = entry.Action
	}
	if entry.Caller != "" {
		formatted["caller"] = entry.Caller
	}
	if len(entry.Fields) > 0 {
		formatted["fields"] = entry.Fields
	}
	if len(entry.Stack) > 0 {
		formatted["stack"] = entry.Stack
	}

	return json.Marshal(formatted)
}

// TextFormatter 文本格式化器
type TextFormatter struct{}

// Format 格式化日志条目为文本
func (f *TextFormatter) Format(entry *LogEntry) ([]byte, error) {
	var parts []string

	// 时间戳
	parts = append(parts, entry.Timestamp.Format("2006-01-02 15:04:05"))

	// 级别
	parts = append(parts, fmt.Sprintf("[%s]", LogLevelString[entry.Level]))

	// 调用者
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Caller))
	}

	// 模块和动作
	if entry.Module != "" && entry.Action != "" {
		parts = append(parts, fmt.Sprintf("[%s:%s]", entry.Module, entry.Action))
	} else if entry.Module != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Module))
	}

	// 消息
	parts = append(parts, entry.Message)

	// 字段
	if len(entry.Fields) > 0 {
		fields := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(fields, ", ")))
	}

	// 堆栈跟踪
	if len(entry.Stack) > 0 {
		parts = append(parts, "\nStack trace:")
		parts = append(parts, strings.Join(entry.Stack, "\n"))
	}

	return []byte(strings.Join(parts, " ") + "\n"), nil
}

// NewEnhancedLogger 创建增强日志记录器
func NewEnhancedLogger() *EnhancedLogger {
	return &EnhancedLogger{
		level:        LogLevelInfo,
		output:       os.Stdout,
		enableCaller: true,
		enableStack:  false,
		formatter:    &JSONFormatter{},
	}
}

// SetLevel 设置日志级别
func (l *EnhancedLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput 设置输出目标
func (l *EnhancedLogger) SetOutput(output io.Writer) {
	l.output = output
}

// SetFormatter 设置格式化器
func (l *EnhancedLogger) SetFormatter(formatter LogFormatter) {
	l.formatter = formatter
}

// EnableCaller 启用调用者信息
func (l *EnhancedLogger) EnableCaller(enable bool) {
	l.enableCaller = enable
}

// EnableStack 启用堆栈跟踪
func (l *EnhancedLogger) EnableStack(enable bool) {
	l.enableStack = enable
}

// log 记录日志
func (l *EnhancedLogger) log(level LogLevel, message string, module string, action string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Module:    module,
		Action:    action,
		Fields:    fields,
	}

	// 添加调用者信息
	if l.enableCaller {
		entry.Caller = l.getCaller()
	}

	// 添加堆栈跟踪（仅对错误级别）
	if l.enableStack && level >= LogLevelError {
		entry.Stack = l.getStackTrace()
	}

	// 格式化并输出
	formatted, err := l.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format log entry: %v\n", err)
		return
	}

	l.output.Write(formatted)
}

// Debug 记录调试日志
func (l *EnhancedLogger) Debug(message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.log(LogLevelDebug, message, "", "", mergedFields)
}

// Info 记录信息日志
func (l *EnhancedLogger) Info(message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.log(LogLevelInfo, message, "", "", mergedFields)
}

// Warning 记录警告日志
func (l *EnhancedLogger) Warning(message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.log(LogLevelWarning, message, "", "", mergedFields)
}

// Error 记录错误日志
func (l *EnhancedLogger) Error(message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.log(LogLevelError, message, "", "", mergedFields)
}

// Fatal 记录致命错误日志
func (l *EnhancedLogger) Fatal(message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.log(LogLevelFatal, message, "", "", mergedFields)
	os.Exit(1)
}

// WithModule 设置模块
func (l *EnhancedLogger) WithModule(module string) *ModuleLogger {
	return &ModuleLogger{
		logger: l,
		module: module,
	}
}

// WithFields 设置字段
func (l *EnhancedLogger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// ModuleLogger 模块日志记录器
type ModuleLogger struct {
	logger *EnhancedLogger
	module string
}

// Debug 记录调试日志
func (ml *ModuleLogger) Debug(message string, action string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	ml.logger.log(LogLevelDebug, message, ml.module, action, mergedFields)
}

// Info 记录信息日志
func (ml *ModuleLogger) Info(message string, action string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	ml.logger.log(LogLevelInfo, message, ml.module, action, mergedFields)
}

// Warning 记录警告日志
func (ml *ModuleLogger) Warning(message string, action string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	ml.logger.log(LogLevelWarning, message, ml.module, action, mergedFields)
}

// Error 记录错误日志
func (ml *ModuleLogger) Error(message string, action string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	ml.logger.log(LogLevelError, message, ml.module, action, mergedFields)
}

// Fatal 记录致命错误日志
func (ml *ModuleLogger) Fatal(message string, action string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	ml.logger.log(LogLevelFatal, message, ml.module, action, mergedFields)
	os.Exit(1)
}

// FieldLogger 字段日志记录器
type FieldLogger struct {
	logger *EnhancedLogger
	fields map[string]interface{}
}

// Debug 记录调试日志
func (fl *FieldLogger) Debug(message string, fields ...map[string]interface{}) {
	mergedFields := fl.mergeFields(fields...)
	fl.logger.log(LogLevelDebug, message, "", "", mergedFields)
}

// Info 记录信息日志
func (fl *FieldLogger) Info(message string, fields ...map[string]interface{}) {
	mergedFields := fl.mergeFields(fields...)
	fl.logger.log(LogLevelInfo, message, "", "", mergedFields)
}

// Warning 记录警告日志
func (fl *FieldLogger) Warning(message string, fields ...map[string]interface{}) {
	mergedFields := fl.mergeFields(fields...)
	fl.logger.log(LogLevelWarning, message, "", "", mergedFields)
}

// Error 记录错误日志
func (fl *FieldLogger) Error(message string, fields ...map[string]interface{}) {
	mergedFields := fl.mergeFields(fields...)
	fl.logger.log(LogLevelError, message, "", "", mergedFields)
}

// Fatal 记录致命错误日志
func (fl *FieldLogger) Fatal(message string, fields ...map[string]interface{}) {
	mergedFields := fl.mergeFields(fields...)
	fl.logger.log(LogLevelFatal, message, "", "", mergedFields)
	os.Exit(1)
}

// mergeFields 合并字段
func (fl *FieldLogger) mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// 复制基础字段
	for k, v := range fl.fields {
		merged[k] = v
	}

	// 合并额外字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			merged[k] = v
		}
	}

	return merged
}

// getCaller 获取调用者信息
func (l *EnhancedLogger) getCaller() string {
	_, file, line, ok := runtime.Caller(3) // 跳过log方法、模块方法、实际调用
	if !ok {
		return ""
	}

	// 只保留文件名，不包含路径
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// getStackTrace 获取堆栈跟踪
func (l *EnhancedLogger) getStackTrace() []string {
	var stack []string
	for i := 3; i < 10; i++ { // 跳过当前函数和调用函数
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", filepath.Base(file), line))
	}
	return stack
}

// 全局日志记录器实例
var GlobalLogger = NewEnhancedLogger()

// 便捷函数
func Debug(message string, fields ...map[string]interface{}) {
	GlobalLogger.Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	GlobalLogger.Info(message, fields...)
}

func Warning(message string, fields ...map[string]interface{}) {
	GlobalLogger.Warning(message, fields...)
}

func Error(message string, fields ...map[string]interface{}) {
	GlobalLogger.Error(message, fields...)
}

func Fatal(message string, fields ...map[string]interface{}) {
	GlobalLogger.Fatal(message, fields...)
}

// 业务日志记录器
type BusinessLogger struct {
	*EnhancedLogger
}

// NewBusinessLogger 创建业务日志记录器
func NewBusinessLogger() *BusinessLogger {
	return &BusinessLogger{
		EnhancedLogger: NewEnhancedLogger(),
	}
}

// LogUserAction 记录用户操作
func (bl *BusinessLogger) LogUserAction(userID uint, action string, resource string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"user_id":  userID,
		"action":   action,
		"resource": resource,
	}

	// 合并详细信息
	for k, v := range details {
		fields[k] = v
	}

	bl.Info("用户操作", map[string]interface{}{"action": "user_action", "data": fields})
}

// LogSystemEvent 记录系统事件
func (bl *BusinessLogger) LogSystemEvent(event string, details map[string]interface{}) {
	bl.Info("系统事件", map[string]interface{}{"event": "system_event", "data": map[string]interface{}{
		"event":   event,
		"details": details,
	}})
}

// LogSecurityEvent 记录安全事件
func (bl *BusinessLogger) LogSecurityEvent(event string, level LogLevel, details map[string]interface{}) {
	fields := map[string]interface{}{
		"event":   event,
		"details": details,
	}

	securityFields := map[string]interface{}{"event": "security_event", "data": fields}

	switch level {
	case LogLevelDebug:
		bl.Debug("安全事件", securityFields)
	case LogLevelInfo:
		bl.Info("安全事件", securityFields)
	case LogLevelWarning:
		bl.Warning("安全事件", securityFields)
	case LogLevelError:
		bl.Error("安全事件", securityFields)
	case LogLevelFatal:
		bl.Fatal("安全事件", securityFields)
	}
}
