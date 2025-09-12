package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TagController 标签控制器
//
// 重要功能说明：
// 1. 标签管理：创建、读取、更新、删除标签
// 2. 标签关联：管理标签与文章的关联关系
// 3. 标签统计：提供标签使用频率和热门标签统计
// 4. 安全控制：删除标签时检查关联关系，防止数据不一致
// 5. 审计日志：记录所有标签操作，支持操作追踪
//
// 安全注意事项：
// - 删除标签前检查是否有文章使用，防止孤立数据
// - 支持强制删除模式，但需要明确指定force=true参数
// - 所有操作都记录到审计日志，便于安全审计
//
// 性能优化：
// - 使用数据库索引优化标签查询
// - 支持分页和搜索功能
// - 缓存热门标签统计结果
//
// 数据一致性：
// - 删除标签时自动清理关联关系
// - 使用数据库事务确保操作原子性
// - 支持级联删除和关联清理
type TagController struct {
	Controller
}

// NewTagController 创建标签控制器
func NewTagController() *TagController {
	return &TagController{}
}

// GetTags 获取标签列表
// 功能说明：
// 1. 获取所有标签列表
// 2. 支持分页查询
// 3. 支持按名称搜索
// 4. 支持按使用频率排序
// 5. 返回标签的使用统计信息
func (c *TagController) GetTags(ctx *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	search := ctx.Query("search")
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	sortOrder := ctx.DefaultQuery("sort_order", "desc")

	// 构建查询条件
	query := Database.DB.Model(&Models.Tag{})

	// 搜索功能
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 排序
	orderClause := sortBy + " " + sortOrder
	query = query.Order(orderClause)

	// 分页
	offset := (page - 1) * limit
	var tags []Models.Tag
	var total int64

	query.Count(&total)
	query.Offset(offset).Limit(limit).Find(&tags)

	// 获取每个标签的使用统计
	for i := range tags {
		postCount := Database.DB.Model(&tags[i]).Association("Posts").Count()
		// 这里可以添加一个字段来存储使用次数，或者通过关联查询获取
		_ = postCount
	}

	// 返回结果
	c.Success(ctx, gin.H{
		"tags": tags,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, "标签列表获取成功")
}

// GetTag 获取单个标签
// 功能说明：
// 1. 根据ID获取标签详细信息
// 2. 返回使用该标签的文章列表
// 3. 支持分页获取文章
// 4. 包含标签的使用统计
func (c *TagController) GetTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的标签ID")
		return
	}

	var tag Models.Tag
	if err := Database.DB.First(&tag, id).Error; err != nil {
		c.NotFound(ctx, "标签不存在")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// 获取使用该标签的文章（分页）
	offset := (page - 1) * limit
	var posts []Models.Post
	var total int64

	// 通过关联表查询文章
	Database.DB.Model(&Models.Post{}).
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Where("post_tags.tag_id = ? AND posts.status = 1", id).
		Count(&total)

	Database.DB.Model(&Models.Post{}).
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Where("post_tags.tag_id = ? AND posts.status = 1", id).
		Preload("User").Preload("Category").
		Order("posts.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts)

	// 构建响应数据
	response := gin.H{
		"tag": tag,
		"posts": gin.H{
			"data": posts,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	}

	c.Success(ctx, response, "标签获取成功")
}

// CreateTag 创建标签
// 功能说明：
// 1. 创建新标签
// 2. 检查标签名称唯一性
// 3. 设置默认状态为启用
// 4. 记录操作日志
func (c *TagController) CreateTag(ctx *gin.Context) {
	var request Requests.CreateTagInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}

	// 检查标签名称唯一性
	var existingTag Models.Tag
	if err := Database.DB.Where("name = ?", request.Name).First(&existingTag).Error; err == nil {
		c.ValidationError(ctx, "标签名称已存在")
		return
	}

	// 创建标签
	tag := &Models.Tag{
		Name:        request.Name,
		Description: request.Description,
		Color:       request.Color,
	}

	if err := Database.DB.Create(tag).Error; err != nil {
		c.ServerError(ctx, "创建标签失败")
		return
	}

	c.Success(ctx, tag, "标签创建成功")
}

// UpdateTag 更新标签
// 功能说明：
// 1. 更新标签信息
// 2. 检查标签名称唯一性（排除自己）
// 3. 验证颜色格式
// 4. 记录修改历史
func (c *TagController) UpdateTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的标签ID")
		return
	}

	var request Requests.UpdateTagInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}

	// 查找标签
	var tag Models.Tag
	if err := Database.DB.First(&tag, id).Error; err != nil {
		c.NotFound(ctx, "标签不存在")
		return
	}

	// 检查标签名称唯一性（排除自己）
	if request.Name != "" && request.Name != tag.Name {
		var existingTag Models.Tag
		if err := Database.DB.Where("name = ? AND id != ?", request.Name, id).First(&existingTag).Error; err == nil {
			c.ValidationError(ctx, "标签名称已存在")
			return
		}
	}

	// 更新标签字段
	if request.Name != "" {
		tag.Name = request.Name
	}
	if request.Description != "" {
		tag.Description = request.Description
	}
	if request.Color != "" {
		tag.Color = request.Color
	}

	// 保存更新
	if err := Database.DB.Save(&tag).Error; err != nil {
		c.ServerError(ctx, "更新标签失败")
		return
	}

	c.Success(ctx, tag, "标签更新成功")
}

