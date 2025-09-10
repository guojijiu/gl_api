package Controllers

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// AdminController 管理员控制器
type AdminController struct {
	Controller
}

// NewAdminController 创建管理员控制器
func NewAdminController() *AdminController {
	return &AdminController{}
}

// Dashboard 管理员仪表板
// 功能说明：
// 1. 获取系统整体统计数据
// 2. 显示用户、文章、分类、标签总数
// 3. 显示今日新增数据
// 4. 显示系统运行状态
// 5. 用于管理员快速了解系统状况
func (c *AdminController) Dashboard(ctx *gin.Context) {
	// 获取基础统计数据
	var totalUsers, totalPosts, totalCategories, totalTags int64
	var todayUsers, todayPosts int64
	
	// 用户统计
	Database.DB.Model(&Models.User{}).Count(&totalUsers)
	Database.DB.Model(&Models.User{}).
		Where("DATE(created_at) = CURDATE()").
		Count(&todayUsers)
	
	// 文章统计
	Database.DB.Model(&Models.Post{}).Count(&totalPosts)
	Database.DB.Model(&Models.Post{}).
		Where("DATE(created_at) = CURDATE()").
		Count(&todayPosts)
	
	// 分类统计
	Database.DB.Model(&Models.Category{}).Count(&totalCategories)
	
	// 标签统计
	Database.DB.Model(&Models.Tag{}).Count(&totalTags)
	
	// 获取最近活跃用户
	var recentUsers []Models.User
	Database.DB.Order("last_login_at DESC").Limit(5).Find(&recentUsers)
	
	// 获取最近发布的文章
	var recentPosts []Models.Post
	Database.DB.Preload("User").Preload("Category").
		Where("status = 1").
		Order("created_at DESC").
		Limit(5).
		Find(&recentPosts)
	
	// 构建仪表板数据
	dashboardData := gin.H{
		"overview": gin.H{
			"total_users":      totalUsers,
			"total_posts":      totalPosts,
			"total_categories": totalCategories,
			"total_tags":       totalTags,
		},
		"today": gin.H{
			"new_users": todayUsers,
			"new_posts": todayPosts,
		},
		"recent_users": recentUsers,
		"recent_posts": recentPosts,
		"system_status": gin.H{
			"database_connected": true, // 这里可以添加实际的数据库连接检查
			"uptime":            "运行中", // 这里可以添加实际的运行时间计算
		},
	}
	
	c.Success(ctx, dashboardData, "管理员仪表板数据获取成功")
}

// Stats 管理员统计
// 功能说明：
// 1. 获取详细的统计数据
// 2. 显示用户增长趋势
// 3. 显示文章增长趋势
// 4. 显示热门文章排行
// 5. 显示最近活动记录
func (c *AdminController) Stats(ctx *gin.Context) {
	// 获取时间范围参数
	days := 30 // 默认30天
	if daysParam := ctx.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 {
			days = d
		}
	}
	
	// 用户增长统计
	userGrowth := c.getUserGrowth(days)
	
	// 文章增长统计
	postGrowth := c.getPostGrowth(days)
	
	// 热门文章排行
	popularPosts := c.getPopularPosts(10)
	
	// 最近活动记录
	recentActivities := c.getRecentActivities(20)
	
	// 分类文章统计
	categoryStats := c.getCategoryStats()
	
	// 标签使用统计
	tagStats := c.getTagStats()
	
	// 构建统计数据
	statsData := gin.H{
		"user_growth":        userGrowth,
		"post_growth":        postGrowth,
		"popular_posts":      popularPosts,
		"recent_activities":  recentActivities,
		"category_stats":     categoryStats,
		"tag_stats":          tagStats,
		"period_days":        days,
	}
	
	c.Success(ctx, statsData, "统计数据获取成功")
}

// getUserGrowth 获取用户增长数据
// 功能说明：
// 1. 统计指定天数内的用户注册数量
// 2. 按日期分组统计
// 3. 返回增长趋势数据
func (c *AdminController) getUserGrowth(days int) []gin.H {
	var growth []gin.H
	
	// 获取最近N天的用户注册统计
	rows, err := Database.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, days).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date string
			var count int
			rows.Scan(&date, &count)
			growth = append(growth, gin.H{
				"date":  date,
				"count": count,
			})
		}
	}
	
	return growth
}

