package Services

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// CircuitBreakerInterface 熔断器接口
type CircuitBreakerInterface interface {
	GetState() string
	GetStats() map[string]interface{}
	RecordResult(success bool, duration time.Duration)
	AllowRequest() bool
}

// MonitoringIntegrationService 监控集成服务
type MonitoringIntegrationService struct {
	BaseService
	storageManager       *Storage.StorageManager
	circuitBreakers      map[string]CircuitBreakerInterface
	performanceMetrics   map[string]interface{}
	mutex                sync.RWMutex
	alertThresholds      map[string]float64
	notificationChannels []NotificationChannel
}

// NotificationChannel 通知通道接口
type NotificationChannel interface {
	SendAlert(alert MonitoringAlert) error
	GetName() string
	IsEnabled() bool
}

// MonitoringAlert 监控告警信息
type MonitoringAlert struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
}

// CircuitBreakerAlert 熔断器告警
type CircuitBreakerAlert struct {
	BreakerName string  `json:"breaker_name"`
	State       string  `json:"state"`
	Failures    uint32  `json:"failures"`
	Requests    uint32  `json:"requests"`
	FailureRate float64 `json:"failure_rate"`
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Threshold  float64 `json:"threshold"`
	Unit       string  `json:"unit"`
	Duration   string  `json:"duration"`
}

// NewMonitoringIntegrationService 创建监控集成服务
func NewMonitoringIntegrationService(storageManager *Storage.StorageManager) *MonitoringIntegrationService {
	return &MonitoringIntegrationService{
		storageManager:     storageManager,
		circuitBreakers:    make(map[string]CircuitBreakerInterface),
		performanceMetrics: make(map[string]interface{}),
		alertThresholds: map[string]float64{
			"response_time":   1000, // 1秒
			"error_rate":      0.05, // 5%
			"cpu_usage":       80.0, // 80%
			"memory_usage":    80.0, // 80%
			"circuit_breaker": 0.5,  // 50%失败率
		},
		notificationChannels: make([]NotificationChannel, 0),
	}
}

// RegisterCircuitBreaker 注册熔断器
func (mis *MonitoringIntegrationService) RegisterCircuitBreaker(name string, breaker CircuitBreakerInterface) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	mis.circuitBreakers[name] = breaker
	mis.storageManager.LogInfo("熔断器已注册", map[string]interface{}{
		"name": name,
		"time": time.Now(),
	})
}

// UnregisterCircuitBreaker 注销熔断器
func (mis *MonitoringIntegrationService) UnregisterCircuitBreaker(name string) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	delete(mis.circuitBreakers, name)
	mis.storageManager.LogInfo("熔断器已注销", map[string]interface{}{
		"name": name,
		"time": time.Now(),
	})
}

// UpdatePerformanceMetrics 更新性能指标
func (mis *MonitoringIntegrationService) UpdatePerformanceMetrics(metrics map[string]interface{}) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	for key, value := range metrics {
		mis.performanceMetrics[key] = value
	}

	// 检查性能告警
	mis.checkPerformanceAlerts(metrics)
}

// checkPerformanceAlerts 检查性能告警
func (mis *MonitoringIntegrationService) checkPerformanceAlerts(metrics map[string]interface{}) {
	for metricName, value := range metrics {
		if threshold, exists := mis.alertThresholds[metricName]; exists {
			if floatValue, ok := value.(float64); ok {
				if floatValue > threshold {
					alert := MonitoringAlert{
						ID:        fmt.Sprintf("perf_%s_%d", metricName, time.Now().Unix()),
						Type:      "performance",
						Severity:  mis.getSeverity(floatValue, threshold),
						Title:     fmt.Sprintf("性能指标告警: %s", metricName),
						Message:   fmt.Sprintf("%s 当前值 %.2f 超过阈值 %.2f", metricName, floatValue, threshold),
						Source:    "performance_monitor",
						Timestamp: time.Now(),
						Metadata: map[string]interface{}{
							"metric_name": metricName,
							"value":       floatValue,
							"threshold":   threshold,
						},
					}

					mis.sendAlert(alert)
				}
			}
		}
	}
}

