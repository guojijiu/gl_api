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
//
// 功能说明：
// 1. 基于IP地址和路径进行速率限制
// 2. 支持自定义限制次数和时间窗口
// 3. 自动重置计数器（滑动窗口）
// 4. 记录超限请求日志
// 5. 返回标准的速率限制响应头
//
// 限流策略：
// - 基于IP地址和请求路径的组合键
// - 滑动窗口：时间窗口到期后自动重置计数器
// - 固定窗口：在时间窗口内限制请求次数
//
// 限流键（Key）：
// - 格式："{IP}:{Path}"
// - 例如："192.168.1.1:/api/users"
// - 不同路径的请求分别计数
//
// 响应头：
// - X-RateLimit-Limit: 限制次数
// - X-RateLimit-Remaining: 剩余次数
// - X-RateLimit-Reset: 重置时间（RFC3339格式）
//
// 并发安全：
// - 使用互斥锁保护共享数据结构
// - 读写操作都需要加锁
//
// 使用场景：
// - 防止API滥用
// - 保护服务器资源
// - 防止DDoS攻击
//
// 注意事项：
// - 限流是基于内存的，服务重启后重置
// - 对于分布式系统，需要使用Redis等共享存储
// - 超限时返回429状态码，符合HTTP标准
func (m *RateLimitMiddleware) Handle(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP地址
		// 用于标识不同的客户端
		clientIP := c.ClientIP()
		
		// 生成限流键：IP + 路径
		// 不同路径的请求分别计数，更精细的限流控制
		key := clientIP + ":" + c.Request.URL.Path
		
		// 加锁保护，避免并发访问导致计数错误
		m.mutex.Lock()
		
		// 获取或创建限流记录
		entry, exists := m.limits[key]
		if !exists {
			// 首次访问，创建新的限流记录
			entry = &RateLimitEntry{
				Count:     0,        // 请求计数
				LastReset: time.Now(), // 最后重置时间
			}
			m.limits[key] = entry
		}
		
		// 检查是否需要重置计数器（滑动窗口）
		// 如果距离上次重置时间超过窗口大小，重置计数器
		if time.Since(entry.LastReset) > window {
			entry.Count = 0
			entry.LastReset = time.Now()
		}
		
		// 检查是否超过限制
		if entry.Count >= maxRequests {
			// 超过限制，记录超限日志
			// 用于安全审计和问题排查
			m.storageManager.LogWarning("Rate limit exceeded", map[string]interface{}{
				"client_ip":    clientIP,
				"path":         c.Request.URL.Path,
				"method":       c.Request.Method,
				"limit":        maxRequests,
				"window":       window.String(),
				"user_agent":   c.Request.UserAgent(),
			})
			
			// 释放锁
			m.mutex.Unlock()
			
			// 返回429 Too Many Requests错误
			// 符合HTTP标准，表示请求频率过高
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests, please try again later",
				"retry_after": window.Seconds(), // 建议的重试时间（秒）
			})
			c.Abort() // 停止后续处理
			return
		}
		
		// 未超过限制，增加计数器
		entry.Count++
		
		// 释放锁
		m.mutex.Unlock()
		
		// 添加标准的速率限制响应头
		// 客户端可以根据这些头信息了解限流状态
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))                    // 限制次数
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-entry.Count))    // 剩余次数
		c.Header("X-RateLimit-Reset", entry.LastReset.Add(window).Format(time.RFC3339)) // 重置时间
		
		// 通过限流检查，继续执行后续中间件和处理器
		c.Next()
	}
}

// Cleanup 清理过期的限制记录
//
// 功能说明：
// 1. 定期清理过期的速率限制记录
// 2. 防止内存泄漏（长期不访问的IP记录会占用内存）
// 3. 在独立的goroutine中运行，不阻塞主流程
// 4. 自动运行清理任务，无需手动调用
//
// 清理策略：
// - 每小时执行一次清理
// - 清理超过1小时未重置的记录
// - 释放不再使用的内存空间
//
// 实现原理：
// - 使用time.Ticker定时触发清理
// - 遍历所有限流记录，删除过期的记录
// - 使用互斥锁保护清理操作
//
// 内存管理：
// - 防止长期不访问的IP记录占用内存
// - 清理后内存使用量会减少
// - 如果IP再次访问，会重新创建记录
//
// 注意事项：
// - defer ticker.Stop()确保ticker资源被释放
// - 清理操作需要加锁，避免并发问题
// - 清理间隔可以根据实际情况调整
func (m *RateLimitMiddleware) Cleanup() {
	// 创建定时器，每小时触发一次清理
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop() // 确保ticker资源被释放
	
	// 无限循环，持续监听定时器
	for range ticker.C {
		// 加锁保护，避免并发清理导致数据不一致
		m.mutex.Lock()
		now := time.Now()
		
		// 遍历所有限流记录
		for key, entry := range m.limits {
			// 清理超过1小时未重置的记录
			// 这些记录通常来自长期不访问的IP
			if now.Sub(entry.LastReset) > time.Hour {
				delete(m.limits, key)
			}
		}
		
		// 释放锁
		m.mutex.Unlock()
	}
}
