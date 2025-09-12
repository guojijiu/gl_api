package Config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret                     string `mapstructure:"secret"`
	SecretKey                  string `mapstructure:"secret_key"`
	ExpireTime                 int    `mapstructure:"expire_time"`                   // 过期时间（小时）
	ExpirationHours            int    `mapstructure:"expiration_hours"`              // 过期时间（小时）
	Issuer                     string `mapstructure:"issuer"`                        // 签发者
	RefreshWindowHours         int    `mapstructure:"refresh_window_hours"`          // 刷新窗口时间（小时）
	RefreshTokenExpirationDays int    `mapstructure:"refresh_token_expiration_days"` // 刷新令牌过期时间（天）
}

// SetDefaults 设置JWT配置默认值
// 功能说明：
// 1. 不设置JWT密钥默认值，强制用户配置
// 2. 设置token过期时间的默认值（24小时）
// 3. 确保JWT配置有合理的默认值
func (j *JWTConfig) SetDefaults() {
	// 不设置JWT密钥默认值，强制用户通过环境变量配置
	// 这样可以避免使用不安全的默认密钥
	viper.SetDefault("jwt.expire_time", 24)
	viper.SetDefault("jwt.expiration_hours", 24)
	viper.SetDefault("jwt.issuer", "cloud-platform-api")
	viper.SetDefault("jwt.refresh_window_hours", 168) // 7天
	viper.SetDefault("jwt.refresh_token_expiration_days", 30)
}

// BindEnvs 绑定JWT环境变量
// 功能说明：
// 1. 绑定JWT_SECRET环境变量到jwt.secret配置
// 2. 绑定JWT_EXPIRE_TIME环境变量到jwt.expire_time配置
// 3. 支持通过环境变量覆盖配置文件中的JWT设置
func (j *JWTConfig) BindEnvs() {
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.secret_key", "JWT_SECRET_KEY")
	viper.BindEnv("jwt.expire_time", "JWT_EXPIRE_TIME")
	viper.BindEnv("jwt.expiration_hours", "JWT_EXPIRATION_HOURS")
	viper.BindEnv("jwt.issuer", "JWT_ISSUER")
	viper.BindEnv("jwt.refresh_window_hours", "JWT_REFRESH_WINDOW_HOURS")
	viper.BindEnv("jwt.refresh_token_expiration_days", "JWT_REFRESH_TOKEN_EXPIRATION_DAYS")
}

// GetJWTConfig 获取JWT配置
// 功能说明：
// 1. 返回全局JWT配置实例
// 2. 如果全局配置未初始化则返回nil
// 3. 用于获取JWT相关的配置信息
func GetJWTConfig() *JWTConfig {
	if globalConfig == nil {
		return nil
	}
	return &globalConfig.JWT
}

// GetExpireDuration 获取过期时间间隔
// 功能说明：
// 1. 将配置的过期时间（小时）转换为Duration类型
// 2. 用于JWT token的过期时间计算
// 3. 返回time.Duration格式的过期时间
func (j *JWTConfig) GetExpireDuration() time.Duration {
	// 优先使用 ExpirationHours，如果没有则使用 ExpireTime
	hours := j.ExpirationHours
	if hours == 0 {
		hours = j.ExpireTime
	}
	return time.Duration(hours) * time.Hour
}

// GetExpireTime 获取过期时间（秒）
// 功能说明：
// 1. 将配置的过期时间（小时）转换为秒数
// 2. 用于JWT token的过期时间计算
// 3. 返回int64格式的过期时间（秒）
func (j *JWTConfig) GetExpireTime() int64 {
	// 优先使用 ExpirationHours，如果没有则使用 ExpireTime
	hours := j.ExpirationHours
	if hours == 0 {
		hours = j.ExpireTime
	}
	return int64(hours * 3600)
}

// Validate 验证JWT配置
func (j *JWTConfig) Validate() error {
	// 优先使用 SecretKey，如果没有则使用 Secret
	secret := j.SecretKey
	if secret == "" {
		secret = j.Secret
	}

	if secret == "" {
		return fmt.Errorf("JWT密钥未配置，请设置JWT_SECRET或JWT_SECRET_KEY环境变量")
	}

	// 检查是否为不安全的默认密钥
	if j.isDefaultSecret() {
		return fmt.Errorf("JWT密钥使用了不安全的默认值，请设置一个强密钥")
	}

	// 检查密钥长度
	if len(secret) < 32 {
		return fmt.Errorf("JWT密钥长度不足，建议至少32个字符，当前长度: %d", len(secret))
	}

	// 检查密钥复杂度
	if err := j.validateSecretComplexity(); err != nil {
		return err
	}

	// 检查密钥是否包含常见弱密钥
	if j.isWeakSecret() {
		return fmt.Errorf("JWT密钥过于简单，请使用更复杂的密钥")
	}

	if j.ExpireTime <= 0 {
		return fmt.Errorf("JWT过期时间配置无效")
	}

	if j.ExpireTime > 8760 { // 超过一年
		return fmt.Errorf("JWT过期时间过长，建议不超过8760小时（一年）")
	}

	return nil
}

// validateSecretComplexity 验证密钥复杂度
func (j *JWTConfig) validateSecretComplexity() error {
	secret := j.SecretKey
	if secret == "" {
		secret = j.Secret
	}

	// 检查是否包含大小写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(secret)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(secret)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(secret)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(secret)

	complexityScore := 0
	if hasUpper {
		complexityScore++
	}
	if hasLower {
		complexityScore++
	}
	if hasDigit {
		complexityScore++
	}
	if hasSpecial {
		complexityScore++
	}

	// 至少需要3种字符类型
	if complexityScore < 3 {
		return fmt.Errorf("JWT密钥复杂度不足，建议包含大小写字母、数字和特殊字符中的至少3种")
	}

	// 检查重复字符
	if j.hasRepeatedChars(secret) {
		return fmt.Errorf("JWT密钥包含过多重复字符，建议使用更随机的密钥")
	}

	return nil
}

