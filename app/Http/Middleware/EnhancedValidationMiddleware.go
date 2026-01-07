package Middleware

import (
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedValidationMiddleware 增强的验证中间件
type EnhancedValidationMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	config         *ValidationConfig
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	MaxRequestSize      int64         `json:"max_request_size"`      // 最大请求体大小（字节）
	MaxFileSize         int64         `json:"max_file_size"`         // 最大文件大小（字节）
	AllowedContentTypes []string      `json:"allowed_content_types"` // 允许的Content-Type
	MaxStringLength     int           `json:"max_string_length"`     // 最大字符串长度
	MaxArrayLength      int           `json:"max_array_length"`      // 最大数组长度
	MaxObjectDepth      int           `json:"max_object_depth"`      // 最大对象深度
	EnableSQLInjection  bool          `json:"enable_sql_injection"`  // 启用SQL注入检测
	EnableXSSProtection bool          `json:"enable_xss_protection"` // 启用XSS防护
	RateLimitPerMinute  int           `json:"rate_limit_per_minute"` // 每分钟请求限制
	RateLimitWindow     time.Duration `json:"rate_limit_window"`     // 速率限制窗口
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewEnhancedValidationMiddleware 创建增强的验证中间件
func NewEnhancedValidationMiddleware(storageManager *Storage.StorageManager, config *ValidationConfig) *EnhancedValidationMiddleware {
	if config == nil {
		config = &ValidationConfig{
			MaxRequestSize:      10 * 1024 * 1024, // 10MB
			MaxFileSize:         5 * 1024 * 1024,  // 5MB
			AllowedContentTypes: []string{"application/json", "multipart/form-data", "application/x-www-form-urlencoded"},
			MaxStringLength:     10000,
			MaxArrayLength:      1000,
			MaxObjectDepth:      10,
			EnableSQLInjection:  true,
			EnableXSSProtection: true,
			RateLimitPerMinute:  100,
			RateLimitWindow:     time.Minute,
		}
	}

	return &EnhancedValidationMiddleware{
		storageManager: storageManager,
		config:         config,
	}
}

// Handle 处理验证
//
// 功能说明：
// 1. 多层安全验证，防止各种攻击
// 2. 按顺序执行验证步骤，任何一步失败都会中止请求
// 3. 记录验证失败的安全日志
// 4. 返回详细的错误信息
//
// 验证步骤（按顺序执行）：
// 1. 请求体大小验证：防止DoS攻击（大文件上传）
// 2. Content-Type验证：确保请求格式正确
// 3. 请求参数验证：验证参数格式和类型
// 4. SQL注入检测：检测SQL注入攻击尝试
// 5. XSS防护：检测跨站脚本攻击尝试
// 6. 速率限制：防止请求频率过高
//
// 安全特性：
// - 多层防护：多个验证步骤，提高安全性
// - 可配置：每个验证步骤都可以通过配置启用/禁用
// - 详细日志：记录所有验证失败事件，用于安全审计
//
// 错误处理：
// - 每个验证步骤失败都会立即返回错误
// - 错误信息包含错误代码、消息和详细信息
// - 记录安全日志，用于问题排查和安全审计
//
// 性能考虑：
// - 验证步骤按顺序执行，早期失败可以快速返回
// - 某些验证步骤（如SQL注入检测）可能较慢
// - 可以根据实际情况调整验证顺序
//
// 注意事项：
// - 验证顺序很重要，应该先验证基础信息，再验证安全相关
// - 验证失败时调用c.Abort()停止后续处理
// - 安全日志应该包含足够的上下文信息
func (m *EnhancedValidationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 验证请求体大小
		// 防止DoS攻击：大文件上传会消耗服务器资源
		// 默认限制：10MB
		if err := m.validateRequestSize(c); err != nil {
			m.handleValidationError(c, "REQUEST_SIZE", "请求体过大", err.Error())
			return
		}

		// 2. 验证Content-Type
		// 确保请求格式正确，防止格式错误导致的问题
		// 允许的类型：application/json、multipart/form-data等
		if err := m.validateContentType(c); err != nil {
			m.handleValidationError(c, "CONTENT_TYPE", "不支持的Content-Type", err.Error())
			return
		}

		// 3. 验证请求参数
		// 验证参数格式、类型、长度等
		// 防止格式错误和类型不匹配
		if err := m.validateRequestParams(c); err != nil {
			m.handleValidationError(c, "PARAMS", "请求参数无效", err.Error())
			return
		}

		// 4. SQL注入检测
		// 检测SQL注入攻击尝试
		// 可以通过配置启用/禁用
		if m.config.EnableSQLInjection {
			if err := m.detectSQLInjection(c); err != nil {
				m.handleValidationError(c, "SQL_INJECTION", "检测到SQL注入尝试", err.Error())
				return
			}
		}

		// 5. XSS防护
		// 检测跨站脚本攻击尝试
		// 可以通过配置启用/禁用
		if m.config.EnableXSSProtection {
			if err := m.detectXSS(c); err != nil {
				m.handleValidationError(c, "XSS", "检测到XSS攻击尝试", err.Error())
				return
			}
		}

		// 6. 速率限制
		// 防止请求频率过高
		// 可以通过配置启用/禁用
		if err := m.checkRateLimit(c); err != nil {
			m.handleValidationError(c, "RATE_LIMIT", "请求频率超限", err.Error())
			return
		}

		// 所有验证通过，继续执行后续中间件和处理器
		c.Next()
	}
}

