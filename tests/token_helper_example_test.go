package tests

import (
	"testing"

	"cloud-platform-api/tests/testsetup"

	"github.com/stretchr/testify/suite"
)

// TokenHelperExampleTestSuite TokenHelper使用示例测试
type TokenHelperExampleTestSuite struct {
	testsetup.TestSuite
}

// TestCreateUserWithToken 测试创建用户并获取token
func (ts *TokenHelperExampleTestSuite) TestCreateUserWithToken() {
	// 创建测试用户并获取token
	userInfo, err := ts.CreateTestUserWithToken("testuser", "test@example.com", "password123", "user")
	ts.Require().NoError(err)
	ts.Require().NotNil(userInfo)
	ts.Require().NotNil(userInfo.User)
	ts.Require().NotEmpty(userInfo.Token)

	// 验证用户信息
	ts.Require().Equal("testuser", userInfo.User.Username)
	ts.Require().Equal("test@example.com", userInfo.User.Email)
	ts.Require().Equal("user", userInfo.User.Role)

	// 验证token有效性
	ts.AssertTokenValid(userInfo.Token)

	// 验证token内容
	claims, err := ts.TokenHelper.ValidateToken(userInfo.Token)
	ts.Require().NoError(err)
	ts.Require().Equal(userInfo.User.ID, claims.UserID)
	ts.Require().Equal(userInfo.User.Username, claims.Username)
	ts.Require().Equal(userInfo.User.Role, claims.Role)
}

// TestCreateAdminUserWithToken 测试创建管理员用户并获取token
func (ts *TokenHelperExampleTestSuite) TestCreateAdminUserWithToken() {
	// 创建管理员用户并获取token
	adminInfo, err := ts.CreateAdminUserWithToken()
	ts.Require().NoError(err)
	ts.Require().NotNil(adminInfo)
	ts.Require().NotNil(adminInfo.User)
	ts.Require().NotEmpty(adminInfo.Token)

	// 验证管理员信息
	ts.Require().Equal("admin", adminInfo.User.Username)
	ts.Require().Equal("admin@test.com", adminInfo.User.Email)
	ts.Require().Equal("admin", adminInfo.User.Role)

	// 验证token有效性
	ts.AssertTokenValid(adminInfo.Token)

	// 验证token内容
	claims, err := ts.TokenHelper.ValidateToken(adminInfo.Token)
	ts.Require().NoError(err)
	ts.Require().Equal(adminInfo.User.ID, claims.UserID)
	ts.Require().Equal(adminInfo.User.Username, claims.Username)
	ts.Require().Equal(adminInfo.User.Role, claims.Role)
}

// TestCreateNormalUserWithToken 测试创建普通用户并获取token
func (ts *TokenHelperExampleTestSuite) TestCreateNormalUserWithToken() {
	// 创建普通用户并获取token
	userInfo, err := ts.CreateNormalUserWithToken()
	ts.Require().NoError(err)
	ts.Require().NotNil(userInfo)
	ts.Require().NotNil(userInfo.User)
	ts.Require().NotEmpty(userInfo.Token)

	// 验证用户信息
	ts.Require().Equal("user", userInfo.User.Username)
	ts.Require().Equal("user@test.com", userInfo.User.Email)
	ts.Require().Equal("user", userInfo.User.Role)

	// 验证token有效性
	ts.AssertTokenValid(userInfo.Token)
}

// TestGetTokenHeaders 测试获取token请求头
func (ts *TokenHelperExampleTestSuite) TestGetTokenHeaders() {
	// 创建测试用户
	userInfo, err := ts.CreateTestUserWithToken("headertest", "header@example.com", "password123", "user")
	ts.Require().NoError(err)

	// 获取token请求头
	headers := ts.GetTestTokenHeaders(userInfo.Token)
	ts.Require().NotNil(headers)
	ts.Require().Contains(headers, "Authorization")
	ts.Require().Contains(headers, "Content-Type")
	ts.Require().Contains(headers, "Accept")

	// 验证Authorization头格式
	ts.Require().Equal("Bearer "+userInfo.Token, headers["Authorization"])
	ts.Require().Equal("application/json", headers["Content-Type"])
	ts.Require().Equal("application/json", headers["Accept"])
}