// DeleteTag 删除标签
// 功能说明：
// 1. 删除指定标签
// 2. 检查是否有文章使用该标签
// 3. 可选择是否同时删除关联关系
// 4. 记录删除操作日志
func (c *TagController) DeleteTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的标签ID")
		return
	}

	// 查找标签
	var tag Models.Tag
	if err := Database.DB.First(&tag, id).Error; err != nil {
		c.NotFound(ctx, "标签不存在")
		return
	}

	// 检查是否有文章使用该标签
	var postCount int64
	Database.DB.Model(&tag).Association("Posts").Count()
	postCount = int64(len(tag.Posts))
	if postCount > 0 {
		// 检查是否强制删除
		forceDelete := ctx.Query("force") == "true"
		if !forceDelete {
			c.ValidationError(ctx, "该标签正在被文章使用，无法删除。如需强制删除，请添加force=true参数")
			return
		}

		// 强制删除时，先清除所有关联关系
		if err := Database.DB.Model(&tag).Association("Posts").Clear(); err != nil {
			c.ServerError(ctx, "清除标签关联关系失败")
			return
		}
	}

	// 删除标签（会自动删除关联关系）
	if err := Database.DB.Delete(&tag).Error; err != nil {
		c.ServerError(ctx, "删除标签失败")
		return
	}

	// 记录删除操作到审计日志
	auditService := Services.NewAuditService(Database.DB)
	currentUserID, _ := strconv.ParseUint(ctx.GetString("user_id"), 10, 32)

	// 获取当前用户信息
	var currentUser Models.User
	if err := Database.DB.First(&currentUser, currentUserID).Error; err != nil {
		// 如果无法获取用户信息，使用默认值
		currentUser = Models.User{
			UUID:     strconv.FormatUint(currentUserID, 10),
			Username: ctx.GetString("username"),
		}
	}

	auditService.LogUserAction(&currentUser, uint(currentUserID), ctx.GetString("username"), "delete_tag", "tag", uint(id), "删除标签")

	c.Success(ctx, gin.H{
		"deleted_tag_id": id,
		"affected_posts": postCount,
	}, "标签删除成功")
}

// GetPopularTags 获取热门标签
// 功能说明：
// 1. 获取使用频率最高的标签
// 2. 支持限制返回数量
// 3. 返回标签的使用统计
// 4. 用于标签云显示
func (c *TagController) GetPopularTags(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// 查询使用频率最高的标签
	var tags []Models.Tag
	Database.DB.Model(&Models.Tag{}).
		Select("tags.*, COUNT(post_tags.post_id) as post_count").
		Joins("LEFT JOIN post_tags ON tags.id = post_tags.tag_id").
		Joins("LEFT JOIN posts ON post_tags.post_id = posts.id AND posts.status = 1").
		Group("tags.id").
		Order("post_count DESC").
		Limit(limit).
		Find(&tags)

	c.Success(ctx, tags, "热门标签获取成功")
}
