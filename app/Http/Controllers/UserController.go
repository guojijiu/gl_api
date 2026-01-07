package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"cloud-platform-api/app/Utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
//
// 功能说明：
// 1. 处理用户相关的HTTP请求
// 2. 提供用户CRUD操作（创建、读取、更新、删除）
// 3. 实现权限控制（管理员和普通用户）
// 4. 处理用户列表查询（支持分页、筛选、搜索）
// 5. 管理用户文章列表
//
// 权限控制：
// - 管理员：可以查看、更新、删除所有用户
// - 普通用户：只能查看和更新自己的信息
//
// 安全特性：
// - 所有敏感信息（密码）都会被清除
// - 用户名和邮箱唯一性验证
// - 删除用户前检查关联数据
// - 记录所有操作到审计日志
//
// 使用场景：
// - 用户管理界面
// - 用户资料编辑
// - 用户列表查询
// - 用户删除操作
type UserController struct {
	Controller
	userService *Services.UserService // 用户服务（处理业务逻辑）
}

// NewUserController 创建用户控制器
//
// 功能说明：
// 1. 初始化用户控制器实例
// 2. 创建UserService服务实例
// 3. 返回配置好的控制器对象
//
// 使用场景：
// - 路由注册时创建控制器实例
// - 测试环境中创建控制器实例
//
// 注意事项：
// - 控制器是无状态的，可以安全地共享
// - 服务实例在控制器创建时初始化
func NewUserController() *UserController {
	return &UserController{
		userService: Services.NewUserService(),
	}
}

// GetUsers 获取用户列表
// 功能说明：
// 1. 获取所有用户列表
// 2. 支持分页查询
// 3. 支持按角色、状态筛选
// 4. 支持按用户名、邮箱搜索
// 5. 支持按注册时间、最后登录时间排序
// 6. 仅管理员可访问
func (c *UserController) GetUsers(ctx *gin.Context) {
	// 权限检查：只有管理员可以查看用户列表
	userRole := ctx.GetString("user_role")
	if userRole != "admin" {
		c.Forbidden(ctx, "只有管理员可以查看用户列表")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	role := ctx.Query("role")
	status := ctx.Query("status")
	search := ctx.Query("search")
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	sortOrder := ctx.DefaultQuery("sort_order", "desc")

	// 构建查询条件
	query := Database.DB.Model(&Models.User{})

	// 角色筛选
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 搜索功能
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 排序
	orderClause := sortBy + " " + sortOrder
	query = query.Order(orderClause)

	// 分页
	offset := (page - 1) * limit
	var users []Models.User
	var total int64

	query.Count(&total)
	query.Offset(offset).Limit(limit).Find(&users)

	// 清除敏感信息
	for i := range users {
		users[i].Password = ""
	}

	// 返回结果
	c.Success(ctx, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, "用户列表获取成功")
}

// GetUser 获取单个用户
// 功能说明：
// 1. 根据ID获取用户详细信息
// 2. 权限控制：只能查看自己的信息或管理员可以查看所有用户
// 3. 返回用户的基本信息和统计
// 4. 包含用户的文章列表
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 获取当前用户信息
	currentUserID := ctx.GetString("user_id")
	currentUserRole := ctx.GetString("user_role")

	// 权限检查：只能查看自己的信息或管理员可以查看所有用户
	if currentUserRole != "admin" && currentUserID != idStr {
		c.Forbidden(ctx, "只能查看自己的用户信息")
		return
	}

	// 查找用户
	var user Models.User
	if err := Database.DB.Preload("Posts").First(&user, id).Error; err != nil {
		c.NotFound(ctx, "用户不存在")
		return
	}

	// 清除敏感信息
	user.Password = ""

	// 获取用户统计信息
	var postCount int64
	Database.DB.Model(&Models.Post{}).Where("user_id = ?", id).Count(&postCount)

	// 构建响应数据
	response := gin.H{
		"user": user,
		"stats": gin.H{
			"post_count": postCount,
		},
	}

	c.Success(ctx, response, "用户信息获取成功")
}

