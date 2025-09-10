package Database

import (
	"cloud-platform-api/app/Models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

var TestDB *gorm.DB

// InitTestDB 初始化测试数据库
func InitTestDB() {
	var err error

	// 使用SQLite作为测试数据库
	TestDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// 自动迁移所有模型
	err = TestDB.AutoMigrate(
		&Models.User{},
		&Models.Category{},
		&Models.Tag{},
		&Models.Post{},
		&Models.AuditLog{},
	)

	if err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}

	// 临时替换主数据库连接
	originalDB := DB
	DB = TestDB

	// 在测试结束后恢复
	defer func() {
		DB = originalDB
	}()
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB() {
	if TestDB != nil {
		sqlDB, err := TestDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// SetupTestData 设置测试数据
func SetupTestData() {
	// 创建测试用户
	testUser := Models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "user",
	}

	TestDB.Create(&testUser)

	// 创建测试分类
	testCategory := Models.Category{
		Name:        "测试分类",
		Description: "这是一个测试分类",
	}

	TestDB.Create(&testCategory)

	// 创建测试标签
	testTag := Models.Tag{
		Name: "测试标签",
	}

	TestDB.Create(&testTag)

	// 创建测试文章
	testPost := Models.Post{
		Title:      "测试文章",
		Content:    "这是测试文章的内容",
		UserID:     testUser.ID,
		CategoryID: testCategory.ID,
		Status:     1,
	}

	TestDB.Create(&testPost)
}

// CleanTestData 清理测试数据
func CleanTestData() {
	TestDB.Exec("DELETE FROM posts")
	TestDB.Exec("DELETE FROM tags")
	TestDB.Exec("DELETE FROM categories")
	TestDB.Exec("DELETE FROM users")
	TestDB.Exec("DELETE FROM audit_logs")
}