// validateRequestSize 验证请求体大小
func (m *EnhancedValidationMiddleware) validateRequestSize(c *gin.Context) error {
	if c.Request.ContentLength > m.config.MaxRequestSize {
		return fmt.Errorf("请求体大小 %d 字节超过限制 %d 字节", c.Request.ContentLength, m.config.MaxRequestSize)
	}
	return nil
}

// validateContentType 验证Content-Type
func (m *EnhancedValidationMiddleware) validateContentType(c *gin.Context) error {
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		return nil // 允许没有Content-Type的请求
	}

	// 检查是否在允许的Content-Type列表中
	for _, allowedType := range m.config.AllowedContentTypes {
		if strings.Contains(contentType, allowedType) {
			return nil
		}
	}

	return fmt.Errorf("不支持的Content-Type: %s", contentType)
}

// validateRequestParams 验证请求参数
func (m *EnhancedValidationMiddleware) validateRequestParams(c *gin.Context) error {
	// 验证查询参数
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if len(value) > m.config.MaxStringLength {
				return fmt.Errorf("参数 %s 的值过长", key)
			}
		}
	}

	// 验证路径参数
	for _, param := range c.Params {
		if len(param.Value) > m.config.MaxStringLength {
			return fmt.Errorf("路径参数 %s 的值过长", param.Key)
		}
	}

	return nil
}

// detectSQLInjection 检测SQL注入
func (m *EnhancedValidationMiddleware) detectSQLInjection(c *gin.Context) error {
	sqlKeywords := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"UNION", "SELECT", "INSERT", "UPDATE", "DELETE",
		"DROP", "CREATE", "ALTER", "EXEC", "EXECUTE",
		"SCRIPT", "VBSCRIPT", "JAVASCRIPT", "ONLOAD",
		"ONERROR", "ONCLICK", "ONMOUSEOVER",
	}

	// 检查查询参数
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			upperValue := strings.ToUpper(value)
			for _, keyword := range sqlKeywords {
				if strings.Contains(upperValue, keyword) {
					return fmt.Errorf("参数 %s 包含可疑的SQL关键字: %s", key, keyword)
				}
			}
		}
	}

	// 检查路径参数
	for _, param := range c.Params {
		upperValue := strings.ToUpper(param.Value)
		for _, keyword := range sqlKeywords {
			if strings.Contains(upperValue, keyword) {
				return fmt.Errorf("路径参数 %s 包含可疑的SQL关键字: %s", param.Key, keyword)
			}
		}
	}

	return nil
}

// detectXSS 检测XSS攻击
func (m *EnhancedValidationMiddleware) detectXSS(c *gin.Context) error {
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"<iframe", "</iframe>", "<object", "</object>",
		"<embed", "</embed>", "<link", "<meta",
		"<style", "</style>", "expression(",
	}

	// 检查查询参数
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			lowerValue := strings.ToLower(value)
			for _, pattern := range xssPatterns {
				if strings.Contains(lowerValue, pattern) {
					return fmt.Errorf("参数 %s 包含可疑的XSS模式: %s", key, pattern)
				}
			}
		}
	}

	// 检查路径参数
	for _, param := range c.Params {
		lowerValue := strings.ToLower(param.Value)
		for _, pattern := range xssPatterns {
			if strings.Contains(lowerValue, pattern) {
				return fmt.Errorf("路径参数 %s 包含可疑的XSS模式: %s", param.Key, pattern)
			}
		}
	}

	return nil
}

// checkRateLimit 检查速率限制
func (m *EnhancedValidationMiddleware) checkRateLimit(c *gin.Context) error {
	// 这里应该实现基于IP或用户的速率限制
	// 暂时返回nil，表示通过检查
	return nil
}

