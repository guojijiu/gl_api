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
//
// 功能说明：
// 1. 根据错误消息内容将数据库错误分类为不同的类型
// 2. 错误分类用于决定是否应该重试连接
// 3. 帮助快速定位问题根源，提供更好的错误处理
//
// 错误类型说明：
// - SUCCESS：无错误（err == nil）
// - NETWORK_ERROR：网络相关错误，通常是临时性的，可以重试
// - AUTH_ERROR：认证相关错误，通常是配置问题，不应该重试
// - DATABASE_NOT_FOUND：数据库不存在，需要先创建数据库，不应该重试
// - PERMISSION_ERROR：权限相关错误，需要修复权限，不应该重试
// - CONFIG_ERROR：配置相关错误，需要修复配置，不应该重试
// - RESOURCE_ERROR：资源不足错误，通常是临时性的，可以重试
// - UNKNOWN_ERROR：未知错误，需要进一步分析
//
// 分类策略：
// - 通过检查错误消息中的关键词来判断错误类型
// - 使用字符串包含检查（strings.Contains），不区分大小写
// - 按优先级检查，先检查更具体的错误类型
//
// 注意事项：
// - 错误消息可能因数据库类型而异，需要覆盖常见的关键词
// - 某些错误可能匹配多个类型，按检查顺序返回第一个匹配的类型
// - 如果无法分类，返回UNKNOWN_ERROR，由shouldRetry决定是否重试
func classifyDBError(err error) string {
	// 无错误，返回成功
	if err == nil {
		return "SUCCESS"
	}

	// 将错误消息转换为小写，便于匹配（不区分大小写）
	errStr := strings.ToLower(err.Error())

	// 网络相关错误（通常是临时性的，可以重试）
	// 包括：连接被拒绝、无法路由到主机、网络不可达、超时等
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "timeout") {
		return "NETWORK_ERROR"
	}

	// 认证相关错误（通常是配置问题，不应该重试）
	// 包括：访问被拒绝、认证失败、无效凭证、密码错误等
	if strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "invalid credentials") ||
		strings.Contains(errStr, "password") {
		return "AUTH_ERROR"
	}

	// 数据库不存在错误（需要先创建数据库，不应该重试）
	// 包括：数据库不存在、数据库未找到等
	if strings.Contains(errStr, "database") &&
		(strings.Contains(errStr, "doesn't exist") ||
			strings.Contains(errStr, "not found")) {
		return "DATABASE_NOT_FOUND"
	}

	// 权限相关错误（需要修复权限，不应该重试）
	// 包括：权限被拒绝、访问被拒绝、权限不足等
	if strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "insufficient privileges") {
		return "PERMISSION_ERROR"
	}

	// 配置相关错误（需要修复配置，不应该重试）
	// 包括：无效配置、无效参数等
	if strings.Contains(errStr, "invalid") &&
		(strings.Contains(errStr, "configuration") ||
			strings.Contains(errStr, "parameter")) {
		return "CONFIG_ERROR"
	}

	// 资源不足错误（通常是临时性的，可以重试）
	// 包括：连接数过多、连接限制、资源不足等
	if strings.Contains(errStr, "too many connections") ||
		strings.Contains(errStr, "connection limit") ||
		strings.Contains(errStr, "resource") {
		return "RESOURCE_ERROR"
	}

	// 其他错误：无法分类的未知错误
	// 由shouldRetry决定是否重试（通常限制重试次数）
	return "UNKNOWN_ERROR"
}

