package Utils

import (
	"cloud-platform-api/app/Config"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTUtils JWT工具类
type JWTUtils struct {
	config *Config.JWTConfig
}

// NewJWTUtils 创建JWT工具实例
func NewJWTUtils(config *Config.JWTConfig) *JWTUtils {
	return &JWTUtils{
		config: config,
	}
}

// 全局JWT工具实例
var globalJWTUtils *JWTUtils

// SetGlobalJWTUtils 设置全局JWT工具实例
func SetGlobalJWTUtils(utils *JWTUtils) {
	globalJWTUtils = utils
}

// GetGlobalJWTUtils 获取全局JWT工具实例
func GetGlobalJWTUtils() *JWTUtils {
	if globalJWTUtils == nil {
		// 如果没有设置全局实例，尝试从配置创建
		config := Config.GetJWTConfig()
		if config != nil {
			globalJWTUtils = NewJWTUtils(config)
		}
	}
	return globalJWTUtils
}

// GenerateToken 生成JWT令牌（全局函数）
func GenerateToken(userID uint, username, email, role string) (string, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return "", errors.New("JWT工具未初始化")
	}
	return utils.GenerateToken(userID, username, email, role)
}

// ValidateToken 验证JWT令牌（全局函数）
func ValidateToken(tokenString string) (*Claims, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return nil, errors.New("JWT工具未初始化")
	}
	return utils.ValidateToken(tokenString)
}

// GeneratePasswordResetToken 生成密码重置令牌（全局函数）
func GeneratePasswordResetToken(userID uint) (string, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return "", errors.New("JWT工具未初始化")
	}
	return utils.GeneratePasswordResetToken(userID)
}

// ValidatePasswordResetToken 验证密码重置令牌（全局函数）
func ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return nil, errors.New("JWT工具未初始化")
	}
	return utils.ValidatePasswordResetToken(tokenString)
}

// GenerateEmailVerificationToken 生成邮箱验证令牌（全局函数）
func GenerateEmailVerificationToken(userID uint) (string, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return "", errors.New("JWT工具未初始化")
	}
	return utils.GenerateEmailVerificationToken(userID)
}

