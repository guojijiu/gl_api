package Migrations

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Migration 迁移记录模型
type Migration struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Migration string    `json:"migration" gorm:"uniqueIndex;not null;size:255"`
	Batch     int       `json:"batch"`
	CreatedAt time.Time `json:"created_at"`
}

// MigrationInterface 迁移接口
type MigrationInterface interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
	GetName() string
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db *gorm.DB
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{
		db: db,
	}
}

// CreateMigrationsTable 创建迁移表
func (m *MigrationManager) CreateMigrationsTable() error {
	// 检查迁移表是否存在
	if !m.db.Migrator().HasTable(&Migration{}) {
		// 如果表不存在，创建迁移表
		log.Println("创建迁移表...")
		return m.db.AutoMigrate(&Migration{})
	}

	// 表已存在，无需重新创建
	log.Println("迁移表已存在，跳过创建")
	return nil
}

// GetMigrations 获取所有迁移
func (m *MigrationManager) GetMigrations() ([]Migration, error) {
	var migrations []Migration
	err := m.db.Order("batch asc, id asc").Find(&migrations).Error
	return migrations, err
}

// GetLastBatch 获取最后一批次号
func (m *MigrationManager) GetLastBatch() (int, error) {
	var migration Migration
	err := m.db.Order("batch desc").First(&migration).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return migration.Batch, nil
}

// RunMigrations 运行迁移
func (m *MigrationManager) RunMigrations() error {
	// 确保迁移表存在
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("创建迁移表失败: %v", err)
	}

	// 获取所有迁移文件
	migrations := m.GetMigrationFiles()

	// 获取已运行的迁移
	ranMigrations, err := m.GetMigrations()
	if err != nil {
		return fmt.Errorf("获取已运行迁移失败: %v", err)
	}

	// 获取最后批次号
	lastBatch, err := m.GetLastBatch()
	if err != nil {
		return fmt.Errorf("获取最后批次失败: %v", err)
	}

	currentBatch := lastBatch + 1
	ranMap := make(map[string]bool)
	for _, ran := range ranMigrations {
		ranMap[ran.Migration] = true
	}

	// 运行未执行的迁移
	for _, migration := range migrations {
		if !ranMap[migration.GetName()] {
			log.Printf("运行迁移: %s", migration.GetName())

			if err := migration.Up(m.db); err != nil {
				return fmt.Errorf("迁移 %s 执行失败: %v", migration.GetName(), err)
			}

			// 记录迁移
			migrationRecord := Migration{
				Migration: migration.GetName(),
				Batch:     currentBatch,
				CreatedAt: time.Now(),
			}

			if err := m.db.Create(&migrationRecord).Error; err != nil {
				return fmt.Errorf("记录迁移失败: %v", err)
			}

			log.Printf("迁移 %s 执行成功", migration.GetName())
		}
	}

	log.Printf("所有迁移执行完成，当前批次: %d", currentBatch)
	return nil
}

// RollbackMigrations 回滚迁移
func (m *MigrationManager) RollbackMigrations(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("回滚步数必须大于0")
	}

	// 获取最后一批次的迁移
	lastBatch, err := m.GetLastBatch()
	if err != nil {
		return fmt.Errorf("获取最后批次失败: %v", err)
	}

	if lastBatch == 0 {
		return fmt.Errorf("没有可回滚的迁移")
	}

	// 获取要回滚的批次
	rollbackBatch := lastBatch - steps + 1
	if rollbackBatch < 1 {
		rollbackBatch = 1
	}

	var migrations []Migration
	err = m.db.Where("batch >= ?", rollbackBatch).Order("batch desc, id desc").Find(&migrations).Error
	if err != nil {
		return fmt.Errorf("获取迁移记录失败: %v", err)
	}

	// 获取所有迁移文件
	migrationFiles := m.GetMigrationFiles()
	migrationMap := make(map[string]MigrationInterface)
	for _, migration := range migrationFiles {
		migrationMap[migration.GetName()] = migration
	}

	// 执行回滚
	for _, migrationRecord := range migrations {
		migration, exists := migrationMap[migrationRecord.Migration]
		if !exists {
			log.Printf("警告: 找不到迁移文件 %s，跳过回滚", migrationRecord.Migration)
			continue
		}

		log.Printf("回滚迁移: %s", migrationRecord.Migration)

		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("迁移 %s 回滚失败: %v", migrationRecord.Migration, err)
		}

		// 删除迁移记录
		if err := m.db.Delete(&migrationRecord).Error; err != nil {
			return fmt.Errorf("删除迁移记录失败: %v", err)
		}

		log.Printf("迁移 %s 回滚成功", migrationRecord.Migration)
	}

	log.Printf("回滚完成，回滚了 %d 个批次", lastBatch-rollbackBatch+1)
	return nil
}

// ResetMigrations 重置所有迁移
func (m *MigrationManager) ResetMigrations() error {
	return m.RollbackMigrations(999) // 回滚所有批次
}

// GetMigrationFiles 获取所有迁移文件
func (m *MigrationManager) GetMigrationFiles() []MigrationInterface {
	return []MigrationInterface{
		&CreateUsersTable{},
		&CreateCategoriesTable{},
		&CreateTagsTable{},
		&CreatePostsTable{},
		&CreateAuditLogsTable{},
	}
}

// GetMigrationStatus 获取迁移状态
func (m *MigrationManager) GetMigrationStatus() (map[string]interface{}, error) {
	ranMigrations, err := m.GetMigrations()
	if err != nil {
		return nil, err
	}

	allMigrations := m.GetMigrationFiles()
	allMap := make(map[string]bool)
	for _, migration := range allMigrations {
		allMap[migration.GetName()] = false
	}

	ranMap := make(map[string]Migration)
	for _, ran := range ranMigrations {
		ranMap[ran.Migration] = ran
		allMap[ran.Migration] = true
	}

	status := map[string]interface{}{
		"total_migrations":   len(allMigrations),
		"ran_migrations":     len(ranMigrations),
		"pending_migrations": len(allMigrations) - len(ranMigrations),
		"last_batch":         0,
		"migrations":         make([]map[string]interface{}, 0),
	}

	if len(ranMigrations) > 0 {
		status["last_batch"] = ranMigrations[len(ranMigrations)-1].Batch
	}

	// 构建迁移详情
	for _, migration := range allMigrations {
		name := migration.GetName()
		migrationInfo := map[string]interface{}{
			"name":   name,
			"status": "pending",
			"batch":  nil,
			"ran_at": nil,
		}

		if ran, exists := ranMap[name]; exists {
			migrationInfo["status"] = "ran"
			migrationInfo["batch"] = ran.Batch
			migrationInfo["ran_at"] = ran.CreatedAt
		}

		status["migrations"] = append(status["migrations"].([]map[string]interface{}), migrationInfo)
	}

	return status, nil
}
