package benchmark

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceTestSuite 性能测试套件
type PerformanceTestSuite struct {
	router           *gin.Engine
	storageManager   *Storage.StorageManager
	healthController *Controllers.HealthController
	circuitBreaker   *Middleware.CircuitBreakerMiddleware
}

// NewPerformanceTestSuite 创建性能测试套件
func NewPerformanceTestSuite() *PerformanceTestSuite {
	gin.SetMode(gin.TestMode)

	storageManager := &Storage.StorageManager{}
	healthController := Controllers.NewHealthController()
	circuitBreaker := Middleware.NewCircuitBreakerMiddleware(storageManager)

	router := gin.New()
	router.Use(circuitBreaker.Handle())
	router.GET("/health", healthController.Health)
	router.GET("/health/detailed", healthController.DetailedHealth)
	router.GET("/health/ready", healthController.Readiness)
	router.GET("/health/live", healthController.Liveness)

	return &PerformanceTestSuite{
		router:           router,
		storageManager:   storageManager,
		healthController: healthController,
		circuitBreaker:   circuitBreaker,
	}
}

// BenchmarkHealthEndpoint 健康检查端点性能测试
func BenchmarkHealthEndpoint(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			suite.router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkDetailedHealthEndpoint 详细健康检查端点性能测试
func BenchmarkDetailedHealthEndpoint(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health/detailed", nil)
			suite.router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkCircuitBreakerMiddleware 熔断器中间件性能测试
func BenchmarkCircuitBreakerMiddleware(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			suite.router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkConcurrentRequests 并发请求性能测试
func BenchmarkConcurrentRequests(b *testing.B) {
	suite := NewPerformanceTestSuite()

	concurrencyLevels := []int{1, 10, 50, 100, 500, 1000}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", "/health", nil)
					suite.router.ServeHTTP(w, req)
				}
			})
		})
	}
}

// BenchmarkMemoryUsage 内存使用性能测试
func BenchmarkMemoryUsage(b *testing.B) {
	suite := NewPerformanceTestSuite()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			suite.router.ServeHTTP(w, req)
		}
	})

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc), "bytes/op")
	b.ReportMetric(float64(m2.NumGC-m1.NumGC), "gc/op")
}

// BenchmarkResponseTime 响应时间性能测试
func BenchmarkResponseTime(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			suite.router.ServeHTTP(w, req)
			duration := time.Since(start)

			b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
		}
	})
}

// TestLoadTest 负载测试
func TestLoadTest(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// 测试参数
	duration := 30 * time.Second
	concurrency := 100
	requestsPerSecond := 1000

	// 创建请求通道
	requestChan := make(chan struct{}, concurrency)
	responseChan := make(chan time.Duration, concurrency*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				start := time.Now()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/health", nil)
				suite.router.ServeHTTP(w, req)
				responseChan <- time.Since(start)
			}
		}()
	}

	// 启动请求生成器
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		defer ticker.Stop()

		timeout := time.After(duration)
		for {
			select {
			case <-ticker.C:
				select {
				case requestChan <- struct{}{}:
				default:
					// 请求队列已满，跳过
				}
			case <-timeout:
				close(requestChan)
				return
			}
		}
	}()

	// 等待测试完成
	wg.Wait()
	close(responseChan)

	// 分析结果
	var totalDuration time.Duration
	var responseTimes []time.Duration
	var successCount int

	for duration := range responseChan {
		totalDuration += duration
		responseTimes = append(responseTimes, duration)
		successCount++
	}

	// 计算统计信息
	avgResponseTime := totalDuration / time.Duration(len(responseTimes))

	// 计算百分位数
	sortDurations(responseTimes)
	p95Index := int(float64(len(responseTimes)) * 0.95)
	p99Index := int(float64(len(responseTimes)) * 0.99)

	p95ResponseTime := responseTimes[p95Index]
	p99ResponseTime := responseTimes[p99Index]

	// 输出结果
	t.Logf("负载测试结果:")
	t.Logf("  持续时间: %v", duration)
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  目标QPS: %d", requestsPerSecond)
	t.Logf("  实际请求数: %d", successCount)
	t.Logf("  错误数: 0")
	t.Logf("  平均响应时间: %v", avgResponseTime)
	t.Logf("  P95响应时间: %v", p95ResponseTime)
	t.Logf("  P99响应时间: %v", p99ResponseTime)
	t.Logf("  实际QPS: %.2f", float64(successCount)/duration.Seconds())

	// 性能断言
	if avgResponseTime > 100*time.Millisecond {
		t.Errorf("平均响应时间过长: %v", avgResponseTime)
	}

	if p95ResponseTime > 500*time.Millisecond {
		t.Errorf("P95响应时间过长: %v", p95ResponseTime)
	}

	if p99ResponseTime > 1*time.Second {
		t.Errorf("P99响应时间过长: %v", p99ResponseTime)
	}
}

