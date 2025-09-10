package Models

import (
	"strings"
)

// Post 文章模型
// 功能说明：
// 1. 文章基本信息管理（标题、内容、摘要等）
// 2. 文章状态管理（发布、草稿）
// 3. 文章统计信息（浏览次数）
// 4. 与用户、分类、标签的关联关系
// 5. 支持文章分类和标签管理
type Post struct {
	BaseModel
	Title      string `json:"title" gorm:"not null;size:200"`     // 文章标题
	Content    string `json:"content" gorm:"type:text"`           // 文章内容
	Summary    string `json:"summary" gorm:"size:500"`            // 文章摘要
	Status     int    `json:"status" gorm:"default:1"`            // 文章状态：1-发布, 0-草稿
	ViewCount  int    `json:"view_count" gorm:"default:0"`        // 浏览次数
	UserID     uint   `json:"user_id"`                            // 作者ID
	CategoryID uint   `json:"category_id"`                        // 分类ID

	// 关联关系
	User     *User     `json:"user" gorm:"foreignKey:UserID"`     // 文章作者
	Category *Category `json:"category" gorm:"foreignKey:CategoryID"` // 文章分类
	Tags     []Tag     `json:"tags" gorm:"many2many:post_tags;"`  // 文章标签
}

// GetTableName 获取表名
// 功能说明：
// 1. 返回文章表的数据库表名
// 2. 用于GORM的数据库操作
func (p *Post) GetTableName() string {
	return "posts"
}

// IsPublished 检查是否已发布
// 功能说明：
// 1. 检查文章是否已发布状态
// 2. 用于文章状态判断和权限控制
// 3. 返回布尔值表示发布状态
func (p *Post) IsPublished() bool {
	return p.Status == 1
}

// IsDraft 检查是否为草稿
// 功能说明：
// 1. 检查文章是否为草稿状态
// 2. 用于文章状态判断和权限控制
// 3. 返回布尔值表示草稿状态
func (p *Post) IsDraft() bool {
	return p.Status == 0
}

// IncrementViewCount 增加浏览次数
// 功能说明：
// 1. 增加文章的浏览次数计数器
// 2. 在文章被访问时调用
// 3. 用于文章热度统计
func (p *Post) IncrementViewCount() {
	p.ViewCount++
}

// GetExcerpt 获取摘要
func (p *Post) GetExcerpt(length int) string {
	if length <= 0 {
		length = 100
	}

	if len(p.Summary) > 0 {
		if len(p.Summary) <= length {
			return p.Summary
		}
		return p.Summary[:length] + "..."
	}

	// 如果没有摘要，从内容中提取
	if len(p.Content) <= length {
		return p.Content
	}
	return p.Content[:length] + "..."
}

// GetReadingTime 获取阅读时间（估算）
func (p *Post) GetReadingTime() int {
	// 假设平均阅读速度为每分钟200字
	wordCount := len(p.Content)
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}
	return readingTime
}

// GetFormattedDate 获取格式化的日期
func (p *Post) GetFormattedDate() string {
	return p.CreatedAt.Format("2006年01月02日")
}

// GetTagsString 获取标签字符串
// 功能说明：
// 1. 将文章的所有标签名称拼接成字符串
// 2. 用逗号分隔多个标签
// 3. 如果没有标签，返回空字符串
// 4. 用于显示和搜索功能
func (p *Post) GetTagsString() string {
	if len(p.Tags) == 0 {
		return ""
	}

	var tagNames []string
	for _, tag := range p.Tags {
		tagNames = append(tagNames, tag.Name)
	}

	// 使用strings.Join进行字符串拼接
	return strings.Join(tagNames, ", ")
}
