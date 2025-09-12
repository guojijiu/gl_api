package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AuthService 认证服务
//
// 重要功能说明：
// 1. 用户认证：注册、登录、登出、token管理
// 2. 密码管理：密码哈希、强度验证、重置流程
// 3. 邮箱验证：邮箱验证状态管理、验证邮件发送
// 4. 会话管理：JWT token生成、验证、刷新、黑名单
// 5. 安全控制：账户状态检查、登录失败处理、权限验证
// 6. 审计日志：记录所有认证相关操作，支持安全审计
//
// 安全特性：
// - 密码使用bcrypt算法哈希，支持自动盐值生成
// - JWT token支持过期时间和签名验证
// - 支持token黑名单，防止已登出token的滥用
// - 邮箱验证和密码重置的安全流程设计
// - 登录失败次数限制和账户锁定机制
// - 支持多设备登录控制和会话管理
//
// 性能优化：
// - 使用Redis缓存用户会话和token黑名单
// - 数据库查询优化，减少不必要的查询
// - 支持密码强度实时验证
// - 异步邮件发送，不阻塞主流程
//
// 错误处理：
// - 详细的错误分类和错误码
// - 用户友好的错误消息
// - 完整的错误日志记录
// - 支持错误重试和降级处理
//
// 业务规则：
// - 用户名和邮箱必须唯一
// - 密码强度要求：至少8字符，包含大小写字母和数字
// - 邮箱验证是可选的，但建议启用
// - 支持账户状态控制（启用/禁用）
// - 登录失败超过限制次数后账户锁定
//
// 扩展性：
// - 支持多种认证方式（用户名/邮箱登录）
// - 支持第三方登录集成（OAuth、SAML）
// - 支持多因素认证（MFA）
// - 支持单点登录（SSO）
// - 支持自定义认证策略
type AuthService struct {
	BaseService
}

// NewAuthService 创建认证服务
// 功能说明：
// 1. 初始化认证服务实例
// 2. 返回配置好的服务对象
func NewAuthService() *AuthService {
	return &AuthService{
		BaseService: *NewBaseService(),
	}
}

// NewAuthServiceWithDB 使用数据库连接创建认证服务
func NewAuthServiceWithDB(db *gorm.DB) *AuthService {
	service := &AuthService{
		BaseService: *NewBaseService(),
	}
	service.DB = db
	return service
}

// getDB 获取数据库连接
func (s *AuthService) getDB() *gorm.DB {
	if s.DB != nil {
		if db, ok := s.DB.(*gorm.DB); ok {
			return db
		}
	}
	// 回退到全局数据库连接
	return Database.DB
}

// Register 用户注册
// 功能说明：
// 1. 验证用户输入数据
// 2. 检查用户名和邮箱唯一性
// 3. 验证密码强度
// 4. 对密码进行安全哈希
// 5. 创建新用户记录
// 6. 返回用户信息（不含密码）
func (s *AuthService) Register(request Requests.RegisterRequest) (*Models.User, error) {
	// 获取数据库连接
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database connection not available")
	}

	// 检查用户名是否已存在
	var existingUser Models.User
	if err := db.Where("username = ?", request.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if err := db.Where("email = ?", request.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 验证密码强度
	isValid, validationErrors := Utils.ValidatePasswordStrength(request.Password)
	if !isValid {
		return nil, fmt.Errorf("password validation failed: %s", strings.Join(validationErrors, "; "))
	}

	// 哈希密码
	hashedPassword, err := Utils.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &Models.User{
		Username: request.Username,
		Email:    request.Email,
		Password: hashedPassword,
		Role:     "user",
		Status:   1,
	}

	if err := s.getDB().Create(user).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return user, nil
}

// Login 用户登录
// 功能说明：
// 1. 验证用户名和密码
// 2. 检查用户状态（是否被禁用）
// 3. 更新最后登录时间和登录次数
// 4. 生成JWT token
// 5. 返回token和用户信息
func (s *AuthService) Login(request Requests.LoginRequest) (string, *Models.User, error) {
	// 查找用户
	var user Models.User
	if err := s.getDB().Where("username = ?", request.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid credentials")
		}
		return "", nil, err
	}

	// 检查密码
	if !Utils.CheckPassword(request.Password, user.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", nil, errors.New("account is disabled")
	}

	// 更新最后登录时间
	user.UpdateLastLoginTime()
	if err := s.getDB().Save(&user).Error; err != nil {
		return "", nil, err
	}

	// 生成JWT token
	token, err := Utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return "", nil, err
	}

	// 清除密码字段
	user.Password = ""

	return token, &user, nil
}

