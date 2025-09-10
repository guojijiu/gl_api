package Middleware

import (
	"cloud-platform-api/app/Storage"
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMiddleware 性能优化中间件
type PerformanceMiddleware struct {
	StorageManager *Storage.StorageManager
}

// NewPerformanceMiddleware 创建性能优化中间件
func NewPerformanceMiddleware(storageManager *Storage.StorageManager) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		StorageManager: storageManager,
	}
}

// Handle 处理性能优化
// 功能说明：
// 1. 设置性能相关的响应头
// 2. 计算和记录请求处理时间
// 3. 监控慢查询并记录日志
// 4. 提供性能优化建议
func (m *PerformanceMiddleware) Handle() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置响应头
		c.Header("X-Response-Time", "")
		c.Header("X-Cache-Control", "public, max-age=300")

		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())

		// 记录慢查询
		if duration > 1*time.Second {
			m.logSlowQuery(c, duration)
		}

		// 记录性能日志
		m.StorageManager.LogInfo("请求性能统计", map[string]interface{}{
			"category":  "requests",
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"duration":  duration.String(),
			"status":    c.Writer.Status(),
			"client_ip": c.ClientIP(),
		})
	})
}

// Timeout 超时控制中间件
func (m *PerformanceMiddleware) Timeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()

		// 检查是否超时
		if ctx.Err() == context.DeadlineExceeded {
			c.AbortWithStatusJSON(408, gin.H{
				"error":   "请求超时",
				"message": "请求处理时间超过限制",
			})
			return
		}
	})
}

// Cache 缓存中间件
func (m *PerformanceMiddleware) Cache(duration time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 设置缓存头
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(duration.Seconds())))
		c.Header("Expires", time.Now().Add(duration).Format(time.RFC1123))

		c.Next()
	})
}

// RateLimit 速率限制中间件
func (m *PerformanceMiddleware) RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	// 使用内存存储限制器
	limiter := make(map[string][]time.Time)

	return gin.HandlerFunc(func(c *gin.Context) {
		key := m.getClientKey(c)
		now := time.Now()

		// 清理过期的记录
		if times, exists := limiter[key]; exists {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < window {
					valid = append(valid, t)
				}
			}
			limiter[key] = valid
		}

		// 检查是否超过限制
		if len(limiter[key]) >= limit {
			c.AbortWithStatusJSON(429, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后再试",
			})
			return
		}

		// 添加当前请求时间
		limiter[key] = append(limiter[key], now)

		c.Next()
	})
}

// getClientKey 获取客户端标识
func (m *PerformanceMiddleware) getClientKey(c *gin.Context) string {
	// 优先使用用户ID
	if userID := c.GetString("user_id"); userID != "" {
		return "user:" + userID
	}

	// 否则使用IP地址
	return "ip:" + c.ClientIP()
}

// logSlowQuery 记录慢查询
func (m *PerformanceMiddleware) logSlowQuery(c *gin.Context, duration time.Duration) {
	// 这里应该实现慢查询日志记录
	fmt.Printf("慢查询: %s %s - 耗时: %v\n", c.Request.Method, c.Request.URL.Path, duration)
}
