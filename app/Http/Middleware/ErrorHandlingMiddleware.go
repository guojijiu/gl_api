package Middleware

import (
	"cloud-platform-api/app/Storage"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorHandlingMiddleware 增强的错误处理中间件
type ErrorHandlingMiddleware struct {
	storageManager *Storage.StorageManager
	config         *ErrorHandlingConfig
}

// ErrorHandlingConfig 错误处理配置
type ErrorHandlingConfig struct {
	EnableDetailedErrors bool `json:"enable_detailed_errors"` // 是否启用详细错误信息
	LogAllErrors         bool `json:"log_all_errors"`         // 是否记录所有错误
	EnableErrorTracking  bool `json:"enable_error_tracking"`  // 是否启用错误追踪
	MaxErrorLogSize      int  `json:"max_error_log_size"`     // 最大错误日志大小
}

// CustomError 自定义错误类型
type CustomError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

// Error 实现error接口
func (e *CustomError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewErrorHandlingMiddleware 创建错误处理中间件
// 功能说明：
// 1. 初始化错误处理中间件
// 2. 配置错误处理策略
// 3. 提供统一的错误响应格式
// 4. 支持错误分类和追踪
func NewErrorHandlingMiddleware(storageManager *Storage.StorageManager, config *ErrorHandlingConfig) *ErrorHandlingMiddleware {
	if config == nil {
		config = &ErrorHandlingConfig{
			EnableDetailedErrors: false, // 生产环境默认关闭
			LogAllErrors:         true,
			EnableErrorTracking:  true,
			MaxErrorLogSize:      1000,
		}
	}

	return &ErrorHandlingMiddleware{
		storageManager: storageManager,
		config:         config,
	}
}

// Handle 处理错误
// 功能说明：
// 1. 捕获panic和未处理的错误
// 2. 分类错误类型
// 3. 记录错误日志
// 4. 返回统一的错误响应
// 5. 支持错误追踪
func (m *ErrorHandlingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				m.handlePanic(c, err)
			}
		}()

		c.Next()

		// 处理其他错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			m.handleError(c, err.Err)
		}
	}
}

// handlePanic 处理panic
func (m *ErrorHandlingMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	// 记录panic日志
	m.storageManager.LogError("应用发生panic", map[string]interface{}{
		"error":       recovered,
		"url":         c.Request.URL.String(),
		"method":      c.Request.Method,
		"client_ip":   c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"stack_trace": string(debug.Stack()),
		"user_id":     c.GetString("user_id"),
	})

	// 返回错误响应
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"message": "Internal server error",
		"error":   "An unexpected error occurred",
		"code":    "INTERNAL_ERROR",
	})
	c.Abort()
}

// handleError 处理普通错误
func (m *ErrorHandlingMiddleware) handleError(c *gin.Context, err error) {
	// 分类错误
	errorType, statusCode, message := m.classifyError(err)

	// 记录错误日志
	if m.config.LogAllErrors {
		m.logError(c, err, errorType, statusCode)
	}

	// 如果还没有响应，返回错误
	if !c.Writer.Written() {
		response := gin.H{
			"success": false,
			"message": message,
			"code":    errorType,
		}

		// 添加详细错误信息（仅在开发环境）
		if m.config.EnableDetailedErrors {
			response["error"] = err.Error()
			if m.config.EnableErrorTracking {
				response["stack"] = string(debug.Stack())
			}
		}

		c.JSON(statusCode, response)
	}
}

// classifyError 分类错误
// 功能说明：
// 1. 根据错误类型返回相应的HTTP状态码
// 2. 提供友好的错误消息
// 3. 返回错误代码用于前端处理
// 4. 支持自定义错误类型扩展
func (m *ErrorHandlingMiddleware) classifyError(err error) (string, int, string) {
	if err == nil {
		return "SUCCESS", http.StatusOK, "Success"
	}

	// 检查常见的错误类型
	switch {
	case isValidationError(err):
		return "VALIDATION_ERROR", http.StatusBadRequest, "请求参数验证失败"
	case isNotFoundError(err):
		return "NOT_FOUND", http.StatusNotFound, "请求的资源不存在"
	case isUnauthorizedError(err):
		return "UNAUTHORIZED", http.StatusUnauthorized, "未授权访问"
	case isForbiddenError(err):
		return "FORBIDDEN", http.StatusForbidden, "禁止访问"
	case isConflictError(err):
		return "CONFLICT", http.StatusConflict, "资源冲突"
	case isRateLimitError(err):
		return "RATE_LIMIT", http.StatusTooManyRequests, "请求频率超限"
	case isDatabaseError(err):
		return "DATABASE_ERROR", http.StatusInternalServerError, "数据库操作失败"
	case isNetworkError(err):
		return "NETWORK_ERROR", http.StatusBadGateway, "网络连接失败"
	case isTimeoutError(err):
		return "TIMEOUT", http.StatusRequestTimeout, "请求超时"
	case isFileError(err):
		return "FILE_ERROR", http.StatusBadRequest, "文件操作失败"
	default:
		return "INTERNAL_ERROR", http.StatusInternalServerError, "服务器内部错误"
	}
}

