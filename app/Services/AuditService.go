package Services

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"time"
)

// AuditService 审计服务
type AuditService struct {
	storageManager *Storage.StorageManager
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Username    string    `json:"username" gorm:"size:50"`
	Action      string    `json:"action" gorm:"size:100;not null"`
	Resource    string    `json:"resource" gorm:"size:50"`
	ResourceID  uint      `json:"resource_id"`
	Description string    `json:"description" gorm:"size:500"`
	IPAddress   string    `json:"ip_address" gorm:"size:45"`
	UserAgent   string    `json:"user_agent" gorm:"size:500"`
	RequestData string    `json:"request_data" gorm:"type:text"`
	ResponseData string   `json:"response_data" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20;default:'success'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewAuditService 创建审计服务
func NewAuditService() *AuditService {
	return &AuditService{}
}

// LogUserAction 记录用户操作
func (s *AuditService) LogUserAction(user *Models.User, userID uint, username, action, resource string, resourceID uint, description string) error {
	auditLog := &AuditLog{
		UserID:      userID,
		Username:    username,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		Status:      "success",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 如果有用户信息，补充用户ID和用户名
	if user != nil {
		auditLog.UserID = user.ID
		auditLog.Username = user.Username
	}

	return s.saveAuditLog(auditLog)
}

// LogSystemEvent 记录系统事件
func (s *AuditService) LogSystemEvent(level, event, description string, metadata map[string]interface{}) error {
	auditLog := &AuditLog{
		UserID:      0, // 系统事件
		Username:    "system",
		Action:      event,
		Resource:    "system",
		ResourceID:  0,
		Description: description,
		Status:      level,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 如果有元数据，转换为JSON存储
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			auditLog.RequestData = string(data)
		}
	}

	return s.saveAuditLog(auditLog)
}

// saveAuditLog 保存审计日志
func (s *AuditService) saveAuditLog(auditLog *AuditLog) error {
	// 保存到数据库
	if err := Database.DB.Create(auditLog).Error; err != nil {
		return fmt.Errorf("保存审计日志到数据库失败: %v", err)
	}

	// 同时记录到文件日志
	logData := map[string]interface{}{
		"audit_id":    auditLog.ID,
		"user_id":     auditLog.UserID,
		"username":    auditLog.Username,
		"action":      auditLog.Action,
		"resource":    auditLog.Resource,
		"resource_id": auditLog.ResourceID,
		"description": auditLog.Description,
		"status":      auditLog.Status,
		"created_at":  auditLog.CreatedAt,
	}

	// 根据状态选择日志级别
	switch auditLog.Status {
	case "error", "critical":
		if s.storageManager != nil {
			s.storageManager.LogError("审计日志", logData)
		}
	case "warning":
		if s.storageManager != nil {
			s.storageManager.LogWarning("审计日志", logData)
		}
	default:
		if s.storageManager != nil {
			s.storageManager.LogInfo("审计日志", logData)
		}
	}

	return nil
}

// GetAuditLogs 获取审计日志
func (s *AuditService) GetAuditLogs(page, limit int, filters map[string]interface{}) ([]AuditLog, int64, error) {
	query := Database.DB.Model(&AuditLog{})

	// 应用筛选条件
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if username, ok := filters["username"].(string); ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}

	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("created_at >= ?", startTime)
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("created_at <= ?", endTime)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	var logs []AuditLog
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error

	return logs, total, err
}
