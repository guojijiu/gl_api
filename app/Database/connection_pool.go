package Database

import (
	"cloud-platform-api/app/Storage"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ConnectionPoolMonitor 连接池监控器
// 功能说明：
// 1. 监控数据库连接池状态
// 2. 检测连接泄漏和性能问题
// 3. 提供连接池统计信息
// 4. 支持告警和自动修复
type ConnectionPoolMonitor struct {
	db             *sql.DB
	storageManager *Storage.StorageManager
	config         *PoolMonitorConfig
	stopChan       chan bool
}

// PoolMonitorConfig 连接池监控配置
type PoolMonitorConfig struct {
	MonitorInterval    time.Duration `json:"monitor_interval"`    // 监控间隔
	MaxIdleConnections int           `json:"max_idle_connections"` // 最大空闲连接数
	MaxOpenConnections int           `json:"max_open_connections"` // 最大打开连接数
	ConnectionTimeout  time.Duration `json:"connection_timeout"`  // 连接超时时间
	AlertThreshold     float64       `json:"alert_threshold"`     // 告警阈值（连接使用率）
	EnableAutoFix      bool          `json:"enable_auto_fix"`     // 启用自动修复
}

// PoolStats 连接池统计信息
type PoolStats struct {
	OpenConnections    int           `json:"open_connections"`    // 当前打开连接数
	InUseConnections   int           `json:"in_use_connections"`  // 正在使用的连接数
	IdleConnections    int           `json:"idle_connections"`    // 空闲连接数
	WaitCount          int64         `json:"wait_count"`          // 等待连接次数
	WaitDuration       time.Duration `json:"wait_duration"`       // 等待连接总时间
	MaxIdleClosed      int64         `json:"max_idle_closed"`     // 因超时关闭的空闲连接数
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"` // 因生命周期关闭的连接数
	MaxOpenConnections int           `json:"max_open_connections"` // 最大打开连接数
	UsageRate          float64       `json:"usage_rate"`          // 连接使用率
	Timestamp          time.Time     `json:"timestamp"`           // 统计时间
}

// NewConnectionPoolMonitor 创建连接池监控器
// 功能说明：
// 1. 初始化连接池监控器
// 2. 设置监控配置参数
// 3. 启动监控协程
// 4. 提供统计信息收集
func NewConnectionPoolMonitor(db *sql.DB, storageManager *Storage.StorageManager) *ConnectionPoolMonitor {
	config := &PoolMonitorConfig{
		MonitorInterval:    5 * time.Minute,
		MaxIdleConnections: 10,
		MaxOpenConnections: 100,
		ConnectionTimeout:  30 * time.Second,
		AlertThreshold:     0.8, // 80%使用率告警
		EnableAutoFix:      false,
	}

	return &ConnectionPoolMonitor{
		db:             db,
		storageManager: storageManager,
		config:         config,
		stopChan:       make(chan bool),
	}
}

// StartMonitoring 开始监控
// 功能说明：
// 1. 启动定期监控协程
// 2. 收集连接池统计信息
// 3. 检测异常情况并告警
// 4. 支持自动修复功能
func (m *ConnectionPoolMonitor) StartMonitoring(interval time.Duration) {
	if interval > 0 {
		m.config.MonitorInterval = interval
	}

	go func() {
		ticker := time.NewTicker(m.config.MonitorInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.monitor()
			case <-m.stopChan:
				log.Println("连接池监控已停止")
				return
			}
		}
	}()

	log.Printf("连接池监控已启动，监控间隔: %v", m.config.MonitorInterval)
}

// StopMonitoring 停止监控
func (m *ConnectionPoolMonitor) StopMonitoring() {
	close(m.stopChan)
}

// monitor 执行监控检查
func (m *ConnectionPoolMonitor) monitor() {
	stats := m.getPoolStats()
	
	// 记录统计信息
	m.storageManager.LogInfo("数据库连接池统计", map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use_connections":  stats.InUseConnections,
		"idle_connections":    stats.IdleConnections,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
		"usage_rate":          fmt.Sprintf("%.2f%%", stats.UsageRate*100),
		"timestamp":           stats.Timestamp,
	})

	// 检查告警条件
	if stats.UsageRate > m.config.AlertThreshold {
		m.alertHighUsage(stats)
	}

	// 检查连接泄漏
	if stats.WaitCount > 0 {
		m.alertConnectionWait(stats)
	}

	// 自动修复（如果启用）
	if m.config.EnableAutoFix {
		m.autoFix(stats)
	}
}

