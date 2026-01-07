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
	TokenHelper    *TokenHelper
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
	storageConfig := &Config.StorageConfig{
		BasePath: storagePath,
	}
	storageManager := Storage.NewStorageManager(storageConfig)

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
//
// 功能说明：
// 1. 初始化测试环境（设置环境变量、加载配置）
// 2. 设置测试数据库连接（使用测试数据库）
// 3. 创建存储管理器（使用测试存储路径）
// 4. 准备测试数据（运行数据库迁移）
// 5. 初始化TokenHelper（用于token相关测试）
//
// 测试环境配置：
// - 使用独立的测试数据库（不污染生产数据）
// - 使用独立的存储路径（不污染生产文件）
// - 使用测试专用的JWT密钥
// - 设置测试模式（减少日志输出）
//
// 初始化流程：
// 1. 设置测试环境变量
// 2. 加载配置（从环境变量）
// 3. 初始化存储管理器
// 4. 初始化数据库连接
// 5. 运行数据库迁移（创建表结构）
// 6. 初始化TokenHelper
// 7. 创建测试上下文
//
// 资源管理：
// - 测试数据库：使用独立的测试数据库
// - 存储路径：使用./storage/test目录
// - 上下文：创建带超时的上下文，防止测试无限运行
//
// 注意事项：
// - 测试环境变量会覆盖生产配置
// - 测试数据库应该在测试后清理
// - 测试文件应该在测试后删除
// - 上下文应该在测试后取消，避免资源泄漏
func (ts *TestSuite) SetupSuite() {
	// 设置测试环境变量
	// 这些变量会覆盖生产配置，确保测试环境独立
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("SERVER_MODE", "test")
	os.Setenv("DATABASE_DRIVER", "mysql")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "3306")
	os.Setenv("DATABASE_USERNAME", "root")
	os.Setenv("DATABASE_PASSWORD", "password")
	os.Setenv("DATABASE_NAME", "test_db")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing-only-32-chars")

	// 加载配置
	// 从环境变量加载配置，覆盖默认值
	Config.LoadConfig()

	// 初始化存储管理器
	// 使用独立的测试存储路径，不污染生产文件
	storagePath := "./storage/test"
	storageConfig := &Config.StorageConfig{
		BasePath: storagePath,
	}
	ts.storageManager = Storage.NewStorageManager(storageConfig)

	// 初始化数据库
	// 使用测试数据库，不污染生产数据
	// 启用SQL日志记录，便于调试
	Database.InitDBWithLogger(ts.storageManager)

	// 运行数据库迁移
	// 创建所有必要的表结构
	// 确保测试环境与生产环境一致
	Database.AutoMigrate()

	// 初始化TokenHelper
	// 用于创建测试用户和生成测试token
	// 简化token相关测试的编写
	ts.TokenHelper = NewTokenHelper()

	// 创建测试上下文
	// 带超时（30秒），防止测试无限运行
	// 超时后自动取消，避免资源泄漏
	ts.ctx, ts.cancel = context.WithTimeout(context.Background(), 30*time.Second)

	log.Println("测试套件初始化完成")
}

