package Models

import (
	"time"
)

// MonitoringMetric 监控指标
type MonitoringMetric struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:50;not null;index" json:"type"`           // 指标类型：system, application, database, cache, business
	Name        string    `gorm:"size:100;not null;index" json:"name"`          // 指标名称
	Value       float64   `gorm:"not null" json:"value"`                         // 指标值
	Unit        string    `gorm:"size:20" json:"unit"`                           // 单位
	Threshold   float64   `gorm:"not null" json:"threshold"`                    // 阈值
	Status      string    `gorm:"size:20;not null;index" json:"status"`         // 状态：normal, warning, critical
	Severity    string    `gorm:"size:20;not null" json:"severity"`              // 严重程度：info, warning, critical, emergency
	Description string    `gorm:"size:500" json:"description"`                  // 描述
	Tags        string    `gorm:"size:1000" json:"tags"`                        // 标签（JSON格式）
	Metadata    string    `gorm:"type:text" json:"metadata"`                    // 元数据（JSON格式）
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`              // 时间戳
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`   // 规则名称
	Description string    `gorm:"size:500" json:"description"`                  // 规则描述
	Type        string    `gorm:"size:50;not null;index" json:"type"`           // 规则类型：threshold, trend, anomaly
	MetricType  string    `gorm:"size:50;not null;index" json:"metric_type"`   // 监控指标类型
	MetricName  string    `gorm:"size:100;not null" json:"metric_name"`         // 监控指标名称
	Condition   string    `gorm:"size:20;not null" json:"condition"`           // 条件：>, <, >=, <=, ==, !=
	Threshold   float64   `gorm:"not null" json:"threshold"`                   // 阈值
	Duration    int       `gorm:"not null;default:1" json:"duration"`          // 持续时间（检查次数）
	Severity    string    `gorm:"size:20;not null" json:"severity"`             // 严重程度
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`        // 是否启用
	Suppression bool      `gorm:"not null;default:false" json:"suppression"`   // 是否抑制
	SuppressionWindow int `gorm:"not null;default:3600" json:"suppression_window"` // 抑制窗口（秒）
	Escalation  bool      `gorm:"not null;default:false" json:"escalation"`    // 是否升级
	EscalationDelay int   `gorm:"not null;default:600" json:"escalation_delay"` // 升级延迟（秒）
	MaxEscalationLevel int `gorm:"not null;default:3" json:"max_escalation_level"` // 最大升级级别
	NotificationChannels string `gorm:"size:500" json:"notification_channels"` // 通知渠道（JSON格式）
	Tags        string    `gorm:"size:1000" json:"tags"`                         // 标签（JSON格式）
	CreatedBy   uint      `gorm:"not null" json:"created_by"`                    // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Alert 告警记录
type Alert struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RuleID      uint      `gorm:"not null;index" json:"rule_id"`                // 告警规则ID
	RuleName    string    `gorm:"size:100;not null" json:"rule_name"`           // 告警规则名称
	Type        string    `gorm:"size:50;not null;index" json:"type"`           // 告警类型
	MetricType  string    `gorm:"size:50;not null;index" json:"metric_type"`   // 监控指标类型
	MetricName  string    `gorm:"size:100;not null" json:"metric_name"`        // 监控指标名称
	Value       float64   `gorm:"not null" json:"value"`                        // 触发值
	Threshold   float64   `gorm:"not null" json:"threshold"`                   // 阈值
	Severity    string    `gorm:"size:20;not null;index" json:"severity"`       // 严重程度
	Status      string    `gorm:"size:20;not null;index" json:"status"`        // 状态：active, acknowledged, resolved, suppressed
	Message     string    `gorm:"size:1000;not null" json:"message"`            // 告警消息
	Description string    `gorm:"size:1000" json:"description"`                 // 详细描述
	Tags        string    `gorm:"size:1000" json:"tags"`                        // 标签（JSON格式）
	Metadata    string    `gorm:"type:text" json:"metadata"`                    // 元数据（JSON格式）
	FiredAt     time.Time `gorm:"not null;index" json:"fired_at"`              // 触发时间
	AcknowledgedAt *time.Time `json:"acknowledged_at"`                         // 确认时间
	AcknowledgedBy *uint     `json:"acknowledged_by"`                          // 确认者ID
	ResolvedAt  *time.Time `json:"resolved_at"`                                // 解决时间
	ResolvedBy  *uint      `json:"resolved_by"`                                 // 解决者ID
	EscalationLevel int    `gorm:"not null;default:0" json:"escalation_level"` // 升级级别
	Suppressed  bool      `gorm:"not null;default:false" json:"suppressed"`    // 是否被抑制
	SuppressedAt *time.Time `json:"suppressed_at"`                             // 抑制时间
	SuppressedBy *uint     `json:"suppressed_by"`                              // 抑制者ID
	SuppressionReason string `gorm:"size:500" json:"suppression_reason"`        // 抑制原因
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NotificationRecord 通知记录
type NotificationRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AlertID     uint      `gorm:"not null;index" json:"alert_id"`              // 告警ID
	Channel     string    `gorm:"size:50;not null;index" json:"channel"`       // 通知渠道：email, webhook, slack, dingtalk, sms
	Recipient   string    `gorm:"size:200;not null" json:"recipient"`          // 接收者
	Subject     string    `gorm:"size:200" json:"subject"`                     // 主题
	Content     string    `gorm:"type:text;not null" json:"content"`           // 内容
	Status      string    `gorm:"size:20;not null;index" json:"status"`        // 状态：pending, sent, failed, retrying
	RetryCount  int       `gorm:"not null;default:0" json:"retry_count"`       // 重试次数
	MaxRetries  int       `gorm:"not null;default:3" json:"max_retries"`      // 最大重试次数
	Error       string    `gorm:"size:500" json:"error"`                       // 错误信息
	SentAt      *time.Time `json:"sent_at"`                                    // 发送时间
	Response    string    `gorm:"size:1000" json:"response"`                   // 响应信息
	Metadata    string    `gorm:"type:text" json:"metadata"`                   // 元数据（JSON格式）
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonitoringDashboard 监控仪表板
type MonitoringDashboard struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`   // 仪表板名称
	Description string    `gorm:"size:500" json:"description"`                 // 描述
	Layout      string    `gorm:"type:text;not null" json:"layout"`          // 布局配置（JSON格式）
	Widgets     string    `gorm:"type:text;not null" json:"widgets"`         // 组件配置（JSON格式）
	RefreshInterval int   `gorm:"not null;default:30" json:"refresh_interval"` // 刷新间隔（秒）
	IsDefault   bool      `gorm:"not null;default:false" json:"is_default"`   // 是否默认仪表板
	IsPublic   bool      `gorm:"not null;default:false" json:"is_public"`     // 是否公开
	CreatedBy   uint      `gorm:"not null" json:"created_by"`                 // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonitoringWidget 监控组件
