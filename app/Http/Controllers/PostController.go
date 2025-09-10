package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"github.com/gin-gonic/gin"
	"strconv"
)

// PostController 文章控制器
type PostController struct {
	Controller
}

// NewPostController 创建文章控制器
func NewPostController() *PostController {
	return &PostController{}
}

// GetPosts 获取文章列表
// 功能说明：
// 1. 支持分页查询文章列表
// 2. 支持按分类、标签、状态筛选
// 3. 支持按创建时间、更新时间排序
// 4. 返回文章的基本信息和关联数据
// 5. 支持搜索功能（标题、内容）
func (c *PostController) GetPosts(ctx *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	categoryID := ctx.Query("category_id")
	tagID := ctx.Query("tag_id")
	status := ctx.Query("status")
	search := ctx.Query("search")
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	sortOrder := ctx.DefaultQuery("sort_order", "desc")

	// 构建查询条件
	query := Database.DB.Model(&Models.Post{}).Preload("User").Preload("Category").Preload("Tags")
	
	// 分类筛选
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	
	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 搜索功能
	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// 标签筛选
	if tagID != "" {
		query = query.Joins("JOIN post_tags ON posts.id = post_tags.post_id").
			Where("post_tags.tag_id = ?", tagID)
	}
	
	// 排序
	orderClause := sortBy + " " + sortOrder
	query = query.Order(orderClause)
	
	// 分页
	offset := (page - 1) * limit
	var posts []Models.Post
	var total int64
	
	query.Count(&total)
	query.Offset(offset).Limit(limit).Find(&posts)
	
	// 返回结果
	c.Success(ctx, gin.H{
		"posts": posts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, "文章列表获取成功")
}

// GetPost 获取单个文章
// 功能说明：
// 1. 根据ID获取文章详细信息
// 2. 自动增加浏览次数
// 3. 返回文章的所有关联数据
// 4. 支持草稿文章的权限控制
func (c *PostController) GetPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的文章ID")
		return
	}
	
	var post Models.Post
	if err := Database.DB.Preload("User").Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		c.NotFound(ctx, "文章不存在")
		return
	}
	
	// 检查文章状态和权限
	if post.Status == 0 { // 草稿状态
		userID, exists := ctx.Get("user_id")
		if !exists {
			c.Forbidden(ctx, "无权访问草稿文章")
			return
		}
		// 统一使用uint类型进行比较
		if userIDUint, ok := userID.(uint); !ok || userIDUint != post.UserID {
			c.Forbidden(ctx, "无权访问草稿文章")
			return
		}
	}
	
	// 增加浏览次数
	post.IncrementViewCount()
	Database.DB.Save(&post)
	
	c.Success(ctx, post, "文章获取成功")
}

// CreatePost 创建文章
// 功能说明：
// 1. 创建新文章
// 2. 验证用户权限
// 3. 处理标签关联
// 4. 设置默认状态为草稿
// 5. 记录操作日志
func (c *PostController) CreatePost(ctx *gin.Context) {
	var request Requests.CreatePostInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}
	
	// 获取当前用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.Unauthorized(ctx, "用户未认证")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		c.Unauthorized(ctx, "用户ID格式错误")
		return
	}
	
	// 验证分类是否存在
	var category Models.Category
	if err := Database.DB.First(&category, request.CategoryID).Error; err != nil {
		c.ValidationError(ctx, "指定的分类不存在")
		return
	}
	
	// 创建文章
	post := &Models.Post{
		Title:      request.Title,
		Content:    request.Content,
		Summary:    request.Summary,
		CategoryID: request.CategoryID,
		UserID:     uint(userIDUint),
		Status:     0, // 默认为草稿状态
	}
	
	if err := Database.DB.Create(post).Error; err != nil {
		c.ServerError(ctx, "创建文章失败")
		return
	}
	
	// 处理标签关联
	if len(request.TagIDs) > 0 {
		var tags []Models.Tag
		Database.DB.Where("id IN ?", request.TagIDs).Find(&tags)
		Database.DB.Model(post).Association("Tags").Append(tags)
	}
	
	// 重新加载关联数据
	Database.DB.Preload("User").Preload("Category").Preload("Tags").First(post, post.ID)
	
	c.Success(ctx, post, "文章创建成功")
}

