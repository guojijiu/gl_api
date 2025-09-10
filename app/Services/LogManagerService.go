package Services

import (
	"cloud-platform-api/app/Config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogManagerService 日志管理器服务
type LogManagerService struct {
	config     *Config.LogConfig
	loggers    map[string]*Logger
	mu         sync.RWMutex
	closed     bool
	asyncQueue chan LogEntry
	stats      *LogStats
	ctx        context.Context
	cancel     context.CancelFunc
}

// LogStats 日志统计信息
type LogStats struct {
	mu           sync.RWMutex
	TotalLogs    int64                     `json:"total_logs"`
	LogsByLevel  map[Config.LogLevel]int64 `json:"logs_by_level"`
	LogsByLogger map[string]int64          `json:"logs_by_logger"`
	Errors       int64                     `json:"errors"`
	Performance  map[string]float64        `json:"performance"`
	LastReset    time.Time                 `json:"last_reset"`
}

// Logger 日志记录器
type Logger struct {
	name      string
	config    interface{}
	writer    io.Writer
	formatter LogFormatter
	level     Config.LogLevel
	enabled   bool
	mu        sync.Mutex
	stats     *LoggerStats
}

// LoggerStats 单个日志记录器统计
type LoggerStats struct {
	mu           sync.RWMutex
	TotalLogs    int64     `json:"total_logs"`
	LastLog      time.Time `json:"last_log"`
	ErrorCount   int64     `json:"error_count"`
	WriteLatency []float64 `json:"write_latency"`
}

// LogEntry 日志条目
type LogEntry struct {
	Logger    string                 `json:"logger"`
	Level     Config.LogLevel        `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	UserID    uint                   `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// CallerInfo 调用者信息
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Package  string `json:"package"`
}

// DailyRotateWriter 按日期轮转的日志写入器
type DailyRotateWriter struct {
	basePath    string
	baseName    string
	currentFile string
	config      Config.LogRotation
	mu          sync.Mutex
	file        *os.File
}

// Write 实现io.Writer接口
func (w *DailyRotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查是否需要轮转（按日期）
	today := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(w.basePath, fmt.Sprintf("%s-%s.log", w.baseName, today))

	if w.currentFile != expectedFile {
		// 关闭旧文件
		if w.file != nil {
			w.file.Close()
			w.file = nil
		}

		// 更新当前文件路径
		w.currentFile = expectedFile
	}

	// 打开或创建新文件
	if w.file == nil {
		// 确保目录存在
		if err := os.MkdirAll(w.basePath, 0755); err != nil {
			return 0, fmt.Errorf("创建日志目录失败: %v", err)
		}

		file, err := os.OpenFile(w.currentFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, fmt.Errorf("打开日志文件失败: %v", err)
		}
		w.file = file
	}

	// 写入数据
	return w.file.Write(p)
}

// Close 关闭文件
func (w *DailyRotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// LogFormatter 日志格式化接口
type LogFormatter interface {
	Format(entry LogEntry) ([]byte, error)
}

// JSONFormatter JSON格式日志
type JSONFormatter struct {
	PrettyPrint bool
	IncludeTime bool
}

// TextFormatter 文本格式日志
type TextFormatter struct {
	ShowTimestamp bool
	ShowCaller    bool
	ShowLevel     bool
	Colorize      bool
}

// NewLogManagerService 创建日志管理器服务
func NewLogManagerService(config *Config.LogConfig) *LogManagerService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &LogManagerService{
		config:     config,
		loggers:    make(map[string]*Logger),
		asyncQueue: make(chan LogEntry, 1000),
		stats: &LogStats{
			LogsByLevel:  make(map[Config.LogLevel]int64),
			LogsByLogger: make(map[string]int64),
			Performance:  make(map[string]float64),
			LastReset:    time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	service.initLoggers()
	go service.processAsyncLogs()
	go service.collectStats()
	go service.monitorPerformance()

	return service
}

// initLoggers 初始化各类型日志记录器
func (s *LogManagerService) initLoggers() {
	if s.config.RequestLog.Enabled {
		s.createLogger("request", s.config.RequestLog, s.config.RequestLog.Path)
	}
	if s.config.SQLLog.Enabled {
		s.createLogger("sql", s.config.SQLLog, s.config.SQLLog.Path)
	}
	if s.config.ErrorLog.Enabled {
		s.createLogger("error", s.config.ErrorLog, s.config.ErrorLog.Path)
	}
	if s.config.AuditLog.Enabled {
		s.createLogger("audit", s.config.AuditLog, s.config.AuditLog.Path)
	}
	if s.config.SecurityLog.Enabled {
		s.createLogger("security", s.config.SecurityLog, s.config.SecurityLog.Path)
	}
	if s.config.BusinessLog.Enabled {
		s.createLogger("business", s.config.BusinessLog, s.config.BusinessLog.Path)
	}
	if s.config.AccessLog.Enabled {
		s.createLogger("access", s.config.AccessLog, s.config.AccessLog.Path)
	}

	// 默认日志记录器已关闭，只记录特定类型的日志
	// s.createDefaultLogger()
}

// createDefaultLogger 创建默认日志记录器
func (s *LogManagerService) createDefaultLogger() {
	defaultPath := filepath.Join(s.config.BasePath, "system")
	if err := os.MkdirAll(defaultPath, 0755); err != nil {
		fmt.Printf("创建默认日志目录失败: %v\n", err)
		return
	}

	// 按日期生成日志文件名
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(defaultPath, fmt.Sprintf("system-%s.log", today))

	// 创建自定义的按日期轮转的writer
	writer := &DailyRotateWriter{
		basePath:    defaultPath,
		baseName:    "system",
		currentFile: logFile,
		config:      s.config.Rotation,
	}

	logger := &Logger{
		name:      "system",
		config:    nil,
		writer:    writer,
		formatter: &JSONFormatter{},
		level:     Config.LogLevelInfo,
		enabled:   true,
		stats:     &LoggerStats{},
	}

	s.mu.Lock()
	s.loggers["system"] = logger
	s.mu.Unlock()
}

// createLogger 创建日志记录器
func (s *LogManagerService) createLogger(name string, config interface{}, path string) {
	logPath := filepath.Join(s.config.BasePath, path)
	if err := os.MkdirAll(logPath, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %s, %v\n", path, err)
		return
	}

	// 按日期生成日志文件名
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logPath, fmt.Sprintf("%s-%s.log", name, today))

	// 创建自定义的按日期轮转的writer
	writer := &DailyRotateWriter{
		basePath:    logPath,
		baseName:    name,
		currentFile: logFile,
		config:      s.config.Rotation,
	}

	var formatter LogFormatter
	switch s.getLogFormat(config) {
	case Config.LogFormatJSON:
		formatter = &JSONFormatter{
			PrettyPrint: s.config.Format == "pretty",
			IncludeTime: true,
		}
	case Config.LogFormatText:
		formatter = &TextFormatter{
			ShowTimestamp: true,
			ShowCaller:    true,
			ShowLevel:     true,
			Colorize:      s.config.Format == "color",
		}
	default:
		formatter = &JSONFormatter{}
	}

	logger := &Logger{
		name:      name,
		config:    config,
		writer:    writer,
		formatter: formatter,
		level:     s.getLogLevel(config),
		enabled:   s.isLoggerEnabled(config),
		stats:     &LoggerStats{},
	}

	s.mu.Lock()
	s.loggers[name] = logger
	s.mu.Unlock()
}

// 获取配置的辅助方法
func (s *LogManagerService) getLogLevel(config interface{}) Config.LogLevel {
	switch c := config.(type) {
	case Config.RequestLogConfig:
		return c.Level
	case Config.SQLLogConfig:
		return c.Level
	case Config.ErrorLogConfig:
		return c.Level
	case Config.AuditLogConfig:
		return c.Level
	case Config.SecurityLogConfig:
		return c.Level
	case Config.BusinessLogConfig:
		return c.Level
	case Config.AccessLogConfig:
		return c.Level
	default:
		return Config.LogLevelInfo
	}
}

func (s *LogManagerService) getLogFormat(config interface{}) Config.LogFormat {
	switch c := config.(type) {
	case Config.RequestLogConfig:
		return c.Format
	case Config.SQLLogConfig:
		return c.Format
	case Config.ErrorLogConfig:
		return c.Format
	case Config.AuditLogConfig:
		return c.Format
	case Config.SecurityLogConfig:
		return c.Format
	case Config.BusinessLogConfig:
		return c.Format
	case Config.AccessLogConfig:
		return c.Format
	default:
		return Config.LogFormatJSON
	}
}

func (s *LogManagerService) isLoggerEnabled(config interface{}) bool {
	switch c := config.(type) {
	case Config.RequestLogConfig:
		return c.Enabled
	case Config.SQLLogConfig:
		return c.Enabled
	case Config.ErrorLogConfig:
		return c.Enabled
	case Config.AuditLogConfig:
		return c.Enabled
	case Config.SecurityLogConfig:
		return c.Enabled
	case Config.BusinessLogConfig:
		return c.Enabled
	case Config.AccessLogConfig:
		return c.Enabled
	default:
		return true
	}
}

// Log 记录日志
func (s *LogManagerService) Log(loggerName string, level Config.LogLevel, message string, fields map[string]interface{}) {
	if s.closed {
		return
	}

	caller := s.getCallerInfo()
	entry := LogEntry{
		Logger:    loggerName,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Fields:    fields,
		Caller:    caller,
	}

	if level >= Config.LogLevelError && s.shouldIncludeStack(loggerName) {
		entry.Stack = s.getStackTrace()
	}

	select {
	case s.asyncQueue <- entry:
		// 异步处理，统计信息在 processAsyncLogs 中更新
	default:
		// 队列满了，同步处理
		s.writeLogSync(entry)
		s.updateStats(entry)
	}
}

// LogWithContext 记录带上下文的日志
func (s *LogManagerService) LogWithContext(ctx context.Context, loggerName string, level Config.LogLevel, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	if ip := ctx.Value("client_ip"); ip != nil {
		fields["ip"] = ip
	}
	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		fields["user_agent"] = userAgent
	}

	s.Log(loggerName, level, message, fields)
}

// 专用日志方法
func (s *LogManagerService) LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, fields map[string]interface{}) {
	if !s.config.RequestLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["method"] = method
	fields["path"] = path
	fields["status_code"] = statusCode
	fields["duration"] = duration.String()
	fields["duration_ms"] = duration.Milliseconds()

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	if ip := ctx.Value("client_ip"); ip != nil {
		fields["ip"] = ip
	}
	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		fields["user_agent"] = userAgent
	}

	s.Log("request", Config.LogLevelInfo, fmt.Sprintf("%s %s - %d", method, path, statusCode), fields)
}

