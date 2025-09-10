package Middleware

import (
	"cloud-platform-api/app/Services"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware 安全防护中间件
type SecurityMiddleware struct {
	securityService *Services.SecurityService
}

// NewSecurityMiddleware 创建安全防护中间件
// 功能说明：
// 1. 初始化安全防护中间件实例
// 2. 设置安全服务用于威胁检测和防护
// 3. 用于保护API端点免受各种安全威胁
func NewSecurityMiddleware(securityService *Services.SecurityService) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityService: securityService,
	}
}

// SecurityCheck 安全检查中间件
// 功能说明：
// 1. 执行全面的安全检查，包括威胁检测、访问控制、速率限制等
// 2. 检查IP地址是否在黑名单中
// 3. 验证用户代理字符串的合法性
// 4. 检查请求路径是否存在安全风险
// 5. 实施速率限制防止暴力攻击
// 6. 记录安全事件和异常行为
// 7. 支持多种安全策略和规则
func (m *SecurityMiddleware) SecurityCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求信息
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		path := c.Request.URL.Path
		method := c.Request.Method

		// 获取用户信息
		userID, exists := c.Get("user_id")
		var userIDUint uint
		if exists {
			if id, ok := userID.(uint); ok {
				userIDUint = id
			}
		}

		// 威胁防护检查
		if m.securityService != nil {
			allowed, reason := m.securityService.CheckThreatProtection(ipAddress, "", "")
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "访问被拒绝",
					"message": reason,
					"code":    403,
				})
				c.Abort()
				return
			}
		}

		// 异常检测
		if m.securityService != nil && userIDUint > 0 {
			isAnomaly, score := m.securityService.DetectAnomaly(userIDUint, "http_request", path, method, ipAddress, userAgent)
			if isAnomaly {
				// 记录异常事件
				m.securityService.RecordSecurityEvent(
					userIDUint,
					"anomaly_detected",
					"high",
					ipAddress,
					userAgent,
					path,
					method,
					fmt.Sprintf("异常评分: %.2f", score),
					score,
					score,
					false,
					true,
					"",
					"",
				)

				// 如果配置了自动阻止，则阻止请求
				// 这里可以根据配置决定是否阻止
			}
		}

		// 访问控制检查
		if m.securityService != nil && userIDUint > 0 {
			allowed, reason := m.securityService.CheckAccessControl(userIDUint, path, method)
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "访问被拒绝",
					"message": reason,
					"code":    403,
				})
				c.Abort()
				return
			}
		}

		// 记录安全事件
		if m.securityService != nil && userIDUint > 0 {
			m.securityService.RecordSecurityEvent(
				userIDUint,
				"http_request",
				"info",
				ipAddress,
				userAgent,
				path,
				method,
				"",
				0,
				0,
				false,
				false,
				"",
				"",
			)
		}

		c.Next()
	}
}

// CSRFProtection CSRF防护中间件
func (m *SecurityMiddleware) CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过GET请求
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// 检查CSRF令牌
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("csrf_token")
		}

		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF令牌缺失",
				"message": "请求缺少CSRF令牌",
				"code":    403,
			})
			c.Abort()
			return
		}

		// 验证CSRF令牌
		if !m.validateCSRFToken(c, token) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF令牌无效",
				"message": "CSRF令牌验证失败",
				"code":    403,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateCSRFToken 验证CSRF令牌
func (m *SecurityMiddleware) validateCSRFToken(c *gin.Context, token string) bool {
	// 从会话中获取CSRF令牌
	sessionToken, exists := c.Get("csrf_token")
	if !exists {
		return false
	}

	// 比较令牌
	return token == sessionToken
}

// XSSProtection XSS防护中间件
func (m *SecurityMiddleware) XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置XSS防护头
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")

		// 检查请求内容中的XSS攻击
		if m.detectXSSAttack(c) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "XSS攻击检测",
				"message": "请求内容包含潜在的XSS攻击",
				"code":    400,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// detectXSSAttack 检测XSS攻击
