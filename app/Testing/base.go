package Testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Services"
)

// TestSuite 测试套件基础结构
type TestSuite struct {
	suite.Suite
	// 测试配置
	Config *Config.TestConfig
	// 测试数据库
	TestDB *gorm.DB
	// SQL Mock
	SQLMock sqlmock.Sqlmock
	// 测试Redis客户端
	TestRedis *redis.Client
	// 测试上下文
	Ctx context.Context
	// 测试Gin引擎
	TestEngine *gin.Engine
	// 测试服务
	Services *TestServices
	// 测试数据清理器
	Cleanup *TestCleanup
}

// TestServices 测试服务集合
type TestServices struct {
	// 用户服务
	UserService *Services.UserService
	// 认证服务
	AuthService *Services.AuthService
	// 标签服务
	TagService *Services.TagService
	// API密钥服务
	ApiKeyService *Services.ApiKeyService
	// 审计服务
	AuditService *Services.AuditService
	// 日志管理服务
	LogManagerService *Services.LogManagerService
	// 日志监控服务
	LogMonitorService *Services.LogMonitorService
	// WebSocket服务
	WebSocketService *Services.WebSocketService
}

// TestCleanup 测试数据清理器
type TestCleanup struct {
	// 需要清理的表
	Tables []string
	// 需要清理的Redis键
	RedisKeys []string
	// 需要清理的文件
	Files []string
}

// SetupSuite 测试套件初始化
func (ts *TestSuite) SetupSuite() {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	
	// 初始化测试配置
	ts.initTestConfig()
	
	// 初始化测试数据库
	ts.initTestDatabase()
	
	// 初始化测试Redis
	ts.initTestRedis()
	
	// 初始化测试服务
	ts.initTestServices()
	
	// 初始化测试引擎
	ts.initTestEngine()
	
	// 初始化测试上下文
	ts.Ctx = context.Background()
	
	// 初始化测试数据清理器
	ts.Cleanup = &TestCleanup{
		Tables:   []string{},
		RedisKeys: []string{},
		Files:    []string{},
	}
}

