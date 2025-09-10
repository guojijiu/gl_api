package Middleware

import (
	"cloud-platform-api/app/Utils"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedErrorHandlingMiddleware 增强的错误处理中间件
type EnhancedErrorHandlingMiddleware struct {
	config *EnhancedErrorHandlingConfig
}

// EnhancedErrorHandlingConfig 增强错误处理配置
type EnhancedErrorHandlingConfig struct {
	EnableDetailedErrors  bool          `json:"enable_detailed_errors"`
	EnableErrorTracking   bool          `json:"enable_error_tracking"`
	EnablePerformanceLog  bool          `json:"enable_performance_log"`
	MaxErrorLogSize       int           `json:"max_error_log_size"`
	SlowRequestThreshold  time.Duration `json:"slow_request_threshold"`
	EnableRequestID       bool          `json:"enable_request_id"`
	EnableUserTracking    bool          `json:"enable_user_tracking"`
	EnableSecurityLogging bool          `json:"enable_security_logging"`
	ErrorResponseFormat   string        `json:"error_response_format"` // json, text
	IncludeStackTrace     bool          `json:"include_stack_trace"`
	EnableErrorMetrics    bool          `json:"enable_error_metrics"`
}

// NewEnhancedErrorHandlingMiddleware 创建增强错误处理中间件
func NewEnhancedErrorHandlingMiddleware(config *EnhancedErrorHandlingConfig) *EnhancedErrorHandlingMiddleware {
	if config == nil {
		config = &EnhancedErrorHandlingConfig{
			EnableDetailedErrors:  false,
			EnableErrorTracking:   true,
			EnablePerformanceLog:  true,
			MaxErrorLogSize:       1000,
			SlowRequestThreshold:  5 * time.Second,
			EnableRequestID:       true,
			EnableUserTracking:    true,
			EnableSecurityLogging: true,
			ErrorResponseFormat:   "json",
			IncludeStackTrace:     false,
			EnableErrorMetrics:    true,
		}
	}

	return &EnhancedErrorHandlingMiddleware{
		config: config,
	}
}

// Handle 处理请求
func (m *EnhancedErrorHandlingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		if m.config.EnableRequestID {
			requestID := m.generateRequestID()
			c.Set("request_id", requestID)
			c.Header("X-Request-ID", requestID)
		}

		// 记录请求开始时间
		startTime := time.Now()
		c.Set("start_time", startTime)

		// 设置上下文
		ctx := m.createContext(c)
		c.Request = c.Request.WithContext(ctx)

		// 记录请求信息
		m.logRequest(c)

		// 处理panic
		defer func() {
			if err := recover(); err != nil {
				m.handlePanic(c, err)
			}
		}()

		// 处理请求
		c.Next()

		// 记录响应信息
		m.logResponse(c, startTime)

		// 处理错误
		if len(c.Errors) > 0 {
			m.handleErrors(c)
		}
	}
}

// createContext 创建请求上下文
func (m *EnhancedErrorHandlingMiddleware) createContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// 添加请求ID
	if requestID, exists := c.Get("request_id"); exists {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}

	// 添加用户ID
	if userID, exists := c.Get("user_id"); exists {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	// 添加IP地址
	ctx = context.WithValue(ctx, "ip_address", c.ClientIP())

	// 添加User-Agent
	ctx = context.WithValue(ctx, "user_agent", c.Request.UserAgent())

	return ctx
}

// logRequest 记录请求信息
func (m *EnhancedErrorHandlingMiddleware) logRequest(c *gin.Context) {
	fields := map[string]interface{}{
		"method":         c.Request.Method,
		"url":            c.Request.URL.String(),
		"user_agent":     c.Request.UserAgent(),
		"ip_address":     c.ClientIP(),
		"referer":        c.Request.Referer(),
		"content_length": c.Request.ContentLength,
	}

	// 添加用户信息
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	// 添加请求ID
	if requestID, exists := c.Get("request_id"); exists {
		fields["request_id"] = requestID
	}

	Utils.WithFields(fields).Info("Request started")
}

