package Utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// ErrorCategory 错误分类
type ErrorCategory string

const (
	CategoryAuth       ErrorCategory = "authentication"
	CategoryValidation ErrorCategory = "validation"
	CategoryDatabase   ErrorCategory = "database"
	CategoryNetwork    ErrorCategory = "network"
	CategoryBusiness   ErrorCategory = "business"
	CategorySystem     ErrorCategory = "system"
	CategoryExternal   ErrorCategory = "external"
	CategorySecurity   ErrorCategory = "security"
)

// APIError 简单的API错误结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"status"`
}

// Error 实现error接口
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// EnhancedError 增强的错误结构
type EnhancedError struct {
	// 基础错误信息
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"status"`

	// 错误分类和严重程度
	Category ErrorCategory `json:"category"`
	Severity ErrorSeverity `json:"severity"`

	// 上下文信息
	Context   map[string]interface{} `json:"context,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`

	// 技术信息
	StackTrace string    `json:"stack_trace,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source,omitempty"`

	// 错误链
	Cause error `json:"-"`

	// 可恢复性
	Recoverable bool `json:"recoverable"`
	Retryable   bool `json:"retryable"`
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap 实现错误链
func (e *EnhancedError) Unwrap() error {
	return e.Cause
}

// NewEnhancedError 创建增强错误
func NewEnhancedError(code, message string, status int) *EnhancedError {
	return &EnhancedError{
		Code:        code,
		Message:     message,
		Status:      status,
		Category:    CategorySystem,
		Severity:    SeverityMedium,
		Context:     make(map[string]interface{}),
		Timestamp:   time.Now(),
		Recoverable: true,
		Retryable:   false,
	}
}

// WithCategory 设置错误分类
func (e *EnhancedError) WithCategory(category ErrorCategory) *EnhancedError {
	e.Category = category
	return e
}

// WithSeverity 设置错误严重程度
func (e *EnhancedError) WithSeverity(severity ErrorSeverity) *EnhancedError {
	e.Severity = severity
	return e
}

// WithDetails 设置详细信息
func (e *EnhancedError) WithDetails(details string) *EnhancedError {
	e.Details = details
	return e
}

// WithContext 添加上下文信息
func (e *EnhancedError) WithContext(key string, value interface{}) *EnhancedError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithUserID 设置用户ID
func (e *EnhancedError) WithUserID(userID string) *EnhancedError {
	e.UserID = userID
	return e
}

// WithRequestID 设置请求ID
func (e *EnhancedError) WithRequestID(requestID string) *EnhancedError {
	e.RequestID = requestID
	return e
}

// WithContextValue 从context.Context中提取值
func (e *EnhancedError) WithContextValue(ctx context.Context) *EnhancedError {
	// 从上下文中提取信息
	if userID, ok := ctx.Value("user_id").(string); ok {
		e.UserID = userID
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		e.RequestID = requestID
	}

	// 添加更多上下文信息
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}

	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		e.Context["ip_address"] = ipAddress
	}

	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		e.Context["user_agent"] = userAgent
	}

	return e
}

// WithCause 设置原因错误
func (e *EnhancedError) WithCause(cause error) *EnhancedError {
	e.Cause = cause
	return e
}

// WithStackTrace 设置堆栈跟踪
func (e *EnhancedError) WithStackTrace() *EnhancedError {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	e.StackTrace = string(buf[:n])
	return e
}

// WithSource 设置错误来源
func (e *EnhancedError) WithSource(source string) *EnhancedError {
	e.Source = source
	return e
}

// SetRecoverable 设置是否可恢复
func (e *EnhancedError) SetRecoverable(recoverable bool) *EnhancedError {
	e.Recoverable = recoverable
	return e
}

// SetRetryable 设置是否可重试
func (e *EnhancedError) SetRetryable(retryable bool) *EnhancedError {
	e.Retryable = retryable
	return e
}

