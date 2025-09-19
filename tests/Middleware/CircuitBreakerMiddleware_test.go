package Middleware

import (
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerMiddleware(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建存储管理器
	storageManager := &Storage.StorageManager{}

	// 创建熔断器中间件
	cbMiddleware := Middleware.NewCircuitBreakerMiddleware(storageManager)

	t.Run("正常请求应该通过", func(t *testing.T) {
		router := gin.New()
		router.Use(cbMiddleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("熔断器状态检查", func(t *testing.T) {
		// 创建熔断器
		config := &Middleware.CircuitBreakerConfig{
			MaxRequests:            2,
			Interval:               1 * time.Second,
			Timeout:                500 * time.Millisecond,
			MaxFailures:            2,
			FailureThreshold:       0.5,
			SuccessThreshold:       0.8,
			RequestVolumeThreshold: 1,
		}

		breaker := Middleware.NewCircuitBreaker(config)

		// 测试初始状态
		assert.Equal(t, Middleware.StateClosed, breaker.GetState())
		assert.True(t, breaker.AllowRequest())

		// 记录失败
		breaker.RecordResult(false, 100*time.Millisecond)
		breaker.RecordResult(false, 100*time.Millisecond)

		// 检查状态变化
		assert.Equal(t, Middleware.StateOpen, breaker.GetState())
		assert.False(t, breaker.AllowRequest())
	})

	t.Run("熔断器统计信息", func(t *testing.T) {
		config := &Middleware.CircuitBreakerConfig{
			MaxRequests:            5,
			Interval:               1 * time.Second,
			Timeout:                500 * time.Millisecond,
			MaxFailures:            3,
			FailureThreshold:       0.5,
			SuccessThreshold:       0.8,
			RequestVolumeThreshold: 1,
		}

		breaker := Middleware.NewCircuitBreaker(config)

		// 记录一些结果
		breaker.RecordResult(true, 100*time.Millisecond)
		breaker.RecordResult(false, 200*time.Millisecond)
		breaker.RecordResult(true, 150*time.Millisecond)

		stats := breaker.GetStats()
		assert.Equal(t, uint32(3), stats["requests"])
		assert.Equal(t, uint32(1), stats["failures"])
		assert.Equal(t, uint32(2), stats["successes"])
	})

	t.Run("熔断器重置", func(t *testing.T) {
		config := &Middleware.CircuitBreakerConfig{
			MaxRequests:            2,
			Interval:               1 * time.Second,
			Timeout:                500 * time.Millisecond,
			MaxFailures:            1,
			FailureThreshold:       0.5,
			SuccessThreshold:       0.8,
			RequestVolumeThreshold: 1,
		}

		breaker := Middleware.NewCircuitBreaker(config)

		// 触发熔断
		breaker.RecordResult(false, 100*time.Millisecond)
		assert.Equal(t, Middleware.StateOpen, breaker.GetState())

		// 重置熔断器
		breaker.Reset()
		assert.Equal(t, Middleware.StateClosed, breaker.GetState())
		assert.True(t, breaker.AllowRequest())
	})
}

func TestCircuitBreakerMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storageManager := &Storage.StorageManager{}
	cbMiddleware := Middleware.NewCircuitBreakerMiddleware(storageManager)

	t.Run("熔断器中间件集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(cbMiddleware.Handle())

		// 添加一个会失败的端点
		router.GET("/fail", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		})

		// 添加一个成功的端点
		router.GET("/success", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 测试失败端点
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// 测试成功端点
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/success", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCircuitBreakerMiddlewareConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storageManager := &Storage.StorageManager{}
	cbMiddleware := Middleware.NewCircuitBreakerMiddleware(storageManager)

	t.Run("并发请求测试", func(t *testing.T) {
		router := gin.New()
		router.Use(cbMiddleware.Handle())
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 并发发送请求
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)
				done <- true
			}()
		}

		// 等待所有请求完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 检查熔断器统计
		stats := cbMiddleware.GetCircuitBreakerStats()
		assert.NotNil(t, stats)
	})
}

func BenchmarkCircuitBreakerMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	storageManager := &Storage.StorageManager{}
	cbMiddleware := Middleware.NewCircuitBreakerMiddleware(storageManager)

	router := gin.New()
	router.Use(cbMiddleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		}
	})
}