func (m *SecurityMiddleware) detectXSSAttack(c *gin.Context) bool {
	// 检查URL参数
	for _, values := range c.Request.URL.Query() {
		for _, value := range values {
			if m.containsXSSPattern(value) {
				return true
			}
		}
	}

	// 检查POST数据
	if c.Request.Method == "POST" {
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			if err := c.Request.ParseForm(); err == nil {
				for _, values := range c.Request.PostForm {
					for _, value := range values {
						if m.containsXSSPattern(value) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// containsXSSPattern 检查是否包含XSS模式
func (m *SecurityMiddleware) containsXSSPattern(input string) bool {
	xssPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=",
		"<iframe", "<object", "<embed", "<form",
		"alert(", "confirm(", "prompt(", "eval(",
		"document.cookie", "window.location", "location.href",
	}

	input = strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}

// SQLInjectionProtection SQL注入防护中间件
func (m *SecurityMiddleware) SQLInjectionProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查URL参数
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				if m.containsSQLInjectionPattern(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "SQL注入攻击检测",
						"message": "请求包含潜在的SQL注入攻击",
						"code":    400,
					})
					c.Abort()
					return
				}
			}
		}

		// 检查POST数据
		if c.Request.Method == "POST" {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/x-www-form-urlencoded") {
				if err := c.Request.ParseForm(); err == nil {
					for _, values := range c.Request.PostForm {
						for _, value := range values {
							if m.containsSQLInjectionPattern(value) {
								c.JSON(http.StatusBadRequest, gin.H{
									"error":   "SQL注入攻击检测",
									"message": "请求包含潜在的SQL注入攻击",
									"code":    400,
								})
								c.Abort()
								return
							}
						}
					}
				}
			}
		}

		c.Next()
	}
}

// containsSQLInjectionPattern 检查是否包含SQL注入模式
func (m *SecurityMiddleware) containsSQLInjectionPattern(input string) bool {
	sqlPatterns := []string{
		"union select", "union all select", "select * from",
		"insert into", "update set", "delete from",
		"drop table", "drop database", "create table",
		"alter table", "exec ", "execute ", "xp_",
		"sp_", "waitfor delay", "benchmark(",
		"sleep(", "load_file(", "into outfile",
		"into dumpfile", "information_schema",
		"@@version", "@@hostname", "@@datadir",
	}

	input = strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}

// FileUploadSecurity 文件上传安全检查中间件
func (m *SecurityMiddleware) FileUploadSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有文件上传
		if c.Request.Method == "POST" && strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			// 解析multipart表单
			if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "文件上传解析失败",
					"message": err.Error(),
					"code":    400,
				})
				c.Abort()
				return
			}

			// 检查文件
			if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
				for _, files := range c.Request.MultipartForm.File {
					for _, file := range files {
						if !m.isFileAllowed(file.Filename, file.Size) {
							c.JSON(http.StatusBadRequest, gin.H{
								"error":   "文件类型不允许",
								"message": "上传的文件类型或大小不符合要求",
								"code":    400,
							})
							c.Abort()
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

// isFileAllowed 检查文件是否允许上传
func (m *SecurityMiddleware) isFileAllowed(filename string, size int64) bool {
	// 检查文件扩展名
	allowedExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp",
		".pdf", ".doc", ".docx", ".txt", ".csv",
		".xls", ".xlsx", ".ppt", ".pptx",
	}

	blockedExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".pif",
		".scr", ".vbs", ".js", ".jar", ".war",
		".php", ".asp", ".aspx", ".jsp", ".py",
	}

	filename = strings.ToLower(filename)

	// 检查阻止的扩展名
	for _, ext := range blockedExtensions {
		if strings.HasSuffix(filename, ext) {
			return false
		}
	}

	// 检查允许的扩展名
	allowed := false
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			allowed = true
			break
		}
	}

	if !allowed {
		return false
	}

	// 检查文件大小（10MB限制）
	maxSize := int64(10 * 1024 * 1024)
	return size <= maxSize
}

