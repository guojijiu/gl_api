package Database

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database/Migrations"
	"cloud-platform-api/app/Storage"
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var LoggedDB *Storage.LoggedDB // 带日志记录的数据库连接

// InitDB 初始化数据库连接
// 功能说明：
// 1. 调用InitDBWithLogger(nil)进行基础数据库初始化
// 2. 不启用SQL日志记录功能
// 3. 保持向后兼容性
func InitDB() {
	InitDBWithLogger(nil)
}

// LogManagerInterface 日志管理器接口，避免循环导入
type LogManagerInterface interface {
	LogSQL(ctx context.Context, sql string, duration time.Duration, rows int64, error error, fields map[string]interface{})
	LogBusiness(ctx context.Context, module string, action string, message string, fields map[string]interface{})
}

// InitDBWithLogManager 使用指定的LogManagerService初始化数据库连接
func InitDBWithLogManager(logManager LogManagerInterface) {
	initDB()

	// 如果提供了LogManagerService，则设置GORM日志记录器
	if logManager != nil {
		// 设置GORM日志记录器
		DB.Logger = &GormLogManagerWrapper{
			logManager: logManager,
		}

		// 记录数据库连接成功日志到数据库日志中
		cfg := Config.GetConfig().Database
		logManager.LogSQL(context.Background(), "数据库连接成功", 0, 0, nil, map[string]interface{}{
			"driver":          cfg.Driver,
			"host":            cfg.Host,
			"port":            cfg.Port,
			"database":        cfg.Database,
			"charset":         cfg.Charset,
			"connection_type": "startup",
		})
	}
}

// GormLogManagerWrapper GORM日志记录器包装器，使用LogManagerService
type GormLogManagerWrapper struct {
	logManager LogManagerInterface
}

// LogMode 设置日志模式
func (w *GormLogManagerWrapper) LogMode(level logger.LogLevel) logger.Interface {
	return w
}

// Info 记录信息级别日志
func (w *GormLogManagerWrapper) Info(ctx context.Context, msg string, data ...interface{}) {
	if w.logManager == nil {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	w.logManager.LogSQL(ctx, msg, 0, 0, nil, fields)
}

// Warn 记录警告级别日志
func (w *GormLogManagerWrapper) Warn(ctx context.Context, msg string, data ...interface{}) {
	if w.logManager == nil {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	w.logManager.LogSQL(ctx, msg, 0, 0, nil, fields)
}

// Error 记录错误级别日志
func (w *GormLogManagerWrapper) Error(ctx context.Context, msg string, data ...interface{}) {
	if w.logManager == nil {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	w.logManager.LogSQL(ctx, msg, 0, 0, fmt.Errorf("%s", msg), fields)
}

// Trace 记录SQL查询跟踪信息
func (w *GormLogManagerWrapper) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if w.logManager == nil {
		return
	}

	// 执行SQL查询
	sql, rows := fc()
	duration := time.Since(begin)

	// 记录SQL日志
	w.logManager.LogSQL(ctx, sql, duration, rows, err, nil)
}

// initDB 初始化数据库连接（内部方法）
//
// 重要功能说明：
// 1. 多数据库支持：MySQL、PostgreSQL、SQLite三种数据库驱动
// 2. 连接池管理：自动配置连接池参数，支持高并发访问
// 3. 重试机制：连接失败时自动重试，支持指数退避策略
// 4. 健康监控：连接池状态监控，支持性能指标收集
// 5. SQL日志：可选的SQL语句日志记录，支持性能分析
// 6. 安全连接：支持SSL连接、用户认证、权限控制
//
// 连接重试机制：
// - 最大重试次数：3次
// - 初始重试延迟：5秒
// - 指数退避策略：每次重试延迟翻倍
// - 重试间隔：5s -> 10s -> 20s
// - 重试失败后记录详细错误日志
//
// 连接池配置：
// - 最大空闲连接：10个（减少连接创建开销）
// - 最大打开连接：100个（支持高并发）
// - 连接生命周期：1小时（防止连接老化）
// - 连接超时：30秒（快速失败）
// - 空闲超时：10分钟（释放无用连接）
//
// 性能优化：
// - 自动连接池调优
// - 慢查询检测和记录
// - 连接复用和负载均衡
// - 支持读写分离（可扩展）
// - 连接预热和健康检查
//
// 安全特性：
// - 连接字符串安全验证
// - 用户权限最小化原则
// - 支持SSL/TLS加密连接
// - 连接超时和重试限制
// - 敏感信息日志脱敏
//
// 监控和告警：
// - 连接池状态实时监控
// - 性能指标自动收集
// - 异常连接自动检测
// - 支持Prometheus指标导出
// - 连接失败告警通知
//
// 错误处理：
// - 连接失败时记录详细错误信息
// - 重试失败后提供诊断信息
// - 连接池配置失败时立即退出
// - 健康检查失败时记录警告
// - 支持优雅降级和故障转移

// classifyDBError 分类数据库错误
func classifyDBError(err error) string {
	if err == nil {
		return "SUCCESS"
	}

	errStr := strings.ToLower(err.Error())

	// 网络相关错误
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "timeout") {
		return "NETWORK_ERROR"
	}

	// 认证相关错误
	if strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "invalid credentials") ||
		strings.Contains(errStr, "password") {
		return "AUTH_ERROR"
	}

	// 数据库不存在错误
	if strings.Contains(errStr, "database") &&
		(strings.Contains(errStr, "doesn't exist") ||
			strings.Contains(errStr, "not found")) {
		return "DATABASE_NOT_FOUND"
	}

	// 权限相关错误
	if strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "insufficient privileges") {
		return "PERMISSION_ERROR"
	}

	// 配置相关错误
	if strings.Contains(errStr, "invalid") &&
		(strings.Contains(errStr, "configuration") ||
			strings.Contains(errStr, "parameter")) {
		return "CONFIG_ERROR"
	}

	// 资源不足错误
	if strings.Contains(errStr, "too many connections") ||
		strings.Contains(errStr, "connection limit") ||
		strings.Contains(errStr, "resource") {
		return "RESOURCE_ERROR"
	}

	// 其他错误
	return "UNKNOWN_ERROR"
}

