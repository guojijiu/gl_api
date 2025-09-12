package Controllers

import (
	"cloud-platform-api/app/Http/Controllers"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建健康检查控制器
	hc := Controllers.NewHealthController()

	t.Run("创建健康检查控制器", func(t *testing.T) {
		assert.NotNil(t, hc)
		assert.NotNil(t, hc.GetCustomChecks())
	})

	t.Run("添加自定义健康检查", func(t *testing.T) {
		hc.AddCustomCheck("test_check", func() error {
			return nil
		})

		checks := hc.GetCustomChecks()
		assert.Contains(t, checks, "test_check")
		assert.NotNil(t, checks["test_check"])
	})

	t.Run("移除自定义健康检查", func(t *testing.T) {
		hc.AddCustomCheck("temp_check", func() error {
			return nil
		})

		checks := hc.GetCustomChecks()
		assert.Contains(t, checks, "temp_check")

		hc.RemoveCustomCheck("temp_check")
		checks = hc.GetCustomChecks()
		assert.NotContains(t, checks, "temp_check")
	})

	t.Run("获取运行时间", func(t *testing.T) {
		uptime := hc.GetUptime()
		assert.Greater(t, uptime, time.Duration(0))

		uptimeStr := hc.GetUptimeString()
		assert.NotEmpty(t, uptimeStr)
		assert.Contains(t, uptimeStr, "s")
	})
}

func TestHealthControllerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hc := Controllers.NewHealthController()

	// 添加测试自定义检查
	hc.AddCustomCheck("success_check", func() error {
		return nil
	})

	hc.AddCustomCheck("failure_check", func() error {
		return assert.AnError
	})

	t.Run("健康检查端点", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", hc.Health)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")

		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "status")
		assert.Contains(t, data, "timestamp")
		assert.Contains(t, data, "version")
		assert.Contains(t, data, "uptime")
		assert.Contains(t, data, "services")
		assert.Contains(t, data, "metrics")
	})

	t.Run("详细健康检查端点", func(t *testing.T) {
		router := gin.New()
		router.GET("/health/detailed", hc.DetailedHealth)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/detailed", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
	})

	t.Run("就绪检查端点", func(t *testing.T) {
		router := gin.New()
		router.GET("/health/ready", hc.ReadinessCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/ready", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
	})

	t.Run("存活检查端点", func(t *testing.T) {
		router := gin.New()
		router.GET("/health/live", hc.LivenessCheck)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/live", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
	})
}

func TestHealthControllerCustomChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hc := Controllers.NewHealthController()

	t.Run("成功自定义检查", func(t *testing.T) {
		hc.AddCustomCheck("success_check", func() error {
			return nil
		})

		router := gin.New()
		router.GET("/health", hc.Health)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		services := data["services"].(map[string]interface{})

		// 检查自定义检查结果
		if successCheck, exists := services["success_check"]; exists {
			checkData := successCheck.(map[string]interface{})
			assert.Equal(t, "healthy", checkData["status"])
			assert.Equal(t, "OK", checkData["message"])
		}
	})

	t.Run("失败自定义检查", func(t *testing.T) {
		hc.AddCustomCheck("failure_check", func() error {
			return assert.AnError
		})

		router := gin.New()
		router.GET("/health", hc.Health)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		// 由于有失败的检查，状态码应该是503
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		services := data["services"].(map[string]interface{})

		// 检查自定义检查结果
		if failureCheck, exists := services["failure_check"]; exists {
			checkData := failureCheck.(map[string]interface{})
			assert.Equal(t, "unhealthy", checkData["status"])
			assert.Contains(t, checkData["message"], "assert.AnError")
		}
	})
}

func TestHealthControllerConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hc := Controllers.NewHealthController()

	// 添加一些自定义检查
	for i := 0; i < 5; i++ {
		checkName := fmt.Sprintf("check_%d", i)
		hc.AddCustomCheck(checkName, func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	t.Run("并发健康检查", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", hc.Health)

		// 并发发送请求
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/health", nil)
				router.ServeHTTP(w, req)
				done <- true
			}()
		}

		// 等待所有请求完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func BenchmarkHealthController(b *testing.B) {
	gin.SetMode(gin.TestMode)

	hc := Controllers.NewHealthController()

	// 添加一些自定义检查
	for i := 0; i < 10; i++ {
		checkName := fmt.Sprintf("check_%d", i)
		hc.AddCustomCheck(checkName, func() error {
			return nil
		})
	}

	router := gin.New()
	router.GET("/health", hc.Health)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func TestHealthControllerEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hc := Controllers.NewHealthController()

	t.Run("空自定义检查列表", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", hc.Health)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("长时间运行的自定义检查", func(t *testing.T) {
		hc.AddCustomCheck("slow_check", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		router := gin.New()
		router.GET("/health", hc.Health)

		start := time.Now()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Greater(t, duration, 100*time.Millisecond)
	})

	t.Run("panic的自定义检查", func(t *testing.T) {
		hc.AddCustomCheck("panic_check", func() error {
			panic("test panic")
		})

		router := gin.New()
		router.GET("/health", hc.Health)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		// 应该处理panic并返回错误状态
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
