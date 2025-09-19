package Controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Controller 基础控制器
//
// 重要功能说明：
// 1. 统一响应格式：提供标准化的成功、错误、验证失败等响应方法
// 2. 分页支持：标准化的分页参数验证和响应格式
// 3. 用户认证：获取当前用户信息和角色，支持权限验证
// 4. 错误处理：统一的HTTP状态码和错误消息处理
// 5. 输入验证：分页参数验证和标准化
//
// 安全特性：
// - 统一的错误响应格式，避免信息泄露
// - 用户身份验证和角色检查
// - 输入参数验证和清理
//
// 响应格式标准：
// - 成功响应：{"success": true, "message": "...", "data": {...}}
// - 错误响应：{"success": false, "message": "...", "error": "..."}
// - 分页响应：包含meta字段，包含分页信息
//
// 使用注意事项：
// - 所有控制器都应该继承此基础控制器
// - 使用统一的响应方法确保API一致性
// - 在需要用户身份的操作前调用GetCurrentUser方法
// - 使用IsAdmin方法检查管理员权限
type Controller struct {
	// 可以在这里添加通用的依赖注入
}

// Success 成功响应
func (c *Controller) Success(ctx *gin.Context, data interface{}, message string) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Error 错误响应
func (c *Controller) Error(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"error":   message,
	})
}

// ValidationError 验证错误响应
func (c *Controller) ValidationError(ctx *gin.Context, errors interface{}) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}

// NotFound 未找到响应
func (c *Controller) NotFound(ctx *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	c.Error(ctx, http.StatusNotFound, message)
}

// Unauthorized 未授权响应
func (c *Controller) Unauthorized(ctx *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.Error(ctx, http.StatusUnauthorized, message)
}

// Forbidden 禁止访问响应
func (c *Controller) Forbidden(ctx *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	c.Error(ctx, http.StatusForbidden, message)
}

// ServerError 服务器错误响应
func (c *Controller) ServerError(ctx *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	c.Error(ctx, http.StatusInternalServerError, message)
}

// TooManyRequests 请求过多响应
// 功能说明：
// 1. 返回429状态码的请求过多响应
// 2. 用于速率限制触发
// 3. 包含重试时间信息
func (c *Controller) TooManyRequests(ctx *gin.Context, message string, retryAfter int) {
	ctx.Header("Retry-After", strconv.Itoa(retryAfter))
	ctx.JSON(http.StatusTooManyRequests, gin.H{
		"success":     false,
		"message":     message,
		"retry_after": retryAfter,
	})
}

// PaginatedSuccess 分页成功响应
// 功能说明：
// 1. 返回分页数据的标准格式
// 2. 包含分页元信息（总数、当前页、每页数量等）
// 3. 用于列表查询接口
func (c *Controller) PaginatedSuccess(ctx *gin.Context, data interface{}, total int64, page, pageSize int, message string) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// Created 创建成功响应
// 功能说明：
// 1. 返回201状态码的创建成功响应
// 2. 用于资源创建接口
// 3. 包含创建的资源信息
func (c *Controller) Created(ctx *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "Resource created successfully"
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// NoContent 无内容响应
// 功能说明：
// 1. 返回204状态码的无内容响应
// 2. 用于删除操作成功
// 3. 不返回任何数据
func (c *Controller) NoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

// GetCurrentUser 获取当前用户信息
// 功能说明：
// 1. 从上下文中获取当前登录用户ID
// 2. 用于需要用户身份的操作
// 3. 如果未登录则返回错误
func (c *Controller) GetCurrentUser(ctx *gin.Context) (uint, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user not authenticated")
	}

	// 支持多种类型转换，增加错误处理
	switch v := userID.(type) {
	case uint:
		if v == 0 {
			return 0, fmt.Errorf("invalid user ID: cannot be zero")
		}
		return v, nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("invalid user ID: must be positive")
		}
		return uint(v), nil
	case int64:
		if v <= 0 {
			return 0, fmt.Errorf("invalid user ID: must be positive")
		}
		return uint(v), nil
	case string:
		// 如果是字符串，尝试转换为uint
		if v == "" {
			return 0, fmt.Errorf("invalid user ID: cannot be empty string")
		}
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			if id == 0 {
				return 0, fmt.Errorf("invalid user ID: cannot be zero")
			}
			return uint(id), nil
		}
		return 0, fmt.Errorf("invalid user ID format: %s", v)
	default:
		return 0, fmt.Errorf("invalid user ID type: %T, value: %v", userID, userID)
	}
}

// GetCurrentUserRole 获取当前用户角色
// 功能说明：
// 1. 从上下文中获取当前用户角色
// 2. 用于权限验证
// 3. 如果未登录则返回空字符串
func (c *Controller) GetCurrentUserRole(ctx *gin.Context) string {
	role, exists := ctx.Get("user_role")
	if !exists {
		return ""
	}

	if roleStr, ok := role.(string); ok {
		return roleStr
	}

	return ""
}

// IsAdmin 检查当前用户是否为管理员
// 功能说明：
// 1. 检查当前用户是否具有管理员权限
// 2. 用于管理员功能访问控制
// 3. 返回布尔值表示权限状态
func (c *Controller) IsAdmin(ctx *gin.Context) bool {
	return c.GetCurrentUserRole(ctx) == "admin"
}

// ValidatePagination 验证分页参数
// 功能说明：
// 1. 验证和标准化分页参数
// 2. 设置默认值和限制范围
// 3. 返回有效的分页参数
func (c *Controller) ValidatePagination(ctx *gin.Context) (page, pageSize int) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)

	// 设置默认值和限制
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	return page, pageSize
}
