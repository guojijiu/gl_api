package Utils

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorLevel 错误级别
type ErrorLevel int

const (
	ErrorLevelDebug ErrorLevel = iota
	ErrorLevelInfo
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelFatal
)

// ErrorLevelString 错误级别字符串映射
var ErrorLevelString = map[ErrorLevel]string{
	ErrorLevelDebug:   "DEBUG",
	ErrorLevelInfo:    "INFO",
	ErrorLevelWarning: "WARNING",
	ErrorLevelError:   "ERROR",
	ErrorLevelFatal:   "FATAL",
}

// EnhancedError 增强错误结构
type EnhancedError struct {
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	Level       ErrorLevel             `json:"level"`
	Timestamp   time.Time              `json:"timestamp"`
	Stack       []string               `json:"stack,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	UserMessage string                 `json:"user_message,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Category    string                 `json:"category"`
	Status      int                    `json:"status"`
	Details     string                 `json:"details,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	return e.Message
}

// String 返回错误字符串表示
func (e *EnhancedError) String() string {
	return fmt.Sprintf("[%s] %s: %s", ErrorLevelString[e.Level], e.Code, e.Message)
}

// WithContext 添加上下文信息
func (e *EnhancedError) WithContext(key string, value interface{}) *EnhancedError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithUserMessage 设置用户友好的错误消息
func (e *EnhancedError) WithUserMessage(message string) *EnhancedError {
	e.UserMessage = message
	return e
}

// SetRetryable 设置是否可重试
func (e *EnhancedError) SetRetryable(retryable bool) *EnhancedError {
	e.Retryable = retryable
	return e
}

// SetCategory 设置错误分类
func (e *EnhancedError) SetCategory(category string) *EnhancedError {
	e.Category = category
	return e
}

// WithContextValue 添加上下文值
func (e *EnhancedError) WithContextValue(key string, value interface{}) *EnhancedError {
	return e.WithContext(key, value)
}

// WithStackTrace 添加堆栈跟踪
func (e *EnhancedError) WithStackTrace() *EnhancedError {
	if e.Stack == nil {
		e.Stack = make([]string, 0)
	}

	// 获取调用堆栈
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	stack := strings.Split(string(buf), "\n")
	e.Stack = append(e.Stack, stack...)
	return e
}

// EnhancedErrorHandler 增强错误处理器
type EnhancedErrorHandler struct {
	enableStack   bool
	enableContext bool
}

// NewEnhancedErrorHandler 创建增强错误处理器
func NewEnhancedErrorHandler() *EnhancedErrorHandler {
	return &EnhancedErrorHandler{
		enableStack:   true,
		enableContext: true,
	}
}

// 全局错误处理器实例
var globalErrorHandler = NewEnhancedErrorHandler()

// 错误分类常量
const (
	CategorySystem     = "system"
	CategoryUser       = "user"
	CategoryDatabase   = "database"
	CategoryNetwork    = "network"
	CategoryAuth       = "auth"
	CategoryValidation = "validation"
)

// 错误严重程度常量
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// 错误分类常量（兼容性）
const (
	ErrorCategorySystem     = CategorySystem
	ErrorCategoryUser       = CategoryUser
	ErrorCategoryDatabase   = CategoryDatabase
	ErrorCategoryNetwork    = CategoryNetwork
	ErrorCategoryAuth       = CategoryAuth
	ErrorCategoryValidation = CategoryValidation
)

// WithFields 创建带字段的错误（全局函数）
func WithFields(message string, fields map[string]interface{}) *EnhancedError {
	return globalErrorHandler.NewError(message, "GENERAL", ErrorLevelError).WithContext("fields", fields)
}

// NewErrorBuilder 创建错误构建器（全局函数）
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		handler: globalErrorHandler,
	}
}

// LogError 记录错误（全局函数）
func LogError(err error, context map[string]interface{}) {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		enhancedErr.WithContext("log_context", context)
		// 这里可以集成日志服务
		fmt.Printf("Error: %s\n", enhancedErr.String())
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

// GetEnhancedError 获取增强错误（全局函数）
func GetEnhancedError(err error) *EnhancedError {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr
	}
	return globalErrorHandler.NewError(err.Error(), "UNKNOWN", ErrorLevelError)
}

// WrapEnhancedError 包装增强错误（全局函数）
func WrapEnhancedError(err error, message string, code string) *EnhancedError {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		enhancedErr.Message = message
		enhancedErr.Code = code
		return enhancedErr
	}
	return globalErrorHandler.NewError(message, code, ErrorLevelError).WithContext("original_error", err.Error())
}

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	handler *EnhancedErrorHandler
	message string
	code    string
	level   ErrorLevel
	context map[string]interface{}
}

// NewErrorBuilder 创建错误构建器
func (h *EnhancedErrorHandler) NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		handler: h,
		level:   ErrorLevelError,
		context: make(map[string]interface{}),
	}
}

// Message 设置错误消息
func (b *ErrorBuilder) Message(message string) *ErrorBuilder {
	b.message = message
	return b
}

// Code 设置错误代码
func (b *ErrorBuilder) Code(code string) *ErrorBuilder {
	b.code = code
	return b
}

// Level 设置错误级别
func (b *ErrorBuilder) Level(level ErrorLevel) *ErrorBuilder {
	b.level = level
	return b
}

// Context 添加上下文
func (b *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	if b.context == nil {
		b.context = make(map[string]interface{})
	}
	b.context[key] = value
	return b
}

// Build 构建错误
func (b *ErrorBuilder) Build() *EnhancedError {
	err := b.handler.NewError(b.message, b.code, b.level)
	for key, value := range b.context {
		err.WithContext(key, value)
	}

	// 从上下文中提取特殊字段
	if status, ok := b.context["status_code"].(int); ok {
		err.Status = status
	}
	if severity, ok := b.context["severity"].(string); ok {
		err.Severity = severity
	}
	if details, ok := b.context["details"].(string); ok {
		err.Details = details
	}
	if category, ok := b.context["category"].(string); ok {
		err.Category = category
	}
	if retryable, ok := b.context["retryable"].(bool); ok {
		err.Retryable = retryable
	}

	return err
}

// Severity 设置严重程度
func (b *ErrorBuilder) Severity(severity string) *ErrorBuilder {
	b.Context("severity", severity)
	return b
}

// Details 设置详细信息
func (b *ErrorBuilder) Details(details string) *ErrorBuilder {
	b.Context("details", details)
	return b
}

// WithContextValue 添加上下文值
func (b *ErrorBuilder) WithContextValue(key string, ctx interface{}) *ErrorBuilder {
	b.Context(key, ctx)
	return b
}

// StackTrace 添加堆栈跟踪
func (b *ErrorBuilder) StackTrace() *ErrorBuilder {
	b.Context("stack_trace", true)
	return b
}

// Source 设置错误源
func (b *ErrorBuilder) Source(source string) *ErrorBuilder {
	b.Context("source", source)
	return b
}

// NewError 创建新的增强错误
func (h *EnhancedErrorHandler) NewError(message string, code string, level ErrorLevel) *EnhancedError {
	err := &EnhancedError{
		Message:   message,
		Code:      code,
		Level:     level,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Retryable: false,
		Category:  "general",
		Status:    500,      // 默认状态码
		Severity:  "medium", // 默认严重程度
	}

	if h.enableStack {
		err.Stack = h.getStackTrace()
	}

	return err
}

// NewValidationError 创建验证错误
func (h *EnhancedErrorHandler) NewValidationError(message string, field string) *EnhancedError {
	return h.NewError(message, "VALIDATION_ERROR", ErrorLevelWarning).
		WithContext("field", field).
		SetCategory("validation").
		WithUserMessage("输入数据验证失败")
}

// NewAuthError 创建认证错误
func (h *EnhancedErrorHandler) NewAuthError(message string) *EnhancedError {
	return h.NewError(message, "AUTH_ERROR", ErrorLevelError).
		SetCategory("authentication").
		WithUserMessage("认证失败，请检查登录信息")
}

// NewPermissionError 创建权限错误
func (h *EnhancedErrorHandler) NewPermissionError(message string) *EnhancedError {
	return h.NewError(message, "PERMISSION_ERROR", ErrorLevelError).
		SetCategory("authorization").
		WithUserMessage("权限不足，无法执行此操作")
}

// NewNotFoundError 创建未找到错误
func (h *EnhancedErrorHandler) NewNotFoundError(resource string) *EnhancedError {
	return h.NewError(fmt.Sprintf("%s not found", resource), "NOT_FOUND", ErrorLevelWarning).
		SetCategory("resource").
		WithUserMessage(fmt.Sprintf("未找到指定的%s", resource))
}

// NewDatabaseError 创建数据库错误
func (h *EnhancedErrorHandler) NewDatabaseError(message string, operation string) *EnhancedError {
	return h.NewError(message, "DATABASE_ERROR", ErrorLevelError).
		WithContext("operation", operation).
		SetCategory("database").
		SetRetryable(true).
		WithUserMessage("数据库操作失败，请稍后重试")
}

// NewNetworkError 创建网络错误
func (h *EnhancedErrorHandler) NewNetworkError(message string, url string) *EnhancedError {
	return h.NewError(message, "NETWORK_ERROR", ErrorLevelError).
		WithContext("url", url).
		SetCategory("network").
		SetRetryable(true).
		WithUserMessage("网络连接失败，请检查网络设置")
}

// NewBusinessError 创建业务错误
func (h *EnhancedErrorHandler) NewBusinessError(message string, code string) *EnhancedError {
	return h.NewError(message, code, ErrorLevelWarning).
		SetCategory("business").
		WithUserMessage(message)
}

// NewSystemError 创建系统错误
func (h *EnhancedErrorHandler) NewSystemError(message string) *EnhancedError {
	return h.NewError(message, "SYSTEM_ERROR", ErrorLevelFatal).
		SetCategory("system").
		WithUserMessage("系统内部错误，请联系管理员")
}

// WrapError 包装现有错误
func (h *EnhancedErrorHandler) WrapError(err error, message string, code string) *EnhancedError {
	if err == nil {
		return nil
	}

	enhancedErr := h.NewError(message, code, ErrorLevelError)
	enhancedErr.Context["original_error"] = err.Error()

	if h.enableStack {
		enhancedErr.Stack = h.getStackTrace()
	}

	return enhancedErr
}

// getStackTrace 获取堆栈跟踪
func (h *EnhancedErrorHandler) getStackTrace() []string {
	var stack []string
	for i := 2; i < 10; i++ { // 跳过当前函数和调用函数
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	return stack
}

// IsRetryable 检查错误是否可重试
func (h *EnhancedErrorHandler) IsRetryable(err error) bool {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr.Retryable
	}
	return false
}

// GetErrorLevel 获取错误级别
func (h *EnhancedErrorHandler) GetErrorLevel(err error) ErrorLevel {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr.Level
	}
	return ErrorLevelError
}

// GetErrorCode 获取错误代码
func (h *EnhancedErrorHandler) GetErrorCode(err error) string {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetUserMessage 获取用户友好的错误消息
func (h *EnhancedErrorHandler) GetUserMessage(err error) string {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		if enhancedErr.UserMessage != "" {
			return enhancedErr.UserMessage
		}
	}
	return err.Error()
}

// LogError 记录错误日志
func (h *EnhancedErrorHandler) LogError(err error, logger interface{}) {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		// 这里可以集成具体的日志记录器
		fmt.Printf("[%s] %s: %s\n",
			ErrorLevelString[enhancedErr.Level],
			enhancedErr.Code,
			enhancedErr.Message)

		if len(enhancedErr.Stack) > 0 {
			fmt.Printf("Stack trace:\n%s\n", strings.Join(enhancedErr.Stack, "\n"))
		}
	} else {
		fmt.Printf("[ERROR] %s\n", err.Error())
	}
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success     bool                   `json:"success"`
	Error       string                 `json:"error"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	UserMessage string                 `json:"user_message,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// ToResponse 转换为错误响应
func (e *EnhancedError) ToResponse() *ErrorResponse {
	return &ErrorResponse{
		Success:     false,
		Error:       e.Message,
		Code:        e.Code,
		Message:     e.Message,
		UserMessage: e.UserMessage,
		Context:     e.Context,
		Timestamp:   e.Timestamp,
	}
}

// 全局错误处理器实例
var GlobalErrorHandler = NewEnhancedErrorHandler()

// 便捷函数
func NewError(message string, code string, level ErrorLevel) *EnhancedError {
	return GlobalErrorHandler.NewError(message, code, level)
}

func NewValidationError(message string, field string) *EnhancedError {
	return GlobalErrorHandler.NewValidationError(message, field)
}

func NewAuthError(message string) *EnhancedError {
	return GlobalErrorHandler.NewAuthError(message)
}

func NewPermissionError(message string) *EnhancedError {
	return GlobalErrorHandler.NewPermissionError(message)
}

func NewNotFoundError(resource string) *EnhancedError {
	return GlobalErrorHandler.NewNotFoundError(resource)
}

func NewDatabaseError(message string, operation string) *EnhancedError {
	return GlobalErrorHandler.NewDatabaseError(message, operation)
}

func NewNetworkError(message string, url string) *EnhancedError {
	return GlobalErrorHandler.NewNetworkError(message, url)
}

func NewBusinessError(message string, code string) *EnhancedError {
	return GlobalErrorHandler.NewBusinessError(message, code)
}

func NewSystemError(message string) *EnhancedError {
	return GlobalErrorHandler.NewSystemError(message)
}

func WrapError(err error, message string, code string) *EnhancedError {
	return GlobalErrorHandler.WrapError(err, message, code)
}
