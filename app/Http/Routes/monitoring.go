package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"

	"github.com/gin-gonic/gin"
)

// RegisterMonitoringRoutes 注册监控告警路由
// 功能说明：
// 1. 注册系统监控指标相关路由
// 2. 注册告警管理和配置路由
// 3. 注册性能监控和分析路由
// 4. 所有路由都需要认证访问
func RegisterMonitoringRoutes(router *gin.Engine, controller *Controllers.MonitoringController) {
	// 监控告警路由组，需要认证
	monitoringGroup := router.Group("/api/v1/monitoring")
	monitoringGroup.Use(Middleware.NewAuthMiddleware().Handle())
	{
		// 监控指标相关路由
		monitoringGroup.GET("/metrics", controller.GetMetrics)

		// 告警相关路由
		monitoringGroup.GET("/alerts", controller.GetAlerts)
		monitoringGroup.POST("/alerts/:id/acknowledge", controller.AcknowledgeAlert)
		monitoringGroup.POST("/alerts/:id/resolve", controller.ResolveAlert)

		// 告警规则相关路由
		monitoringGroup.GET("/alert-rules", controller.GetAlertRules)
		monitoringGroup.POST("/alert-rules", controller.CreateAlertRule)
		monitoringGroup.PUT("/alert-rules/:id", controller.UpdateAlertRule)
		monitoringGroup.DELETE("/alert-rules/:id", controller.DeleteAlertRule)

		// 系统健康状态
		monitoringGroup.GET("/health", controller.GetSystemHealth)

		// 通知记录
		monitoringGroup.GET("/notifications", controller.GetNotificationRecords)

		// 监控统计信息
		monitoringGroup.GET("/stats", controller.GetMonitoringStats)
	}
}
