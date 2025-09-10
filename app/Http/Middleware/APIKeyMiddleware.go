package Middleware

import (
	"cloud-platform-api/app/Storage"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strings"
	"time"
)

// APIKeyConfig API密钥配置
type APIKeyConfig struct {
	SecretKey string        `json:"secret_key"`
	ExpireTime time.Duration `json:"expire_time"`
	MaxRequests int         `json:"max_requests"`
}

// APIKeyMiddleware API密钥认证中间件
type APIKeyMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	config         *APIKeyConfig
	apiKeys        map[string]*APIKeyInfo
}

// APIKeyInfo API密钥信息
type APIKeyInfo struct {
	Key         string    `json:"key"`
	Secret      string    `json:"secret"`
	UserID      uint      `json:"user_id"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsed    time.Time `json:"last_used"`
	RequestCount int      `json:"request_count"`
	IsActive    bool      `json:"is_active"`
}

// NewAPIKeyMiddleware 创建API密钥认证中间件
// 功能说明：
// 1. 初始化API密钥认证中间件
// 2. 提供API密钥验证功能
// 3. 支持密钥过期和权限控制
// 4. 记录API密钥使用日志
func NewAPIKeyMiddleware(storageManager *Storage.StorageManager, config *APIKeyConfig) *APIKeyMiddleware {
	middleware := &APIKeyMiddleware{
		storageManager: storageManager,
		config:         config,
		apiKeys:        make(map[string]*APIKeyInfo),
	}
	
	// 初始化默认API密钥（仅用于开发环境）
	if config != nil && config.SecretKey != "" {
		middleware.initDefaultAPIKeys()
	}
	
	return middleware
}

// Handle 处理API密钥认证
// 功能说明：
// 1. 验证API密钥的有效性
// 2. 检查密钥是否过期
// 3. 验证请求签名
// 4. 记录API密钥使用情况
func (m *APIKeyMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取API密钥
		apiKey := m.extractAPIKey(c)
		if apiKey == "" {
			m.logAPIKeyError(c, "missing_api_key", "缺少API密钥")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少API密钥",
				"error":   "请在请求头中包含有效的API密钥",
			})
			c.Abort()
			return
		}

		// 验证API密钥
		keyInfo, err := m.validateAPIKey(apiKey)
		if err != nil {
			m.logAPIKeyError(c, "invalid_api_key", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的API密钥",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 验证请求签名
		if err := m.validateSignature(c, keyInfo); err != nil {
			m.logAPIKeyError(c, "invalid_signature", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "请求签名无效",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 检查请求频率限制
		if err := m.checkRateLimit(keyInfo); err != nil {
			m.logAPIKeyError(c, "rate_limit_exceeded", err.Error())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求频率超限",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 更新API密钥使用信息
		m.updateAPIKeyUsage(keyInfo)

		// 设置用户信息到上下文
		c.Set("api_key_user_id", keyInfo.UserID)
		c.Set("api_key_permissions", keyInfo.Permissions)
		c.Set("api_key_info", keyInfo)

		// 记录API密钥使用日志
		m.logAPIKeyUsage(c, keyInfo)

		c.Next()
	}
}

// RequirePermission 要求特定权限
// 功能说明：
// 1. 检查API密钥是否具有指定权限
// 2. 支持多个权限的AND逻辑
// 3. 记录权限检查结果
func (m *APIKeyMiddleware) RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyInfo, exists := c.Get("api_key_info")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "需要API密钥认证",
			})
			c.Abort()
			return
		}

		apiKeyInfo := keyInfo.(*APIKeyInfo)
		
		// 检查是否具有所需权限
		hasPermission := m.checkPermissions(apiKeyInfo.Permissions, permissions)
		if !hasPermission {
			m.logAPIKeyError(c, "insufficient_permissions", "权限不足")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "需要权限: " + strings.Join(permissions, " 和 "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractAPIKey 提取API密钥
func (m *APIKeyMiddleware) extractAPIKey(c *gin.Context) string {
	// 1. 从请求头提取
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// 2. 从Authorization头提取
	if auth := c.GetHeader("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// 3. 从查询参数提取
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// validateAPIKey 验证API密钥
func (m *APIKeyMiddleware) validateAPIKey(apiKey string) (*APIKeyInfo, error) {
	// 从缓存或数据库获取API密钥信息
	keyInfo, exists := m.apiKeys[apiKey]
	if !exists {
		// 这里应该从数据库查询API密钥信息
		return nil, fmt.Errorf("API密钥不存在")
	}

	// 检查密钥是否激活
	if !keyInfo.IsActive {
		return nil, fmt.Errorf("API密钥已禁用")
	}

	// 检查密钥是否过期
	if !keyInfo.ExpiresAt.IsZero() && time.Now().After(keyInfo.ExpiresAt) {
		return nil, fmt.Errorf("API密钥已过期")
	}

	return keyInfo, nil
}

// validateSignature 验证请求签名
func (m *APIKeyMiddleware) validateSignature(c *gin.Context, keyInfo *APIKeyInfo) error {
	// 获取签名
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		return fmt.Errorf("缺少请求签名")
	}

	// 获取时间戳
	timestamp := c.GetHeader("X-Timestamp")
	if timestamp == "" {
		return fmt.Errorf("缺少时间戳")
	}

	// 验证时间戳
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("时间戳格式无效")
	}

	// 检查时间戳是否在有效期内（5分钟）
	if time.Since(ts) > 5*time.Minute {
		return fmt.Errorf("请求已过期")
	}

	// 生成签名
	expectedSignature := m.generateSignature(c, keyInfo.Secret, timestamp)
	if signature != expectedSignature {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}

// generateSignature 生成请求签名
func (m *APIKeyMiddleware) generateSignature(c *gin.Context, secret, timestamp string) string {
	// 构建签名字符串
	var params []string
	
	// 添加请求方法
	params = append(params, c.Request.Method)
	
	// 添加请求路径
	params = append(params, c.Request.URL.Path)
	
	// 添加查询参数
	if c.Request.URL.RawQuery != "" {
		params = append(params, c.Request.URL.RawQuery)
	}
	
	// 添加时间戳
	params = append(params, timestamp)
	
	// 添加请求体（如果有）
	if c.Request.Body != nil {
		// 这里应该读取请求体内容
		// 为了简化，暂时使用空字符串
		params = append(params, "")
	}
	
	// 排序参数
	sort.Strings(params)
	
	// 构建签名字符串
	signString := strings.Join(params, "&")
	
	// 使用HMAC-SHA256生成签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signString))
	
	return hex.EncodeToString(h.Sum(nil))
}

// checkRateLimit 检查请求频率限制
func (m *APIKeyMiddleware) checkRateLimit(keyInfo *APIKeyInfo) error {
	if m.config != nil && keyInfo.RequestCount >= m.config.MaxRequests {
		return fmt.Errorf("请求频率超限")
	}
	return nil
}

// checkPermissions 检查权限
func (m *APIKeyMiddleware) checkPermissions(userPermissions, requiredPermissions []string) bool {
	permissionMap := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionMap[perm] = true
	}
	
	for _, required := range requiredPermissions {
		if !permissionMap[required] {
			return false
		}
	}
	
	return true
}

// updateAPIKeyUsage 更新API密钥使用信息
func (m *APIKeyMiddleware) updateAPIKeyUsage(keyInfo *APIKeyInfo) {
	keyInfo.LastUsed = time.Now()
	keyInfo.RequestCount++
}

// initDefaultAPIKeys 初始化默认API密钥
func (m *APIKeyMiddleware) initDefaultAPIKeys() {
	// 仅用于开发环境的默认API密钥
	defaultKey := &APIKeyInfo{
		Key:         "dev_api_key_123456",
		Secret:      m.config.SecretKey,
		UserID:      1,
		Permissions: []string{"read", "write", "admin"},
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0), // 1年后过期
		LastUsed:    time.Now(),
		RequestCount: 0,
		IsActive:    true,
	}
	
	m.apiKeys[defaultKey.Key] = defaultKey
}

// logAPIKeyUsage 记录API密钥使用日志
func (m *APIKeyMiddleware) logAPIKeyUsage(c *gin.Context, keyInfo *APIKeyInfo) {
	m.storageManager.LogInfo("API密钥使用", map[string]interface{}{
		"api_key":    keyInfo.Key,
		"user_id":    keyInfo.UserID,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})
}

// logAPIKeyError 记录API密钥错误日志
func (m *APIKeyMiddleware) logAPIKeyError(c *gin.Context, errorType, reason string) {
	m.storageManager.LogWarning("API密钥错误", map[string]interface{}{
		"error_type": errorType,
		"reason":     reason,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})
}

// CreateAPIKey 创建新的API密钥
func (m *APIKeyMiddleware) CreateAPIKey(userID uint, permissions []string, expiresAt time.Time) (*APIKeyInfo, error) {
	// 生成API密钥
	apiKey := m.generateAPIKey()
	secret := m.generateSecret()
	
	keyInfo := &APIKeyInfo{
		Key:         apiKey,
		Secret:      secret,
		UserID:      userID,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		LastUsed:    time.Now(),
		RequestCount: 0,
		IsActive:    true,
	}
	
	// 保存到内存（实际应该保存到数据库）
	m.apiKeys[apiKey] = keyInfo
	
	// 记录创建日志
	m.storageManager.LogInfo("API密钥已创建", map[string]interface{}{
		"api_key":     apiKey,
		"user_id":     userID,
		"permissions": permissions,
		"expires_at":  expiresAt,
	})
	
	return keyInfo, nil
}

// generateAPIKey 生成API密钥
func (m *APIKeyMiddleware) generateAPIKey() string {
	// 这里应该使用更安全的随机生成方法
	return fmt.Sprintf("api_key_%d", time.Now().UnixNano())
}

// generateSecret 生成密钥
func (m *APIKeyMiddleware) generateSecret() string {
	// 这里应该使用更安全的随机生成方法
	return fmt.Sprintf("secret_%d", time.Now().UnixNano())
}

// RevokeAPIKey 撤销API密钥
func (m *APIKeyMiddleware) RevokeAPIKey(apiKey string) error {
	keyInfo, exists := m.apiKeys[apiKey]
	if !exists {
		return fmt.Errorf("API密钥不存在")
	}
	
	keyInfo.IsActive = false
	
	// 记录撤销日志
	m.storageManager.LogInfo("API密钥已撤销", map[string]interface{}{
		"api_key": apiKey,
		"user_id": keyInfo.UserID,
	})
	
	return nil
}
