package Middleware

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	EnableRateLimit     bool          `json:"enable_rate_limit"`
	DefaultLimit        int           `json:"default_limit"`        // 默认限制次数
	DefaultWindow       time.Duration `json:"default_window"`       // 默认时间窗口
	EnableIPBlacklist   bool          `json:"enable_ip_blacklist"`  // 启用IP黑名单
	BlacklistThreshold  int           `json:"blacklist_threshold"`  // 黑名单阈值
	BlacklistDuration   time.Duration `json:"blacklist_duration"`   // 黑名单持续时间
	EnableWhitelist     bool          `json:"enable_whitelist"`     // 启用IP白名单
	WhitelistIPs        []string      `json:"whitelist_ips"`        // 白名单IP列表
	EnableUserRateLimit bool          `json:"enable_user_rate_limit"` // 启用用户级别限流
	UserLimitMultiplier float64       `json:"user_limit_multiplier"`  // 用户限流倍数
}

// RateLimitRecord 限流记录
type RateLimitRecord struct {
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// BlacklistRecord 黑名单记录
type BlacklistRecord struct {
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	AddedAt   time.Time `json:"added_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AdvancedRateLimitMiddleware 高级限流中间件
type AdvancedRateLimitMiddleware struct {
	storageManager *Storage.StorageManager
	config         *RateLimitConfig
	rateLimits     map[string]*RateLimitRecord
	blacklist      map[string]*BlacklistRecord
	mutex          sync.RWMutex
}

// NewAdvancedRateLimitMiddleware 创建高级限流中间件
// 功能说明：
// 1. 初始化高级限流中间件
// 2. 支持多种限流策略
// 3. 支持IP黑名单和白名单
// 4. 支持用户级别限流
// 5. 提供详细的限流统计
func NewAdvancedRateLimitMiddleware(storageManager *Storage.StorageManager, config *RateLimitConfig) *AdvancedRateLimitMiddleware {
	if config == nil {
		config = &RateLimitConfig{
			EnableRateLimit:     true,
			DefaultLimit:        100,
			DefaultWindow:       1 * time.Minute,
			EnableIPBlacklist:   true,
			BlacklistThreshold:  1000,
			BlacklistDuration:   1 * time.Hour,
			EnableWhitelist:     false,
			WhitelistIPs:        []string{},
			EnableUserRateLimit: true,
			UserLimitMultiplier: 2.0,
		}
	}

	middleware := &AdvancedRateLimitMiddleware{
		storageManager: storageManager,
		config:         config,
		rateLimits:     make(map[string]*RateLimitRecord),
		blacklist:      make(map[string]*BlacklistRecord),
	}

	// 启动清理协程
	go middleware.cleanupExpiredRecords()

	return middleware
}

// Handle 处理限流
// 功能说明：
// 1. 检查IP是否在黑名单中
// 2. 检查IP是否在白名单中
// 3. 应用限流规则
// 4. 记录限流统计
// 5. 自动管理黑名单
func (m *AdvancedRateLimitMiddleware) Handle(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.EnableRateLimit {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		userID := c.GetString("user_id")

		// 1. 检查IP黑名单
		if m.config.EnableIPBlacklist && m.isIPBlacklisted(clientIP) {
			m.logRateLimitEvent(c, "ip_blacklisted", "IP在黑名单中")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "IP已被限制访问",
				"code":    "IP_BLACKLISTED",
			})
			c.Abort()
			return
		}

		// 2. 检查IP白名单
		if m.config.EnableWhitelist && m.isIPWhitelisted(clientIP) {
			c.Next()
			return
		}

		// 3. 应用限流规则
		key := m.generateRateLimitKey(clientIP, userID)
		limit, window := m.getRateLimitConfig(limit, window, userID)

		if !m.checkRateLimit(key, limit, window) {
			m.logRateLimitEvent(c, "rate_limit_exceeded", fmt.Sprintf("超过限流: %d/%s", limit, window))
			
			// 检查是否需要加入黑名单
			if m.config.EnableIPBlacklist {
				m.checkAndAddToBlacklist(clientIP)
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求频率超限",
				"code":    "RATE_LIMIT_EXCEEDED",
				"retry_after": m.getRetryAfter(key, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HandleWithCustomConfig 处理自定义配置的限流
func (m *AdvancedRateLimitMiddleware) HandleWithCustomConfig(config map[string]RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.EnableRateLimit {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		userID := c.GetString("user_id")
		path := c.Request.URL.Path

		// 获取路径特定的配置
		pathConfig, exists := config[path]
		if !exists {
			pathConfig = RateLimitConfig{
				DefaultLimit:  m.config.DefaultLimit,
				DefaultWindow: m.config.DefaultWindow,
			}
		}

		// 应用限流
		key := m.generateRateLimitKey(clientIP, userID)
		limit, window := m.getRateLimitConfig(pathConfig.DefaultLimit, pathConfig.DefaultWindow, userID)

		if !m.checkRateLimit(key, limit, window) {
			m.logRateLimitEvent(c, "rate_limit_exceeded", fmt.Sprintf("路径 %s 超过限流", path))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求频率超限",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查限流
func (m *AdvancedRateLimitMiddleware) checkRateLimit(key string, limit int, window time.Duration) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	record, exists := m.rateLimits[key]

	if !exists {
		// 创建新记录
		m.rateLimits[key] = &RateLimitRecord{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return true
	}

	// 检查时间窗口
	if now.Sub(record.FirstSeen) > window {
		// 重置记录
		record.Count = 1
		record.FirstSeen = now
		record.LastSeen = now
		return true
	}

	// 检查是否超过限制
	if record.Count >= limit {
		return false
	}

	// 增加计数
	record.Count++
	record.LastSeen = now
	return true
}

// generateRateLimitKey 生成限流键
func (m *AdvancedRateLimitMiddleware) generateRateLimitKey(clientIP, userID string) string {
	if m.config.EnableUserRateLimit && userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}
	return fmt.Sprintf("ip:%s", clientIP)
}

// getRateLimitConfig 获取限流配置
func (m *AdvancedRateLimitMiddleware) getRateLimitConfig(limit int, window time.Duration, userID string) (int, time.Duration) {
	if m.config.EnableUserRateLimit && userID != "" {
		// 用户级别限流，给予更高的限制
		userLimit := int(float64(limit) * m.config.UserLimitMultiplier)
		return userLimit, window
	}
	return limit, window
}

// isIPBlacklisted 检查IP是否在黑名单中
func (m *AdvancedRateLimitMiddleware) isIPBlacklisted(ip string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	record, exists := m.blacklist[ip]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(record.ExpiresAt) {
		delete(m.blacklist, ip)
		return false
	}

	return true
}

// isIPWhitelisted 检查IP是否在白名单中
func (m *AdvancedRateLimitMiddleware) isIPWhitelisted(ip string) bool {
	for _, whitelistIP := range m.config.WhitelistIPs {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

// checkAndAddToBlacklist 检查并添加到黑名单
func (m *AdvancedRateLimitMiddleware) checkAndAddToBlacklist(ip string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查当前IP的请求次数
	key := fmt.Sprintf("ip:%s", ip)
	record, exists := m.rateLimits[key]
	if !exists {
		return
	}

	// 如果超过黑名单阈值，添加到黑名单
	if record.Count >= m.config.BlacklistThreshold {
		m.blacklist[ip] = &BlacklistRecord{
			IP:        ip,
			Reason:    "超过限流阈值",
			AddedAt:   time.Now(),
			ExpiresAt: time.Now().Add(m.config.BlacklistDuration),
		}

		m.storageManager.LogWarning("IP已加入黑名单", map[string]interface{}{
			"ip":           ip,
			"reason":       "超过限流阈值",
			"count":        record.Count,
			"threshold":    m.config.BlacklistThreshold,
			"expires_at":   time.Now().Add(m.config.BlacklistDuration),
		})
	}
}

// getRetryAfter 获取重试时间
func (m *AdvancedRateLimitMiddleware) getRetryAfter(key string, window time.Duration) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	record, exists := m.rateLimits[key]
	if !exists {
		return 0
	}

	remaining := window - time.Since(record.FirstSeen)
	if remaining < 0 {
		return 0
	}

	return int(remaining.Seconds())
}

// cleanupExpiredRecords 清理过期记录
func (m *AdvancedRateLimitMiddleware) cleanupExpiredRecords() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()

		now := time.Now()

		// 清理过期的限流记录
		for key, record := range m.rateLimits {
			if now.Sub(record.LastSeen) > 1*time.Hour {
				delete(m.rateLimits, key)
			}
		}

		// 清理过期的黑名单记录
		for ip, record := range m.blacklist {
			if now.After(record.ExpiresAt) {
				delete(m.blacklist, ip)
			}
		}

		m.mutex.Unlock()
	}
}

// logRateLimitEvent 记录限流事件
func (m *AdvancedRateLimitMiddleware) logRateLimitEvent(c *gin.Context, eventType, details string) {
	m.storageManager.LogWarning("限流事件", map[string]interface{}{
		"event_type": eventType,
		"details":    details,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"method":     c.Request.Method,
		"url":        c.Request.URL.String(),
		"user_id":    c.GetString("user_id"),
	})
}

// GetRateLimitStats 获取限流统计
func (m *AdvancedRateLimitMiddleware) GetRateLimitStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"active_rate_limits": len(m.rateLimits),
		"blacklisted_ips":    len(m.blacklist),
		"config":             m.config,
	}

	// 统计最活跃的IP
	ipStats := make(map[string]int)
	for key, record := range m.rateLimits {
		if len(key) > 3 && key[:3] == "ip:" {
			ip := key[3:]
			ipStats[ip] = record.Count
		}
	}

	stats["ip_statistics"] = ipStats
	return stats
}

// AddToBlacklist 手动添加到黑名单
func (m *AdvancedRateLimitMiddleware) AddToBlacklist(ip, reason string, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.blacklist[ip] = &BlacklistRecord{
		IP:        ip,
		Reason:    reason,
		AddedAt:   time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	m.storageManager.LogWarning("IP手动加入黑名单", map[string]interface{}{
		"ip":         ip,
		"reason":     reason,
		"duration":   duration.String(),
		"expires_at": time.Now().Add(duration),
	})
}

// RemoveFromBlacklist 从黑名单中移除
func (m *AdvancedRateLimitMiddleware) RemoveFromBlacklist(ip string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.blacklist[ip]; exists {
		delete(m.blacklist, ip)
		m.storageManager.LogInfo("IP从黑名单中移除", map[string]interface{}{
			"ip": ip,
		})
	}
}

// GetBlacklist 获取黑名单
func (m *AdvancedRateLimitMiddleware) GetBlacklist() []*BlacklistRecord {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var blacklist []*BlacklistRecord
	for _, record := range m.blacklist {
		blacklist = append(blacklist, record)
	}
	return blacklist
}