// TearDownSuite 测试套件清理
//
// 功能说明：
// 1. 清理测试数据（测试用户、测试数据等）
// 2. 关闭数据库连接（释放连接池资源）
// 3. 清理临时文件（删除测试存储目录）
// 4. 释放资源（取消上下文、关闭服务等）
//
// 清理流程：
// 1. 清理测试用户（通过TokenHelper）
// 2. 关闭数据库连接（释放连接池）
// 3. 取消上下文（释放goroutine资源）
// 4. 清理测试存储目录（删除测试文件）
//
// 资源释放：
// - 数据库连接：关闭连接池，释放所有连接
// - 上下文：取消上下文，停止所有相关goroutine
// - 文件系统：删除测试目录，释放磁盘空间
// - 内存：清理测试数据，释放内存
//
// 错误处理：
// - 清理失败时记录错误但不中断流程
// - 确保所有资源都被尝试清理
// - 某些资源可能已经被清理，需要检查nil
//
// 注意事项：
// - 清理顺序很重要，应该先清理依赖资源
// - 清理失败不应该导致panic，只记录错误
// - 某些资源（如数据库）可能在其他地方已关闭
// - 测试存储目录可能包含重要文件，需要谨慎删除
func (ts *TestSuite) TearDownSuite() {
	// 清理测试用户
	// 删除所有通过TokenHelper创建的测试用户
	// 防止测试数据污染数据库
	if ts.TokenHelper != nil {
		if err := ts.TokenHelper.CleanupTestUsers(); err != nil {
			log.Printf("清理测试用户失败: %v", err)
		}
	}

	// 关闭数据库连接
	// 释放连接池资源，关闭所有数据库连接
	// 如果连接已经关闭，不会报错
	if err := Database.CloseDB(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	// 取消上下文
	// 停止所有使用此上下文的goroutine
	// 释放goroutine资源，避免goroutine泄漏
	if ts.cancel != nil {
		ts.cancel()
	}

	// 清理测试存储目录
	// 删除所有测试文件，释放磁盘空间
	// 注意：这会删除整个测试存储目录
	if err := os.RemoveAll("./storage/test"); err != nil {
		log.Printf("清理测试存储目录失败: %v", err)
	}

	log.Println("测试套件清理完成")
}

// SetupTest 单个测试初始化
//
// 功能说明：
// 1. 为每个测试准备干净的环境
// 2. 清理上一个测试的数据
// 3. 重置测试状态
// 4. 确保测试之间的隔离性
//
// 测试隔离：
// - 每个测试都应该在干净的环境中运行
// - 清理上一个测试的数据，避免测试之间的相互影响
// - 重置所有状态，确保测试的可重复性
//
// 清理内容：
// - 数据库表：清空所有测试数据
// - 测试文件：删除所有测试文件
// - 缓存：清理所有缓存数据
//
// 执行时机：
// - 在每个测试方法执行前自动调用
// - 由testify/suite框架自动管理
// - 确保每个测试都有干净的环境
//
// 注意事项：
// - 清理操作应该快速，避免影响测试性能
// - 清理失败不应该导致测试失败
// - 某些数据可能需要保留（如基础配置数据）
func (ts *TestSuite) SetupTest() {
	// 清理数据库表
	// 删除所有测试数据，确保每个测试从干净状态开始
	// 使用TRUNCATE或DELETE清空表，比DROP TABLE更快
	ts.cleanupDatabase()

	// 清理测试文件
	// 删除所有测试文件，释放磁盘空间
	// 确保文件系统状态干净
	ts.cleanupTestFiles()
}

// TearDownTest 单个测试清理
//
// 功能说明：
// 1. 清理当前测试产生的数据
// 2. 重置测试状态
// 3. 准备下一个测试的环境
//
// 清理目的：
// - 防止测试数据影响后续测试
// - 确保测试之间的隔离性
// - 释放测试占用的资源
//
// 执行时机：
// - 在每个测试方法执行后自动调用
// - 由testify/suite框架自动管理
// - 即使测试失败也会执行
//
// 清理内容：
// - 数据库数据：删除测试创建的数据
// - 临时文件：删除测试创建的文件
// - 缓存数据：清理测试使用的缓存
//
// 注意事项：
// - 清理操作应该快速，避免影响测试性能
// - 清理失败不应该导致测试失败
// - 某些资源（如数据库连接）不应该在这里关闭
func (ts *TestSuite) TearDownTest() {
	// 清理测试数据
	// 删除当前测试创建的所有数据
	// 确保下一个测试从干净状态开始
	ts.cleanupDatabase()
}

// cleanupDatabase 清理数据库
//
// 功能说明：
// 1. 清空所有测试数据表（删除所有记录）
// 2. 为下一个测试准备干净的环境
// 3. 确保测试之间的数据隔离
//
// 清理策略：
// - 使用DELETE FROM清空表（比DROP TABLE更快）
// - 保留表结构，只删除数据
// - 清理失败时记录错误但不中断流程
//
// 清理的表：
// - users: 用户表
// - posts: 文章表
// - categories: 分类表
// - tags: 标签表
// - audit_logs: 审计日志表
//
// 使用场景：
// - 每个测试前清理（SetupTest）
// - 每个测试后清理（TearDownTest）
// - 确保测试数据不相互影响
//
// 注意事项：
// - 只清理测试数据，不删除表结构
// - 清理失败不会导致测试失败
// - 如果数据库连接不存在，直接返回
// - 某些表可能有外键约束，需要按顺序清理
func (ts *TestSuite) cleanupDatabase() {
	// 获取数据库连接
	// 如果连接不存在，直接返回
	db := Database.GetDB()
	if db == nil {
		return
	}

	// 定义需要清理的表
	// 按顺序清理，避免外键约束问题
	tables := []string{"users", "posts", "categories", "tags", "audit_logs"}

	// 循环清理每个表
	// 使用DELETE FROM清空表，保留表结构
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			// 清理失败时记录错误，但不中断流程
			// 某些表可能不存在或已被清理
			log.Printf("清理表 %s 失败: %v", table, err)
		}
	}
}

