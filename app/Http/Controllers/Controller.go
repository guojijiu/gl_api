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
//
// 功能说明：
// 1. 返回标准的成功响应格式
// 2. 统一所有成功响应的格式，确保API一致性
// 3. 使用200 OK状态码
//
// 响应格式：
// {
//   "success": true,
//   "message": "操作成功",
//   "data": {...}
// }
//
// 使用场景：
// - 创建资源成功
// - 更新资源成功
// - 查询资源成功
// - 删除资源成功
//
// 注意事项：
// - 所有成功响应都应该使用此方法
// - message应该简洁明了，描述操作结果
// - data可以是任何类型（对象、数组、字符串等）
func (c *Controller) Success(ctx *gin.Context, data interface{}, message string) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,  // 操作成功标志
		"message": message, // 成功消息
		"data":    data,    // 返回的数据
	})
}

// Error 错误响应
//
// 功能说明：
// 1. 返回标准的错误响应格式
// 2. 统一所有错误响应的格式，确保API一致性
// 3. 支持自定义HTTP状态码
//
// 响应格式：
// {
//   "success": false,
//   "message": "错误消息",
//   "error": "错误消息"
// }
//
// 使用场景：
// - 业务逻辑错误
// - 数据验证失败
// - 权限不足
// - 资源不存在
//
// 状态码说明：
// - 400 Bad Request: 请求参数错误
// - 401 Unauthorized: 未授权
// - 403 Forbidden: 禁止访问
// - 404 Not Found: 资源不存在
// - 500 Internal Server Error: 服务器错误
//
// 注意事项：
// - 所有错误响应都应该使用此方法或特定的错误方法
// - message应该用户友好，避免泄露系统信息
// - 生产环境不应该返回详细的错误堆栈
func (c *Controller) Error(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"success": false,  // 操作失败标志
		"message": message, // 错误消息
		"error":   message, // 错误详情（与message相同）
	})
}

// ValidationError 验证错误响应
//
// 功能说明：
// 1. 返回标准的验证错误响应格式
// 2. 用于请求参数验证失败的情况
// 3. 包含详细的验证错误信息
//
// 响应格式：
// {
//   "success": false,
//   "message": "Validation failed",
//   "errors": {...}  // 验证错误详情（可以是对象或数组）
// }
//
// 参数说明：
// - errors: 验证错误信息（可以是map[string]string或[]string）
//
// 使用场景：
// - 请求参数格式错误
// - 必填字段缺失
// - 字段类型不匹配
// - 字段值不符合规则
//
// 注意事项：
// - 使用400 Bad Request状态码
// - errors应该包含具体的验证错误信息
// - 错误信息应该用户友好，便于前端显示
func (c *Controller) ValidationError(ctx *gin.Context, errors interface{}) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}

// NotFound 未找到响应
//
// 功能说明：
// 1. 返回404 Not Found错误响应
// 2. 用于资源不存在的情况
// 3. 支持自定义错误消息
//
// 参数说明：
// - message: 错误消息（如果为空，使用默认消息）
//
// 使用场景：
// - 用户不存在
// - 文章不存在
// - 资源ID无效
// - 路由不存在
//
// 注意事项：
// - 使用404 Not Found状态码
// - 消息应该简洁明了
// - 不应该泄露系统内部信息
func (c *Controller) NotFound(ctx *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	c.Error(ctx, http.StatusNotFound, message)
}

// Unauthorized 未授权响应
//
// 功能说明：
// 1. 返回401 Unauthorized错误响应
// 2. 用于未认证或认证失败的情况
// 3. 支持自定义错误消息
//
// 参数说明：
// - message: 错误消息（如果为空，使用默认消息）
//
// 使用场景：
// - 未提供认证token
// - Token无效或过期
// - 认证失败
//
// 注意事项：
// - 使用401 Unauthorized状态码
// - 客户端应该重新登录
// - 不应该泄露认证失败的具体原因
func (c *Controller) Unauthorized(ctx *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.Error(ctx, http.StatusUnauthorized, message)
}

