package Middleware

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware JWT认证中间件
type JWTAuthMiddleware struct {
	BaseMiddleware
	tokenBlacklistService *Services.TokenBlacklistService
}

// Claims JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTAuthMiddleware 创建JWT认证中间件
func NewJWTAuthMiddleware() *JWTAuthMiddleware {
	// 初始化Token黑名单服务
	var redisService *Services.RedisService
	redisConfig := Config.GetConfig().Redis
	if redisConfig.Host != "" {
		redisService = Services.NewRedisService(&Services.RedisConfig{
			Host:     redisConfig.Host,
			Port:     redisConfig.Port,
			Password: redisConfig.Password,
			DB:       redisConfig.Database,
		})
	}

	tokenBlacklistService := Services.NewTokenBlacklistService(redisService)

	return &JWTAuthMiddleware{
		tokenBlacklistService: tokenBlacklistService,
	}
}

// Handle 处理JWT认证
func (m *JWTAuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少认证令牌",
				"code":    "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查Bearer格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的认证格式",
				"code":    "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 检查token是否在黑名单中
		if m.tokenBlacklistService != nil && m.tokenBlacklistService.IsBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "令牌已失效",
				"code":    "TOKEN_BLACKLISTED",
			})
			c.Abort()
			return
		}

		// 解析和验证token
		claims, err := m.parseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的认证令牌",
				"code":    "INVALID_TOKEN",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 检查token是否过期
		if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "令牌已过期",
				"code":    "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// 验证用户是否存在且状态正常
		var user Models.User
		if err := Database.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户不存在",
				"code":    "USER_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.IsActive() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户账户已被禁用",
				"code":    "USER_DISABLED",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("jwt_token", tokenString)

		c.Next()
	}
}

// parseToken 解析JWT token
func (m *JWTAuthMiddleware) parseToken(tokenString string) (*Claims, error) {
	// 获取JWT密钥
	jwtSecret := Config.GetConfig().JWT.Secret
	if jwtSecret == "" {
		return nil, Utils.NewError("JWT密钥未配置", "CONFIG_ERROR", Utils.ErrorLevelError)
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, Utils.NewError("无效的签名方法", "INVALID_SIGNATURE", Utils.ErrorLevelError)
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token是否有效
	if !token.Valid {
		return nil, Utils.NewError("无效的token", "INVALID_TOKEN", Utils.ErrorLevelError)
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, Utils.NewError("无法解析token声明", "INVALID_CLAIMS", Utils.ErrorLevelError)
	}

	return claims, nil
}

// GenerateToken 生成JWT token
func (m *JWTAuthMiddleware) GenerateToken(user *Models.User) (string, error) {
	// 获取JWT配置
	jwtConfig := Config.GetConfig().JWT
	if jwtConfig.Secret == "" {
		return "", Utils.NewError("JWT密钥未配置", "CONFIG_ERROR", Utils.ErrorLevelError)
	}

	// 设置过期时间
	expirationTime := time.Now().Add(time.Duration(jwtConfig.ExpirationHours) * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cloud-platform-api",
			Subject:   user.UUID,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RefreshToken 刷新JWT token
func (m *JWTAuthMiddleware) RefreshToken(tokenString string) (string, error) {
	// 解析现有token
	claims, err := m.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查token是否即将过期（剩余时间少于1小时）
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", Utils.NewError("token尚未到刷新时间", "TOKEN_NOT_READY", Utils.ErrorLevelError)
	}

	// 获取用户信息
	var user Models.User
	if err := Database.DB.First(&user, claims.UserID).Error; err != nil {
		return "", Utils.NewError("用户不存在", "USER_NOT_FOUND", Utils.ErrorLevelError)
	}

	// 生成新token
	return m.GenerateToken(&user)
}

// RequireRole 要求特定角色的中间件
func (m *JWTAuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		m.Handle()(c)

		// 如果已经中止，直接返回
		if c.IsAborted() {
			return
		}

		// 检查用户角色
		userRole := c.GetString("user_role")
		if userRole != requiredRole && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"code":    "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
func (m *JWTAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}
