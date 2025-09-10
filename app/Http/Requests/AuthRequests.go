package Requests

import (
	"regexp"
	"strings"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// Validate 验证注册请求
func (r *RegisterRequest) Validate() []string {
	var errors []string

	// 用户名验证
	if len(r.Username) < 3 || len(r.Username) > 50 {
		errors = append(errors, "用户名长度必须在3-50个字符之间")
	}
	
	// 用户名格式验证（只允许字母、数字、下划线）
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(r.Username) {
		errors = append(errors, "用户名只能包含字母、数字和下划线")
	}

	// 邮箱验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		errors = append(errors, "邮箱格式不正确")
	}

	// 密码强度验证
	if len(r.Password) < 6 {
		errors = append(errors, "密码长度至少6个字符")
	}
	
	// 检查密码是否包含数字和字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.Password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(r.Password)
	if !hasLetter || !hasDigit {
		errors = append(errors, "密码必须包含字母和数字")
	}

	return errors
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Validate 验证登录请求
func (r *LoginRequest) Validate() []string {
	var errors []string

	if strings.TrimSpace(r.Username) == "" {
		errors = append(errors, "用户名不能为空")
	}

	if strings.TrimSpace(r.Password) == "" {
		errors = append(errors, "密码不能为空")
	}

	return errors
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// Validate 验证更新资料请求
func (r *UpdateProfileRequest) Validate() []string {
	var errors []string

	// 用户名验证
	if r.Username != "" {
		if len(r.Username) < 3 || len(r.Username) > 50 {
			errors = append(errors, "用户名长度必须在3-50个字符之间")
		}
		
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !usernameRegex.MatchString(r.Username) {
			errors = append(errors, "用户名只能包含字母、数字和下划线")
		}
	}

	// 邮箱验证
	if r.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(r.Email) {
			errors = append(errors, "邮箱格式不正确")
		}
	}

	// 头像URL验证
	if r.Avatar != "" {
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(r.Avatar) {
			errors = append(errors, "头像URL格式不正确")
		}
	}

	return errors
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// Validate 验证密码重置请求
func (r *PasswordResetRequest) Validate() []string {
	var errors []string

	if strings.TrimSpace(r.Token) == "" {
		errors = append(errors, "重置令牌不能为空")
	}

	if len(r.NewPassword) < 6 {
		errors = append(errors, "新密码长度至少6个字符")
	}
	
	// 检查密码强度
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.NewPassword)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(r.NewPassword)
	if !hasLetter || !hasDigit {
		errors = append(errors, "新密码必须包含字母和数字")
	}

	return errors
}

// PasswordChangeRequest 密码修改请求
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}

// Validate 验证密码修改请求
func (r *PasswordChangeRequest) Validate() []string {
	var errors []string

	// 验证当前密码
	if strings.TrimSpace(r.CurrentPassword) == "" {
		errors = append(errors, "当前密码不能为空")
	}

	// 验证新密码
	if len(r.NewPassword) < 8 {
		errors = append(errors, "新密码长度至少8个字符")
	}
	
	// 检查密码强度
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.NewPassword)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(r.NewPassword)
	if !hasLetter || !hasDigit {
		errors = append(errors, "新密码必须包含字母和数字")
	}

	// 检查新密码是否与当前密码相同
	if r.CurrentPassword == r.NewPassword {
		errors = append(errors, "新密码不能与当前密码相同")
	}

	return errors
}

// EmailVerificationRequest 邮箱验证请求
type EmailVerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

// Validate 验证邮箱验证请求
func (r *EmailVerificationRequest) Validate() []string {
	var errors []string

	// 验证token
	if strings.TrimSpace(r.Token) == "" {
		errors = append(errors, "验证token不能为空")
	}

	// 检查token长度（JWT token通常很长）
	if len(r.Token) < 50 {
		errors = append(errors, "验证token格式不正确")
	}

	return errors
}