// getSeverity 获取告警严重程度
func (mis *MonitoringIntegrationService) getSeverity(value, threshold float64) string {
	ratio := value / threshold
	if ratio >= 2.0 {
		return "critical"
	} else if ratio >= 1.5 {
		return "high"
	} else if ratio >= 1.2 {
		return "medium"
	} else {
		return "low"
	}
}

// MonitorCircuitBreakers 监控熔断器状态
func (mis *MonitoringIntegrationService) MonitorCircuitBreakers() {
	mis.mutex.RLock()
	breakers := make(map[string]CircuitBreakerInterface)
	for name, breaker := range mis.circuitBreakers {
		breakers[name] = breaker
	}
	mis.mutex.RUnlock()

	for name, breaker := range breakers {
		stats := breaker.GetStats()
		state := stats["state"].(string)

		// 检查熔断器状态变化
		if state == "open" {
			alert := MonitoringAlert{
				ID:        fmt.Sprintf("cb_%s_%d", name, time.Now().Unix()),
				Type:      "circuit_breaker",
				Severity:  "high",
				Title:     fmt.Sprintf("熔断器开启: %s", name),
				Message:   fmt.Sprintf("熔断器 %s 已开启，失败率过高", name),
				Source:    "circuit_breaker_monitor",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"breaker_name": name,
					"state":        state,
					"failures":     stats["failures"],
					"requests":     stats["requests"],
				},
			}

			mis.sendAlert(alert)
		}
	}
}

// AddNotificationChannel 添加通知通道
func (mis *MonitoringIntegrationService) AddNotificationChannel(channel NotificationChannel) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	mis.notificationChannels = append(mis.notificationChannels, channel)
	mis.storageManager.LogInfo("通知通道已添加", map[string]interface{}{
		"channel": channel.GetName(),
		"time":    time.Now(),
	})
}

// RemoveNotificationChannel 移除通知通道
func (mis *MonitoringIntegrationService) RemoveNotificationChannel(channelName string) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	for i, channel := range mis.notificationChannels {
		if channel.GetName() == channelName {
			mis.notificationChannels = append(mis.notificationChannels[:i], mis.notificationChannels[i+1:]...)
			break
		}
	}
}

// sendAlert 发送告警
func (mis *MonitoringIntegrationService) sendAlert(alert MonitoringAlert) {
	mis.mutex.RLock()
	channels := make([]NotificationChannel, len(mis.notificationChannels))
	copy(channels, mis.notificationChannels)
	mis.mutex.RUnlock()

	// 记录告警到日志
	mis.storageManager.LogWarning("系统告警", map[string]interface{}{
		"alert_id":  alert.ID,
		"type":      alert.Type,
		"severity":  alert.Severity,
		"title":     alert.Title,
		"message":   alert.Message,
		"source":    alert.Source,
		"timestamp": alert.Timestamp,
		"metadata":  alert.Metadata,
	})

	// 发送到所有启用的通知通道
	for _, channel := range channels {
		if channel.IsEnabled() {
			go func(ch NotificationChannel) {
				if err := ch.SendAlert(alert); err != nil {
					mis.storageManager.LogError("发送告警失败", map[string]interface{}{
						"channel":  ch.GetName(),
						"alert_id": alert.ID,
						"error":    err.Error(),
					})
				}
			}(channel)
		}
	}
}