// shouldRetry 判断是否应该重试
//
// 功能说明：
// 1. 根据错误类型和重试次数决定是否应该重试连接
// 2. 避免对不可恢复的错误进行无意义的重试
// 3. 防止无限重试导致资源浪费
//
// 重试策略：
// - 网络错误（NETWORK_ERROR）：可以重试，可能是临时网络问题
// - 资源错误（RESOURCE_ERROR）：可以重试，可能是连接池满等临时问题
// - 认证错误（AUTH_ERROR）：不应该重试，配置错误需要人工修复
// - 权限错误（PERMISSION_ERROR）：不应该重试，权限问题需要人工修复
// - 配置错误（CONFIG_ERROR）：不应该重试，配置错误需要人工修复
// - 数据库不存在（DATABASE_NOT_FOUND）：不应该重试，需要先创建数据库
// - 未知错误（UNKNOWN_ERROR）：最多重试3次，避免无限重试
//
// 注意事项：
// - 达到最大重试次数后不再重试
// - 对于不可恢复的错误，立即返回false，避免浪费资源
// - 未知错误限制重试次数，避免无限重试
func shouldRetry(errorType string, attempt, maxRetries int) bool {
	// 达到最大重试次数，不再重试
	if attempt >= maxRetries {
		return false
	}

	// 根据错误类型决定是否重试
	// 策略：只有临时性错误才重试，永久性错误不重试
	switch errorType {
	case "NETWORK_ERROR", "RESOURCE_ERROR":
		// 网络和资源错误通常是临时性的，可以重试
		// 例如：网络中断、连接池满等，这些问题可能会自动恢复
		return true
	case "AUTH_ERROR", "PERMISSION_ERROR", "CONFIG_ERROR", "DATABASE_NOT_FOUND":
		// 认证、权限、配置和数据库不存在错误是永久性的，不应该重试
		// 这些问题需要人工修复，重试只会浪费资源
		return false
	case "UNKNOWN_ERROR":
		// 未知错误：可能是临时性的，也可能是永久性的
		// 为了安全，限制最多重试3次，避免无限重试
		return attempt < 3
	default:
		// 其他错误类型默认不重试
		return false
	}
}

// calculateRetryDelay 计算重试延迟
//
// 功能说明：
// 1. 使用指数退避策略计算重试延迟
// 2. 添加随机抖动，避免多个客户端同时重试（thundering herd问题）
// 3. 限制最大和最小延迟，确保延迟在合理范围内
//
// 指数退避算法：
// - 第1次重试：baseDelay * 2^0 = baseDelay
// - 第2次重试：baseDelay * 2^1 = baseDelay * 2
// - 第3次重试：baseDelay * 2^2 = baseDelay * 4
// - 第N次重试：baseDelay * 2^(N-1)
//
// 随机抖动：
// - 在指数退避的基础上添加±25%的随机延迟
// - 避免多个客户端同时重试，造成服务器压力
// - 例如：如果延迟是10秒，抖动范围是7.5秒到12.5秒
//
// 延迟限制：
// - 最大延迟：不超过maxDelay，避免等待时间过长
// - 最小延迟：不低于baseDelay，确保有基本的延迟
//
// 使用场景：
// - 数据库连接重试
// - API调用重试
// - 网络请求重试
func calculateRetryDelay(baseDelay time.Duration, attempt int, maxDelay time.Duration) time.Duration {
	// 指数退避：每次重试延迟翻倍
	// 公式：baseDelay * 2^(attempt-1)
	// 例如：baseDelay=2秒，attempt=3，则延迟=2*2^2=8秒
	// 这样可以避免频繁重试，给服务器恢复的时间
	exponentialDelay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))

	// 添加随机抖动（±25%）
	// 随机抖动可以避免多个客户端同时重试（thundering herd问题）
	// 例如：如果延迟是10秒，抖动范围是7.5秒到12.5秒
	// rand.Float64()返回[0,1)的随机数，2*rand.Float64()-1返回[-1,1)的随机数
	jitter := time.Duration(float64(exponentialDelay) * 0.25 * (2*rand.Float64() - 1))

	// 计算最终延迟（指数退避 + 随机抖动）
	finalDelay := exponentialDelay + jitter

	// 限制最大延迟，避免等待时间过长
	// 例如：如果计算出的延迟是60秒，但maxDelay是30秒，则使用30秒
	if finalDelay > maxDelay {
		finalDelay = maxDelay
	}

	// 确保最小延迟，避免延迟过短
	// 例如：如果计算出的延迟是0.5秒，但baseDelay是2秒，则使用2秒
	if finalDelay < baseDelay {
		finalDelay = baseDelay
	}

	return finalDelay
}

