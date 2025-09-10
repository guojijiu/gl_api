package Models

import (
	"time"
)

// PerformanceMetric 性能指标基础模型
type PerformanceMetric struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;index" json:"name"`          // 指标名称
	Value       float64   `gorm:"not null" json:"value"`                        // 指标值
	Unit        string    `gorm:"size:20" json:"unit"`                           // 单位
	Category    string    `gorm:"size:50;not null;index" json:"category"`       // 指标分类
	Status      string    `gorm:"size:20;not null;index" json:"status"`         // 状态：normal, warning, critical
	Severity    string    `gorm:"size:20;not null" json:"severity"`              // 严重程度：info, warning, critical
	Description string    `gorm:"size:500" json:"description"`                   // 描述
	Tags        string    `gorm:"size:1000" json:"tags"`                         // 标签（JSON格式）
	Metadata    string    `gorm:"type:text" json:"metadata"`                     // 元数据（JSON格式）
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`               // 时间戳
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemResourceMetric 系统资源指标
type SystemResourceMetric struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CPUUsage        float64   `gorm:"not null" json:"cpu_usage"`                 // CPU使用率
	MemoryUsage     float64   `gorm:"not null" json:"memory_usage"`              // 内存使用率
	DiskUsage       float64   `gorm:"not null" json:"disk_usage"`                // 磁盘使用率
	NetworkIn       float64   `gorm:"not null" json:"network_in"`                // 网络入流量
	NetworkOut      float64   `gorm:"not null" json:"network_out"`               // 网络出流量
	LoadAverage     float64   `gorm:"not null" json:"load_average"`              // 系统负载
	ProcessCount    int       `gorm:"not null" json:"process_count"`             // 进程数量
	ThreadCount     int       `gorm:"not null" json:"thread_count"`              // 线程数量
	OpenFiles       int       `gorm:"not null" json:"open_files"`                // 打开文件数
	Status          string    `gorm:"size:20;not null;index" json:"status"`      // 状态
	Severity        string    `gorm:"size:20;not null" json:"severity"`           // 严重程度
	Description     string    `gorm:"size:500" json:"description"`                // 描述
	Tags            string    `gorm:"size:1000" json:"tags"`                      // 标签
	Metadata        string    `gorm:"type:text" json:"metadata"`                  // 元数据
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`            // 时间戳
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ApplicationMetric 应用程序指标
type ApplicationMetric struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ResponseTime    float64   `gorm:"not null" json:"response_time"`             // 响应时间
	Throughput      float64   `gorm:"not null" json:"throughput"`                // 吞吐量
	ErrorRate       float64   `gorm:"not null" json:"error_rate"`                // 错误率
	SuccessRate     float64   `gorm:"not null" json:"success_rate"`              // 成功率
	ActiveUsers     int       `gorm:"not null" json:"active_users"`               // 活跃用户数
	ConcurrentUsers int       `gorm:"not null" json:"concurrent_users"`          // 并发用户数
	RequestCount    int       `gorm:"not null" json:"request_count"`             // 请求数量
	Endpoint        string    `gorm:"size:200;not null;index" json:"endpoint"`   // API端点
	Method          string    `gorm:"size:10;not null" json:"method"`            // HTTP方法
	Status          string    `gorm:"size:20;not null;index" json:"status"`      // 状态
	Severity        string    `gorm:"size:20;not null" json:"severity"`           // 严重程度
	Description     string    `gorm:"size:500" json:"description"`                // 描述
	Tags            string    `gorm:"size:1000" json:"tags"`                      // 标签
	Metadata        string    `gorm:"type:text" json:"metadata"`                  // 元数据
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`            // 时间戳
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BusinessMetric 业务指标
type BusinessMetric struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MetricName      string    `gorm:"size:100;not null;index" json:"metric_name"` // 指标名称
	Value           float64   `gorm:"not null" json:"value"`                       // 指标值
	Unit            string    `gorm:"size:20" json:"unit"`                          // 单位
	Category        string    `gorm:"size:50;not null;index" json:"category"`      // 业务分类
	BusinessUnit    string    `gorm:"size:100;not null;index" json:"business_unit"` // 业务单元
	Target          float64   `gorm:"not null" json:"target"`                       // 目标值
	Threshold       float64   `gorm:"not null" json:"threshold"`                    // 阈值
	Trend           string    `gorm:"size:20;not null" json:"trend"`                // 趋势：up, down, stable
	Status          string    `gorm:"size:20;not null;index" json:"status"`         // 状态
	Severity        string    `gorm:"size:20;not null" json:"severity"`              // 严重程度
	Description     string    `gorm:"size:500" json:"description"`                   // 描述
	Tags            string    `gorm:"size:1000" json:"tags"`                         // 标签
	Metadata        string    `gorm:"type:text" json:"metadata"`                     // 元数据
	Timestamp       time.Time `gorm:"not null;index" json:"timestamp"`               // 时间戳
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MetricID    uint      `gorm:"not null;index" json:"metric_id"`                // 指标ID
	MetricType  string    `gorm:"size:50;not null;index" json:"metric_type"`      // 指标类型
	MetricName  string    `gorm:"size:100;not null" json:"metric_name"`           // 指标名称
	Value       float64   `gorm:"not null" json:"value"`                           // 触发值
	Threshold   float64   `gorm:"not null" json:"threshold"`                       // 阈值
	Condition   string    `gorm:"size:20;not null" json:"condition"`               // 触发条件
	Severity    string    `gorm:"size:20;not null;index" json:"severity"`          // 严重程度
	Status      string    `gorm:"size:20;not null;index" json:"status"`            // 状态：active, acknowledged, resolved
	Message     string    `gorm:"size:1000;not null" json:"message"`               // 告警消息
	Description string    `gorm:"size:1000" json:"description"`                    // 详细描述
	Tags        string    `gorm:"size:1000" json:"tags"`                           // 标签
	Metadata    string    `gorm:"type:text" json:"metadata"`                       // 元数据
	FiredAt     time.Time `gorm:"not null;index" json:"fired_at"`                  // 触发时间
	AcknowledgedAt *time.Time `json:"acknowledged_at"`                             // 确认时间
	AcknowledgedBy *uint     `json:"acknowledged_by"`                              // 确认者ID
	ResolvedAt  *time.Time `json:"resolved_at"`                                    // 解决时间
	ResolvedBy  *uint      `json:"resolved_by"`                                     // 解决者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ReportName      string    `gorm:"size:100;not null" json:"report_name"`        // 报告名称
	ReportType      string    `gorm:"size:50;not null;index" json:"report_type"`   // 报告类型
	Period          string    `gorm:"size:20;not null" json:"period"`               // 报告周期：hourly, daily, weekly, monthly
	StartTime       time.Time `gorm:"not null" json:"start_time"`                   // 开始时间
	EndTime         time.Time `gorm:"not null" json:"end_time"`                     // 结束时间
	Summary         string    `gorm:"type:text" json:"summary"`                     // 报告摘要
	Details         string    `gorm:"type:text" json:"details"`                     // 详细内容
	Status          string    `gorm:"size:20;not null;index" json:"status"`         // 状态：generating, completed, failed
	GeneratedBy     uint      `gorm:"not null" json:"generated_by"`                 // 生成者ID
	Tags            string    `gorm:"size:1000" json:"tags"`                         // 标签
	Metadata        string    `gorm:"type:text" json:"metadata"`                     // 元数据
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
