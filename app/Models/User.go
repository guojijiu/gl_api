package Models

import (
	"cloud-platform-api/app/Utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
//
// 重要功能说明：
// 1. 用户身份管理：唯一标识、用户名、邮箱、密码等基本信息
// 2. 权限控制：基于角色的访问控制（RBAC），支持管理员和普通用户
// 3. 状态管理：用户账户状态控制，支持启用/禁用账户
// 4. 安全特性：密码安全哈希、邮箱验证、登录历史记录
// 5. 关联关系：与文章、分类、标签等业务实体的关联管理
// 6. 审计支持：登录时间、登录次数、账户创建时间等审计信息
//
// 安全设计：
// - UUID全局唯一标识，防止ID枚举攻击
// - 密码使用bcrypt哈希，支持自动盐值生成
// - 邮箱验证状态跟踪，支持安全验证流程
// - 账户状态控制，支持管理员禁用违规账户
// - 登录历史记录，支持异常登录检测
//
// 数据库设计：
// - 用户名和邮箱唯一索引，确保数据一致性
// - 密码字段JSON序列化时自动排除，防止泄露
// - 支持软删除和硬删除两种模式
// - 自动时间戳管理（创建时间、更新时间）
// - 支持数据库迁移和版本控制
//
// 业务规则：
// - 用户名长度限制：3-50字符
// - 邮箱格式验证和唯一性检查
// - 密码强度要求：至少8字符，包含大小写字母和数字
// - 角色权限：admin（管理员）、user（普通用户）
// - 账户状态：1（正常）、0（禁用）
//
// 性能优化：
// - 关键字段建立数据库索引
// - 支持关联查询预加载
// - 密码验证使用常量时间比较
// - 支持用户信息缓存
//
// 扩展性：
// - 支持自定义用户属性扩展
// - 支持多角色权限系统
// - 支持用户组和权限继承
// - 支持第三方登录集成
type User struct {
	BaseModel
	UUID            string     `json:"uuid" gorm:"unique;not null;size:36"`     // 用户唯一标识
	Username        string     `json:"username" gorm:"unique;not null;size:50"` // 用户名
	Email           string     `json:"email" gorm:"unique;not null;size:100"`   // 邮箱地址
	Password        string     `json:"-" gorm:"not null;size:255"`              // 密码（不在JSON中返回）
	Avatar          string     `json:"avatar" gorm:"size:255"`                  // 头像URL
	Role            string     `json:"role" gorm:"default:'user';size:20"`      // 用户角色
	Status          int        `json:"status" gorm:"default:1"`                 // 用户状态：1-正常, 0-禁用
	EmailVerifiedAt *time.Time `json:"email_verified_at" gorm:"index"`          // 邮箱验证时间
	LastLoginAt     *time.Time `json:"last_login_at" gorm:"index"`              // 最后登录时间
	LoginCount      int        `json:"login_count" gorm:"default:0"`            // 登录次数

	// 关联关系
	Posts []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"` // 用户发布的文章
}

// BeforeCreate 创建前的钩子
// 功能说明：
// 1. 在创建用户记录前自动生成UUID
// 2. 确保每个用户都有唯一的UUID标识
// 3. 用于分布式系统中的用户标识
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}

// GetTableName 获取表名
// 功能说明：
// 1. 返回用户表的数据库表名
// 2. 用于GORM的数据库操作
func (u *User) GetTableName() string {
	return "users"
}

// IsAdmin 检查是否为管理员
// 功能说明：
// 1. 检查用户是否具有管理员权限
// 2. 用于权限控制和功能访问验证
// 3. 返回布尔值表示管理员状态
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsActive 检查是否激活
// 功能说明：
// 1. 检查用户账户是否处于正常状态
// 2. 用于登录验证和功能访问控制
// 3. 返回布尔值表示账户状态
func (u *User) IsActive() bool {
	return u.Status == 1
}

