package Middleware

import (
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitEntry 速率限制条目
type RateLimitEntry struct {
	Count     int
	LastReset time.Time
}

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	limits         map[string]*RateLimitEntry
	mutex          sync.RWMutex
}

// NewRateLimitMiddleware 创建速率限制中间件
// 功能说明：
// 1. 初始化速率限制中间件实例
// 2. 创建内存存储的限制记录映射
// 3. 启动自动清理协程防止内存泄漏
// 4. 返回配置好的中间件实例
func NewRateLimitMiddleware(storageManager *Storage.StorageManager) *RateLimitMiddleware {
	middleware := &RateLimitMiddleware{
		storageManager: storageManager,
		limits:         make(map[string]*RateLimitEntry),
	}
	
	// 启动自动清理协程
	go middleware.Cleanup()
	
	return middleware
}

// Handle 处理速率限制
// 功能说明：
// 1. 基于IP地址进行速率限制
// 2. 支持自定义限制次数和时间窗口
// 3. 自动重置计数器
// 4. 记录超限请求日志
func (m *RateLimitMiddleware) Handle(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		key := clientIP + ":" + c.Request.URL.Path
		
		m.mutex.Lock()
		entry, exists := m.limits[key]
		if !exists {
			entry = &RateLimitEntry{
				Count:     0,
				LastReset: time.Now(),
			}
			m.limits[key] = entry
		}
		
		// 检查是否需要重置计数器
		if time.Since(entry.LastReset) > window {
			entry.Count = 0
			entry.LastReset = time.Now()
		}
		
		// 检查是否超过限制
		if entry.Count >= maxRequests {
			// 记录超限日志
			m.storageManager.LogWarning("Rate limit exceeded", map[string]interface{}{
				"client_ip":    clientIP,
				"path":         c.Request.URL.Path,
				"method":       c.Request.Method,
				"limit":        maxRequests,
				"window":       window.String(),
				"user_agent":   c.Request.UserAgent(),
			})
			
			m.mutex.Unlock()
			
			// 返回429错误
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests, please try again later",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}
		
		// 增加计数器
		entry.Count++
		m.mutex.Unlock()
		
		// 添加响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-entry.Count))
		c.Header("X-RateLimit-Reset", entry.LastReset.Add(window).Format(time.RFC3339))
		
		c.Next()
	}
}

// Cleanup 清理过期的限制记录
// 功能说明：
// 1. 定期清理过期的速率限制记录
// 2. 防止内存泄漏
// 3. 自动运行清理任务
func (m *RateLimitMiddleware) Cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		
		for key, entry := range m.limits {
			// 清理超过1小时的记录
			if now.Sub(entry.LastReset) > time.Hour {
				delete(m.limits, key)
			}
		}
		
		m.mutex.Unlock()
	}
}
