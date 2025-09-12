package Utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// PasswordUtils 密码工具类
type PasswordUtils struct {
	// 密码强度配置
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSpecial   bool
	SpecialChars     string
}

// NewPasswordUtils 创建密码工具实例
func NewPasswordUtils() *PasswordUtils {
	return &PasswordUtils{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
		SpecialChars:     "!@#$%^&*()_+-=[]{}|;:,.<>?",
	}
}

// 全局密码工具实例
var globalPasswordUtils = NewPasswordUtils()

// HashPassword 哈希密码（全局函数）
func HashPassword(password string) (string, error) {
	return globalPasswordUtils.HashPassword(password)
}

// CheckPassword 检查密码（全局函数）
func CheckPassword(password, encodedHash string) bool {
	valid, _ := globalPasswordUtils.VerifyPassword(password, encodedHash)
	return valid
}

// ValidatePasswordStrength 验证密码强度（全局函数）
func ValidatePasswordStrength(password string) (bool, []string) {
	return globalPasswordUtils.ValidatePasswordStrength(password)
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig 默认密码配置
var DefaultPasswordConfig = &PasswordConfig{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword 哈希密码
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	// 生成随机盐
	salt, err := p.generateRandomBytes(DefaultPasswordConfig.SaltLength)
	if err != nil {
		return "", fmt.Errorf("生成盐失败: %v", err)
	}

	// 使用Argon2id哈希密码
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		DefaultPasswordConfig.Iterations,
		DefaultPasswordConfig.Memory,
		DefaultPasswordConfig.Parallelism,
		DefaultPasswordConfig.KeyLength,
	)

	// 编码为base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 返回格式化的哈希字符串
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultPasswordConfig.Memory,
		DefaultPasswordConfig.Iterations,
		DefaultPasswordConfig.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword 验证密码
func (p *PasswordUtils) VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析哈希字符串
	config, salt, hash, err := p.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同参数哈希输入密码
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// 使用constant time比较防止时序攻击
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// ValidatePasswordStrength 验证密码强度
func (p *PasswordUtils) ValidatePasswordStrength(password string) (bool, []string) {
	var errors []string

	// 检查长度
	if len(password) < p.MinLength {
		errors = append(errors, fmt.Sprintf("密码长度至少需要%d个字符", p.MinLength))
	}

	// 检查大写字母
	if p.RequireUppercase && !p.hasUppercase(password) {
		errors = append(errors, "密码必须包含至少一个大写字母")
	}

	// 检查小写字母
	if p.RequireLowercase && !p.hasLowercase(password) {
		errors = append(errors, "密码必须包含至少一个小写字母")
	}

	// 检查数字
	if p.RequireNumbers && !p.hasNumbers(password) {
		errors = append(errors, "密码必须包含至少一个数字")
	}

	// 检查特殊字符
	if p.RequireSpecial && !p.hasSpecialChars(password) {
		errors = append(errors, fmt.Sprintf("密码必须包含至少一个特殊字符: %s", p.SpecialChars))
	}

	return len(errors) == 0, errors
}

// CalculatePasswordStrength 计算密码强度分数
func (p *PasswordUtils) CalculatePasswordStrength(password string) int {
	score := 0

	// 长度分数
	length := len(password)
	if length >= 12 {
		score += 25
	} else if length >= 8 {
		score += 15
	} else if length >= 6 {
		score += 5
	}

	// 字符类型分数
	if p.hasUppercase(password) {
		score += 15
	}
	if p.hasLowercase(password) {
		score += 15
	}
	if p.hasNumbers(password) {
		score += 15
	}
	if p.hasSpecialChars(password) {
		score += 20
	}

	// 复杂度分数
	if p.hasRepeatedChars(password) {
		score -= 10
	}
	if p.hasSequentialChars(password) {
		score -= 10
	}

	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// GeneratePassword 生成安全密码
func (p *PasswordUtils) GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"

	password := make([]byte, length)
	for i := range password {
		randomBytes, err := p.generateRandomBytes(1)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomBytes[0]%byte(len(charset))]
	}

	return string(password), nil
}

// CheckPasswordHistory 检查密码历史（防止重复使用）
func (p *PasswordUtils) CheckPasswordHistory(password string, history []string) bool {
	for _, oldPassword := range history {
		if match, _ := p.VerifyPassword(password, oldPassword); match {
			return false
		}
	}
	return true
}

// 辅助方法

// generateRandomBytes 生成随机字节
func (p *PasswordUtils) generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// decodeHash 解析哈希字符串
func (p *PasswordUtils) decodeHash(encodedHash string) (config *PasswordConfig, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("无效的哈希格式")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("不支持的哈希算法")
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("不兼容的版本")
	}

	config = &PasswordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	config.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	config.KeyLength = uint32(len(hash))

	return config, salt, hash, nil
}

// hasUppercase 检查是否包含大写字母
func (p *PasswordUtils) hasUppercase(password string) bool {
	for _, char := range password {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

// hasLowercase 检查是否包含小写字母
func (p *PasswordUtils) hasLowercase(password string) bool {
	for _, char := range password {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

// hasNumbers 检查是否包含数字
func (p *PasswordUtils) hasNumbers(password string) bool {
	for _, char := range password {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// hasSpecialChars 检查是否包含特殊字符
func (p *PasswordUtils) hasSpecialChars(password string) bool {
	for _, char := range password {
		if strings.ContainsRune(p.SpecialChars, char) {
			return true
		}
	}
	return false
}

// hasRepeatedChars 检查是否有重复字符
func (p *PasswordUtils) hasRepeatedChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

// hasSequentialChars 检查是否有连续字符
func (p *PasswordUtils) hasSequentialChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i+1] == password[i]+1 && password[i+2] == password[i]+2 {
			return true
		}
	}
	return false
}

// ValidatePasswordPolicy 验证密码策略
func (p *PasswordUtils) ValidatePasswordPolicy(password string, policy map[string]interface{}) (bool, []string) {
	var errors []string

	// 检查最小长度
	if minLength, ok := policy["min_length"].(int); ok {
		if len(password) < minLength {
			errors = append(errors, fmt.Sprintf("密码长度至少需要%d个字符", minLength))
		}
	}

	// 检查最大长度
	if maxLength, ok := policy["max_length"].(int); ok {
		if len(password) > maxLength {
			errors = append(errors, fmt.Sprintf("密码长度不能超过%d个字符", maxLength))
		}
	}

	// 检查是否包含用户名
	if preventUsername, ok := policy["prevent_username"].(bool); ok && preventUsername {
		if username, ok := policy["username"].(string); ok {
			if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
				errors = append(errors, "密码不能包含用户名")
			}
		}
	}

	// 检查常见密码
	if preventCommon, ok := policy["prevent_common"].(bool); ok && preventCommon {
		commonPasswords := []string{
			"password", "123456", "123456789", "qwerty", "abc123",
			"password123", "admin", "root", "user", "guest",
		}
		for _, common := range commonPasswords {
			if strings.EqualFold(password, common) {
				errors = append(errors, "不能使用常见密码")
				break
			}
		}
	}

	return len(errors) == 0, errors
}

// GetPasswordStrengthLevel 获取密码强度等级
func (p *PasswordUtils) GetPasswordStrengthLevel(password string) string {
	score := p.CalculatePasswordStrength(password)

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