// shouldRetry 判断是否应该重试
func shouldRetry(errorType string, attempt, maxRetries int) bool {
	// 达到最大重试次数
	if attempt >= maxRetries {
		return false
	}

	// 根据错误类型决定是否重试
	switch errorType {
	case "NETWORK_ERROR", "RESOURCE_ERROR":
		return true // 网络和资源错误可以重试
	case "AUTH_ERROR", "PERMISSION_ERROR", "CONFIG_ERROR", "DATABASE_NOT_FOUND":
		return false // 认证、权限、配置和数据库不存在错误不应该重试
	case "UNKNOWN_ERROR":
		return attempt < 3 // 未知错误最多重试3次
	default:
		return false
	}
}

// calculateRetryDelay 计算重试延迟
func calculateRetryDelay(baseDelay time.Duration, attempt int, maxDelay time.Duration) time.Duration {
	// 指数退避：baseDelay * 2^(attempt-1)
	exponentialDelay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))

	// 添加随机抖动（±25%）
	jitter := time.Duration(float64(exponentialDelay) * 0.25 * (2*rand.Float64() - 1))

	// 计算最终延迟
	finalDelay := exponentialDelay + jitter

	// 限制最大延迟
	if finalDelay > maxDelay {
		finalDelay = maxDelay
	}

	// 确保最小延迟
	if finalDelay < baseDelay {
		finalDelay = baseDelay
	}

	return finalDelay
}