// hasRepeatedChars 检查是否包含过多重复字符
func (j *JWTConfig) hasRepeatedChars(secret string) bool {
	charCount := make(map[rune]int)
	for _, char := range secret {
		charCount[char]++
		if charCount[char] > len(secret)/4 { // 同一字符超过总长度的1/4
			return true
		}
	}
	return false
}

// isWeakSecret 检查是否为弱密钥
func (j *JWTConfig) isWeakSecret() bool {
	secret := j.SecretKey
	if secret == "" {
		secret = j.Secret
	}
	secret = strings.ToLower(secret)

	// 常见弱密钥列表（只检查完整的弱密钥，不检查包含关系）
	weakSecrets := []string{
		"password", "123456", "admin", "secret", "key", "token",
		"jwt", "auth", "login", "user", "test", "demo",
		"default", "changeme", "password123", "admin123",
		"qwerty", "abc123", "123456789", "password1",
		"welcome", "hello", "world", "example", "sample",
		"your-secret-key", "changeme", "password",
	}

	// 检查是否为完整的弱密钥
	for _, weak := range weakSecrets {
		if secret == weak {
			return true
		}
	}

	// 检查是否为纯数字
	if regexp.MustCompile(`^[0-9]+$`).MatchString(secret) {
		return true
	}

	// 检查是否为纯字母
	if regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(secret) {
		return true
	}

	// 检查是否为连续字符
	if j.isSequential(secret) {
		return true
	}

	return false
}

// isSequential 检查是否为连续字符
func (j *JWTConfig) isSequential(secret string) bool {
	if len(secret) < 3 {
		return false
	}

	// 检查数字序列
	if regexp.MustCompile(`123|234|345|456|567|678|789|890|987|876|765|654|543|432|321|210`).MatchString(secret) {
		return true
	}

	// 检查字母序列
	if regexp.MustCompile(`abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz`).MatchString(secret) {
		return true
	}

	return false
}

// IsSecretDefault 检查是否为默认密钥
func (j *JWTConfig) IsSecretDefault() bool {
	return j.isDefaultSecret()
}

// isDefaultSecret 检查是否为不安全的默认密钥
func (j *JWTConfig) isDefaultSecret() bool {
	secret := j.SecretKey
	if secret == "" {
		secret = j.Secret
	}

	defaultSecrets := []string{
		"your-secret-key",
		"your-super-secret-jwt-key-change-in-production",
		"your-super-secret-jwt-key-change-in-production-must-be-at-least-32-characters-long",
		"change-in-production",
		"jwt-secret-key",
		"default-jwt-secret",
		"secret-key",
		"jwt-secret",
		"api-secret",
		"app-secret",
	}

	for _, defaultSecret := range defaultSecrets {
		if secret == defaultSecret {
			return true
		}
	}

	return false
}

// GenerateSecureSecret 生成安全的JWT密钥
func (j *JWTConfig) GenerateSecureSecret() (string, error) {
	// 生成64字节的随机密钥（更安全）
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %v", err)
	}

	// 转换为十六进制字符串
	secret := hex.EncodeToString(bytes)

	// 添加一些特殊字符增加复杂度
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	secret = secret + string(specialChars[bytes[0]%uint8(len(specialChars))])

	return secret, nil
}

// ValidateProductionConfig 验证生产环境配置
func (j *JWTConfig) ValidateProductionConfig() error {
	// 检查环境变量
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	if env == "production" || env == "prod" {
		// 生产环境额外验证
		if j.isDefaultSecret() {
			return fmt.Errorf("生产环境不能使用默认JWT密钥，请设置强密钥")
		}

		// 生产环境要求更高的密钥强度
		if len(j.Secret) < 64 {
			return fmt.Errorf("生产环境JWT密钥长度不足，建议至少64个字符，当前长度: %d", len(j.Secret))
		}

		// 检查密钥强度评分
		strength := j.GetSecretStrength()
		if strength < 80 {
			return fmt.Errorf("生产环境JWT密钥强度不足，当前强度: %d/100，建议至少80分", strength)
		}
	}

	return j.Validate()
}

// GetSecretStrength 获取密钥强度评分
func (j *JWTConfig) GetSecretStrength() int {
	secret := j.SecretKey
	if secret == "" {
		secret = j.Secret
	}
	score := 0

	// 长度评分 (0-40分)
	length := len(secret)
	if length >= 64 {
		score += 40
	} else if length >= 48 {
		score += 30
	} else if length >= 32 {
		score += 20
	} else if length >= 16 {
		score += 10
	}

	// 复杂度评分 (0-40分)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(secret)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(secret)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(secret)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(secret)

	complexityScore := 0
	if hasUpper {
		complexityScore++
	}
	if hasLower {
		complexityScore++
	}
	if hasDigit {
		complexityScore++
	}
	if hasSpecial {
		complexityScore++
	}

	score += complexityScore * 10

	// 随机性评分 (0-20分)
	if !j.hasRepeatedChars(secret) && !j.isSequential(secret) && !j.isWeakSecret() {
		score += 20
	} else if !j.isWeakSecret() {
		score += 10
	}

	return score
}

// GetSecretStrengthText 获取密钥强度文本描述
func (j *JWTConfig) GetSecretStrengthText() string {
	score := j.GetSecretStrength()

	switch {
	case score >= 80:
		return "非常强"
	case score >= 60:
		return "强"
	case score >= 40:
		return "中等"
	case score >= 20:
		return "弱"
	default:
		return "非常弱"
	}
}