// UpdatePost 更新文章
// 功能说明：
// 1. 更新文章信息
// 2. 验证用户权限（只能更新自己的文章或管理员可以更新所有文章）
// 3. 处理标签关联的更新
// 4. 记录修改历史
func (c *PostController) UpdatePost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的文章ID")
		return
	}
	
	var request Requests.UpdatePostInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}
	
	// 获取当前用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.Unauthorized(ctx, "用户未认证")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		c.Unauthorized(ctx, "用户ID格式错误")
		return
	}
	
	userRole := ctx.GetString("user_role")
	
	// 查找文章
	var post Models.Post
	if err := Database.DB.First(&post, id).Error; err != nil {
		c.NotFound(ctx, "文章不存在")
		return
	}
	
	// 权限检查：只能更新自己的文章或管理员可以更新所有文章
	if userRole != "admin" && userIDUint != post.UserID {
		c.Forbidden(ctx, "只能更新自己的文章")
		return
	}
	
	// 更新文章字段
	if request.Title != "" {
		post.Title = request.Title
	}
	if request.Content != "" {
		post.Content = request.Content
	}
	if request.Summary != "" {
		post.Summary = request.Summary
	}
	if request.CategoryID != 0 {
		// 验证分类是否存在
		var category Models.Category
		if err := Database.DB.First(&category, request.CategoryID).Error; err != nil {
			c.ValidationError(ctx, "指定的分类不存在")
			return
		}
		post.CategoryID = request.CategoryID
	}
	if request.Status != nil {
		post.Status = *request.Status
	}
	
	// 保存更新
	if err := Database.DB.Save(&post).Error; err != nil {
		c.ServerError(ctx, "更新文章失败")
		return
	}
	
	// 处理标签关联
	if request.TagIDs != nil {
		// 清除现有标签关联
		Database.DB.Model(&post).Association("Tags").Clear()
		// 添加新的标签关联
		if len(request.TagIDs) > 0 {
			var tags []Models.Tag
			Database.DB.Where("id IN ?", request.TagIDs).Find(&tags)
			Database.DB.Model(&post).Association("Tags").Append(tags)
		}
	}
	
	// 重新加载关联数据
	Database.DB.Preload("User").Preload("Category").Preload("Tags").First(&post, post.ID)
	
	c.Success(ctx, post, "文章更新成功")
}

// DeletePost 删除文章
// 功能说明：
// 1. 删除指定文章
// 2. 验证用户权限（只能删除自己的文章或管理员可以删除所有文章）
// 3. 同时删除相关的标签关联
// 4. 记录删除操作日志
func (c *PostController) DeletePost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的文章ID")
		return
	}
	
	// 获取当前用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.Unauthorized(ctx, "用户未认证")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		c.Unauthorized(ctx, "用户ID格式错误")
		return
	}
	
	userRole := ctx.GetString("user_role")
	
	// 查找文章
	var post Models.Post
	if err := Database.DB.First(&post, id).Error; err != nil {
		c.NotFound(ctx, "文章不存在")
		return
	}
	
	// 权限检查：只能删除自己的文章或管理员可以删除所有文章
	if userRole != "admin" && userIDUint != post.UserID {
		c.Forbidden(ctx, "只能删除自己的文章")
		return
	}
	
	// 删除文章（会自动删除关联的标签关系）
	if err := Database.DB.Delete(&post).Error; err != nil {
		c.ServerError(ctx, "删除文章失败")
		return
	}
	
	c.Success(ctx, nil, "文章删除成功")
}

