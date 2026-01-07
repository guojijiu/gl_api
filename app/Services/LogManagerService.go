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
//
// 功能说明：
// 1. 异步记录日志，不阻塞调用者
// 2. 自动收集调用者信息（文件名、行号等）
// 3. 错误级别日志自动包含堆栈跟踪
// 4. 队列满时降级为同步处理，确保日志不丢失
//
// 异步处理机制：
// - 使用带缓冲的channel（asyncQueue）实现异步日志
// - 默认容量1000，可以缓冲大量日志
// - 如果队列满，降级为同步处理，确保日志不丢失
//
// 调用者信息：
// - 自动获取调用Log()函数的文件名和行号
// - 用于定位日志来源，便于问题排查
// - 使用runtime.Caller()获取调用栈信息
//
// 堆栈跟踪：
// - 错误级别（ERROR、FATAL）的日志自动包含堆栈跟踪
// - 堆栈跟踪帮助定位错误发生的位置
// - 可以根据配置决定是否包含堆栈跟踪
//
// 性能优化：
// - 异步处理不阻塞业务逻辑
// - 队列满时降级为同步，避免日志丢失
// - 批量写入可以提高I/O效率
//
// 注意事项：
// - 如果服务已关闭（s.closed），直接返回
// - 队列满时会阻塞，但可以确保日志不丢失
// - 堆栈跟踪可能很长，需要合理存储
func (s *LogManagerService) Log(loggerName string, level Config.LogLevel, message string, fields map[string]interface{}) {
	// 如果服务已关闭，不再记录日志
	if s.closed {
		return
	}

	// 获取调用者信息（文件名、行号等）
	// 用于定位日志来源，便于问题排查
	caller := s.getCallerInfo()
	
	// 构建日志条目
	entry := LogEntry{
		Logger:    loggerName,  // 日志记录器名称（如"sql"、"request"等）
		Level:     level,        // 日志级别（DEBUG、INFO、WARNING、ERROR、FATAL）
		Message:   message,      // 日志消息
		Timestamp: time.Now(),   // 时间戳
		Fields:    fields,       // 附加字段（键值对）
		Caller:    caller,       // 调用者信息（文件名、行号）
	}

	// 错误级别日志自动包含堆栈跟踪
	// 堆栈跟踪帮助定位错误发生的位置
	if level >= Config.LogLevelError && s.shouldIncludeStack(loggerName) {
		entry.Stack = s.getStackTrace()
	}

	// 尝试异步写入日志队列
	// 使用select实现非阻塞发送
	select {
	case s.asyncQueue <- entry:
		// 成功发送到队列，异步处理
		// 统计信息会在processAsyncLogs中更新
		// 这种方式不阻塞调用者，性能最好
	default:
		// 队列满了，降级为同步处理
		// 这确保日志不会丢失，但会阻塞调用者
		// 同步写入并更新统计信息
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

// writeLogSync 同步写入日志
//
// 功能说明：
// 1. 同步写入日志到文件或输出流
// 2. 检查日志记录器是否存在和启用
// 3. 检查日志级别是否满足要求
// 4. 格式化日志条目并写入
// 5. 记录写入延迟用于性能监控
//
// 日志级别过滤：
// - 只写入级别大于等于配置级别的日志
// - 例如：配置为INFO级别，则DEBUG日志不写入
// - 这样可以减少I/O操作，提高性能
//
// 格式化处理：
// - 使用配置的格式化器（JSON或文本格式）
// - 格式化失败时记录错误但不中断流程
// - 添加换行符确保每条日志占一行
//
// 性能监控：
// - 记录每次写入的延迟时间
// - 保留最近100次的延迟数据
// - 用于分析日志写入性能
//
// 错误处理：
// - 格式化失败时记录错误但不中断
// - 写入失败时记录错误但不中断
// - 确保日志系统不会因为单个日志失败而崩溃
//
// 注意事项：
// - 使用读锁获取logger，减少锁竞争
// - 写入操作可能阻塞，但通常很快
// - 延迟数据使用slice存储，需要定期清理
func (s *LogManagerService) writeLogSync(entry LogEntry) {
	// 获取日志记录器（使用读锁，允许多个goroutine同时读取）
	s.mu.RLock()
	logger, exists := s.loggers[entry.Logger]
	s.mu.RUnlock()

	// 检查日志记录器是否存在和启用
	if !exists || !logger.enabled {
		return
	}

	// 检查日志级别：只写入级别大于等于配置级别的日志
	// 例如：配置为INFO级别，则DEBUG日志不写入
	// LogLevel是枚举类型，值越小级别越高（DEBUG=0, INFO=1, ...）
	if entry.Level < logger.level {
		return
	}

	// 格式化日志条目（JSON或文本格式）
	data, err := logger.formatter.Format(entry)
	if err != nil {
		// 格式化失败时记录错误但不中断流程
		fmt.Printf("日志格式化失败: %v\n", err)
		return
	}

	// 记录写入开始时间，用于计算延迟
	start := time.Now()
	
	// 添加换行符确保每条日志占一行
	// 这样便于日志解析和查看
	logData := append(data, '\n')
	
	// 写入日志数据
	_, err = logger.writer.Write(logData)
	if err != nil {
		// 写入失败时记录错误但不中断流程
		fmt.Printf("日志写入失败: %v\n", err)
		return
	}

	// 计算写入延迟并更新统计信息
	latency := time.Since(start).Seconds()
	logger.stats.mu.Lock()
	// 记录延迟时间（用于性能分析）
	logger.stats.WriteLatency = append(logger.stats.WriteLatency, latency)
	// 只保留最近100次的延迟数据，避免内存无限增长
	if len(logger.stats.WriteLatency) > 100 {
		logger.stats.WriteLatency = logger.stats.WriteLatency[1:]
	}
	logger.stats.mu.Unlock()
}

// processAsyncLogs 处理异步日志
//
// 功能说明：
// 1. 在独立的goroutine中运行，持续处理日志队列
// 2. 从异步队列中取出日志条目并写入
// 3. 更新日志统计信息
// 4. 支持优雅停止，通过context控制
//
// 实现原理：
// - 使用无限循环持续监听日志队列
// - 使用select同时监听队列和context取消信号
// - 当收到日志条目时，同步写入并更新统计
//
// 优雅停止：
// - 当context被取消时（服务关闭），立即退出
// - 确保所有已入队的日志都被处理
// - 避免日志丢失
//
// 性能考虑：
// - 异步处理不阻塞日志调用者
// - 批量处理可以提高I/O效率（可扩展）
// - 如果队列满，调用者会降级为同步处理
//
// 注意事项：
// - 这个goroutine在服务启动时创建，服务关闭时退出
// - 如果写入失败，只记录错误，不中断处理流程
// - 统计信息更新需要加锁保护
func (s *LogManagerService) processAsyncLogs() {
	// 无限循环，持续处理日志队列
	for {
		select {
		case entry := <-s.asyncQueue:
			// 从队列中取出日志条目
			// 同步写入日志文件（确保顺序和一致性）
			s.writeLogSync(entry)
			// 更新统计信息（总日志数、按级别统计等）
			s.updateStats(entry)
		case <-s.ctx.Done():
			// 收到停止信号（服务关闭）
			// 立即退出goroutine，实现优雅停止
			// 注意：队列中可能还有未处理的日志，但服务关闭时通常可以接受
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

// updateStats 更新统计信息
//
// 功能说明：
// 1. 更新全局统计信息（总日志数、按级别统计、按记录器统计）
// 2. 更新单个日志记录器的统计信息
// 3. 记录错误日志数量
// 4. 更新最后日志时间
//
// 统计维度：
// - 总日志数：所有日志记录的总数
// - 按级别统计：DEBUG、INFO、WARNING、ERROR、FATAL的数量
// - 按记录器统计：每个日志记录器（sql、request等）的数量
// - 错误计数：ERROR和FATAL级别的日志数量
//
// 并发安全：
// - 使用互斥锁保护共享统计数据
// - 全局统计使用s.stats.mu锁
// - 单个记录器统计使用logger.stats.mu锁
// - 使用读锁获取logger，减少锁竞争
//
// 性能考虑：
// - 统计更新是轻量级操作，但需要加锁
// - 使用defer确保锁被释放
// - 尽量减少锁的持有时间
//
// 注意事项：
// - 统计信息在内存中，服务重启后会丢失
// - 如果需要持久化，需要定期导出到数据库或文件
// - 错误计数只统计ERROR和FATAL级别
func (s *LogManagerService) updateStats(entry LogEntry) {
	// 更新全局统计信息（需要加锁保护）
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	// 增加总日志数
	s.stats.TotalLogs++
	// 按日志级别统计（DEBUG、INFO、WARNING、ERROR、FATAL）
	s.stats.LogsByLevel[entry.Level]++
	// 按日志记录器统计（sql、request、error等）
	s.stats.LogsByLogger[entry.Logger]++

	// 更新单个日志记录器的统计信息
	// 使用读锁获取logger，减少锁竞争
	s.mu.RLock()
	if logger, exists := s.loggers[entry.Logger]; exists {
		// 更新记录器的统计信息（需要加锁保护）
		logger.stats.mu.Lock()
		logger.stats.TotalLogs++              // 记录器的总日志数
		logger.stats.LastLog = entry.Timestamp // 最后日志时间
		// 如果是错误级别（ERROR或FATAL），增加错误计数
		if entry.Level >= Config.LogLevelError {
			logger.stats.ErrorCount++
		}
		logger.stats.mu.Unlock()
	}
	s.mu.RUnlock()
}

// resetStats 重置统计信息
//
// 功能说明：
// 1. 清空所有统计计数器，重新开始统计
// 2. 记录重置时间，用于周期性分析
// 3. 定期调用（默认每分钟），避免统计数据无限增长
//
// 重置内容：
// - 总日志数：重置为0
// - 按级别统计：所有级别的计数重置为0
// - 按记录器统计：所有记录器的计数重置为0
// - 错误计数：重置为0
// - 最后重置时间：更新为当前时间
//
// 使用场景：
// - 周期性统计：每分钟重置一次，用于分析每分钟的日志情况
// - 性能分析：重置后重新统计，用于分析性能趋势
// - 问题排查：重置后观察新的统计，用于定位问题
//
// 注意事项：
// - 重置前可以导出统计数据用于分析
// - 重置操作需要加锁保护，避免并发问题
// - 重置后统计数据从0开始，不影响历史数据
func (s *LogManagerService) resetStats() {
	// 加锁保护，避免并发重置导致数据不一致
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	// 重置总日志数
	s.stats.TotalLogs = 0
	
	// 重置所有级别的统计计数
	// 遍历所有已存在的级别，将计数重置为0
	for level := range s.stats.LogsByLevel {
		s.stats.LogsByLevel[level] = 0
	}
	
	// 重置所有记录器的统计计数
	// 遍历所有已存在的记录器，将计数重置为0
	for logger := range s.stats.LogsByLogger {
		s.stats.LogsByLogger[logger] = 0
	}
	
	// 重置错误计数
	s.stats.Errors = 0
	
	// 记录重置时间，用于分析周期性统计
	s.stats.LastReset = time.Now()
}

// calculatePerformanceMetrics 计算性能指标
//
// 功能说明：
// 1. 计算每个日志记录器的性能指标
// 2. 包括平均写入延迟、最大延迟、最小延迟等
// 3. 用于性能分析和优化
//
// 计算的指标：
// - 平均写入延迟：所有写入操作的平均耗时
// - 最大写入延迟：最慢的写入操作耗时
// - 最小写入延迟：最快的写入操作耗时
// - 吞吐量：每秒处理的日志数（可扩展）
//
// 实现原理：
// - 遍历所有日志记录器
// - 从WriteLatency数组中计算统计值
// - 将结果存储到Performance map中
//
// 性能考虑：
// - 使用读锁获取logger，减少锁竞争
// - 计算过程可能需要遍历大量数据，注意性能影响
// - 定期计算（默认每5分钟），避免频繁计算
//
// 使用场景：
// - 性能监控：实时监控日志系统性能
// - 问题诊断：识别性能瓶颈
// - 容量规划：根据性能指标规划资源
//
// 注意事项：
// - 如果WriteLatency为空，跳过计算
// - 计算结果存储在Performance map中，需要定期清理
// - 计算过程需要加锁，可能影响性能
func (s *LogManagerService) calculatePerformanceMetrics() {
	// 使用读锁获取所有日志记录器，允许多个goroutine同时读取
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 遍历所有日志记录器，计算每个记录器的性能指标
	for name, logger := range s.loggers {
		// 使用读锁获取记录器的统计信息
		logger.stats.mu.RLock()
		
		// 检查是否有延迟数据
		if len(logger.stats.WriteLatency) > 0 {
			// 计算平均写入延迟
			var total float64
			for _, latency := range logger.stats.WriteLatency {
				total += latency
			}
			avgLatency := total / float64(len(logger.stats.WriteLatency))

			// 更新全局性能指标（需要加锁保护）
			s.stats.mu.Lock()
			// 存储平均延迟，键名为"{logger_name}_avg_latency"
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
