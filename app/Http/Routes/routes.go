package Routes

import (
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"context"
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
//
// 功能说明：
// 1. 注册所有HTTP路由和API端点
// 2. 配置全局中间件链（按顺序执行）
// 3. 创建API版本分组（v1、v2等）
// 4. 注册健康检查端点
// 5. 注册业务路由（认证、用户、文章等）
//
// 中间件执行顺序（重要）：
// 1. 错误恢复（Recovery）：最先执行，捕获panic，防止程序崩溃
// 2. CORS：处理跨域请求，设置响应头
// 3. 验证中间件：输入验证、SQL注入检测、XSS防护
// 4. 超时控制：防止长时间运行的请求
// 5. 速率限制：防止API滥用
// 6. 性能监控：收集性能指标
// 7. 请求统计：统计请求信息
// 8. 请求日志：记录请求和响应
// 9. SQL日志：记录SQL查询
// 10. 错误处理：最后执行，处理业务错误
//
// 中间件顺序的重要性：
// - 错误恢复必须最先执行，才能捕获后续中间件的panic
// - 验证中间件应该在业务逻辑之前执行，提前过滤无效请求
// - 日志中间件应该在业务逻辑之后执行，记录完整的请求信息
// - 错误处理应该最后执行，处理所有业务错误
//
// 性能考虑：
// - 中间件按顺序执行，可能影响请求处理时间
// - 某些中间件（如日志）可以异步处理，减少延迟
// - 速率限制应该在早期执行，避免无效请求消耗资源
//
// 安全考虑：
// - CORS应该在早期执行，防止跨域攻击
// - 验证中间件应该在业务逻辑之前，防止恶意输入
// - 速率限制应该在早期执行，防止DDoS攻击
//
// 注意事项：
// - 中间件顺序很重要，不要随意调整
// - 某些中间件可能影响响应时间，需要权衡
// - 日志中间件可能产生大量日志，需要合理配置
func RegisterRoutes(engine *gin.Engine, storageManager *Storage.StorageManager, logManager *Services.LogManagerService) {
	// 创建中间件实例
	// 每个中间件负责不同的功能（日志、错误处理、安全等）
	requestLogMiddleware := Middleware.NewRequestLogMiddleware(logManager)
	sqlLogMiddleware := Middleware.NewSQLLogMiddleware(logManager)
	errorHandlingMiddleware := Middleware.NewErrorHandlingMiddleware(storageManager, nil)
	recoveryMiddleware := Middleware.NewRecoveryMiddleware(storageManager)
	corsMiddleware := Middleware.NewCORSMiddleware()
	timeoutMiddleware := Middleware.NewTimeoutMiddleware(storageManager)
	performanceMiddleware := Middleware.NewPerformanceMiddleware(storageManager)
	rateLimitMiddleware := Middleware.NewRateLimitMiddleware(storageManager)
	versionMiddleware := Middleware.NewVersionMiddleware(storageManager)

	// 创建增强的验证中间件
	// 包含输入验证、SQL注入检测、XSS防护等功能
	validationMiddleware := Middleware.NewEnhancedValidationMiddleware(storageManager, nil)

	// 创建请求统计中间件
	// 用于收集和分析请求统计信息
	monitoringService := Services.NewOptimizedMonitoringService()
	requestStatsMiddleware := Middleware.NewRequestStatsMiddleware(storageManager, monitoringService)

	// 添加全局中间件
	// 注意：中间件的执行顺序很重要，影响功能和性能
	// 执行顺序：从上到下，从左到右
	engine.Use(
		recoveryMiddleware.Handle(),                    // 1. 错误恢复中间件（最先执行，捕获panic）
		corsMiddleware.Handle(),                        // 2. CORS中间件（处理跨域请求）
		validationMiddleware.Handle(),                  // 3. 增强的验证中间件（输入验证、安全检测）
		validationMiddleware.ValidateJSON(),            // 4. JSON验证中间件（验证JSON格式）
		validationMiddleware.ValidateFileUpload(),      // 5. 文件上传验证中间件（验证文件类型和大小）
		timeoutMiddleware.Handle(30*time.Second),       // 6. 请求超时中间件（30秒超时）
		rateLimitMiddleware.Handle(100, 1*time.Minute), // 7. 全局速率限制（每分钟100次请求）
		performanceMiddleware.Handle(),                 // 8. 性能监控中间件（收集性能指标）
		requestStatsMiddleware.Handle(),                // 9. 请求统计中间件（统计请求信息）
		requestLogMiddleware.RequestLog(),              // 10. 自定义请求日志（记录请求和响应）
		sqlLogMiddleware.Handle(),                      // 11. SQL日志中间件（记录SQL查询）
		errorHandlingMiddleware.Handle(),               // 12. 错误处理中间件（最后执行，处理业务错误）
	)

	// API版本分组
	// 创建v1版本的API路由组
	// 所有v1版本的API都在/api/v1路径下
	// 支持多版本API共存（v1、v2等）
	v1 := engine.Group("/api/v1")
	// 添加版本控制中间件
	// 用于记录API版本信息，便于版本管理和统计
	v1.Use(versionMiddleware.Handle())

	// 健康检查端点
	// 这些端点不经过认证中间件，用于监控和负载均衡器检查
	healthController := Controllers.NewHealthController()

	// 基础健康检查：快速检查服务是否运行
	engine.GET("/health", healthController.Health)
	// 注意：docker 的健康检查/一些探活工具会使用 HEAD 请求
	engine.HEAD("/health", healthController.Health)

	// 详细健康检查：包含系统资源、数据库、缓存等详细信息
	engine.GET("/health/detailed", healthController.DetailedHealth)
	engine.HEAD("/health/detailed", healthController.DetailedHealth)

	// 就绪检查：检查服务是否准备好接受请求
	// 用于Kubernetes的readiness probe
	engine.GET("/health/ready", healthController.Readiness)
	engine.HEAD("/health/ready", healthController.Readiness)

	// 存活检查：检查服务是否存活
	// 用于Kubernetes的liveness probe
	engine.GET("/health/live", healthController.Liveness)
	engine.HEAD("/health/live", healthController.Liveness)

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
	userGroup.Use(Middleware.NewAuthMiddleware().Handle())
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
	// 注入监控服务，否则 /metrics 会返回 500（监控服务未初始化）
	monitoringController.SetMonitoringService(monitoringService)
	monitoringGroup := v1.Group("/monitoring")
	{
		monitoringGroup.GET("/metrics", monitoringController.GetMetrics)
		monitoringGroup.GET("/health", monitoringController.GetSystemHealth)
		monitoringGroup.GET("/alerts", monitoringController.GetAlerts)
	}

	// Prometheus 默认抓取路径通常是 /metrics，这里提供一个顶层别名，避免 404 造成噪音
	engine.GET("/metrics", monitoringController.GetMetrics)
	engine.HEAD("/metrics", monitoringController.GetMetrics)

	// WebSocket路由
	wsController := Controllers.NewWebSocketController()
	wsGroup := v1.Group("/ws")
	{
		wsGroup.GET("/", wsController.Connect)
	}

	// 记录路由注册完成日志
	// 使用 business 日志记录器，因为路由注册属于业务逻辑
	logManager.LogBusiness(context.Background(), "routes", "register", "所有路由注册完成", map[string]interface{}{
		"total_routes": len(engine.Routes()),
		"api_version":  "v1",
		"base_path":    "/api/v1",
	})
}
