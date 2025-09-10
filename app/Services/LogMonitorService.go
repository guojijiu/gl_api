package Services

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Models"
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

// LogMonitorService 日志监控服务
//
// 重要功能说明：
// 1. 实时日志监控：监控日志级别、频率、异常模式
// 2. 智能告警：基于规则和阈值的自动告警
// 3. 日志分析：模式识别、异常检测、趋势分析
// 4. 性能监控：日志写入性能、系统资源使用
// 5. 报告生成：定期报告、实时统计、历史分析
// 6. 告警通知：邮件、Webhook、Slack等多种通知方式
type LogMonitorService struct {
	logManager *LogManagerService
	config     *Config.LogConfig
	ctx        context.Context
	cancel     context.CancelFunc
	
	// 监控规则
	rules      []*LogRule
	rulesMu    sync.RWMutex
	
	// 告警状态
	alerts     map[string]*Models.Alert
	alertsMu   sync.RWMutex
	
	// 统计信息
	stats      *MonitorStats
	statsMu    sync.RWMutex
	
	// 通知器
	notifiers  []LogNotifier
}

// LogRule 日志监控规则
type LogRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	
	// 匹配条件
	Logger      string                 `json:"logger"`      // 日志记录器名称
	Level       Config.LogLevel        `json:"level"`       // 日志级别
	Pattern     string                 `json:"pattern"`     // 正则表达式模式
	Keywords    []string               `json:"keywords"`    // 关键词列表
	Fields      map[string]interface{} `json:"fields"`      // 字段匹配
	
	// 触发条件
	Threshold   int                    `json:"threshold"`   // 触发阈值
	TimeWindow  time.Duration          `json:"time_window"` // 时间窗口
	Count       int                    `json:"count"`       // 当前计数
	
	// 告警配置
	AlertLevel  string                 `json:"alert_level"` // 告警级别
	Message     string                 `json:"message"`     // 告警消息
	Actions     []string               `json:"actions"`     // 执行动作
	
	// 状态
	LastTrigger time.Time              `json:"last_trigger"`
	TriggerCount int64                 `json:"trigger_count"`
	
	// 编译后的正则表达式
	regex       *regexp.Regexp
}

// 使用 Models.Alert 代替本地定义

// MonitorStats 监控统计信息
type MonitorStats struct {
	mu              sync.RWMutex
	TotalRules      int                    `json:"total_rules"`
	ActiveRules     int                    `json:"active_rules"`
	TotalAlerts     int64                  `json:"total_alerts"`
	ActiveAlerts    int                    `json:"active_alerts"`
	ResolvedAlerts  int64                  `json:"resolved_alerts"`
	RulesByLogger   map[string]int         `json:"rules_by_logger"`
	AlertsByLevel   map[string]int64       `json:"alerts_by_level"`
	LastReset       time.Time              `json:"last_reset"`
}

// LogNotifier 日志通知接口
type LogNotifier interface {
	Notify(alert *Models.Alert) error
	Name() string
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	config *Config.EmailConfig
}

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	url     string
	headers map[string]string
}

// SlackNotifier Slack通知器
type SlackNotifier struct {
	webhookURL string
	channel    string
}

// NewLogMonitorService 创建日志监控服务
func NewLogMonitorService(logManager *LogManagerService, config *Config.LogConfig) *LogMonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &LogMonitorService{
		logManager: logManager,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
		rules:      make([]*LogRule, 0),
		alerts:     make(map[string]*Models.Alert),
		stats: &MonitorStats{
			RulesByLogger: make(map[string]int),
			AlertsByLevel: make(map[string]int64),
			LastReset:     time.Now(),
		},
		notifiers: make([]LogNotifier, 0),
	}
	
	// 初始化默认规则
	service.initDefaultRules()
	
	// 初始化通知器
	service.initNotifiers()
	
	// 启动监控
	go service.startMonitoring()
	
	return service
}