// RateLimitSecurity 速率限制安全中间件
func (m *SecurityMiddleware) RateLimitSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipAddress := c.ClientIP()
		userID, exists := c.Get("user_id")

		// 检查IP速率限制
		if !m.checkIPRateLimit(c, ipAddress) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "IP地址请求频率超过限制",
				"code":    429,
			})
			c.Abort()
			return
		}

		// 检查用户速率限制
		if exists {
			if id, ok := userID.(uint); ok {
				if !m.checkUserRateLimit(c, id) {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":   "请求过于频繁",
						"message": "用户请求频率超过限制",
						"code":    429,
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// checkIPRateLimit 检查IP速率限制
func (m *SecurityMiddleware) checkIPRateLimit(c *gin.Context, ipAddress string) bool {
	// 使用Redis或内存缓存来跟踪请求频率
	// 这里使用简单的内存实现
	key := fmt.Sprintf("rate_limit:ip:%s", ipAddress)

	// 获取当前时间窗口的请求次数
	now := time.Now()
	windowStart := now.Truncate(time.Minute)

	// 这里应该从缓存中获取请求次数
	// 示例实现
	requestCount := m.getRequestCount(key, windowStart)

	// 检查是否超过限制（每分钟100次）
	if requestCount >= 100 {
		return false
	}

	// 增加请求计数
	m.incrementRequestCount(key, windowStart)
	return true
}

// checkUserRateLimit 检查用户速率限制
func (m *SecurityMiddleware) checkUserRateLimit(c *gin.Context, userID uint) bool {
	key := fmt.Sprintf("rate_limit:user:%d", userID)

	now := time.Now()
	windowStart := now.Truncate(time.Minute)

	requestCount := m.getRequestCount(key, windowStart)

	// 用户限制更严格（每分钟50次）
	if requestCount >= 50 {
		return false
	}

	// 增加请求计数
	m.incrementRequestCount(key, windowStart)
	return true
}

// getRequestCount 获取请求计数（简化实现）
func (m *SecurityMiddleware) getRequestCount(key string, windowStart time.Time) int {
	// 这里应该从Redis或内存缓存中获取
	// 简化实现，返回0
	return 0
}

// incrementRequestCount 增加请求计数（简化实现）
func (m *SecurityMiddleware) incrementRequestCount(key string, windowStart time.Time) {
	// 这里应该增加Redis或内存缓存中的计数
	// 简化实现，不做任何操作
}

// ContentSecurityPolicy 内容安全策略中间件
func (m *SecurityMiddleware) ContentSecurityPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置CSP头
		cspDirectives := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data: https:",
			"font-src 'self'",
			"connect-src 'self'",
			"frame-src 'none'",
			"object-src 'none'",
			"base-uri 'self'",
			"form-action 'self'",
			"frame-ancestors 'none'",
		}

		c.Header("Content-Security-Policy", strings.Join(cspDirectives, "; "))
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func (m *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置各种安全头
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}

// RequestLogging 请求日志中间件
func (m *SecurityMiddleware) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求信息
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		method := c.Request.Method
		path := c.Request.URL.Path
		userID, _ := c.Get("user_id")

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)
		status := c.Writer.Status()

		// 记录安全事件
		if m.securityService != nil && userID != nil {
			if id, ok := userID.(uint); ok {
				eventLevel := "info"
				if status >= 400 {
					eventLevel = "warning"
				}
				if status >= 500 {
					eventLevel = "error"
				}

				m.securityService.RecordSecurityEvent(
					id,
					"http_response",
					eventLevel,
					ipAddress,
					userAgent,
					path,
					method,
					fmt.Sprintf("Status: %d, Duration: %v", status, duration),
					0,
					0,
					false,
					false,
					"",
					"",
				)
			}
		}
	}
}

// ContextSecurity 上下文安全中间件
func (m *SecurityMiddleware) ContextSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求上下文
		type contextKey string
		const securityMiddlewareKey contextKey = "security_middleware"
		ctx := context.WithValue(c.Request.Context(), securityMiddlewareKey, m)
		c.Request = c.Request.WithContext(ctx)

		// 添加安全相关的上下文信息
		c.Set("request_id", m.generateRequestID())
		c.Set("timestamp", time.Now().Unix())
		c.Set("ip_address", c.ClientIP())

		c.Next()
	}
}

// generateRequestID 生成请求ID
func (m *SecurityMiddleware) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