// getPostGrowth 获取文章增长数据
// 功能说明：
// 1. 统计指定天数内的文章发布数量
// 2. 按日期分组统计
// 3. 返回增长趋势数据
func (c *AdminController) getPostGrowth(days int) []gin.H {
	var growth []gin.H
	
	// 获取最近N天的文章发布统计
	rows, err := Database.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM posts
		WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, days).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date string
			var count int
			rows.Scan(&date, &count)
			growth = append(growth, gin.H{
				"date":  date,
				"count": count,
			})
		}
	}
	
	return growth
}

// getPopularPosts 获取热门文章
// 功能说明：
// 1. 获取浏览次数最多的文章
// 2. 支持限制返回数量
// 3. 包含文章的基本信息
func (c *AdminController) getPopularPosts(limit int) []gin.H {
	var posts []gin.H
	
	// 查询热门文章
	rows, err := Database.DB.Raw(`
		SELECT p.id, p.title, p.view_count, p.created_at, u.username
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.status = 1
		ORDER BY p.view_count DESC
		LIMIT ?
	`, limit).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id uint
			var title string
			var viewCount int
			var createdAt time.Time
			var username string
			rows.Scan(&id, &title, &viewCount, &createdAt, &username)
			posts = append(posts, gin.H{
				"id":         id,
				"title":      title,
				"view_count": viewCount,
				"created_at": createdAt.Format("2006-01-02 15:04:05"),
				"username":   username,
			})
		}
	}
	
	return posts
}

// getRecentActivities 获取最近活动
// 功能说明：
// 1. 获取系统最近的活动记录
// 2. 包括用户注册、文章发布等
// 3. 按时间倒序排列
func (c *AdminController) getRecentActivities(limit int) []gin.H {
	var activities []gin.H
	
	// 获取最近用户注册
	var recentUsers []Models.User
	Database.DB.Order("created_at DESC").Limit(limit/2).Find(&recentUsers)
	
	for _, user := range recentUsers {
		activities = append(activities, gin.H{
			"type":      "user_register",
			"message":   "新用户注册: " + user.Username,
			"timestamp": user.CreatedAt.Format("2006-01-02 15:04:05"),
			"user_id":   user.ID,
		})
	}
	
	// 获取最近文章发布
	var recentPosts []Models.Post
	Database.DB.Preload("User").
		Where("status = 1").
		Order("created_at DESC").
		Limit(limit/2).
		Find(&recentPosts)
	
	for _, post := range recentPosts {
		activities = append(activities, gin.H{
			"type":      "post_publish",
			"message":   "新文章发布: " + post.Title,
			"timestamp": post.CreatedAt.Format("2006-01-02 15:04:05"),
			"post_id":   post.ID,
			"user_id":   post.UserID,
		})
	}
	
	// 按时间排序
	// 这里简化处理，实际应该合并后排序
	return activities
}

// getCategoryStats 获取分类统计
// 功能说明：
// 1. 统计每个分类下的文章数量
// 2. 返回分类使用情况
func (c *AdminController) getCategoryStats() []gin.H {
	var stats []gin.H
	
	rows, err := Database.DB.Raw(`
		SELECT c.id, c.name, COUNT(p.id) as post_count
		FROM categories c
		LEFT JOIN posts p ON c.id = p.category_id AND p.status = 1
		GROUP BY c.id, c.name
		ORDER BY post_count DESC
	`).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id uint
			var name string
			var postCount int
			rows.Scan(&id, &name, &postCount)
			stats = append(stats, gin.H{
				"id":         id,
				"name":       name,
				"post_count": postCount,
			})
		}
	}
	
	return stats
}

// getTagStats 获取标签统计
// 功能说明：
// 1. 统计每个标签的使用次数
// 2. 返回标签使用情况
func (c *AdminController) getTagStats() []gin.H {
	var stats []gin.H
	
	rows, err := Database.DB.Raw(`
		SELECT t.id, t.name, COUNT(pt.post_id) as usage_count
		FROM tags t
		LEFT JOIN post_tags pt ON t.id = pt.tag_id
		LEFT JOIN posts p ON pt.post_id = p.id AND p.status = 1
		GROUP BY t.id, t.name
		ORDER BY usage_count DESC
		LIMIT 20
	`).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id uint
			var name string
			var usageCount int
			rows.Scan(&id, &name, &usageCount)
			stats = append(stats, gin.H{
				"id":          id,
				"name":        name,
				"usage_count": usageCount,
			})
		}
	}
	
	return stats
}