// handleValidationError 处理验证错误
//
// 功能说明：
// 1. 记录验证失败的安全日志
// 2. 返回统一的错误响应格式
// 3. 中止后续处理，防止不安全请求继续执行
//
// 日志记录：
// - 记录验证失败的所有相关信息
// - 包括错误代码、消息、详细信息
// - 包括客户端IP、User-Agent、URL、方法等
// - 用于安全审计和问题排查
//
// 错误响应：
// - 统一的错误响应格式
// - 包含错误代码，便于前端处理
// - 包含错误消息和详细信息
// - 返回400 Bad Request状态码
//
// 安全考虑：
// - 记录详细的安全日志，用于威胁分析
// - 不返回敏感信息，避免信息泄露
// - 中止后续处理，防止不安全请求继续执行
//
// 注意事项：
// - 必须调用c.Abort()停止后续处理
// - 日志应该包含足够的上下文信息
// - 错误消息应该用户友好，但不要泄露系统信息
func (m *EnhancedValidationMiddleware) handleValidationError(c *gin.Context, code, message, details string) {
	// 记录安全日志
	// 用于安全审计和问题排查
	// 包含所有相关信息，便于分析攻击模式
	m.storageManager.LogWarning("请求验证失败", map[string]interface{}{
		"code":       code,                      // 错误代码（如SQL_INJECTION、XSS等）
		"message":    message,                   // 错误消息
		"details":    details,                   // 详细信息
		"client_ip":  c.ClientIP(),              // 客户端IP（用于追踪攻击来源）
		"user_agent": c.Request.UserAgent(),     // User-Agent（用于识别客户端）
		"url":        c.Request.URL.String(),    // 请求URL
		"method":     c.Request.Method,          // HTTP方法
		"timestamp":  time.Now(),                 // 时间戳
	})

	// 返回错误响应
	// 统一的错误响应格式，便于前端处理
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,  // 请求失败
		"message": message, // 用户友好的错误消息
		"code":    code,    // 错误代码（用于前端错误处理）
		"error":   details, // 详细错误信息
	})
	
	// 中止后续处理，防止不安全请求继续执行
	c.Abort()
}

// ValidateJSON 验证JSON请求体
func (m *EnhancedValidationMiddleware) ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// 读取请求体
		body, err := c.GetRawData()
		if err != nil {
			m.handleValidationError(c, "READ_BODY", "读取请求体失败", err.Error())
			return
		}

		// 验证JSON格式
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			m.handleValidationError(c, "INVALID_JSON", "JSON格式无效", err.Error())
			return
		}

		// 验证JSON深度
		if err := m.validateJSONDepth(jsonData, 0); err != nil {
			m.handleValidationError(c, "JSON_DEPTH", "JSON对象深度超限", err.Error())
			return
		}

		// 将请求体重新设置到上下文中
		c.Request.Body = &bodyReader{data: body}

		c.Next()
	}
}

// validateJSONDepth 验证JSON深度
func (m *EnhancedValidationMiddleware) validateJSONDepth(data interface{}, depth int) error {
	if depth > m.config.MaxObjectDepth {
		return fmt.Errorf("JSON对象深度 %d 超过限制 %d", depth, m.config.MaxObjectDepth)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			if err := m.validateJSONDepth(value, depth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(v) > m.config.MaxArrayLength {
			return fmt.Errorf("数组长度 %d 超过限制 %d", len(v), m.config.MaxArrayLength)
		}
		for _, value := range v {
			if err := m.validateJSONDepth(value, depth+1); err != nil {
				return err
			}
		}
	case string:
		if len(v) > m.config.MaxStringLength {
			return fmt.Errorf("字符串长度 %d 超过限制 %d", len(v), m.config.MaxStringLength)
		}
	}

	return nil
}

// bodyReader 实现io.ReadCloser接口
type bodyReader struct {
	data []byte
	pos  int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bodyReader) Close() error {
	return nil
}

// ValidateFileUpload 验证文件上传
func (m *EnhancedValidationMiddleware) ValidateFileUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有文件上传
		form, err := c.MultipartForm()
		if err != nil {
			c.Next()
			return
		}

		// 验证文件大小
		for _, files := range form.File {
			for _, file := range files {
				if file.Size > m.config.MaxFileSize {
					m.handleValidationError(c, "FILE_SIZE", "文件大小超限",
						fmt.Sprintf("文件 %s 大小 %d 字节超过限制 %d 字节", file.Filename, file.Size, m.config.MaxFileSize))
					return
				}

				// 验证文件类型
				if err := m.validateFileType(file.Filename); err != nil {
					m.handleValidationError(c, "FILE_TYPE", "文件类型不支持", err.Error())
					return
				}
			}
		}

		c.Next()
	}
}

// validateFileType 验证文件类型
func (m *EnhancedValidationMiddleware) validateFileType(filename string) error {
	allowedExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx",
		".txt", ".csv", ".xlsx", ".zip", ".rar",
	}

	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("文件类型 %s 不在允许列表中", ext)
}

// GetConfig 获取配置
func (m *EnhancedValidationMiddleware) GetConfig() *ValidationConfig {
	return m.config
}

// UpdateConfig 更新配置
func (m *EnhancedValidationMiddleware) UpdateConfig(config *ValidationConfig) {
	m.config = config
}