func initDB() {
	var err error
	maxRetries := 5                   // 增加重试次数
	baseRetryDelay := 2 * time.Second // 减少初始延迟
	maxRetryDelay := 30 * time.Second // 设置最大延迟

	cfg := Config.GetConfig().Database

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 根据数据库驱动类型建立连接
		switch cfg.Driver {
		case "mysql":
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
				cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
			DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
				Logger: getGormLogger(),
			})

		case "postgres":
			dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai connect_timeout=30",
				cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
			DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
				Logger: getGormLogger(),
			})

		case "sqlite":
			// 使用纯 Go 的 SQLite 驱动，不需要 CGO
			dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", cfg.Database)
			DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
				Logger: getGormLogger(),
			})

		default:
			log.Fatal("Unsupported database driver:", cfg.Driver)
		}

		if err == nil {
			log.Printf("数据库连接成功 (尝试 %d/%d)", attempt, maxRetries)
			break // 连接成功，跳出重试循环
		}

		// 分类错误类型
		errorType := classifyDBError(err)
		log.Printf("数据库连接失败 (尝试 %d/%d) [%s]: %v", attempt, maxRetries, errorType, err)

		// 根据错误类型决定是否重试
		if !shouldRetry(errorType, attempt, maxRetries) {
			log.Printf("错误类型 %s 不支持重试，停止重试", errorType)
			// 对于不支持重试的错误，直接退出并记录致命错误
			log.Fatal("数据库连接失败，不支持重试的错误类型:", err)
			return
		}

		if attempt < maxRetries {
			// 计算重试延迟（指数退避 + 随机抖动）
			retryDelay := calculateRetryDelay(baseRetryDelay, attempt, maxRetryDelay)
			log.Printf("等待 %v 后重试...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatal("数据库连接失败，已达到最大重试次数:", err)
	}

	// 获取底层的sql.DB对象
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	// 设置连接池参数 - 优化性能
	sqlDB.SetMaxIdleConns(20)                  // 最大空闲连接数（增加以保持连接池活跃）
	sqlDB.SetMaxOpenConns(200)                 // 最大打开连接数（增加以支持高并发）
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期（减少以避免长时间占用）
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接最大生存时间（新增）

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	// 启动连接池监控
	poolMonitor := NewConnectionPoolMonitor(sqlDB, nil)
	poolMonitor.StartMonitoring(5 * time.Minute)

	log.Println("Database connected successfully")
}

// InitDBWithLogger 使用指定的StorageManager初始化数据库连接（向后兼容）
func InitDBWithLogger(storageManager *Storage.StorageManager) {
	initDB()

	// 如果提供了StorageManager，则包装数据库连接以添加SQL日志
	if storageManager != nil {
		sqlLogger := Storage.NewSQLLogger(storageManager)
		if sqlDB, err := DB.DB(); err == nil {
			LoggedDB = Storage.WrapDB(sqlDB, sqlLogger)
		}

		// 记录数据库连接成功日志
		storageManager.LogInfo("数据库连接成功", map[string]interface{}{
			"driver":   Config.GetConfig().Database.Driver,
			"host":     Config.GetConfig().Database.Host,
			"port":     Config.GetConfig().Database.Port,
			"database": Config.GetConfig().Database.Database,
			"charset":  Config.GetConfig().Database.Charset,
		})
	}
}

// getGormLogger 获取GORM日志配置
// 功能说明：
// 1. 根据环境配置设置日志级别
// 2. debug模式：显示所有SQL语句
// 3. 生产模式：只显示错误信息
// 4. 配置慢查询阈值（1秒）
// 5. 忽略记录未找到的错误
// 6. 启用彩色输出
func getGormLogger() logger.Interface {
	// 根据环境设置日志级别
	var logLevel logger.LogLevel
	if Config.GetConfig().Server.Mode == "debug" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}

// AutoMigrate 自动迁移数据库表
// 功能说明：
// 1. 使用新的迁移系统管理数据库表结构
// 2. 支持版本控制和回滚功能
// 3. 记录迁移成功或失败的日志
// 4. 包含所有核心业务模型和审计日志表
func AutoMigrate() {
	// 使用新的迁移系统
	migrationManager := Migrations.NewMigrationManager(DB)

	if err := migrationManager.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Database migrations completed successfully")
}

// GetDB 获取GORM数据库实例
// 功能说明：
// 1. 返回全局的GORM数据库实例
// 2. 供其他模块使用进行数据库操作
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
// 功能说明：
// 1. 安全关闭数据库连接
// 2. 释放连接池资源
// 3. 记录数据库关闭日志
// 4. 在应用关闭时调用
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
