package Middleware

import (
	"cloud-platform-api/app/Services"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMonitoringMiddleware 性能监控中间件
type PerformanceMonitoringMiddleware struct {
	monitoringService *Services.PerformanceMonitoringService
	excludePaths      map[string]bool
}

// NewPerformanceMonitoringMiddleware 创建性能监控中间件
func NewPerformanceMonitoringMiddleware(service *Services.PerformanceMonitoringService, excludePaths []string) *PerformanceMonitoringMiddleware {
	excludeMap := make(map[string]bool)
	for _, path := range excludePaths {
		excludeMap[path] = true
	}
	
	return &PerformanceMonitoringMiddleware{
		monitoringService: service,
		excludePaths:      excludeMap,
	}
}

// Handler 性能监控中间件处理器
func (m *PerformanceMonitoringMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要排除此路径
		if m.shouldExcludePath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()
		
		// 设置监控上下文
		c.Set("monitor_start_time", startTime)
		c.Set("monitor_path", c.Request.URL.Path)
		c.Set("monitor_method", c.Request.Method)
		
		// 记录请求开始
		m.recordRequestStart(c)
		
		// 处理请求
		c.Next()
		
		// 记录请求结束
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		
		// 记录性能指标
		m.recordRequestMetrics(c, duration)
	}
}

