package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Utils"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// AuthServiceTestSuite 认证服务测试套件
// 功能说明：
// 1. 测试用户注册功能
// 2. 测试用户登录功能
// 3. 测试密码重置功能
// 4. 测试邮箱验证功能
// 5. 测试Token刷新功能
type AuthServiceTestSuite struct {
	suite.Suite
	authService *Services.AuthService
}

// SetupSuite 测试套件初始化
func (ats *AuthServiceTestSuite) SetupSuite() {
	// 设置测试环境变量
	os.Setenv("SERVER_MODE", "test")
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DATABASE", ":memory:")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing-only-32-chars")

	// 加载配置
	Config.LoadConfig()

	// 初始化数据库
	Database.InitDB()
	Database.AutoMigrate()

	// 初始化认证服务
	ats.authService = Services.NewAuthService()
}

// TestRegister 测试用户注册
// 功能说明：
// 1. 测试正常注册流程
// 2. 测试重复用户名注册
// 3. 测试重复邮箱注册
// 4. 测试密码强度验证
func (ats *AuthServiceTestSuite) TestRegister() {
	// 测试正常注册
	request := Requests.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := ats.authService.Register(request)
	ats.Require().NoError(err)
	ats.Require().NotNil(user)
	ats.Require().Equal("testuser", user.Username)
	ats.Require().Equal("test@example.com", user.Email)
	ats.Require().Equal("user", user.Role)
	ats.Require().Equal(1, user.Status)

	// 测试重复用户名注册
	_, err = ats.authService.Register(request)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "username already exists")

	// 测试重复邮箱注册
	request2 := Requests.RegisterRequest{
		Username: "testuser2",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = ats.authService.Register(request2)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "email already exists")

	// 测试密码强度验证
	weakRequest := Requests.RegisterRequest{
		Username: "weakuser",
		Email:    "weak@example.com",
		Password: "123",
	}
	_, err = ats.authService.Register(weakRequest)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "password")
}

// TestLogin 测试用户登录
// 功能说明：
// 1. 测试正常登录流程
// 2. 测试错误密码登录
// 3. 测试不存在的用户登录
// 4. 测试禁用账户登录
func (ats *AuthServiceTestSuite) TestLogin() {
	// 先创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "logintest",
		Email:    "login@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 测试正常登录
	loginRequest := Requests.LoginRequest{
		Username: "logintest",
		Password: "password123",
	}

	token, loginUser, err := ats.authService.Login(loginRequest)
	ats.Require().NoError(err)
	ats.Require().NotEmpty(token)
	ats.Require().NotNil(loginUser)
	ats.Require().Equal(user.ID, loginUser.ID)

	// 测试错误密码
	wrongPasswordRequest := Requests.LoginRequest{
		Username: "logintest",
		Password: "wrongpassword",
	}
	_, _, err = ats.authService.Login(wrongPasswordRequest)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "invalid credentials")

	// 测试不存在的用户
	nonexistentRequest := Requests.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	_, _, err = ats.authService.Login(nonexistentRequest)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "invalid credentials")
}

// TestGetProfile 测试获取用户资料
// 功能说明：
// 1. 测试正常获取资料
// 2. 测试获取不存在的用户资料
func (ats *AuthServiceTestSuite) TestGetProfile() {
	// 创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "profiletest",
		Email:    "profile@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 测试正常获取资料
	profile, err := ats.authService.GetProfile(fmt.Sprintf("%d", user.ID))
	ats.Require().NoError(err)
	ats.Require().NotNil(profile)
	ats.Require().Equal(user.ID, profile.ID)
	ats.Require().Equal(user.Username, profile.Username)
	ats.Require().Equal(user.Email, profile.Email)
	ats.Require().Empty(profile.Password) // 密码不应该返回

	// 测试获取不存在的用户资料
	_, err = ats.authService.GetProfile("99999")
	ats.Require().Error(err)
}

