package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Services"

	"github.com/gin-gonic/gin"
)

// RegisterPerformanceMonitoringRoutes 注册性能监控路由
func RegisterPerformanceMonitoringRoutes(router *gin.Engine, controller *Controllers.PerformanceMonitoringController) {
	// 性能监控路由组，需要认证
	perfGroup := router.Group("/api/v1/performance")
	perfGroup.Use(Middleware.NewAuthMiddleware().Handle())
	{
		// 当前指标
		perfGroup.GET("/current", controller.GetCurrentMetrics)

		// 历史指标
		perfGroup.GET("/metrics", controller.GetMetricsByTimeRange)

		// 自定义指标
		perfGroup.POST("/custom-metrics", controller.RecordCustomMetric)

		// 告警相关路由
		alertGroup := perfGroup.Group("/alerts")
		{
			// 活跃告警
			alertGroup.GET("/active", controller.GetActiveAlerts)

			// 告警历史
			alertGroup.GET("/history", controller.GetAlertHistory)

			// 确认告警
			alertGroup.POST("/:id/acknowledge", controller.AcknowledgeAlert)
		}

		// 告警规则相关路由
		ruleGroup := perfGroup.Group("/alert-rules")
		{
			// 创建告警规则
			ruleGroup.POST("", controller.CreateAlertRule)

			// 更新告警规则
			ruleGroup.PUT("/:id", controller.UpdateAlertRule)

			// 删除告警规则
			ruleGroup.DELETE("/:id", controller.DeleteAlertRule)
		}

		// 统计信息
		perfGroup.GET("/stats", controller.GetMonitoringStats)

		// 系统健康状态
		perfGroup.GET("/health", controller.GetSystemHealth)
	}

	// 公开的健康检查端点（不需要认证）
	router.GET("/health", controller.GetSystemHealth)
}

// RegisterPerformanceMiddleware 注册性能监控中间件
func RegisterPerformanceMiddleware(router *gin.Engine, monitoringService *Services.OptimizedMonitoringService) {
	// 排除的路径
	excludePaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/api/v1/performance/health",
	}

	// 创建性能监控中间件
	perfMiddleware := Middleware.NewPerformanceMonitoringMiddleware(monitoringService, excludePaths)

	// 全局应用性能监控中间件
	router.Use(perfMiddleware.Handler())

	// 应用其他性能监控中间件
	router.Use(Middleware.BusinessMetricsMiddleware(monitoringService))
	router.Use(Middleware.DatabaseMetricsMiddleware(monitoringService))
	router.Use(Middleware.CacheMetricsMiddleware(monitoringService))
}
