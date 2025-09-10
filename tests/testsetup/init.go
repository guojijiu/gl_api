package testsetup

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Routes"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/bootstrap"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

var Router *bootstrap.Router

// TestSuite 测试套件基类
// 功能说明：
// 1. 提供测试环境的初始化和清理
// 2. 包含数据库连接和存储管理器
// 3. 提供通用的测试工具方法
// 4. 支持测试数据的准备和清理
// 5. 集成TokenHelper用于token相关测试
type TestSuite struct {
	suite.Suite
	storageManager *Storage.StorageManager
	tokenHelper    *TokenHelper
	ctx            context.Context
	cancel         context.CancelFunc
}

// Init 初始化测试环境
// 功能说明：
// 1. 设置测试环境变量
// 2. 初始化测试数据库
// 3. 创建测试存储管理器
// 4. 注册测试路由
// 5. 准备测试数据
func Init() {
	// 设置测试环境变量
	setTestEnvironment()

	// 加载配置
	Config.LoadConfig()

	// 初始化测试存储管理器
	storagePath := filepath.Join(".", "storage", "test")
	storageManager := Storage.NewStorageManager(storagePath)

	// 初始化测试数据库
	Database.InitDBWithLogger(storageManager)

	// 运行数据库迁移
	Database.AutoMigrate()

	// 创建路由引擎
	Router = bootstrap.NewRouter()

	// 注册路由
	Routes.RegisterRoutes(Router.Engine, storageManager, nil)

	// 准备测试数据
	prepareTestData()
}

// SetupSuite 测试套件初始化
// 功能说明：
// 1. 初始化测试环境
// 2. 设置测试数据库连接
// 3. 创建存储管理器
// 4. 准备测试数据
// 5. 初始化TokenHelper
func (ts *TestSuite) SetupSuite() {
	// 设置测试环境变量
	os.Setenv("SERVER_MODE", "test")
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DATABASE", ":memory:")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing-only-32-chars")

	// 初始化存储管理器
	storagePath := "./storage/test"
	ts.storageManager = Storage.NewStorageManager(storagePath)

	// 初始化数据库
	Database.InitDBWithLogger(ts.storageManager)

	// 运行数据库迁移
	Database.AutoMigrate()

	// 初始化TokenHelper
	ts.tokenHelper = NewTokenHelper()

	// 创建上下文
	ts.ctx, ts.cancel = context.WithTimeout(context.Background(), 30*time.Second)

	log.Println("测试套件初始化完成")
}

// TearDownSuite 测试套件清理
// 功能说明：
// 1. 清理测试数据
// 2. 关闭数据库连接
// 3. 清理临时文件
// 4. 释放资源
func (ts *TestSuite) TearDownSuite() {
	// 清理测试用户
	if ts.tokenHelper != nil {
		if err := ts.tokenHelper.CleanupTestUsers(); err != nil {
			log.Printf("清理测试用户失败: %v", err)
		}
	}

	// 关闭数据库连接
	if err := Database.CloseDB(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	// 取消上下文
	if ts.cancel != nil {
		ts.cancel()
	}

	// 清理测试存储目录
	if err := os.RemoveAll("./storage/test"); err != nil {
		log.Printf("清理测试存储目录失败: %v", err)
	}

	log.Println("测试套件清理完成")
}

// SetupTest 单个测试初始化
// 功能说明：
// 1. 为每个测试准备干净的环境
// 2. 清理测试数据
// 3. 重置测试状态
func (ts *TestSuite) SetupTest() {
	// 清理数据库表
	ts.cleanupDatabase()

	// 清理测试文件
	ts.cleanupTestFiles()
}

// TearDownTest 单个测试清理
// 功能说明：
// 1. 清理测试产生的数据
// 2. 重置测试状态
// 3. 准备下一个测试
func (ts *TestSuite) TearDownTest() {
	// 清理测试数据
	ts.cleanupDatabase()
}

// cleanupDatabase 清理数据库
func (ts *TestSuite) cleanupDatabase() {
	db := Database.GetDB()
	if db == nil {
		return
	}

	tables := []string{"users", "posts", "categories", "tags", "audit_logs"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			log.Printf("清理表 %s 失败: %v", table, err)
		}
	}
}