// initDefaultRules 初始化默认监控规则
func (s *LogMonitorService) initDefaultRules() {
	defaultRules := []*LogRule{
		{
			ID:          "error_threshold",
			Name:        "错误日志阈值",
			Description: "监控错误日志数量，超过阈值时告警",
			Enabled:     true,
			Logger:      "error",
			Level:       Config.LogLevelError,
			Threshold:   10,
			TimeWindow:  5 * time.Minute,
			AlertLevel:  "warning",
			Message:     "错误日志数量过多，请检查系统状态",
			Actions:     []string{"email", "webhook"},
		},
		{
			ID:          "slow_query_detection",
			Name:        "慢查询检测",
			Description: "检测SQL慢查询",
			Enabled:     true,
			Logger:      "sql",
			Level:       Config.LogLevelWarning,
			Keywords:    []string{"slow_query"},
			Threshold:   5,
			TimeWindow:  1 * time.Minute,
			AlertLevel:  "warning",
			Message:     "检测到多个慢查询，请优化数据库性能",
			Actions:     []string{"email"},
		},
		{
			ID:          "security_events",
			Name:        "安全事件监控",
			Description: "监控安全相关日志",
			Enabled:     true,
			Logger:      "security",
			Level:       Config.LogLevelWarning,
			Threshold:   1,
			TimeWindow:  1 * time.Minute,
			AlertLevel:  "critical",
			Message:     "检测到安全事件，请立即处理",
			Actions:     []string{"email", "slack", "webhook"},
		},
		{
			ID:          "high_error_rate",
			Name:        "高错误率检测",
			Description: "监控错误率是否过高",
			Enabled:     true,
			Logger:      "*",
			Level:       Config.LogLevelError,
			Threshold:   20,
			TimeWindow:  10 * time.Minute,
			AlertLevel:  "critical",
			Message:     "系统错误率过高，可能存在严重问题",
			Actions:     []string{"email", "slack", "webhook"},
		},
	}
	
	for _, rule := range defaultRules {
		s.AddRule(rule)
	}
}

// initNotifiers 初始化通知器
func (s *LogMonitorService) initNotifiers() {
	// 这里应该从配置中读取通知器配置
	// 暂时添加示例通知器
	
	// 邮件通知器
	if s.config.ErrorLog.NotifyEmail != "" {
		emailNotifier := &EmailNotifier{
			config: &Config.EmailConfig{
				Host:     "localhost",
				Port:     587,
				Username: "noreply@example.com",
				Password: "password",
				From:     "noreply@example.com",
			},
		}
		s.notifiers = append(s.notifiers, emailNotifier)
	}
	
	// Webhook通知器
	webhookNotifier := &WebhookNotifier{
		url: "http://localhost:8080/webhook/logs",
		headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	s.notifiers = append(s.notifiers, webhookNotifier)
	
	// Slack通知器
	slackNotifier := &SlackNotifier{
		webhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
		channel:    "#alerts",
	}
	s.notifiers = append(s.notifiers, slackNotifier)
}

// AddRule 添加监控规则
func (s *LogMonitorService) AddRule(rule *LogRule) error {
	// 编译正则表达式
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("正则表达式编译失败: %v", err)
		}
		rule.regex = regex
	}
	
	s.rulesMu.Lock()
	defer s.rulesMu.Unlock()
	
	s.rules = append(s.rules, rule)
	
	// 更新统计
	s.stats.mu.Lock()
	s.stats.TotalRules++
	if rule.Enabled {
		s.stats.ActiveRules++
	}
	if rule.Logger != "*" {
		s.stats.RulesByLogger[rule.Logger]++
	}
	s.stats.mu.Unlock()
	
	return nil
}

// RemoveRule 移除监控规则
func (s *LogMonitorService) RemoveRule(ruleID string) error {
	s.rulesMu.Lock()
	defer s.rulesMu.Unlock()
	
	for i, rule := range s.rules {
		if rule.ID == ruleID {
			// 更新统计
			s.stats.mu.Lock()
			if rule.Enabled {
				s.stats.ActiveRules--
			}
			if rule.Logger != "*" {
				s.stats.RulesByLogger[rule.Logger]--
			}
			s.stats.mu.Unlock()
			
			// 移除规则
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			return nil
		}
	}
	
	return fmt.Errorf("规则不存在: %s", ruleID)
}

