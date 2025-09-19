package Middleware

import (
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestStatsMiddleware 请求统计中间件
type RequestStatsMiddleware struct {
	BaseMiddleware
	storageManager    *Storage.StorageManager
	monitoringService *Services.OptimizedMonitoringService
	requestCount      int64
	errorCount        int64
	totalResponseTime int64
	responseTimeCount int64
	activeUsers       int64
	mu                sync.RWMutex
	userSessions      map[string]time.Time
	lastCleanup       time.Time
}

// NewRequestStatsMiddleware 创建请求统计中间件
func NewRequestStatsMiddleware(storageManager *Storage.StorageManager, monitoringService *Services.OptimizedMonitoringService) *RequestStatsMiddleware {
	return &RequestStatsMiddleware{
		storageManager:    storageManager,
		monitoringService: monitoringService,
		userSessions:      make(map[string]time.Time),
		lastCleanup:       time.Now(),
	}
}

// Handle 处理请求统计
func (m *RequestStatsMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 增加请求计数
		atomic.AddInt64(&m.requestCount, 1)

		// 记录活跃用户
		if userID := c.GetString("user_id"); userID != "" {
			m.recordActiveUser(userID)
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		responseTime := time.Since(startTime)
		atomic.AddInt64(&m.totalResponseTime, responseTime.Nanoseconds())
		atomic.AddInt64(&m.responseTimeCount, 1)

		// 记录错误
		if c.Writer.Status() >= 400 {
			atomic.AddInt64(&m.errorCount, 1)
		}

		// 更新监控服务缓存
		m.updateMonitoringCache()

		// 定期清理过期用户会话
		if time.Since(m.lastCleanup) > 5*time.Minute {
			m.cleanupExpiredSessions()
		}
	}
}

// recordActiveUser 记录活跃用户
func (m *RequestStatsMiddleware) recordActiveUser(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.userSessions[userID] = time.Now()
	atomic.StoreInt64(&m.activeUsers, int64(len(m.userSessions)))
}

// cleanupExpiredSessions 清理过期用户会话
func (m *RequestStatsMiddleware) cleanupExpiredSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiredUsers := make([]string, 0)

	for userID, lastSeen := range m.userSessions {
		if now.Sub(lastSeen) > 30*time.Minute { // 30分钟无活动视为过期
			expiredUsers = append(expiredUsers, userID)
		}
	}

	for _, userID := range expiredUsers {
		delete(m.userSessions, userID)
	}

	atomic.StoreInt64(&m.activeUsers, int64(len(m.userSessions)))
	m.lastCleanup = now
}

// updateMonitoringCache 更新监控服务缓存
func (m *RequestStatsMiddleware) updateMonitoringCache() {
	// 更新请求计数
	m.monitoringService.AddMetric("request_count", atomic.LoadInt64(&m.requestCount), map[string]string{
		"type": "counter",
	})

	// 更新错误计数
	m.monitoringService.AddMetric("error_count", atomic.LoadInt64(&m.errorCount), map[string]string{
		"type": "counter",
	})

	// 更新活跃用户数
	m.monitoringService.AddMetric("active_users", atomic.LoadInt64(&m.activeUsers), map[string]string{
		"type": "gauge",
	})

	// 计算平均响应时间
	if count := atomic.LoadInt64(&m.responseTimeCount); count > 0 {
		totalTime := atomic.LoadInt64(&m.totalResponseTime)
		avgTime := time.Duration(totalTime / count)
		m.monitoringService.AddMetric("avg_response_time", avgTime, map[string]string{
			"type": "gauge",
		})
	}
}

// GetStats 获取统计信息
func (m *RequestStatsMiddleware) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"request_count":       atomic.LoadInt64(&m.requestCount),
		"error_count":         atomic.LoadInt64(&m.errorCount),
		"active_users":        atomic.LoadInt64(&m.activeUsers),
		"response_time_count": atomic.LoadInt64(&m.responseTimeCount),
	}

	// 计算平均响应时间
	if count := atomic.LoadInt64(&m.responseTimeCount); count > 0 {
		totalTime := atomic.LoadInt64(&m.totalResponseTime)
		avgTime := time.Duration(totalTime / count)
		stats["avg_response_time"] = avgTime.String()
		stats["avg_response_time_ms"] = avgTime.Milliseconds()
	}

	return stats
}

// ResetStats 重置统计信息
func (m *RequestStatsMiddleware) ResetStats() {
	atomic.StoreInt64(&m.requestCount, 0)
	atomic.StoreInt64(&m.errorCount, 0)
	atomic.StoreInt64(&m.totalResponseTime, 0)
	atomic.StoreInt64(&m.responseTimeCount, 0)

	m.mu.Lock()
	m.userSessions = make(map[string]time.Time)
	m.mu.Unlock()

	atomic.StoreInt64(&m.activeUsers, 0)
}