// TestGetAdminTokenHeaders 测试获取管理员token请求头
func (ts *TokenHelperExampleTestSuite) TestGetAdminTokenHeaders() {
	// 获取管理员token请求头
	headers, err := ts.GetAdminTokenHeaders()
	ts.Require().NoError(err)
	ts.Require().NotNil(headers)

	// 验证请求头内容
	ts.Require().Contains(headers, "Authorization")
	ts.Require().Contains(headers, "Content-Type")
	ts.Require().Contains(headers, "Accept")
	ts.Require().Contains(headers, "X-Test-User-ID")
	ts.Require().Contains(headers, "X-Test-User-Role")

	// 验证Authorization头格式
	ts.Require().Contains(headers["Authorization"], "Bearer ")
	ts.Require().Equal("application/json", headers["Content-Type"])
	ts.Require().Equal("application/json", headers["Accept"])
	ts.Require().Equal("admin", headers["X-Test-User-Role"])
}

// TestGetUserTokenHeaders 测试获取普通用户token请求头
func (ts *TokenHelperExampleTestSuite) TestGetUserTokenHeaders() {
	// 获取普通用户token请求头
	headers, err := ts.GetUserTokenHeaders()
	ts.Require().NoError(err)
	ts.Require().NotNil(headers)

	// 验证请求头内容
	ts.Require().Contains(headers, "Authorization")
	ts.Require().Contains(headers, "Content-Type")
	ts.Require().Contains(headers, "Accept")
	ts.Require().Contains(headers, "X-Test-User-ID")
	ts.Require().Contains(headers, "X-Test-User-Role")

	// 验证Authorization头格式
	ts.Require().Contains(headers["Authorization"], "Bearer ")
	ts.Require().Equal("application/json", headers["Content-Type"])
	ts.Require().Equal("application/json", headers["Accept"])
	ts.Require().Equal("user", headers["X-Test-User-Role"])
}

// TestTokenValidation 测试token验证
func (ts *TokenHelperExampleTestSuite) TestTokenValidation() {
	// 创建有效token
	userInfo, err := ts.CreateTestUserWithToken("validationtest", "validation@example.com", "password123", "user")
	ts.Require().NoError(err)

	// 验证有效token
	ts.AssertTokenValid(userInfo.Token)

	// 验证无效token
	ts.AssertTokenInvalid("invalid.token.here")
	ts.AssertTokenInvalid("")
	ts.AssertTokenInvalid("Bearer invalid")
}

// TestCreateExpiredToken 测试创建过期token
func (ts *TokenHelperExampleTestSuite) TestCreateExpiredToken() {
	// 创建测试用户
	userInfo, err := ts.CreateTestUserWithToken("expiredtest", "expired@example.com", "password123", "user")
	ts.Require().NoError(err)

	// 创建过期token
	expiredToken, err := ts.TokenHelper.CreateExpiredToken(userInfo.User)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(expiredToken)

	// 验证过期token无效
	ts.AssertTokenInvalid(expiredToken)
}

// TestCreateInvalidToken 测试创建无效token
func (ts *TokenHelperExampleTestSuite) TestCreateInvalidToken() {
	// 创建无效token
	invalidToken := ts.TokenHelper.CreateInvalidToken()
	ts.Require().NotEmpty(invalidToken)

	// 验证无效token
	ts.AssertTokenInvalid(invalidToken)
}

// TestCreateMultipleUsersWithTokens 测试创建多个用户并获取token
func (ts *TokenHelperExampleTestSuite) TestCreateMultipleUsersWithTokens() {
	// 创建多个用户
	users, err := ts.TokenHelper.CreateMultipleUsersWithTokens(3)
	ts.Require().NoError(err)
	ts.Require().Len(users, 3)

	// 验证每个用户
	for username, userInfo := range users {
		ts.Require().NotNil(userInfo)
		ts.Require().NotNil(userInfo.User)
		ts.Require().NotEmpty(userInfo.Token)

		// 验证token有效性
		ts.AssertTokenValid(userInfo.Token)

		// 验证用户名格式
		ts.Require().Contains(username, "testuser")
	}
}

// TestTokenHelperExample 运行TokenHelper示例测试
func TestTokenHelperExample(t *testing.T) {
	suite.Run(t, new(TokenHelperExampleTestSuite))
}
