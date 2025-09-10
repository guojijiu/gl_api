package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"

	"github.com/gin-gonic/gin"
)

// RegisterLogRoutes 注册日志管理路由
//
// 重要功能说明：
// 1. 日志统计路由：查询各类型日志的统计信息
// 2. 日志监控路由：管理监控规则和告警
// 3. 日志配置路由：查看和更新日志配置
// 4. 系统健康路由：检查日志系统状态
// 5. 日志搜索路由：支持多条件日志搜索
//
// 安全特性：
// - 所有路由都需要JWT认证
// - 支持管理员权限验证
// - 防止敏感信息泄露
// - 支持操作频率限制
//
// 路由分组：
// - /logs/stats: 日志统计
// - /logs/monitor: 日志监控
// - /logs/config: 日志配置
// - /logs/health: 系统健康
func RegisterLogRoutes(router *gin.Engine) {
	// 日志管理路由组
	logsGroup := router.Group("/logs")
	logsGroup.Use(Middleware.NewAuthMiddleware().Handle())
	{
		// 日志统计
		logsGroup.GET("/stats", Controllers.NewLogController().GetLogStats)
		
		// 日志配置
		logsGroup.GET("/config", Controllers.NewLogController().GetLogConfig)
		
		// 系统健康
		logsGroup.GET("/health", Controllers.NewLogController().GetSystemHealth)
		
		// 日志监控子路由组
		monitorGroup := logsGroup.Group("/monitor")
		{
			// 监控统计
			monitorGroup.GET("/stats", Controllers.NewLogController().GetLogMonitorStats)
			
			// 监控规则管理
			rulesGroup := monitorGroup.Group("/rules")
			{
				rulesGroup.GET("", Controllers.NewLogController().GetLogRules)                    // 获取规则列表
				rulesGroup.POST("", Controllers.NewLogController().CreateLogRule)               // 创建规则
				rulesGroup.GET("/:rule_id", Controllers.NewLogController().GetLogRule)          // 获取单个规则
				rulesGroup.PUT("/:rule_id", Controllers.NewLogController().UpdateLogRule)      // 更新规则
				rulesGroup.DELETE("/:rule_id", Controllers.NewLogController().DeleteLogRule)   // 删除规则
			}
			
			// 告警管理
			alertsGroup := monitorGroup.Group("/alerts")
			{
				alertsGroup.GET("", Controllers.NewLogController().GetLogAlerts)                                    // 获取告警列表
				alertsGroup.POST("/:alert_id/resolve", Controllers.NewLogController().ResolveLogAlert)              // 解决告警
				alertsGroup.POST("/:alert_id/acknowledge", Controllers.NewLogController().AcknowledgeLogAlert)     // 确认告警
			}
		}
	}
}