// UpdateRule 更新监控规则
func (s *LogMonitorService) UpdateRule(ruleID string, updates map[string]interface{}) error {
	s.rulesMu.Lock()
	defer s.rulesMu.Unlock()
	
	for _, rule := range s.rules {
		if rule.ID == ruleID {
			// 更新字段
			for key, value := range updates {
				switch key {
				case "enabled":
					if enabled, ok := value.(bool); ok {
						oldEnabled := rule.Enabled
						rule.Enabled = enabled
						
						// 更新统计
						s.stats.mu.Lock()
						if enabled && !oldEnabled {
							s.stats.ActiveRules++
						} else if !enabled && oldEnabled {
							s.stats.ActiveRules--
						}
						s.stats.mu.Unlock()
					}
				case "threshold":
					if threshold, ok := value.(int); ok {
						rule.Threshold = threshold
					}
				case "time_window":
					if timeWindow, ok := value.(time.Duration); ok {
						rule.TimeWindow = timeWindow
					}
				case "message":
					if message, ok := value.(string); ok {
						rule.Message = message
					}
				}
			}
			return nil
		}
	}
	
	return fmt.Errorf("规则不存在: %s", ruleID)
}

// GetRules 获取所有监控规则
func (s *LogMonitorService) GetRules() []*LogRule {
	s.rulesMu.RLock()
	defer s.rulesMu.RUnlock()
	
	rules := make([]*LogRule, len(s.rules))
	copy(rules, s.rules)
	return rules
}

// GetRule 获取指定规则
func (s *LogMonitorService) GetRule(ruleID string) *LogRule {
	s.rulesMu.RLock()
	defer s.rulesMu.RUnlock()
	
	for _, rule := range s.rules {
		if rule.ID == ruleID {
			return rule
		}
	}
	return nil
}

// startMonitoring 启动监控
func (s *LogMonitorService) startMonitoring() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.checkRules()
		case <-s.ctx.Done():
			return
		}
	}
}

// checkRules 检查所有规则
func (s *LogMonitorService) checkRules() {
	s.rulesMu.RLock()
	rules := make([]*LogRule, len(s.rules))
	copy(rules, s.rules)
	s.rulesMu.RUnlock()
	
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		s.checkRule(rule)
	}
}

// checkRule 检查单个规则
func (s *LogMonitorService) checkRule(rule *LogRule) {
	// 获取日志统计
	stats := s.logManager.GetStats()
	
	// 检查是否触发
	triggered := false
	
	// 根据规则类型检查
	switch {
	case rule.Logger != "*":
		// 检查特定日志记录器
		if count, exists := stats.LogsByLogger[rule.Logger]; exists {
			if count >= int64(rule.Threshold) {
				triggered = true
			}
		}
	case rule.Level != "":
		// 检查特定日志级别
		if count, exists := stats.LogsByLevel[rule.Level]; exists {
			if count >= int64(rule.Threshold) {
				triggered = true
			}
		}
	}
	
	if triggered {
		s.triggerAlert(rule)
	}
}

// triggerAlert 触发告警
func (s *LogMonitorService) triggerAlert(rule *LogRule) {
	// 检查是否在时间窗口内已经触发过
	if time.Since(rule.LastTrigger) < rule.TimeWindow {
		return
	}
	
	// 创建告警
	alert := &Models.Alert{
		RuleName:    rule.Name,
		Type:        "log_monitoring",
		MetricType:  "log_count",
		MetricName:  rule.Name,
		Value:       float64(rule.TriggerCount + 1),
		Threshold:   float64(rule.Threshold),
		Severity:    rule.AlertLevel,
		Status:      "active",
		Message:     rule.Message,
		Description: rule.Description,
		FiredAt:     time.Now(),
		Metadata:    fmt.Sprintf(`{"rule_id":"%s","threshold":%d,"time_window":"%s"}`, rule.ID, rule.Threshold, rule.TimeWindow.String()),
	}
	
	// 保存告警
	s.alertsMu.Lock()
	alertID := fmt.Sprintf("alert_%s_%d", rule.ID, time.Now().Unix())
	s.alerts[alertID] = alert
	s.alertsMu.Unlock()
	
	// 更新规则状态
	rule.LastTrigger = time.Now()
	rule.TriggerCount++
	
	// 更新统计
	s.stats.mu.Lock()
	s.stats.TotalAlerts++
	s.stats.ActiveAlerts++
	s.stats.AlertsByLevel[rule.AlertLevel]++
	s.stats.mu.Unlock()
	
	// 发送通知
	go s.sendNotifications(alert, alertID)
	
	// 记录告警日志
	s.logManager.LogSecurity(
		context.Background(),
		fmt.Sprintf("监控规则触发: %s", rule.Name),
		Config.LogLevelWarning,
		map[string]interface{}{
			"rule_id":    rule.ID,
			"alert_id":   alertID,
			"alert_level": rule.AlertLevel,
			"message":    rule.Message,
		},
	)
}

