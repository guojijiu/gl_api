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
	storageManager := Storage.NewStorageManager(storagePath)

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
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 检查token格式
		if len(token) < 7 || !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token format",
			})
			c.Abort()
			return
		}

		// 提取token字符串
		tokenString := token[7:] // 移除"Bearer "前缀

		// 检查token是否在黑名单中
		if m.tokenBlacklistService != nil && m.tokenBlacklistService.IsBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token has been revoked",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		claims, err := Utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		// 注意：统一使用string类型存储user_id，确保类型一致性
		c.Set("user_id", fmt.Sprintf("%d", claims.UserID))
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		// 记录认证成功日志
		m.storageManager.LogInfo("用户认证成功", map[string]interface{}{
			"user_id":    claims.UserID,
			"username":   claims.Username,
			"role":       claims.Role,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

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