// getPoolStats 获取连接池统计信息
func (m *ConnectionPoolMonitor) getPoolStats() *PoolStats {
	stats := &PoolStats{
		OpenConnections:    m.db.Stats().OpenConnections,
		InUseConnections:   m.db.Stats().InUse,
		IdleConnections:    m.db.Stats().Idle,
		WaitCount:          m.db.Stats().WaitCount,
		WaitDuration:       m.db.Stats().WaitDuration,
		MaxIdleClosed:      m.db.Stats().MaxIdleClosed,
		MaxLifetimeClosed:  m.db.Stats().MaxLifetimeClosed,
		MaxOpenConnections: m.db.Stats().MaxOpenConnections,
		Timestamp:          time.Now(),
	}

	// 计算使用率
	if stats.MaxOpenConnections > 0 {
		stats.UsageRate = float64(stats.InUseConnections) / float64(stats.MaxOpenConnections)
	}

	return stats
}

// alertHighUsage 高使用率告警
func (m *ConnectionPoolMonitor) alertHighUsage(stats *PoolStats) {
	message := fmt.Sprintf("数据库连接池使用率过高: %.2f%%", stats.UsageRate*100)
	
	m.storageManager.LogWarning("数据库连接池告警", map[string]interface{}{
		"alert_type":     "high_usage",
		"usage_rate":     stats.UsageRate,
		"threshold":      m.config.AlertThreshold,
		"open_connections": stats.OpenConnections,
		"in_use_connections": stats.InUseConnections,
		"message":        message,
	})

	log.Printf("⚠️  %s", message)
}

// alertConnectionWait 连接等待告警
func (m *ConnectionPoolMonitor) alertConnectionWait(stats *PoolStats) {
	message := fmt.Sprintf("数据库连接等待次数: %d, 总等待时间: %v", stats.WaitCount, stats.WaitDuration)
	
	m.storageManager.LogWarning("数据库连接等待告警", map[string]interface{}{
		"alert_type":    "connection_wait",
		"wait_count":    stats.WaitCount,
		"wait_duration": stats.WaitDuration.String(),
		"message":       message,
	})

	log.Printf("⚠️  %s", message)
}

// autoFix 自动修复
func (m *ConnectionPoolMonitor) autoFix(stats *PoolStats) {
	// 如果使用率过高，尝试增加最大连接数
	if stats.UsageRate > 0.9 && stats.MaxOpenConnections < 200 {
		newMax := stats.MaxOpenConnections + 20
		m.db.SetMaxOpenConns(newMax)
		
		m.storageManager.LogInfo("自动修复：增加最大连接数", map[string]interface{}{
			"old_max": stats.MaxOpenConnections,
			"new_max": newMax,
			"reason":  "high_usage_rate",
		})
		
		log.Printf("🔧 自动修复：最大连接数从 %d 增加到 %d", stats.MaxOpenConnections, newMax)
	}

	// 如果空闲连接过多，减少最大空闲连接数
	if stats.IdleConnections > stats.MaxOpenConnections/2 {
		newMaxIdle := stats.MaxOpenConnections / 4
		m.db.SetMaxIdleConns(newMaxIdle)
		
		m.storageManager.LogInfo("自动修复：减少最大空闲连接数", map[string]interface{}{
			"old_max_idle": stats.IdleConnections,
			"new_max_idle": newMaxIdle,
			"reason":       "too_many_idle_connections",
		})
		
		log.Printf("🔧 自动修复：最大空闲连接数从 %d 减少到 %d", stats.IdleConnections, newMaxIdle)
	}
}

// GetPoolStats 获取当前连接池统计信息
func (m *ConnectionPoolMonitor) GetPoolStats() *PoolStats {
	return m.getPoolStats()
}

// SetConfig 更新监控配置
func (m *ConnectionPoolMonitor) SetConfig(config *PoolMonitorConfig) {
	m.config = config
}

// HealthCheck 健康检查
func (m *ConnectionPoolMonitor) HealthCheck() map[string]interface{} {
	stats := m.getPoolStats()
	
	health := map[string]interface{}{
		"status":           "healthy",
		"usage_rate":       stats.UsageRate,
		"open_connections": stats.OpenConnections,
		"in_use_connections": stats.InUseConnections,
		"idle_connections": stats.IdleConnections,
		"wait_count":       stats.WaitCount,
		"last_check":       stats.Timestamp,
	}

	// 判断健康状态
	if stats.UsageRate > 0.9 {
		health["status"] = "warning"
		health["message"] = "连接池使用率过高"
	} else if stats.WaitCount > 10 {
		health["status"] = "warning"
		health["message"] = "连接等待次数过多"
	} else if stats.UsageRate > 0.95 {
		health["status"] = "critical"
		health["message"] = "连接池即将耗尽"
	}

	return health
}
