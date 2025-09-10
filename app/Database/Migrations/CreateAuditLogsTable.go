package Migrations

import (
	"cloud-platform-api/app/Models"
	"gorm.io/gorm"
)

// CreateAuditLogsTable 创建审计日志表迁移
type CreateAuditLogsTable struct{}

// GetName 获取迁移名称
func (m *CreateAuditLogsTable) GetName() string {
	return "2024_01_01_000005_create_audit_logs_table"
}

// Up 执行迁移
func (m *CreateAuditLogsTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Models.AuditLog{})
}

// Down 回滚迁移
func (m *CreateAuditLogsTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&Models.AuditLog{})
}
