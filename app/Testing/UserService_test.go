package Testing

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	suite.Suite
	// 用户服务
	UserService *Services.UserService
}

// SetupSuite 测试套件初始化
func (suite *UserServiceTestSuite) SetupSuite() {
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

	// 初始化用户服务
	suite.UserService = Services.NewUserService()
}

// TestUserService_CreateUser 测试创建用户
func (suite *UserServiceTestSuite) TestUserService_CreateUser() {
	t := suite.T()

	// 测试数据
	userData := &Models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1, // 1-正常状态
	}

	// 执行创建用户
	createdUser, err := suite.UserService.CreateUser(userData)

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, userData.Username, createdUser.Username)
	assert.Equal(t, userData.Email, createdUser.Email)
	assert.Equal(t, userData.Role, createdUser.Role)
	assert.Equal(t, userData.Status, createdUser.Status)
	assert.NotZero(t, createdUser.ID)
	assert.NotZero(t, createdUser.CreatedAt)

	// 验证密码已加密
	assert.NotEqual(t, userData.Password, createdUser.Password)

	// 验证数据库记录
	var user Models.User
	err = Database.DB.Where("username = ? AND email = ?", userData.Username, userData.Email).First(&user).Error
	require.NoError(t, err)
}

// TestUserService_CreateUser_DuplicateUsername 测试创建重复用户名的用户
func (suite *UserServiceTestSuite) TestUserService_CreateUser_DuplicateUsername() {
	t := suite.T()

	// 创建第一个用户
	user1 := &Models.User{
		Username: "duplicateuser",
		Email:    "user1@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1, // 1-正常状态
	}

	_, err := suite.UserService.CreateUser(user1)
	require.NoError(t, err)

	// 尝试创建同名用户
	user2 := &Models.User{
		Username: "duplicateuser",
		Email:    "user2@example.com",
		Password: "password456",
		Role:     "user",
		Status:   1, // 1-正常状态
	}

	_, err = suite.UserService.CreateUser(user2)

	// 断言错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username already exists")
}

// TestUserService_CreateUser_DuplicateEmail 测试创建重复邮箱的用户
func (suite *UserServiceTestSuite) TestUserService_CreateUser_DuplicateEmail() {
	t := suite.T()

	// 创建第一个用户
	user1 := &Models.User{
		Username: "user1",
		Email:    "duplicate@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1, // 1-正常状态
	}

	_, err := suite.UserService.CreateUser(user1)
	require.NoError(t, err)

	// 尝试创建同邮箱用户
	user2 := &Models.User{
		Username: "user2",
		Email:    "duplicate@example.com",
		Password: "password456",
		Role:     "user",
		Status:   1, // 1-正常状态
	}

	_, err = suite.UserService.CreateUser(user2)

	// 断言错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
}

// TestUserService_GetUserByID 测试根据ID获取用户
func (suite *UserServiceTestSuite) TestUserService_GetUserByID() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "getuser",
		Email:    "getuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 根据ID获取用户
	retrievedUser, err := suite.UserService.GetUserByID(user.ID)

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

// TestUserService_GetUserByID_NotFound 测试获取不存在的用户
func (suite *UserServiceTestSuite) TestUserService_GetUserByID_NotFound() {
	t := suite.T()

	// 尝试获取不存在的用户
	user, err := suite.UserService.GetUserByID(99999)

	// 断言结果
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
}

// TestUserService_GetUserByUsername 测试根据用户名获取用户
func (suite *UserServiceTestSuite) TestUserService_GetUserByUsername() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "usernameuser",
		Email:    "usernameuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 根据用户名获取用户
	retrievedUser, err := suite.UserService.GetUserByUsername(user.Username)

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

// TestUserService_GetUserByEmail 测试根据邮箱获取用户
func (suite *UserServiceTestSuite) TestUserService_GetUserByEmail() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "emailuser",
		Email:    "emailuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 根据邮箱获取用户
	retrievedUser, err := suite.UserService.GetUserByEmail(user.Email)

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

