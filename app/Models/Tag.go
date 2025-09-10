package Models

// Tag 标签模型
// 功能说明：
// 1. 标签基本信息管理（名称、描述、颜色等）
// 2. 与文章的多对多关联关系
// 3. 支持标签颜色自定义
// 4. 标签使用统计和热门标签识别
type Tag struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;size:50"`      // 标签名称
	Description string `json:"description" gorm:"size:200"`       // 标签描述
	Color       string `json:"color" gorm:"size:20"`              // 标签颜色（用于前端显示）

	// 关联关系
	Posts []Post `json:"posts,omitempty" gorm:"many2many:post_tags;"` // 使用该标签的文章
}

// GetTableName 获取表名
// 功能说明：
// 1. 返回标签表的数据库表名
// 2. 用于GORM的数据库操作
func (t *Tag) GetTableName() string {
	return "tags"
}

// GetPostsCount 获取文章数量
// 功能说明：
// 1. 获取使用该标签的文章数量
// 2. 用于标签热度统计
// 3. 返回整数表示文章数量
func (t *Tag) GetPostsCount() int {
	return len(t.Posts)
}

// GetColorOrDefault 获取颜色或默认颜色
// 功能说明：
// 1. 获取标签的显示颜色
// 2. 如果标签没有设置颜色，返回默认灰色
// 3. 用于前端界面显示
func (t *Tag) GetColorOrDefault() string {
	if t.Color != "" {
		return t.Color
	}
	return "#6c757d" // 默认灰色
}

// IsPopular 检查是否为热门标签（文章数量大于5）
// 功能说明：
// 1. 检查标签是否为热门标签
// 2. 基于使用该标签的文章数量判断
// 3. 用于热门标签推荐和显示
func (t *Tag) IsPopular() bool {
	return t.GetPostsCount() > 5
}