// sendNotifications 发送通知
func (s *LogMonitorService) sendNotifications(alert *Models.Alert, alertID string) {
	for _, notifier := range s.notifiers {
		go func(n LogNotifier) {
			if err := n.Notify(alert); err != nil {
				s.logManager.LogError(
					context.Background(),
					err,
					fmt.Sprintf("通知发送失败: %s", n.Name()),
					map[string]interface{}{
						"notifier": n.Name(),
						"alert_id": alertID,
					},
				)
			}
		}(notifier)
	}
}

// GetAlerts 获取告警列表
func (s *LogMonitorService) GetAlerts(status string, limit int) []*Models.Alert {
	s.alertsMu.RLock()
	defer s.alertsMu.RUnlock()
	
	var alerts []*Models.Alert
	for _, alert := range s.alerts {
		if status == "" || alert.Status == status {
			alerts = append(alerts, alert)
		}
	}
	
	// 按时间排序（按触发时间倒序）
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].FiredAt.After(alerts[j].FiredAt)
	})
	
	if len(alerts) > limit {
		alerts = alerts[:limit]
	}
	
	return alerts
}

// ResolveAlert 解决告警
func (s *LogMonitorService) ResolveAlert(alertID string) error {
	s.alertsMu.Lock()
	defer s.alertsMu.Unlock()
	
	alert, exists := s.alerts[alertID]
	if !exists {
		return fmt.Errorf("告警不存在: %s", alertID)
	}
	
	if alert.Status == "resolved" {
		return fmt.Errorf("告警已经解决: %s", alertID)
	}
	
	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	
	// 更新统计
	s.stats.mu.Lock()
	s.stats.ActiveAlerts--
	s.stats.ResolvedAlerts++
	s.stats.mu.Unlock()
	
	return nil
}

// AcknowledgeAlert 确认告警
func (s *LogMonitorService) AcknowledgeAlert(alertID string) error {
	s.alertsMu.Lock()
	defer s.alertsMu.Unlock()
	
	alert, exists := s.alerts[alertID]
	if !exists {
		return fmt.Errorf("告警不存在: %s", alertID)
	}
	
	if alert.Status == "acknowledged" {
		return fmt.Errorf("告警已经确认: %s", alertID)
	}
	
	now := time.Now()
	alert.Status = "acknowledged"
	alert.AcknowledgedAt = &now
	
	return nil
}

// GetStats 获取监控统计信息
func (s *LogMonitorService) GetStats() *MonitorStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()
	
	stats := &MonitorStats{
		TotalRules:     s.stats.TotalRules,
		ActiveRules:    s.stats.ActiveRules,
		TotalAlerts:    s.stats.TotalAlerts,
		ActiveAlerts:   s.stats.ActiveAlerts,
		ResolvedAlerts: s.stats.ResolvedAlerts,
		RulesByLogger:  make(map[string]int),
		AlertsByLevel:  make(map[string]int64),
		LastReset:      s.stats.LastReset,
	}
	
	for logger, count := range s.stats.RulesByLogger {
		stats.RulesByLogger[logger] = count
	}
	
	for level, count := range s.stats.AlertsByLevel {
		stats.AlertsByLevel[level] = count
	}
	
	return stats
}

// Close 关闭监控服务
func (s *LogMonitorService) Close() error {
	s.cancel()
	return nil
}

// 通知器实现

// Notify 邮件通知器
func (n *EmailNotifier) Notify(alert *Models.Alert) error {
	// 这里应该实现邮件发送逻辑
	// 暂时只打印日志
	fmt.Printf("邮件通知: %s - %s\n", alert.Severity, alert.Message)
	return nil
}

func (n *EmailNotifier) Name() string {
	return "email"
}

// Notify Webhook通知器
func (n *WebhookNotifier) Notify(alert *Models.Alert) error {
	// 这里应该实现HTTP POST请求
	// 暂时只打印日志
	fmt.Printf("Webhook通知: %s - %s\n", alert.Severity, alert.Message)
	return nil
}

func (n *WebhookNotifier) Name() string {
	return "webhook"
}

// Notify Slack通知器
func (n *SlackNotifier) Notify(alert *Models.Alert) error {
	// 这里应该实现Slack API调用
	// 暂时只打印日志
	fmt.Printf("Slack通知: %s - %s\n", alert.Severity, alert.Message)
	return nil
}

func (n *SlackNotifier) Name() string {
	return "slack"
}
