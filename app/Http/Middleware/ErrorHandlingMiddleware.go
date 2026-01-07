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
//
// 实现原理：
// - 使用defer + recover捕获panic，防止程序崩溃
// - 在c.Next()后检查Gin的错误列表，处理业务错误
// - 根据错误类型返回相应的HTTP状态码和错误消息
//
// 错误处理流程：
// 1. 使用defer recover捕获panic（运行时错误）
// 2. 执行后续中间件和处理器（c.Next()）
// 3. 检查Gin错误列表，处理业务错误
// 4. 根据错误类型分类并返回相应响应
//
// 注意事项：
// - recover只能捕获当前goroutine的panic
// - 如果后续中间件在goroutine中panic，需要在该goroutine中recover
// - c.Errors是Gin收集的错误列表，由中间件和处理器添加
// - 如果响应已写入（c.Writer.Written()），不再重复写入
func (m *ErrorHandlingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用defer + recover捕获panic
		// recover只能捕获当前goroutine的panic
		// 如果后续中间件在goroutine中panic，需要在那个goroutine中recover
		defer func() {
			// 捕获panic（运行时错误，如空指针、数组越界等）
			if err := recover(); err != nil {
				// 处理panic，记录堆栈信息并返回500错误
				m.handlePanic(c, err)
			}
		}()

		// 执行后续中间件和处理器
		// 如果发生错误，会被添加到c.Errors中
		c.Next()

		// 处理业务错误（非panic错误）
		// Gin的错误列表由中间件和处理器通过c.Error()添加
		if len(c.Errors) > 0 {
			// 获取最后一个错误（通常是最相关的错误）
			err := c.Errors.Last()
			// 处理错误，返回相应的HTTP响应
			m.handleError(c, err.Err)
		}
	}
}

// handlePanic 处理panic
//
// 功能说明：
// 1. 捕获并处理运行时panic（如空指针、数组越界等）
// 2. 记录详细的panic信息，包括堆栈跟踪
// 3. 返回500内部服务器错误响应
// 4. 中止后续处理，防止重复响应
//
// 记录的信息：
// - panic的值（通常是错误信息或对象）
// - 请求URL、方法、客户端IP、User-Agent
// - 完整的堆栈跟踪（用于定位问题）
// - 用户ID（如果已认证）
//
// 安全考虑：
// - 堆栈跟踪包含敏感信息，只在开发环境显示
// - 生产环境只返回通用错误消息，避免信息泄露
// - 详细的错误信息只记录到日志，不返回给客户端
//
// 注意事项：
// - 必须调用c.Abort()停止后续处理
// - 如果响应已写入，不应再次写入
// - 堆栈跟踪可能很长，需要合理存储
func (m *ErrorHandlingMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	// 记录panic日志（包含详细信息用于问题诊断）
	// 注意：堆栈跟踪包含敏感信息，只在日志中记录，不返回给客户端
	m.storageManager.LogError("应用发生panic", map[string]interface{}{
		"error":       recovered,                    // panic的值
		"url":         c.Request.URL.String(),      // 请求URL
		"method":      c.Request.Method,             // HTTP方法
		"client_ip":   c.ClientIP(),                 // 客户端IP
		"user_agent":  c.Request.UserAgent(),        // User-Agent
		"stack_trace": string(debug.Stack()),        // 完整的堆栈跟踪
		"user_id":     c.GetString("user_id"),       // 用户ID（如果已认证）
	})

	// 返回500内部服务器错误响应
	// 注意：不返回详细的错误信息，避免信息泄露
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"message": "Internal server error",
		"error":   "An unexpected error occurred",
		"code":    "INTERNAL_ERROR",
	})
	
	// 中止后续处理，防止重复响应
	// 如果不调用Abort()，后续中间件可能继续执行并写入响应
	c.Abort()
}

// handleError 处理普通错误
//
// 功能说明：
// 1. 分类错误类型，确定HTTP状态码和错误消息
// 2. 根据配置决定是否记录错误日志
// 3. 返回统一的错误响应格式
// 4. 在开发环境可以返回详细错误信息
//
// 错误分类：
// - 根据错误类型（如验证错误、未找到错误等）返回相应状态码
// - 提供用户友好的错误消息
// - 返回错误代码供前端处理
//
// 响应格式：
// - success: false（表示请求失败）
// - message: 用户友好的错误消息
// - code: 错误代码（用于前端错误处理）
// - error: 详细错误信息（仅在开发环境）
// - stack: 堆栈跟踪（仅在开发环境且启用错误追踪）
//
// 注意事项：
// - 检查c.Writer.Written()避免重复写入响应
// - 详细错误信息只在开发环境返回，生产环境不返回
// - 堆栈跟踪可能很长，需要合理处理
func (m *ErrorHandlingMiddleware) handleError(c *gin.Context, err error) {
	// 分类错误：根据错误类型确定HTTP状态码和错误消息
	// 例如：验证错误返回400，未找到错误返回404，权限错误返回403
	errorType, statusCode, message := m.classifyError(err)

	// 根据配置决定是否记录错误日志
	// LogAllErrors为true时记录所有错误，false时只记录严重错误
	if m.config.LogAllErrors {
		m.logError(c, err, errorType, statusCode)
	}

	// 如果响应还没有写入，返回错误响应
	// 检查c.Writer.Written()避免重复写入响应
	// 如果响应已写入（如部分成功的情况），不再写入错误响应
	if !c.Writer.Written() {
		// 构建错误响应
		response := gin.H{
			"success": false,  // 请求失败
			"message": message, // 用户友好的错误消息
			"code":    errorType, // 错误代码（用于前端错误处理）
		}

		// 添加详细错误信息（仅在开发环境）
		// 生产环境不返回详细错误信息，避免信息泄露
		if m.config.EnableDetailedErrors {
			response["error"] = err.Error() // 原始错误信息
			// 如果启用错误追踪，添加堆栈跟踪
			if m.config.EnableErrorTracking {
				response["stack"] = string(debug.Stack())
			}
		}

		// 返回错误响应
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
