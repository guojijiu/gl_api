package Controllers

import (
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
//
// 重要功能说明：
// 1. 用户注册：支持用户名、邮箱、密码注册，包含密码强度验证
// 2. 用户登录：JWT token认证，支持登录状态记录和统计
// 3. 用户登出：token黑名单机制，支持多设备登录控制
// 4. 资料管理：用户资料查看和更新，支持头像上传
// 5. 密码管理：密码重置、邮箱验证、token刷新
// 6. 安全控制：账户状态检查、登录失败处理、会话管理
//
// 安全特性：
// - 密码使用bcrypt安全哈希存储
// - JWT token支持过期时间和自动刷新
// - 登录失败次数限制和账户锁定
// - 邮箱验证和密码重置的安全流程
// - 支持token黑名单，防止token滥用
//
// 性能优化：
// - 使用Redis缓存用户会话和token黑名单
// - 数据库查询优化，减少不必要的查询
// - 支持密码强度实时验证
//
// 错误处理：
// - 统一的错误响应格式
// - 详细的错误日志记录
// - 用户友好的错误消息
// - 支持多语言错误提示
//
// 使用注意事项：
// - 注册和登录接口有速率限制保护
// - 密码重置需要邮箱验证
// - 邮箱验证是可选的，但建议启用
// - 支持多设备同时登录
type AuthController struct {
	authService *Services.AuthService
}

// NewAuthController 创建认证控制器
// 功能说明：
// 1. 初始化认证控制器实例
// 2. 创建AuthService服务实例
// 3. 返回配置好的控制器对象
func NewAuthController() *AuthController {
	return &AuthController{
		authService: Services.NewAuthService(),
	}
}

// Register 用户注册
// 功能说明：
// 1. 接收用户注册请求（用户名、邮箱、密码）
// 2. 验证请求数据的有效性
// 3. 调用AuthService进行用户注册
// 4. 返回注册结果和用户信息（不含密码）
// 5. 支持用户名和邮箱唯一性检查
// 6. 自动对密码进行安全哈希处理
// 7. 增强的输入验证和密码强度检查
func (c *AuthController) Register(ctx *gin.Context) {
	var request Requests.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	// 增强的请求验证
	if validationErrors := request.Validate(); len(validationErrors) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求验证失败",
			"errors":  validationErrors,
		})
		return
	}

	user, err := c.authService.Register(request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "注册失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "注册成功",
		"data":    user,
	})
}

// Login 用户登录
// 功能说明：
// 1. 接收用户登录请求（用户名、密码）
// 2. 验证用户凭据的有效性
// 3. 检查用户账户状态（是否被禁用）
// 4. 生成JWT token用于后续认证
// 5. 更新用户最后登录时间和登录次数
// 6. 返回token和用户信息
//
// 安全措施：
// - 密码验证使用bcrypt安全哈希比较
// - 检查用户状态防止禁用账户登录
// - 更新登录时间用于安全审计
// - 生成安全的JWT token
// - 清除返回数据中的敏感信息（密码）
//
// 错误处理：
// - 用户名不存在时返回统一错误信息（不泄露用户存在性）
// - 密码错误时返回统一错误信息
// - 账户被禁用时明确提示
// - 数据库错误时记录详细日志
//
// 性能考虑：
// - 使用数据库索引优化用户名查询
// - 密码验证使用常量时间比较
// - 最小化数据库查询次数
func (c *AuthController) Login(ctx *gin.Context) {
	var request Requests.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	token, user, err := c.authService.Login(request)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "登录失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// Logout 用户登出
// 功能说明：
// 1. 处理用户登出请求
// 2. 将当前token加入黑名单
// 3. 清理用户会话数据
// 4. 记录登出日志
// 5. 需要有效的JWT token才能访问
func (c *AuthController) Logout(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	err := c.authService.Logout(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "登出失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// GetProfile 获取用户资料
// 功能说明：
// 1. 获取当前登录用户的详细资料
// 2. 排除敏感信息（如密码）
// 3. 返回完整的用户信息
// 4. 需要有效的JWT token才能访问
// 5. 支持用户ID、用户名、邮箱、头像等信息
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	user, err := c.authService.GetProfile(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户不存在",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// UpdateProfile 更新用户资料
// 功能说明：
// 1. 更新当前登录用户的资料信息
// 2. 支持更新用户名、邮箱、头像等字段
// 3. 验证数据唯一性（用户名、邮箱不能重复）
// 4. 需要有效的JWT token才能访问
// 5. 返回更新后的用户信息
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	var request Requests.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	user, err := c.authService.UpdateProfile(userID, request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    user,
	})
}

// RefreshToken 刷新Token
// 功能说明：
// 1. 刷新当前JWT token的有效期
// 2. 验证当前token的有效性
// 3. 检查用户是否仍然存在且有效
// 4. 生成新的token，延长有效期
// 5. 用于保持用户登录状态
// 6. 支持Bearer token格式
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Token不能为空",
		})
		return
	}

	newToken, err := c.authService.RefreshToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Token刷新失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token刷新成功",
		"data": gin.H{
			"token": newToken,
		},
	})
}

// RequestPasswordReset 请求密码重置
// 功能说明：
// 1. 接收密码重置请求（邮箱地址）
// 2. 验证邮箱是否存在且有效
// 3. 生成密码重置token
// 4. 发送重置邮件到用户邮箱
// 5. 记录重置请求日志
// 6. 支持邮件发送失败重试机制
func (c *AuthController) RequestPasswordReset(ctx *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "邮箱格式无效",
			"error":   err.Error(),
		})
		return
	}

	err := c.authService.RequestPasswordReset(request.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "密码重置请求失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码重置邮件已发送",
	})
}

// ResetPassword 重置密码
// 功能说明：
// 1. 接收密码重置请求（重置token、新密码）
// 2. 验证重置token的有效性
// 3. 更新用户密码（安全哈希处理）
// 4. 清除重置token
// 5. 记录密码重置操作日志
// 6. 支持密码强度验证
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	err := c.authService.ResetPassword(request.Token, request.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "密码重置失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码重置成功",
	})
}

// SendEmailVerification 发送邮箱验证
// 功能说明：
// 1. 为当前登录用户发送邮箱验证邮件
// 2. 生成邮箱验证token
// 3. 发送验证邮件到用户邮箱
// 4. 记录验证请求日志
// 5. 需要有效的JWT token才能访问
// 6. 支持邮件发送失败重试机制
func (c *AuthController) SendEmailVerification(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	err := c.authService.SendEmailVerification(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "邮箱验证邮件发送失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "邮箱验证邮件已发送",
	})
}

// VerifyEmail 验证邮箱
// 功能说明：
// 1. 接收邮箱验证请求（验证token）
// 2. 验证token的有效性
// 3. 更新用户邮箱验证状态
// 4. 记录验证操作日志
// 5. 清除验证token
// 6. 支持验证状态查询
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "验证token无效",
			"error":   err.Error(),
		})
		return
	}

	err := c.authService.VerifyEmail(request.Token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "邮箱验证失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "邮箱验证成功",
	})
}