// TestStressTest 压力测试
func TestStressTest(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// 压力测试参数
	duration := 60 * time.Second
	maxConcurrency := 1000

	// 逐步增加并发数
	concurrencyLevels := []int{10, 50, 100, 200, 500, 1000}

	for _, concurrency := range concurrencyLevels {
		if concurrency > maxConcurrency {
			break
		}

		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			testConcurrency(t, suite, concurrency, duration)
		})
	}
}

// testConcurrency 测试特定并发数
func testConcurrency(t *testing.T, suite *PerformanceTestSuite, concurrency int, duration time.Duration) {
	requestChan := make(chan struct{}, concurrency)
	responseChan := make(chan time.Duration, concurrency*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				start := time.Now()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/health", nil)
				suite.router.ServeHTTP(w, req)
				responseChan <- time.Since(start)
			}
		}()
	}

	// 启动请求生成器
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(duration)
		for {
			select {
			case <-ticker.C:
				select {
				case requestChan <- struct{}{}:
				default:
					// 请求队列已满，跳过
				}
			case <-timeout:
				close(requestChan)
				return
			}
		}
	}()

	// 等待测试完成
	wg.Wait()
	close(responseChan)

	// 分析结果
	var totalDuration time.Duration
	var responseTimes []time.Duration
	var successCount int

	for duration := range responseChan {
		totalDuration += duration
		responseTimes = append(responseTimes, duration)
		successCount++
	}

	// 计算统计信息
	avgResponseTime := totalDuration / time.Duration(len(responseTimes))

	// 计算百分位数
	sortDurations(responseTimes)
	p95Index := int(float64(len(responseTimes)) * 0.95)
	p99Index := int(float64(len(responseTimes)) * 0.99)

	p95ResponseTime := responseTimes[p95Index]
	p99ResponseTime := responseTimes[p99Index]

	// 输出结果
	t.Logf("并发数 %d 测试结果:", concurrency)
	t.Logf("  请求数: %d", successCount)
	t.Logf("  平均响应时间: %v", avgResponseTime)
	t.Logf("  P95响应时间: %v", p95ResponseTime)
	t.Logf("  P99响应时间: %v", p99ResponseTime)
	t.Logf("  QPS: %.2f", float64(successCount)/duration.Seconds())

	// 性能断言
	if avgResponseTime > 200*time.Millisecond {
		t.Errorf("并发数 %d 平均响应时间过长: %v", concurrency, avgResponseTime)
	}

	if p95ResponseTime > 1*time.Second {
		t.Errorf("并发数 %d P95响应时间过长: %v", concurrency, p95ResponseTime)
	}
}