// UpdateUser 更新用户
// 功能说明：
// 1. 更新用户的基本信息
// 2. 支持更新用户名、邮箱、头像等字段
// 3. 验证用户权限（只能更新自己的信息或管理员可以更新所有用户）
// 4. 检查字段唯一性约束
// 5. 支持密码更新（需要验证原密码）
func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 获取当前用户信息
	currentUserID := ctx.GetString("user_id")
	currentUserRole := ctx.GetString("user_role")

	// 权限检查：只能更新自己的信息或管理员可以更新所有用户
	if currentUserRole != "admin" && currentUserID != idStr {
		c.Forbidden(ctx, "只能更新自己的用户信息")
		return
	}

	// 解析请求数据
	var request struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Avatar      string `json:"avatar"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
		Status      *int   `json:"status"`
		Role        string `json:"role"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}

	// 查找用户
	var user Models.User
	if err := Database.DB.First(&user, id).Error; err != nil {
		c.NotFound(ctx, "用户不存在")
		return
	}

	// 检查用户名唯一性（排除自己）
	if request.Username != "" && request.Username != user.Username {
		var existingUser Models.User
		if err := Database.DB.Where("username = ? AND id != ?", request.Username, id).First(&existingUser).Error; err == nil {
			c.ValidationError(ctx, "用户名已存在")
			return
		}
	}

	// 检查邮箱唯一性（排除自己）
	if request.Email != "" && request.Email != user.Email {
		var existingUser Models.User
		if err := Database.DB.Where("email = ? AND id != ?", request.Email, id).First(&existingUser).Error; err == nil {
			c.ValidationError(ctx, "邮箱已存在")
			return
		}
	}

	// 更新基本信息
	if request.Username != "" {
		user.Username = request.Username
	}
	if request.Email != "" {
		user.Email = request.Email
	}
	if request.Avatar != "" {
		user.Avatar = request.Avatar
	}

	// 处理密码更新
	if request.NewPassword != "" {
		// 验证原密码
		if !Utils.CheckPassword(request.OldPassword, user.Password) {
			c.ValidationError(ctx, "原密码错误")
			return
		}

		// 设置新密码
		if err := user.SetPassword(request.NewPassword); err != nil {
			c.ServerError(ctx, "密码更新失败")
			return
		}
	}

	// 管理员可以更新状态和角色
	if currentUserRole == "admin" {
		if request.Status != nil {
			user.Status = *request.Status
		}
		if request.Role != "" {
			user.Role = request.Role
		}
	}

	// 保存更新
	if err := Database.DB.Save(&user).Error; err != nil {
		c.ServerError(ctx, "更新用户信息失败")
		return
	}

	// 清除敏感信息
	user.Password = ""

	c.Success(ctx, user, "用户信息更新成功")
}

// DeleteUser 删除用户
// 功能说明：
// 1. 删除指定的用户
// 2. 仅管理员可以删除用户
// 3. 检查用户是否有文章（有文章时不允许删除）
// 4. 可选择是否同时删除用户的文章
// 5. 记录删除操作日志
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 权限检查：只有管理员可以删除用户
	userRole := ctx.GetString("user_role")
	if userRole != "admin" {
		c.Forbidden(ctx, "只有管理员可以删除用户")
		return
	}

	// 不能删除自己
	currentUserID := ctx.GetString("user_id")
	if currentUserID == idStr {
		c.ValidationError(ctx, "不能删除自己的账户")
		return
	}

	// 查找用户
	var user Models.User
	if err := Database.DB.First(&user, id).Error; err != nil {
		c.NotFound(ctx, "用户不存在")
		return
	}

	// 不能删除超级管理员
	if user.Role == "admin" {
		c.ValidationError(ctx, "不能删除管理员账户")
		return
	}

	// 检查用户是否有文章
	var postCount int64
	Database.DB.Model(&Models.Post{}).Where("user_id = ?", id).Count(&postCount)
	if postCount > 0 {
		// 检查是否强制删除
		forceDelete := ctx.Query("force") == "true"
		if !forceDelete {
			c.ValidationError(ctx, "该用户有文章，无法删除。如需强制删除，请添加force=true参数")
			return
		}

		// 强制删除时，先删除用户的文章
		if err := Database.DB.Where("user_id = ?", id).Delete(&Models.Post{}).Error; err != nil {
			c.ServerError(ctx, "删除用户文章失败")
			return
		}
	}

	// 检查用户是否有评论（如果有评论表）
	// var commentCount int64
	// Database.DB.Model(&Models.Comment{}).Where("user_id = ?", id).Count(&commentCount)
	// if commentCount > 0 {
	//     // 删除用户评论
	//     Database.DB.Where("user_id = ?", id).Delete(&Models.Comment{})
	// }

	// 删除用户相关的其他数据（如用户设置、偏好等）
	// 这里可以根据实际的数据模型进行扩展

	// 删除用户
	if err := Database.DB.Delete(&user).Error; err != nil {
		c.ServerError(ctx, "删除用户失败")
		return
	}

	// 记录删除操作到审计日志
	auditService := Services.NewAuditService(Database.DB)
	currentUserIDUint, _ := strconv.ParseUint(currentUserID, 10, 32)
	auditService.LogUserAction(nil, uint(currentUserIDUint), ctx.GetString("username"), "delete_user", "user", uint(id), "删除用户")

	c.Success(ctx, gin.H{
		"deleted_user_id": id,
		"deleted_posts":   postCount,
	}, "用户删除成功")
}

// GetUserPosts 获取用户的文章列表
// 功能说明：
// 1. 获取指定用户的文章列表
// 2. 支持分页查询
// 3. 支持按状态筛选
// 4. 权限控制：只能查看自己的文章或公开文章
func (c *UserController) GetUserPosts(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 获取当前用户信息
	currentUserID := ctx.GetString("user_id")
	currentUserRole := ctx.GetString("user_role")

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	status := ctx.Query("status")

	// 构建查询条件
	query := Database.DB.Model(&Models.Post{}).Where("user_id = ?", id)

	// 权限控制：只能查看自己的文章或公开文章
	if currentUserRole != "admin" && currentUserID != idStr {
		query = query.Where("status = 1") // 只查看公开文章
	} else if status != "" {
		query = query.Where("status = ?", status)
	}

	// 分页
	offset := (page - 1) * limit
	var posts []Models.Post
	var total int64

	query.Count(&total)
	query.Preload("Category").Preload("Tags").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts)

	// 返回结果
	c.Success(ctx, gin.H{
		"posts": posts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, "用户文章列表获取成功")
}