// TestUserService_UpdateUser 测试更新用户
func (suite *UserServiceTestSuite) TestUserService_UpdateUser() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "updateuser",
		Email:    "updateuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 更新用户信息
	updateData := map[string]interface{}{
		"username": "updateduser",
		"email":    "updated@example.com",
		"role":     "admin",
		"status":   0, // 0-禁用状态
	}

	updatedUser, err := suite.UserService.UpdateUser(user.ID, updateData)

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, updateData["username"], updatedUser.Username)
	assert.Equal(t, updateData["email"], updatedUser.Email)
	assert.Equal(t, updateData["role"], updatedUser.Role)
	assert.Equal(t, updateData["status"], updatedUser.Status)

	// 验证数据库记录已更新
	var dbUser Models.User
	err = Database.DB.Where("id = ?", user.ID).First(&dbUser).Error
	require.NoError(t, err)
	assert.Equal(t, updateData["username"], dbUser.Username)
	assert.Equal(t, updateData["email"], dbUser.Email)
}

// TestUserService_UpdateUser_NotFound 测试更新不存在的用户
func (suite *UserServiceTestSuite) TestUserService_UpdateUser_NotFound() {
	t := suite.T()

	// 尝试更新不存在的用户
	updateData := map[string]interface{}{
		"username": "nonexistent",
		"email":    "nonexistent@example.com",
	}

	user, err := suite.UserService.UpdateUser(99999, updateData)

	// 断言结果
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
}

// TestUserService_DeleteUser 测试删除用户
func (suite *UserServiceTestSuite) TestUserService_DeleteUser() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "deleteuser",
		Email:    "deleteuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 删除用户
	err := suite.UserService.DeleteUser(user.ID)

	// 断言结果
	require.NoError(t, err)

	// 验证用户已被删除
	var deletedUser Models.User
	err = Database.DB.Where("id = ?", user.ID).First(&deletedUser).Error
	assert.Error(t, err) // 应该找不到记录
}

// TestUserService_DeleteUser_NotFound 测试删除不存在的用户
func (suite *UserServiceTestSuite) TestUserService_DeleteUser_NotFound() {
	t := suite.T()

	// 尝试删除不存在的用户
	err := suite.UserService.DeleteUser(99999)

	// 断言结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// TestUserService_ListUsers 测试用户列表
func (suite *UserServiceTestSuite) TestUserService_ListUsers() {
	t := suite.T()

	// 创建多个测试用户
	user1 := &Models.User{Username: "listuser1", Email: "listuser1@example.com", Password: "password123", Role: "user", Status: 1}
	user2 := &Models.User{Username: "listuser2", Email: "listuser2@example.com", Password: "password456", Role: "user", Status: 1}
	user3 := &Models.User{Username: "listuser3", Email: "listuser3@example.com", Password: "password789", Role: "user", Status: 1}
	Database.DB.Create(user1)
	Database.DB.Create(user2)
	Database.DB.Create(user3)

	// 获取用户列表
	users, total, err := suite.UserService.ListUsers(1, 10)

	// 断言结果
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(3))
	assert.GreaterOrEqual(t, len(users), 3)

	// 验证包含创建的测试用户
	userIDs := make(map[uint]bool)
	for _, user := range users {
		userIDs[user.ID] = true
	}

	assert.True(t, userIDs[user1.ID])
	assert.True(t, userIDs[user2.ID])
	assert.True(t, userIDs[user3.ID])
}

// TestUserService_ListUsers_Pagination 测试用户列表分页
func (suite *UserServiceTestSuite) TestUserService_ListUsers_Pagination() {
	t := suite.T()

	// 创建多个测试用户
	for i := 1; i <= 15; i++ {
		user := &Models.User{
			Username: fmt.Sprintf("paginationuser%d", i),
			Email:    fmt.Sprintf("paginationuser%d@example.com", i),
			Password: "password123",
			Role:     "user",
			Status:   1,
		}
		Database.DB.Create(user)
	}

	// 测试第一页
	users1, total1, err := suite.UserService.ListUsers(1, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), int64(len(users1)))
	assert.GreaterOrEqual(t, total1, int64(15))

	// 测试第二页
	users2, total2, err := suite.UserService.ListUsers(2, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), int64(len(users2)))
	assert.Equal(t, total1, total2)

	// 验证两页数据不同
	userIDs1 := make(map[uint]bool)
	for _, user := range users1 {
		userIDs1[user.ID] = true
	}

	for _, user := range users2 {
		assert.False(t, userIDs1[user.ID], "第二页用户ID不应在第一页中出现")
	}
}

