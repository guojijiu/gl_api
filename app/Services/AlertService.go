package Services

import (
	"fmt"
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertChannel 告警渠道
type AlertChannel string

const (
	AlertChannelEmail   AlertChannel = "email"
	AlertChannelSlack   AlertChannel = "slack"
	AlertChannelWebhook AlertChannel = "webhook"
)

// AlertRule 告警规则
type AlertRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Metric      string         `json:"metric"`
	Condition   string         `json:"condition"`
	Threshold   float64        `json:"threshold"`
	Duration    time.Duration  `json:"duration"`
	Level       AlertLevel     `json:"level"`
	Channels    []AlertChannel `json:"channels"`
	Enabled     bool           `json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Alert 告警实例
type Alert struct {
	ID         string     `json:"id"`
	RuleID     string     `json:"rule_id"`
	Level      AlertLevel `json:"level"`
	Message    string     `json:"message"`
	Metric     string     `json:"metric"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// AlertService 告警服务
type AlertService struct {
	emailService      *EmailService
	monitoringService *OptimizedMonitoringService
	rules             map[string]*AlertRule
	alerts            map[string]*Alert
}

// NewAlertService 创建告警服务
func NewAlertService(emailService *EmailService, monitoringService *OptimizedMonitoringService) *AlertService {
	return &AlertService{
		emailService:      emailService,
		monitoringService: monitoringService,
		rules:             make(map[string]*AlertRule),
		alerts:            make(map[string]*Alert),
	}
}

// AddRule 添加告警规则
func (a *AlertService) AddRule(rule *AlertRule) error {
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	a.rules[rule.ID] = rule

	return nil
}

// GetRules 获取所有告警规则
func (a *AlertService) GetRules() []*AlertRule {
	rules := make([]*AlertRule, 0, len(a.rules))
	for _, rule := range a.rules {
		rules = append(rules, rule)
	}
	return rules
}

// CheckAlerts 检查告警
func (a *AlertService) CheckAlerts() error {
	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		value, err := a.getMetricValue(rule.Metric)
		if err != nil {
			continue
		}

		if a.shouldTriggerAlert(rule, value) {
			a.triggerAlert(rule, value)
		} else {
			a.resolveAlert(rule)
		}
	}

	return nil
}

// getMetricValue 获取指标值
func (a *AlertService) getMetricValue(metric string) (float64, error) {
	switch metric {
	case "cpu_usage":
		return 75.0, nil
	case "memory_usage":
		return 80.0, nil
	case "error_rate":
		return 2.5, nil
	case "response_time":
		return 150.0, nil
	default:
		return 0.0, fmt.Errorf("未知指标: %s", metric)
	}
}

// shouldTriggerAlert 检查是否应该触发告警
func (a *AlertService) shouldTriggerAlert(rule *AlertRule, value float64) bool {
	switch rule.Condition {
	case ">":
		return value > rule.Threshold
	case ">=":
		return value >= rule.Threshold
	case "<":
		return value < rule.Threshold
	case "<=":
		return value <= rule.Threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (a *AlertService) triggerAlert(rule *AlertRule, value float64) {
	alertID := fmt.Sprintf("%s_%d", rule.ID, time.Now().Unix())

	alert := &Alert{
		ID:        alertID,
		RuleID:    rule.ID,
		Level:     rule.Level,
		Message:   fmt.Sprintf("指标 %s 当前值为 %.2f，超过阈值 %.2f", rule.Metric, value, rule.Threshold),
		Metric:    rule.Metric,
		Value:     value,
		Threshold: rule.Threshold,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	a.alerts[alertID] = alert
	a.sendAlertNotifications(alert, rule)
}

// resolveAlert 恢复告警
func (a *AlertService) resolveAlert(rule *AlertRule) {
	for _, alert := range a.alerts {
		if alert.RuleID == rule.ID && alert.Status == "active" {
			now := time.Now()
			alert.Status = "resolved"
			alert.ResolvedAt = &now
			a.sendResolveNotifications(alert, rule)
		}
	}
}

// sendAlertNotifications 发送告警通知
func (a *AlertService) sendAlertNotifications(alert *Alert, rule *AlertRule) {
	for _, channel := range rule.Channels {
		switch channel {
		case AlertChannelEmail:
			a.sendEmailAlert(alert, rule)
		}
	}
}

// sendResolveNotifications 发送恢复通知
func (a *AlertService) sendResolveNotifications(alert *Alert, rule *AlertRule) {
	for _, channel := range rule.Channels {
		switch channel {
		case AlertChannelEmail:
			a.sendEmailResolve(alert, rule)
		}
	}
}

// sendEmailAlert 发送邮件告警
func (a *AlertService) sendEmailAlert(alert *Alert, rule *AlertRule) {
	subject := fmt.Sprintf("[%s] 系统告警: %s", string(alert.Level), rule.Name)
	body := fmt.Sprintf(`
告警详情:
- 规则名称: %s
- 告警级别: %s
- 告警时间: %s
- 告警消息: %s
- 指标名称: %s
- 当前值: %.2f
- 阈值: %.2f
`, rule.Name, string(alert.Level), alert.CreatedAt.Format("2006-01-02 15:04:05"),
		alert.Message, alert.Metric, alert.Value, alert.Threshold)

	a.emailService.SendNotificationEmail("admin@example.com", subject, body)
}

// sendEmailResolve 发送邮件恢复通知
func (a *AlertService) sendEmailResolve(alert *Alert, rule *AlertRule) {
	subject := fmt.Sprintf("[恢复] 系统告警已恢复: %s", rule.Name)
	body := fmt.Sprintf(`
告警恢复:
- 规则名称: %s
- 恢复时间: %s
- 指标名称: %s
- 当前值: %.2f
- 阈值: %.2f
`, rule.Name, alert.ResolvedAt.Format("2006-01-02 15:04:05"),
		alert.Metric, alert.Value, alert.Threshold)

	a.emailService.SendNotificationEmail("admin@example.com", subject, body)
}

// GetAlerts 获取告警列表
func (a *AlertService) GetAlerts(status string, limit int) []*Alert {
	alerts := make([]*Alert, 0)

	for _, alert := range a.alerts {
		if status == "" || alert.Status == status {
			alerts = append(alerts, alert)
		}
	}

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	return alerts
}

// GetAlertStats 获取告警统计
func (a *AlertService) GetAlertStats() map[string]interface{} {
	stats := make(map[string]interface{})

	activeCount := 0
	resolvedCount := 0
	enabledRules := 0

	for _, alert := range a.alerts {
		if alert.Status == "active" {
			activeCount++
		} else {
			resolvedCount++
		}
	}

	for _, rule := range a.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	stats["total_alerts"] = len(a.alerts)
	stats["active_alerts"] = activeCount
	stats["resolved_alerts"] = resolvedCount
	stats["total_rules"] = len(a.rules)
	stats["enabled_rules"] = enabledRules

	return stats
}