// Forbidden 禁止访问响应
//
// 功能说明：
// 1. 返回403 Forbidden错误响应
// 2. 用于权限不足的情况
// 3. 支持自定义错误消息
//
// 参数说明：
// - message: 错误消息（如果为空，使用默认消息）
//
// 使用场景：
// - 用户权限不足
// - 角色不允许访问
// - 资源访问受限
//
// 注意事项：
// - 使用403 Forbidden状态码
// - 与401的区别：401是未认证，403是已认证但权限不足
// - 不应该泄露权限系统的详细信息
func (c *Controller) Forbidden(ctx *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	c.Error(ctx, http.StatusForbidden, message)
}

// ServerError 服务器错误响应
//
// 功能说明：
// 1. 返回500 Internal Server Error错误响应
// 2. 用于服务器内部错误
// 3. 支持自定义错误消息
//
// 参数说明：
// - message: 错误消息（如果为空，使用默认消息）
//
// 使用场景：
// - 数据库操作失败
// - 业务逻辑错误
// - 系统异常
//
// 注意事项：
// - 使用500 Internal Server Error状态码
// - 生产环境不应该返回详细的错误信息
// - 应该记录详细的错误日志用于调试
// - 用户看到的消息应该友好，不泄露技术细节
func (c *Controller) ServerError(ctx *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	c.Error(ctx, http.StatusInternalServerError, message)
}

// TooManyRequests 请求过多响应
//
// 功能说明：
// 1. 返回429 Too Many Requests错误响应
// 2. 用于速率限制触发的情况
// 3. 包含重试时间信息（Retry-After头）
//
// 参数说明：
// - message: 错误消息
// - retryAfter: 重试等待时间（秒）
//
// 响应格式：
// {
//   "success": false,
//   "message": "请求频率超限",
//   "retry_after": 60
// }
// 响应头：Retry-After: 60
//
// 使用场景：
// - API速率限制触发
// - 防止API滥用
// - 保护系统资源
//
// 注意事项：
// - 使用429 Too Many Requests状态码
// - 必须设置Retry-After响应头
// - retryAfter应该合理设置，不要过长
// - 客户端应该根据Retry-After头等待后重试
func (c *Controller) TooManyRequests(ctx *gin.Context, message string, retryAfter int) {
	ctx.Header("Retry-After", strconv.Itoa(retryAfter))
	ctx.JSON(http.StatusTooManyRequests, gin.H{
		"success":     false,
		"message":     message,
		"retry_after": retryAfter,
	})
}

// PaginatedSuccess 分页成功响应
//
// 功能说明：
// 1. 返回分页数据的标准格式
// 2. 包含分页元信息（总数、当前页、每页数量等）
// 3. 用于列表查询接口
//
// 响应格式：
// {
//   "success": true,
//   "message": "查询成功",
//   "data": [...],
//   "meta": {
//     "total": 100,
//     "page": 1,
//     "page_size": 10,
//     "total_pages": 10
//   }
// }
//
// 参数说明：
// - data: 当前页的数据列表
// - total: 总记录数
// - page: 当前页码（从1开始）
// - pageSize: 每页记录数
// - message: 成功消息
//
// 使用场景：
// - 用户列表查询
// - 文章列表查询
// - 订单列表查询
// - 任何需要分页的列表接口
//
// 注意事项：
// - page应该从1开始，不是0
// - total_pages会自动计算：ceil(total / pageSize)
// - 如果total为0，total_pages为0
// - data通常是数组类型
func (c *Controller) PaginatedSuccess(ctx *gin.Context, data interface{}, total int64, page, pageSize int, message string) {
	// 计算总页数
	// 使用整数除法，向上取整：ceil(total / pageSize) = (total + pageSize - 1) / pageSize
	totalPages := (int(total) + pageSize - 1) / pageSize
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages, // 使用计算好的总页数
		},
	})
}

