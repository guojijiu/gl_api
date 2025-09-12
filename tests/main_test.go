package tests

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/tests/testsetup"
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TestMain: 在本包所有测试前统一初始化
func TestMain(m *testing.M) {
	// 初始化测试环境（只执行一次）
	testsetup.Init()
	// 执行测试
	os.Exit(m.Run())
}

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
	tokenHelper    *testsetup.TokenHelper
	ctx            context.Context
	cancel         context.CancelFunc
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
	storageConfig := &Config.StorageConfig{
		BasePath: storagePath,
	}
	ts.storageManager = Storage.NewStorageManager(storageConfig)

	// 初始化数据库
	Database.InitDBWithLogger(ts.storageManager)

	// 运行数据库迁移
	Database.AutoMigrate()

	// 初始化TokenHelper
	ts.tokenHelper = testsetup.NewTokenHelper()

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
	tables := []string{"users", "posts", "categories", "tags", "audit_logs"}
	for _, table := range tables {
		if err := Database.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
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

// CreateTestUser 创建测试用户
// 功能说明：
// 1. 创建用于测试的用户数据
// 2. 返回用户信息供测试使用
// 3. 支持自定义用户属性
func (ts *TestSuite) CreateTestUser(username, email, password string) map[string]interface{} {
	userData := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
		"role":     "user",
		"status":   1,
	}

	return userData
}

// CreateTestUserWithToken 创建测试用户并返回token
// 功能说明：
// 1. 创建测试用户
// 2. 生成有效的JWT token
// 3. 返回用户信息和token
func (ts *TestSuite) CreateTestUserWithToken(username, email, password, role string) (*testsetup.UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateTestUserWithToken(username, email, password, role)
	if err != nil {
		return nil, err
	}

	return &testsetup.UserTokenInfo{
		User:  user,
		Token: token,
	}, nil
}

// CreateAdminUserWithToken 创建管理员用户并返回token
func (ts *TestSuite) CreateAdminUserWithToken() (*testsetup.UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateAdminUserWithToken()
	if err != nil {
		return nil, err
	}

	return &testsetup.UserTokenInfo{
		User:  user,
		Token: token,
	}, nil
}

// CreateNormalUserWithToken 创建普通用户并返回token
func (ts *TestSuite) CreateNormalUserWithToken() (*testsetup.UserTokenInfo, error) {
	user, token, err := ts.tokenHelper.CreateNormalUserWithToken()
	if err != nil {
		return nil, err
	}

	return &testsetup.UserTokenInfo{
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

// CreateTestPost 创建测试文章
// 功能说明：
// 1. 创建用于测试的文章数据
// 2. 返回文章信息供测试使用
// 3. 支持自定义文章属性
func (ts *TestSuite) CreateTestPost(title, content string, userID uint) map[string]interface{} {
	postData := map[string]interface{}{
		"title":   title,
		"content": content,
		"user_id": userID,
		"status":  1,
	}

	return postData
}

// AssertResponseSuccess 断言响应成功
// 功能说明：
// 1. 检查API响应是否成功
// 2. 验证响应格式
// 3. 提供详细的错误信息
func (ts *TestSuite) AssertResponseSuccess(response map[string]interface{}) {
	ts.Require().NotNil(response)
	ts.Require().Contains(response, "success")
	ts.Require().True(response["success"].(bool))
}

// AssertResponseError 断言响应错误
// 功能说明：
// 1. 检查API响应是否包含错误
// 2. 验证错误格式
// 3. 提供详细的错误信息
func (ts *TestSuite) AssertResponseError(response map[string]interface{}) {
	ts.Require().NotNil(response)
	ts.Require().Contains(response, "success")
	ts.Require().False(response["success"].(bool))
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

// RunTests 运行所有测试
func RunTests(t *testing.T) {
	// 运行测试套件
	suite.Run(t, new(TestSuite))
}