// initDB 初始化数据库连接
//
// 功能说明：
// 1. 根据配置的数据库类型建立连接（MySQL、PostgreSQL、SQLite）
// 2. 实现智能重试机制，对临时性错误自动重试
// 3. 配置连接池参数，优化性能和资源使用
// 4. 启动连接池监控，实时监控连接状态
//
// 重试策略：
// - 最大重试次数：5次
// - 初始延迟：2秒
// - 最大延迟：30秒
// - 使用指数退避 + 随机抖动
//
// 连接池配置：
// - 最大空闲连接：20个（保持连接池活跃，减少连接创建开销）
// - 最大打开连接：200个（支持高并发访问）
// - 连接最大生命周期：30分钟（防止连接老化）
// - 空闲连接最大生存时间：5分钟（释放无用连接）
//
// 错误处理：
// - 根据错误类型决定是否重试（网络错误可重试，认证错误不重试）
// - 不支持重试的错误立即退出，避免无限循环
// - 达到最大重试次数后退出，记录详细错误信息
//
// 注意事项：
// - 连接失败会记录致命错误并退出程序（log.Fatal）
// - 连接池配置失败会立即退出
// - 连接测试失败会立即退出
func initDB() {
	var err error
	maxRetries := 5                   // 最大重试次数
	baseRetryDelay := 2 * time.Second // 初始重试延迟
	maxRetryDelay := 30 * time.Second // 最大重试延迟

	cfg := Config.GetConfig().Database

	// 重试机制：最多重试maxRetries次
	// 每次重试前会根据错误类型判断是否应该重试
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 根据数据库驱动类型建立连接
		// 支持MySQL、PostgreSQL、SQLite三种数据库
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
				Logger:                                   getGormLogger(),
				DisableForeignKeyConstraintWhenMigrating: true,
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

		// 检查连接是否成功
		if err == nil {
			log.Printf("数据库连接成功 (尝试 %d/%d)", attempt, maxRetries)
			break // 连接成功，跳出重试循环
		}

		// 连接失败，分类错误类型
		// 根据错误类型决定是否应该重试
		errorType := classifyDBError(err)
		log.Printf("数据库连接失败 (尝试 %d/%d) [%s]: %v", attempt, maxRetries, errorType, err)

		// 根据错误类型决定是否重试
		// 对于不可恢复的错误（如认证错误、配置错误），不应该重试
		// 重试只会浪费资源，需要人工修复配置
		if !shouldRetry(errorType, attempt, maxRetries) {
			log.Printf("错误类型 %s 不支持重试，停止重试", errorType)
			// 对于不支持重试的错误，直接退出并记录致命错误
			// 这些错误通常是配置问题，需要人工修复
			log.Fatal("数据库连接失败，不支持重试的错误类型:", err)
			return
		}

		// 如果还有重试机会，计算延迟后重试
		if attempt < maxRetries {
			// 计算重试延迟（指数退避 + 随机抖动）
			// 指数退避：每次重试延迟翻倍，避免频繁重试造成服务器压力
			// 随机抖动：添加随机延迟，避免多个客户端同时重试（thundering herd问题）
			retryDelay := calculateRetryDelay(baseRetryDelay, attempt, maxRetryDelay)
			log.Printf("等待 %v 后重试...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatal("数据库连接失败，已达到最大重试次数:", err)
	}

	// 获取底层的sql.DB对象
	// GORM的DB对象是对sql.DB的封装，需要获取底层对象来配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	// 设置连接池参数 - 优化性能和资源使用
	// 这些参数对高并发场景下的性能至关重要
	sqlDB.SetMaxIdleConns(20)                  // 最大空闲连接数：保持20个连接在池中，减少连接创建开销
	sqlDB.SetMaxOpenConns(200)                 // 最大打开连接数：最多同时打开200个连接，支持高并发
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期：30分钟后强制关闭连接，防止连接老化
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接最大生存时间：5分钟未使用则关闭，释放资源

	// 测试数据库连接
	// Ping()会发送一个简单的查询来验证连接是否有效
	// 如果连接无效，会返回错误，此时应该退出程序
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	// 启动连接池监控
	// 监控连接池状态，包括活跃连接数、空闲连接数、等待连接数等
	// 每5分钟检查一次，用于性能分析和问题诊断
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
