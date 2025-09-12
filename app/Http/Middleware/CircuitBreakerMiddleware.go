package Middleware

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	MaxRequests            uint32        `json:"max_requests"`             // 半开状态最大请求数
	Interval               time.Duration `json:"interval"`                 // 熔断器重置间隔
	Timeout                time.Duration `json:"timeout"`                  // 熔断器超时时间
	MaxFailures            uint32        `json:"max_failures"`             // 最大失败次数
	FailureThreshold       float64       `json:"failure_threshold"`        // 失败率阈值
	SuccessThreshold       float64       `json:"success_threshold"`        // 成功率阈值
	RequestVolumeThreshold uint32        `json:"request_volume_threshold"` // 请求量阈值
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	state         CircuitBreakerState
	failures      uint32
	successes     uint32
	requests      uint32
	lastFailTime  time.Time
	lastResetTime time.Time
	mutex         sync.RWMutex
}

// CircuitBreakerMiddleware 熔断器中间件
type CircuitBreakerMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	breakers       map[string]*CircuitBreaker
	mutex          sync.RWMutex
}

// NewCircuitBreakerMiddleware 创建熔断器中间件
func NewCircuitBreakerMiddleware(storageManager *Storage.StorageManager) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		storageManager: storageManager,
		breakers:       make(map[string]*CircuitBreaker),
	}
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			MaxRequests:            5,
			Interval:               60 * time.Second,
			Timeout:                30 * time.Second,
			MaxFailures:            5,
			FailureThreshold:       0.5,
			SuccessThreshold:       0.8,
			RequestVolumeThreshold: 10,
		}
	}

	return &CircuitBreaker{
		config:        config,
		state:         StateClosed,
		lastResetTime: time.Now(),
	}
}

// Handle 熔断器中间件处理函数
func (m *CircuitBreakerMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取熔断器键（基于路径和方法）
		key := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

		// 获取或创建熔断器
		breaker := m.getOrCreateBreaker(key)

		// 检查熔断器状态
		if !breaker.AllowRequest() {
			m.storageManager.LogWarning("熔断器阻止请求", map[string]interface{}{
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
				"state":  breaker.GetState(),
			})

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "服务暂时不可用，请稍后重试",
				"code":    "CIRCUIT_BREAKER_OPEN",
			})
			c.Abort()
			return
		}

		// 记录请求开始
		start := time.Now()

		// 执行请求
		c.Next()

		// 记录请求结果
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 判断请求是否成功
		success := statusCode < 500

		// 更新熔断器状态
		breaker.RecordResult(success, duration)

		// 记录熔断器状态变化
		if breaker.GetState() == StateOpen {
			m.storageManager.LogWarning("熔断器状态变为开启", map[string]interface{}{
				"path":     c.Request.URL.Path,
				"method":   c.Request.Method,
				"failures": breaker.GetFailures(),
				"requests": breaker.GetRequests(),
			})
		}
	}
}

// getOrCreateBreaker 获取或创建熔断器
func (m *CircuitBreakerMiddleware) getOrCreateBreaker(key string) *CircuitBreaker {
	m.mutex.RLock()
	breaker, exists := m.breakers[key]
	m.mutex.RUnlock()

	if !exists {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		// 双重检查
		if breaker, exists = m.breakers[key]; !exists {
			breaker = NewCircuitBreaker(nil)
			m.breakers[key] = breaker
		}
	}

	return breaker
}

// AllowRequest 检查是否允许请求
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	// 检查是否需要重置熔断器
	if cb.state == StateOpen && now.Sub(cb.lastFailTime) > cb.config.Interval {
		cb.state = StateHalfOpen
		cb.failures = 0
		cb.successes = 0
		cb.requests = 0
		cb.lastResetTime = now
	}

	// 检查熔断器状态
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return false
	case StateHalfOpen:
		return cb.requests < cb.config.MaxRequests
	default:
		return true
	}
}

// RecordResult 记录请求结果
func (cb *CircuitBreaker) RecordResult(success bool, duration time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.requests++

	if success {
		cb.successes++
	} else {
		cb.failures++
		cb.lastFailTime = time.Now()
	}

	// 检查是否需要改变状态
	cb.checkState()
}

// checkState 检查并更新熔断器状态
func (cb *CircuitBreaker) checkState() {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		// 检查是否需要开启熔断器
		if cb.requests >= cb.config.RequestVolumeThreshold {
			failureRate := float64(cb.failures) / float64(cb.requests)
			if failureRate >= cb.config.FailureThreshold || cb.failures >= cb.config.MaxFailures {
				cb.state = StateOpen
				cb.lastFailTime = now
			}
		}

	case StateHalfOpen:
		// 检查是否需要关闭或重新开启熔断器
		if cb.requests >= cb.config.MaxRequests {
			successRate := float64(cb.successes) / float64(cb.requests)
			if successRate >= cb.config.SuccessThreshold {
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
				cb.requests = 0
				cb.lastResetTime = now
			} else {
				cb.state = StateOpen
				cb.lastFailTime = now
			}
		}
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailures 获取失败次数
func (cb *CircuitBreaker) GetFailures() uint32 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failures
}

// GetRequests 获取请求次数
func (cb *CircuitBreaker) GetRequests() uint32 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.requests
}

// GetStats 获取熔断器统计信息
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	var stateStr string
	switch cb.state {
	case StateClosed:
		stateStr = "closed"
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	}

	return map[string]interface{}{
		"state":           stateStr,
		"failures":        cb.failures,
		"successes":       cb.successes,
		"requests":        cb.requests,
		"last_fail_time":  cb.lastFailTime,
		"last_reset_time": cb.lastResetTime,
	}
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
	cb.lastResetTime = time.Now()
}

// GetCircuitBreakerStats 获取所有熔断器统计信息
func (m *CircuitBreakerMiddleware) GetCircuitBreakerStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	for key, breaker := range m.breakers {
		stats[key] = breaker.GetStats()
	}

	return stats
}

// ResetCircuitBreaker 重置指定熔断器
func (m *CircuitBreakerMiddleware) ResetCircuitBreaker(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	breaker, exists := m.breakers[key]
	if !exists {
		return fmt.Errorf("熔断器不存在: %s", key)
	}

	breaker.Reset()
	return nil
}

// ResetAllCircuitBreakers 重置所有熔断器
func (m *CircuitBreakerMiddleware) ResetAllCircuitBreakers() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, breaker := range m.breakers {
		breaker.Reset()
	}
}
