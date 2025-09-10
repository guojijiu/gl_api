package Services

import (
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Storage"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SecurityAuditService 安全审计服务
// 功能说明：
// 1. API密钥管理和验证
// 2. 安全事件审计日志
// 3. 安全扫描和漏洞检测
// 4. 访问控制和安全策略
// 5. 安全报告生成
type SecurityAuditService struct {
	storageManager *Storage.StorageManager
	apiKeys        map[string]*APIKey
	mutex          sync.RWMutex
	config         *SecurityAuditConfig
}

// SecurityAuditConfig 安全审计配置
type SecurityAuditConfig struct {
	EnableAPIKeyAuth     bool          `json:"enable_api_key_auth"`     // 启用API密钥认证
	APIKeyExpireTime     time.Duration `json:"api_key_expire_time"`     // API密钥过期时间
	MaxAPIKeysPerUser    int           `json:"max_api_keys_per_user"`   // 每个用户最大API密钥数
	EnableSecurityScan   bool          `json:"enable_security_scan"`    // 启用安全扫描
	ScanInterval         time.Duration `json:"scan_interval"`           // 扫描间隔
	EnableAuditLog       bool          `json:"enable_audit_log"`        // 启用审计日志
	AuditLogRetention    time.Duration `json:"audit_log_retention"`     // 审计日志保留时间
	EnableRateLimit      bool          `json:"enable_rate_limit"`       // 启用速率限制
	MaxRequestsPerMinute int           `json:"max_requests_per_minute"` // 每分钟最大请求数
}

// APIKey API密钥信息
type APIKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	KeyHash     string    `json:"-" gorm:"size:64;uniqueIndex"`
	Permissions []string  `json:"permissions" gorm:"serializer:json"`
	LastUsed    time.Time `json:"last_used"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// SecurityEvent 安全事件
type SecurityEvent struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	EventType   string                 `json:"event_type" gorm:"size:50"`
	UserID      uint                   `json:"user_id"`
	IPAddress   string                 `json:"ip_address" gorm:"size:45"`
	UserAgent   string                 `json:"user_agent" gorm:"size:500"`
	Resource    string                 `json:"resource" gorm:"size:200"`
	Action      string                 `json:"action" gorm:"size:50"`
	Status      string                 `json:"status" gorm:"size:20"`
	Details     map[string]interface{} `json:"details" gorm:"serializer:json"`
	RiskLevel   string                 `json:"risk_level" gorm:"size:20"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SecurityScanResult 安全扫描结果
type SecurityScanResult struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	ScanType    string                 `json:"scan_type" gorm:"size:50"`
	Target      string                 `json:"target" gorm:"size:200"`
	Vulnerabilities []Vulnerability    `json:"vulnerabilities" gorm:"serializer:json"`
	RiskScore   int                    `json:"risk_score"`
	Status      string                 `json:"status" gorm:"size:20"`
	ScanTime    time.Time              `json:"scan_time"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Vulnerability 漏洞信息
type Vulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Recommendation string `json:"recommendation"`
}

// NewSecurityAuditService 创建安全审计服务
// 功能说明：
// 1. 初始化安全审计服务
// 2. 设置默认安全配置
// 3. 启动安全监控
// 4. 定期安全扫描
func NewSecurityAuditService(storageManager *Storage.StorageManager) *SecurityAuditService {
	config := &SecurityAuditConfig{
		EnableAPIKeyAuth:     true,
		APIKeyExpireTime:     365 * 24 * time.Hour, // 1年
		MaxAPIKeysPerUser:    10,
		EnableSecurityScan:   true,
		ScanInterval:         24 * time.Hour, // 24小时
		EnableAuditLog:       true,
		AuditLogRetention:    90 * 24 * time.Hour, // 90天
		EnableRateLimit:      true,
		MaxRequestsPerMinute: 100,
	}

	service := &SecurityAuditService{
		storageManager: storageManager,
		apiKeys:        make(map[string]*APIKey),
		config:         config,
	}

	// 启动安全监控
	if config.EnableSecurityScan {
		go service.startSecurityScanning()
	}

	// 启动审计日志清理
	if config.EnableAuditLog {
		go service.startAuditLogCleanup()
	}

	return service
}

// CreateAPIKey 创建API密钥
// 功能说明：
// 1. 为用户生成新的API密钥
// 2. 设置密钥权限和过期时间
// 3. 安全存储密钥哈希
// 4. 返回密钥信息
func (sas *SecurityAuditService) CreateAPIKey(userID uint, name string, permissions []string) (*APIKey, string, error) {
	sas.mutex.Lock()
	defer sas.mutex.Unlock()

	// 检查用户API密钥数量限制
	existingKeys, err := sas.getUserAPIKeys(userID)
	if err != nil {
		return nil, "", err
	}

	if len(existingKeys) >= sas.config.MaxAPIKeysPerUser {
		return nil, "", fmt.Errorf("用户已达到最大API密钥数量限制: %d", sas.config.MaxAPIKeysPerUser)
	}

	// 生成API密钥
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}
	apiKey := hex.EncodeToString(keyBytes)

	// 计算密钥哈希
	keyHash := sha256.Sum256([]byte(apiKey))
	keyHashStr := hex.EncodeToString(keyHash[:])

	// 创建API密钥记录
	apiKeyRecord := &APIKey{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHashStr,
		Permissions: permissions,
		LastUsed:    time.Now(),
		ExpiresAt:   time.Now().Add(sas.config.APIKeyExpireTime),
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	// 保存到数据库
	if err := Database.DB.Create(apiKeyRecord).Error; err != nil {
		return nil, "", err
	}

	// 添加到内存缓存
	sas.apiKeys[keyHashStr] = apiKeyRecord

	// 记录安全事件
	sas.logSecurityEvent(&SecurityEvent{
		ID:        uuid.New().String(),
		EventType: "api_key_created",
		UserID:    userID,
		Resource:  "api_key",
		Action:    "create",
		Status:    "success",
		Details: map[string]interface{}{
			"key_name": name,
			"permissions": permissions,
		},
		RiskLevel: "low",
		CreatedAt: time.Now(),
	})

	return apiKeyRecord, apiKey, nil
}

// ValidateAPIKey 验证API密钥
// 功能说明：
// 1. 验证API密钥的有效性
// 2. 检查密钥是否过期
// 3. 验证用户权限
// 4. 更新最后使用时间
func (sas *SecurityAuditService) ValidateAPIKey(apiKey string, requiredPermissions []string) (*APIKey, error) {
	// 计算密钥哈希
	keyHash := sha256.Sum256([]byte(apiKey))
	keyHashStr := hex.EncodeToString(keyHash[:])

	sas.mutex.RLock()
	keyRecord, exists := sas.apiKeys[keyHashStr]
	sas.mutex.RUnlock()

	if !exists {
		// 从数据库加载
		if err := Database.DB.Where("key_hash = ?", keyHashStr).First(&keyRecord).Error; err != nil {
			return nil, fmt.Errorf("无效的API密钥")
		}
		sas.mutex.Lock()
		sas.apiKeys[keyHashStr] = keyRecord
		sas.mutex.Unlock()
	}

	// 检查密钥是否激活
	if !keyRecord.IsActive {
		return nil, fmt.Errorf("API密钥已被禁用")
	}

	// 检查密钥是否过期
	if time.Now().After(keyRecord.ExpiresAt) {
		return nil, fmt.Errorf("API密钥已过期")
	}

	// 检查权限
	if !sas.hasPermissions(keyRecord.Permissions, requiredPermissions) {
		return nil, fmt.Errorf("API密钥权限不足")
	}

	// 更新最后使用时间
	keyRecord.LastUsed = time.Now()
	Database.DB.Save(keyRecord)

	return keyRecord, nil
}

// hasPermissions 检查是否具有所需权限
func (sas *SecurityAuditService) hasPermissions(userPermissions, requiredPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}

	permissionMap := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionMap[perm] = true
	}

	for _, required := range requiredPermissions {
		if !permissionMap[required] {
			return false
		}
	}

	return true
}

// RevokeAPIKey 撤销API密钥
// 功能说明：
// 1. 禁用指定的API密钥
// 2. 从内存缓存中移除
// 3. 记录撤销事件
func (sas *SecurityAuditService) RevokeAPIKey(keyID string, userID uint) error {
	sas.mutex.Lock()
	defer sas.mutex.Unlock()

	// 查找并禁用密钥
	var keyRecord APIKey
	if err := Database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&keyRecord).Error; err != nil {
		return fmt.Errorf("API密钥不存在")
	}

	keyRecord.IsActive = false
	if err := Database.DB.Save(&keyRecord).Error; err != nil {
		return err
	}

	// 从内存缓存中移除
	delete(sas.apiKeys, keyRecord.KeyHash)

	// 记录安全事件
	sas.logSecurityEvent(&SecurityEvent{
		ID:        uuid.New().String(),
		EventType: "api_key_revoked",
		UserID:    userID,
		Resource:  "api_key",
		Action:    "revoke",
		Status:    "success",
		Details: map[string]interface{}{
			"key_id": keyID,
		},
		RiskLevel: "low",
		CreatedAt: time.Now(),
	})

	return nil
}

// GetUserAPIKeys 获取用户的API密钥列表
func (sas *SecurityAuditService) GetUserAPIKeys(userID uint) ([]*APIKey, error) {
	return sas.getUserAPIKeys(userID)
}

// getUserAPIKeys 内部方法：获取用户API密钥
func (sas *SecurityAuditService) getUserAPIKeys(userID uint) ([]*APIKey, error) {
	var keys []*APIKey
	if err := Database.DB.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// LogSecurityEvent 记录安全事件
// 功能说明：
// 1. 记录各种安全事件
// 2. 评估事件风险等级
// 3. 存储到数据库和日志
// 4. 触发安全告警
func (sas *SecurityAuditService) LogSecurityEvent(event *SecurityEvent) {
	if !sas.config.EnableAuditLog {
		return
	}

	// 设置事件ID
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// 评估风险等级
	if event.RiskLevel == "" {
		event.RiskLevel = sas.assessRiskLevel(event)
	}

	// 保存到数据库
	if err := Database.DB.Create(event).Error; err != nil {
		log.Printf("保存安全事件失败: %v", err)
	}

	// 记录到日志
	sas.storageManager.LogWarning("安全事件", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"user_id":    event.UserID,
		"ip_address": event.IPAddress,
		"resource":   event.Resource,
		"action":     event.Action,
		"status":     event.Status,
		"risk_level": event.RiskLevel,
		"details":    event.Details,
	})

	// 高风险事件触发告警
	if event.RiskLevel == "high" || event.RiskLevel == "critical" {
		sas.triggerSecurityAlert(event)
	}
}

// logSecurityEvent 内部方法：记录安全事件
func (sas *SecurityAuditService) logSecurityEvent(event *SecurityEvent) {
	sas.LogSecurityEvent(event)
}

// assessRiskLevel 评估事件风险等级
func (sas *SecurityAuditService) assessRiskLevel(event *SecurityEvent) string {
	// 根据事件类型和详情评估风险等级
	switch event.EventType {
	case "login_failed", "unauthorized_access":
		return "medium"
	case "api_key_compromised", "sql_injection_attempt":
		return "high"
	case "admin_access", "data_breach":
		return "critical"
	default:
		return "low"
	}
}

// triggerSecurityAlert 触发安全告警
func (sas *SecurityAuditService) triggerSecurityAlert(event *SecurityEvent) {
	// 发送邮件告警
	_ = NewEmailService(&EmailConfig{
		Host:     "localhost",
		Port:     587,
		Username: "security@example.com",
		Password: "password",
		From:     "security@example.com",
		UseTLS:   true,
	})

	subject := fmt.Sprintf("安全告警: %s", event.EventType)
	_ = fmt.Sprintf(`
		安全事件详情:
		- 事件类型: %s
		- 用户ID: %d
		- IP地址: %s
		- 资源: %s
		- 操作: %s
		- 风险等级: %s
		- 时间: %s
	`, event.EventType, event.UserID, event.IPAddress, event.Resource, event.Action, event.RiskLevel, event.CreatedAt.Format(time.RFC3339))

	// 这里应该发送给安全管理员
	// emailService.SendEmail("admin@example.com", subject, body)

	// 记录告警日志
	sas.storageManager.LogError("安全告警触发", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"risk_level": event.RiskLevel,
		"subject":    subject,
	})
}

// PerformSecurityScan 执行安全扫描
// 功能说明：
// 1. 扫描系统安全漏洞
// 2. 检查配置安全性
// 3. 生成扫描报告
// 4. 提供修复建议
func (sas *SecurityAuditService) PerformSecurityScan() (*SecurityScanResult, error) {
	if !sas.config.EnableSecurityScan {
		return nil, fmt.Errorf("安全扫描功能已禁用")
	}

	result := &SecurityScanResult{
		ID:       uuid.New().String(),
		ScanType: "comprehensive",
		Target:   "system",
		Status:   "completed",
		ScanTime: time.Now(),
		CreatedAt: time.Now(),
	}

	// 执行各种安全检查
	vulnerabilities := []Vulnerability{}

	// 检查数据库连接安全
	if err := sas.checkDatabaseSecurity(&vulnerabilities); err != nil {
		log.Printf("数据库安全检查失败: %v", err)
	}

	// 检查API密钥安全
	if err := sas.checkAPIKeySecurity(&vulnerabilities); err != nil {
		log.Printf("API密钥安全检查失败: %v", err)
	}

	// 检查配置安全
	if err := sas.checkConfigurationSecurity(&vulnerabilities); err != nil {
		log.Printf("配置安全检查失败: %v", err)
	}

	result.Vulnerabilities = vulnerabilities
	result.RiskScore = sas.calculateRiskScore(vulnerabilities)

	// 保存扫描结果
	if err := Database.DB.Create(result).Error; err != nil {
		log.Printf("保存安全扫描结果失败: %v", err)
	}

	// 记录扫描完成事件
	sas.logSecurityEvent(&SecurityEvent{
		ID:        uuid.New().String(),
		EventType: "security_scan_completed",
		Resource:  "system",
		Action:    "scan",
		Status:    "success",
		Details: map[string]interface{}{
			"scan_id":     result.ID,
			"vulnerabilities_count": len(vulnerabilities),
			"risk_score":  result.RiskScore,
		},
		RiskLevel: "low",
		CreatedAt: time.Now(),
	})

	return result, nil
}

// checkDatabaseSecurity 检查数据库安全
func (sas *SecurityAuditService) checkDatabaseSecurity(vulnerabilities *[]Vulnerability) error {
	// 检查数据库连接是否使用SSL
	// 这里简化处理，实际应该检查具体的数据库配置
	*vulnerabilities = append(*vulnerabilities, Vulnerability{
		Type:        "database_security",
		Severity:    "medium",
		Description: "建议启用数据库SSL连接",
		Location:    "database_config",
		Recommendation: "在数据库配置中启用SSL连接",
	})
	return nil
}

// checkAPIKeySecurity 检查API密钥安全
func (sas *SecurityAuditService) checkAPIKeySecurity(vulnerabilities *[]Vulnerability) error {
	// 检查过期的API密钥
	var expiredKeys []APIKey
	if err := Database.DB.Where("expires_at < ? AND is_active = ?", time.Now(), true).Find(&expiredKeys).Error; err != nil {
		return err
	}

	if len(expiredKeys) > 0 {
		*vulnerabilities = append(*vulnerabilities, Vulnerability{
			Type:        "api_key_security",
			Severity:    "medium",
			Description: fmt.Sprintf("发现 %d 个过期的API密钥", len(expiredKeys)),
			Location:    "api_keys",
			Recommendation: "清理过期的API密钥",
		})
	}

	return nil
}

// checkConfigurationSecurity 检查配置安全
func (sas *SecurityAuditService) checkConfigurationSecurity(vulnerabilities *[]Vulnerability) error {
	// 检查默认配置
	*vulnerabilities = append(*vulnerabilities, Vulnerability{
		Type:        "configuration_security",
		Severity:    "low",
		Description: "建议定期更新安全配置",
		Location:    "app_config",
		Recommendation: "定期审查和更新安全配置参数",
	})
	return nil
}

// calculateRiskScore 计算风险评分
func (sas *SecurityAuditService) calculateRiskScore(vulnerabilities []Vulnerability) int {
	score := 0
	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case "critical":
			score += 10
		case "high":
			score += 7
		case "medium":
			score += 4
		case "low":
			score += 1
		}
	}
	return score
}

// startSecurityScanning 启动定期安全扫描
func (sas *SecurityAuditService) startSecurityScanning() {
	ticker := time.NewTicker(sas.config.ScanInterval)
	defer ticker.Stop()

	for range ticker.C {
		if _, err := sas.PerformSecurityScan(); err != nil {
			log.Printf("定期安全扫描失败: %v", err)
		}
	}
}

// startAuditLogCleanup 启动审计日志清理
func (sas *SecurityAuditService) startAuditLogCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoffTime := time.Now().Add(-sas.config.AuditLogRetention)
		
		// 清理过期的安全事件
		if err := Database.DB.Where("created_at < ?", cutoffTime).Delete(&SecurityEvent{}).Error; err != nil {
			log.Printf("清理审计日志失败: %v", err)
		}

		// 清理过期的扫描结果
		if err := Database.DB.Where("created_at < ?", cutoffTime).Delete(&SecurityScanResult{}).Error; err != nil {
			log.Printf("清理扫描结果失败: %v", err)
		}
	}
}

// GetSecurityReport 获取安全报告
// 功能说明：
// 1. 生成安全状态报告
// 2. 包含事件统计和趋势
// 3. 提供安全建议
// 4. 支持定期报告生成
func (sas *SecurityAuditService) GetSecurityReport() map[string]interface{} {
	// 获取最近的安全事件统计
	var eventStats []struct {
		EventType string `json:"event_type"`
		Count     int64  `json:"count"`
		RiskLevel string `json:"risk_level"`
	}

	Database.DB.Model(&SecurityEvent{}).
		Select("event_type, count(*) as count, risk_level").
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Group("event_type, risk_level").
		Find(&eventStats)

	// 获取API密钥统计
	var apiKeyCount int64
	Database.DB.Model(&APIKey{}).Where("is_active = ?", true).Count(&apiKeyCount)

	// 获取最近的扫描结果
	var lastScan SecurityScanResult
	Database.DB.Order("created_at desc").First(&lastScan)

	report := map[string]interface{}{
		"generated_at":     time.Now(),
		"event_statistics": eventStats,
		"active_api_keys":  apiKeyCount,
		"last_scan":        lastScan,
		"security_status":  sas.getSecurityStatus(),
	}

	return report
}

// getSecurityStatus 获取安全状态
func (sas *SecurityAuditService) getSecurityStatus() string {
	// 检查最近的高风险事件
	var highRiskEvents int64
	Database.DB.Model(&SecurityEvent{}).
		Where("risk_level IN (?) AND created_at > ?", []string{"high", "critical"}, time.Now().AddDate(0, 0, -7)).
		Count(&highRiskEvents)

	if highRiskEvents > 10 {
		return "critical"
	} else if highRiskEvents > 5 {
		return "warning"
	} else {
		return "secure"
	}
}