// cleanupTestFiles 清理测试文件
//
// 功能说明：
// 1. 删除所有测试文件目录（包括子目录和文件）
// 2. 为下一个测试准备干净的文件系统环境
// 3. 确保测试之间的文件隔离
//
// 清理策略：
// - 使用RemoveAll删除整个目录（包括所有子目录和文件）
// - 清理失败时记录错误但不中断流程
// - 如果目录不存在，不会报错
//
// 清理的目录：
// - ./storage/test/app/public: 公共文件目录
// - ./storage/test/app/private: 私有文件目录
// - ./storage/test/temp: 临时文件目录
// - ./storage/test/logs: 日志文件目录
//
// 使用场景：
// - 每个测试前清理（SetupTest）
// - 确保测试文件不相互影响
// - 释放磁盘空间
//
// 注意事项：
// - 删除操作不可逆，请确保是测试目录
// - 清理失败不会导致测试失败
// - 如果目录不存在，RemoveAll不会报错
// - 某些文件可能正在被使用，清理可能失败
func (ts *TestSuite) cleanupTestFiles() {
	// 定义需要清理的测试目录
	// 包括公共文件、私有文件、临时文件和日志文件
	testDirs := []string{
		"./storage/test/app/public",  // 公共文件目录
		"./storage/test/app/private", // 私有文件目录
		"./storage/test/temp",        // 临时文件目录
		"./storage/test/logs",        // 日志文件目录
	}

	// 循环清理每个目录
	// 使用RemoveAll删除整个目录及其内容
	for _, dir := range testDirs {
		if err := os.RemoveAll(dir); err != nil {
			// 清理失败时记录错误，但不中断流程
			// 某些目录可能不存在或正在被使用
			log.Printf("清理测试目录 %s 失败: %v", dir, err)
		}
	}
}

// CreateTestUserWithToken 创建测试用户并返回token
//
// 功能说明：
// 1. 创建测试用户（保存到数据库）
// 2. 生成有效的JWT token（包含用户信息）
// 3. 返回用户信息和token（便于测试使用）
//
// 使用场景：
// - 需要认证的API测试
// - 权限相关的测试
// - 用户相关的业务逻辑测试
//
// 返回信息：
// - User: 创建的用户对象（包含ID、用户名、邮箱、角色等）
// - Token: JWT token字符串（用于API请求认证）
//
// 注意事项：
// - 创建的测试用户应该在测试后清理
// - Token的有效期由JWT配置决定
// - 用户名和邮箱必须唯一，重复创建会失败
// - 密码会被哈希存储，不会返回明文
func (ts *TestSuite) CreateTestUserWithToken(username, email, password, role string) (*UserTokenInfo, error) {
	// 通过TokenHelper创建测试用户和token
	// TokenHelper封装了用户创建和token生成的逻辑
	user, token, err := ts.TokenHelper.CreateTestUserWithToken(username, email, password, role)
	if err != nil {
		return nil, err
	}

	// 返回用户信息和token
	// 便于测试代码直接使用
	return &UserTokenInfo{
		User:  user,  // 用户对象
		Token: token, // JWT token
	}, nil
}

// CreateAdminUserWithToken 创建管理员用户并返回token
func (ts *TestSuite) CreateAdminUserWithToken() (*UserTokenInfo, error) {
	user, token, err := ts.TokenHelper.CreateAdminUserWithToken()
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
	user, token, err := ts.TokenHelper.CreateNormalUserWithToken()
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
	return ts.TokenHelper.GetAdminTokenHeaders()
}

// GetUserTokenHeaders 获取普通用户token请求头
func (ts *TestSuite) GetUserTokenHeaders() (map[string]string, error) {
	return ts.TokenHelper.GetUserTokenHeaders()
}

// GetTestTokenHeaders 获取测试用的请求头
func (ts *TestSuite) GetTestTokenHeaders(token string) map[string]string {
	return ts.TokenHelper.GetTestTokenHeaders(token)
}

