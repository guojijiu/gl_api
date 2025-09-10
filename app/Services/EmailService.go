package Services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

// EmailService 邮件服务
type EmailService struct {
	config *EmailConfig
}

// NewEmailService 创建邮件服务
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendNotificationEmail 发送通知邮件
func (s *EmailService) SendNotificationEmail(to, subject, body string) error {
	return s.sendEmail(to, subject, body, "text/html")
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *EmailService) SendPasswordResetEmail(to, resetToken, username string) error {
	subject := "密码重置请求"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.getBaseURL(), resetToken)
	
	body := fmt.Sprintf(`
		<h2>密码重置</h2>
		<p>您好 %s，</p>
		<p>请点击以下链接重置您的密码：</p>
		<p><a href="%s">重置密码</a></p>
		<p>此链接将在1小时后过期。</p>
	`, username, resetLink)

	return s.sendEmail(to, subject, body, "text/html")
}

// SendEmailVerificationEmail 发送邮箱验证邮件
func (s *EmailService) SendEmailVerificationEmail(to, verificationToken, username string) error {
	subject := "邮箱验证"
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.getBaseURL(), verificationToken)
	
	body := fmt.Sprintf(`
		<h2>邮箱验证</h2>
		<p>您好 %s，</p>
		<p>请点击以下链接验证您的邮箱：</p>
		<p><a href="%s">验证邮箱</a></p>
		<p>此链接将在24小时后过期。</p>
	`, username, verificationLink)

	return s.sendEmail(to, subject, body, "text/html")
}

// sendEmail 发送邮件
func (s *EmailService) sendEmail(to, subject, body, contentType string) error {
	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=UTF-8", contentType)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	var err error
	for i := 0; i < 3; i++ {
		if s.config.UseTLS {
			err = s.sendEmailWithTLS(to, addr, auth, message)
		} else {
			err = smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message))
		}

		if err == nil {
			break
		}

		if i < 2 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return err
}

// sendEmailWithTLS 使用TLS发送邮件
func (s *EmailService) sendEmailWithTLS(to, addr string, auth smtp.Auth, message string) error {
	conn, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = conn.StartTLS(&tls.Config{ServerName: s.config.Host}); err != nil {
		return err
	}

	if err = conn.Auth(auth); err != nil {
		return err
	}

	if err = conn.Mail(s.config.From); err != nil {
		return err
	}

	if err = conn.Rcpt(to); err != nil {
		return err
	}

	w, err := conn.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	return err
}

// getBaseURL 获取基础URL
func (s *EmailService) getBaseURL() string {
	return "http://localhost:8080"
}
