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
//
// 功能说明：
// 1. 为测试提供有效的JWT token（简化测试编写）
// 2. 创建测试用户并生成token（自动处理用户创建和token生成）
// 3. 支持不同角色的用户token（管理员、普通用户等）
// 4. 提供token验证功能（验证token有效性）
// 5. 支持批量创建测试用户（并发测试场景）
// 6. 提供测试用的请求头（简化HTTP测试）
//
// 设计目的：
// - 简化测试代码：减少重复的用户创建和token生成代码
// - 提高测试效率：封装常用操作，提高测试编写速度
// - 统一测试标准：确保所有测试使用相同的token生成逻辑
// - 支持多种场景：支持单个用户、批量用户、不同角色等场景
//
// 使用场景：
// - API认证测试：需要token的API接口测试
// - 权限测试：测试不同角色的权限控制
// - 并发测试：批量创建用户进行并发测试
// - 集成测试：需要真实用户和token的集成测试
//
// 注意事项：
// - 创建的测试用户应该在测试后清理
// - Token使用测试环境的JWT密钥生成
// - 批量创建用户时注意数据库性能
// - Token的有效期由JWT配置决定
type TokenHelper struct {
	authService *Services.AuthService // 认证服务，用于用户注册和认证
}

// NewTokenHelper 创建Token辅助工具
//
// 功能说明：
// 1. 初始化TokenHelper实例
// 2. 创建认证服务实例
// 3. 返回可用的TokenHelper
//
// 注意事项：
// - 每个测试套件应该创建一个TokenHelper实例
// - TokenHelper依赖AuthService，需要确保AuthService已初始化
// - TokenHelper是线程安全的，可以在并发测试中使用
func NewTokenHelper() *TokenHelper {
	return &TokenHelper{
		authService: Services.NewAuthService(),
	}
}

// CreateTestUserWithToken 创建测试用户并返回token
//
// 功能说明：
// 1. 创建测试用户（通过AuthService注册）
// 2. 设置用户角色（如果指定了管理员角色）
// 3. 生成有效的JWT token（包含用户信息）
// 4. 返回用户对象和token字符串
//
// 参数说明：
// - username: 用户名（必须唯一）
// - email: 邮箱地址（必须唯一）
// - password: 密码（会被哈希存储）
// - role: 用户角色（"admin"或"user"）
//
// 返回信息：
// - *Models.User: 创建的用户对象（包含ID、用户名、邮箱、角色等）
// - string: JWT token字符串（用于API请求认证）
// - error: 错误信息（如果创建失败）
//
// 执行流程：
// 1. 通过AuthService注册新用户
// 2. 如果角色是"admin"，更新用户角色
// 3. 使用JWT工具生成token（包含用户ID、用户名、邮箱、角色）
// 4. 返回用户对象和token
//
// 使用场景：
// - 需要特定角色的用户进行权限测试
// - 需要自定义用户信息的测试
// - 需要多个不同用户的并发测试
//
// 注意事项：
// - 用户名和邮箱必须唯一，重复创建会失败
// - 密码会被哈希存储，不会返回明文
// - 管理员角色需要单独更新，因为注册时默认是普通用户
// - Token的有效期由JWT配置决定
// - 创建的测试用户应该在测试后清理
func (th *TokenHelper) CreateTestUserWithToken(username, email, password, role string) (*Models.User, string, error) {
	// 创建用户
	// 通过AuthService注册，确保使用正确的业务逻辑
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
	// 注册时默认是普通用户，需要单独更新为管理员
	if role == "admin" {
		if err := Database.DB.Model(user).Update("role", "admin").Error; err != nil {
			return nil, "", fmt.Errorf("更新用户角色失败: %v", err)
		}
		user.Role = "admin"
	}

	// 生成token
	// 使用JWT工具生成包含用户信息的token
	// Token包含用户ID、用户名、邮箱、角色等信息
	jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
	token, err := jwtUtils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %v", err)
	}

	return user, token, nil
}