// 错误类型检查函数
func isValidationError(err error) bool {
	return err != nil && (contains(err.Error(), "validation") ||
		contains(err.Error(), "invalid") ||
		contains(err.Error(), "required") ||
		contains(err.Error(), "binding"))
}

func isNotFoundError(err error) bool {
	return err != nil && (contains(err.Error(), "not found") ||
		contains(err.Error(), "record not found") ||
		contains(err.Error(), "不存在"))
}

func isUnauthorizedError(err error) bool {
	return err != nil && (contains(err.Error(), "unauthorized") ||
		contains(err.Error(), "invalid token") ||
		contains(err.Error(), "未授权"))
}

func isForbiddenError(err error) bool {
	return err != nil && (contains(err.Error(), "forbidden") ||
		contains(err.Error(), "access denied") ||
		contains(err.Error(), "禁止"))
}

func isConflictError(err error) bool {
	return err != nil && (contains(err.Error(), "conflict") ||
		contains(err.Error(), "already exists") ||
		contains(err.Error(), "duplicate") ||
		contains(err.Error(), "冲突"))
}

func isRateLimitError(err error) bool {
	return err != nil && (contains(err.Error(), "rate limit") ||
		contains(err.Error(), "too many requests") ||
		contains(err.Error(), "频率限制"))
}

func isDatabaseError(err error) bool {
	return err != nil && (contains(err.Error(), "database") ||
		contains(err.Error(), "sql") ||
		contains(err.Error(), "connection") ||
		contains(err.Error(), "数据库"))
}

func isNetworkError(err error) bool {
	return err != nil && (contains(err.Error(), "network") ||
		contains(err.Error(), "connection refused") ||
		contains(err.Error(), "timeout") ||
		contains(err.Error(), "网络"))
}

func isTimeoutError(err error) bool {
	return err != nil && (contains(err.Error(), "timeout") ||
		contains(err.Error(), "deadline exceeded") ||
		contains(err.Error(), "超时"))
}

func isFileError(err error) bool {
	return err != nil && (contains(err.Error(), "file") ||
		contains(err.Error(), "directory") ||
		contains(err.Error(), "permission denied") ||
		contains(err.Error(), "文件"))
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// logError 记录错误日志
func (m *ErrorHandlingMiddleware) logError(c *gin.Context, err error, errorType string, statusCode int) {
	logData := map[string]interface{}{
		"error_type":  errorType,
		"status_code": statusCode,
		"error":       err.Error(),
		"url":         c.Request.URL.String(),
		"method":      c.Request.Method,
		"client_ip":   c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"user_id":     c.GetString("user_id"),
	}

	// 根据错误类型选择日志级别
	switch statusCode {
	case http.StatusBadRequest, http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden:
		m.storageManager.LogWarning("请求错误", logData)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		m.storageManager.LogError("服务器错误", logData)
	default:
		m.storageManager.LogError("未知错误", logData)
	}
}

// CreateCustomError 创建自定义错误
// 功能说明：
// 1. 创建标准化的错误对象
// 2. 支持错误代码和消息
// 3. 可选的详细信息和堆栈跟踪
func (m *ErrorHandlingMiddleware) CreateCustomError(code, message, details string) *CustomError {
	err := &CustomError{
		Code:    code,
		Message: message,
		Details: details,
	}

	if m.config.EnableErrorTracking {
		err.Stack = string(debug.Stack())
	}

	return err
}

// WrapError 包装错误
// 功能说明：
// 1. 为现有错误添加上下文信息
// 2. 保持原始错误信息
// 3. 添加额外的错误代码
func (m *ErrorHandlingMiddleware) WrapError(err error, code, message string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %s - %w", code, message, err)
}

// IsCustomError 检查是否为自定义错误
func (m *ErrorHandlingMiddleware) IsCustomError(err error) bool {
	var customErr *CustomError
	return errors.As(err, &customErr)
}

// GetCustomError 获取自定义错误
func (m *ErrorHandlingMiddleware) GetCustomError(err error) (*CustomError, bool) {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr, true
	}
	return nil, false
}