// ValidateEmailVerificationToken 验证邮箱验证令牌（全局函数）
func ValidateEmailVerificationToken(tokenString string) (*Claims, error) {
	utils := GetGlobalJWTUtils()
	if utils == nil {
		return nil, errors.New("JWT工具未初始化")
	}
	return utils.ValidateEmailVerificationToken(tokenString)
}

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
//
// 功能说明：
// 1. 生成包含用户信息的JWT令牌
// 2. 使用HS256算法签名
// 3. 设置过期时间、签发时间、生效时间等标准声明
// 4. 包含用户ID、用户名、邮箱、角色等自定义声明
//
// 令牌结构：
// - Header: 算法类型（HS256）
// - Payload: 用户信息和标准声明
// - Signature: 使用密钥签名的哈希值
//
// 标准声明（Registered Claims）：
// - ExpiresAt: 过期时间（从配置读取）
// - IssuedAt: 签发时间（当前时间）
// - NotBefore: 生效时间（当前时间）
// - Issuer: 签发者（从配置读取）
// - Subject: 主题（用户ID）
//
// 自定义声明（Custom Claims）：
// - UserID: 用户ID
// - Username: 用户名
// - Email: 邮箱
// - Role: 用户角色
//
// 安全考虑：
// - 使用HS256算法，密钥必须保密
// - 过期时间应该合理设置（通常1-24小时）
// - 密钥应该足够复杂，防止被破解
// - 令牌不应该包含敏感信息（如密码）
//
// 使用场景：
// - 用户登录后生成访问令牌
// - 用于后续API请求的身份验证
// - 支持无状态认证
//
// 注意事项：
// - 密钥泄露会导致所有令牌失效
// - 过期时间过短会影响用户体验
// - 过期时间过长会增加安全风险
// - 令牌应该通过HTTPS传输
func (j *JWTUtils) GenerateToken(userID uint, username, email, role string) (string, error) {
	// 设置过期时间
	// 从配置读取过期小时数，默认24小时
	expirationTime := time.Now().Add(time.Duration(j.config.ExpirationHours) * time.Hour)

	// 创建JWT声明（Claims）
	// 包含用户信息和标准JWT声明
	claims := &Claims{
		UserID:   userID,   // 用户ID
		Username: username, // 用户名
		Email:    email,    // 邮箱
		Role:     role,     // 用户角色
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),    // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),    // 生效时间
			Issuer:    j.config.Issuer,                    // 签发者
			Subject:   fmt.Sprintf("%d", userID),          // 主题（用户ID）
		},
	}

	// 创建JWT令牌
	// 使用HS256算法（HMAC-SHA256）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名令牌
	// 签名用于验证令牌的完整性和真实性
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("签名令牌失败: %v", err)
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
//
// 功能说明：
// 1. 解析JWT令牌字符串
// 2. 验证签名方法和签名有效性
// 3. 验证令牌的过期时间和生效时间
// 4. 返回令牌中的用户信息
//
// 验证步骤：
// 1. 解析令牌：将字符串解析为JWT对象
// 2. 验证签名方法：确保使用HS256算法
// 3. 验证签名：使用密钥验证签名是否有效
// 4. 验证过期时间：检查令牌是否已过期
// 5. 验证生效时间：检查令牌是否已生效
//
// 安全验证：
// - 签名验证：确保令牌未被篡改
// - 过期验证：防止使用过期令牌
// - 生效验证：防止使用未生效令牌
// - 算法验证：防止算法替换攻击
//
// 错误处理：
// - 解析失败：返回解析错误
// - 签名无效：返回签名错误
// - 已过期：返回过期错误
// - 未生效：返回未生效错误
//
// 使用场景：
// - 中间件验证请求中的JWT令牌
// - API端点验证用户身份
// - 刷新令牌前验证当前令牌
//
// 注意事项：
// - 验证失败应该返回明确的错误信息
// - 不应该泄露详细的验证失败原因（安全考虑）
// - 过期令牌应该被拒绝，不能刷新
// - 验证过程应该快速，避免影响性能
func (j *JWTUtils) ValidateToken(tokenString string) (*Claims, error) {
	// 解析令牌
	// ParseWithClaims会解析令牌并验证签名
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		// 确保使用HS256算法，防止算法替换攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		// 返回密钥，用于验证签名
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %v", err)
	}

	// 验证令牌有效性
	// token.Valid会检查签名是否有效
	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	// 获取声明
	// 将Claims转换为自定义的Claims类型
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无法解析令牌声明")
	}

	// 验证过期时间
	// 如果令牌已过期，拒绝使用
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("令牌已过期")
	}

	// 验证生效时间
	// 如果令牌尚未生效，拒绝使用
	if claims.NotBefore != nil && time.Now().Before(claims.NotBefore.Time) {
		return nil, errors.New("令牌尚未生效")
	}

	// 验证通过，返回声明
	return claims, nil
}

// RefreshToken 刷新JWT令牌
func (j *JWTUtils) RefreshToken(tokenString string) (string, error) {
	// 验证当前令牌
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("验证令牌失败: %v", err)
	}

	// 检查是否在刷新窗口内
	refreshWindow := time.Duration(j.config.RefreshWindowHours) * time.Hour
	if time.Until(claims.ExpiresAt.Time) > refreshWindow {
		return "", errors.New("令牌尚未到刷新时间")
	}

	// 生成新令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// ExtractClaims 提取令牌声明（不验证签名）