// shouldExcludePath 检查是否应该排除此路径
func (m *PerformanceMonitoringMiddleware) shouldExcludePath(path string) bool {
	// 直接匹配
	if m.excludePaths[path] {
		return true
	}
	
	// 前缀匹配
	for excludePath := range m.excludePaths {
		if strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	
	return false
}

// recordRequestStart 记录请求开始
func (m *PerformanceMonitoringMiddleware) recordRequestStart(c *gin.Context) {
	// 增加活跃连接数
	labels := map[string]string{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"user_agent": c.Request.UserAgent(),
		"remote_ip": c.ClientIP(),
	}
	
	// 记录请求开始指标
	if m.monitoringService != nil {
		m.monitoringService.RecordCustomMetric(
			"http",
			"request_start",
			1.0,
			labels,
		)
	}
}

// recordRequestMetrics 记录请求指标
func (m *PerformanceMonitoringMiddleware) recordRequestMetrics(c *gin.Context, duration time.Duration) {
	if m.monitoringService == nil {
		return
	}

	statusCode := c.Writer.Status()
	isError := statusCode >= 400
	
	// 基础标签
	labels := map[string]string{
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status_code": strconv.Itoa(statusCode),
		"user_agent":  c.Request.UserAgent(),
		"remote_ip":   c.ClientIP(),
	}
	
	// 添加用户信息（如果有）
	if userID, exists := c.Get("user_id"); exists {
		labels["user_id"] = getUserIDString(userID)
	}
	
	if username, exists := c.Get("username"); exists {
		labels["username"] = fmt.Sprintf("%v", username)
	}
	
	// 记录响应时间
	m.monitoringService.RecordCustomMetric(
		"http",
		"response_time",
		float64(duration.Milliseconds()),
		labels,
	)
	
	// 记录请求总数
	m.monitoringService.RecordCustomMetric(
		"http",
		"request_total",
		1.0,
		labels,
	)
	
	// 记录错误请求
	if isError {
		m.monitoringService.RecordCustomMetric(
			"http",
			"request_errors",
			1.0,
			labels,
		)
	}
	
	// 记录请求大小
	if contentLength := c.Request.ContentLength; contentLength > 0 {
		m.monitoringService.RecordCustomMetric(
			"http",
			"request_size_bytes",
			float64(contentLength),
			labels,
		)
	}
	
	// 记录响应大小
	responseSize := c.Writer.Size()
	if responseSize > 0 {
		m.monitoringService.RecordCustomMetric(
			"http",
			"response_size_bytes",
			float64(responseSize),
			labels,
		)
	}
	
	// 记录慢请求
	if duration > 1*time.Second {
		slowLabels := make(map[string]string)
		for k, v := range labels {
			slowLabels[k] = v
		}
		slowLabels["duration_ms"] = strconv.FormatInt(duration.Milliseconds(), 10)
		
		m.monitoringService.RecordCustomMetric(
			"http",
			"slow_requests",
			1.0,
			slowLabels,
		)
	}
	
	// 记录特定状态码
	m.recordStatusCodeMetrics(statusCode, labels)
	
	// 记录路径特定指标
	m.recordPathSpecificMetrics(c.Request.URL.Path, duration, labels)
}

// recordStatusCodeMetrics 记录状态码指标
func (m *PerformanceMonitoringMiddleware) recordStatusCodeMetrics(statusCode int, labels map[string]string) {
	statusLabels := make(map[string]string)
	for k, v := range labels {
		statusLabels[k] = v
	}
	
	switch {
	case statusCode >= 200 && statusCode < 300:
		m.monitoringService.RecordCustomMetric("http", "requests_2xx", 1.0, statusLabels)
	case statusCode >= 300 && statusCode < 400:
		m.monitoringService.RecordCustomMetric("http", "requests_3xx", 1.0, statusLabels)
	case statusCode >= 400 && statusCode < 500:
		m.monitoringService.RecordCustomMetric("http", "requests_4xx", 1.0, statusLabels)
		// 记录客户端错误详情
		statusLabels["error_type"] = "client_error"
		m.monitoringService.RecordCustomMetric("http", "error_details", 1.0, statusLabels)
	case statusCode >= 500:
		m.monitoringService.RecordCustomMetric("http", "requests_5xx", 1.0, statusLabels)
		// 记录服务器错误详情
		statusLabels["error_type"] = "server_error"
		m.monitoringService.RecordCustomMetric("http", "error_details", 1.0, statusLabels)
	}
}

// recordPathSpecificMetrics 记录路径特定指标
func (m *PerformanceMonitoringMiddleware) recordPathSpecificMetrics(path string, duration time.Duration, labels map[string]string) {
	pathLabels := make(map[string]string)
	for k, v := range labels {
		pathLabels[k] = v
	}
	
	// 为API路径记录特殊指标
	if strings.HasPrefix(path, "/api/") {
		pathLabels["api_version"] = extractAPIVersion(path)
		pathLabels["endpoint_type"] = categorizeEndpoint(path)
		
		m.monitoringService.RecordCustomMetric(
			"api",
			"endpoint_calls",
			1.0,
			pathLabels,
		)
		
		m.monitoringService.RecordCustomMetric(
			"api",
			"endpoint_response_time",
			float64(duration.Milliseconds()),
			pathLabels,
		)
	}
	
	// 为认证相关路径记录指标
	if isAuthPath(path) {
		pathLabels["auth_type"] = categorizeAuthPath(path)
		
		m.monitoringService.RecordCustomMetric(
			"auth",
			"auth_attempts",
			1.0,
			pathLabels,
		)
	}
	
	// 为文件上传/下载路径记录指标
	if isFilePath(path) {
		pathLabels["file_operation"] = categorizeFileOperation(path)
		
		m.monitoringService.RecordCustomMetric(
			"file",
			"file_operations",
			1.0,
			pathLabels,
		)
	}
}

// WebSocketMetricsMiddleware WebSocket性能监控中间件
func WebSocketMetricsMiddleware(monitoringService *Services.PerformanceMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if monitoringService == nil {
			c.Next()
			return
		}

		startTime := time.Now()
		
		// 记录WebSocket连接尝试
		labels := map[string]string{
			"connection_type": "websocket",
			"remote_ip":       c.ClientIP(),
			"user_agent":      c.Request.UserAgent(),
		}
		
		monitoringService.RecordCustomMetric(
			"websocket",
			"connection_attempts",
			1.0,
			labels,
		)
		
		c.Next()
		
		// 记录连接处理时间
		duration := time.Since(startTime)
		monitoringService.RecordCustomMetric(
			"websocket",
			"connection_duration",
			float64(duration.Milliseconds()),
			labels,
		)
	}
}

// DatabaseMetricsMiddleware 数据库性能监控中间件
func DatabaseMetricsMiddleware(monitoringService *Services.PerformanceMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if monitoringService == nil {
			c.Next()
			return
		}

		// 设置数据库监控上下文
		c.Set("db_monitor_service", monitoringService)
		c.Set("db_query_start_time", time.Now())
		
		c.Next()
	}
}

// CacheMetricsMiddleware 缓存性能监控中间件
func CacheMetricsMiddleware(monitoringService *Services.PerformanceMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if monitoringService == nil {
			c.Next()
			return
		}

		// 设置缓存监控上下文
		c.Set("cache_monitor_service", monitoringService)
		
		c.Next()
	}
}

