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
	ID        uint           `json:"id" gorm:"primarykey"`              // 主键ID
	CreatedAt time.Time      `json:"created_at"`                        // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                        // 更新时间
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
//
// 功能说明：
// 1. 检查模型是否已被软删除
// 2. 用于数据恢复和状态判断
// 3. 软删除的数据不会从数据库中物理删除，只是标记为已删除
//
// 软删除机制：
// - DeletedAt字段不为空表示已删除
// - 已删除的数据在查询时会被自动过滤
// - 可以使用Unscoped()查询已删除的数据
// - 支持数据恢复（将DeletedAt设置为NULL）
//
// 使用场景：
// - 检查数据是否可用
// - 数据恢复前检查状态
// - 业务逻辑中的状态判断
//
// 注意事项：
// - 软删除的数据仍然占用数据库空间
// - 需要定期清理过期的软删除数据
// - 某些操作可能需要硬删除（物理删除）
func (m *BaseModel) IsDeleted() bool {
	// DeletedAt.Valid为true表示已删除
	// GORM的软删除机制会自动设置此字段
	return m.DeletedAt.Valid
}

// GetTableName 获取表名（子类可以重写）
//
// 功能说明：
// 1. 获取模型的数据库表名
// 2. 子类可以重写此方法自定义表名
// 3. 用于GORM的数据库操作
//
// 表名规则：
// - 默认：GORM根据模型名称自动生成表名（复数形式）
// - 自定义：子类可以重写此方法返回自定义表名
// - 示例：User模型默认表名为"users"，可以自定义为"user"
//
// 使用场景：
// - 需要自定义表名（如历史表、分表等）
// - 需要兼容现有数据库表名
// - 需要多租户表名隔离
//
// 注意事项：
// - 表名应该在模型定义时确定，不应该动态变化
// - 表名应该符合数据库命名规范
// - 自定义表名需要确保数据库中存在该表
func (m *BaseModel) GetTableName() string {
	// 默认返回空字符串，使用GORM的默认表名规则
	// 子类可以重写此方法返回自定义表名
	return ""
}