// logResponse 记录响应信息
func (m *EnhancedErrorHandlingMiddleware) logResponse(c *gin.Context, startTime time.Time) {
	duration := time.Since(startTime)

	fields := map[string]interface{}{
		"status_code":   c.Writer.Status(),
		"duration":      duration.String(),
		"duration_ms":   duration.Milliseconds(),
		"response_size": c.Writer.Size(),
	}

	// 添加用户信息
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	// 添加请求ID
	if requestID, exists := c.Get("request_id"); exists {
		fields["request_id"] = requestID
	}

	// 检查是否为慢请求
	if duration > m.config.SlowRequestThreshold {
		fields["slow_request"] = true
		Utils.WithFields(fields).Warn("Slow request detected")
	} else {
		Utils.WithFields(fields).Info("Request completed")
	}
}

// handlePanic 处理panic
func (m *EnhancedErrorHandlingMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	// 创建增强错误
	err := Utils.NewErrorBuilder("PANIC", "Application panic occurred", http.StatusInternalServerError).
		Category(Utils.CategorySystem).
		Severity(Utils.SeverityCritical).
		Details(fmt.Sprintf("%v", recovered)).
		WithContextValue(c.Request.Context()).
		StackTrace().
		Source(m.getCallerInfo()).
		SetRecoverable(false).
		Build()

	// 记录错误
	Utils.LogError(err)

	// 返回错误响应
	m.sendErrorResponse(c, err)
	c.Abort()
}

// handleErrors 处理错误
func (m *EnhancedErrorHandlingMiddleware) handleErrors(c *gin.Context) {
	for _, ginErr := range c.Errors {
		enhancedErr := m.convertToEnhancedError(ginErr.Err, c)
		Utils.LogError(enhancedErr)

		// 只处理第一个错误
		if !c.Writer.Written() {
			m.sendErrorResponse(c, enhancedErr)
		}
		break
	}
}

// convertToEnhancedError 转换为增强错误
func (m *EnhancedErrorHandlingMiddleware) convertToEnhancedError(err error, c *gin.Context) *Utils.EnhancedError {
	// 如果已经是增强错误，直接返回
	if enhancedErr, ok := Utils.GetEnhancedError(err); ok {
		return enhancedErr
	}

	// 根据错误类型创建增强错误
	enhancedErr := m.classifyAndCreateError(err, c)

	// 添加上下文信息
	enhancedErr = enhancedErr.WithContextValue(c.Request.Context())

	// 添加请求信息
	enhancedErr = enhancedErr.WithContext("method", c.Request.Method).
		WithContext("url", c.Request.URL.String()).
		WithContext("ip_address", c.ClientIP()).
		WithContext("user_agent", c.Request.UserAgent())

	// 添加堆栈跟踪
	if m.config.IncludeStackTrace {
		enhancedErr = enhancedErr.WithStackTrace()
	}

	return enhancedErr
}

// classifyAndCreateError 分类并创建错误
func (m *EnhancedErrorHandlingMiddleware) classifyAndCreateError(err error, c *gin.Context) *Utils.EnhancedError {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return Utils.NewErrorBuilder("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest).
			Category(Utils.CategoryValidation).
			Severity(Utils.SeverityLow).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "不存在"):
		return Utils.NewErrorBuilder("NOT_FOUND", "Resource not found", http.StatusNotFound).
			Category(Utils.CategoryBusiness).
			Severity(Utils.SeverityLow).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "未授权"):
		return Utils.NewErrorBuilder("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized).
			Category(Utils.CategoryAuth).
			Severity(Utils.SeverityMedium).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "禁止"):
		return Utils.NewErrorBuilder("FORBIDDEN", "Forbidden access", http.StatusForbidden).
			Category(Utils.CategoryAuth).
			Severity(Utils.SeverityMedium).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		return Utils.NewErrorBuilder("CONFLICT", "Resource conflict", http.StatusConflict).
			Category(Utils.CategoryBusiness).
			Severity(Utils.SeverityLow).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many requests"):
		return Utils.NewErrorBuilder("RATE_LIMIT", "Rate limit exceeded", http.StatusTooManyRequests).
			Category(Utils.CategorySecurity).
			Severity(Utils.SeverityMedium).
			Details(err.Error()).
			Build()

	case strings.Contains(errStr, "database") || strings.Contains(errStr, "sql"):
		return Utils.NewErrorBuilder("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError).
			Category(Utils.CategoryDatabase).
			Severity(Utils.SeverityHigh).
			Details(err.Error()).
			SetRetryable(true).
			Build()

	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return Utils.NewErrorBuilder("NETWORK_ERROR", "Network error", http.StatusBadGateway).
			Category(Utils.CategoryNetwork).
			Severity(Utils.SeverityHigh).
			Details(err.Error()).
			SetRetryable(true).
			Build()

	case strings.Contains(errStr, "timeout"):
		return Utils.NewErrorBuilder("TIMEOUT", "Request timeout", http.StatusRequestTimeout).
			Category(Utils.CategoryNetwork).
			Severity(Utils.SeverityMedium).
			Details(err.Error()).
			SetRetryable(true).
			Build()

	default:
		return Utils.NewErrorBuilder("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError).
			Category(Utils.CategorySystem).
			Severity(Utils.SeverityHigh).
			Details(err.Error()).
			Build()
	}
}

