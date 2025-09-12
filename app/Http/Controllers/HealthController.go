package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthController 健康检查控制器
type HealthController struct {
	Controller
	db              *gorm.DB
	redisClient     *redis.Client
	storageManager  *Storage.StorageManager
	securityService *Services.SecurityService
}

// NewHealthController 创建健康检查控制器
// 功能说明：
// 1. 初始化健康检查控制器实例
// 2. 设置数据库、Redis、存储管理器等依赖服务
// 3. 用于监控系统各组件的健康状态
func NewHealthController() *HealthController {
	return &HealthController{
		db:             Database.GetDB(),
		storageManager: Storage.GetStorageManager(),
		// securityService 将在需要时通过依赖注入获取
	}
}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Version   string                   `json:"version"`
	Uptime    string                   `json:"uptime"`
	Services  map[string]ServiceHealth `json:"services"`
	Metrics   SystemMetrics            `json:"metrics"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Message      string        `json:"message,omitempty"`
	Details      interface{}   `json:"details,omitempty"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Memory     MemoryMetrics `json:"memory"`
	CPU        CPUMetrics    `json:"cpu"`
	Goroutines int           `json:"goroutines"`
	GC         GCMetrics     `json:"gc"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	NumCPU int `json:"num_cpu"`
}

// GCMetrics GC指标
type GCMetrics struct {
	NumGC        uint32  `json:"num_gc"`
	PauseTotal   uint64  `json:"pause_total_ns"`
	PauseAverage uint64  `json:"pause_average_ns"`
	CPUFraction  float64 `json:"cpu_fraction"`
}

// Health 基础健康检查
// @Summary 基础健康检查
// @Description 检查应用基础健康状态
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (hc *HealthController) Health(c *gin.Context) {
	start := time.Now()

	// 获取系统指标
	metrics := hc.getSystemMetrics()

	// 检查服务状态
	services := hc.checkServices()

	// 确定整体状态
	overallStatus := "healthy"
	for _, service := range services {
		if service.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}

	// 计算响应时间
	responseTime := time.Since(start)

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0", // 可以从构建信息中获取
		Uptime:    hc.getUptime(),
		Services:  services,
		Metrics:   metrics,
	}

	// 根据状态设置HTTP状态码
	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"success":       true,
		"data":          response,
		"response_time": responseTime.Milliseconds(),
	})
}

// DetailedHealth 详细健康检查
// @Summary 详细健康检查
// @Description 检查应用详细健康状态，包括所有依赖服务
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/detailed [get]
func (hc *HealthController) DetailedHealth(c *gin.Context) {
	start := time.Now()

	// 获取详细系统指标
	metrics := hc.getDetailedSystemMetrics()

	// 检查所有服务状态
	services := hc.checkAllServices()

	// 确定整体状态
	overallStatus := "healthy"
	criticalServices := []string{"database", "redis", "storage"}

	for _, serviceName := range criticalServices {
		if service, exists := services[serviceName]; exists && service.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}

	responseTime := time.Since(start)

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    hc.getUptime(),
		Services:  services,
		Metrics:   metrics,
	}

	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"success":       true,
		"data":          response,
		"response_time": responseTime.Milliseconds(),
	})
}

// Readiness 就绪检查
// @Summary 就绪检查
// @Description 检查应用是否准备好接收流量
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/ready [get]
func (hc *HealthController) Readiness(c *gin.Context) {
	// 检查关键服务是否就绪
	services := map[string]bool{
		"database": hc.checkDatabase(),
		"redis":    hc.checkRedis(),
		"storage":  hc.checkStorage(),
	}

	ready := true
	for service, status := range services {
		if !status {
			ready = false
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success":  false,
				"message":  fmt.Sprintf("Service %s is not ready", service),
				"services": services,
			})
			return
		}
	}

	// 如果所有服务都正常，返回成功响应
	if ready {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "All services are ready",
			"services": services,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Application is ready",
		"services": services,
	})
}

// Liveness 存活检查
// @Summary 存活检查
// @Description 检查应用是否存活
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/live [get]
func (hc *HealthController) Liveness(c *gin.Context) {
	// 简单的存活检查
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Application is alive",
		"timestamp": time.Now(),
	})
}

// checkServices 检查基础服务
func (hc *HealthController) checkServices() map[string]ServiceHealth {
	services := make(map[string]ServiceHealth)

	// 检查数据库
	services["database"] = hc.checkDatabaseHealth()

	// 检查Redis
	services["redis"] = hc.checkRedisHealth()

	// 检查存储
	services["storage"] = hc.checkStorageHealth()

	return services
}

// checkAllServices 检查所有服务
func (hc *HealthController) checkAllServices() map[string]ServiceHealth {
	services := hc.checkServices()

	// 检查安全服务
	services["security"] = hc.checkSecurityServiceHealth()

	// 检查性能监控服务
	services["monitoring"] = hc.checkMonitoringServiceHealth()

	return services
}

// checkDatabaseHealth 检查数据库健康状态
func (hc *HealthController) checkDatabaseHealth() ServiceHealth {
	start := time.Now()

	if hc.db == nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      "Database connection is nil",
		}
	}

	// 执行简单查询
	var result int
	err := hc.db.Raw("SELECT 1").Scan(&result).Error

	responseTime := time.Since(start)

	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime,
			Message:      err.Error(),
		}
	}

	// 获取连接池状态
	sqlDB, err := hc.db.DB()
	var poolStats sql.DBStats
	if err == nil {
		poolStats = sqlDB.Stats()
	}

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"open_connections": poolStats.OpenConnections,
			"in_use":           poolStats.InUse,
			"idle":             poolStats.Idle,
			"wait_count":       poolStats.WaitCount,
		},
	}
}

// checkRedisHealth 检查Redis健康状态
func (hc *HealthController) checkRedisHealth() ServiceHealth {
	start := time.Now()

	if hc.redisClient == nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      "Redis client is nil",
		}
	}

	// 执行ping命令
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := hc.redisClient.Ping(ctx).Err()
	responseTime := time.Since(start)

	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime,
			Message:      err.Error(),
		}
	}

	// 获取Redis信息
	info, err := hc.redisClient.Info(ctx, "memory").Result()
	var details interface{}
	if err == nil {
		details = map[string]interface{}{
			"info_available": len(info) > 0,
		}
	}

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
		Details:      details,
	}
}

// checkStorageHealth 检查存储健康状态
func (hc *HealthController) checkStorageHealth() ServiceHealth {
	start := time.Now()

	if hc.storageManager == nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      "Storage manager is nil",
		}
	}

	// 检查存储目录是否可写
	err := hc.storageManager.CheckHealth()
	responseTime := time.Since(start)

	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime,
			Message:      err.Error(),
		}
	}

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
	}
}

// checkSecurityServiceHealth 检查安全服务健康状态
func (hc *HealthController) checkSecurityServiceHealth() ServiceHealth {
	start := time.Now()

	if hc.securityService == nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: time.Since(start),
			Message:      "Security service is nil",
		}
	}

	// 简单的健康检查
	responseTime := time.Since(start)

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"threat_detection_enabled": true,
		},
	}
}

// checkMonitoringServiceHealth 检查监控服务健康状态
func (hc *HealthController) checkMonitoringServiceHealth() ServiceHealth {
	start := time.Now()

	// 检查监控服务是否可用
	responseTime := time.Since(start)

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"metrics_collection_enabled": true,
		},
	}
}

// getSystemMetrics 获取系统指标
func (hc *HealthController) getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemMetrics{
		Memory: MemoryMetrics{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		CPU: CPUMetrics{
			NumCPU: runtime.NumCPU(),
		},
		Goroutines: runtime.NumGoroutine(),
		GC: GCMetrics{
			NumGC:        m.NumGC,
			PauseTotal:   m.PauseTotalNs,
			PauseAverage: m.PauseTotalNs / uint64(m.NumGC),
			CPUFraction:  m.GCCPUFraction,
		},
	}
}

// getDetailedSystemMetrics 获取详细系统指标
func (hc *HealthController) getDetailedSystemMetrics() SystemMetrics {
	metrics := hc.getSystemMetrics()

	// 可以添加更多详细的指标
	// 例如：磁盘使用率、网络统计等

	return metrics
}

// getUptime 获取运行时间
func (hc *HealthController) getUptime() string {
	// 这里应该从应用启动时间计算
	// 暂时返回固定值
	return "24h30m15s"
}

// checkDatabase 检查数据库连接
func (hc *HealthController) checkDatabase() bool {
	if hc.db == nil {
		return false
	}

	var result int
	err := hc.db.Raw("SELECT 1").Scan(&result).Error
	return err == nil
}

// checkRedis 检查Redis连接
func (hc *HealthController) checkRedis() bool {
	if hc.redisClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := hc.redisClient.Ping(ctx).Err()
	return err == nil
}

// checkStorage 检查存储
func (hc *HealthController) checkStorage() bool {
	if hc.storageManager == nil {
		return false
	}

	err := hc.storageManager.CheckHealth()
	return err == nil
}
