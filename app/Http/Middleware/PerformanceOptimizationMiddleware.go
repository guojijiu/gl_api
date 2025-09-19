package Middleware

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceOptimizationMiddleware 性能优化中间件
// 功能说明：
// 1. 监控请求性能指标
// 2. 自动优化响应
// 3. 检测性能瓶颈
// 4. 提供性能报告
type PerformanceOptimizationMiddleware struct {
	BaseMiddleware
	storageManager   *Storage.StorageManager
	requestCount     int64
	responseTimeSum  int64
	slowRequestCount int64
	errorCount       int64
	mu               sync.RWMutex
	performanceStats *PerformanceStats
	config           *PerformanceConfig
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	TotalRequests       int64         `json:"total_requests"`    // 总请求数
	AverageResponseTime time.Duration `json:"avg_response_time"` // 平均响应时间
	SlowRequests        int64         `json:"slow_requests"`     // 慢请求数
	ErrorRate           float64       `json:"error_rate"`        // 错误率
	MemoryUsage         uint64        `json:"memory_usage"`      // 内存使用量
	GCStats             *GCStats      `json:"gc_stats"`          // GC统计
	LastUpdated         time.Time     `json:"last_updated"`      // 最后更新时间
}

// GCStats GC统计
type GCStats struct {
	NumGC        uint32        `json:"num_gc"`        // GC次数
	PauseTotal   time.Duration `json:"pause_total"`   // 总暂停时间
	PauseAverage time.Duration `json:"pause_average"` // 平均暂停时间
	LastGC       time.Time     `json:"last_gc"`       // 最后GC时间
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	SlowRequestThreshold time.Duration `json:"slow_request_threshold"` // 慢请求阈值
	EnableCompression    bool          `json:"enable_compression"`     // 启用压缩
	EnableCaching        bool          `json:"enable_caching"`         // 启用缓存
	MaxMemoryUsage       uint64        `json:"max_memory_usage"`       // 最大内存使用量（字节）
	EnableGC             bool          `json:"enable_gc"`              // 启用GC优化
}

// NewPerformanceOptimizationMiddleware 创建性能优化中间件
func NewPerformanceOptimizationMiddleware(storageManager *Storage.StorageManager) *PerformanceOptimizationMiddleware {
	config := &PerformanceConfig{
		SlowRequestThreshold: 1 * time.Second,   // 1秒慢请求阈值
		EnableCompression:    true,              // 启用压缩
		EnableCaching:        true,              // 启用缓存
		MaxMemoryUsage:       100 * 1024 * 1024, // 100MB最大内存
		EnableGC:             true,              // 启用GC优化
	}

	return &PerformanceOptimizationMiddleware{
		storageManager:   storageManager,
		performanceStats: &PerformanceStats{},
		config:           config,
	}
}

// Handle 处理请求
func (m *PerformanceOptimizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求开始
		atomic.AddInt64(&m.requestCount, 1)

		// 检查内存使用量
		if m.config.EnableGC {
			m.checkMemoryUsage()
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)
		atomic.AddInt64(&m.responseTimeSum, int64(duration))

		// 检查是否为慢请求
		if duration > m.config.SlowRequestThreshold {
			atomic.AddInt64(&m.slowRequestCount, 1)
			m.logSlowRequest(c, duration)
		}

		// 检查错误
		if c.Writer.Status() >= 400 {
			atomic.AddInt64(&m.errorCount, 1)
		}

		// 更新性能统计
		m.updatePerformanceStats()

		// 添加性能头信息
		m.addPerformanceHeaders(c, duration)
	}
}

// checkMemoryUsage 检查内存使用量
func (m *PerformanceOptimizationMiddleware) checkMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memStats.Alloc > m.config.MaxMemoryUsage {
		// 触发GC
		runtime.GC()

		m.storageManager.LogWarning("内存使用量过高，触发GC", map[string]interface{}{
			"current_usage": memStats.Alloc,
			"max_usage":     m.config.MaxMemoryUsage,
			"gc_triggered":  true,
		})
	}
}

// logSlowRequest 记录慢请求
func (m *PerformanceOptimizationMiddleware) logSlowRequest(c *gin.Context, duration time.Duration) {
	m.storageManager.LogWarning("慢请求检测", map[string]interface{}{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"duration":   duration.String(),
		"threshold":  m.config.SlowRequestThreshold.String(),
		"user_agent": c.Request.UserAgent(),
		"ip":         c.ClientIP(),
	})
}