// BusinessMetricsMiddleware 业务指标监控中间件
func BusinessMetricsMiddleware(monitoringService *Services.PerformanceMonitoringService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if monitoringService == nil {
			c.Next()
			return
		}

		// 记录用户活动
		if userID, exists := c.Get("user_id"); exists {
			labels := map[string]string{
				"user_id":    getUserIDString(userID),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"user_agent": c.Request.UserAgent(),
			}
			
			monitoringService.RecordCustomMetric(
				"business",
				"user_activity",
				1.0,
				labels,
			)
		}
		
		c.Next()
		
		// 记录业务操作完成
		if c.Writer.Status() < 400 {
			labels := map[string]string{
				"operation": categorizeBusinessOperation(c.Request.URL.Path),
				"status":    "success",
			}
			
			monitoringService.RecordCustomMetric(
				"business",
				"operations_completed",
				1.0,
				labels,
			)
		}
	}
}

// 辅助函数

// getUserIDString 获取用户ID字符串
func getUserIDString(userID interface{}) string {
	switch v := userID.(type) {
	case string:
		return v
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return "unknown"
	}
}

// extractAPIVersion 提取API版本
func extractAPIVersion(path string) string {
	if strings.HasPrefix(path, "/api/v1/") {
		return "v1"
	} else if strings.HasPrefix(path, "/api/v2/") {
		return "v2"
	}
	return "unknown"
}

// categorizeEndpoint 分类端点
func categorizeEndpoint(path string) string {
	switch {
	case strings.Contains(path, "/auth"):
		return "auth"
	case strings.Contains(path, "/user"):
		return "user"
	case strings.Contains(path, "/tag"):
		return "tag"
	case strings.Contains(path, "/file"):
		return "file"
	case strings.Contains(path, "/websocket") || strings.Contains(path, "/ws"):
		return "websocket"
	case strings.Contains(path, "/monitor"):
		return "monitoring"
	case strings.Contains(path, "/log"):
		return "logging"
	case strings.Contains(path, "/query-optimization"):
		return "query_optimization"
	default:
		return "other"
	}
}

// isAuthPath 检查是否为认证路径
func isAuthPath(path string) bool {
	authPaths := []string{"/api/v1/auth/", "/login", "/register", "/logout"}
	for _, authPath := range authPaths {
		if strings.Contains(path, authPath) {
			return true
		}
	}
	return false
}

// categorizeAuthPath 分类认证路径
func categorizeAuthPath(path string) string {
	switch {
	case strings.Contains(path, "/login"):
		return "login"
	case strings.Contains(path, "/register"):
		return "register"
	case strings.Contains(path, "/logout"):
		return "logout"
	case strings.Contains(path, "/refresh"):
		return "refresh"
	case strings.Contains(path, "/forgot"):
		return "forgot_password"
	case strings.Contains(path, "/reset"):
		return "reset_password"
	default:
		return "other"
	}
}

// isFilePath 检查是否为文件路径
func isFilePath(path string) bool {
	filePaths := []string{"/upload", "/download", "/file"}
	for _, filePath := range filePaths {
		if strings.Contains(path, filePath) {
			return true
		}
	}
	return false
}

// categorizeFileOperation 分类文件操作
func categorizeFileOperation(path string) string {
	switch {
	case strings.Contains(path, "/upload"):
		return "upload"
	case strings.Contains(path, "/download"):
		return "download"
	case strings.Contains(path, "/delete"):
		return "delete"
	case strings.Contains(path, "/list"):
		return "list"
	default:
		return "other"
	}
}

// categorizeBusinessOperation 分类业务操作
func categorizeBusinessOperation(path string) string {
	switch {
	case strings.Contains(path, "/user"):
		return "user_management"
	case strings.Contains(path, "/tag"):
		return "tag_management"
	case strings.Contains(path, "/auth"):
		return "authentication"
	case strings.Contains(path, "/file"):
		return "file_management"
	case strings.Contains(path, "/websocket"):
		return "realtime_communication"
	case strings.Contains(path, "/monitor"):
		return "system_monitoring"
	case strings.Contains(path, "/log"):
		return "log_management"
	default:
		return "other"
	}
}
