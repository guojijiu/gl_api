package Middleware

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"time"
)

// HealthCheckMiddleware 健康检查中间件
type HealthCheckMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
}

// NewHealthCheckMiddleware 创建健康检查中间件
// 功能说明：
// 1. 初始化健康检查中间件实例
// 2. 用于监控应用运行状态
// 3. 检查数据库连接状态
// 4. 监控系统资源使用情况
func NewHealthCheckMiddleware(storageManager *Storage.StorageManager) *HealthCheckMiddleware {
	return &HealthCheckMiddleware{
		storageManager: storageManager,
	}
}

// Handle 处理健康检查请求
// 功能说明：
// 1. 检查数据库连接状态
// 2. 监控内存使用情况
// 3. 检查应用运行时间
// 4. 返回详细的健康状态信息
func (m *HealthCheckMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 检查数据库连接
		dbStatus := "healthy"
		db := Database.GetDB()
		if db == nil {
			dbStatus = "unhealthy"
		} else {
			sqlDB, err := db.DB()
			if err != nil {
				dbStatus = "unhealthy"
			} else {
				if err := sqlDB.Ping(); err != nil {
					dbStatus = "unhealthy"
				}
			}
		}

		// 获取内存统计
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// 计算响应时间
		duration := time.Since(start)

		// 构建健康状态响应
		healthStatus := gin.H{
			"status":        "healthy",
			"timestamp":     time.Now().Format(time.RFC3339),
			"uptime":        time.Since(start).String(),
			"response_time": duration.String(),
			"services": gin.H{
				"database": dbStatus,
			},
			"system": gin.H{
				"memory_alloc":       memStats.Alloc,
				"memory_total_alloc": memStats.TotalAlloc,
				"memory_sys":         memStats.Sys,
				"num_goroutines":     runtime.NumGoroutine(),
			},
		}

		// 如果数据库不健康，整体状态为不健康
		if dbStatus == "unhealthy" {
			healthStatus["status"] = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, healthStatus)
		} else {
			c.JSON(http.StatusOK, healthStatus)
		}

		// 记录健康检查日志
		m.storageManager.LogInfo("健康检查", map[string]interface{}{
			"status":          healthStatus["status"],
			"response_time":   duration.String(),
			"database_status": dbStatus,
		})
	}
}