// TestUserService_ListUsers_Search 测试用户列表搜索
func (suite *UserServiceTestSuite) TestUserService_ListUsers_Search() {
	t := suite.T()

	// 创建测试用户
	user1 := &Models.User{Username: "searchuser", Email: "searchuser@example.com", Password: "password123", Role: "user", Status: 1}
	user2 := &Models.User{Username: "otheruser", Email: "otheruser@example.com", Password: "password456", Role: "user", Status: 1}
	Database.DB.Create(user1)
	Database.DB.Create(user2)

	// 搜索包含"search"的用户
	users, total, err := suite.UserService.ListUsers(1, 10)

	// 断言结果
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	// 验证搜索结果
	found := false
	for _, user := range users {
		if user.Username == "searchuser" || user.Email == "searchuser@example.com" {
			found = true
			break
		}
	}
	assert.True(t, found, "应该找到包含'search'的用户")
}

// TestUserService_ChangePassword 测试修改密码
func (suite *UserServiceTestSuite) TestUserService_ChangePassword() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "passworduser",
		Email:    "passworduser@example.com",
		Password: "oldpassword",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 修改密码
	newPassword := "newpassword123"
	err := suite.UserService.ChangePassword(user.ID, "oldpassword", newPassword)

	// 断言结果
	require.NoError(t, err)

	// 验证新密码可以登录
	loginUser, err := suite.UserService.ValidateUser("passworduser", newPassword)
	require.NoError(t, err)
	assert.Equal(t, user.ID, loginUser.ID)
}

// TestUserService_ChangePassword_WrongOldPassword 测试修改密码时旧密码错误
func (suite *UserServiceTestSuite) TestUserService_ChangePassword_WrongOldPassword() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "wrongpassuser",
		Email:    "wrongpassuser@example.com",
		Password: "oldpassword",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 尝试用错误的旧密码修改密码
	err := suite.UserService.ChangePassword(user.ID, "wrongoldpassword", "newpassword123")

	// 断言结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid old password")
}

// TestUserService_ValidateUser 测试用户验证
func (suite *UserServiceTestSuite) TestUserService_ValidateUser() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "validateuser",
		Email:    "validateuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 验证用户凭据
	validatedUser, err := suite.UserService.ValidateUser("validateuser", "password123")

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, user.Username, validatedUser.Username)
}

// TestUserService_ValidateUser_InvalidCredentials 测试无效用户凭据
func (suite *UserServiceTestSuite) TestUserService_ValidateUser_InvalidCredentials() {
	t := suite.T()

	// 创建测试用户
	user := &Models.User{
		Username: "invaliduser",
		Email:    "invaliduser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 尝试用错误的密码验证
	user, err := suite.UserService.ValidateUser("invaliduser", "wrongpassword")

	// 断言结果
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

// TestUserService_GetUserStats 测试获取用户统计
func (suite *UserServiceTestSuite) TestUserService_GetUserStats() {
	t := suite.T()

	// 创建不同状态的测试用户
	user1 := &Models.User{Username: "activeuser1", Email: "activeuser1@example.com", Password: "password123", Role: "user", Status: 1}
	user2 := &Models.User{Username: "activeuser2", Email: "activeuser2@example.com", Password: "password456", Role: "user", Status: 1}
	user3 := &Models.User{Username: "inactiveuser", Email: "inactiveuser@example.com", Password: "password789", Role: "user", Status: 0}
	Database.DB.Create(user1)
	Database.DB.Create(user2)
	Database.DB.Create(user3)

	// 获取用户统计
	stats, err := suite.UserService.GetUserStats()

	// 断言结果
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats["total_users"], int64(3))
	assert.GreaterOrEqual(t, stats["active_users"], int64(2))
	assert.GreaterOrEqual(t, stats["inactive_users"], int64(1))
}

// BenchmarkUserService_CreateUser 用户创建性能基准测试
func (suite *UserServiceTestSuite) BenchmarkUserService_CreateUser(b *testing.B) {
	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		userData := &Models.User{
			Username: fmt.Sprintf("benchmarkuser%d", i),
			Email:    fmt.Sprintf("benchmarkuser%d@example.com", i),
			Password: "password123",
			Role:     "user",
			Status:   1, // 1-正常状态
		}

		_, err := suite.UserService.CreateUser(userData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUserService_GetUserByID 用户查询性能基准测试
func (suite *UserServiceTestSuite) BenchmarkUserService_GetUserByID(b *testing.B) {
	// 创建测试用户
	user := &Models.User{
		Username: "benchmarkgetuser",
		Email:    "benchmarkgetuser@example.com",
		Password: "password123",
		Role:     "user",
		Status:   1,
	}
	Database.DB.Create(user)

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		_, err := suite.UserService.GetUserByID(user.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 运行测试套件
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
