package Migrations

import (
	"cloud-platform-api/app/Models"
	"gorm.io/gorm"
)

// CreateTagsTable 创建标签表迁移
type CreateTagsTable struct{}

// GetName 获取迁移名称
func (m *CreateTagsTable) GetName() string {
	return "2024_01_01_000004_create_tags_table"
}

// Up 执行迁移
func (m *CreateTagsTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Models.Tag{})
}

// Down 回滚迁移
func (m *CreateTagsTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Models.Tag{})
}