// TestUpdateProfile 测试更新用户资料
// 功能说明：
// 1. 测试正常更新资料
// 2. 测试更新为重复用户名
// 3. 测试更新为重复邮箱
func (ats *AuthServiceTestSuite) TestUpdateProfile() {
	// 创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "updatetest",
		Email:    "update@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 创建另一个用户用于测试重复
	registerRequest2 := Requests.RegisterRequest{
		Username: "updatetest2",
		Email:    "update2@example.com",
		Password: "password123",
	}
	_, err = ats.authService.Register(registerRequest2)
	ats.Require().NoError(err)

	// 测试正常更新
	updateRequest := Requests.UpdateProfileRequest{
		Username: "updateduser",
		Email:    "updated@example.com",
		Avatar:   "https://example.com/avatar.jpg",
	}

	updatedUser, err := ats.authService.UpdateProfile(fmt.Sprintf("%d", user.ID), updateRequest)
	ats.Require().NoError(err)
	ats.Require().NotNil(updatedUser)
	ats.Require().Equal("updateduser", updatedUser.Username)
	ats.Require().Equal("updated@example.com", updatedUser.Email)
	ats.Require().Equal("https://example.com/avatar.jpg", updatedUser.Avatar)

	// 测试更新为重复用户名
	duplicateUsernameRequest := Requests.UpdateProfileRequest{
		Username: "updatetest2",
	}
	_, err = ats.authService.UpdateProfile(fmt.Sprintf("%d", user.ID), duplicateUsernameRequest)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "username already exists")

	// 测试更新为重复邮箱
	duplicateEmailRequest := Requests.UpdateProfileRequest{
		Email: "update2@example.com",
	}
	_, err = ats.authService.UpdateProfile(fmt.Sprintf("%d", user.ID), duplicateEmailRequest)
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "email already exists")
}

// TestRefreshToken 测试Token刷新
// 功能说明：
// 1. 测试正常Token刷新
// 2. 测试无效Token刷新
func (ats *AuthServiceTestSuite) TestRefreshToken() {
	// 创建测试用户并登录
	registerRequest := Requests.RegisterRequest{
		Username: "refreshtest",
		Email:    "refresh@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	loginRequest := Requests.LoginRequest{
		Username: "refreshtest",
		Password: "password123",
	}
	token, _, err := ats.authService.Login(loginRequest)
	ats.Require().NoError(err)

	// 测试正常Token刷新
	newToken, err := ats.authService.RefreshToken("Bearer " + token)
	ats.Require().NoError(err)
	ats.Require().NotEmpty(newToken)
	ats.Require().NotEqual(token, newToken)

	// 验证新Token
	claims, err := Utils.ValidateToken(newToken)
	ats.Require().NoError(err)
	ats.Require().Equal(user.ID, claims.UserID)
	ats.Require().Equal(user.Username, claims.Username)

	// 测试无效Token刷新
	_, err = ats.authService.RefreshToken("invalid-token")
	ats.Require().Error(err)
}

// TestPasswordReset 测试密码重置
// 功能说明：
// 1. 测试密码重置请求
// 2. 测试密码重置执行
func (ats *AuthServiceTestSuite) TestPasswordReset() {
	// 创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "resetuser",
		Email:    "reset@example.com",
		Password: "password123",
	}
	_, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 测试密码重置请求
	err = ats.authService.RequestPasswordReset("reset@example.com")
	ats.Require().NoError(err)

	// 测试不存在的邮箱
	err = ats.authService.RequestPasswordReset("nonexistent@example.com")
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "email not found")
}

// TestEmailVerification 测试邮箱验证
// 功能说明：
// 1. 测试邮箱验证请求
// 2. 测试邮箱验证执行
func (ats *AuthServiceTestSuite) TestEmailVerification() {
	// 创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "verifyuser",
		Email:    "verify@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 测试邮箱验证请求
	err = ats.authService.SendEmailVerification(fmt.Sprintf("%d", user.ID))
	ats.Require().NoError(err)

	// 测试已验证的邮箱
	user.VerifyEmail()
	if err := Database.DB.Save(user).Error; err != nil {
		ats.T().Fatal(err)
	}

	err = ats.authService.SendEmailVerification(fmt.Sprintf("%d", user.ID))
	ats.Require().Error(err)
	ats.Require().Contains(err.Error(), "email already verified")
}

