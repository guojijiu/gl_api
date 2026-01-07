package Database

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database/Migrations"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// TestDB 测试数据库实例
var TestDB *gorm.DB

// InitTestDB 初始化测试数据库
//
// 功能说明：
// 1. 创建内存SQLite数据库（不依赖外部数据库）
// 2. 配置连接池参数（优化性能）
// 3. 运行数据库迁移（创建表结构）
// 4. 返回GORM数据库实例（用于测试）
//
// 数据库类型：
// - 使用内存SQLite数据库（:memory:）
// - 不需要外部数据库服务
// - 测试结束后数据自动清除
// - 支持外键约束和WAL模式
//
// 连接池配置：
// - MaxIdleConns: 10（最大空闲连接数）
// - MaxOpenConns: 100（最大打开连接数）
// - ConnMaxLifetime: 1小时（连接最大生存时间）
// - ConnMaxIdleTime: 10分钟（连接最大空闲时间）
//
// 数据库特性：
// - 外键约束：启用（_pragma=foreign_keys(1)）
// - 日志模式：WAL（Write-Ahead Logging，提高并发性能）
// - 缓存模式：共享缓存（cache=shared）
//
// 使用场景：
// - 单元测试：快速、隔离的测试环境
// - 集成测试：不需要外部数据库依赖
// - CI/CD：无需配置数据库服务
//
// 注意事项：
// - 数据库是单例模式，多次调用返回同一实例
// - 内存数据库在程序退出时自动清除
// - 迁移失败会导致程序退出（Fatal）
// - 测试时日志级别为Silent，减少输出
func InitTestDB() *gorm.DB {
	// 如果已初始化，直接返回
	// 单例模式，避免重复初始化
	if TestDB != nil {
		return TestDB
	}

	// 使用纯Go的SQLite驱动创建内存数据库
	// DSN参数说明：
	// - file::memory:: 内存数据库
	// - cache=shared: 共享缓存模式
	// - _pragma=foreign_keys(1): 启用外键约束
	// - _pragma=journal_mode(WAL): WAL日志模式
	dsn := "file::memory:?cache=shared&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: getTestGormLogger(), // 使用测试日志配置（静默模式）
	})
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// 设置连接池参数
	// 获取底层sql.DB实例以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)              // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)             // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)    // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 连接最大空闲时间

	// 运行数据库迁移
	// 创建所有必要的表结构
	migrationManager := Migrations.NewMigrationManager(db)
	if err := migrationManager.RunMigrations(); err != nil {
		log.Fatal("Failed to run test migrations:", err)
	}

	// 保存数据库实例
	TestDB = db
	log.Println("Test database initialized successfully")
	return db
}

// getTestGormLogger 获取测试用的GORM日志配置
func getTestGormLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent, // 测试时静默日志
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// CloseTestDB 关闭测试数据库
func CloseTestDB() error {
	if TestDB != nil {
		sqlDB, err := TestDB.DB()
		if err != nil {
			return err
		}
		err = sqlDB.Close()
		TestDB = nil
		return err
	}
	return nil
}

// GetTestDB 获取测试数据库实例
func GetTestDB() *gorm.DB {
	if TestDB == nil {
		return InitTestDB()
	}
	return TestDB
}

// ResetTestDB 重置测试数据库（清空所有数据）
//
// 功能说明：
// 1. 删除所有测试表（DROP TABLE IF EXISTS）
// 2. 重新运行数据库迁移（重建表结构）
// 3. 为测试提供全新的数据库环境
//
// 重置策略：
// - 使用DROP TABLE IF EXISTS删除表（如果存在）
// - 删除所有表后重新运行迁移
// - 确保表结构是最新的
//
// 删除的表：
// - users: 用户表
// - posts: 文章表
// - categories: 分类表
// - tags: 标签表
// - api_keys: API密钥表
// - audit_logs: 审计日志表
// - performance_metrics: 性能指标表
// - security_events: 安全事件表
//
// 使用场景：
// - 测试套件初始化：确保干净的数据库环境
// - 测试失败恢复：重置数据库状态
// - 数据库结构变更：重新创建表结构
//
// 注意事项：
// - 数据库必须已初始化（TestDB != nil）
// - 删除操作不可逆，所有数据会丢失
// - 删除表时会自动处理外键约束
// - 迁移失败会返回错误
func ResetTestDB() error {
	// 检查数据库是否已初始化
	if TestDB == nil {
		return fmt.Errorf("test database not initialized")
	}

	// 删除所有表
	// 使用DROP TABLE IF EXISTS，如果表不存在不会报错
	// 按顺序删除，避免外键约束问题
	tables := []string{
		"users",
		"posts",
		"categories",
		"tags",
		"api_keys",
		"audit_logs",
		"performance_metrics",
		"security_events",
	}

	for _, table := range tables {
		if err := TestDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			return fmt.Errorf("删除表 %s 失败: %v", table, err)
		}
	}

	// 重新运行迁移
	// 重建所有表结构，确保是最新的
	migrationManager := Migrations.NewMigrationManager(TestDB)
	return migrationManager.RunMigrations()
}

// MockDatabaseConfig 创建模拟数据库配置
func MockDatabaseConfig() *Config.DatabaseConfig {
	return &Config.DatabaseConfig{
		Driver:            "postgres",
		Host:              "localhost",
		Port:              "5432",
		Username:          "test_user",
		Password:          "test_password",
		Database:          "test_db",
		Charset:           "utf8",
		MaxOpenConns:      10,
		MaxIdleConns:      5,
		ConnMaxLifetime:   time.Hour,
		ConnMaxIdleTime:   10 * time.Minute,
		ConnectionTimeout: 30,
		ReadTimeout:       30,
		WriteTimeout:      30,
	}
}

// SetupTestEnvironment 设置测试环境
func SetupTestEnvironment() {
	// 设置测试环境变量
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_USERNAME", "test_user")
	os.Setenv("DATABASE_PASSWORD", "test_password")
	os.Setenv("DATABASE_NAME", "test_db")
}

// CleanupTestEnvironment 清理测试环境
func CleanupTestEnvironment() {
	// 清理环境变量
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("DATABASE_DRIVER")
	os.Unsetenv("DATABASE_HOST")
	os.Unsetenv("DATABASE_PORT")
	os.Unsetenv("DATABASE_USERNAME")
	os.Unsetenv("DATABASE_PASSWORD")
	os.Unsetenv("DATABASE_NAME")
}

// TestContext 创建测试上下文
//
// 功能说明：
// 1. 创建带超时的测试上下文（30秒）
// 2. 返回context和cancel函数
// 3. 防止测试无限运行
//
// 超时设置：
// - 超时时间：30秒
// - 超时后context会自动取消
// - 防止测试无限等待
//
// 返回信息：
// - context.Context: 带超时的上下文
// - context.CancelFunc: 取消函数（可以提前取消）
//
// 使用场景：
// - 数据库操作测试：设置操作超时
// - HTTP请求测试：设置请求超时
// - 并发测试：控制goroutine执行时间
//
// 注意事项：
// - 调用者应该在使用完context后调用cancel函数
// - 不调用cancel会导致context泄漏（go vet会警告）
// - 超时后context会自动取消，但最好显式调用cancel
// - 可以使用defer cancel()确保资源释放
//
// 示例用法：
//   ctx, cancel := TestContext()
//   defer cancel() // 确保资源释放
//   // 使用ctx进行测试
func TestContext() (context.Context, context.CancelFunc) {
	// 创建带超时的上下文
	// 30秒超时，防止测试无限运行
	return context.WithTimeout(context.Background(), 30*time.Second)
}
