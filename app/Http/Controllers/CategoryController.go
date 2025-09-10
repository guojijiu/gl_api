package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"github.com/gin-gonic/gin"
	"strconv"
)

// CategoryController 分类控制器
type CategoryController struct {
	Controller
}

// NewCategoryController 创建分类控制器
func NewCategoryController() *CategoryController {
	return &CategoryController{}
}

// GetCategories 获取分类列表
// 功能说明：
// 1. 获取所有分类的树形结构
// 2. 支持按状态筛选（启用/禁用）
// 3. 支持按排序字段排序
// 4. 返回分类的层级关系
// 5. 包含每个分类的文章数量统计
func (c *CategoryController) GetCategories(ctx *gin.Context) {
	status := ctx.Query("status")
	
	// 构建查询条件
	query := Database.DB.Model(&Models.Category{}).Preload("Children").Preload("Posts")
	
	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 只获取根分类
	query = query.Where("parent_id IS NULL").Order("sort ASC, created_at ASC")
	
	var categories []Models.Category
	if err := query.Find(&categories).Error; err != nil {
		c.ServerError(ctx, "获取分类列表失败")
		return
	}
	
	// 递归加载子分类
	for i := range categories {
		c.loadCategoryChildren(&categories[i])
	}
	
	c.Success(ctx, categories, "分类列表获取成功")
}

// GetCategory 获取单个分类
// 功能说明：
// 1. 根据ID获取分类详细信息
// 2. 返回分类的完整层级路径
// 3. 包含子分类和文章列表
// 4. 支持分页获取文章
func (c *CategoryController) GetCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的分类ID")
		return
	}
	
	var category Models.Category
	if err := Database.DB.Preload("Parent").Preload("Children").Preload("Posts").First(&category, id).Error; err != nil {
		c.NotFound(ctx, "分类不存在")
		return
	}
	
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	
	// 获取分类下的文章（分页）
	offset := (page - 1) * limit
	var posts []Models.Post
	var total int64
	
	Database.DB.Model(&Models.Post{}).Where("category_id = ? AND status = 1", id).Count(&total)
	Database.DB.Where("category_id = ? AND status = 1", id).
		Preload("User").Preload("Tags").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts)
	
	// 构建响应数据
	response := gin.H{
		"category": category,
		"full_path": category.GetFullPath(),
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
	
	c.Success(ctx, response, "分类获取成功")
}

// CreateCategory 创建分类
// 功能说明：
// 1. 创建新分类
// 2. 验证父分类是否存在
// 3. 检查分类名称唯一性
// 4. 设置默认状态为启用
// 5. 记录操作日志
func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var request Requests.CreateCategoryInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}
	
	// 检查分类名称唯一性
	var existingCategory Models.Category
	if err := Database.DB.Where("name = ?", request.Name).First(&existingCategory).Error; err == nil {
		c.ValidationError(ctx, "分类名称已存在")
		return
	}
	
	// 验证父分类是否存在
	if request.ParentID != nil {
		var parentCategory Models.Category
		if err := Database.DB.First(&parentCategory, *request.ParentID).Error; err != nil {
			c.ValidationError(ctx, "指定的父分类不存在")
			return
		}
	}
	
	// 创建分类
	category := &Models.Category{
		Name:        request.Name,
		Description: request.Description,
		ParentID:    request.ParentID,
		Sort:        request.Sort,
		Status:      1, // 默认为启用状态
	}
	
	if err := Database.DB.Create(category).Error; err != nil {
		c.ServerError(ctx, "创建分类失败")
		return
	}
	
	// 重新加载关联数据
	Database.DB.Preload("Parent").Preload("Children").First(category, category.ID)
	
	c.Success(ctx, category, "分类创建成功")
}

// UpdateCategory 更新分类
// 功能说明：
// 1. 更新分类信息
// 2. 验证父分类不能设置为自己或自己的子分类
// 3. 检查分类名称唯一性（排除自己）
// 4. 处理分类层级变更
func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的分类ID")
		return
	}
	
	var request Requests.UpdateCategoryInput
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}
	
	// 查找分类
	var category Models.Category
	if err := Database.DB.First(&category, id).Error; err != nil {
		c.NotFound(ctx, "分类不存在")
		return
	}
	
	// 检查分类名称唯一性（排除自己）
	if request.Name != "" && request.Name != category.Name {
		var existingCategory Models.Category
		if err := Database.DB.Where("name = ? AND id != ?", request.Name, id).First(&existingCategory).Error; err == nil {
			c.ValidationError(ctx, "分类名称已存在")
			return
		}
	}
	
	// 验证父分类
	if request.ParentID != nil {
		// 不能设置自己为父分类
		if *request.ParentID == uint(id) {
			c.ValidationError(ctx, "不能设置自己为父分类")
			return
		}
		
		// 检查是否设置为自己的子分类
		if c.isDescendant(uint(id), *request.ParentID) {
			c.ValidationError(ctx, "不能设置自己的子分类为父分类")
			return
		}
		
		// 验证父分类是否存在
		var parentCategory Models.Category
		if err := Database.DB.First(&parentCategory, *request.ParentID).Error; err != nil {
			c.ValidationError(ctx, "指定的父分类不存在")
			return
		}
	}
	
	// 更新分类字段
	if request.Name != "" {
		category.Name = request.Name
	}
	if request.Description != "" {
		category.Description = request.Description
	}
	if request.ParentID != nil {
		category.ParentID = request.ParentID
	}
	if request.Sort != 0 {
		category.Sort = request.Sort
	}
	if request.Status != nil {
		category.Status = *request.Status
	}
	
	// 保存更新
	if err := Database.DB.Save(&category).Error; err != nil {
		c.ServerError(ctx, "更新分类失败")
		return
	}
	
	// 重新加载关联数据
	Database.DB.Preload("Parent").Preload("Children").First(&category, category.ID)
	
	c.Success(ctx, category, "分类更新成功")
}

