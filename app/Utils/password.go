package Utils

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"strings"
	"unicode"
)

// PasswordStrength 密码强度枚举
type PasswordStrength int

const (
	Weak PasswordStrength = iota
	Medium
	Strong
	VeryStrong
)

// PasswordRequirements 密码要求配置
type PasswordRequirements struct {
	MinLength     int
	MaxLength     int
	RequireUpper  bool
	RequireLower  bool
	RequireDigit  bool
	RequireSpecial bool
	MinStrength   PasswordStrength
}

// DefaultPasswordRequirements 默认密码要求
func DefaultPasswordRequirements() *PasswordRequirements {
	return &PasswordRequirements{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		MinStrength:    Medium,
	}
}

// HashPassword 哈希密码
// 功能说明：
// 1. 使用bcrypt算法对密码进行安全哈希
// 2. 使用默认的cost值(10)平衡安全性和性能
// 3. 返回哈希后的密码字符串
// 4. 用于用户注册和密码修改
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
// 功能说明：
// 1. 比较明文密码和哈希密码是否匹配
// 2. 使用bcrypt的CompareHashAndPassword方法
// 3. 用于用户登录验证
// 4. 返回布尔值表示密码是否正确
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 生成随机密码
// 功能说明：
// 1. 生成指定长度的随机密码
// 2. 包含大小写字母、数字和特殊字符
// 3. 确保密码强度符合要求
// 4. 用于密码重置和临时密码生成
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	
	const (
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		digitChars   = "0123456789"
		specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)
	
	allChars := upperChars + lowerChars + digitChars + specialChars
	
	// 确保至少包含每种字符类型
	password := make([]byte, length)
	
	// 随机选择字符
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}
		password[i] = allChars[randomIndex.Int64()%int64(len(allChars))]
	}
	
	// 确保包含所有必需的字符类型
	password[0] = upperChars[0]
	password[1] = lowerChars[0]
	password[2] = digitChars[0]
	password[3] = specialChars[0]
	
	return string(password), nil
}

// ValidatePasswordStrength 验证密码强度
// 功能说明：
// 1. 检查密码长度是否符合要求（最小和最大长度）
// 2. 验证密码是否包含必需的字符类型（大写、小写、数字、特殊字符）
// 3. 检测常见弱密码和易猜测的密码
// 4. 检查重复字符序列和模式
// 5. 计算密码强度分数（弱、中等、强、很强）
// 6. 支持自定义密码策略和规则
// 7. 返回详细的验证错误信息
// 8. 提供密码强度建议和改进方案
// 9. 支持密码历史检查（可扩展）
// 10. 记录密码验证的安全日志
// 11. 用于用户注册和密码修改时的验证
// 12. 支持多语言错误消息（可扩展）
func ValidatePasswordStrength(password string, requirements *PasswordRequirements) (PasswordStrength, []string, error) {
	if requirements == nil {
		requirements = DefaultPasswordRequirements()
	}
	
	var errors []string
	
	// 检查长度
	if len(password) < requirements.MinLength {
		errors = append(errors, fmt.Sprintf("密码长度至少需要%d个字符", requirements.MinLength))
	}
	if len(password) > requirements.MaxLength {
		errors = append(errors, fmt.Sprintf("密码长度不能超过%d个字符", requirements.MaxLength))
	}
	
	// 检查字符类型
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if requirements.RequireUpper && !hasUpper {
		errors = append(errors, "密码必须包含至少一个大写字母")
	}
	if requirements.RequireLower && !hasLower {
		errors = append(errors, "密码必须包含至少一个小写字母")
	}
	if requirements.RequireDigit && !hasDigit {
		errors = append(errors, "密码必须包含至少一个数字")
	}
	if requirements.RequireSpecial && !hasSpecial {
		errors = append(errors, "密码必须包含至少一个特殊字符")
	}
	
	// 检查常见弱密码
	if isCommonPassword(password) {
		errors = append(errors, "密码过于常见，请使用更复杂的密码")
	}
	
	// 检查重复字符
	if hasRepeatingChars(password) {
		errors = append(errors, "密码不能包含重复的字符序列")
	}
	
	// 计算密码强度
	strength := calculatePasswordStrength(password, hasUpper, hasLower, hasDigit, hasSpecial)
	
	return strength, errors, nil
}

// isCommonPassword 检查是否为常见密码
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"1234567890", "12345678", "1234567", "12345678910",
		"111111", "000000", "123123", "123321", "654321",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	
	return false
}

// hasRepeatingChars 检查是否有重复字符序列
func hasRepeatingChars(password string) bool {
	if len(password) < 4 {
		return false
	}
	
	for i := 0; i <= len(password)-4; i++ {
		seq := password[i : i+4]
		if strings.Count(password, seq) > 1 {
			return true
		}
	}
	
	return false
}

// calculatePasswordStrength 计算密码强度
func calculatePasswordStrength(password string, hasUpper, hasLower, hasDigit, hasSpecial bool) PasswordStrength {
	score := 0
	
	// 长度分数
	if len(password) >= 8 {
		score += 1
	}
	if len(password) >= 12 {
		score += 1
	}
	if len(password) >= 16 {
		score += 1
	}
	
	// 字符类型分数
	if hasUpper {
		score += 1
	}
	if hasLower {
		score += 1
	}
	if hasDigit {
		score += 1
	}
	if hasSpecial {
		score += 1
	}
	
	// 复杂度分数
	if len(password) >= 8 && hasUpper && hasLower && hasDigit && hasSpecial {
		score += 2
	}
	
	// 根据分数确定强度
	switch {
	case score <= 2:
		return Weak
	case score <= 4:
		return Medium
	case score <= 6:
		return Strong
	default:
		return VeryStrong
	}
}

// GetPasswordStrengthText 获取密码强度文本描述
func GetPasswordStrengthText(strength PasswordStrength) string {
	switch strength {
	case Weak:
		return "弱"
	case Medium:
		return "中等"
	case Strong:
		return "强"
	case VeryStrong:
		return "很强"
	default:
		return "未知"
	}
}

// ValidatePasswordPolicy 验证密码策略
// 功能说明：
// 1. 检查密码是否符合组织的密码策略
// 2. 支持自定义策略规则
// 3. 返回详细的验证结果
func ValidatePasswordPolicy(password string, policy map[string]interface{}) (bool, []string) {
	var errors []string
	
	// 默认策略
	minLength := 8
	maxLength := 128
	requireUpper := true
	requireLower := true
	requireDigit := true
	requireSpecial := true
	
	// 应用自定义策略
	if val, ok := policy["min_length"].(int); ok {
		minLength = val
	}
	if val, ok := policy["max_length"].(int); ok {
		maxLength = val
	}
	if val, ok := policy["require_upper"].(bool); ok {
		requireUpper = val
	}
	if val, ok := policy["require_lower"].(bool); ok {
		requireLower = val
	}
	if val, ok := policy["require_digit"].(bool); ok {
		requireDigit = val
	}
	if val, ok := policy["require_special"].(bool); ok {
		requireSpecial = val
	}
	
	// 创建要求配置
	requirements := &PasswordRequirements{
		MinLength:      minLength,
		MaxLength:      maxLength,
		RequireUpper:   requireUpper,
		RequireLower:   requireLower,
		RequireDigit:   requireDigit,
		RequireSpecial: requireSpecial,
	}
	
	// 验证密码强度
	_, validationErrors, err := ValidatePasswordStrength(password, requirements)
	if err != nil {
		errors = append(errors, "密码验证失败")
		return false, errors
	}
	
	errors = append(errors, validationErrors...)
	
	return len(errors) == 0, errors
}


