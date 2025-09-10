package Models

// Category 分类模型
// 功能说明：
// 1. 分类基本信息管理（名称、描述等）
// 2. 支持树形结构（父子分类关系）
// 3. 分类排序和状态管理
// 4. 与文章的关联关系
// 5. 支持分类层级管理
type Category struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;size:50"`      // 分类名称
	Description string `json:"description" gorm:"size:200"`       // 分类描述
	ParentID    *uint  `json:"parent_id"`                         // 父分类ID（支持树形结构）
	Sort        int    `json:"sort" gorm:"default:0"`             // 排序权重
	Status      int    `json:"status" gorm:"default:1"`           // 分类状态：1-启用, 0-禁用

	// 关联关系
	Parent   *Category  `json:"parent" gorm:"foreignKey:ParentID"`   // 父分类
	Children []Category `json:"children" gorm:"foreignKey:ParentID"` // 子分类列表
	Posts    []Post     `json:"posts,omitempty" gorm:"foreignKey:CategoryID"` // 分类下的文章
}

// GetTableName 获取表名
func (c *Category) GetTableName() string {
	return "categories"
}

// IsActive 检查是否激活
func (c *Category) IsActive() bool {
	return c.Status == 1
}

// IsRoot 检查是否为根分类
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren 检查是否有子分类
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetPostsCount 获取文章数量
func (c *Category) GetPostsCount() int {
	return len(c.Posts)
}

// GetFullPath 获取完整路径
// 功能说明：
// 1. 获取分类的完整层级路径
// 2. 从根分类到当前分类的完整路径
// 3. 用 " > " 分隔各级分类
// 4. 避免递归调用导致的无限循环
func (c *Category) GetFullPath() string {
	if c.IsRoot() {
		return c.Name
	}

	if c.Parent != nil {
		return c.Parent.GetFullPath() + " > " + c.Name
	}

	return c.Name
}