// ToJSON 转换为JSON字符串
func (e *EnhancedError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// LogLevel 获取日志级别
func (e *EnhancedError) LogLevel() string {
	switch e.Severity {
	case SeverityCritical:
		return "error"
	case SeverityHigh:
		return "error"
	case SeverityMedium:
		return "warn"
	case SeverityLow:
		return "info"
	default:
		return "info"
	}
}

// ShouldAlert 判断是否需要告警
func (e *EnhancedError) ShouldAlert() bool {
	return e.Severity == SeverityCritical || e.Severity == SeverityHigh
}

// 预定义的增强错误
var (
	// 认证相关错误
	ErrInvalidTokenEnhanced = NewEnhancedError("INVALID_TOKEN", "Invalid or expired token", http.StatusUnauthorized).
				WithCategory(CategoryAuth).WithSeverity(SeverityMedium)

	ErrTokenExpiredEnhanced = NewEnhancedError("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized).
				WithCategory(CategoryAuth).WithSeverity(SeverityMedium)

	ErrUnauthorizedEnhanced = NewEnhancedError("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized).
				WithCategory(CategoryAuth).WithSeverity(SeverityMedium)

	// 验证相关错误
	ErrValidationEnhanced = NewEnhancedError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest).
				WithCategory(CategoryValidation).WithSeverity(SeverityLow)

	ErrInvalidInputEnhanced = NewEnhancedError("INVALID_INPUT", "Invalid input data", http.StatusBadRequest).
				WithCategory(CategoryValidation).WithSeverity(SeverityLow)

	// 数据库相关错误
	ErrDatabaseEnhanced = NewEnhancedError("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError).
				WithCategory(CategoryDatabase).WithSeverity(SeverityHigh).SetRetryable(true)

	ErrConnectionFailedEnhanced = NewEnhancedError("CONNECTION_FAILED", "Database connection failed", http.StatusInternalServerError).
					WithCategory(CategoryDatabase).WithSeverity(SeverityHigh).SetRetryable(true)

	// 业务相关错误
	ErrBusinessLogicEnhanced = NewEnhancedError("BUSINESS_ERROR", "Business logic error", http.StatusBadRequest).
					WithCategory(CategoryBusiness).WithSeverity(SeverityMedium)

	ErrResourceNotFoundEnhanced = NewEnhancedError("RESOURCE_NOT_FOUND", "Resource not found", http.StatusNotFound).
					WithCategory(CategoryBusiness).WithSeverity(SeverityLow)

	// 系统相关错误
	ErrInternalServerEnhanced = NewEnhancedError("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError).
					WithCategory(CategorySystem).WithSeverity(SeverityCritical)

	ErrServiceUnavailableEnhanced = NewEnhancedError("SERVICE_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable).
					WithCategory(CategorySystem).WithSeverity(SeverityHigh).SetRetryable(true)

	// 安全相关错误
	ErrSecurityViolationEnhanced = NewEnhancedError("SECURITY_VIOLATION", "Security violation detected", http.StatusForbidden).
					WithCategory(CategorySecurity).WithSeverity(SeverityHigh)

	ErrRateLimitEnhanced = NewEnhancedError("RATE_LIMIT", "Rate limit exceeded", http.StatusTooManyRequests).
				WithCategory(CategorySecurity).WithSeverity(SeverityMedium)
)

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	err *EnhancedError
}

// NewErrorBuilder 创建错误构建器
func NewErrorBuilder(code, message string, status int) *ErrorBuilder {
	return &ErrorBuilder{
		err: NewEnhancedError(code, message, status),
	}
}

// Category 设置分类
func (b *ErrorBuilder) Category(category ErrorCategory) *ErrorBuilder {
	b.err.Category = category
	return b
}

// Severity 设置严重程度
func (b *ErrorBuilder) Severity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// Details 设置详细信息
func (b *ErrorBuilder) Details(details string) *ErrorBuilder {
	b.err.Details = details
	return b
}

// Context 添加上下文
func (b *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	if b.err.Context == nil {
		b.err.Context = make(map[string]interface{})
	}
	b.err.Context[key] = value
	return b
}

// UserID 设置用户ID
func (b *ErrorBuilder) UserID(userID string) *ErrorBuilder {
	b.err.UserID = userID
	return b
}

// RequestID 设置请求ID
func (b *ErrorBuilder) RequestID(requestID string) *ErrorBuilder {
	b.err.RequestID = requestID
	return b
}

// Cause 设置原因
func (b *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// StackTrace 添加堆栈跟踪
func (b *ErrorBuilder) StackTrace() *ErrorBuilder {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	b.err.StackTrace = string(buf[:n])
	return b
}

// Source 设置来源
func (b *ErrorBuilder) Source(source string) *ErrorBuilder {
	b.err.Source = source
	return b
}

// Recoverable 设置可恢复性
func (b *ErrorBuilder) Recoverable(recoverable bool) *ErrorBuilder {
	b.err.Recoverable = recoverable
	return b
}

// Retryable 设置可重试性
func (b *ErrorBuilder) Retryable(retryable bool) *ErrorBuilder {
	b.err.Retryable = retryable
	return b
}

// SetRetryable 设置可重试性（别名方法）
func (b *ErrorBuilder) SetRetryable(retryable bool) *ErrorBuilder {
	b.err.Retryable = retryable
	return b
}

// SetRecoverable 设置可恢复性（别名方法）
func (b *ErrorBuilder) SetRecoverable(recoverable bool) *ErrorBuilder {
	b.err.Recoverable = recoverable
	return b
}

// WithContextValue 从context.Context中提取值
func (b *ErrorBuilder) WithContextValue(ctx context.Context) *ErrorBuilder {
	// 从上下文中提取信息
	if userID, ok := ctx.Value("user_id").(string); ok {
		b.err.UserID = userID
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		b.err.RequestID = requestID
	}

	// 添加更多上下文信息
	if b.err.Context == nil {
		b.err.Context = make(map[string]interface{})
	}

	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		b.err.Context["ip_address"] = ipAddress
	}

	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		b.err.Context["user_agent"] = userAgent
	}

	return b
}

// Build 构建错误
func (b *ErrorBuilder) Build() *EnhancedError {
	return b.err
}

// WrapEnhancedError 包装现有错误
func WrapEnhancedError(err error, code, message string) *EnhancedError {
	if err == nil {
		return nil
	}

	enhancedErr := NewEnhancedError(code, message, http.StatusInternalServerError).
		WithCause(err).
		WithStackTrace()

	// 尝试从现有错误中提取信息
	if apiErr, ok := err.(*APIError); ok {
		enhancedErr.Status = apiErr.Status
		enhancedErr.Code = apiErr.Code
		enhancedErr.Message = apiErr.Message
		enhancedErr.Details = apiErr.Details
	}

	return enhancedErr
}

// WrapWithContext 使用上下文包装错误
func WrapWithContext(ctx context.Context, err error, code, message string) *EnhancedError {
	enhancedErr := WrapEnhancedError(err, code, message)

	// 从上下文中提取信息
	if userID, ok := ctx.Value("user_id").(string); ok {
		enhancedErr.UserID = userID
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		enhancedErr.RequestID = requestID
	}

	// 添加更多上下文信息
	enhancedErr.Context["timestamp"] = time.Now()
	enhancedErr.Context["goroutine_id"] = getGoroutineID()

	return enhancedErr
}

// getGoroutineID 获取goroutine ID
func getGoroutineID() string {
	buf := make([]byte, 64)
	buf = buf[:runtime.Stack(buf, false)]
	// 提取goroutine ID
	lines := strings.Split(string(buf), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "unknown"
}

// IsEnhancedError 检查是否为增强错误
func IsEnhancedError(err error) bool {
	_, ok := err.(*EnhancedError)
	return ok
}

// GetEnhancedError 获取增强错误
func GetEnhancedError(err error) (*EnhancedError, bool) {
	enhancedErr, ok := err.(*EnhancedError)
	return enhancedErr, ok
}

// ErrorCollector 错误收集器
type ErrorCollector struct {
	errors []*EnhancedError
}

// NewErrorCollector 创建错误收集器
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]*EnhancedError, 0),
	}
}