// sendErrorResponse 发送错误响应
func (m *EnhancedErrorHandlingMiddleware) sendErrorResponse(c *gin.Context, err *Utils.EnhancedError) {
	response := gin.H{
		"success": false,
		"message": err.Message,
		"code":    err.Code,
		"status":  err.Status,
	}

	// 添加详细信息
	if m.config.EnableDetailedErrors {
		response["details"] = err.Details
		response["category"] = err.Category
		response["severity"] = err.Severity

		if err.RequestID != "" {
			response["request_id"] = err.RequestID
		}
	}

	// 添加堆栈跟踪
	if m.config.IncludeStackTrace && err.StackTrace != "" {
		response["stack_trace"] = err.StackTrace
	}

	// 添加重试信息
	if err.Retryable {
		response["retryable"] = true
	}

	// 设置响应头
	c.Header("X-Error-Code", err.Code)
	c.Header("X-Error-Category", string(err.Category))

	if err.RequestID != "" {
		c.Header("X-Request-ID", err.RequestID)
	}

	// 发送响应
	if m.config.ErrorResponseFormat == "text" {
		c.String(err.Status, "%s: %s", err.Code, err.Message)
	} else {
		c.JSON(err.Status, response)
	}
}

// generateRequestID 生成请求ID
func (m *EnhancedErrorHandlingMiddleware) generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// getCallerInfo 获取调用者信息
func (m *EnhancedErrorHandlingMiddleware) getCallerInfo() string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}

	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Sprintf("%s:%d %s", file, line, funcName)
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(c *gin.Context, err *Utils.EnhancedError)
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct{}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(c *gin.Context, err *Utils.EnhancedError) {
	Utils.LogError(err)

	response := gin.H{
		"success": false,
		"message": err.Message,
		"code":    err.Code,
		"status":  err.Status,
	}

	c.JSON(err.Status, response)
}

// CustomErrorHandler 自定义错误处理器
type CustomErrorHandler struct {
	handlers map[Utils.ErrorCategory]func(c *gin.Context, err *Utils.EnhancedError)
}

// NewCustomErrorHandler 创建自定义错误处理器
func NewCustomErrorHandler() *CustomErrorHandler {
	return &CustomErrorHandler{
		handlers: make(map[Utils.ErrorCategory]func(c *gin.Context, err *Utils.EnhancedError)),
	}
}

// RegisterHandler 注册错误处理器
func (h *CustomErrorHandler) RegisterHandler(category Utils.ErrorCategory, handler func(c *gin.Context, err *Utils.EnhancedError)) {
	h.handlers[category] = handler
}

// HandleError 处理错误
func (h *CustomErrorHandler) HandleError(c *gin.Context, err *Utils.EnhancedError) {
	if handler, exists := h.handlers[err.Category]; exists {
		handler(c, err)
	} else {
		// 使用默认处理器
		defaultHandler := &DefaultErrorHandler{}
		defaultHandler.HandleError(c, err)
	}
}
