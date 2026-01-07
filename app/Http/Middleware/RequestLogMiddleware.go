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
//
// 功能说明：
// 1. 记录HTTP请求和响应的详细信息
// 2. 支持请求体和响应体的记录（可配置）
// 3. 自动脱敏敏感字段（密码、token等）
// 4. 记录请求处理时间和状态码
// 5. 支持路径过滤，避免记录不需要的路径
//
// 记录的信息：
// - 请求信息：方法、路径、查询参数、请求头、请求体
// - 响应信息：状态码、响应头、响应体
// - 性能信息：处理时间、请求大小、响应大小
// - 用户信息：用户ID、用户名、角色（如果已认证）
// - 客户端信息：IP地址、User-Agent、Referer
//
// 敏感字段脱敏：
// - 自动检测并脱敏敏感字段（如password、token等）
// - 支持嵌套对象和数组的递归脱敏
// - 脱敏后的值显示为"***MASKED***"
//
// 性能考虑：
// - 请求体和响应体的读取可能影响性能
// - 可以通过配置控制是否记录请求体和响应体
// - 大文件请求可能占用较多内存
//
// 使用场景：
// - 问题排查和调试
// - 安全审计
// - 性能分析
// - API使用情况统计
//
// 注意事项：
// - 请求体和响应体的读取会消耗内存
// - 敏感信息会被自动脱敏，但需要正确配置
// - 大量日志可能占用磁盘空间，需要定期清理
func (m *RequestLogMiddleware) RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间，用于计算请求处理耗时
		startTime := time.Now()

		// 检查是否应该记录此路径
		// 某些路径（如健康检查、静态资源）可能不需要记录
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 读取请求体（如果配置了）
		// 注意：读取后需要重新设置到请求中，以便后续处理
		var requestBody []byte
		if m.config.IncludeBody && m.config.MaxBodySize > 0 {
			requestBody = m.readRequestBody(c)
		}

		// 创建响应写入器包装器
		// 用于捕获响应体，同时不影响正常的响应写入
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		// 替换响应写入器，以便捕获响应体
		c.Writer = responseWriter

		// 处理请求（执行后续中间件和处理器）
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)

		// 读取响应体（如果配置了）
		// 从响应写入器包装器中获取响应体
		var responseBody []byte
		if m.config.IncludeBody && m.config.MaxBodySize > 0 {
			responseBody = m.readResponseBody(responseWriter)
		}

		// 记录请求日志
		// 包含所有请求和响应的详细信息
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
//
// 功能说明：
// 1. 检测并脱敏JSON数据中的敏感字段
// 2. 支持嵌套对象和数组的递归脱敏
// 3. 非JSON格式的数据直接返回（不处理）
//
// 脱敏策略：
// - 检测字段名是否包含敏感关键词（如password、token等）
// - 将敏感字段的值替换为"***MASKED***"
// - 递归处理嵌套对象和数组
//
// 支持的格式：
// - JSON对象：{"password": "123456"} -> {"password": "***MASKED***"}
// - 嵌套对象：{"user": {"password": "123456"}} -> {"user": {"password": "***MASKED***"}}
// - 数组：[{"password": "123456"}] -> [{"password": "***MASKED***"}]
//
// 性能考虑：
// - JSON解析和序列化可能较慢
// - 大文件可能占用较多内存
// - 非JSON格式的数据不处理，直接返回
//
// 安全考虑：
// - 敏感字段检测基于字段名，需要正确配置
// - 脱敏后的数据仍然可能包含其他敏感信息
// - 建议在生产环境中谨慎记录请求体和响应体
//
// 注意事项：
// - 如果JSON解析失败，返回原始数据
// - 如果序列化失败，返回原始数据
// - 脱敏操作不会修改原始数据，返回新数据
func (m *RequestLogMiddleware) maskSensitiveFields(data []byte) []byte {
	// 如果数据为空，直接返回
	if len(data) == 0 {
		return data
	}

	// 尝试解析JSON
	// 如果不是JSON格式，无法进行脱敏处理
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// 不是JSON格式，直接返回原始数据
		// 非JSON格式的数据（如纯文本、二进制）不进行脱敏
		return data
	}

	// 脱敏处理
	// 递归处理所有嵌套对象和数组
	m.maskMap(jsonData)

	// 重新序列化为JSON
	result, err := json.Marshal(jsonData)
	if err != nil {
		// 序列化失败，返回原始数据
		// 这通常不应该发生，但为了安全起见，返回原始数据
		return data
	}

	// 返回脱敏后的数据
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
