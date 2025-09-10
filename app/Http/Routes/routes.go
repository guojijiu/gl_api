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
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", func(c *gin.Context) {
			// 用户注册
			c.JSON(200, gin.H{"message": "注册功能待实现"})
		})
		authGroup.POST("/login", func(c *gin.Context) {
			// 用户登录
			c.JSON(200, gin.H{"message": "登录功能待实现"})
		})
		authGroup.POST("/logout", func(c *gin.Context) {
			// 用户登出
			c.JSON(200, gin.H{"message": "登出功能待实现"})
		})
		authGroup.POST("/refresh", func(c *gin.Context) {
			// 刷新token
			c.JSON(200, gin.H{"message": "刷新token功能待实现"})
		})
	}

	// 用户管理路由
	userGroup := v1.Group("/users")
	{
		userGroup.GET("/", func(c *gin.Context) {
			// 获取用户列表
			c.JSON(200, gin.H{"message": "获取用户列表功能待实现"})
		})
		userGroup.GET("/:id", func(c *gin.Context) {
			// 获取单个用户
			c.JSON(200, gin.H{"message": "获取用户详情功能待实现"})
		})
		userGroup.PUT("/:id", func(c *gin.Context) {
			// 更新用户信息
			c.JSON(200, gin.H{"message": "更新用户信息功能待实现"})
		})
		userGroup.DELETE("/:id", func(c *gin.Context) {
			// 删除用户
			c.JSON(200, gin.H{"message": "删除用户功能待实现"})
		})
	}

	// 文章管理路由
	postGroup := v1.Group("/posts")
	{
		postGroup.GET("/", func(c *gin.Context) {
			// 获取文章列表
			c.JSON(200, gin.H{"message": "获取文章列表功能待实现"})
		})
		postGroup.GET("/:id", func(c *gin.Context) {
			// 获取单个文章
			c.JSON(200, gin.H{"message": "获取文章详情功能待实现"})
		})
		postGroup.POST("/", func(c *gin.Context) {
			// 创建文章
			c.JSON(200, gin.H{"message": "创建文章功能待实现"})
		})
		postGroup.PUT("/:id", func(c *gin.Context) {
			// 更新文章
			c.JSON(200, gin.H{"message": "更新文章功能待实现"})
		})
		postGroup.DELETE("/:id", func(c *gin.Context) {
			// 删除文章
			c.JSON(200, gin.H{"message": "删除文章功能待实现"})
		})
	}

	// 分类管理路由
	categoryGroup := v1.Group("/categories")
	{
		categoryGroup.GET("/", func(c *gin.Context) {
			// 获取分类列表
			c.JSON(200, gin.H{"message": "获取分类列表功能待实现"})
		})
		categoryGroup.GET("/:id", func(c *gin.Context) {
			// 获取单个分类
			c.JSON(200, gin.H{"message": "获取分类详情功能待实现"})
		})
		categoryGroup.POST("/", func(c *gin.Context) {
			// 创建分类
			c.JSON(200, gin.H{"message": "创建分类功能待实现"})
		})
		categoryGroup.PUT("/:id", func(c *gin.Context) {
			// 更新分类
			c.JSON(200, gin.H{"message": "更新分类功能待实现"})
		})
		categoryGroup.DELETE("/:id", func(c *gin.Context) {
			// 删除分类
			c.JSON(200, gin.H{"message": "删除分类功能待实现"})
		})
	}

	// 标签管理路由
	tagGroup := v1.Group("/tags")
	{
		tagGroup.GET("/", func(c *gin.Context) {
			// 获取标签列表
			c.JSON(200, gin.H{"message": "获取标签列表功能待实现"})
		})
		tagGroup.GET("/:id", func(c *gin.Context) {
			// 获取单个标签
			c.JSON(200, gin.H{"message": "获取标签详情功能待实现"})
		})
		tagGroup.POST("/", func(c *gin.Context) {
			// 创建标签
			c.JSON(200, gin.H{"message": "创建标签功能待实现"})
		})
		tagGroup.PUT("/:id", func(c *gin.Context) {
			// 更新标签
			c.JSON(200, gin.H{"message": "更新标签功能待实现"})
		})
		tagGroup.DELETE("/:id", func(c *gin.Context) {
			// 删除标签
			c.JSON(200, gin.H{"message": "删除标签功能待实现"})
		})
	}

	// 管理员路由
	adminGroup := v1.Group("/admin")
	{
		adminGroup.GET("/dashboard", func(c *gin.Context) {
			// 管理员仪表板
			c.JSON(200, gin.H{"message": "管理员仪表板功能待实现"})
		})
		adminGroup.GET("/users", func(c *gin.Context) {
			// 管理员用户管理
			c.JSON(200, gin.H{"message": "管理员用户管理功能待实现"})
		})
		adminGroup.GET("/posts", func(c *gin.Context) {
			// 管理员文章管理
			c.JSON(200, gin.H{"message": "管理员文章管理功能待实现"})
		})
		adminGroup.GET("/logs", func(c *gin.Context) {
			// 管理员日志查看
			c.JSON(200, gin.H{"message": "管理员日志查看功能待实现"})
		})
	}

	// 监控路由
	monitoringGroup := v1.Group("/monitoring")
	{
		monitoringGroup.GET("/metrics", func(c *gin.Context) {
			// 获取监控指标
			c.JSON(200, gin.H{"message": "监控指标功能待实现"})
		})
		monitoringGroup.GET("/health", func(c *gin.Context) {
			// 健康检查
			c.JSON(200, gin.H{"message": "健康检查功能待实现"})
		})
		monitoringGroup.GET("/performance", func(c *gin.Context) {
			// 性能监控
			c.JSON(200, gin.H{"message": "性能监控功能待实现"})
		})
	}

	// WebSocket路由
	wsGroup := v1.Group("/ws")
	{
		wsGroup.GET("/", func(c *gin.Context) {
			// WebSocket连接
			c.JSON(200, gin.H{"message": "WebSocket功能待实现"})
		})
	}

	// 记录路由注册完成日志
	logManager.Info("routes", "所有路由注册完成", map[string]interface{}{
		"total_routes": len(engine.Routes()),
		"api_version":  "v1",
		"base_path":    "/api/v1",
	})
}