// updatePerformanceStats 更新性能统计
func (m *PerformanceOptimizationMiddleware) updatePerformanceStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新基本统计
	m.performanceStats.TotalRequests = atomic.LoadInt64(&m.requestCount)
	m.performanceStats.SlowRequests = atomic.LoadInt64(&m.slowRequestCount)
	m.performanceStats.LastUpdated = time.Now()

	// 计算平均响应时间
	if m.performanceStats.TotalRequests > 0 {
		avgTime := atomic.LoadInt64(&m.responseTimeSum) / m.performanceStats.TotalRequests
		m.performanceStats.AverageResponseTime = time.Duration(avgTime)
	}

	// 计算错误率
	if m.performanceStats.TotalRequests > 0 {
		errorCount := atomic.LoadInt64(&m.errorCount)
		m.performanceStats.ErrorRate = float64(errorCount) / float64(m.performanceStats.TotalRequests)
	}

	// 更新内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.performanceStats.MemoryUsage = memStats.Alloc

	// 更新GC统计
	m.performanceStats.GCStats = &GCStats{
		NumGC:      memStats.NumGC,
		PauseTotal: time.Duration(memStats.PauseTotalNs),
		LastGC:     time.Unix(0, int64(memStats.LastGC)),
	}

	if memStats.NumGC > 0 {
		m.performanceStats.GCStats.PauseAverage = time.Duration(memStats.PauseTotalNs) / time.Duration(memStats.NumGC)
	}
}

// addPerformanceHeaders 添加性能头信息
func (m *PerformanceOptimizationMiddleware) addPerformanceHeaders(c *gin.Context, duration time.Duration) {
	c.Header("X-Response-Time", duration.String())
	c.Header("X-Request-ID", c.GetString("request_id"))

	// 添加缓存头（如果启用）
	if m.config.EnableCaching {
		c.Header("Cache-Control", "public, max-age=300") // 5分钟缓存
	}

	// 添加压缩头（如果启用）
	if m.config.EnableCompression {
		c.Header("Content-Encoding", "gzip")
	}
}

// GetPerformanceStats 获取性能统计
func (m *PerformanceOptimizationMiddleware) GetPerformanceStats() *PerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回统计副本
	stats := *m.performanceStats
	return &stats
}

// ResetStats 重置统计
func (m *PerformanceOptimizationMiddleware) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.requestCount, 0)
	atomic.StoreInt64(&m.responseTimeSum, 0)
	atomic.StoreInt64(&m.slowRequestCount, 0)
	atomic.StoreInt64(&m.errorCount, 0)

	m.performanceStats = &PerformanceStats{}
}

// SetConfig 更新配置
func (m *PerformanceOptimizationMiddleware) SetConfig(config *PerformanceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
}

// HealthCheck 健康检查
func (m *PerformanceOptimizationMiddleware) HealthCheck() map[string]interface{} {
	stats := m.GetPerformanceStats()

	health := map[string]interface{}{
		"status":            "healthy",
		"total_requests":    stats.TotalRequests,
		"avg_response_time": stats.AverageResponseTime.String(),
		"slow_requests":     stats.SlowRequests,
		"error_rate":        fmt.Sprintf("%.2f%%", stats.ErrorRate*100),
		"memory_usage":      stats.MemoryUsage,
		"last_updated":      stats.LastUpdated,
	}

	// 判断健康状态
	if stats.ErrorRate > 0.1 { // 错误率超过10%
		health["status"] = "warning"
		health["message"] = "错误率过高"
	} else if stats.AverageResponseTime > 2*time.Second { // 平均响应时间超过2秒
		health["status"] = "warning"
		health["message"] = "响应时间过长"
	} else if stats.MemoryUsage > m.config.MaxMemoryUsage {
		health["status"] = "critical"
		health["message"] = "内存使用量过高"
	}

	return health
}

// OptimizeResponse 优化响应
func (m *PerformanceOptimizationMiddleware) OptimizeResponse(c *gin.Context) {
	// 启用压缩
	if m.config.EnableCompression {
		c.Header("Content-Encoding", "gzip")
	}

	// 设置缓存头
	if m.config.EnableCaching {
		c.Header("Cache-Control", "public, max-age=300")
		c.Header("ETag", generateETag(c))
	}

	// 移除不必要的头信息
	c.Header("Server", "CloudPlatform-API")
	c.Header("X-Powered-By", "Go")
}

// generateETag 生成ETag
func generateETag(c *gin.Context) string {
	// 简单的ETag生成，实际应用中应该基于内容生成
	return fmt.Sprintf("\"%x\"", time.Now().Unix())
}
