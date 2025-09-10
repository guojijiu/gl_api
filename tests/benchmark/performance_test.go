package benchmark

import (
	"bytes"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Routes"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Storage"
	"cloud-platform-api/app/Utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 基准测试配置
var (
	testDB    *gorm.DB
	router    *gin.Engine
	authToken string
	testUser  *Models.User
)

// 初始化基准测试环境
func init() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 初始化数据库
	setupTestDatabase()

	// 初始化路由
	setupTestRouter()

	// 创建测试用户和token
	setupTestUser()
}

// setupTestDatabase 设置测试数据库
func setupTestDatabase() {
	var err error
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// 自动迁移
	err = testDB.AutoMigrate(
		&Models.User{},
		&Models.Post{},
		&Models.Category{},
		&Models.Tag{},
		&Models.AuditLog{},
		&Models.MonitoringMetric{},
		&Models.Alert{},
		&Models.AlertRule{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate database: %v", err))
	}

	// 设置全局数据库连接
	Database.SetDB(testDB)
}

// setupTestRouter 设置测试路由
func setupTestRouter() {
	// 创建存储管理器
	storageManager := Storage.NewStorageManager("./test_storage")

	// 创建路由
	router = gin.New()
	Routes.RegisterRoutes(router, storageManager, nil)
}

// setupTestUser 创建测试用户
func setupTestUser() {
	// 创建测试用户
	testUser = &Models.User{
		Username: "benchmark_user",
		Email:    "benchmark@example.com",
		Password: "benchmark_password",
		Role:     "user",
		Status:   "active",
	}

	// 保存到数据库
	if err := testDB.Create(testUser).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test user: %v", err))
	}

	// 生成JWT token
	token, err := Utils.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate token: %v", err))
	}

	authToken = token
}

// BenchmarkUserRegistration 用户注册性能测试
func BenchmarkUserRegistration(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 准备请求数据
		userData := map[string]interface{}{
			"username":         fmt.Sprintf("user_%d", i),
			"email":            fmt.Sprintf("user_%d@example.com", i),
			"password":         "password123",
			"confirm_password": "password123",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkUserLogin 用户登录性能测试
func BenchmarkUserLogin(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 准备请求数据
		loginData := map[string]interface{}{
			"username": "benchmark_user",
			"password": "benchmark_password",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkGetUserProfile 获取用户资料性能测试
func BenchmarkGetUserProfile(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", testUser.ID), nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkCreatePost 创建文章性能测试
func BenchmarkCreatePost(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 准备请求数据
		postData := map[string]interface{}{
			"title":   fmt.Sprintf("Benchmark Post %d", i),
			"content": fmt.Sprintf("This is benchmark post content %d", i),
			"excerpt": fmt.Sprintf("Benchmark post excerpt %d", i),
			"status":  "published",
		}

		jsonData, _ := json.Marshal(postData)
		req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkGetPosts 获取文章列表性能测试
func BenchmarkGetPosts(b *testing.B) {
	// 预先创建一些测试文章
	for i := 0; i < 100; i++ {
		post := &Models.Post{
			Title:    fmt.Sprintf("Test Post %d", i),
			Content:  fmt.Sprintf("Test content %d", i),
			Excerpt:  fmt.Sprintf("Test excerpt %d", i),
			Status:   "published",
			AuthorID: testUser.ID,
		}
		testDB.Create(post)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/posts?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkHealthCheck 健康检查性能测试
func BenchmarkHealthCheck(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)

		// 执行请求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkConcurrentRequests 并发请求性能测试
func BenchmarkConcurrentRequests(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/health", nil)

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkDatabaseOperations 数据库操作性能测试
func BenchmarkDatabaseOperations(b *testing.B) {
	b.Run("CreateUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &Models.User{
				Username: fmt.Sprintf("db_user_%d", i),
				Email:    fmt.Sprintf("db_user_%d@example.com", i),
				Password: "password123",
				Role:     "user",
				Status:   "active",
			}
			testDB.Create(user)
		}
	})

	b.Run("QueryUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var user Models.User
			testDB.First(&user, testUser.ID)
		}
	})

	b.Run("UpdateUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testDB.Model(testUser).Update("updated_at", time.Now())
		}
	})
}

// BenchmarkJWTGeneration JWT生成性能测试
func BenchmarkJWTGeneration(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Utils.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
		if err != nil {
			b.Errorf("Failed to generate JWT: %v", err)
		}
	}
}

// BenchmarkJWTValidation JWT验证性能测试
func BenchmarkJWTValidation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Utils.ValidateToken(authToken)
		if err != nil {
			b.Errorf("Failed to validate JWT: %v", err)
		}
	}
}

// BenchmarkPasswordHashing 密码哈希性能测试
func BenchmarkPasswordHashing(b *testing.B) {
	password := "benchmark_password"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Utils.HashPassword(password)
		if err != nil {
			b.Errorf("Failed to hash password: %v", err)
		}
	}
}

// BenchmarkPasswordVerification 密码验证性能测试
func BenchmarkPasswordVerification(b *testing.B) {
	password := "benchmark_password"
	hashedPassword, _ := Utils.HashPassword(password)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := Utils.CheckPassword(password, hashedPassword)
		if err != nil {
			b.Errorf("Failed to verify password: %v", err)
		}
	}
}

// BenchmarkMemoryUsage 内存使用性能测试
func BenchmarkMemoryUsage(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建大量数据
		users := make([]Models.User, 1000)
		for j := 0; j < 1000; j++ {
			users[j] = Models.User{
				Username: fmt.Sprintf("memory_user_%d_%d", i, j),
				Email:    fmt.Sprintf("memory_user_%d_%d@example.com", i, j),
				Password: "password123",
				Role:     "user",
				Status:   "active",
			}
		}

		// 模拟处理
		_ = users
	}
}

// BenchmarkConcurrentDatabaseWrites 并发数据库写入性能测试
func BenchmarkConcurrentDatabaseWrites(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			user := &Models.User{
				Username: fmt.Sprintf("concurrent_user_%d", i),
				Email:    fmt.Sprintf("concurrent_user_%d@example.com", i),
				Password: "password123",
				Role:     "user",
				Status:   "active",
			}
			testDB.Create(user)
			i++
		}
	})
}

// 清理测试环境
func TestMain(m *testing.M) {
	// 运行测试
	code := m.Run()

	// 清理测试数据
	os.RemoveAll("./test_storage")

	// 退出
	os.Exit(code)
}