func (j *JWTUtils) ExtractClaims(tokenString string) (*Claims, error) {
	// 解析令牌但不验证签名
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %v", err)
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无法解析令牌声明")
	}

	return claims, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTUtils) IsTokenExpired(tokenString string) bool {
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenExpiration 获取令牌过期时间
func (j *JWTUtils) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("令牌没有过期时间")
	}

	return claims.ExpiresAt.Time, nil
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTUtils) GenerateRefreshToken(userID uint) (string, error) {
	// 刷新令牌使用更长的过期时间
	expirationTime := time.Now().Add(time.Duration(j.config.RefreshTokenExpirationDays) * 24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("refresh_%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("签名刷新令牌失败: %v", err)
	}

	return tokenString, nil
}

// ValidateRefreshToken 验证刷新令牌
func (j *JWTUtils) ValidateRefreshToken(tokenString string) (uint, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	// 检查是否是刷新令牌
	if !j.isRefreshToken(claims) {
		return 0, errors.New("不是有效的刷新令牌")
	}

	return claims.UserID, nil
}

// isRefreshToken 检查是否是刷新令牌
func (j *JWTUtils) isRefreshToken(claims *Claims) bool {
	return claims.Subject != "" &&
		len(claims.Subject) > 8 &&
		claims.Subject[:8] == "refresh_"
}

// BlacklistToken 将令牌加入黑名单
func (j *JWTUtils) BlacklistToken(tokenString string) error {
	// 这里可以集成Redis或其他存储来维护黑名单
	// 示例实现：将令牌的哈希值存储到黑名单中
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return err
	}

	// 计算令牌哈希
	tokenHash := j.hashToken(tokenString)

	// 存储到黑名单（这里需要集成实际的存储）
	_ = tokenHash
	_ = claims

	return nil
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func (j *JWTUtils) IsTokenBlacklisted(tokenString string) bool {
	// 这里可以集成Redis或其他存储来检查黑名单
	// 示例实现：检查令牌哈希是否在黑名单中
	tokenHash := j.hashToken(tokenString)

	// 检查黑名单（这里需要集成实际的存储）
	_ = tokenHash

	return false
}

// hashToken 计算令牌哈希
func (j *JWTUtils) hashToken(tokenString string) string {
	// 使用SHA256计算令牌哈希
	// 这里简化实现，实际应该使用更安全的方法
	return fmt.Sprintf("%x", tokenString)
}

// GeneratePasswordResetToken 生成密码重置令牌
func (j *JWTUtils) GeneratePasswordResetToken(userID uint) (string, error) {
	// 设置较短的过期时间（1小时）
	expirationTime := time.Now().Add(1 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: "",
		Email:    "",
		Role:     "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("password_reset_%d", userID),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 获取密钥
	secretKey := j.config.SecretKey
	if secretKey == "" {
		secretKey = j.config.Secret
	}

	// 签名令牌
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("生成密码重置令牌失败: %v", err)
	}

	return tokenString, nil
}

// ValidatePasswordResetToken 验证密码重置令牌
func (j *JWTUtils) ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 获取密钥
		secretKey := j.config.SecretKey
		if secretKey == "" {
			secretKey = j.config.Secret
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查是否是密码重置令牌
		if !strings.HasPrefix(claims.Subject, "password_reset_") {
			return nil, errors.New("invalid password reset token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateEmailVerificationToken 生成邮箱验证令牌
func (j *JWTUtils) GenerateEmailVerificationToken(userID uint) (string, error) {
	// 设置较短的过期时间（24小时）
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: "",
		Email:    "",
		Role:     "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("email_verification_%d", userID),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 获取密钥
	secretKey := j.config.SecretKey
	if secretKey == "" {
		secretKey = j.config.Secret
	}

	// 签名令牌
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("生成邮箱验证令牌失败: %v", err)
	}

	return tokenString, nil
}

// ValidateEmailVerificationToken 验证邮箱验证令牌
func (j *JWTUtils) ValidateEmailVerificationToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 获取密钥
		secretKey := j.config.SecretKey
		if secretKey == "" {
			secretKey = j.config.Secret
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查是否是邮箱验证令牌
		if !strings.HasPrefix(claims.Subject, "email_verification_") {
			return nil, errors.New("invalid email verification token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetTokenInfo 获取令牌信息
func (j *JWTUtils) GetTokenInfo(tokenString string) (map[string]interface{}, error) {
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"email":      claims.Email,
		"role":       claims.Role,
		"issued_at":  claims.IssuedAt.Time,
		"expires_at": claims.ExpiresAt.Time,
		"issuer":     claims.Issuer,
		"subject":    claims.Subject,
	}

	return info, nil
}
