package Migrations

import (
	"cloud-platform-api/app/Models"
	"gorm.io/gorm"
)

// CreatePostsTable 创建文章表迁移
type CreatePostsTable struct{}

// GetName 获取迁移名称
func (m *CreatePostsTable) GetName() string {
	return "2024_01_01_000002_create_posts_table"
}

// Up 执行迁移
func (m *CreatePostsTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Models.Post{})
}

// Down 回滚迁移
func (m *CreatePostsTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Models.Post{})
}
