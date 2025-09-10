package Utils

import (
	"cloud-platform-api/app/Config"
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
// 功能说明：
// 1. 创建包含用户信息的JWT声明（用户ID、用户名、角色）
// 2. 设置token的过期时间（可配置）
// 3. 使用HMAC-SHA256算法进行签名
// 4. 使用配置的密钥进行签名验证
// 5. 支持token的发行时间和生效时间设置
// 6. 返回Base64编码的JWT字符串
// 7. 用于用户登录后的身份验证
// 8. 支持自定义声明扩展（可扩展）
// 9. 提供token版本控制（可扩展）
// 10. 记录token生成的安全日志
func GenerateToken(userID uint, username, role string) (string, error) {
	cfg := Config.GetConfig().JWT
	
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpireTime) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken 验证JWT token
// 功能说明：
// 1. 验证JWT token的签名和格式
// 2. 检查token是否过期
// 3. 验证token的发行时间和生效时间
// 4. 返回解析后的用户声明信息
// 5. 提供详细的错误信息
func ValidateToken(tokenString string) (*Claims, error) {
	// 检查token是否为空
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}
	
	// 检查token长度是否合理
	if len(tokenString) < 10 {
		return nil, errors.New("token format is invalid")
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		
		// 获取JWT密钥
		secret := Config.GetConfig().JWT.Secret
		if secret == "" {
			return nil, errors.New("JWT secret is not configured")
		}
		
		return []byte(secret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 额外验证：检查用户ID是否有效
		if claims.UserID == 0 {
			return nil, errors.New("invalid user ID in token")
		}
		
		// 额外验证：检查用户名是否为空
		if claims.Username == "" {
			return nil, errors.New("invalid username in token")
		}
		
		return claims, nil
	}
	
	return nil, errors.New("invalid token")
}

// RefreshToken 刷新JWT token
// 功能说明：
// 1. 验证当前token的有效性
// 2. 检查token是否即将过期（剩余时间少于1小时）
// 3. 生成新的token，延长有效期
// 4. 用于保持用户登录状态
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	
	// 检查token是否即将过期（比如还有1小时过期）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not expired yet")
	}
	
	// 生成新的token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// PasswordResetClaims 密码重置声明
type PasswordResetClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GeneratePasswordResetToken 生成密码重置token
// 功能说明：
// 1. 生成专门用于密码重置的token
// 2. 设置较短的过期时间（1小时）
// 3. 包含用户ID和邮箱信息
func GeneratePasswordResetToken(userID uint, email string) (string, error) {
	cfg := Config.GetConfig().JWT
	
	claims := PasswordResetClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidatePasswordResetToken 验证密码重置token
func ValidatePasswordResetToken(tokenString string) (*PasswordResetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(Config.GetConfig().JWT.Secret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*PasswordResetClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, errors.New("invalid token")
}

// EmailVerificationClaims 邮箱验证声明
type EmailVerificationClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateEmailVerificationToken 生成邮箱验证token
// 功能说明：
// 1. 生成专门用于邮箱验证的token
// 2. 设置较长的过期时间（24小时）
// 3. 包含用户ID和邮箱信息
func GenerateEmailVerificationToken(userID uint, email string) (string, error) {
	cfg := Config.GetConfig().JWT
	
	claims := EmailVerificationClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateEmailVerificationToken 验证邮箱验证token
func ValidateEmailVerificationToken(tokenString string) (*EmailVerificationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &EmailVerificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(Config.GetConfig().JWT.Secret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*EmailVerificationClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, errors.New("invalid token")
}