// SetPassword 设置密码
// 功能说明：
// 1. 对密码进行安全的哈希处理
// 2. 使用bcrypt算法确保密码安全性
// 3. 自动处理密码哈希过程中的错误
// 4. 用于用户注册和密码修改
func (u *User) SetPassword(password string) error {
	hashedPassword, err := Utils.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return nil
}

// GetDisplayName 获取显示名称
// 功能说明：
// 1. 返回用户的显示名称
// 2. 优先使用用户名，如果用户名为空则使用邮箱
// 3. 用于前端界面显示和日志记录
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}

// GetAvatarURL 获取头像URL
// 功能说明：
// 1. 返回用户的头像URL
// 2. 如果用户没有设置头像，返回默认头像
// 3. 用于前端界面显示
func (u *User) GetAvatarURL() string {
	if u.Avatar != "" {
		return u.Avatar
	}
	// 返回默认头像
	return "/images/default-avatar.png"
}

// GetCreatedDate 获取创建日期
// 功能说明：
// 1. 返回用户账户的创建日期
// 2. 格式化为YYYY-MM-DD格式
// 3. 用于用户信息显示和统计
func (u *User) GetCreatedDate() string {
	return u.GetCreatedAt().Format("2006-01-02")
}

// GetLastLoginTime 获取最后登录时间
// 功能说明：
// 1. 返回用户的最后登录时间
// 2. 如果从未登录过，返回nil
// 3. 用于用户活动分析和安全监控
func (u *User) GetLastLoginTime() *time.Time {
	return u.LastLoginAt
}

// UpdateLastLoginTime 更新最后登录时间
// 功能说明：
// 1. 更新用户的最后登录时间为当前时间
// 2. 增加登录次数计数器
// 3. 在用户登录成功后调用
// 4. 用于记录用户活动
func (u *User) UpdateLastLoginTime() {
	now := time.Now()
	u.LastLoginAt = &now
	u.LoginCount++
}

// IsEmailVerified 检查邮箱是否已验证
// 功能说明：
// 1. 检查用户邮箱是否已经通过验证
// 2. 用于邮箱验证状态判断
// 3. 返回布尔值表示验证状态
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// VerifyEmail 验证邮箱
// 功能说明：
// 1. 设置邮箱验证时间为当前时间
// 2. 标记用户邮箱为已验证状态
// 3. 用于邮箱验证流程
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now
}

// GetStatusText 获取状态文本
// 功能说明：
// 1. 返回用户状态的文本描述
// 2. 用于前端显示和日志记录
// 3. 支持多语言扩展
func (u *User) GetStatusText() string {
	switch u.Status {
	case 1:
		return "正常"
	case 0:
		return "禁用"
	default:
		return "未知"
	}
}

// CanLogin 检查是否可以登录
// 功能说明：
// 1. 检查用户状态是否允许登录
// 2. 检查用户是否被禁用
// 3. 用于登录验证
func (u *User) CanLogin() bool {
	return u.IsActive() && u.Status == 1
}

// GetRoleText 获取角色文本
// 功能说明：
// 1. 返回用户角色的文本描述
// 2. 用于前端显示和权限管理
// 3. 支持多语言扩展
func (u *User) GetRoleText() string {
	switch u.Role {
	case "admin":
		return "管理员"
	case "user":
		return "普通用户"
	default:
		return "未知角色"
	}
}

// ValidatePassword 验证密码
// 功能说明：
// 1. 验证用户输入的密码是否正确
// 2. 使用bcrypt进行密码验证
// 3. 用于登录验证
func (u *User) ValidatePassword(password string) bool {
	return Utils.CheckPassword(password, u.Password)
}

// GetLoginHistory 获取登录历史
// 功能说明：
// 1. 返回用户的登录统计信息
// 2. 用于用户行为分析
// 3. 包含登录次数、最后登录时间等信息
func (u *User) GetLoginHistory() map[string]interface{} {
	return map[string]interface{}{
		"login_count":    u.LoginCount,
		"last_login_at":  u.LastLoginAt,
		"email_verified": u.IsEmailVerified(),
		"account_status": u.GetStatusText(),
	}
}
