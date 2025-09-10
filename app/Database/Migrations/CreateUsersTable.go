package Migrations

import (
	"cloud-platform-api/app/Models"
	"gorm.io/gorm"
)

// CreateUsersTable 创建用户表迁移
type CreateUsersTable struct{}

// GetName 获取迁移名称
func (m *CreateUsersTable) GetName() string {
	return "2024_01_01_000001_create_users_table"
}

// Up 执行迁移
func (m *CreateUsersTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Models.User{})
}

// Down 回滚迁移
func (m *CreateUsersTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Models.User{})
}