// GetMonitoringStatus 获取监控状态
func (mis *MonitoringIntegrationService) GetMonitoringStatus() map[string]interface{} {
	mis.mutex.RLock()
	defer mis.mutex.RUnlock()

	status := map[string]interface{}{
		"circuit_breakers":      make(map[string]interface{}),
		"performance_metrics":   mis.performanceMetrics,
		"alert_thresholds":      mis.alertThresholds,
		"notification_channels": len(mis.notificationChannels),
		"timestamp":             time.Now(),
	}

	// 收集熔断器状态
	for name, breaker := range mis.circuitBreakers {
		status["circuit_breakers"].(map[string]interface{})[name] = breaker.GetStats()
	}

	return status
}

// SetAlertThreshold 设置告警阈值
func (mis *MonitoringIntegrationService) SetAlertThreshold(metricName string, threshold float64) {
	mis.mutex.Lock()
	defer mis.mutex.Unlock()

	mis.alertThresholds[metricName] = threshold
	mis.storageManager.LogInfo("告警阈值已更新", map[string]interface{}{
		"metric_name": metricName,
		"threshold":   threshold,
		"time":        time.Now(),
	})
}

// GetAlertThresholds 获取告警阈值
func (mis *MonitoringIntegrationService) GetAlertThresholds() map[string]float64 {
	mis.mutex.RLock()
	defer mis.mutex.RUnlock()

	thresholds := make(map[string]float64)
	for key, value := range mis.alertThresholds {
		thresholds[key] = value
	}

	return thresholds
}

// StartMonitoring 启动监控
func (mis *MonitoringIntegrationService) StartMonitoring() {
	go mis.monitoringLoop()
	mis.storageManager.LogInfo("监控服务已启动", map[string]interface{}{
		"time": time.Now(),
	})
}

// monitoringLoop 监控循环
func (mis *MonitoringIntegrationService) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 监控熔断器
		mis.MonitorCircuitBreakers()

		// 检查系统资源
		mis.checkSystemResources()
	}
}

// checkSystemResources 检查系统资源
func (mis *MonitoringIntegrationService) checkSystemResources() {
	// 这里可以添加系统资源检查逻辑
	// 例如：CPU使用率、内存使用率、磁盘空间等

	// 示例：检查内存使用率
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	if memoryUsage > mis.alertThresholds["memory_usage"] {
		alert := MonitoringAlert{
			ID:        fmt.Sprintf("mem_%d", time.Now().Unix()),
			Type:      "system_resource",
			Severity:  mis.getSeverity(memoryUsage, mis.alertThresholds["memory_usage"]),
			Title:     "内存使用率告警",
			Message:   fmt.Sprintf("内存使用率 %.2f%% 超过阈值 %.2f%%", memoryUsage, mis.alertThresholds["memory_usage"]),
			Source:    "system_monitor",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"memory_usage": memoryUsage,
				"threshold":    mis.alertThresholds["memory_usage"],
			},
		}

		mis.sendAlert(alert)
	}
}

// GetAlertHistory 获取告警历史
func (mis *MonitoringIntegrationService) GetAlertHistory(limit int) ([]MonitoringAlert, error) {
	// 这里可以从数据库或日志中获取告警历史
	// 暂时返回空列表
	return []MonitoringAlert{}, nil
}

// ResolveAlert 解决告警
func (mis *MonitoringIntegrationService) ResolveAlert(alertID string) error {
	// 这里可以实现告警解决逻辑
	// 暂时只记录日志
	mis.storageManager.LogInfo("告警已解决", map[string]interface{}{
		"alert_id": alertID,
		"time":     time.Now(),
	})

	return nil
}

// 全局监控集成服务
var globalMonitoringIntegrationService *MonitoringIntegrationService

// InitMonitoringIntegration 初始化监控集成
func InitMonitoringIntegration(storageManager *Storage.StorageManager) {
	globalMonitoringIntegrationService = NewMonitoringIntegrationService(storageManager)
	globalMonitoringIntegrationService.StartMonitoring()
}

// GetMonitoringIntegrationService 获取全局监控集成服务
func GetMonitoringIntegrationService() *MonitoringIntegrationService {
	return globalMonitoringIntegrationService
}
