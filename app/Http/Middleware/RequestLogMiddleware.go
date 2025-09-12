package Middleware

import (
	"bytes"
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"bufio"

	"github.com/gin-gonic/gin"
)

// RequestLogMiddleware 请求日志中间件
//
// 重要功能说明：
// 1. 记录HTTP请求和响应的详细信息
// 2. 支持请求体和响应体的记录（可配置）
// 3. 自动脱敏敏感字段（密码、token等）
// 4. 记录请求处理时间和状态码
// 5. 支持路径过滤和字段过滤
// 6. 异步日志记录，不影响请求性能
// 7. 可配置的日志格式和存储位置
type RequestLogMiddleware struct {
	BaseMiddleware
	logManager *Services.LogManagerService
	config     *Config.RequestLogConfig
}

// NewRequestLogMiddleware 创建请求日志中间件
func NewRequestLogMiddleware(logManager *Services.LogManagerService) *RequestLogMiddleware {
	return &RequestLogMiddleware{
		logManager: logManager,
		config:     &logManager.GetConfig().RequestLog,
	}
}

// RequestLog 请求日志中间件处理函数
func (m *RequestLogMiddleware) RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 检查是否应该记录此路径
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 读取请求体
		var requestBody []byte
		if m.config.IncludeBody && m.config.MaxBodySize > 0 {
			requestBody = m.readRequestBody(c)
		}

		// 创建响应写入器包装器
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)

		// 读取响应体
		var responseBody []byte
		if m.config.IncludeBody && m.config.MaxBodySize > 0 {
			responseBody = m.readResponseBody(responseWriter)
		}

		// 记录请求日志
		m.logRequest(c, startTime, duration, requestBody, responseBody)
	}
}

// shouldSkipPath 检查是否应该跳过记录此路径
func (m *RequestLogMiddleware) shouldSkipPath(path string) bool {
	for _, filterPath := range m.config.FilterPaths {
		if strings.HasPrefix(path, filterPath) {
			return true
		}
	}
	return false
}

// readRequestBody 读取请求体
func (m *RequestLogMiddleware) readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}

	// 限制读取大小
	limitedReader := io.LimitReader(c.Request.Body, int64(m.config.MaxBodySize*1024))
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil
	}

	// 重新设置请求体，以便后续处理
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 脱敏处理
	return m.maskSensitiveFields(body)
}

// readResponseBody 读取响应体
func (m *RequestLogMiddleware) readResponseBody(writer *responseBodyWriter) []byte {
	body := writer.body.Bytes()
	if len(body) > m.config.MaxBodySize*1024 {
		body = body[:m.config.MaxBodySize*1024]
	}
	return m.maskSensitiveFields(body)
}

// maskSensitiveFields 脱敏敏感字段
func (m *RequestLogMiddleware) maskSensitiveFields(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// 尝试解析JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// 不是JSON格式，直接返回
		return data
	}

	// 脱敏处理
	m.maskMap(jsonData)

	// 重新序列化
	result, err := json.Marshal(jsonData)
	if err != nil {
		return data
	}

	return result
}

// maskMap 递归脱敏map中的敏感字段
func (m *RequestLogMiddleware) maskMap(data map[string]interface{}) {
	for key, value := range data {
		// 检查是否是敏感字段
		if m.isSensitiveField(key) {
			data[key] = "***MASKED***"
			continue
		}

		// 递归处理嵌套map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			m.maskMap(nestedMap)
		}

		// 递归处理数组
		if nestedArray, ok := value.([]interface{}); ok {
			m.maskArray(nestedArray)
		}
	}
}

// maskArray 递归脱敏数组中的敏感字段
func (m *RequestLogMiddleware) maskArray(data []interface{}) {
	for _, item := range data {
		if nestedMap, ok := item.(map[string]interface{}); ok {
			m.maskMap(nestedMap)
		}
		if nestedArray, ok := item.([]interface{}); ok {
			m.maskArray(nestedArray)
		}
	}
}

// isSensitiveField 检查是否是敏感字段
func (m *RequestLogMiddleware) isSensitiveField(fieldName string) bool {
	fieldNameLower := strings.ToLower(fieldName)
	for _, maskField := range m.config.MaskFields {
		if strings.Contains(fieldNameLower, strings.ToLower(maskField)) {
			return true
		}
	}
	return false
}

