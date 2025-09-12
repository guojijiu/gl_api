package Controllers

import (
	"bytes"
	"cloud-platform-api/app/Http/Controllers"
	"cloud-platform-api/tests/testsetup"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// AuthControllerTestSuite 认证控制器测试套件
type AuthControllerTestSuite struct {
	testsetup.TestSuite
	authController *Controllers.AuthController
}

// SetupSuite 测试套件初始化
func (ats *AuthControllerTestSuite) SetupSuite() {
	// 调用父类SetupSuite
	ats.TestSuite.SetupSuite()
	ats.authController = Controllers.NewAuthController()
}

// TestRegister 测试用户注册
func (ats *AuthControllerTestSuite) TestRegister() {
	// 测试正常注册
	registerData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)
	ats.Require().Contains(response, "data")

	// 验证返回的用户数据
	userData := response["data"].(map[string]interface{})
	ats.Require().Equal("testuser", userData["username"])
	ats.Require().Equal("test@example.com", userData["email"])
	ats.Require().NotContains(userData, "password") // 密码不应该返回
}

// TestLogin 测试用户登录
func (ats *AuthControllerTestSuite) TestLogin() {
	// 先创建测试用户
	_, err := ats.CreateTestUserWithToken("logintest", "login@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试正常登录
	loginData := map[string]interface{}{
		"username": "logintest",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)

	// 验证返回的token
	data := response["data"].(map[string]interface{})
	ats.Require().Contains(data, "token")
	token := data["token"].(string)
	ats.Require().NotEmpty(token)

	// 验证token有效性
	ats.AssertTokenValid(token)

	// 验证返回的用户信息
	ats.Require().Contains(data, "user")
	userData := data["user"].(map[string]interface{})
	ats.Require().Equal("logintest", userData["username"])
	ats.Require().Equal("login@example.com", userData["email"])
}

// TestGetProfile 测试获取用户资料
func (ats *AuthControllerTestSuite) TestGetProfile() {
	// 创建测试用户并获取token
	userInfo, err := ats.CreateTestUserWithToken("profiletest", "profile@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试获取用户资料（需要认证）
	req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)

	// 验证返回的用户资料
	ats.Require().Contains(response, "data")
	userData := response["data"].(map[string]interface{})
	ats.Require().Equal("profiletest", userData["username"])
	ats.Require().Equal("profile@example.com", userData["email"])
	ats.Require().NotContains(userData, "password") // 密码不应该返回
}

// TestGetProfileUnauthorized 测试未授权获取用户资料
func (ats *AuthControllerTestSuite) TestGetProfileUnauthorized() {
	// 测试未授权访问
	req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Content-Type", "application/json")
	// 不设置Authorization头

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseError(response)
}

// TestGetProfileInvalidToken 测试无效token获取用户资料
func (ats *AuthControllerTestSuite) TestGetProfileInvalidToken() {
	// 测试无效token
	req, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseError(response)
}

// TestUpdateProfile 测试更新用户资料
func (ats *AuthControllerTestSuite) TestUpdateProfile() {
	// 创建测试用户并获取token
	userInfo, err := ats.CreateTestUserWithToken("updatetest", "update@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试更新用户资料
	updateData := map[string]interface{}{
		"username": "updateduser",
		"email":    "updated@example.com",
		"avatar":   "https://example.com/avatar.jpg",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/api/v1/auth/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)

	// 验证更新后的用户资料
	ats.Require().Contains(response, "data")
	userData := response["data"].(map[string]interface{})
	ats.Require().Equal("updateduser", userData["username"])
	ats.Require().Equal("updated@example.com", userData["email"])
	ats.Require().Equal("https://example.com/avatar.jpg", userData["avatar"])
}

// TestRefreshToken 测试Token刷新
func (ats *AuthControllerTestSuite) TestRefreshToken() {
	// 创建测试用户并获取token
	userInfo, err := ats.CreateTestUserWithToken("refreshtest", "refresh@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试Token刷新
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)

	// 验证返回的新token
	ats.Require().Contains(response, "data")
	data := response["data"].(map[string]interface{})
	ats.Require().Contains(data, "token")
	newToken := data["token"].(string)
	ats.Require().NotEmpty(newToken)
	ats.Require().NotEqual(userInfo.Token, newToken) // 新token应该不同

	// 验证新token有效性
	ats.AssertTokenValid(newToken)
}

// TestLogout 测试用户登出
func (ats *AuthControllerTestSuite) TestLogout() {
	// 创建测试用户并获取token
	userInfo, err := ats.CreateTestUserWithToken("logouttest", "logout@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 验证原始token有效
	ats.AssertTokenValid(userInfo.Token)

	// 测试用户登出
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)

	// 验证token已被加入黑名单（通过尝试使用该token访问受保护的接口）
	req2, _ := http.NewRequest("GET", "/api/v1/auth/profile", nil)
	req2.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w2, req2)

	// 应该返回未授权错误
	ats.Require().Equal(http.StatusUnauthorized, w2.Code)
}

// TestPasswordReset 测试密码重置
func (ats *AuthControllerTestSuite) TestPasswordReset() {
	// 创建测试用户
	_, err := ats.CreateTestUserWithToken("resetuser", "reset@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试密码重置请求
	resetData := map[string]interface{}{
		"email": "reset@example.com",
	}

	jsonData, _ := json.Marshal(resetData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/password/reset-request", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)
}

// TestEmailVerification 测试邮箱验证
func (ats *AuthControllerTestSuite) TestEmailVerification() {
	// 创建测试用户并获取token
	userInfo, err := ats.CreateTestUserWithToken("verifyuser", "verify@example.com", "password123", "user")
	ats.Require().NoError(err)

	// 测试邮箱验证请求
	req, _ := http.NewRequest("POST", "/api/v1/auth/email/verify-request", nil)
	req.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	ats.Require().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	ats.Require().NoError(err)

	ats.AssertResponseSuccess(response)
}

// TestAdminAccess 测试管理员权限
func (ats *AuthControllerTestSuite) TestAdminAccess() {
	// 创建管理员用户
	adminInfo, err := ats.CreateAdminUserWithToken()
	ats.Require().NoError(err)

	// 测试管理员访问管理员接口
	req, _ := http.NewRequest("GET", "/api/v1/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+adminInfo.Token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w, req)

	// 管理员应该能够访问
	ats.Require().Equal(http.StatusOK, w.Code)

	// 创建普通用户
	userInfo, err := ats.CreateNormalUserWithToken()
	ats.Require().NoError(err)

	// 测试普通用户访问管理员接口
	req2, _ := http.NewRequest("GET", "/api/v1/admin/dashboard", nil)
	req2.Header.Set("Authorization", "Bearer "+userInfo.Token)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	testsetup.Router.ServeHTTP(w2, req2)

	// 普通用户应该被拒绝访问
	ats.Require().Equal(http.StatusForbidden, w2.Code)
}

// TestAuthController 运行认证控制器测试
func TestAuthController(t *testing.T) {
	suite.Run(t, new(AuthControllerTestSuite))
}