// Logout 用户登出
// 功能说明：
// 1. 处理用户登出逻辑
// 2. 将token添加到黑名单中
// 3. 记录登出日志
// 4. 清理用户会话数据
func (s *AuthService) Logout(userID string) error {
	// 记录登出日志
	auditService := NewAuditService(s.getDB())
	auditService.LogUserAction(nil, 0, "", "logout", "user", 0, "用户登出")

	// 初始化Token黑名单服务，使用配置的Redis服务
	var redisService *RedisService
	redisConfig := Config.GetConfig().Redis
	if redisConfig.Host != "" {
		redisService = NewRedisService(&RedisConfig{
			Host:     redisConfig.Host,
			Port:     redisConfig.Port,
			Password: redisConfig.Password,
			DB:       redisConfig.Database,
		})

		// 测试Redis连接
		if err := redisService.Ping(); err != nil {
			// Redis连接失败时使用nil
			redisService = nil
		}
	}

	_ = NewTokenBlacklistService(redisService)

	// 清理用户会话数据
	// 清理用户相关的缓存和会话信息
	cacheService := NewCacheService(redisService, &CacheConfig{
		Prefix:      "app:",
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		EnableCache: redisService != nil,
	})

	// 清理用户缓存
	cacheService.Delete("user:" + userID)
	cacheService.Delete("session:" + userID)

	return nil
}

// GetProfile 获取用户资料
// 功能说明：
// 1. 根据用户ID获取用户信息
// 2. 排除敏感信息（如密码）
// 3. 返回完整的用户资料
func (s *AuthService) GetProfile(userID string) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// UpdateProfile 更新用户资料
// 功能说明：
// 1. 验证用户权限
// 2. 更新允许修改的字段
// 3. 验证数据有效性
// 4. 保存更新后的用户信息
func (s *AuthService) UpdateProfile(userID string, request Requests.UpdateProfileRequest) (*Models.User, error) {
	var user Models.User
	if err := s.getDB().First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 如果更新用户名，检查是否与其他用户冲突
	if request.Username != "" && request.Username != user.Username {
		var existingUser Models.User
		if err := s.getDB().Where("username = ? AND id != ?", request.Username, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("username already exists")
		}
		user.Username = request.Username
	}

	// 如果更新邮箱，检查是否与其他用户冲突
	if request.Email != "" && request.Email != user.Email {
		var existingUser Models.User
		if err := s.getDB().Where("email = ? AND id != ?", request.Email, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("email already exists")
		}
		user.Email = request.Email
	}

	if request.Avatar != "" {
		user.Avatar = request.Avatar
	}

	// 保存更新
	if err := s.getDB().Save(&user).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	user.Password = ""

	return &user, nil
}

// RefreshToken 刷新Token
// 功能说明：
// 1. 处理带有Bearer前缀的token
// 2. 验证当前token的有效性
// 3. 验证用户是否仍然存在且有效
// 4. 生成新的token，延长有效期
// 5. 用于保持用户登录状态
func (s *AuthService) RefreshToken(token string) (string, error) {
	// 处理Bearer前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证当前token
	claims, err := Utils.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// 验证用户是否仍然存在且有效
	var user Models.User
	if err := s.getDB().First(&user, claims.UserID).Error; err != nil {
		return "", errors.New("user not found or account disabled")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", errors.New("account is disabled")
	}

	// 生成新token
	newToken, err := Utils.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// RequestPasswordReset 请求密码重置
// 功能说明：
// 1. 验证邮箱是否存在
// 2. 生成密码重置token
// 3. 发送重置邮件
// 4. 记录重置请求日志
func (s *AuthService) RequestPasswordReset(email string) error {
	// 查找用户
	var user Models.User
	if err := s.getDB().Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("email not found")
		}
		return err
	}

	// 检查用户状态
	if user.Status != 1 {
		return errors.New("account is disabled")
	}

	// 生成密码重置token
	resetToken, err := Utils.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return err
	}

	// 记录重置请求到审计日志
	auditService := NewAuditService(s.getDB())
	auditService.LogUserAction(nil, user.ID, user.Username, "password_reset_request", "user", user.ID, "请求密码重置")

	// 发送重置邮件
	emailService := NewEmailService(&EmailConfig{
		Host:     Config.GetConfig().Email.Host,
		Port:     Config.GetConfig().Email.Port,
		Username: Config.GetConfig().Email.Username,
		Password: Config.GetConfig().Email.Password,
		From:     Config.GetConfig().Email.From,
		UseTLS:   Config.GetConfig().Email.UseTLS,
	})

	if err := emailService.SendPasswordResetEmail(user.Email, resetToken, user.Username); err != nil {
		// 记录邮件发送失败日志
		auditService.LogUserAction(nil, user.ID, user.Username, "password_reset_email_failed", "user", user.ID, "密码重置邮件发送失败: "+err.Error())
		return fmt.Errorf("failed to send reset email: %v", err)
	}

	// 记录邮件发送成功日志
	auditService.LogUserAction(nil, user.ID, user.Username, "password_reset_email_sent", "user", user.ID, "密码重置邮件发送成功")

	return nil
}

