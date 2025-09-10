package Services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"runtime"
	"strings"
	"sync"
	"time"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"gorm.io/gorm"
)

// MonitoringService 监控告警服务
type MonitoringService struct {
	db           *gorm.DB
	config       *Config.MonitoringConfig
	alertRules   map[uint]*Models.AlertRule
	alerts       map[uint]*Models.Alert
	notifications map[uint]*Models.NotificationRecord
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	stopChan     chan struct{}
	metricsChan  chan *Models.MonitoringMetric
	alertChan    chan *Models.Alert
	notificationChan chan *Models.NotificationRecord
}

// NewMonitoringService 创建监控告警服务
func NewMonitoringService(db *gorm.DB, config *Config.MonitoringConfig) *MonitoringService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &MonitoringService{
		db:           db,
		config:       config,
		alertRules:   make(map[uint]*Models.AlertRule),
		alerts:       make(map[uint]*Models.Alert),
		notifications: make(map[uint]*Models.NotificationRecord),
		ctx:          ctx,
		cancel:       cancel,
		stopChan:     make(chan struct{}),
		metricsChan:  make(chan *Models.MonitoringMetric, 1000),
		alertChan:    make(chan *Models.Alert, 100),
		notificationChan: make(chan *Models.NotificationRecord, 100),
	}

	// 初始化数据库表
	service.initDatabase()
	
	// 加载告警规则
	service.loadAlertRules()
	
	// 启动监控服务
	go service.startMonitoring()
	
	// 启动告警处理
	go service.startAlertProcessor()
	
	// 启动通知处理
	go service.startNotificationProcessor()
	
	// 启动数据清理
	go service.startDataCleanup()

	return service
}

// initDatabase 初始化数据库表
func (s *MonitoringService) initDatabase() {
	// 自动迁移表结构
	s.db.AutoMigrate(
		&Models.MonitoringMetric{},
		&Models.AlertRule{},
		&Models.Alert{},
		&Models.NotificationRecord{},
		&Models.MonitoringDashboard{},
		&Models.MonitoringWidget{},
		&Models.MonitoringSchedule{},
		&Models.MonitoringReport{},
		&Models.MonitoringEvent{},
	)
}

// GetDB 获取数据库连接
func (s *MonitoringService) GetDB() *gorm.DB {
	return s.db
}

// loadAlertRules 加载告警规则
func (s *MonitoringService) loadAlertRules() {
	var rules []Models.AlertRule
	if err := s.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		log.Printf("加载告警规则失败: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.alertRules = make(map[uint]*Models.AlertRule)
	for i := range rules {
		s.alertRules[rules[i].ID] = &rules[i]
	}
	
	log.Printf("加载了 %d 个告警规则", len(rules))
}

// startMonitoring 启动监控
func (s *MonitoringService) startMonitoring() {
	// 检查检查间隔，如果为0则使用默认值
	checkInterval := s.config.BaseConfig.CheckInterval
	if checkInterval <= 0 {
		checkInterval = 5 * time.Minute // 默认5分钟
	}
	
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics 收集监控指标
func (s *MonitoringService) collectMetrics() {
	// 系统监控
	if s.config.SystemMonitoring.Enabled {
		s.collectSystemMetrics()
	}

	// 应用监控
	if s.config.ApplicationMonitoring.Enabled {
		s.collectApplicationMetrics()
	}

	// 数据库监控
	if s.config.DatabaseMonitoring.Enabled {
		s.collectDatabaseMetrics()
	}

	// 缓存监控
	if s.config.CacheMonitoring.Enabled {
		s.collectCacheMetrics()
	}

	// 业务监控
	if s.config.BusinessMonitoring.Enabled {
		s.collectBusinessMetrics()
	}
}

// collectSystemMetrics 收集系统指标
func (s *MonitoringService) collectSystemMetrics() {
	// CPU使用率
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		s.recordMetric("system", "cpu_usage", cpuPercent[0], s.config.SystemMonitoring.CPUThreshold, "%")
	}

	// 内存使用率
	if vmstat, err := mem.VirtualMemory(); err == nil {
		s.recordMetric("system", "memory_usage", vmstat.UsedPercent, s.config.SystemMonitoring.MemoryThreshold, "%")
	}

	// 磁盘使用率
	if partitions, err := disk.Partitions(false); err == nil {
		for _, partition := range partitions {
			if usage, err := disk.Usage(partition.Mountpoint); err == nil {
				s.recordMetric("system", "disk_usage_"+partition.Device, usage.UsedPercent, s.config.SystemMonitoring.DiskThreshold, "%")
			}
		}
	}

	// 网络流量
	if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
		s.recordMetric("system", "network_bytes_sent", float64(netIO[0].BytesSent), s.config.SystemMonitoring.NetworkThreshold, "bytes")
		s.recordMetric("system", "network_bytes_recv", float64(netIO[0].BytesRecv), s.config.SystemMonitoring.NetworkThreshold, "bytes")
	}

	// 进程数量
	if processes, err := process.Processes(); err == nil {
		s.recordMetric("system", "process_count", float64(len(processes)), float64(s.config.SystemMonitoring.ProcessThreshold), "count")
	}

	// 系统负载（暂时注释掉，因为gopsutil版本兼容性问题）
	// if loadAvg, err := cpu.LoadAvg(); err == nil {
	// 	s.recordMetric("system", "load_average_1m", loadAvg.Load1, s.config.SystemMonitoring.LoadAverageThreshold, "load")
	// 	s.recordMetric("system", "load_average_5m", loadAvg.Load5, s.config.SystemMonitoring.LoadAverageThreshold, "load")
	// 	s.recordMetric("system", "load_average_15m", loadAvg.Load15, s.config.SystemMonitoring.LoadAverageThreshold, "load")
	// }
}

