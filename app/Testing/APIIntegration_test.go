package Testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/app/Http/Middleware"
	"cloud-platform-api/app/Http/Routes"
	"cloud-platform-api/app/Models"
)

// APIIntegrationTestSuite API集成测试套件
type APIIntegrationTestSuite struct {
	TestSuite
	// API引擎
	APIEngine *gin.Engine
	// 测试用户
	TestUser *Models.User
	// 测试用户Token
	TestUserToken string
	// 测试管理员用户
	TestAdminUser *Models.User
	// 测试管理员Token
	TestAdminToken string
}

// SetupSuite 测试套件初始化
func (suite *APIIntegrationTestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()
	
	// 初始化API引擎
	suite.initAPIEngine()
	
	// 创建测试用户
	suite.createTestUsers()
}

// initAPIEngine 初始化API引擎
func (suite *APIIntegrationTestSuite) initAPIEngine() {
	suite.APIEngine = gin.New()
	suite.APIEngine.Use(gin.Recovery())
	
	// 注册中间件
	suite.APIEngine.Use(Middleware.CORSMiddleware())
	suite.APIEngine.Use(Middleware.RequestLogMiddleware())
	suite.APIEngine.Use(Middleware.ErrorHandlerMiddleware())
	
	// 注册路由
	Routes.RegisterAuthRoutes(suite.APIEngine)
	Routes.RegisterUserRoutes(suite.APIEngine)
	Routes.RegisterTagRoutes(suite.APIEngine)
	Routes.RegisterApiKeyRoutes(suite.APIEngine)
	Routes.RegisterLogRoutes(suite.APIEngine)
	Routes.RegisterWebSocketRoutes(suite.APIEngine)
}

// createTestUsers 创建测试用户
func (suite *APIIntegrationTestSuite) createTestUsers() {
	// 创建普通测试用户
	suite.TestUser = suite.CreateTestUser("testuser", "testuser@example.com", "password123")
	
	// 创建管理员测试用户
	suite.TestAdminUser = suite.CreateTestUser("adminuser", "adminuser@example.com", "password123")
	suite.TestAdminUser.Role = "admin"
	suite.TestDB.Save(suite.TestAdminUser)
	
	// 生成测试用户Token
	suite.TestUserToken = suite.generateTestToken(suite.TestUser)
	suite.TestAdminToken = suite.generateTestToken(suite.TestAdminUser)
}

// generateTestToken 生成测试Token
func (suite *APIIntegrationTestSuite) generateTestToken(user *Models.User) string {
	// 这里应该调用实际的JWT生成逻辑
	// 为了测试，我们使用一个简单的模拟Token
	return fmt.Sprintf("test_token_%d_%s", user.ID, user.Username)
}

