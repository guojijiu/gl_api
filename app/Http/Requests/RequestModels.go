package Requests

// RegisterInput 注册输入
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginInput 登录输入
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileInput 更新资料输入
type UpdateProfileInput struct {
	Username string `json:"username" binding:"min=3,max=50"`
	Email    string `json:"email" binding:"email"`
	Avatar   string `json:"avatar"`
}

// CreatePostInput 创建文章输入
type CreatePostInput struct {
	Title      string `json:"title" binding:"required,min=1,max=200"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id" binding:"required"`
	TagIDs     []uint `json:"tag_ids"`
}

// UpdatePostInput 更新文章输入
type UpdatePostInput struct {
	Title      string `json:"title" binding:"min=1,max=200"`
	Content    string `json:"content"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
	Status     *int   `json:"status"`
}

// CreateCategoryInput 创建分类输入
type CreateCategoryInput struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	ParentID    *uint  `json:"parent_id"`
	Sort        int    `json:"sort"`
}

// UpdateCategoryInput 更新分类输入
type UpdateCategoryInput struct {
	Name        string `json:"name" binding:"min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	ParentID    *uint  `json:"parent_id"`
	Sort        int    `json:"sort"`
	Status      *int   `json:"status"`
}

// CreateTagInput 创建标签输入
type CreateTagInput struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	Color       string `json:"color" binding:"max=20"`
}

// UpdateTagInput 更新标签输入
type UpdateTagInput struct {
	Name        string `json:"name" binding:"min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	Color       string `json:"color" binding:"max=20"`
}