// collectApplicationMetrics 收集应用指标
func (s *MonitoringService) collectApplicationMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 内存使用
	s.recordMetric("application", "memory_alloc", float64(m.Alloc), 0, "bytes")
	s.recordMetric("application", "memory_total_alloc", float64(m.TotalAlloc), 0, "bytes")
	s.recordMetric("application", "memory_sys", float64(m.Sys), 0, "bytes")
	s.recordMetric("application", "memory_heap_alloc", float64(m.HeapAlloc), 0, "bytes")
	s.recordMetric("application", "memory_heap_sys", float64(m.HeapSys), 0, "bytes")

	// Goroutine数量
			s.recordMetric("application", "goroutine_count", float64(runtime.NumGoroutine()), float64(s.config.ApplicationMonitoring.GoroutineThreshold), "count")

	// GC统计
			s.recordMetric("application", "gc_count", float64(m.NumGC), 0, "count")
			s.recordMetric("application", "gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), float64(s.config.ApplicationMonitoring.GCThreshold), "ns")

	// 内存泄漏检测（简单实现）
	if m.Alloc > 0 {
		leakRatio := float64(m.Alloc) / float64(m.Sys) * 100
		s.recordMetric("application", "memory_leak_ratio", leakRatio, s.config.ApplicationMonitoring.MemoryLeakThreshold, "%")
	}
}