// Add 添加错误
func (c *ErrorCollector) Add(err *EnhancedError) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

// AddError 添加普通错误
func (c *ErrorCollector) AddError(err error, code, message string) {
	if err != nil {
		enhancedErr := WrapEnhancedError(err, code, message)
		c.errors = append(c.errors, enhancedErr)
	}
}

// HasErrors 检查是否有错误
func (c *ErrorCollector) HasErrors() bool {
	return len(c.errors) > 0
}

// GetErrors 获取所有错误
func (c *ErrorCollector) GetErrors() []*EnhancedError {
	return c.errors
}

// GetCriticalErrors 获取严重错误
func (c *ErrorCollector) GetCriticalErrors() []*EnhancedError {
	var critical []*EnhancedError
	for _, err := range c.errors {
		if err.Severity == SeverityCritical {
			critical = append(critical, err)
		}
	}
	return critical
}

// GetRetryableErrors 获取可重试错误
func (c *ErrorCollector) GetRetryableErrors() []*EnhancedError {
	var retryable []*EnhancedError
	for _, err := range c.errors {
		if err.Retryable {
			retryable = append(retryable, err)
		}
	}
	return retryable
}

// Clear 清空错误
func (c *ErrorCollector) Clear() {
	c.errors = c.errors[:0]
}

// ToJSON 转换为JSON
func (c *ErrorCollector) ToJSON() string {
	data, _ := json.Marshal(c.errors)
	return string(data)
}