func (s *LogManagerService) LogSQL(ctx context.Context, sql string, duration time.Duration, rows int64, error error, fields map[string]interface{}) {
	if !s.config.SQLLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["sql"] = sql
	fields["duration"] = duration.String()
	fields["duration_ms"] = duration.Milliseconds()
	fields["rows"] = rows

	if error != nil {
		fields["error"] = error.Error()
	}

	if duration > s.config.SQLLog.SlowThreshold {
		fields["slow_query"] = true
		fields["slow_threshold"] = s.config.SQLLog.SlowThreshold.String()
	}

	level := Config.LogLevelInfo
	if error != nil {
		level = Config.LogLevelError
	} else if duration > s.config.SQLLog.SlowThreshold {
		level = Config.LogLevelWarning
	}

	s.Log("sql", level, "SQL执行", fields)
}

func (s *LogManagerService) LogError(ctx context.Context, error error, message string, fields map[string]interface{}) {
	if !s.config.ErrorLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["error"] = error.Error()
	fields["error_type"] = fmt.Sprintf("%T", error)

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	s.Log("error", Config.LogLevelError, message, fields)
}

func (s *LogManagerService) LogAudit(ctx context.Context, action string, resource string, resourceID interface{}, fields map[string]interface{}) {
	if !s.config.AuditLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["action"] = action
	fields["resource"] = resource
	fields["resource_id"] = resourceID

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	if ip := ctx.Value("client_ip"); ip != nil {
		fields["ip"] = ip
	}

	s.Log("audit", Config.LogLevelInfo, fmt.Sprintf("审计: %s %s", action, resource), fields)
}

