package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"

	"github.com/gin-gonic/gin"
)

// RegisterSecurityRoutes 注册安全防护路由
func RegisterSecurityRoutes(router *gin.Engine, controller *Controllers.SecurityController) {
	// 安全防护路由组，需要认证
	securityGroup := router.Group("/api/v1/security")
	securityGroup.Use(Middleware.NewAuthMiddleware().Handle())
	{
		// 安全事件相关路由
		securityGroup.GET("/events", controller.GetSecurityEvents)
		
		// 威胁情报相关路由
		securityGroup.GET("/threats", controller.GetThreatIntelligence)
		
		// 登录尝试相关路由
		securityGroup.GET("/login-attempts", controller.GetLoginAttempts)
		
		// 账户锁定相关路由
		securityGroup.GET("/account-lockouts", controller.GetAccountLockouts)
		// 账户管理
		// TODO: 实现账户解锁功能
		
		// 安全告警相关路由
		securityGroup.GET("/alerts", controller.GetSecurityAlerts)
		// TODO: 实现告警确认功能
		// TODO: 实现告警解决功能
		
		// 安全报告相关路由
		securityGroup.GET("/reports", controller.GetSecurityReports)
		// TODO: 实现安全报告生成功能
		
		// 安全指标相关路由
		// TODO: 实现安全指标获取功能
		
		// 安全仪表板相关路由
		// TODO: 实现安全仪表板功能
	}
}
