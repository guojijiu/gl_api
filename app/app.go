package app

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Routes"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/bootstrap"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App 应用结构体
type App struct {
	Config         *Config.Config
	Router         *bootstrap.Router
	StorageManager *Storage.StorageManager
	LogManager     *Services.LogManagerService
}

// NewApp 创建新的应用实例
// 功能说明：
// 1. 加载应用配置（服务器、数据库、JWT、Redis、存储、邮件等）
// 2. 验证配置的有效性和完整性
// 3. 初始化存储管理器（创建storage目录结构）
// 4. 初始化Redis服务和缓存服务（支持降级到内存缓存）
// 5. 使用StorageManager初始化数据库（启用SQL日志记录）
// 6. 创建Gin路由引擎和中间件链
// 7. 注册所有HTTP路由和API端点
// 8. 执行数据库自动迁移（表结构同步）
// 9. 预热缓存（如果Redis可用）
// 10. 记录应用启动成功的详细日志信息
// 11. 返回完整的应用实例供启动使用
// 12. 支持优雅关闭和资源清理
//
// 错误处理：
// - 配置验证失败时立即退出
// - Redis连接失败时降级到内存缓存
// - 数据库连接失败时重试3次后退出
// - 缓存预热失败时记录警告但不影响启动
//
// 资源管理：
// - 自动创建必要的存储目录
// - 初始化数据库连接池
// - 设置Redis连接参数
// - 配置日志记录器
//
// 安全考虑：
// - 验证JWT密钥强度
// - 检查数据库连接安全性
// - 验证Redis连接参数
// - 确保存储目录权限正确
func NewApp() *App {
	// 加载配置
	Config.LoadConfig()

	// 验证配置
	if err := Config.ValidateConfig(); err != nil {
		log.Fatal("配置验证失败:", err)
	}

	// 初始化存储管理器
	storageConfig := Config.GetStorageConfig()
	storageManager := Storage.NewStorageManager(storageConfig)

	// 初始化日志管理器服务（确保日志按目录分离）
	logManager := Services.NewLogManagerService(&Config.GetConfig().Log)

	// 记录日志系统初始化信息
	log.Printf("日志系统初始化完成，基础路径: %s", Config.GetConfig().Log.BasePath)
	log.Printf("请求日志: %v, SQL日志: %v, 错误日志: %v",
		Config.GetConfig().Log.RequestLog.Enabled,
		Config.GetConfig().Log.SQLLog.Enabled,
		Config.GetConfig().Log.ErrorLog.Enabled)

	// 初始化Redis服务
	var redisService *Services.RedisService
	redisConfig := Config.GetConfig().Redis
	if redisConfig.Host != "" {
		redisService = Services.NewRedisService(&Services.RedisConfig{
			Host:     redisConfig.Host,
			Port:     redisConfig.Port,
			Password: redisConfig.Password,
			DB:       redisConfig.Database,
		})

		// 测试Redis连接
		if err := redisService.Ping(); err != nil {
			log.Printf("警告: Redis连接失败: %v", err)
			redisService = nil
		} else {
			log.Println("✅ Redis服务连接成功")
		}
	}

	// 初始化缓存服务
	cacheService := Services.NewCacheService(redisService, &Services.CacheConfig{
		Prefix:      "app:",
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		EnableCache: redisService != nil,
	})

	// 使用LogManagerService初始化数据库（启用SQL日志）
	Database.InitDBWithLogManager(logManager)

	// 初始化应用
	app := &App{
		Config:         Config.GetConfig(),
		Router:         bootstrap.NewRouter(),
		StorageManager: storageManager,
		LogManager:     logManager,
	}

	// 注册路由
	Routes.RegisterRoutes(app.Router.Engine, storageManager, logManager)

	// 自动迁移数据库表
	Database.AutoMigrate()

	// 预热缓存（如果Redis可用）
	if redisService != nil {
		if err := cacheService.WarmCache(); err != nil {
			log.Printf("警告: 缓存预热失败: %v", err)
		}
	}

	// 记录应用启动日志到对应的日志类型中
	startupCtx := context.Background()

	// 记录系统启动信息到业务日志
	logManager.LogBusiness(startupCtx, "system", "startup", "应用启动成功", map[string]interface{}{
		"port":          app.Config.Server.Port,
		"mode":          app.Config.Server.Mode,
		"storage_path":  storageConfig.BasePath,
		"log_base_path": Config.GetConfig().Log.BasePath,
	})

	// 记录数据库信息到数据库日志
	logManager.LogSQL(startupCtx, "应用启动 - 数据库配置", 0, 0, nil, map[string]interface{}{
		"driver":       app.Config.Database.Driver,
		"host":         app.Config.Database.Host,
		"port":         app.Config.Database.Port,
		"database":     app.Config.Database.Database,
		"charset":      app.Config.Database.Charset,
		"startup_type": "application",
	})

	// 记录缓存信息到业务日志
	logManager.LogBusiness(startupCtx, "cache", "startup", "缓存服务初始化", map[string]interface{}{
		"redis_enabled": redisService != nil,
		"cache_enabled": cacheService != nil,
		"cache_type": func() string {
			if redisService != nil {
				return "redis"
			}
			return "memory"
		}(),
	})

	return app
}

// Run 启动应用
// 功能说明：
// 1. 创建HTTP服务器实例
// 2. 配置优雅关闭超时时间（30秒）
// 3. 监听系统信号（SIGINT、SIGTERM）
// 4. 启动服务器监听指定端口
// 5. 收到关闭信号时优雅关闭服务器
// 6. 记录服务器启动和关闭日志
func (app *App) Run() error {
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + app.Config.Server.Port,
		Handler: app.Router.Engine,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %s", app.Config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// 关闭数据库连接
	if err := Database.CloseDB(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	// 记录应用关闭日志到对应的日志类型中
	shutdownCtx := context.Background()

	// 记录系统关闭信息到业务日志
	app.LogManager.LogBusiness(shutdownCtx, "system", "shutdown", "应用已优雅关闭", map[string]interface{}{
		"shutdown_time": time.Now().Format(time.RFC3339),
	})

	// 记录数据库关闭信息到数据库日志
	app.LogManager.LogSQL(shutdownCtx, "应用关闭 - 数据库连接关闭", 0, 0, nil, map[string]interface{}{
		"shutdown_type": "application",
		"shutdown_time": time.Now().Format(time.RFC3339),
	})

	// 关闭日志管理器
	if err := app.LogManager.Close(); err != nil {
		log.Printf("Error closing log manager: %v", err)
	}

	log.Println("Server exited")
	return nil
}

// GetConfig 获取配置
func (app *App) GetConfig() *Config.Config {
	return app.Config
}

// GetRouter 获取路由
func (app *App) GetRouter() *bootstrap.Router {
	return app.Router
}

// GetLogManager 获取日志管理器
func (app *App) GetLogManager() *Services.LogManagerService {
	return app.LogManager
}