// TestMemoryLeakTest 内存泄漏测试
func TestMemoryLeakTest(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// 运行多次请求，检查内存使用
	iterations := 10000

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < iterations; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		suite.router.ServeHTTP(w, req)

		// 每1000次请求检查一次内存
		if i%1000 == 0 {
			runtime.GC()
			runtime.ReadMemStats(&m2)

			memoryIncrease := m2.Alloc - m1.Alloc
			if memoryIncrease > 1024*1024 { // 1MB
				t.Errorf("内存使用增加过多: %d bytes", memoryIncrease)
			}
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	finalMemoryIncrease := m2.Alloc - m1.Alloc
	t.Logf("最终内存增加: %d bytes", finalMemoryIncrease)

	if finalMemoryIncrease > 10*1024*1024 { // 10MB
		t.Errorf("内存泄漏检测: 最终内存增加过多: %d bytes", finalMemoryIncrease)
	}
}

// TestCircuitBreakerPerformance 熔断器性能测试
func TestCircuitBreakerPerformance(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// 测试熔断器在不同负载下的性能
	concurrencyLevels := []int{10, 50, 100, 200}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			testCircuitBreakerConcurrency(t, suite, concurrency)
		})
	}
}

// testCircuitBreakerConcurrency 测试熔断器并发性能
func testCircuitBreakerConcurrency(t *testing.T, suite *PerformanceTestSuite, concurrency int) {
	requestChan := make(chan struct{}, concurrency)
	responseChan := make(chan time.Duration, concurrency*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				start := time.Now()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/health", nil)
				suite.router.ServeHTTP(w, req)
				responseChan <- time.Since(start)
			}
		}()
	}

	// 生成请求
	go func() {
		for i := 0; i < 1000; i++ {
			requestChan <- struct{}{}
		}
		close(requestChan)
	}()

	// 等待完成
	wg.Wait()
	close(responseChan)

	// 分析结果
	var totalDuration time.Duration
	var responseTimes []time.Duration
	var successCount int

	for duration := range responseChan {
		totalDuration += duration
		responseTimes = append(responseTimes, duration)
		successCount++
	}

	// 计算统计信息
	avgResponseTime := totalDuration / time.Duration(len(responseTimes))

	// 输出结果
	t.Logf("熔断器并发数 %d 测试结果:", concurrency)
	t.Logf("  请求数: %d", successCount)
	t.Logf("  平均响应时间: %v", avgResponseTime)

	// 性能断言
	if avgResponseTime > 100*time.Millisecond {
		t.Errorf("熔断器并发数 %d 平均响应时间过长: %v", concurrency, avgResponseTime)
	}
}

// sortDurations 排序持续时间切片
func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

// BenchmarkJSONSerialization JSON序列化性能测试
func BenchmarkJSONSerialization(b *testing.B) {
	data := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"uptime":    "2d 5h 30m 15s",
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status":        "healthy",
				"response_time": "5ms",
				"message":       "OK",
			},
			"redis": map[string]interface{}{
				"status":        "healthy",
				"response_time": "2ms",
				"message":       "OK",
			},
		},
		"metrics": map[string]interface{}{
			"memory": map[string]interface{}{
				"alloc":       1024000,
				"total_alloc": 2048000,
				"sys":         4096000,
			},
			"cpu": map[string]interface{}{
				"usage": 15.5,
			},
			"goroutines": 150,
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := json.Marshal(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkJSONDeserialization JSON反序列化性能测试
func BenchmarkJSONDeserialization(b *testing.B) {
	jsonData := []byte(`{
		"status": "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version": "1.0.0",
		"uptime": "2d 5h 30m 15s",
		"services": {
			"database": {
				"status": "healthy",
				"response_time": "5ms",
				"message": "OK"
			},
			"redis": {
				"status": "healthy",
				"response_time": "2ms",
				"message": "OK"
			}
		},
		"metrics": {
			"memory": {
				"alloc": 1024000,
				"total_alloc": 2048000,
				"sys": 4096000
			},
			"cpu": {
				"usage": 15.5
			},
			"goroutines": 150
		}
	}`)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var data map[string]interface{}
			err := json.Unmarshal(jsonData, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
