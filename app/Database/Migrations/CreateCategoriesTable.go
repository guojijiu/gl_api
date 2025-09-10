package Migrations

import (
	"cloud-platform-api/app/Models"
	"gorm.io/gorm"
)

// CreateCategoriesTable 创建分类表迁移
type CreateCategoriesTable struct{}

// GetName 获取迁移名称
func (m *CreateCategoriesTable) GetName() string {
	return "2024_01_01_000003_create_categories_table"
}

// Up 执行迁移
func (m *CreateCategoriesTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Models.Category{})
}

// Down 回滚迁移
func (m *CreateCategoriesTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Models.Category{})
}