// cleanupTestFiles 清理测试文件
func (ts *TestSuite) cleanupTestFiles() {
	testDirs := []string{
		"./storage/test/app/public",
		"./storage/test/app/private",
		"./storage/test/temp",
		"./storage/test/logs",
	}

	for _, dir := range testDirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("清理测试目录 %s 失败: %v", dir, err)
		}
	}
}

// CreateTestUserWithToken 创建测试用户并返回token
// 功能说明：
// 1. 创建测试用户
// 2. 生成有效的JWT token
// 3. 返回用户信息和token
func (ts *TestSuite) CreateTestUserWithToken(username, email, password, role string) (*UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateTestUserWithToken(username, email, password, role)
	if err != nil {
		return nil, err
	}

	return &UserTokenInfo{
		User:  user,
		Token: token,
	}, nil
}

// CreateAdminUserWithToken 创建管理员用户并返回token
func (ts *TestSuite) CreateAdminUserWithToken() (*UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateAdminUserWithToken()
	if err != nil {
		return nil, err
	}

	return &UserTokenInfo{
		User:  user,
		Token: token,
	}, nil
}

// CreateNormalUserWithToken 创建普通用户并返回token
func (ts *TestSuite) CreateNormalUserWithToken() (*UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateNormalUserWithToken()
	if err != nil {
		return nil, err
	}

	return &UserTokenInfo{
		User:  user,
		Token: token,
	}, nil
}

// GetAdminTokenHeaders 获取管理员token请求头
func (ts *TestSuite) GetAdminTokenHeaders() (map[string]string, error) {
	return ts.tokenHelper.GetAdminTokenHeaders()
}

// GetUserTokenHeaders 获取普通用户token请求头
func (ts *TestSuite) GetUserTokenHeaders() (map[string]string, error) {
	return ts.tokenHelper.GetUserTokenHeaders()
}

// GetTestTokenHeaders 获取测试用的请求头
func (ts *TestSuite) GetTestTokenHeaders(token string) map[string]string {
	return ts.tokenHelper.GetTestTokenHeaders(token)
}

// ValidateToken 验证token有效性
func (ts *TestSuite) ValidateToken(token string) error {
	_, err := ts.tokenHelper.ValidateToken(token)
	return err
}

// AssertTokenValid 断言token有效
func (ts *TestSuite) AssertTokenValid(token string) {
	err := ts.ValidateToken(token)
	ts.Require().NoError(err, "Token应该有效")
}

// AssertTokenInvalid 断言token无效
func (ts *TestSuite) AssertTokenInvalid(token string) {
	err := ts.ValidateToken(token)
	ts.Require().Error(err, "Token应该无效")
}

// setTestEnvironment 设置测试环境变量
func setTestEnvironment() {
	envVars := map[string]string{
		"SERVER_MODE":      "test",
		"DB_DRIVER":        "sqlite",
		"DB_DATABASE":      ":memory:",
		"JWT_SECRET":       "test-jwt-secret-key-for-testing-only-32-chars",
		"REDIS_HOST":       "",
		"LOG_LEVEL":        "error",
		"TEST_ENVIRONMENT": "true",
	}

	for key, value := range envVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// prepareTestData 准备测试数据
func prepareTestData() {
	// 这里可以添加一些基础的测试数据
	// 例如：创建测试用户、测试分类等
}

// Cleanup 清理测试环境
func Cleanup() {
	// 清理测试数据库
	if err := Database.CloseDB(); err != nil {
		// 记录错误但不中断测试
	}

	// 清理测试存储目录
	testStoragePath := filepath.Join(".", "storage", "test")
	if err := os.RemoveAll(testStoragePath); err != nil {
		// 记录错误但不中断测试
	}
}

// GetTestStoragePath 获取测试存储路径
func GetTestStoragePath() string {
	return filepath.Join(".", "storage", "test")
}

// GetTestDatabase 获取测试数据库实例
func GetTestDatabase() *gorm.DB {
	return Database.GetDB()
}