func (s *LogManagerService) LogSecurity(ctx context.Context, event string, level Config.LogLevel, fields map[string]interface{}) {
	if !s.config.SecurityLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["security_event"] = event
	fields["security_level"] = string(level)

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	if ip := ctx.Value("client_ip"); ip != nil {
		fields["ip"] = ip
	}

	s.Log("security", level, fmt.Sprintf("安全事件: %s", event), fields)
}

func (s *LogManagerService) LogBusiness(ctx context.Context, module string, action string, message string, fields map[string]interface{}) {
	if !s.config.BusinessLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["module"] = module
	fields["action"] = action

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	s.Log("business", Config.LogLevelInfo, message, fields)
}

func (s *LogManagerService) LogAccess(ctx context.Context, method, path string, statusCode int, userAgent string, fields map[string]interface{}) {
	if !s.config.AccessLog.Enabled {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["method"] = method
	fields["path"] = path
	fields["status_code"] = statusCode
	fields["user_agent"] = userAgent

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	if ip := ctx.Value("client_ip"); ip != nil {
		fields["ip"] = ip
	}

	s.Log("access", Config.LogLevelInfo, fmt.Sprintf("访问: %s %s", method, path), fields)
}

// 辅助方法
func (s *LogManagerService) getCallerInfo() *CallerInfo {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return nil
	}

	funcObj := runtime.FuncForPC(pc)
	if funcObj == nil {
		return nil
	}

	funcName := funcObj.Name()
	lastDot := strings.LastIndex(funcName, ".")
	if lastDot != -1 {
		funcName = funcName[lastDot+1:]
	}

	return &CallerInfo{
		File:     filepath.Base(file),
		Line:     line,
		Function: funcName,
		Package:  filepath.Dir(file),
	}
}