// collectDatabaseMetrics 收集数据库指标
func (s *MonitoringService) collectDatabaseMetrics() {
	// 连接数统计
	var connectionCount int64
	s.db.Raw("SELECT COUNT(*) FROM information_schema.processlist").Scan(&connectionCount)
			s.recordMetric("database", "connection_count", float64(connectionCount), float64(s.config.DatabaseMonitoring.ConnectionThreshold), "count")

	// 慢查询统计
	var slowQueryCount int64
	s.db.Raw("SELECT COUNT(*) FROM information_schema.processlist WHERE TIME > ?", s.config.DatabaseMonitoring.SlowQueryThreshold.Seconds()).Scan(&slowQueryCount)
			s.recordMetric("database", "slow_query_count", float64(slowQueryCount), 0, "count")

	// 表大小统计
	var tableSizes []struct {
		TableName string  `json:"table_name"`
		Size      float64 `json:"size"`
	}
	s.db.Raw(`
		SELECT 
			table_name,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) AS size
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&tableSizes)

	for _, table := range tableSizes {
		s.recordMetric("database", "table_size_"+table.TableName, table.Size, float64(s.config.DatabaseMonitoring.TableSizeThreshold)/1024/1024, "MB")
	}
}

// collectCacheMetrics 收集缓存指标
func (s *MonitoringService) collectCacheMetrics() {
	// 这里需要根据实际使用的缓存系统来实现
	// 示例：Redis指标收集
	if s.config.CacheMonitoring.Enabled {
		// 这里应该实现Redis指标收集
		// 由于没有Redis连接，这里只是示例
		s.recordMetric("cache", "redis_connected", 1, 0, "status")
	}
}

// collectBusinessMetrics 收集业务指标
func (s *MonitoringService) collectBusinessMetrics() {
	// 用户活跃度
	var activeUsers int64
	s.db.Model(&Models.User{}).Where("last_login_at > ?", time.Now().Add(-24*time.Hour)).Count(&activeUsers)
			s.recordMetric("business", "active_users_24h", float64(activeUsers), float64(s.config.BusinessMonitoring.UserActivityThreshold), "count")

	// API调用次数（这里需要从日志或其他地方统计）
			s.recordMetric("business", "api_calls_total", 0, float64(s.config.BusinessMonitoring.APIUsageThreshold), "count")

		// 错误日志数量（暂时使用固定值，因为LogStatistics模型不存在）
	errorLogCount := int64(0)
	s.recordMetric("business", "error_logs_1h", float64(errorLogCount), float64(s.config.BusinessMonitoring.ErrorLogThreshold), "count")

	// 安全事件数量
	var securityEventCount int64
	s.db.Model(&Models.SecurityEvent{}).Where("created_at > ?", time.Now().Add(-1*time.Hour)).Count(&securityEventCount)
			s.recordMetric("business", "security_events_1h", float64(securityEventCount), float64(s.config.BusinessMonitoring.SecurityEventThreshold), "count")
}

// recordMetric 记录监控指标
func (s *MonitoringService) recordMetric(metricType, name string, value, threshold float64, unit string) {
	metric := &Models.MonitoringMetric{
		Type:        metricType,
		Name:        name,
		Value:       value,
		Unit:        unit,
		Threshold:   threshold,
		Status:      "normal",
		Severity:    "info",
		Description: fmt.Sprintf("%s: %f %s", name, value, unit),
		Timestamp:   time.Now(),
	}

	// 判断状态
	if threshold > 0 {
		if value >= threshold {
			metric.Status = "critical"
			metric.Severity = "critical"
		} else if value >= threshold*0.8 {
			metric.Status = "warning"
			metric.Severity = "warning"
		}
	}

	// 保存到数据库
	if err := s.db.Create(metric).Error; err != nil {
		log.Printf("保存监控指标失败: %v", err)
		return
	}

	// 发送到指标通道
	select {
	case s.metricsChan <- metric:
	default:
		log.Printf("指标通道已满，丢弃指标: %s", name)
	}

	// 检查告警规则
	s.checkAlertRules(metric)
}

// checkAlertRules 检查告警规则
func (s *MonitoringService) checkAlertRules(metric *Models.MonitoringMetric) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rule := range s.alertRules {
		if rule.MetricType == metric.Type && rule.MetricName == metric.Name {
			if s.evaluateRule(rule, metric) {
				s.triggerAlert(rule, metric)
			}
		}
	}
}

// evaluateRule 评估告警规则
func (s *MonitoringService) evaluateRule(rule *Models.AlertRule, metric *Models.MonitoringMetric) bool {
	switch rule.Condition {
	case ">":
		return metric.Value > rule.Threshold
	case ">=":
		return metric.Value >= rule.Threshold
	case "<":
		return metric.Value < rule.Threshold
	case "<=":
		return metric.Value <= rule.Threshold
	case "==":
		return metric.Value == rule.Threshold
	case "!=":
		return metric.Value != rule.Threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (s *MonitoringService) triggerAlert(rule *Models.AlertRule, metric *Models.MonitoringMetric) {
	// 检查是否在抑制期内
	if rule.Suppression {
		if s.isSuppressed(rule.ID) {
			return
		}
	}

	alert := &Models.Alert{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Type:        rule.Type,
		MetricType:  metric.Type,
		MetricName:  metric.Name,
		Value:       metric.Value,
		Threshold:   rule.Threshold,
		Severity:    rule.Severity,
		Status:      "active",
		Message:     fmt.Sprintf("指标 %s 触发告警: 当前值 %.2f %s, 阈值 %.2f", metric.Name, metric.Value, metric.Unit, rule.Threshold),
		Description: fmt.Sprintf("监控指标 %s 超过阈值 %.2f，当前值为 %.2f %s", metric.Name, rule.Threshold, metric.Value, metric.Unit),
		FiredAt:     time.Now(),
		EscalationLevel: 0,
		Suppressed:  false,
	}

	// 保存告警
	if err := s.db.Create(alert).Error; err != nil {
		log.Printf("保存告警失败: %v", err)
		return
	}

	// 发送到告警通道
	select {
	case s.alertChan <- alert:
	default:
		log.Printf("告警通道已满，丢弃告警: %s", alert.Message)
	}

	// 发送通知
	s.sendNotifications(alert)
}

// isSuppressed 检查是否被抑制
func (s *MonitoringService) isSuppressed(ruleID uint) bool {
	var count int64
	s.db.Model(&Models.Alert{}).
		Where("rule_id = ? AND fired_at > ? AND suppressed = ?", 
			ruleID, 
			time.Now().Add(-time.Duration(s.config.AlertConfig.SuppressionWindow)*time.Second),
			false).
		Count(&count)
	
	return count > 0
}

// startAlertProcessor 启动告警处理器
func (s *MonitoringService) startAlertProcessor() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case alert := <-s.alertChan:
			s.processAlert(alert)
		}
	}
}

// processAlert 处理告警
func (s *MonitoringService) processAlert(alert *Models.Alert) {
	// 检查是否需要升级
	if s.shouldEscalate(alert) {
		s.escalateAlert(alert)
	}

	// 检查是否自动解决
	if s.config.AlertConfig.AutoResolveEnabled {
		go s.autoResolveAlert(alert)
	}
}

// shouldEscalate 检查是否需要升级
func (s *MonitoringService) shouldEscalate(alert *Models.Alert) bool {
	if !s.config.AlertConfig.EscalationEnabled {
		return false
	}

	if alert.EscalationLevel >= s.config.AlertConfig.MaxEscalationLevel {
		return false
	}

	// 检查是否超过升级延迟时间
	return time.Since(alert.FiredAt) > s.config.AlertConfig.EscalationDelay
}

// escalateAlert 升级告警
func (s *MonitoringService) escalateAlert(alert *Models.Alert) {
	alert.EscalationLevel++
	
	if err := s.db.Save(alert).Error; err != nil {
		log.Printf("升级告警失败: %v", err)
		return
	}

	// 重新发送通知
	s.sendNotifications(alert)
}

// autoResolveAlert 自动解决告警
func (s *MonitoringService) autoResolveAlert(alert *Models.Alert) {
	time.Sleep(s.config.AlertConfig.AutoResolveDelay)
	
	// 检查告警是否仍然存在
	var currentAlert Models.Alert
	if err := s.db.First(&currentAlert, alert.ID).Error; err != nil {
		return
	}

	if currentAlert.Status == "active" {
		now := time.Now()
		currentAlert.Status = "resolved"
		currentAlert.ResolvedAt = &now
		
		if err := s.db.Save(&currentAlert).Error; err != nil {
			log.Printf("自动解决告警失败: %v", err)
		}
	}
}

// startNotificationProcessor 启动通知处理器
func (s *MonitoringService) startNotificationProcessor() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case notification := <-s.notificationChan:
			s.processNotification(notification)
		}
	}
}

// processNotification 处理通知
func (s *MonitoringService) processNotification(notification *Models.NotificationRecord) {
	// 发送通知
	success := s.sendNotification(notification)
	
	if success {
		now := time.Now()
		notification.Status = "sent"
		notification.SentAt = &now
	} else {
		notification.Status = "failed"
		notification.RetryCount++
		
		// 重试逻辑
		if notification.RetryCount < notification.MaxRetries {
			notification.Status = "retrying"
			go s.retryNotification(notification)
		}
	}

	// 更新通知记录
	if err := s.db.Save(notification).Error; err != nil {
		log.Printf("更新通知记录失败: %v", err)
	}
}

// sendNotification 发送通知
func (s *MonitoringService) sendNotification(notification *Models.NotificationRecord) bool {
	switch notification.Channel {
	case "email":
		return s.sendEmailNotification(notification)
	case "webhook":
		return s.sendWebhookNotification(notification)
	case "slack":
		return s.sendSlackNotification(notification)
	case "dingtalk":
		return s.sendDingTalkNotification(notification)
	case "sms":
		return s.sendSMSNotification(notification)
	default:
		log.Printf("不支持的通知渠道: %s", notification.Channel)
		return false
	}
}

// sendEmailNotification 发送邮件通知
func (s *MonitoringService) sendEmailNotification(notification *Models.NotificationRecord) bool {
	if !s.config.NotificationConfig.Email.Enabled {
		return false
	}

	auth := smtp.PlainAuth("", 
		s.config.NotificationConfig.Email.Username, 
		s.config.NotificationConfig.Email.Password, 
		s.config.NotificationConfig.Email.SMTPHost)

	to := strings.Split(s.config.NotificationConfig.Email.ToAddresses, ",")
	
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", 
		strings.Join(to, ","), 
		notification.Subject, 
		notification.Content)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.config.NotificationConfig.Email.SMTPHost, s.config.NotificationConfig.Email.SMTPPort),
		auth,
		s.config.NotificationConfig.Email.FromAddress,
		to,
		[]byte(msg))

	if err != nil {
		notification.Error = err.Error()
		return false
	}

	return true
}

// sendWebhookNotification 发送Webhook通知
func (s *MonitoringService) sendWebhookNotification(notification *Models.NotificationRecord) bool {
	if !s.config.NotificationConfig.Webhook.Enabled {
		return false
	}

	client := &http.Client{
		Timeout: s.config.NotificationConfig.Webhook.Timeout,
	}

	payload := map[string]interface{}{
		"alert_id":   notification.AlertID,
		"subject":    notification.Subject,
		"content":    notification.Content,
		"timestamp":  time.Now().Unix(),
		"recipient":  notification.Recipient,
	}

	jsonData, _ := json.Marshal(payload)
	
	req, err := http.NewRequest(s.config.NotificationConfig.Webhook.Method, 
		s.config.NotificationConfig.Webhook.URL, 
		bytes.NewBuffer(jsonData))
	if err != nil {
		notification.Error = err.Error()
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	
	// 添加自定义头部
	if s.config.NotificationConfig.Webhook.Headers != "" {
		headers := strings.Split(s.config.NotificationConfig.Webhook.Headers, ",")
		for _, header := range headers {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		notification.Error = err.Error()
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	notification.Response = string(body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}

	notification.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	return false
}

// sendSlackNotification 发送Slack通知
func (s *MonitoringService) sendSlackNotification(notification *Models.NotificationRecord) bool {
	if !s.config.NotificationConfig.Slack.Enabled {
		return false
	}

	payload := map[string]interface{}{
		"text":      notification.Content,
		"username":  s.config.NotificationConfig.Slack.Username,
		"icon_emoji": s.config.NotificationConfig.Slack.IconEmoji,
	}

	if s.config.NotificationConfig.Slack.Channel != "" {
		payload["channel"] = s.config.NotificationConfig.Slack.Channel
	}

	jsonData, _ := json.Marshal(payload)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(s.config.NotificationConfig.Slack.WebhookURL, 
		"application/json", 
		bytes.NewBuffer(jsonData))
	if err != nil {
		notification.Error = err.Error()
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	notification.Response = string(body)

	if resp.StatusCode == 200 {
		return true
	}

	notification.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	return false
}

// sendDingTalkNotification 发送钉钉通知
func (s *MonitoringService) sendDingTalkNotification(notification *Models.NotificationRecord) bool {
	if !s.config.NotificationConfig.DingTalk.Enabled {
		return false
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": notification.Content,
		},
	}

	// 添加@功能
	if s.config.NotificationConfig.DingTalk.AtMobiles != "" {
		mobiles := strings.Split(s.config.NotificationConfig.DingTalk.AtMobiles, ",")
		payload["at"] = map[string]interface{}{
			"atMobiles": mobiles,
			"isAtAll":   false,
		}
	}

	jsonData, _ := json.Marshal(payload)
	
	// 计算签名
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, s.config.NotificationConfig.DingTalk.Secret)
	h := hmac.New(sha256.New, []byte(s.config.NotificationConfig.DingTalk.Secret))
	h.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	url := fmt.Sprintf("%s&timestamp=%d&sign=%s", 
		s.config.NotificationConfig.DingTalk.WebhookURL, 
		timestamp, 
		sign)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		notification.Error = err.Error()
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	notification.Response = string(body)

	if resp.StatusCode == 200 {
		return true
	}

	notification.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	return false
}

// sendSMSNotification 发送短信通知
func (s *MonitoringService) sendSMSNotification(notification *Models.NotificationRecord) bool {
	if !s.config.NotificationConfig.SMS.Enabled {
		return false
	}

	// 这里需要根据具体的短信服务商来实现
	// 示例：阿里云短信服务
	log.Printf("发送短信通知到: %s, 内容: %s", notification.Recipient, notification.Content)
	return true
}

// retryNotification 重试通知
func (s *MonitoringService) retryNotification(notification *Models.NotificationRecord) {
	time.Sleep(time.Duration(notification.RetryCount) * time.Second)
	
	select {
	case s.notificationChan <- notification:
	default:
		log.Printf("通知通道已满，无法重试通知: %d", notification.ID)
	}
}

// sendNotifications 发送通知
func (s *MonitoringService) sendNotifications(alert *Models.Alert) {
	// 解析通知渠道
	var channels []string
	if err := json.Unmarshal([]byte(alert.Tags), &channels); err != nil {
		channels = []string{"email"} // 默认使用邮件
	}

	for _, channel := range channels {
		notification := &Models.NotificationRecord{
			AlertID:    alert.ID,
			Channel:    channel,
			Recipient:  s.getRecipient(channel),
			Subject:    fmt.Sprintf("[%s] %s", alert.Severity, alert.RuleName),
			Content:    alert.Message,
			Status:     "pending",
			RetryCount: 0,
			MaxRetries: s.config.NotificationConfig.Webhook.RetryCount,
		}

		// 保存通知记录
		if err := s.db.Create(notification).Error; err != nil {
			log.Printf("保存通知记录失败: %v", err)
			continue
		}

		// 发送到通知通道
		select {
		case s.notificationChan <- notification:
		default:
			log.Printf("通知通道已满，丢弃通知: %s", channel)
		}
	}
}

// getRecipient 获取接收者
func (s *MonitoringService) getRecipient(channel string) string {
	switch channel {
	case "email":
		return s.config.NotificationConfig.Email.ToAddresses
	case "slack":
		return s.config.NotificationConfig.Slack.Channel
	case "dingtalk":
		return s.config.NotificationConfig.DingTalk.AtMobiles
	case "sms":
		return s.config.NotificationConfig.SMS.PhoneNumbers
	default:
		return ""
	}
}

// startDataCleanup 启动数据清理
func (s *MonitoringService) startDataCleanup() {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOldData()
		}
	}
}

// cleanupOldData 清理旧数据
func (s *MonitoringService) cleanupOldData() {
	cutoffTime := time.Now().Add(-s.config.BaseConfig.RetentionPeriod)

	// 清理监控指标
	if err := s.db.Where("timestamp < ?", cutoffTime).Delete(&Models.MonitoringMetric{}).Error; err != nil {
		log.Printf("清理监控指标失败: %v", err)
	}

	// 清理告警记录
	if err := s.db.Where("fired_at < ?", cutoffTime).Delete(&Models.Alert{}).Error; err != nil {
		log.Printf("清理告警记录失败: %v", err)
	}

	// 清理通知记录
	if err := s.db.Where("created_at < ?", cutoffTime).Delete(&Models.NotificationRecord{}).Error; err != nil {
		log.Printf("清理通知记录失败: %v", err)
	}

	// 清理监控事件
	if err := s.db.Where("timestamp < ?", cutoffTime).Delete(&Models.MonitoringEvent{}).Error; err != nil {
		log.Printf("清理监控事件失败: %v", err)
	}

	log.Printf("数据清理完成，清理时间点: %s", cutoffTime.Format("2006-01-02 15:04:05"))
}

// Stop 停止监控服务
func (s *MonitoringService) Stop() {
	s.cancel()
	close(s.stopChan)
	log.Println("监控告警服务已停止")
}

// GetMetrics 获取监控指标
func (s *MonitoringService) GetMetrics(metricType, name string, limit int) ([]Models.MonitoringMetric, error) {
	var metrics []Models.MonitoringMetric
	query := s.db.Model(&Models.MonitoringMetric{})
	
	if metricType != "" {
		query = query.Where("type = ?", metricType)
	}
	if name != "" {
		query = query.Where("name = ?", name)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("timestamp DESC").Find(&metrics).Error
	return metrics, err
}

// GetAlerts 获取告警记录
func (s *MonitoringService) GetAlerts(status, severity string, limit int) ([]Models.Alert, error) {
	var alerts []Models.Alert
	query := s.db.Model(&Models.Alert{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("fired_at DESC").Find(&alerts).Error
	return alerts, err
}

// AcknowledgeAlert 确认告警
func (s *MonitoringService) AcknowledgeAlert(alertID uint, userID uint) error {
	now := time.Now()
	return s.db.Model(&Models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":           "acknowledged",
			"acknowledged_at":  &now,
			"acknowledged_by":  userID,
		}).Error
}

// ResolveAlert 解决告警
func (s *MonitoringService) ResolveAlert(alertID uint, userID uint) error {
	now := time.Now()
	return s.db.Model(&Models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": &now,
			"resolved_by": userID,
		}).Error
}

// CreateAlertRule 创建告警规则
func (s *MonitoringService) CreateAlertRule(rule *Models.AlertRule) error {
	if err := s.db.Create(rule).Error; err != nil {
		return err
	}
	
	// 重新加载告警规则
	s.loadAlertRules()
	return nil
}

// UpdateAlertRule 更新告警规则
func (s *MonitoringService) UpdateAlertRule(rule *Models.AlertRule) error {
	if err := s.db.Save(rule).Error; err != nil {
		return err
	}
	
	// 重新加载告警规则
	s.loadAlertRules()
	return nil
}

// DeleteAlertRule 删除告警规则
func (s *MonitoringService) DeleteAlertRule(ruleID uint) error {
	if err := s.db.Delete(&Models.AlertRule{}, ruleID).Error; err != nil {
		return err
	}
	
	// 重新加载告警规则
	s.loadAlertRules()
	return nil
}

// GetSystemHealth 获取系统健康状态
func (s *MonitoringService) GetSystemHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"metrics":   make(map[string]interface{}),
		"alerts":    make(map[string]interface{}),
	}

	// 检查活跃告警数量
	var activeAlerts int64
	s.db.Model(&Models.Alert{}).Where("status = ?", "active").Count(&activeAlerts)
	health["alerts"].(map[string]interface{})["active_count"] = activeAlerts

	if activeAlerts > 0 {
		health["status"] = "warning"
	}

	// 检查严重告警数量
	var criticalAlerts int64
	s.db.Model(&Models.Alert{}).Where("status = ? AND severity = ?", "active", "critical").Count(&criticalAlerts)
	health["alerts"].(map[string]interface{})["critical_count"] = criticalAlerts

	if criticalAlerts > 0 {
		health["status"] = "critical"
	}

	// 获取最新指标
	latestMetrics, _ := s.GetMetrics("", "", 10)
	health["metrics"].(map[string]interface{})["latest"] = latestMetrics

	return health
}