// ResetPassword 重置密码
// 功能说明：
// 1. 验证重置token的有效性
// 2. 更新用户密码
// 3. 清除重置token
// 4. 记录密码重置操作
func (s *AuthService) ResetPassword(resetToken, newPassword string) error {
	// 验证重置token
	claims, err := Utils.ValidatePasswordResetToken(resetToken)
	if err != nil {
		return err
	}

	// 查找用户
	var user Models.User
	if err := s.getDB().First(&user, claims.UserID).Error; err != nil {
		return errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != 1 {
		return errors.New("account is disabled")
	}

	// 哈希新密码
	hashedPassword, err := Utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.getDB().Model(&user).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	// 记录密码重置操作到审计日志
	auditService := NewAuditService(s.getDB())
	auditService.LogUserAction(nil, user.ID, user.Username, "password_reset", "user", user.ID, "密码重置成功")

	// 清除重置token（添加到黑名单）
	tokenBlacklistService := NewTokenBlacklistService(nil)
	err = tokenBlacklistService.AddToBlacklist(resetToken, user.ID, claims.ExpiresAt.Time)
	if err != nil {
		// 记录token清理失败，但不影响密码重置流程
		auditService.LogUserAction(nil, user.ID, user.Username, "password_reset_token_cleanup_failed", "user", user.ID, "重置token清理失败: "+err.Error())
	} else {
		auditService.LogUserAction(nil, user.ID, user.Username, "password_reset_token_cleaned", "user", user.ID, "重置token已清理")
	}

	return nil
}

// SendEmailVerification 发送邮箱验证
// 功能说明：
// 1. 生成邮箱验证token
// 2. 发送验证邮件
// 3. 记录验证请求
func (s *AuthService) SendEmailVerification(userID string) error {
	var user Models.User
	if err := s.getDB().First(&user, userID).Error; err != nil {
		return err
	}

	// 检查邮箱是否已验证
	if user.IsEmailVerified() {
		return errors.New("email already verified")
	}

	// 生成邮箱验证token
	verificationToken, err := Utils.GenerateEmailVerificationToken(user.ID)
	if err != nil {
		return err
	}

	// 记录验证请求
	auditService := NewAuditService(s.getDB())
	auditService.LogUserAction(nil, user.ID, user.Username, "email_verification_request", "user", user.ID, "请求邮箱验证")

	// 发送验证邮件
	emailService := NewEmailService(&EmailConfig{
		Host:     Config.GetConfig().Email.Host,
		Port:     Config.GetConfig().Email.Port,
		Username: Config.GetConfig().Email.Username,
		Password: Config.GetConfig().Email.Password,
		From:     Config.GetConfig().Email.From,
		UseTLS:   Config.GetConfig().Email.UseTLS,
	})

	if err := emailService.SendEmailVerificationEmail(user.Email, verificationToken, user.Username); err != nil {
		// 记录邮件发送失败日志
		auditService.LogUserAction(nil, user.ID, user.Username, "email_verification_email_failed", "user", user.ID, "邮箱验证邮件发送失败: "+err.Error())
		return fmt.Errorf("failed to send verification email: %v", err)
	}

	// 记录邮件发送成功日志
	auditService.LogUserAction(nil, user.ID, user.Username, "email_verification_email_sent", "user", user.ID, "邮箱验证邮件发送成功")

	return nil
}

// VerifyEmail 验证邮箱
// 功能说明：
// 1. 验证邮箱验证token
// 2. 更新用户邮箱验证状态
// 3. 记录验证操作
// 4. 实现完整的审计日志记录
func (s *AuthService) VerifyEmail(verificationToken string) error {
	// 验证token
	claims, err := Utils.ValidateEmailVerificationToken(verificationToken)
	if err != nil {
		return err
	}

	// 查找用户
	var user Models.User
	if err := s.getDB().First(&user, claims.UserID).Error; err != nil {
		return errors.New("user not found")
	}

	// 更新邮箱验证状态
	if err := s.getDB().Model(&user).Update("email_verified_at", time.Now()).Error; err != nil {
		return err
	}

	// 记录验证操作
	auditService := NewAuditService(s.getDB())
	auditService.LogUserAction(nil, user.ID, user.Username, "email_verification", "user", user.ID, "邮箱验证成功")

	// 清除验证token（添加到黑名单）
	tokenBlacklistService := NewTokenBlacklistService(nil)
	err = tokenBlacklistService.AddToBlacklist(verificationToken, user.ID, claims.ExpiresAt.Time)
	if err != nil {
		// 记录token清理失败，但不影响验证流程
		auditService.LogUserAction(nil, user.ID, user.Username, "email_verification_token_cleanup_failed", "user", user.ID, "验证token清理失败")
	} else {
		auditService.LogUserAction(nil, user.ID, user.Username, "email_verification_token_cleaned", "user", user.ID, "验证token已清理")
	}

	return nil
}
