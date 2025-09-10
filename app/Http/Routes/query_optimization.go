package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"

	"github.com/gin-gonic/gin"
)

// RegisterQueryOptimizationRoutes 注册查询优化路由
// 功能说明：
// 1. 注册慢查询管理相关路由
// 2. 注册查询统计和分析路由
// 3. 注册索引建议和优化路由
// 4. 所有路由都需要认证访问
func RegisterQueryOptimizationRoutes(router *gin.Engine, controller *Controllers.QueryOptimizationController) {
	// 查询优化路由组，需要认证
	queryOptGroup := router.Group("/api/v1/query-optimization")
	queryOptGroup.Use(Middleware.NewAuthMiddleware().Handle())
	{
		// 慢查询相关路由
		queryOptGroup.GET("/slow-queries", controller.GetSlowQueries)

		// 查询统计相关路由
		queryOptGroup.GET("/query-statistics", controller.GetQueryStatistics)

		// 索引建议相关路由
		queryOptGroup.GET("/index-suggestions", controller.GetIndexSuggestions)
		queryOptGroup.POST("/index-suggestions/:suggestion_id/apply", controller.ApplyIndexSuggestion)
		queryOptGroup.POST("/index-suggestions/:suggestion_id/reject", controller.RejectIndexSuggestion)

		// 性能报告相关路由
		queryOptGroup.GET("/performance-report", controller.GetPerformanceReport)

		// 优化报告生成
		queryOptGroup.POST("/generate-report", controller.GenerateOptimizationReport)

		// 优化摘要
		queryOptGroup.GET("/summary", controller.GetOptimizationSummary)
	}
}
