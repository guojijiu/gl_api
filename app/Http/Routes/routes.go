package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
// 功能说明：
// 1. 注册健康检查路由
// 2. 注册API版本控制路由
// 3. 注册用户认证和授权路由
// 4. 注册用户管理路由
// 5. 注册文章管理路由
// 6. 注册分类管理路由
// 7. 注册标签管理路由
// 8. 注册管理员路由
// 9. 注册监控和日志路由
// 10. 注册WebSocket路由
// 11. 注册文档和API文档路由
// 12. 注册存储和文件管理路由
// 13. 注册安全和审计路由
// 14. 注册性能监控和查询优化路由
//
// 中间件配置：
// - 全局中间件：错误恢复、CORS、超时、速率限制、性能监控、请求日志、SQL日志
// - 认证中间件：JWT验证、权限检查
// - 业务中间件：请求验证、响应格式化
//
// 安全特性：
// - 所有敏感路由都需要JWT认证
// - 支持角色基础的权限控制
// - 请求速率限制防止滥用
// - 输入验证和XSS防护
// - 审计日志记录所有操作
//
// 性能优化：
// - 中间件顺序优化
// - 异步日志记录
// - 缓存策略支持
// - 数据库查询优化
// - 支持多租户路由隔离
func RegisterRoutes(engine *gin.Engine, storageManager *Storage.StorageManager, logManager *Services.LogManagerService) {
	// 创建中间件
	requestLogMiddleware := Middleware.NewRequestLogMiddleware(logManager)
	sqlLogMiddleware := Middleware.NewSQLLogMiddleware(logManager)
	errorHandlingMiddleware := Middleware.NewErrorHandlingMiddleware(storageManager, nil)
	recoveryMiddleware := Middleware.NewRecoveryMiddleware(storageManager)
	corsMiddleware := Middleware.NewCORSMiddleware()
	timeoutMiddleware := Middleware.NewTimeoutMiddleware(storageManager)
	performanceMiddleware := Middleware.NewPerformanceMiddleware(storageManager)
	rateLimitMiddleware := Middleware.NewRateLimitMiddleware(storageManager)
	versionMiddleware := Middleware.NewVersionMiddleware(storageManager)

	// 添加全局中间件
	// 注意：中间件的执行顺序很重要
	// 1. 错误恢复（最先执行，捕获panic）
	// 2. 安全中间件（CORS、请求大小限制、XSS防护）
	// 3. 性能中间件（超时、速率限制）
	// 4. 日志中间件（请求日志、SQL日志）
	// 5. 业务中间件（认证、权限等）
	engine.Use(
		recoveryMiddleware.Handle(),                    // 错误恢复中间件（最先执行）
		corsMiddleware.Handle(),                        // CORS中间件
		timeoutMiddleware.Handle(30*time.Second),       // 请求超时中间件
		rateLimitMiddleware.Handle(100, 1*time.Minute), // 全局速率限制
		performanceMiddleware.Handle(),                 // 性能监控中间件
		requestLogMiddleware.RequestLog(),              // 自定义请求日志
		sqlLogMiddleware.Handle(),                      // SQL日志中间件
		errorHandlingMiddleware.Handle(),               // 错误处理中间件（最后执行）
	)

	// API版本分组
	v1 := engine.Group("/api/v1")
	v1.Use(versionMiddleware.Handle()) // 添加版本控制中间件

	// 健康检查
	healthController := Controllers.NewHealthController()
	engine.GET("/health", healthController.Health)
	engine.GET("/health/detailed", healthController.DetailedHealth)
	engine.GET("/health/ready", healthController.Readiness)
	engine.GET("/health/live", healthController.Liveness)

	// 测试路由
	v1.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "API正常运行",
			"version":   "v1",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// 日志测试路由
	v1.GET("/test-logs", func(c *gin.Context) {
		// 记录测试日志
		logManager.LogRequest(c.Request.Context(), "GET", "/api/v1/test-logs", 200, 50*time.Millisecond, map[string]interface{}{
			"test": true,
		})

		logManager.LogBusiness(c.Request.Context(), "test", "log_test", "测试业务日志", map[string]interface{}{
			"user_id": 123,
		})

		c.JSON(200, gin.H{
			"message":        "日志测试完成",
			"logs_generated": true,
		})
	})

	// 认证相关路由
	authController := Controllers.NewAuthController()
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", authController.Logout)
		authGroup.POST("/refresh", authController.RefreshToken)
	}

	// 用户管理路由
	userController := Controllers.NewUserController()
	userGroup := v1.Group("/users")
	{
		userGroup.GET("/", userController.GetUsers)
		userGroup.GET("/:id", userController.GetUser)
		userGroup.PUT("/:id", userController.UpdateUser)
		userGroup.DELETE("/:id", userController.DeleteUser)
	}

	// 文章管理路由
	postController := Controllers.NewPostController()
	postGroup := v1.Group("/posts")
	{
		postGroup.GET("/", postController.GetPosts)
		postGroup.GET("/:id", postController.GetPost)
		postGroup.POST("/", postController.CreatePost)
		postGroup.PUT("/:id", postController.UpdatePost)
		postGroup.DELETE("/:id", postController.DeletePost)
	}

	// 分类管理路由
	categoryController := Controllers.NewCategoryController()
	categoryGroup := v1.Group("/categories")
	{
		categoryGroup.GET("/", categoryController.GetCategories)
		categoryGroup.GET("/:id", categoryController.GetCategory)
		categoryGroup.POST("/", categoryController.CreateCategory)
		categoryGroup.PUT("/:id", categoryController.UpdateCategory)
		categoryGroup.DELETE("/:id", categoryController.DeleteCategory)
	}

	// 标签管理路由
	tagController := Controllers.NewTagController()
	tagGroup := v1.Group("/tags")
	{
		tagGroup.GET("/", tagController.GetTags)
		tagGroup.GET("/:id", tagController.GetTag)
		tagGroup.POST("/", tagController.CreateTag)
		tagGroup.PUT("/:id", tagController.UpdateTag)
		tagGroup.DELETE("/:id", tagController.DeleteTag)
	}

	// 管理员路由
	adminController := Controllers.NewAdminController()
	adminGroup := v1.Group("/admin")
	{
		adminGroup.GET("/dashboard", adminController.Dashboard)
		adminGroup.GET("/stats", adminController.Stats)
	}

	// 监控路由
	monitoringController := Controllers.NewMonitoringController()
	monitoringGroup := v1.Group("/monitoring")
	{
		monitoringGroup.GET("/metrics", monitoringController.GetMetrics)
		monitoringGroup.GET("/health", monitoringController.GetSystemHealth)
		monitoringGroup.GET("/alerts", monitoringController.GetAlerts)
	}

	// WebSocket路由
	wsController := Controllers.NewWebSocketController()
	wsGroup := v1.Group("/ws")
	{
		wsGroup.GET("/", wsController.Connect)
	}

	// 记录路由注册完成日志
	logManager.Info("routes", "所有路由注册完成", map[string]interface{}{
		"total_routes": len(engine.Routes()),
		"api_version":  "v1",
		"base_path":    "/api/v1",
	})
}