// logRequest 记录请求日志
func (m *RequestLogMiddleware) logRequest(c *gin.Context, startTime time.Time, duration time.Duration, requestBody, responseBody []byte) {
	// 构建日志字段
	fields := map[string]interface{}{
		"method":         c.Request.Method,
		"path":           c.Request.URL.Path,
		"query":          c.Request.URL.RawQuery,
		"status_code":    c.Writer.Status(),
		"duration_ms":    duration.Milliseconds(),
		"duration_ns":    duration.Nanoseconds(),
		"client_ip":      c.ClientIP(),
		"user_agent":     c.Request.UserAgent(),
		"content_length": c.Request.ContentLength,
		"referer":        c.Request.Referer(),
		"protocol":       c.Request.Proto,
		"host":           c.Request.Host,
	}

	// 添加用户信息（如果可用）
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}
	if username, exists := c.Get("username"); exists {
		fields["username"] = username
	}
	if userRole, exists := c.Get("user_role"); exists {
		fields["user_role"] = userRole
	}

	// 添加请求头信息
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	fields["request_headers"] = headers

	// 添加响应头信息
	responseHeaders := make(map[string]string)
	for key, values := range c.Writer.Header() {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}
	fields["response_headers"] = responseHeaders

	// 添加请求体（如果配置了）
	if m.config.IncludeBody && len(requestBody) > 0 {
		fields["request_body"] = string(requestBody)
	}

	// 添加响应体（如果配置了）
	if m.config.IncludeBody && len(responseBody) > 0 {
		fields["response_body"] = string(responseBody)
	}

	// 添加错误信息（如果有）
	if len(c.Errors) > 0 {
		errors := make([]string, 0, len(c.Errors))
		for _, err := range c.Errors {
			errors = append(errors, err.Error())
		}
		fields["errors"] = errors
	}

	// 添加时间信息
	fields["start_time"] = startTime.Format(time.RFC3339Nano)
	fields["end_time"] = time.Now().Format(time.RFC3339Nano)

	// 记录日志
	m.logManager.LogRequest(c.Request.Context(), c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration, fields)
}

// buildLogMessage 构建日志消息
func (m *RequestLogMiddleware) buildLogMessage(c *gin.Context, duration time.Duration) string {
	statusCode := c.Writer.Status()
	method := c.Request.Method
	path := c.Request.URL.Path
	durationStr := duration.String()

	// 根据状态码选择消息格式
	switch {
	case statusCode >= 500:
		return fmt.Sprintf("HTTP %s %s %d - %s (ERROR)", method, path, statusCode, durationStr)
	case statusCode >= 400:
		return fmt.Sprintf("HTTP %s %s %d - %s (WARNING)", method, path, statusCode, durationStr)
	case statusCode >= 300:
		return fmt.Sprintf("HTTP %s %s %d - %s (REDIRECT)", method, path, statusCode, durationStr)
	default:
		return fmt.Sprintf("HTTP %s %s %d - %s", method, path, statusCode, durationStr)
	}
}

// responseBodyWriter 响应体写入器包装器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应体
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 写入响应字符串
func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// WriteHeader 写入响应头
func (w *responseBodyWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

// Status 获取响应状态码
func (w *responseBodyWriter) Status() int {
	return w.ResponseWriter.Status()
}

// Size 获取响应大小
func (w *responseBodyWriter) Size() int {
	return w.ResponseWriter.Size()
}

// WriteHeaderNow 立即写入响应头
func (w *responseBodyWriter) WriteHeaderNow() {
	w.ResponseWriter.WriteHeaderNow()
}

// CloseNotify 关闭通知
func (w *responseBodyWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.CloseNotify()
}

// Flush 刷新
func (w *responseBodyWriter) Flush() {
	w.ResponseWriter.Flush()
}

// Hijack 劫持连接
func (w *responseBodyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.Hijack()
}

// Pusher 推送器
func (w *responseBodyWriter) Pusher() http.Pusher {
	return w.ResponseWriter.Pusher()
}