// Created 创建成功响应
//
// 功能说明：
// 1. 返回201 Created成功响应
// 2. 用于资源创建接口（POST请求）
// 3. 包含创建的资源信息
//
// 响应格式：
// {
//   "success": true,
//   "message": "Resource created successfully",
//   "data": {...}  // 创建的资源对象
// }
//
// 参数说明：
// - data: 创建的资源对象（包含ID等新生成的信息）
// - message: 成功消息（如果为空，使用默认消息）
//
// 使用场景：
// - 用户注册成功
// - 文章创建成功
// - 订单创建成功
// - 任何资源创建操作
//
// 注意事项：
// - 使用201 Created状态码（符合RESTful规范）
// - data应该包含创建后的完整资源信息
// - 通常包含新生成的ID和其他系统生成的字段
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
//
// 功能说明：
// 1. 返回204 No Content响应
// 2. 用于删除操作成功或更新操作成功
// 3. 不返回任何响应体（符合HTTP规范）
//
// 使用场景：
// - 资源删除成功（DELETE请求）
// - 资源更新成功（PUT/PATCH请求，不需要返回数据）
// - 批量操作成功
//
// 注意事项：
// - 使用204 No Content状态码（符合RESTful规范）
// - 不返回任何响应体，只返回状态码
// - 客户端应该根据状态码判断操作是否成功
// - 与200 OK的区别：204表示成功但无内容返回
func (c *Controller) NoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

// GetCurrentUser 获取当前用户信息
//
// 功能说明：
// 1. 从上下文中获取当前登录用户ID
// 2. 支持多种类型转换（uint、int、int64、string）
// 3. 验证用户ID的有效性（不能为0或负数）
// 4. 如果未登录或用户ID无效则返回错误
//
// 返回信息：
// - uint: 用户ID（如果成功）
// - error: 错误信息（如果未登录、用户ID不存在或格式无效）
//
// 类型转换支持：
// - uint: 直接返回（验证不为0）
// - int: 转换为uint（验证大于0）
// - int64: 转换为uint（验证大于0）
// - string: 解析为uint（验证不为空且不为0）
//
// 使用场景：
// - 需要用户身份的操作（创建、更新、删除）
// - 权限验证（检查是否为资源所有者）
// - 审计日志（记录操作用户）
//
// 注意事项：
// - 用户ID必须通过认证中间件设置到上下文中
// - 用户ID不能为0或负数
// - 如果类型转换失败，返回详细的错误信息
// - 建议在需要用户身份的操作前调用此方法
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
//
// 功能说明：
// 1. 从上下文中获取当前用户角色
// 2. 用于权限验证和角色检查
// 3. 如果未登录或角色未设置则返回空字符串
//
// 返回信息：
// - string: 用户角色（如"admin"、"user"等）
//
// 使用场景：
// - 权限验证（检查是否为管理员）
// - 角色基础访问控制（RBAC）
// - 功能权限判断
//
// 注意事项：
// - 角色必须通过认证中间件设置到上下文中
// - 如果未登录，返回空字符串（不会报错）
// - 建议与IsAdmin方法配合使用
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
//
// 功能说明：
// 1. 检查当前用户是否具有管理员权限
// 2. 用于管理员功能访问控制
// 3. 返回布尔值表示权限状态
//
// 返回信息：
// - bool: true表示是管理员，false表示不是管理员
//
// 使用场景：
// - 管理员功能访问控制
// - 权限验证（在控制器方法中）
// - 条件判断（是否显示管理员功能）
//
// 注意事项：
// - 依赖于GetCurrentUserRole方法
// - 如果用户未登录，返回false
// - 角色必须完全匹配"admin"（区分大小写）
func (c *Controller) IsAdmin(ctx *gin.Context) bool {
	return c.GetCurrentUserRole(ctx) == "admin"
}

// ValidatePagination 验证分页参数
//
// 功能说明：
// 1. 从请求中提取分页参数（page和page_size）
// 2. 验证和标准化分页参数（设置默认值和限制范围）
// 3. 返回有效的分页参数（确保在合理范围内）
//
// 参数来源：
// - page: 从查询参数"page"获取（默认值：1）
// - page_size: 从查询参数"page_size"获取（默认值：10）
//
// 验证规则：
// - page: 最小值为1，小于1时设置为1
// - page_size: 最小值为1（小于1时设置为10），最大值为100（超过100时设置为100）
//
// 返回信息：
// - page: 验证后的页码（从1开始）
// - pageSize: 验证后的每页数量（1-100之间）
//
// 使用场景：
// - 列表查询接口（用户列表、文章列表等）
// - 需要分页的数据查询
// - 统一分页参数验证
//
// 注意事项：
// - page从1开始，不是0
// - page_size有最大值限制（100），防止一次性查询过多数据
// - 如果参数解析失败，使用默认值
// - 建议在所有列表查询接口中使用此方法
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