// ValidateToken 验证token有效性
func (ts *TestSuite) ValidateToken(token string) error {
	_, err := ts.TokenHelper.ValidateToken(token)
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
		"DB_DRIVER":        "mysql",
		"DB_HOST":          "localhost",
		"DB_PORT":          "3306",
		"DB_USERNAME":      "root",
		"DB_PASSWORD":      "password",
		"DB_DATABASE":      "test_db",
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
// prepareTestData 准备测试数据
//
// 功能说明：
// 1. 创建基础的测试数据（如测试用户、测试分类等）
// 2. 为测试提供初始数据
// 3. 确保测试有可用的数据
//
// 使用场景：
// - 测试套件初始化：创建基础测试数据
// - 集成测试：需要真实数据的测试
// - 性能测试：需要大量数据的测试
//
// 注意事项：
// - 目前为空实现，可以根据需要添加
// - 创建的数据应该在测试后清理
// - 数据应该符合测试需求，避免过多
func prepareTestData() {
	// 这里可以添加一些基础的测试数据
	// 例如：创建测试用户、测试分类等
	// 目前为空实现，可以根据测试需求添加
}

// Cleanup 清理测试环境
//
// 功能说明：
// 1. 关闭测试数据库连接（释放连接池资源）
// 2. 删除测试存储目录（清理测试文件）
// 3. 释放测试占用的资源
//
// 清理内容：
// - 数据库连接：关闭连接池，释放所有连接
// - 存储目录：删除整个测试存储目录（./storage/test）
//
// 使用场景：
// - 测试套件清理：在所有测试完成后清理
// - 测试失败恢复：清理残留的测试数据
// - CI/CD环境：确保测试环境干净
//
// 注意事项：
// - 清理失败不会中断流程，只记录错误
// - 某些资源可能已经被清理，不会报错
// - 删除存储目录会删除所有测试文件
// - 建议在测试套件的TearDownSuite中调用
func Cleanup() {
	// 清理测试数据库
	// 关闭连接池，释放所有数据库连接
	if err := Database.CloseDB(); err != nil {
		// 记录错误但不中断测试
		// 某些情况下数据库可能已经关闭
	}

	// 清理测试存储目录
	// 删除整个测试存储目录，包括所有子目录和文件
	testStoragePath := filepath.Join(".", "storage", "test")
	if err := os.RemoveAll(testStoragePath); err != nil {
		// 记录错误但不中断测试
		// 某些文件可能正在被使用或已被删除
	}
}

// GetTestStoragePath 获取测试存储路径
//
// 功能说明：
// 1. 返回测试存储的基础路径
// 2. 统一管理测试文件路径
// 3. 便于测试代码使用
//
// 返回路径：
// - ./storage/test（相对于项目根目录）
//
// 使用场景：
// - 测试文件操作：需要知道测试文件路径
// - 测试数据准备：创建测试文件
// - 测试清理：删除测试文件
//
// 注意事项：
// - 路径是相对于项目根目录的
// - 路径可能不存在，需要先创建
// - 测试后应该清理此目录
func GetTestStoragePath() string {
	// 返回测试存储路径
	// 使用filepath.Join确保跨平台兼容性
	return filepath.Join(".", "storage", "test")
}

// GetTestDatabase 获取测试数据库实例
//
// 功能说明：
// 1. 返回测试数据库的GORM实例
// 2. 提供统一的数据库访问接口
// 3. 便于测试代码使用
//
// 返回信息：
// - *gorm.DB: GORM数据库实例（用于数据库操作）
//
// 使用场景：
// - 数据库操作测试：执行数据库查询和操作
// - 数据准备：创建测试数据
// - 数据验证：验证数据库状态
//
// 注意事项：
// - 数据库必须已初始化（通过Init或SetupSuite）
// - 返回的实例是全局单例
// - 测试后应该清理测试数据
func GetTestDatabase() *gorm.DB {
	// 返回全局数据库实例
	// 通过Database.GetDB()获取已初始化的数据库
	return Database.GetDB()
}

// AssertResponseSuccess 断言响应成功
func (ts *TestSuite) AssertResponseSuccess(response map[string]interface{}) {
	ts.Require().NotNil(response)
	ts.Require().Equal("success", response["status"])
}

// AssertResponseError 断言响应错误
func (ts *TestSuite) AssertResponseError(response map[string]interface{}) {
	ts.Require().NotNil(response)
	ts.Require().Equal("error", response["status"])
}

// CreateTestUserWithToken 创建测试用户并返回token（包级别函数）
func CreateTestUserWithToken(username, email, password, role string) (*UserTokenInfo, error) {
	// 创建临时TestSuite实例
	ts := &TestSuite{}
	ts.SetupSuite()
	defer ts.TearDownSuite()

	return ts.CreateTestUserWithToken(username, email, password, role)
}