// DeleteCategory 删除分类
// 功能说明：
// 1. 删除指定分类
// 2. 检查是否有子分类（有子分类时不允许删除）
// 3. 检查是否有文章（有文章时不允许删除）
// 4. 处理分类下的文章（可选择移动到其他分类）
func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ValidationError(ctx, "无效的分类ID")
		return
	}
	
	// 查找分类
	var category Models.Category
	if err := Database.DB.Preload("Children").First(&category, id).Error; err != nil {
		c.NotFound(ctx, "分类不存在")
		return
	}
	
	// 检查是否有子分类
	if len(category.Children) > 0 {
		// 检查是否强制删除
		forceDelete := ctx.Query("force") == "true"
		if !forceDelete {
			c.ValidationError(ctx, "该分类下有子分类，无法删除。如需强制删除，请添加force=true参数")
			return
		}
		
		// 强制删除时，先删除所有子分类
		if err := c.deleteCategoryRecursively(uint(id)); err != nil {
			c.ServerError(ctx, "删除子分类失败")
			return
		}
	}
	
	// 检查是否有文章
	var postCount int64
	Database.DB.Model(&Models.Post{}).Where("category_id = ?", id).Count(&postCount)
	if postCount > 0 {
		// 获取目标分类ID（用于移动文章）
		targetCategoryID := ctx.Query("move_to")
		if targetCategoryID != "" {
			// 验证目标分类是否存在
			var targetCategory Models.Category
			if err := Database.DB.First(&targetCategory, targetCategoryID).Error; err != nil {
				c.ValidationError(ctx, "指定的目标分类不存在")
				return
			}
			
			// 移动文章到目标分类
			if err := Database.DB.Model(&Models.Post{}).Where("category_id = ?", id).Update("category_id", targetCategoryID).Error; err != nil {
				c.ServerError(ctx, "移动文章失败")
				return
			}
		} else {
			// 检查是否强制删除
			forceDelete := ctx.Query("force") == "true"
			if !forceDelete {
				c.ValidationError(ctx, "该分类下有文章，无法删除。请指定move_to参数移动文章，或添加force=true参数强制删除")
				return
			}
			
			// 强制删除时，先删除所有文章
			if err := Database.DB.Where("category_id = ?", id).Delete(&Models.Post{}).Error; err != nil {
				c.ServerError(ctx, "删除文章失败")
				return
			}
		}
	}
	
	// 删除分类
	if err := Database.DB.Delete(&category).Error; err != nil {
		c.ServerError(ctx, "删除分类失败")
		return
	}
	
	// 记录删除操作到审计日志
	auditService := Services.NewAuditService()
	currentUserID, _ := strconv.ParseUint(ctx.GetString("user_id"), 10, 32)
	auditService.LogUserAction(nil, uint(currentUserID), ctx.GetString("username"), "delete_category", "category", uint(id), "删除分类")
	
	c.Success(ctx, gin.H{
		"deleted_category_id": id,
		"deleted_posts":       postCount,
		"deleted_children":    len(category.Children),
	}, "分类删除成功")
}

// deleteCategoryRecursively 递归删除分类
// 功能说明：
// 1. 递归删除分类及其所有子分类
// 2. 处理每个子分类下的文章
// 3. 确保数据完整性
func (c *CategoryController) deleteCategoryRecursively(categoryID uint) error {
	// 查找子分类
	var children []Models.Category
	if err := Database.DB.Where("parent_id = ?", categoryID).Find(&children).Error; err != nil {
		return err
	}
	
	// 递归删除子分类
	for _, child := range children {
		if err := c.deleteCategoryRecursively(child.ID); err != nil {
			return err
		}
	}
	
	// 删除当前分类下的文章
	if err := Database.DB.Where("category_id = ?", categoryID).Delete(&Models.Post{}).Error; err != nil {
		return err
	}
	
	// 删除当前分类
	return Database.DB.Delete(&Models.Category{}, categoryID).Error
}

// loadCategoryChildren 递归加载子分类
// 功能说明：
// 1. 递归加载分类的所有子分类
// 2. 构建完整的分类树形结构
// 3. 按排序字段排序
func (c *CategoryController) loadCategoryChildren(category *Models.Category) {
	Database.DB.Where("parent_id = ?", category.ID).Order("sort ASC, created_at ASC").Find(&category.Children)
	
	for i := range category.Children {
		c.loadCategoryChildren(&category.Children[i])
	}
}

// isDescendant 检查是否为后代分类
// 功能说明：
// 1. 检查指定分类是否为另一个分类的后代
// 2. 用于防止循环引用
// 3. 递归检查父分类关系
func (c *CategoryController) isDescendant(categoryID, parentID uint) bool {
	var category Models.Category
	if err := Database.DB.First(&category, parentID).Error; err != nil {
		return false
	}
	
	if category.ParentID == nil {
		return false
	}
	
	if *category.ParentID == categoryID {
		return true
	}
	
	return c.isDescendant(categoryID, *category.ParentID)
}

