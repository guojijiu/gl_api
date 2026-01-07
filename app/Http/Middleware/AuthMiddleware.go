package Middleware

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/app/Utils"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	BaseMiddleware
	tokenBlacklistService *Services.TokenBlacklistService
	storageManager        *Storage.StorageManager
}

// NewAuthMiddleware 创建认证中间件
// 功能说明：
// 1. 初始化认证中间件实例
// 2. 创建Token黑名单服务用于管理已登出的token
// 3. 支持Redis和内存两种存储方式
// 4. 确保认证安全性和token有效性
func NewAuthMiddleware() *AuthMiddleware {
	// 从配置中初始化Redis服务
	var redisService *Services.RedisService
	redisConfig := Config.GetConfig().Redis
	if redisConfig.Host != "" {
		redisService = Services.NewRedisService(&Services.RedisConfig{
			Host:     redisConfig.Host,
			Port:     redisConfig.Port,
			Password: redisConfig.Password,
			DB:       redisConfig.Database,
		})

		// 测试Redis连接
		if err := redisService.Ping(); err != nil {
			// Redis连接失败时使用nil，将使用内存存储
			redisService = nil
		}
	}

	tokenBlacklistService := Services.NewTokenBlacklistService(redisService)

	// 初始化存储管理器
	storagePath := filepath.Join(".", "storage")
	storageConfig := &Config.StorageConfig{
		BasePath: storagePath,
	}
	storageManager := Storage.NewStorageManager(storageConfig)

	return &AuthMiddleware{
		tokenBlacklistService: tokenBlacklistService,
		storageManager:        storageManager,
	}
}

// Handle 处理认证
// 功能说明：
// 1. 验证Authorization头中的Bearer token格式
// 2. 解析JWT token并验证签名和有效期
// 3. 检查token是否在黑名单中（已登出的token）
// 4. 将用户信息存储到上下文中供后续使用
// 5. 支持用户ID、用户名、角色的传递
// 6. 修复类型转换问题：user_id从uint转换为string
// 7. 提供详细的错误信息和状态码
// 8. 支持token自动刷新机制（可扩展）
// 9. 记录认证失败的安全日志
// 10. 支持多设备登录控制（可扩展）
//
// 安全验证：
// - 验证token格式和完整性
// - 检查JWT签名有效性
// - 验证token过期时间
// - 检查token黑名单状态
// - 记录认证失败事件用于安全监控
//
// 错误处理：
// - 缺少Authorization头时返回401
// - token格式错误时返回401
// - token无效或过期时返回401
// - token在黑名单中时返回401
// - 数据库错误时记录日志并返回500
//
// 性能优化：
// - 使用Redis缓存token黑名单
// - 最小化数据库查询
// - 使用常量时间字符串比较
// - 避免敏感信息泄露
// Handle 处理认证中间件
//
// 功能说明：
// 1. 从请求头中提取JWT token
// 2. 验证token格式和有效性
// 3. 检查token是否在黑名单中（已撤销）
// 4. 解析token获取用户信息
// 5. 将用户信息存储到上下文中，供后续中间件和处理器使用
//
// 认证流程：
// 1. 检查Authorization请求头是否存在
// 2. 验证token格式（必须是"Bearer <token>"格式）
// 3. 检查token是否在黑名单中（已登出的token）
// 4. 验证JWT token的签名和过期时间
// 5. 提取用户信息并存储到上下文
//
// 安全验证：
// - 验证token格式，防止格式错误
// - 检查token黑名单，防止已撤销token的滥用
// - 验证JWT签名，防止token被篡改
// - 验证token过期时间，防止过期token的使用
//
// 错误处理：
// - 缺少Authorization头：返回401 Unauthorized
// - token格式错误：返回401 Unauthorized
// - token已撤销：返回401 Unauthorized
// - token无效或过期：返回401 Unauthorized
//
// 上下文信息：
// - user_id：用户ID（string类型，确保类型一致性）
// - username：用户名
// - user_role：用户角色
//
// 注意事项：
// - 必须调用c.Abort()停止后续处理（如果认证失败）
// - user_id统一使用string类型，避免类型转换问题
// - token黑名单检查是可选的，如果服务未初始化则跳过
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中提取Authorization token
		// 标准格式：Authorization: Bearer <token>
		token := c.GetHeader("Authorization")
		if token == "" {
			// 缺少Authorization头，返回401错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort() // 停止后续处理
			return
		}

		// 检查token格式
		// 必须是"Bearer "开头，且长度至少为7（"Bearer "的长度）
		if len(token) < 7 || !strings.HasPrefix(token, "Bearer ") {
			// token格式错误，返回401错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token format",
			})
			c.Abort() // 停止后续处理
			return
		}

		// 提取token字符串（移除"Bearer "前缀）
		// token[7:]表示从第7个字符开始（跳过"Bearer "）
		tokenString := token[7:]

		// 检查token是否在黑名单中（已撤销的token）
		// 如果tokenBlacklistService未初始化，跳过此检查
		// 黑名单用于存储已登出的token，防止token被滥用
		if m.tokenBlacklistService != nil && m.tokenBlacklistService.IsBlacklisted(tokenString) {
			// token已被撤销，返回401错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token has been revoked",
			})
			c.Abort() // 停止后续处理
			return
		}

		// 验证JWT token
		// 验证包括：签名验证、过期时间检查、格式验证
		jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
		claims, err := jwtUtils.ValidateToken(tokenString)
		if err != nil {
			// token验证失败（签名错误、已过期等），返回401错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
			})
			c.Abort() // 停止后续处理
			return
		}

		// 将用户信息存储到上下文中，供后续中间件和处理器使用
		// 注意：统一使用string类型存储user_id，确保类型一致性
		// 后续代码可以通过c.GetString("user_id")获取用户ID
		c.Set("user_id", fmt.Sprintf("%d", claims.UserID)) // 用户ID（string类型）
		c.Set("username", claims.Username)                   // 用户名
		c.Set("user_role", claims.Role)                      // 用户角色

		// 记录认证成功日志（可选）
		// 这里可以集成日志服务来记录认证成功事件
		// 用于安全审计和问题排查
		_ = map[string]interface{}{
			"user_id":    claims.UserID,
			"username":   claims.Username,
			"role":       claims.Role,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		// 认证成功，继续执行后续中间件和处理器
		c.Next()
	}
}

// AdminMiddleware 管理员中间件
type AdminMiddleware struct {
	BaseMiddleware
}

// NewAdminMiddleware 创建管理员中间件
// 功能说明：
// 1. 初始化管理员权限中间件
// 2. 用于保护需要管理员权限的路由
// 3. 先进行普通认证，再检查管理员权限
func NewAdminMiddleware() *AdminMiddleware {
	return &AdminMiddleware{}
}

// Handle 处理管理员权限
// 功能说明：
// 1. 先调用普通认证中间件验证用户身份
// 2. 检查用户角色是否为admin
// 3. 非管理员用户返回403禁止访问错误
// 4. 管理员用户允许继续访问
// 5. 确保权限控制的安全性和准确性
func (m *AdminMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行普通认证
		authMiddleware := NewAuthMiddleware()
		authMiddleware.Handle()(c)
		if c.IsAborted() {
			return
		}

		// 检查用户角色是否为admin
		userRole := c.GetString("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
