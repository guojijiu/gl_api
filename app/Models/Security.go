package Models

import (
	"time"

	"gorm.io/gorm"
)

// SecurityEvent 安全事件模型
type SecurityEvent struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	EventType       string         `json:"event_type" gorm:"type:varchar(50);not null;index"`           // 事件类型
	EventLevel      string         `json:"event_level" gorm:"type:varchar(20);not null;index"`         // 事件级别
	UserID          *uint          `json:"user_id" gorm:"index"`                                         // 用户ID
	Username        string         `json:"username" gorm:"type:varchar(100)"`                          // 用户名
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45);index"`                   // IP地址
	UserAgent       string         `json:"user_agent" gorm:"type:text"`                                 // 用户代理
	Resource        string         `json:"resource" gorm:"type:varchar(255)"`                          // 资源
	Action          string         `json:"action" gorm:"type:varchar(100)"`                             // 操作
	Details         string         `json:"details" gorm:"type:text"`                                     // 详细信息
	RiskScore       float64        `json:"risk_score" gorm:"type:decimal(5,2);default:0"`              // 风险评分
	AnomalyScore    float64        `json:"anomaly_score" gorm:"type:decimal(5,2);default:0"`           // 异常评分
	Blocked         bool           `json:"blocked" gorm:"default:false"`                               // 是否被阻止
	Alerted         bool           `json:"alerted" gorm:"default:false"`                                 // 是否已告警
	Location        string         `json:"location" gorm:"type:varchar(255)"`                           // 地理位置
	DeviceInfo      string         `json:"device_info" gorm:"type:text"`                                 // 设备信息
	SessionID       string         `json:"session_id" gorm:"type:varchar(100);index"`                    // 会话ID
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// ThreatIntelligence 威胁情报模型
type ThreatIntelligence struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Source          string         `json:"source" gorm:"type:varchar(100);not null"`                   // 情报源
	ThreatType      string         `json:"threat_type" gorm:"type:varchar(50);not null;index"`         // 威胁类型
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45);index"`                   // IP地址
	Domain          string         `json:"domain" gorm:"type:varchar(255);index"`                       // 域名
	URL             string         `json:"url" gorm:"type:text"`                                        // URL
	Hash            string         `json:"hash" gorm:"type:varchar(64);index"`                          // 文件哈希
	Confidence      float64        `json:"confidence" gorm:"type:decimal(5,2);default:0"`              // 置信度
	Severity        string         `json:"severity" gorm:"type:varchar(20);not null;index"`            // 严重程度
	Description     string         `json:"description" gorm:"type:text"`                               // 描述
	Tags            string         `json:"tags" gorm:"type:text"`                                      // 标签
	FirstSeen       time.Time      `json:"first_seen"`                                                  // 首次发现时间
	LastSeen        time.Time      `json:"last_seen"`                                                   // 最后发现时间
	UpdateInterval  time.Duration  `json:"update_interval" gorm:"type:bigint"`                         // 更新间隔
	Active          bool           `json:"active" gorm:"default:true"`                                 // 是否活跃
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AccessControl 访问控制模型
type AccessControl struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`                               // 用户ID
	Resource        string         `json:"resource" gorm:"type:varchar(255);not null;index"`            // 资源
	Action          string         `json:"action" gorm:"type:varchar(100);not null"`                    // 操作
	Permission      string         `json:"permission" gorm:"type:varchar(20);not null;index"`           // 权限(allow/deny)
	Condition       string         `json:"condition" gorm:"type:text"`                                  // 条件
	TimeRestriction string         `json:"time_restriction" gorm:"type:text"`                           // 时间限制
	LocationRestriction string     `json:"location_restriction" gorm:"type:text"`                      // 位置限制
	DeviceRestriction string       `json:"device_restriction" gorm:"type:text"`                        // 设备限制
	Priority        int            `json:"priority" gorm:"default:0"`                                   // 优先级
	Active          bool           `json:"active" gorm:"default:true"`                                  // 是否活跃
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// PasswordHistory 密码历史模型
type PasswordHistory struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`                                // 用户ID
	PasswordHash    string         `json:"password_hash" gorm:"type:varchar(255);not null"`             // 密码哈希
	ChangedAt       time.Time      `json:"changed_at"`                                                  // 更改时间
	ChangedBy       uint           `json:"changed_by" gorm:"index"`                                    // 更改者ID
	Reason          string         `json:"reason" gorm:"type:varchar(100)"`                             // 更改原因
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45)"`                          // IP地址
	UserAgent       string         `json:"user_agent" gorm:"type:text"`                                 // 用户代理
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// LoginAttempt 登录尝试模型
type LoginAttempt struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Username        string         `json:"username" gorm:"type:varchar(100);not null;index"`            // 用户名
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45);not null;index"`           // IP地址
	UserAgent       string         `json:"user_agent" gorm:"type:text"`                                 // 用户代理
	Success         bool           `json:"success" gorm:"not null"`                                    // 是否成功
	FailureReason   string         `json:"failure_reason" gorm:"type:varchar(255)"`                     // 失败原因
	AttemptTime     time.Time      `json:"attempt_time"`                                               // 尝试时间
	Location        string         `json:"location" gorm:"type:varchar(255)"`                          // 地理位置
	DeviceInfo      string         `json:"device_info" gorm:"type:text"`                               // 设备信息
	RiskScore       float64        `json:"risk_score" gorm:"type:decimal(5,2);default:0"`             // 风险评分
	Blocked         bool           `json:"blocked" gorm:"default:false"`                              // 是否被阻止
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AccountLockout 账户锁定模型
type AccountLockout struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`                               // 用户ID
	Username        string         `json:"username" gorm:"type:varchar(100);not null;index"`            // 用户名
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45);not null;index"`          // IP地址
	LockoutType     string         `json:"lockout_type" gorm:"type:varchar(20);not null"`               // 锁定类型
	Reason          string         `json:"reason" gorm:"type:varchar(255)"`                            // 锁定原因
	LockoutTime     time.Time      `json:"lockout_time"`                                               // 锁定时间
	ExpiryTime      time.Time      `json:"expiry_time"`                                                // 过期时间
	AttemptCount    int            `json:"attempt_count" gorm:"default:0"`                             // 尝试次数
	Active          bool           `json:"active" gorm:"default:true"`                                  // 是否活跃
	UnlockedBy      *uint          `json:"unlocked_by" gorm:"index"`                                   // 解锁者ID
	UnlockTime      *time.Time     `json:"unlock_time"`                                                // 解锁时间
	UnlockReason    string         `json:"unlock_reason" gorm:"type:varchar(255)"`                     // 解锁原因
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SecurityPolicy 安全策略模型
type SecurityPolicy struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"type:varchar(100);not null;unique"`              // 策略名称
	Description     string         `json:"description" gorm:"type:text"`                               // 策略描述
	PolicyType      string         `json:"policy_type" gorm:"type:varchar(50);not null;index"`         // 策略类型
	Rules           string         `json:"rules" gorm:"type:text"`                                     // 策略规则(JSON)
	Priority        int            `json:"priority" gorm:"default:0"`                                   // 优先级
	Active          bool           `json:"active" gorm:"default:true"`                                  // 是否活跃
	AppliedTo       string         `json:"applied_to" gorm:"type:varchar(255)"`                         // 应用范围
	CreatedBy       uint           `json:"created_by" gorm:"index"`                                     // 创建者ID
	UpdatedBy       uint           `json:"updated_by" gorm:"index"`                                    // 更新者ID
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SecurityAlert 安全告警模型
type SecurityAlert struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	AlertType       string         `json:"alert_type" gorm:"type:varchar(50);not null;index"`           // 告警类型
	Severity        string         `json:"severity" gorm:"type:varchar(20);not null;index"`             // 严重程度
	Title           string         `json:"title" gorm:"type:varchar(255);not null"`                    // 告警标题
	Description     string         `json:"description" gorm:"type:text"`                               // 告警描述
	Source          string         `json:"source" gorm:"type:varchar(100)"`                             // 告警源
	UserID          *uint          `json:"user_id" gorm:"index"`                                        // 用户ID
	IPAddress       string         `json:"ip_address" gorm:"type:varchar(45);index"`                   // IP地址
	Resource        string         `json:"resource" gorm:"type:varchar(255)"`                           // 资源
	Details         string         `json:"details" gorm:"type:text"`                                    // 详细信息
	RiskScore       float64        `json:"risk_score" gorm:"type:decimal(5,2);default:0"`              // 风险评分
	Status          string         `json:"status" gorm:"type:varchar(20);default:'open';index"`        // 状态
	Acknowledged    bool           `json:"acknowledged" gorm:"default:false"`                          // 是否已确认
	AcknowledgedBy  *uint          `json:"acknowledged_by" gorm:"index"`                              // 确认者ID
	AcknowledgedAt  *time.Time     `json:"acknowledged_at"`                                            // 确认时间
	Resolved        bool           `json:"resolved" gorm:"default:false"`                              // 是否已解决
	ResolvedBy      *uint          `json:"resolved_by" gorm:"index"`                                   // 解决者ID
	ResolvedAt      *time.Time     `json:"resolved_at"`                                                // 解决时间
	ResolutionNotes string         `json:"resolution_notes" gorm:"type:text"`                          // 解决说明
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SecurityReport 安全报告模型
type SecurityReport struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ReportType      string         `json:"report_type" gorm:"type:varchar(50);not null;index"`          // 报告类型
	Title           string         `json:"title" gorm:"type:varchar(255);not null"`                     // 报告标题
	Description     string         `json:"description" gorm:"type:text"`                                // 报告描述
	Period          string         `json:"period" gorm:"type:varchar(50)"`                               // 报告周期
	StartDate       time.Time      `json:"start_date"`                                                  // 开始日期
	EndDate         time.Time      `json:"end_date"`                                                    // 结束日期
	Content         string         `json:"content" gorm:"type:text"`                                    // 报告内容(JSON)
	Summary         string         `json:"summary" gorm:"type:text"`                                    // 报告摘要
	GeneratedBy     uint           `json:"generated_by" gorm:"index"`                                   // 生成者ID
	Status          string         `json:"status" gorm:"type:varchar(20);default:'draft';index"`       // 状态
	Published       bool           `json:"published" gorm:"default:false"`                             // 是否已发布
	PublishedAt     *time.Time     `json:"published_at"`                                               // 发布时间
	Recipients      string         `json:"recipients" gorm:"type:text"`                                // 接收者(JSON)
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SecurityMetrics 安全指标模型
type SecurityMetrics struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	MetricName      string         `json:"metric_name" gorm:"type:varchar(100);not null;index"`         // 指标名称
	MetricValue     float64        `json:"metric_value" gorm:"type:decimal(10,2);not null"`             // 指标值
	MetricUnit      string         `json:"metric_unit" gorm:"type:varchar(20)"`                         // 指标单位
	Category        string         `json:"category" gorm:"type:varchar(50);index"`                      // 分类
	TimeWindow      string         `json:"time_window" gorm:"type:varchar(20)"`                         // 时间窗口
	RecordedAt      time.Time      `json:"recorded_at"`                                                 // 记录时间
	Threshold       float64        `json:"threshold" gorm:"type:decimal(10,2)"`                        // 阈值
	AlertTriggered  bool           `json:"alert_triggered" gorm:"default:false"`                       // 是否触发告警
	Tags            string         `json:"tags" gorm:"type:text"`                                       // 标签
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (SecurityEvent) TableName() string {
	return "security_events"
}

func (ThreatIntelligence) TableName() string {
	return "threat_intelligence"
}

func (AccessControl) TableName() string {
	return "access_controls"
}

func (PasswordHistory) TableName() string {
	return "password_history"
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func (AccountLockout) TableName() string {
	return "account_lockouts"
}

func (SecurityPolicy) TableName() string {
	return "security_policies"
}

func (SecurityAlert) TableName() string {
	return "security_alerts"
}

func (SecurityReport) TableName() string {
	return "security_reports"
}

func (SecurityMetrics) TableName() string {
	return "security_metrics"
}
