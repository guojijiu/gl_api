package Models

import (
	"time"
	"gorm.io/gorm"
)

// Model 基础模型接口
// 功能说明：
// 1. 定义所有模型必须实现的基础方法
// 2. 提供统一的模型操作接口
// 3. 支持软删除和时间戳管理
type Model interface {
	GetID() uint
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() gorm.DeletedAt
}

// BaseModel 基础模型结构
// 功能说明：
// 1. 所有模型的基类，提供通用字段
// 2. 包含主键ID、创建时间、更新时间、删除时间
// 3. 支持GORM的软删除功能
// 4. 自动管理时间戳字段
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`           // 主键ID
	CreatedAt time.Time      `json:"created_at"`                     // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                     // 更新时间
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"` // 删除时间（软删除）
}

// GetID 获取ID
// 功能说明：
// 1. 获取模型的主键ID
// 2. 用于模型标识和关联查询
func (m *BaseModel) GetID() uint {
	return m.ID
}

// GetCreatedAt 获取创建时间
// 功能说明：
// 1. 获取模型的创建时间
// 2. 用于时间排序和统计
func (m *BaseModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// GetUpdatedAt 获取更新时间
// 功能说明：
// 1. 获取模型的最后更新时间
// 2. 用于缓存失效和数据同步
func (m *BaseModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

// GetDeletedAt 获取删除时间
// 功能说明：
// 1. 获取模型的软删除时间
// 2. 用于软删除状态判断
func (m *BaseModel) GetDeletedAt() gorm.DeletedAt {
	return m.DeletedAt
}

// IsDeleted 检查是否已删除
// 功能说明：
// 1. 检查模型是否已被软删除
// 2. 用于数据恢复和状态判断
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}

// GetTableName 获取表名（子类可以重写）
// 功能说明：
// 1. 获取模型的数据库表名
// 2. 子类可以重写此方法自定义表名
// 3. 用于GORM的数据库操作
func (m *BaseModel) GetTableName() string {
	return ""
}

