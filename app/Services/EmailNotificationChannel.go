package Services

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailNotificationChannel 邮件通知通道
type EmailNotificationChannel struct {
	BaseService
	storageManager *Storage.StorageManager
	emailService   *EmailService
	enabled        bool
	recipients     []string
	template       *template.Template
}

// NewEmailNotificationChannel 创建邮件通知通道
func NewEmailNotificationChannel(storageManager *Storage.StorageManager, emailService *EmailService, recipients []string) *EmailNotificationChannel {
	// 创建邮件模板
	tmpl := template.Must(template.New("alert_email").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>系统告警通知</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background-color: {{.SeverityColor}}; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { padding: 20px; }
        .alert-details { background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin: 15px 0; }
        .metadata { background-color: #e9ecef; padding: 10px; border-radius: 4px; font-family: monospace; font-size: 12px; }
        .footer { background-color: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; color: #6c757d; }
        .severity-critical { background-color: #dc3545; }
        .severity-high { background-color: #fd7e14; }
        .severity-medium { background-color: #ffc107; color: #212529; }
        .severity-low { background-color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header severity-{{.Severity}}">
            <h1>{{.Title}}</h1>
            <p>告警时间: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        </div>
        <div class="content">
            <h2>告警详情</h2>
            <div class="alert-details">
                <p><strong>告警类型:</strong> {{.Type}}</p>
                <p><strong>严重程度:</strong> {{.Severity}}</p>
                <p><strong>来源:</strong> {{.Source}}</p>
                <p><strong>消息:</strong> {{.Message}}</p>
            </div>
            
            {{if .Metadata}}
            <h3>元数据</h3>
            <div class="metadata">
                {{range $key, $value := .Metadata}}
                <div><strong>{{$key}}:</strong> {{$value}}</div>
                {{end}}
            </div>
            {{end}}
            
            <h3>建议操作</h3>
            <ul>
                <li>检查系统日志获取更多信息</li>
                <li>联系系统管理员</li>
                <li>查看监控仪表板了解系统状态</li>
            </ul>
        </div>
        <div class="footer">
            <p>此邮件由云平台监控系统自动发送</p>
            <p>告警ID: {{.ID}}</p>
        </div>
    </div>
</body>
</html>
`))

	return &EmailNotificationChannel{
		storageManager: storageManager,
		emailService:   emailService,
		enabled:        true,
		recipients:     recipients,
		template:       tmpl,
	}
}

// SendAlert 发送告警邮件
func (enc *EmailNotificationChannel) SendAlert(alert MonitoringAlert) error {
	if !enc.enabled {
		return fmt.Errorf("邮件通知通道已禁用")
	}

	// 准备邮件数据
	emailData := struct {
		MonitoringAlert
		SeverityColor string
	}{
		MonitoringAlert: alert,
		SeverityColor:   enc.getSeverityColor(alert.Severity),
	}

	// 渲染邮件模板
	var htmlContent string
	var buf strings.Builder
	if err := enc.template.Execute(&buf, emailData); err != nil {
		return fmt.Errorf("渲染邮件模板失败: %v", err)
	}
	htmlContent = buf.String()

	// 发送邮件
	subject := fmt.Sprintf("[%s] %s", alert.Severity, alert.Title)

	for _, recipient := range enc.recipients {
		if err := enc.emailService.sendEmail(recipient, subject, htmlContent, "text/html"); err != nil {
			enc.storageManager.LogError("发送告警邮件失败", map[string]interface{}{
				"recipient": recipient,
				"alert_id":  alert.ID,
				"error":     err.Error(),
			})
			return err
		}
	}

	enc.storageManager.LogInfo("告警邮件已发送", map[string]interface{}{
		"alert_id":   alert.ID,
		"recipients": enc.recipients,
		"severity":   alert.Severity,
		"timestamp":  time.Now(),
	})

	return nil
}

// getSeverityColor 获取严重程度对应的颜色
func (enc *EmailNotificationChannel) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "low"
	}
}

// GetName 获取通道名称
func (enc *EmailNotificationChannel) GetName() string {
	return "email"
}

// IsEnabled 检查是否启用
func (enc *EmailNotificationChannel) IsEnabled() bool {
	return enc.enabled
}

// SetEnabled 设置启用状态
func (enc *EmailNotificationChannel) SetEnabled(enabled bool) {
	enc.enabled = enabled
	enc.storageManager.LogInfo("邮件通知通道状态已更新", map[string]interface{}{
		"enabled":   enabled,
		"timestamp": time.Now(),
	})
}

// AddRecipient 添加收件人
func (enc *EmailNotificationChannel) AddRecipient(email string) {
	enc.recipients = append(enc.recipients, email)
	enc.storageManager.LogInfo("邮件收件人已添加", map[string]interface{}{
		"email":     email,
		"timestamp": time.Now(),
	})
}

// RemoveRecipient 移除收件人
func (enc *EmailNotificationChannel) RemoveRecipient(email string) {
	for i, recipient := range enc.recipients {
		if recipient == email {
			enc.recipients = append(enc.recipients[:i], enc.recipients[i+1:]...)
			break
		}
	}
	enc.storageManager.LogInfo("邮件收件人已移除", map[string]interface{}{
		"email":     email,
		"timestamp": time.Now(),
	})
}

// GetRecipients 获取收件人列表
func (enc *EmailNotificationChannel) GetRecipients() []string {
	return enc.recipients
}

// SetRecipients 设置收件人列表
func (enc *EmailNotificationChannel) SetRecipients(recipients []string) {
	enc.recipients = recipients
	enc.storageManager.LogInfo("邮件收件人列表已更新", map[string]interface{}{
		"recipients": recipients,
		"timestamp":  time.Now(),
	})
}

// TestConnection 测试邮件连接
func (enc *EmailNotificationChannel) TestConnection() error {
	if !enc.enabled {
		return fmt.Errorf("邮件通知通道已禁用")
	}

	// 发送测试邮件
	testAlert := MonitoringAlert{
		ID:        fmt.Sprintf("test_%d", time.Now().Unix()),
		Type:      "test",
		Severity:  "low",
		Title:     "邮件通知测试",
		Message:   "这是一封测试邮件，用于验证邮件通知通道是否正常工作。",
		Source:    "system_test",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	return enc.SendAlert(testAlert)
}

// GetStatus 获取通道状态
func (enc *EmailNotificationChannel) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"name":       enc.GetName(),
		"enabled":    enc.enabled,
		"recipients": enc.recipients,
		"timestamp":  time.Now(),
	}
}