type MonitoringWidget struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DashboardID uint      `gorm:"not null;index" json:"dashboard_id"`         // 仪表板ID
	Name        string    `gorm:"size:100;not null" json:"name"`              // 组件名称
	Type        string    `gorm:"size:50;not null" json:"type"`               // 组件类型：chart, gauge, table, text, alert
	Config      string    `gorm:"type:text;not null" json:"config"`         // 配置（JSON格式）
	Position    string    `gorm:"size:100;not null" json:"position"`        // 位置（JSON格式）
	Size        string    `gorm:"size:100;not null" json:"size"`              // 大小（JSON格式）
	DataSource  string    `gorm:"size:200" json:"data_source"`               // 数据源
	Query       string    `gorm:"type:text" json:"query"`                     // 查询条件
	RefreshInterval int   `gorm:"not null;default:30" json:"refresh_interval"` // 刷新间隔（秒）
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`       // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonitoringSchedule 监控调度
type MonitoringSchedule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`  // 调度名称
	Description string    `gorm:"size:500" json:"description"`                 // 描述
	Type        string    `gorm:"size:50;not null" json:"type"`                // 调度类型：interval, cron, event
	Expression  string    `gorm:"size:200;not null" json:"expression"`        // 表达式
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`        // 是否启用
	LastRun     *time.Time `json:"last_run"`                                  // 上次运行时间
	NextRun     *time.Time `json:"next_run"`                                  // 下次运行时间
	RunCount    int       `gorm:"not null;default:0" json:"run_count"`       // 运行次数
	SuccessCount int      `gorm:"not null;default:0" json:"success_count"`    // 成功次数
	FailureCount int      `gorm:"not null;default:0" json:"failure_count"`    // 失败次数
	CreatedBy   uint      `gorm:"not null" json:"created_by"`                 // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonitoringReport 监控报告
type MonitoringReport struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`              // 报告名称
	Description string    `gorm:"size:500" json:"description"`                // 描述
	Type        string    `gorm:"size:50;not null" json:"type"`               // 报告类型：daily, weekly, monthly, custom
	ScheduleID  *uint     `gorm:"index" json:"schedule_id"`                   // 调度ID
	Template    string    `gorm:"type:text;not null" json:"template"`        // 模板配置（JSON格式）
	Parameters  string    `gorm:"type:text" json:"parameters"`                // 参数（JSON格式）
	Format      string    `gorm:"size:20;not null;default:'pdf'" json:"format"` // 格式：pdf, html, json, csv
	Recipients  string    `gorm:"size:1000" json:"recipients"`                // 接收者（JSON格式）
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`       // 是否启用
	LastGenerated *time.Time `json:"last_generated"`                         // 上次生成时间
	CreatedBy   uint      `gorm:"not null" json:"created_by"`                 // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonitoringEvent 监控事件
type MonitoringEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:50;not null;index" json:"type"`         // 事件类型
	Category    string    `gorm:"size:50;not null;index" json:"category"`     // 事件分类
	Source      string    `gorm:"size:100;not null" json:"source"`           // 事件源
	Severity    string    `gorm:"size:20;not null;index" json:"severity"`      // 严重程度
	Message     string    `gorm:"size:1000;not null" json:"message"`         // 事件消息
	Description string    `gorm:"size:1000" json:"description"`               // 详细描述
	Data        string    `gorm:"type:text" json:"data"`                      // 事件数据（JSON格式）
	Tags        string    `gorm:"size:1000" json:"tags"`                       // 标签（JSON格式）
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`          // 时间戳
	Processed   bool      `gorm:"not null;default:false" json:"processed"`    // 是否已处理
	ProcessedAt *time.Time `json:"processed_at"`                             // 处理时间
	ProcessedBy *uint      `json:"processed_by"`                              // 处理者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (MonitoringMetric) TableName() string {
	return "monitoring_metrics"
}

func (AlertRule) TableName() string {
	return "monitoring_alert_rules"
}

func (Alert) TableName() string {
	return "monitoring_alerts"
}

func (NotificationRecord) TableName() string {
	return "monitoring_notification_records"
}

func (MonitoringDashboard) TableName() string {
	return "monitoring_dashboards"
}

func (MonitoringWidget) TableName() string {
	return "monitoring_widgets"
}

func (MonitoringSchedule) TableName() string {
	return "monitoring_schedules"
}

func (MonitoringReport) TableName() string {
	return "monitoring_reports"
}

func (MonitoringEvent) TableName() string {
	return "monitoring_events"
}