func (s *LogManagerService) getStackTrace() string {
	var stack []string
	for i := 1; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		funcObj := runtime.FuncForPC(pc)
		if funcObj == nil {
			continue
		}

		stack = append(stack, fmt.Sprintf("%s:%d %s", filepath.Base(file), line, funcObj.Name()))
	}

	return strings.Join(stack, "\n")
}

func (s *LogManagerService) shouldIncludeStack(loggerName string) bool {
	switch loggerName {
	case "error", "security":
		return true
	default:
		return false
	}
}

func (s *LogManagerService) writeLogSync(entry LogEntry) {
	s.mu.RLock()
	logger, exists := s.loggers[entry.Logger]
	s.mu.RUnlock()

	if !exists || !logger.enabled {
		return
	}

	if entry.Level < logger.level {
		return
	}

	data, err := logger.formatter.Format(entry)
	if err != nil {
		fmt.Printf("日志格式化失败: %v\n", err)
		return
	}

	start := time.Now()
	// 添加换行符确保每条日志占一行
	logData := append(data, '\n')
	_, err = logger.writer.Write(logData)
	if err != nil {
		fmt.Printf("日志写入失败: %v\n", err)
		return
	}

	latency := time.Since(start).Seconds()
	logger.stats.mu.Lock()
	logger.stats.WriteLatency = append(logger.stats.WriteLatency, latency)
	if len(logger.stats.WriteLatency) > 100 {
		logger.stats.WriteLatency = logger.stats.WriteLatency[1:]
	}
	logger.stats.mu.Unlock()
}

func (s *LogManagerService) processAsyncLogs() {
	for {
		select {
		case entry := <-s.asyncQueue:
			s.writeLogSync(entry)
			s.updateStats(entry)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *LogManagerService) collectStats() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.resetStats()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *LogManagerService) monitorPerformance() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.calculatePerformanceMetrics()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *LogManagerService) updateStats(entry LogEntry) {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	s.stats.TotalLogs++
	s.stats.LogsByLevel[entry.Level]++
	s.stats.LogsByLogger[entry.Logger]++

	s.mu.RLock()
	if logger, exists := s.loggers[entry.Logger]; exists {
		logger.stats.mu.Lock()
		logger.stats.TotalLogs++
		logger.stats.LastLog = entry.Timestamp
		if entry.Level >= Config.LogLevelError {
			logger.stats.ErrorCount++
		}
		logger.stats.mu.Unlock()
	}
	s.mu.RUnlock()
}

func (s *LogManagerService) resetStats() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	s.stats.TotalLogs = 0
	for level := range s.stats.LogsByLevel {
		s.stats.LogsByLevel[level] = 0
	}
	for logger := range s.stats.LogsByLogger {
		s.stats.LogsByLogger[logger] = 0
	}
	s.stats.Errors = 0
	s.stats.LastReset = time.Now()
}

func (s *LogManagerService) calculatePerformanceMetrics() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, logger := range s.loggers {
		logger.stats.mu.RLock()
		if len(logger.stats.WriteLatency) > 0 {
			var total float64
			for _, latency := range logger.stats.WriteLatency {
				total += latency
			}
			avgLatency := total / float64(len(logger.stats.WriteLatency))

			s.stats.mu.Lock()
			s.stats.Performance[fmt.Sprintf("%s_avg_latency", name)] = avgLatency
			s.stats.mu.Unlock()
		}
		logger.stats.mu.RUnlock()
	}
}