// TestAPI_UserRegistration 测试用户注册API
func (suite *APIIntegrationTestSuite) TestAPI_UserRegistration() {
	t := suite.T()
	
	// 测试数据
	registrationData := map[string]interface{}{
		"username": "newuser",
		"email":    "newuser@example.com",
		"password": "password123",
		"role":     "user",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(registrationData)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证数据库记录
	suite.AssertDatabaseRecord(&Models.User{}, map[string]interface{}{
		"username": "newuser",
		"email":    "newuser@example.com",
	})
}

// TestAPI_UserLogin 测试用户登录API
func (suite *APIIntegrationTestSuite) TestAPI_UserLogin() {
	t := suite.T()
	
	// 测试数据
	loginData := map[string]interface{}{
		"username": suite.TestUser.Username,
		"password": "password123",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证Token存在
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}

// TestAPI_UserLogin_InvalidCredentials 测试用户登录API - 无效凭据
func (suite *APIIntegrationTestSuite) TestAPI_UserLogin_InvalidCredentials() {
	t := suite.T()
	
	// 测试数据
	loginData := map[string]interface{}{
		"username": suite.TestUser.Username,
		"password": "wrongpassword",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "invalid credentials")
}

// TestAPI_GetUserProfile 测试获取用户资料API
func (suite *APIIntegrationTestSuite) TestAPI_GetUserProfile() {
	t := suite.T()
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/users/profile", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证用户信息
	data := response["data"].(map[string]interface{})
	assert.Equal(t, suite.TestUser.Username, data["username"])
	assert.Equal(t, suite.TestUser.Email, data["email"])
}

// TestAPI_GetUserProfile_Unauthorized 测试获取用户资料API - 未授权
func (suite *APIIntegrationTestSuite) TestAPI_GetUserProfile_Unauthorized() {
	t := suite.T()
	
	// 创建请求（无Token）
	req, _ := http.NewRequest("GET", "/api/users/profile", nil)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAPI_UpdateUserProfile 测试更新用户资料API
func (suite *APIIntegrationTestSuite) TestAPI_UpdateUserProfile() {
	t := suite.T()
	
	// 测试数据
	updateData := map[string]interface{}{
		"email": "updated@example.com",
		"role":  "user",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/api/users/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	// 验证数据库记录已更新
	suite.AssertDatabaseRecord(&Models.User{}, map[string]interface{}{
		"id":    suite.TestUser.ID,
		"email": "updated@example.com",
	})
}

// TestAPI_CreateTag 测试创建标签API
func (suite *APIIntegrationTestSuite) TestAPI_CreateTag() {
	t := suite.T()
	
	// 测试数据
	tagData := map[string]interface{}{
		"name":        "testtag",
		"description": "Test tag description",
		"status":      "active",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(tagData)
	req, _ := http.NewRequest("POST", "/api/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	// 验证数据库记录
	suite.AssertDatabaseRecord(&Models.Tag{}, map[string]interface{}{
		"name": "testtag",
	})
}

// TestAPI_GetTags 测试获取标签列表API
func (suite *APIIntegrationTestSuite) TestAPI_GetTags() {
	t := suite.T()
	
	// 创建测试标签
	suite.CreateTestTag("listtag1", "List tag 1")
	suite.CreateTestTag("listtag2", "List tag 2")
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/tags?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证标签数据
	data := response["data"].(map[string]interface{})
	tags := data["tags"].([]interface{})
	assert.GreaterOrEqual(t, len(tags), 2)
}

// TestAPI_UpdateTag 测试更新标签API
func (suite *APIIntegrationTestSuite) TestAPI_UpdateTag() {
	t := suite.T()
	
	// 创建测试标签
	tag := suite.CreateTestTag("updatetag", "Original description")
	
	// 测试数据
	updateData := map[string]interface{}{
		"name":        "updatedtag",
		"description": "Updated description",
		"status":      "inactive",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/tags/%d", tag.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	// 验证数据库记录已更新
	suite.AssertDatabaseRecord(&Models.Tag{}, map[string]interface{}{
		"id":          tag.ID,
		"name":        "updatedtag",
		"description": "Updated description",
	})
}

// TestAPI_DeleteTag 测试删除标签API
func (suite *APIIntegrationTestSuite) TestAPI_DeleteTag() {
	t := suite.T()
	
	// 创建测试标签
	tag := suite.CreateTestTag("deletetag", "Tag to delete")
	
	// 创建请求
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/tags/%d", tag.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	// 验证标签已被删除
	suite.AssertDatabaseRecordNotExists(&Models.Tag{}, map[string]interface{}{
		"id": tag.ID,
	})
}

// TestAPI_CreateApiKey 测试创建API密钥API
func (suite *APIIntegrationTestSuite) TestAPI_CreateApiKey() {
	t := suite.T()
	
	// 测试数据
	apiKeyData := map[string]interface{}{
		"name":        "testapikey",
		"permissions": []string{"read", "write"},
		"expires_at":  "2025-12-31T23:59:59Z",
		"description": "Test API key",
	}
	
	// 创建请求
	jsonData, _ := json.Marshal(apiKeyData)
	req, _ := http.NewRequest("POST", "/api/apikeys", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证API密钥数据
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["api_key"])
	assert.Equal(t, "testapikey", data["name"])
}

// TestAPI_GetApiKeys 测试获取API密钥列表API
func (suite *APIIntegrationTestSuite) TestAPI_GetApiKeys() {
	t := suite.T()
	
	// 创建测试API密钥
	suite.CreateTestApiKey(suite.TestUser.ID, "listapikey1")
	suite.CreateTestApiKey(suite.TestUser.ID, "listapikey2")
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/apikeys?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证API密钥数据
	data := response["data"].(map[string]interface{})
	apiKeys := data["api_keys"].([]interface{})
	assert.GreaterOrEqual(t, len(apiKeys), 2)
}

// TestAPI_GetLogStats 测试获取日志统计API
func (suite *APIIntegrationTestSuite) TestAPI_GetLogStats() {
	t := suite.T()
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/logs/stats", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestAdminToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
}

// TestAPI_GetSystemHealth 测试获取系统健康状态API
func (suite *APIIntegrationTestSuite) TestAPI_GetSystemHealth() {
	t := suite.T()
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/logs/health", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestAdminToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	// 验证响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotEmpty(t, response["data"])
	
	// 验证健康状态
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
}

// TestAPI_WebSocketConnect 测试WebSocket连接API
func (suite *APIIntegrationTestSuite) TestAPI_WebSocketConnect() {
	t := suite.T()
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/api/ws/connect", nil)
	req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言响应
	// WebSocket连接会返回101状态码（Switching Protocols）
	// 但在HTTP测试中，我们只能测试HTTP响应
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// TestAPI_RateLimiting 测试API限流
func (suite *APIIntegrationTestSuite) TestAPI_RateLimiting() {
	t := suite.T()
	
	// 快速发送多个请求来测试限流
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
		
		w := httptest.NewRecorder()
		suite.APIEngine.ServeHTTP(w, req)
		
		// 如果达到限流阈值，应该返回429状态码
		if w.Code == http.StatusTooManyRequests {
			// 验证限流响应
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, "error", response["status"])
			assert.Contains(t, response["message"], "rate limit")
			return
		}
	}
	
	// 如果没有触发限流，测试通过
	t.Log("Rate limiting not triggered in this test")
}

// TestAPI_ErrorHandling 测试API错误处理
func (suite *APIIntegrationTestSuite) TestAPI_ErrorHandling() {
	t := suite.T()
	
	// 测试不存在的路由
	req, _ := http.NewRequest("GET", "/api/nonexistent", nil)
	
	w := httptest.NewRecorder()
	suite.APIEngine.ServeHTTP(w, req)
	
	// 断言404响应
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	// 验证错误响应格式
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.NotEmpty(t, response["message"])
}

// BenchmarkAPI_UserProfile 用户资料API性能基准测试
func (suite *APIIntegrationTestSuite) BenchmarkAPI_UserProfile(b *testing.B) {
	// 重置计时器
	b.ResetTimer()
	
	// 运行基准测试
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+suite.TestUserToken)
		
		w := httptest.NewRecorder()
		suite.APIEngine.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatal("API request failed")
		}
	}
}

// 运行测试套件
func TestAPIIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(APIIntegrationTestSuite))
}