// TearDownSuite 测试套件清理
func (ts *TestSuite) TearDownSuite() {
	// 清理测试数据
	ts.cleanupTestData()
	
	// 关闭测试数据库连接
	if ts.TestDB != nil {
		sqlDB, err := ts.TestDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	
	// 关闭测试Redis连接
	if ts.TestRedis != nil {
		ts.TestRedis.Close()
	}
}

// SetupTest 单个测试初始化
func (ts *TestSuite) SetupTest() {
	// 开始测试事务
	if ts.Config.Database.RollbackTransactions {
		ts.TestDB = ts.TestDB.Begin()
	}
}

// TearDownTest 单个测试清理
func (ts *TestSuite) TearDownTest() {
	// 回滚测试事务
	if ts.Config.Database.RollbackTransactions && ts.TestDB != nil {
		ts.TestDB.Rollback()
	}
}

// initTestConfig 初始化测试配置
func (ts *TestSuite) initTestConfig() {
	ts.Config = &Config.TestConfig{}
	ts.Config.SetDefaults()
	
	// 从环境变量加载配置
	ts.Config.BindEnvs()
	
	// 验证配置
	err := ts.Config.Validate()
	if err != nil {
		panic(fmt.Sprintf("invalid test config: %v", err))
	}
}

// initTestDatabase 初始化测试数据库
func (ts *TestSuite) initTestDatabase() {
	var err error
	
	if ts.Config.Database.Type == "sqlite" && ts.Config.Database.InMemory {
		// 使用内存SQLite数据库
		ts.TestDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	} else {
		// 使用配置的数据库
		ts.TestDB, err = gorm.Open(sqlite.Open(ts.Config.Database.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}
	
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}
	
	// 自动迁移数据库表
	err = ts.TestDB.AutoMigrate(
		// 在这里添加需要迁移的模型
	)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate test database: %v", err))
	}
	
	// 加载测试数据种子
	if ts.Config.Database.SeedFile != "" {
		ts.loadTestDataSeed()
	}
}

// initTestRedis 初始化测试Redis
func (ts *TestSuite) initTestRedis() {
	if ts.Config.Cache.Type == "redis" {
		ts.TestRedis = redis.NewClient(&redis.Options{
			Addr:         ts.Config.Cache.Redis.Addr,
			Password:     ts.Config.Cache.Redis.Password,
			DB:           ts.Config.Cache.Redis.DB,
			DialTimeout:  ts.Config.Cache.Redis.DialTimeout,
			ReadTimeout:  ts.Config.Cache.Redis.ReadTimeout,
			WriteTimeout: ts.Config.Cache.Redis.WriteTimeout,
		})
		
		// 测试Redis连接
		_, err := ts.TestRedis.Ping(ts.Ctx).Result()
		if err != nil {
			panic(fmt.Sprintf("failed to connect to test redis: %v", err))
		}
	}
}

// initTestServices 初始化测试服务
func (ts *TestSuite) initTestServices() {
	ts.Services = &TestServices{
		UserService:        Services.NewUserService(ts.TestDB),
		AuthService:        Services.NewAuthService(ts.TestDB),
		TagService:         Services.NewTagService(ts.TestDB),
		ApiKeyService:      Services.NewApiKeyService(ts.TestDB),
		AuditService:       Services.NewAuditService(ts.TestDB),
		LogManagerService:  Services.NewLogManagerService(),
		LogMonitorService:  Services.NewLogMonitorService(),
		WebSocketService:   Services.NewWebSocketService(),
	}
}

// initTestEngine 初始化测试Gin引擎
func (ts *TestSuite) initTestEngine() {
	ts.TestEngine = gin.New()
	ts.TestEngine.Use(gin.Recovery())
}

// loadTestDataSeed 加载测试数据种子
func (ts *TestDatabase) loadTestDataSeed() {
	if ts.Config.Database.SeedFile == "" {
		return
	}
	
	// 检查种子文件是否存在
	if _, err := os.Stat(ts.Config.Database.SeedFile); os.IsNotExist(err) {
		return
	}
	
	// 读取种子文件
	seedData, err := os.ReadFile(ts.Config.Database.SeedFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read seed file: %v", err))
	}
	
	// 执行种子SQL
	err = ts.TestDB.Exec(string(seedData)).Error
	if err != nil {
		panic(fmt.Sprintf("failed to execute seed SQL: %v", err))
	}
}

// cleanupTestData 清理测试数据
func (ts *TestSuite) cleanupTestData() {
	if !ts.Config.Base.CleanupTestData {
		return
	}
	
	// 清理数据库表
	for _, table := range ts.Cleanup.Tables {
		ts.TestDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
	
	// 清理Redis键
	if ts.TestRedis != nil {
		for _, key := range ts.Cleanup.RedisKeys {
			ts.TestRedis.Del(ts.Ctx, key)
		}
	}
	
	// 清理文件
	for _, file := range ts.Cleanup.Files {
		os.Remove(file)
	}
}

// CreateTestUser 创建测试用户
func (ts *TestSuite) CreateTestUser(username, email, password string) *Models.User {
	user := &Models.User{
		Username: username,
		Email:    email,
		Password: password,
		Role:     "user",
		Status:   "active",
	}
	
	err := ts.TestDB.Create(user).Error
	assert.NoError(ts.T(), err)
	
	// 添加到清理列表
	ts.Cleanup.Tables = append(ts.Cleanup.Tables, "users")
	
	return user
}

// CreateTestTag 创建测试标签
func (ts *TestSuite) CreateTestTag(name, description string) *Models.Tag {
	tag := &Models.Tag{
		Name:        name,
		Description: description,
		Status:      "active",
	}
	
	err := ts.TestDB.Create(tag).Error
	assert.NoError(ts.T(), err)
	
	// 添加到清理列表
	ts.Cleanup.Tables = append(ts.Cleanup.Tables, "tags")
	
	return tag
}

// CreateTestApiKey 创建测试API密钥
func (ts *TestSuite) CreateTestApiKey(userID uint, name string) *Models.ApiKey {
	apiKey := &Models.ApiKey{
		UserID:      userID,
		Name:        name,
		KeyHash:     "test_hash",
		Permissions: []string{"read", "write"},
		Status:      "active",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	
	err := ts.TestDB.Create(apiKey).Error
	assert.NoError(ts.T(), err)
	
	// 添加到清理列表
	ts.Cleanup.Tables = append(ts.Cleanup.Tables, "api_keys")
	
	return apiKey
}

// GetTestContext 获取测试上下文
func (ts *TestSuite) GetTestContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(nil)
	return ctx
}

// GetTestContextWithUser 获取带用户信息的测试上下文
func (ts *TestSuite) GetTestContextWithUser(user *Models.User) *gin.Context {
	ctx := ts.GetTestContext()
	ctx.Set("user_id", user.ID)
	ctx.Set("username", user.Username)
	ctx.Set("user_role", user.Role)
	return ctx
}

// AssertResponseSuccess 断言响应成功
func (ts *TestSuite) AssertResponseSuccess(response *gin.Context) {
	assert.Equal(ts.T(), 200, response.Writer.Status())
}

// AssertResponseError 断言响应错误
func (ts *TestSuite) AssertResponseError(response *gin.Context, expectedStatus int) {
	assert.Equal(ts.T(), expectedStatus, response.Writer.Status())
}

// AssertDatabaseRecord 断言数据库记录存在
func (ts *TestSuite) AssertDatabaseRecord(model interface{}, conditions map[string]interface{}) {
	var count int64
	query := ts.TestDB.Model(model)
	
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	
	err := query.Count(&count).Error
	assert.NoError(ts.T(), err)
	assert.Greater(ts.T(), count, int64(0))
}

// AssertDatabaseRecordNotExists 断言数据库记录不存在
func (ts *TestSuite) AssertDatabaseRecordNotExists(model interface{}, conditions map[string]interface{}) {
	var count int64
	query := ts.TestDB.Model(model)
	
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	
	err := query.Count(&count).Error
	assert.NoError(ts.T(), err)
	assert.Equal(ts.T(), int64(0), count)
}

// WaitForCondition 等待条件满足
func (ts *TestSuite) WaitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// CreateTempFile 创建临时文件
func (ts *TestSuite) CreateTempFile(content string) string {
	tmpFile, err := os.CreateTemp("", "test_*")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp file: %v", err))
	}
	
	_, err = tmpFile.WriteString(content)
	if err != nil {
		panic(fmt.Sprintf("failed to write to temp file: %v", err))
	}
	
	tmpFile.Close()
	
	// 添加到清理列表
	ts.Cleanup.Files = append(ts.Cleanup.Files, tmpFile.Name())
	
	return tmpFile.Name()
}

// CreateTempDir 创建临时目录
func (ts *TestSuite) CreateTempDir() string {
	tmpDir, err := os.MkdirTemp("", "test_*")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp dir: %v", err))
	}
	
	// 添加到清理列表
	ts.Cleanup.Files = append(ts.Cleanup.Files, tmpDir)
	
	return tmpDir
}