// 公共方法
func (s *LogManagerService) GetStats() *LogStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	stats := &LogStats{
		TotalLogs:    s.stats.TotalLogs,
		LogsByLevel:  make(map[Config.LogLevel]int64),
		LogsByLogger: make(map[string]int64),
		Errors:       s.stats.Errors,
		Performance:  make(map[string]float64),
		LastReset:    s.stats.LastReset,
	}

	for level, count := range s.stats.LogsByLevel {
		stats.LogsByLevel[level] = count
	}
	for logger, count := range s.stats.LogsByLogger {
		stats.LogsByLogger[logger] = count
	}
	for metric, value := range s.stats.Performance {
		stats.Performance[metric] = value
	}

	return stats
}

func (s *LogManagerService) GetLoggerStats(loggerName string) *LoggerStats {
	s.mu.RLock()
	logger, exists := s.loggers[loggerName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	logger.stats.mu.RLock()
	defer logger.stats.mu.RUnlock()

	stats := &LoggerStats{
		TotalLogs:    logger.stats.TotalLogs,
		LastLog:      logger.stats.LastLog,
		ErrorCount:   logger.stats.ErrorCount,
		WriteLatency: make([]float64, len(logger.stats.WriteLatency)),
	}

	copy(stats.WriteLatency, logger.stats.WriteLatency)
	return stats
}

func (s *LogManagerService) Close() error {
	s.closed = true
	s.cancel()

	time.Sleep(100 * time.Millisecond)

	s.mu.RLock()
	for _, logger := range s.loggers {
		if closer, ok := logger.writer.(io.Closer); ok {
			closer.Close()
		}
	}
	s.mu.RUnlock()

	return nil
}

// 格式化器实现
func (f *JSONFormatter) Format(entry LogEntry) ([]byte, error) {
	if f.PrettyPrint {
		return json.MarshalIndent(entry, "", "  ")
	}
	return json.Marshal(entry)
}

func (f *TextFormatter) Format(entry LogEntry) ([]byte, error) {
	var parts []string

	if f.ShowTimestamp {
		parts = append(parts, entry.Timestamp.Format("2006-01-02 15:04:05"))
	}

	if f.ShowLevel {
		levelStr := string(entry.Level)
		if f.Colorize {
			levelStr = f.colorizeLevel(levelStr, entry.Level)
		}
		parts = append(parts, levelStr)
	}

	if f.ShowCaller && entry.Caller != nil {
		parts = append(parts, fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line))
	}

	parts = append(parts, entry.Message)

	if len(entry.Fields) > 0 {
		fieldsStr := f.formatFields(entry.Fields)
		parts = append(parts, fieldsStr)
	}

	return []byte(strings.Join(parts, " ")), nil
}

func (f *TextFormatter) colorizeLevel(level string, logLevel Config.LogLevel) string {
	switch logLevel {
	case Config.LogLevelDebug:
		return "\033[36m" + level + "\033[0m"
	case Config.LogLevelInfo:
		return "\033[32m" + level + "\033[0m"
	case Config.LogLevelWarning:
		return "\033[33m" + level + "\033[0m"
	case Config.LogLevelError:
		return "\033[31m" + level + "\033[0m"
	case Config.LogLevelFatal:
		return "\033[35m" + level + "\033[0m"
	default:
		return level
	}
}

func (f *TextFormatter) formatFields(fields map[string]interface{}) string {
	var parts []string
	for key, value := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// 便捷方法
func (s *LogManagerService) Debug(loggerName, message string, fields map[string]interface{}) {
	s.Log(loggerName, Config.LogLevelDebug, message, fields)
}

func (s *LogManagerService) Info(loggerName, message string, fields map[string]interface{}) {
	s.Log(loggerName, Config.LogLevelInfo, message, fields)
}

func (s *LogManagerService) Warning(loggerName, message string, fields map[string]interface{}) {
	s.Log(loggerName, Config.LogLevelWarning, message, fields)
}

func (s *LogManagerService) Error(loggerName, message string, fields map[string]interface{}) {
	s.Log(loggerName, Config.LogLevelError, message, fields)
}

func (s *LogManagerService) Fatal(loggerName, message string, fields map[string]interface{}) {
	s.Log(loggerName, Config.LogLevelFatal, message, fields)
}

// GetConfig 获取日志配置
func (s *LogManagerService) GetConfig() *Config.LogConfig {
	return s.config
}
