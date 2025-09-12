package testsetup

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Utils"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenHelper Token辅助工具
// 功能说明：
// 1. 为测试提供有效的JWT token
// 2. 创建测试用户并生成token
// 3. 支持不同角色的用户token
// 4. 提供token验证功能
type TokenHelper struct {
	authService *Services.AuthService
}

// NewTokenHelper 创建Token辅助工具
func NewTokenHelper() *TokenHelper {
	return &TokenHelper{
		authService: Services.NewAuthService(),
	}
}

// CreateTestUserWithToken 创建测试用户并返回token
// 功能说明：
// 1. 创建测试用户
// 2. 生成有效的JWT token
// 3. 返回用户信息和token
func (th *TokenHelper) CreateTestUserWithToken(username, email, password, role string) (*Models.User, string, error) {
	// 创建用户
	registerRequest := Requests.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	user, err := th.authService.Register(registerRequest)
	if err != nil {
		return nil, "", fmt.Errorf("创建测试用户失败: %v", err)
	}

	// 如果需要管理员角色，更新用户角色
	if role == "admin" {
		if err := Database.DB.Model(user).Update("role", "admin").Error; err != nil {
			return nil, "", fmt.Errorf("更新用户角色失败: %v", err)
		}
		user.Role = "admin"
	}

	// 生成token
	jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
	token, err := jwtUtils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %v", err)
	}

	return user, token, nil
}

// CreateAdminUserWithToken 创建管理员用户并返回token
func (th *TokenHelper) CreateAdminUserWithToken() (*Models.User, string, error) {
	return th.CreateTestUserWithToken("admin", "admin@test.com", "admin123", "admin")
}

// CreateNormalUserWithToken 创建普通用户并返回token
func (th *TokenHelper) CreateNormalUserWithToken() (*Models.User, string, error) {
	return th.CreateTestUserWithToken("user", "user@test.com", "user123", "user")
}

// GenerateTokenForUser 为现有用户生成token
// 功能说明：
// 1. 为已存在的用户生成token
// 2. 支持自定义过期时间
// 3. 验证用户状态
func (th *TokenHelper) GenerateTokenForUser(user *Models.User) (string, error) {
	// 检查用户状态
	if !user.IsActive() {
		return "", fmt.Errorf("用户账户已禁用")
	}

	// 生成token
	jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
	token, err := jwtUtils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return "", fmt.Errorf("生成token失败: %v", err)
	}

	return token, nil
}

// ValidateToken 验证token有效性
// 功能说明：
// 1. 验证token格式和签名
// 2. 检查token是否过期
// 3. 返回用户信息
func (th *TokenHelper) ValidateToken(token string) (*Utils.Claims, error) {
	// 处理Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证token
	jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
	claims, err := jwtUtils.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("token验证失败: %v", err)
	}

	return claims, nil
}

// CreateExpiredToken 创建已过期的token（用于测试）
func (th *TokenHelper) CreateExpiredToken(user *Models.User) (string, error) {
	// 使用自定义的过期时间（过去的时间）
	claims := &Utils.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1小时前过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // 2小时前签发
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := Config.GetConfig().JWT.Secret
	return token.SignedString([]byte(secret))
}

// CreateInvalidToken 创建无效的token（用于测试）
func (th *TokenHelper) CreateInvalidToken() string {
	return "invalid.token.here"
}

// GetBearerToken 获取带Bearer前缀的token
func (th *TokenHelper) GetBearerToken(token string) string {
	return "Bearer " + token
}

// CreateMultipleUsersWithTokens 创建多个测试用户并返回token
// 功能说明：
// 1. 批量创建测试用户
// 2. 为每个用户生成token
// 3. 返回用户和token映射
func (th *TokenHelper) CreateMultipleUsersWithTokens(count int) (map[string]*UserTokenInfo, error) {
	users := make(map[string]*UserTokenInfo)

	for i := 1; i <= count; i++ {
		username := fmt.Sprintf("testuser%d", i)
		email := fmt.Sprintf("testuser%d@example.com", i)
		password := "password123"

		user, token, err := th.CreateTestUserWithToken(username, email, password, "user")
		if err != nil {
			return nil, fmt.Errorf("创建用户 %d 失败: %v", i, err)
		}

		users[username] = &UserTokenInfo{
			User:  user,
			Token: token,
		}
	}

	return users, nil
}

// UserTokenInfo 用户Token信息
type UserTokenInfo struct {
	User  *Models.User
	Token string
}

// CleanupTestUsers 清理测试用户
func (th *TokenHelper) CleanupTestUsers() error {
	// 清理所有测试用户
	if err := Database.DB.Where("email LIKE ?", "%@test.com").Delete(&Models.User{}).Error; err != nil {
		return fmt.Errorf("清理测试用户失败: %v", err)
	}

	if err := Database.DB.Where("email LIKE ?", "%@example.com").Delete(&Models.User{}).Error; err != nil {
		return fmt.Errorf("清理测试用户失败: %v", err)
	}

	return nil
}

// GetTestTokenHeaders 获取测试用的请求头
// 功能说明：
// 1. 为HTTP测试提供标准的认证头
// 2. 支持不同的认证方式
// 3. 返回完整的请求头映射
func (th *TokenHelper) GetTestTokenHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
}

// GetAdminTokenHeaders 获取管理员token请求头
func (th *TokenHelper) GetAdminTokenHeaders() (map[string]string, error) {
	admin, token, err := th.CreateAdminUserWithToken()
	if err != nil {
		return nil, err
	}

	headers := th.GetTestTokenHeaders(token)
	headers["X-Test-User-ID"] = fmt.Sprintf("%d", admin.ID)
	headers["X-Test-User-Role"] = admin.Role

	return headers, nil
}

// GetUserTokenHeaders 获取普通用户token请求头
func (th *TokenHelper) GetUserTokenHeaders() (map[string]string, error) {
	user, token, err := th.CreateNormalUserWithToken()
	if err != nil {
		return nil, err
	}

	headers := th.GetTestTokenHeaders(token)
	headers["X-Test-User-ID"] = fmt.Sprintf("%d", user.ID)
	headers["X-Test-User-Role"] = user.Role

	return headers, nil
}

// GenerateTestToken 为测试用户生成token
// 功能说明：
// 1. 为测试用户数据生成有效的JWT token
// 2. 支持自定义用户信息
// 3. 返回可用于测试的token字符串
func (th *TokenHelper) GenerateTestToken(userData map[string]interface{}) string {
	// 创建测试用的Claims
	claims := &Utils.Claims{
		UserID:   1, // 测试用户ID
		Username: userData["username"].(string),
		Email:    userData["email"].(string),
		Role:     userData["role"].(string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := Config.GetConfig().JWT.Secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		// 如果生成失败，返回一个测试用的假token
		return "test-token-" + userData["username"].(string)
	}

	return tokenString
}
