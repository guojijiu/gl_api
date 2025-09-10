package Models

import (
	"encoding/json"
	"time"
)

// AuditLog 审计日志模型
// 功能说明：
// 1. 记录用户的重要操作（登录、注册、密码修改、数据变更等）
// 2. 支持操作类型分类和级别设置
// 3. 记录操作前后的数据变化
// 4. 支持IP地址和用户代理记录
// 5. 用于安全审计和合规要求
type AuditLog struct {
	BaseModel
	UserID      uint        `json:"user_id" gorm:"index"`                    // 操作用户ID
	Username    string      `json:"username" gorm:"size:50"`                 // 操作用户名
	Action      string      `json:"action" gorm:"size:100;not null"`        // 操作类型
	Level       string      `json:"level" gorm:"size:20;default:'info'"`    // 日志级别：info, warning, error
	Resource    string      `json:"resource" gorm:"size:100"`               // 操作资源（如：user, post, category）
	ResourceID  uint        `json:"resource_id" gorm:"index"`               // 资源ID
	Description string      `json:"description" gorm:"size:500"`            // 操作描述
	IPAddress   string      `json:"ip_address" gorm:"size:45"`              // IP地址
	UserAgent   string      `json:"user_agent" gorm:"size:500"`             // 用户代理
	RequestID   string      `json:"request_id" gorm:"size:100"`             // 请求ID
	Status      string      `json:"status" gorm:"size:20;default:'success'"` // 操作状态：success, failed
	ErrorMsg    string      `json:"error_msg" gorm:"size:500"`              // 错误信息
	BeforeData  string      `json:"before_data" gorm:"type:text"`           // 操作前数据（JSON格式）
	AfterData   string      `json:"after_data" gorm:"type:text"`            // 操作后数据（JSON格式）
	Metadata    string      `json:"metadata" gorm:"type:text"`              // 额外元数据（JSON格式）
	CreatedAt   time.Time   `json:"created_at" gorm:"index"`                // 创建时间
	
	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogLevel 审计日志级别
const (
	AuditLevelInfo    = "info"
	AuditLevelWarning = "warning"
	AuditLevelError   = "error"
)

// AuditAction 审计操作类型
const (
	// 用户相关操作
	AuditActionUserLogin        = "user.login"
	AuditActionUserLogout       = "user.logout"
	AuditActionUserRegister     = "user.register"
	AuditActionUserUpdate       = "user.update"
	AuditActionUserDelete       = "user.delete"
	AuditActionPasswordReset    = "user.password_reset"
	AuditActionEmailVerify      = "user.email_verify"
	
	// 文章相关操作
	AuditActionPostCreate       = "post.create"
	AuditActionPostUpdate       = "post.update"
	AuditActionPostDelete       = "post.delete"
	AuditActionPostPublish      = "post.publish"
	AuditActionPostUnpublish    = "post.unpublish"
	
	// 分类相关操作
	AuditActionCategoryCreate   = "category.create"
	AuditActionCategoryUpdate   = "category.update"
	AuditActionCategoryDelete   = "category.delete"
	
	// 标签相关操作
	AuditActionTagCreate        = "tag.create"
	AuditActionTagUpdate        = "tag.update"
	AuditActionTagDelete        = "tag.delete"
	
	// 系统相关操作
	AuditActionSystemConfig     = "system.config"
	AuditActionSystemBackup     = "system.backup"
	AuditActionSystemRestore    = "system.restore"
	AuditActionSystemMaintenance = "system.maintenance"
)

// AuditStatus 审计状态
const (
	AuditStatusSuccess = "success"
	AuditStatusFailed  = "failed"
)

// NewAuditLog 创建新的审计日志
func NewAuditLog(userID uint, username, action, resource string, resourceID uint) *AuditLog {
	return &AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Level:      AuditLevelInfo,
		Status:     AuditStatusSuccess,
		CreatedAt:  time.Now(),
	}
}

// SetLevel 设置日志级别
func (a *AuditLog) SetLevel(level string) *AuditLog {
	a.Level = level
	return a
}

// SetDescription 设置操作描述
func (a *AuditLog) SetDescription(description string) *AuditLog {
	a.Description = description
	return a
}

// SetIPAddress 设置IP地址
func (a *AuditLog) SetIPAddress(ipAddress string) *AuditLog {
	a.IPAddress = ipAddress
	return a
}

// SetUserAgent 设置用户代理
func (a *AuditLog) SetUserAgent(userAgent string) *AuditLog {
	a.UserAgent = userAgent
	return a
}

// SetRequestID 设置请求ID
func (a *AuditLog) SetRequestID(requestID string) *AuditLog {
	a.RequestID = requestID
	return a
}

// SetError 设置错误信息
func (a *AuditLog) SetError(errorMsg string) *AuditLog {
	a.ErrorMsg = errorMsg
	a.Status = AuditStatusFailed
	a.Level = AuditLevelError
	return a
}

// SetBeforeData 设置操作前数据
func (a *AuditLog) SetBeforeData(data interface{}) *AuditLog {
	if jsonData, err := json.Marshal(data); err == nil {
		a.BeforeData = string(jsonData)
	}
	return a
}

// SetAfterData 设置操作后数据
func (a *AuditLog) SetAfterData(data interface{}) *AuditLog {
	if jsonData, err := json.Marshal(data); err == nil {
		a.AfterData = string(jsonData)
	}
	return a
}

// SetMetadata 设置元数据
func (a *AuditLog) SetMetadata(metadata interface{}) *AuditLog {
	if jsonData, err := json.Marshal(metadata); err == nil {
		a.Metadata = string(jsonData)
	}
	return a
}