// CreateAdminUserWithToken 创建管理员用户并返回token
//
// 功能说明：
// 1. 创建默认的管理员用户（用户名：admin）
// 2. 生成管理员角色的JWT token
// 3. 返回管理员用户对象和token
//
// 使用场景：
// - 需要管理员权限的API测试
// - 权限控制相关的测试
// - 管理员功能的集成测试
//
// 注意事项：
// - 默认用户名是"admin"，如果已存在会失败
// - 默认邮箱是"admin@test.com"，如果已存在会失败
// - 创建的测试用户应该在测试后清理
func (th *TokenHelper) CreateAdminUserWithToken() (*Models.User, string, error) {
	return th.CreateTestUserWithToken("admin", "admin@test.com", "admin123", "admin")
}

// CreateNormalUserWithToken 创建普通用户并返回token
//
// 功能说明：
// 1. 创建默认的普通用户（用户名：user）
// 2. 生成普通用户角色的JWT token
// 3. 返回普通用户对象和token
//
// 使用场景：
// - 需要普通用户权限的API测试
// - 用户功能相关的测试
// - 权限限制相关的测试
//
// 注意事项：
// - 默认用户名是"user"，如果已存在会失败
// - 默认邮箱是"user@test.com"，如果已存在会失败
// - 创建的测试用户应该在测试后清理
func (th *TokenHelper) CreateNormalUserWithToken() (*Models.User, string, error) {
	return th.CreateTestUserWithToken("user", "user@test.com", "user123", "user")
}

// GenerateTokenForUser 为现有用户生成token
//
// 功能说明：
// 1. 为已存在的用户生成token（不创建新用户）
// 2. 验证用户状态（确保用户账户未被禁用）
// 3. 生成包含用户信息的JWT token
//
// 使用场景：
// - 为已存在的用户生成token进行测试
// - 测试用户状态验证逻辑
// - 测试token生成功能
//
// 参数说明：
// - user: 已存在的用户对象（必须有效且未被禁用）
//
// 返回信息：
// - string: JWT token字符串
// - error: 错误信息（如果用户被禁用或生成失败）
//
// 注意事项：
// - 用户必须存在且未被禁用
// - Token的有效期由JWT配置决定
// - 不会修改用户信息，只生成token
func (th *TokenHelper) GenerateTokenForUser(user *Models.User) (string, error) {
	// 检查用户状态
	// 确保用户账户未被禁用，只有活跃用户才能生成token
	if !user.IsActive() {
		return "", fmt.Errorf("用户账户已禁用")
	}

	// 生成token
	// 使用JWT工具生成包含用户信息的token
	jwtUtils := Utils.NewJWTUtils(&Config.GetConfig().JWT)
	token, err := jwtUtils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return "", fmt.Errorf("生成token失败: %v", err)
	}

	return token, nil
}

