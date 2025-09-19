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
// 使用内存SQLite数据库进行测试，避免外部依赖
func InitTestDB() *gorm.DB {
	if TestDB != nil {
		return TestDB
	}

	// 使用纯Go的SQLite驱动创建内存数据库
	dsn := "file::memory:?cache=shared&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: getTestGormLogger(),
	})
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// 运行迁移
	migrationManager := Migrations.NewMigrationManager(db)
	if err := migrationManager.RunMigrations(); err != nil {
		log.Fatal("Failed to run test migrations:", err)
	}

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
func ResetTestDB() error {
	if TestDB == nil {
		return fmt.Errorf("test database not initialized")
	}

	// 删除所有表
	if err := TestDB.Exec("DROP TABLE IF EXISTS users").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS posts").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS categories").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS tags").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS api_keys").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS audit_logs").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS performance_metrics").Error; err != nil {
		return err
	}
	if err := TestDB.Exec("DROP TABLE IF EXISTS security_events").Error; err != nil {
		return err
	}

	// 重新运行迁移
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
func TestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return ctx
}
