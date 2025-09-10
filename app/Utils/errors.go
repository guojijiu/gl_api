package Utils

import (
	"errors"
	"fmt"
	"net/http"
)

// 预定义错误类型
var (
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrForbidden       = errors.New("forbidden access")
	ErrValidation      = errors.New("validation failed")
	ErrConflict        = errors.New("resource conflict")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrInternalServer  = errors.New("internal server error")
	ErrBadRequest      = errors.New("bad request")
	ErrTooManyRequests = errors.New("too many requests")
)

// APIError API错误结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"status"`
}

// Error 实现error接口
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError 创建新的API错误
// 功能说明：
// 1. 创建标准化的API错误结构
// 2. 设置错误代码、消息和HTTP状态码
// 3. 用于统一的错误响应格式
func NewAPIError(code, message string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewAPIErrorWithDetails 创建带详情的API错误
// 功能说明：
// 1. 创建包含详细信息的API错误结构
// 2. 设置错误代码、消息、详细信息和HTTP状态码
// 3. 用于提供更详细的错误信息给客户端
func NewAPIErrorWithDetails(code, message, details string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
		Status:  status,
	}
}

// 预定义的API错误
var (
	// 认证相关错误
	ErrInvalidToken = NewAPIError("INVALID_TOKEN", "Invalid or expired token", http.StatusUnauthorized)
	ErrTokenExpired = NewAPIError("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrTokenMissing = NewAPIError("TOKEN_MISSING", "Authorization token is required", http.StatusUnauthorized)

	// 用户相关错误
	ErrUserNotFound    = NewAPIError("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrUserExists      = NewAPIError("USER_EXISTS", "User already exists", http.StatusConflict)
	ErrInvalidPassword = NewAPIError("INVALID_PASSWORD", "Invalid password", http.StatusUnauthorized)
	ErrPasswordTooWeak = NewAPIError("PASSWORD_TOO_WEAK", "Password does not meet requirements", http.StatusBadRequest)

	// 权限相关错误
	ErrInsufficientPermissions = NewAPIError("INSUFFICIENT_PERMISSIONS", "Insufficient permissions", http.StatusForbidden)
	ErrAdminRequired           = NewAPIError("ADMIN_REQUIRED", "Admin privileges required", http.StatusForbidden)

	// 资源相关错误
	ErrResourceNotFound = NewAPIError("RESOURCE_NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrResourceExists   = NewAPIError("RESOURCE_EXISTS", "Resource already exists", http.StatusConflict)
	ErrResourceLocked   = NewAPIError("RESOURCE_LOCKED", "Resource is locked", http.StatusLocked)

	// 验证相关错误
	ErrInvalidInput  = NewAPIError("INVALID_INPUT", "Invalid input data", http.StatusBadRequest)
	ErrMissingField  = NewAPIError("MISSING_FIELD", "Required field is missing", http.StatusBadRequest)
	ErrInvalidFormat = NewAPIError("INVALID_FORMAT", "Invalid data format", http.StatusBadRequest)

	// 系统相关错误
	ErrDatabaseError      = NewAPIError("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError)
	ErrExternalService    = NewAPIError("EXTERNAL_SERVICE_ERROR", "External service error", http.StatusBadGateway)
	ErrServiceUnavailable = NewAPIError("SERVICE_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable)

	// 限制相关错误
	ErrRateLimitExceeded = NewAPIError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", http.StatusTooManyRequests)
	ErrQuotaExceeded     = NewAPIError("QUOTA_EXCEEDED", "Quota exceeded", http.StatusTooManyRequests)
)

// WrapError 包装错误
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapAPIError 包装API错误
func WrapAPIError(err error, apiErr *APIError) *APIError {
	if err == nil {
		return apiErr
	}
	return NewAPIErrorWithDetails(apiErr.Code, apiErr.Message, err.Error(), apiErr.Status)
}

// IsAPIError 检查是否是API错误
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError 获取API错误
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

// GetHTTPStatus 获取HTTP状态码
func GetHTTPStatus(err error) int {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.Status
	}
	return http.StatusInternalServerError
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.Message
	}
	if err != nil {
		return err.Error()
	}
	return "Unknown error"
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors 验证错误集合
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error 实现error接口
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", ve.Errors[0].Message)
}

// Add 添加验证错误
func (ve *ValidationErrors) Add(field, message, value string) {
	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors 检查是否有错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// NewValidationErrors 创建验证错误集合
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// BusinessError 业务错误
type BusinessError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error 实现error接口
func (be *BusinessError) Error() string {
	return be.Message
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WithContext 添加上下文信息
func (be *BusinessError) WithContext(key string, value interface{}) *BusinessError {
	be.Context[key] = value
	return be
}