// ValidateToken 验证token有效性
//
// 功能说明：
// 1. 验证token格式和签名（确保token未被篡改）
// 2. 检查token是否过期（确保token仍然有效）
// 3. 返回token中的用户信息（Claims）
//
// 参数说明：
// - token: JWT token字符串（可以带或不带"Bearer "前缀）
//
// 返回信息：
// - *Utils.Claims: token中的用户信息（用户ID、用户名、邮箱、角色等）
// - error: 错误信息（如果token无效、过期或签名错误）
//
// 验证流程：
// 1. 处理Bearer前缀（如果存在则移除）
// 2. 验证token签名（确保token未被篡改）
// 3. 检查token是否过期（确保token仍然有效）
// 4. 解析token中的Claims（提取用户信息）
//
// 使用场景：
// - 测试token验证逻辑
// - 测试token过期处理
// - 测试token签名验证
//
// 注意事项：
// - Token必须使用正确的JWT密钥签名
// - Token必须未过期
// - Token格式必须正确
func (th *TokenHelper) ValidateToken(token string) (*Utils.Claims, error) {
	// 处理Bearer前缀
	// 如果token以"Bearer "开头，移除前缀
	// 支持两种格式：带前缀和不带前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证token
	// 使用JWT工具验证token的有效性
	// 包括签名验证、过期检查等
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
//
// 功能说明：
// 1. 批量创建测试用户（指定数量）
// 2. 为每个用户生成token（自动生成）
// 3. 返回用户和token映射（以用户名为key）
//
// 参数说明：
// - count: 要创建的用户数量（必须大于0）
//
// 返回信息：
// - map[string]*UserTokenInfo: 用户名到用户信息的映射
// - error: 错误信息（如果创建失败）
//
// 用户命名规则：
// - 用户名：testuser1, testuser2, ..., testuserN
// - 邮箱：testuser1@example.com, testuser2@example.com, ..., testuserN@example.com
// - 密码：统一为"password123"
// - 角色：统一为"user"（普通用户）
//
// 使用场景：
// - 并发测试：需要多个用户进行并发测试
// - 压力测试：需要大量用户进行压力测试
// - 批量操作测试：测试批量用户操作功能
//
// 注意事项：
// - 创建的用户数量应该合理，避免过多影响性能
// - 所有用户使用相同的密码，便于测试
// - 创建的用户应该在测试后清理
// - 如果某个用户创建失败，会返回错误并停止创建
func (th *TokenHelper) CreateMultipleUsersWithTokens(count int) (map[string]*UserTokenInfo, error) {
	// 创建用户映射
	// 使用用户名作为key，便于查找
	users := make(map[string]*UserTokenInfo)

	// 循环创建指定数量的用户
	for i := 1; i <= count; i++ {
		// 生成唯一的用户名和邮箱
		username := fmt.Sprintf("testuser%d", i)
		email := fmt.Sprintf("testuser%d@example.com", i)
		password := "password123"

		// 创建用户并生成token
		user, token, err := th.CreateTestUserWithToken(username, email, password, "user")
		if err != nil {
			return nil, fmt.Errorf("创建用户 %d 失败: %v", i, err)
		}

		// 保存用户信息到映射
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
//
// 功能说明：
// 1. 删除所有测试用户（通过邮箱模式匹配）
// 2. 清理测试数据，防止污染数据库
// 3. 支持两种邮箱模式（@test.com和@example.com）
//
// 清理规则：
// - 删除所有邮箱以"@test.com"结尾的用户
// - 删除所有邮箱以"@example.com"结尾的用户
// - 使用软删除（如果模型支持）或硬删除
//
// 使用场景：
// - 测试套件清理时调用
// - 测试完成后清理测试数据
// - 防止测试数据影响后续测试
//
// 注意事项：
// - 只删除测试用户，不会删除生产用户
// - 删除操作不可逆，请谨慎使用
// - 如果测试用户被其他表引用，可能需要先删除关联数据
// - 建议在测试套件的TearDownSuite中调用
func (th *TokenHelper) CleanupTestUsers() error {
	// 清理所有测试用户
	// 通过邮箱模式匹配，删除所有测试用户
	// 使用LIKE查询匹配邮箱模式
	if err := Database.DB.Where("email LIKE ?", "%@test.com").Delete(&Models.User{}).Error; err != nil {
		return fmt.Errorf("清理测试用户失败: %v", err)
	}

	// 清理所有使用@example.com邮箱的测试用户
	// 支持多种测试邮箱模式
	if err := Database.DB.Where("email LIKE ?", "%@example.com").Delete(&Models.User{}).Error; err != nil {
		return fmt.Errorf("清理测试用户失败: %v", err)
	}

	return nil
}

// GetTestTokenHeaders 获取测试用的请求头
//
// 功能说明：
// 1. 为HTTP测试提供标准的认证头（包含Bearer token）
// 2. 设置标准的Content-Type和Accept头
// 3. 返回完整的请求头映射（可直接用于HTTP请求）
//
// 参数说明：
// - token: JWT token字符串（会自动添加"Bearer "前缀）
//
// 返回信息：
// - map[string]string: 请求头映射，包含：
//   - Authorization: Bearer token（用于认证）
//   - Content-Type: application/json（请求体类型）
//   - Accept: application/json（响应类型）
//
// 使用场景：
// - HTTP API测试：需要认证的API请求
// - 集成测试：需要完整请求头的测试
// - 单元测试：模拟HTTP请求
//
// 注意事项：
// - Token会自动添加"Bearer "前缀
// - 请求头符合RESTful API标准
// - 可以直接用于gin.HTTPTest或http.NewRequest
func (th *TokenHelper) GetTestTokenHeaders(token string) map[string]string {
	// 返回标准的HTTP请求头
	// 包含认证、内容类型和接受类型
	return map[string]string{
		"Authorization": "Bearer " + token, // Bearer token认证
		"Content-Type":  "application/json", // JSON请求体
		"Accept":        "application/json", // JSON响应
	}
}

// GetAdminTokenHeaders 获取管理员token请求头
//
// 功能说明：
// 1. 创建管理员用户并生成token
// 2. 生成包含管理员token的请求头
// 3. 添加额外的测试头（用户ID和角色）
//
// 返回信息：
// - map[string]string: 请求头映射，包含：
//   - Authorization: Bearer token（管理员token）
//   - Content-Type: application/json
//   - Accept: application/json
//   - X-Test-User-ID: 用户ID（用于测试）
//   - X-Test-User-Role: 用户角色（admin）
// - error: 错误信息（如果创建失败）
//
// 使用场景：
// - 需要管理员权限的API测试
// - 权限控制相关的测试
// - 管理员功能的集成测试
//
// 注意事项：
// - 每次调用都会创建新的管理员用户
// - 创建的用户应该在测试后清理
// - X-Test-User-ID和X-Test-User-Role是测试专用头
func (th *TokenHelper) GetAdminTokenHeaders() (map[string]string, error) {
	// 创建管理员用户并生成token
	admin, token, err := th.CreateAdminUserWithToken()
	if err != nil {
		return nil, err
	}

	// 获取标准请求头
	headers := th.GetTestTokenHeaders(token)
	
	// 添加测试专用的请求头
	// 便于测试代码获取用户信息
	headers["X-Test-User-ID"] = fmt.Sprintf("%d", admin.ID)
	headers["X-Test-User-Role"] = admin.Role

	return headers, nil
}

// GetUserTokenHeaders 获取普通用户token请求头
//
// 功能说明：
// 1. 创建普通用户并生成token
// 2. 生成包含普通用户token的请求头
// 3. 添加额外的测试头（用户ID和角色）
//
// 返回信息：
// - map[string]string: 请求头映射，包含：
//   - Authorization: Bearer token（普通用户token）
//   - Content-Type: application/json
//   - Accept: application/json
//   - X-Test-User-ID: 用户ID（用于测试）
//   - X-Test-User-Role: 用户角色（user）
// - error: 错误信息（如果创建失败）
//
// 使用场景：
// - 需要普通用户权限的API测试
// - 用户功能相关的测试
// - 权限限制相关的测试
//
// 注意事项：
// - 每次调用都会创建新的普通用户
// - 创建的用户应该在测试后清理
// - X-Test-User-ID和X-Test-User-Role是测试专用头
func (th *TokenHelper) GetUserTokenHeaders() (map[string]string, error) {
	// 创建普通用户并生成token
	user, token, err := th.CreateNormalUserWithToken()
	if err != nil {
		return nil, err
	}

	// 获取标准请求头
	headers := th.GetTestTokenHeaders(token)
	
	// 添加测试专用的请求头
	// 便于测试代码获取用户信息
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
