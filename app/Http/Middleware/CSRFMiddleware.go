package Middleware

import (
	"cloud-platform-api/app/Storage"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFConfig CSRF配置
type CSRFConfig struct {
	SecretKey     string        `json:"secret_key"`
	TokenLength   int           `json:"token_length"`
	TokenExpire   time.Duration `json:"token_expire"`
	HeaderName    string        `json:"header_name"`
	FormFieldName string        `json:"form_field_name"`
	CookieName    string        `json:"cookie_name"`
}

// CSRFMiddleware CSRF保护中间件
type CSRFMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	config         *CSRFConfig
	tokens         sync.Map // 使用sync.Map替代普通map，解决并发安全问题
}

// CSRFToken CSRF令牌信息
type CSRFToken struct {
	Token     string    `json:"token"`
	UserID    uint      `json:"user_id"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// NewCSRFMiddleware 创建CSRF保护中间件
// 功能说明：
// 1. 初始化CSRF保护中间件
// 2. 生成和验证CSRF令牌
// 3. 防止跨站请求伪造攻击
// 4. 支持令牌过期和一次性使用
func NewCSRFMiddleware(storageManager *Storage.StorageManager, config *CSRFConfig) *CSRFMiddleware {
	if config == nil {
		config = &CSRFConfig{
			SecretKey:     "csrf_secret_key_change_in_production",
			TokenLength:   32,
			TokenExpire:   1 * time.Hour,
			HeaderName:    "X-CSRF-Token",
			FormFieldName: "_csrf_token",
			CookieName:    "csrf_token",
		}
	}

	middleware := &CSRFMiddleware{
		storageManager: storageManager,
		config:         config,
		// tokens字段使用sync.Map，不需要初始化
	}

	// 启动令牌清理协程
	go middleware.cleanupExpiredTokens()

	return middleware
}

// Handle 处理CSRF保护
// 功能说明：
// 1. 为GET请求生成CSRF令牌
// 2. 为POST/PUT/DELETE请求验证CSRF令牌
// 3. 设置CSRF令牌到Cookie和响应头
// 4. 记录CSRF验证日志
func (m *CSRFMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要CSRF保护的路由
		if m.shouldSkipCSRF(c) {
			c.Next()
			return
		}

		// 获取用户ID和会话ID
		userID := m.getUserID(c)
		sessionID := m.getSessionID(c)

		switch c.Request.Method {
		case "GET":
			// 为GET请求生成CSRF令牌
			token := m.generateCSRFToken(userID, sessionID)
			m.setCSRFToken(c, token)
			c.Next()

		case "POST", "PUT", "DELETE", "PATCH":
			// 为修改请求验证CSRF令牌
			if err := m.validateCSRFToken(c, userID, sessionID); err != nil {
				m.logCSRFError(c, "token_validation_failed", err.Error())
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "CSRF验证失败",
					"error":   err.Error(),
				})
				c.Abort()
				return
			}

			// 验证成功，继续处理
			m.logCSRFSuccess(c, "token_validation_success")
			c.Next()

		default:
			// 其他方法跳过CSRF验证
			c.Next()
		}
	}
}

// shouldSkipCSRF 检查是否应该跳过CSRF保护
func (m *CSRFMiddleware) shouldSkipCSRF(c *gin.Context) bool {
	// 跳过API路由
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return true
	}

	// 跳过静态文件
	if strings.HasPrefix(c.Request.URL.Path, "/static/") ||
		strings.HasPrefix(c.Request.URL.Path, "/assets/") {
		return true
	}

	// 跳过健康检查
	if c.Request.URL.Path == "/health" {
		return true
	}

	// 跳过OPTIONS请求
	if c.Request.Method == "OPTIONS" {
		return true
	}

	return false
}

// generateCSRFToken 生成CSRF令牌
func (m *CSRFMiddleware) generateCSRFToken(userID uint, sessionID string) *CSRFToken {
	// 生成随机令牌
	tokenBytes := make([]byte, m.config.TokenLength)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// 创建令牌信息
	csrfToken := &CSRFToken{
		Token:     token,
		UserID:    userID,
		SessionID: sessionID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.config.TokenExpire),
		Used:      false,
	}

	// 存储令牌（使用sync.Map的Store方法）
	m.tokens.Store(token, csrfToken)

	// 记录令牌生成日志
	m.storageManager.LogInfo("CSRF令牌已生成", map[string]interface{}{
		"token":      token,
		"user_id":    userID,
		"session_id": sessionID,
		"expires_at": csrfToken.ExpiresAt,
	})

	return csrfToken
}

// validateCSRFToken 验证CSRF令牌
func (m *CSRFMiddleware) validateCSRFToken(c *gin.Context, userID uint, sessionID string) error {
	// 获取CSRF令牌
	token := m.extractCSRFToken(c)
	if token == "" {
		return fmt.Errorf("缺少CSRF令牌")
	}

	// 查找令牌（使用sync.Map的Load方法）
	tokenValue, exists := m.tokens.Load(token)
	if !exists {
		return fmt.Errorf("CSRF令牌不存在")
	}
	csrfToken, ok := tokenValue.(*CSRFToken)
	if !ok {
		return fmt.Errorf("CSRF令牌类型错误")
	}

	// 检查令牌是否过期
	if time.Now().After(csrfToken.ExpiresAt) {
		// 清理过期令牌（使用sync.Map的Delete方法）
		m.tokens.Delete(token)
		return fmt.Errorf("CSRF令牌已过期")
	}

	// 检查令牌是否已被使用
	if csrfToken.Used {
		return fmt.Errorf("CSRF令牌已被使用")
	}

	// 检查用户ID是否匹配
	if csrfToken.UserID != userID {
		return fmt.Errorf("CSRF令牌用户不匹配")
	}

	// 检查会话ID是否匹配
	if csrfToken.SessionID != sessionID {
		return fmt.Errorf("CSRF令牌会话不匹配")
	}

	// 标记令牌为已使用
	csrfToken.Used = true

	// 清理已使用的令牌（使用sync.Map的Delete方法）
	m.tokens.Delete(token)

	return nil
}

// extractCSRFToken 提取CSRF令牌
func (m *CSRFMiddleware) extractCSRFToken(c *gin.Context) string {
	// 1. 从请求头提取
	if token := c.GetHeader(m.config.HeaderName); token != "" {
		return token
	}

	// 2. 从表单字段提取
	if token := c.PostForm(m.config.FormFieldName); token != "" {
		return token
	}

	// 3. 从查询参数提取
	if token := c.Query(m.config.FormFieldName); token != "" {
		return token
	}

	// 4. 从Cookie提取
	if cookie, err := c.Cookie(m.config.CookieName); err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// setCSRFToken 设置CSRF令牌
func (m *CSRFMiddleware) setCSRFToken(c *gin.Context, token *CSRFToken) {
	// 设置响应头
	c.Header(m.config.HeaderName, token.Token)

	// 设置Cookie
	c.SetCookie(
		m.config.CookieName,
		token.Token,
		int(m.config.TokenExpire.Seconds()),
		"/",
		"",
		false, // 生产环境应该设置为true（HTTPS）
		true,  // HttpOnly
	)

	// 将令牌添加到响应数据中（用于前端获取）
	c.Set("csrf_token", token.Token)
}

// getUserID 获取用户ID
func (m *CSRFMiddleware) getUserID(c *gin.Context) uint {
	// 从JWT认证中获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}

	// 从API密钥认证中获取用户ID
	if userID, exists := c.Get("api_key_user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}

	return 0
}

// getSessionID 获取会话ID
func (m *CSRFMiddleware) getSessionID(c *gin.Context) string {
	// 从Cookie中获取会话ID
	if sessionID, err := c.Cookie("session_id"); err == nil {
		return sessionID
	}

	// 从请求头中获取会话ID
	if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
		return sessionID
	}

	// 生成临时会话ID
	return fmt.Sprintf("temp_session_%d", time.Now().Unix())
}

// cleanupExpiredTokens 清理过期的令牌
func (m *CSRFMiddleware) cleanupExpiredTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		cleaned := 0
		totalCount := 0

		// 使用sync.Map的Range方法遍历所有令牌
		m.tokens.Range(func(key, value interface{}) bool {
			totalCount++
			token := key.(string)
			csrfToken := value.(*CSRFToken)

			if now.After(csrfToken.ExpiresAt) || csrfToken.Used {
				m.tokens.Delete(token)
				cleaned++
			}
			return true // 继续遍历
		})

		if cleaned > 0 {
			m.storageManager.LogInfo("CSRF令牌清理完成", map[string]interface{}{
				"cleaned_count":   cleaned,
				"remaining_count": totalCount - cleaned,
			})
		}
	}
}

// logCSRFError 记录CSRF错误日志
func (m *CSRFMiddleware) logCSRFError(c *gin.Context, errorType, reason string) {
	m.storageManager.LogWarning("CSRF验证失败", map[string]interface{}{
		"error_type": errorType,
		"reason":     reason,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})
}

// logCSRFSuccess 记录CSRF成功日志
func (m *CSRFMiddleware) logCSRFSuccess(c *gin.Context, action string) {
	m.storageManager.LogInfo("CSRF验证成功", map[string]interface{}{
		"action":    action,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
	})
}

// GetCSRFToken 获取CSRF令牌（用于前端）
func (m *CSRFMiddleware) GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}

// RefreshCSRFToken 刷新CSRF令牌
func (m *CSRFMiddleware) RefreshCSRFToken(c *gin.Context) *CSRFToken {
	userID := m.getUserID(c)
	sessionID := m.getSessionID(c)

	// 生成新令牌
	token := m.generateCSRFToken(userID, sessionID)

	// 设置令牌
	m.setCSRFToken(c, token)

	return token
}

// GetCSRFStats 获取CSRF统计信息
func (m *CSRFMiddleware) GetCSRFStats() map[string]interface{} {
	now := time.Now()
	activeTokens := 0
	expiredTokens := 0
	usedTokens := 0
	totalTokens := 0

	// 使用sync.Map的Range方法遍历所有令牌
	m.tokens.Range(func(key, value interface{}) bool {
		totalTokens++
		token := value.(*CSRFToken)

		if token.Used {
			usedTokens++
		} else if now.After(token.ExpiresAt) {
			expiredTokens++
		} else {
			activeTokens++
		}
		return true // 继续遍历
	})

	return map[string]interface{}{
		"total_tokens":   totalTokens,
		"active_tokens":  activeTokens,
		"expired_tokens": expiredTokens,
		"used_tokens":    usedTokens,
		"token_expire":   m.config.TokenExpire.String(),
		"token_length":   m.config.TokenLength,
	}
}