// TestLogout 测试用户登出
// 功能说明：
// 1. 测试正常登出流程
func (ats *AuthServiceTestSuite) TestLogout() {
	// 创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "logoutuser",
		Email:    "logout@example.com",
		Password: "password123",
	}
	user, err := ats.authService.Register(registerRequest)
	ats.Require().NoError(err)

	// 测试正常登出
	err = ats.authService.Logout(fmt.Sprintf("%d", user.ID))
	ats.Require().NoError(err)
}

// TestPasswordValidation 测试密码验证
// 功能说明：
// 1. 测试密码强度验证
// 2. 测试密码哈希和验证
func (ats *AuthServiceTestSuite) TestPasswordValidation() {
	// 测试密码哈希
	password := "testpassword123"
	hashedPassword, err := Utils.HashPassword(password)
	ats.Require().NoError(err)
	ats.Require().NotEmpty(hashedPassword)
	ats.Require().NotEqual(password, hashedPassword)

	// 测试密码验证
	isValid := Utils.CheckPassword(password, hashedPassword)
	ats.Require().True(isValid)

	// 测试错误密码
	isValid = Utils.CheckPassword("wrongpassword", hashedPassword)
	ats.Require().False(isValid)
}

// TestUserModel 测试用户模型
// 功能说明：
// 1. 测试用户模型方法
// 2. 测试用户状态检查
func (ats *AuthServiceTestSuite) TestUserModel() {
	// 创建用户模型
	user := &Models.User{
		Username: "modeltest",
		Email:    "model@example.com",
		Role:     "user",
		Status:   1,
	}

	// 测试用户状态检查
	ats.Require().True(user.IsActive())
	ats.Require().False(user.IsAdmin())
	ats.Require().False(user.IsEmailVerified())

	// 测试管理员用户
	adminUser := &Models.User{
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Status:   1,
	}
	ats.Require().True(adminUser.IsAdmin())

	// 测试禁用用户
	disabledUser := &Models.User{
		Username: "disabled",
		Email:    "disabled@example.com",
		Role:     "user",
		Status:   0,
	}
	ats.Require().False(disabledUser.IsActive())

	// 测试邮箱验证
	user.VerifyEmail()
	ats.Require().True(user.IsEmailVerified())

	// 测试显示名称
	ats.Require().Equal("modeltest", user.GetDisplayName())

	// 测试头像URL
	ats.Require().Equal("/images/default-avatar.png", user.GetAvatarURL())
	user.Avatar = "https://example.com/avatar.jpg"
	ats.Require().Equal("https://example.com/avatar.jpg", user.GetAvatarURL())
}

// BenchmarkRegister 注册性能测试
func (ats *AuthServiceTestSuite) BenchmarkRegister(b *testing.B) {
	for i := 0; i < b.N; i++ {
		request := Requests.RegisterRequest{
			Username: fmt.Sprintf("benchuser%d", i),
			Email:    fmt.Sprintf("bench%d@example.com", i),
			Password: "password123",
		}
		_, err := ats.authService.Register(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLogin 登录性能测试
func (ats *AuthServiceTestSuite) BenchmarkLogin(b *testing.B) {
	// 先创建测试用户
	registerRequest := Requests.RegisterRequest{
		Username: "benchlogin",
		Email:    "benchlogin@example.com",
		Password: "password123",
	}
	_, err := ats.authService.Register(registerRequest)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loginRequest := Requests.LoginRequest{
			Username: "benchlogin",
			Password: "password123",
		}
		_, _, err := ats.authService.Login(loginRequest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestAuthService 运行认证服务测试
func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
