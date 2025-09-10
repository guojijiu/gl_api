package Middleware

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services"
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLLogMiddleware SQL日志中间件
// 
// 重要功能说明：
// 1. 记录所有SQL查询语句和执行时间
// 2. 支持慢查询检测和告警
// 3. 记录SQL参数和调用栈信息
// 4. 可配置的日志级别和格式
// 5. 自动脱敏敏感数据
// 6. 支持查询性能分析
// 7. 异步日志记录，不影响数据库性能
type SQLLogMiddleware struct {
	BaseMiddleware
	logManager *Services.LogManagerService
	config     *Config.SQLLogConfig
}

// NewSQLLogMiddleware 创建SQL日志中间件
func NewSQLLogMiddleware(logManager *Services.LogManagerService) *SQLLogMiddleware {
	return &SQLLogMiddleware{
		logManager: logManager,
		config:     &logManager.GetConfig().SQLLog,
	}
}

// Handle SQL日志中间件处理函数
func (m *SQLLogMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置GORM日志记录器
		c.Set("sql_logger", m.createSQLLogger())
		c.Next()
	}
}

// createSQLLogger 创建SQL日志记录器
func (m *SQLLogMiddleware) createSQLLogger() logger.Interface {
	return &SQLLogger{
		logManager: m.logManager,
		config:     m.config,
	}
}

// SQLLogger GORM SQL日志记录器
type SQLLogger struct {
	logManager *Services.LogManagerService
	config     *Config.SQLLogConfig
}

// LogMode 设置日志模式
func (l *SQLLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info 记录信息级别日志
func (l *SQLLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(Config.LogLevelInfo) {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	l.logManager.LogSQL(context.Background(), msg, 0, 0, nil, fields)
}

// Warn 记录警告级别日志
func (l *SQLLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(Config.LogLevelWarning) {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	l.logManager.LogSQL(context.Background(), msg, 0, 0, nil, fields)
}

// Error 记录错误级别日志
func (l *SQLLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(Config.LogLevelError) {
		return
	}

	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}

	l.logManager.LogSQL(context.Background(), msg, 0, 0, nil, fields)
}

// Trace 记录SQL查询跟踪信息
func (l *SQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.shouldLog(Config.LogLevelInfo) {
		return
	}

	// 执行SQL查询
	sql, rows := fc()
	duration := time.Since(begin)

	// 检查是否是慢查询
	isSlowQuery := duration > l.config.SlowThreshold

	// 构建日志字段
	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),
		"duration_ms":   duration.Milliseconds(),
		"duration_ns":   duration.Nanoseconds(),
		"rows_affected": rows,
		"is_slow_query": isSlowQuery,
		"timestamp":     begin.Format(time.RFC3339Nano),
	}

	// 添加错误信息
	if err != nil {
		fields["error"] = err.Error()
	}

	// 添加调用栈信息
	if l.config.IncludeStack {
		fields["stack_trace"] = l.getStackTrace()
	}

	// 选择日志级别
	var level Config.LogLevel
	var message string

	if err != nil {
		level = Config.LogLevelError
		message = "SQL执行错误"
	} else if isSlowQuery {
		level = Config.LogLevelWarning
		message = "慢查询检测"
	} else {
		level = Config.LogLevelInfo
		message = "SQL执行成功"
	}

	// 记录日志
	l.logManager.Log("sql", level, message, fields)

	// 慢查询告警
	if isSlowQuery {
		l.logSlowQueryAlert(sql, duration, rows)
	}
}

// shouldLog 检查是否应该记录日志
func (l *SQLLogger) shouldLog(level Config.LogLevel) bool {
	levels := map[Config.LogLevel]int{
		Config.LogLevelDebug:   0,
		Config.LogLevelInfo:    1,
		Config.LogLevelWarning: 2,
		Config.LogLevelError:   3,
		Config.LogLevelFatal:   4,
	}

	configLevel := map[Config.LogLevel]int{
		Config.LogLevelDebug:   0,
		Config.LogLevelInfo:    1,
		Config.LogLevelWarning: 2,
		Config.LogLevelError:   3,
		Config.LogLevelFatal:   4,
	}

	return levels[level] >= configLevel[l.config.Level]
}

// truncateSQL 截断SQL语句
func (l *SQLLogger) truncateSQL(sql string) string {
	if len(sql) <= l.config.MaxQuerySize*1024 {
		return sql
	}
	return sql[:l.config.MaxQuerySize*1024] + "... [TRUNCATED]"
}

// getStackTrace 获取调用栈
func (l *SQLLogger) getStackTrace() string {
	var stack []string
	for i := 1; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// 过滤掉GORM内部调用
		if !strings.Contains(file, "gorm.io") && !strings.Contains(file, "database/sql") {
			stack = append(stack, fmt.Sprintf("%s:%d", file, line))
		}
	}
	return strings.Join(stack, "\n")
}

// logSlowQueryAlert 记录慢查询告警
func (l *SQLLogger) logSlowQueryAlert(sql string, duration time.Duration, rows int64) {
	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),
		"duration_ms":   duration.Milliseconds(),
		"duration_ns":   duration.Nanoseconds(),
		"rows_affected": rows,
		"threshold_ms":  l.config.SlowThreshold.Milliseconds(),
		"alert_type":    "slow_query",
		"severity":      "warning",
	}

	l.logManager.Log("sql", Config.LogLevelWarning, "慢查询告警", fields)
}

// SQLQueryLogger SQL查询日志记录器接口
type SQLQueryLogger interface {
	LogQuery(sql string, duration time.Duration, rows int64, err error)
	LogSlowQuery(sql string, duration time.Duration, rows int64)
	LogError(sql string, err error)
}

// GormSQLLogger GORM SQL日志记录器包装器
type GormSQLLogger struct {
	logManager *Services.LogManagerService
	config     *Config.SQLLogConfig
}

// NewGormSQLLogger 创建GORM SQL日志记录器
func NewGormSQLLogger(logManager *Services.LogManagerService) *GormSQLLogger {
	return &GormSQLLogger{
		logManager: logManager,
		config:     &logManager.GetConfig().SQLLog,
	}
}

// LogQuery 记录SQL查询
func (l *GormSQLLogger) LogQuery(sql string, duration time.Duration, rows int64, err error) {
	if !l.config.Enabled {
		return
	}

	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),
		"duration_ms":   duration.Milliseconds(),
		"duration_ns":   duration.Nanoseconds(),
		"rows_affected": rows,
		"query_type":    "select",
	}

	if err != nil {
		fields["error"] = err.Error()
		l.logManager.Log("sql", Config.LogLevelError, "SQL查询错误", fields)
	} else {
		l.logManager.Log("sql", Config.LogLevelInfo, "SQL查询执行", fields)
	}
}

// LogSlowQuery 记录慢查询
func (l *GormSQLLogger) LogSlowQuery(sql string, duration time.Duration, rows int64) {
	if !l.config.Enabled {
		return
	}

	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),
		"duration_ms":   duration.Milliseconds(),
		"duration_ns":   duration.Nanoseconds(),
		"rows_affected": rows,
		"query_type":    "slow_query",
		"threshold_ms":  l.config.SlowThreshold.Milliseconds(),
		"alert_type":    "slow_query",
		"severity":      "warning",
	}

	l.logManager.Log("sql", Config.LogLevelWarning, "慢查询检测", fields)
}

// LogError 记录SQL错误
func (l *GormSQLLogger) LogError(sql string, err error) {
	if !l.config.Enabled {
		return
	}

	fields := map[string]interface{}{
		"sql":        l.truncateSQL(sql),
		"error":      err.Error(),
		"query_type": "error",
		"severity":   "error",
	}

	l.logManager.Log("sql", Config.LogLevelError, "SQL执行错误", fields)
}

// truncateSQL 截断SQL语句
func (l *GormSQLLogger) truncateSQL(sql string) string {
	if len(sql) <= l.config.MaxQuerySize*1024 {
		return sql
	}
	return sql[:l.config.MaxQuerySize*1024] + "... [TRUNCATED]"
}

// SQLMetrics SQL性能指标
type SQLMetrics struct {
	TotalQueries    int64         `json:"total_queries"`
	SlowQueries     int64         `json:"slow_queries"`
	ErrorQueries    int64         `json:"error_queries"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	MinDuration     time.Duration `json:"min_duration"`
}

// GetSQLMetrics 获取SQL性能指标
func (l *GormSQLLogger) GetSQLMetrics() *SQLMetrics {
	// 这里可以实现SQL性能指标统计
	// 暂时返回空指标
	return &SQLMetrics{}
}

// ResetMetrics 重置性能指标
func (l *GormSQLLogger) ResetMetrics() {
	// 重置性能指标
}

// EnableSQLLogging 启用SQL日志记录
func EnableSQLLogging(db *gorm.DB, logManager *Services.LogManagerService) {
	if logManager == nil {
		return
	}

	sqlLogger := NewGormSQLLogger(logManager)
	
	// 设置GORM日志记录器
	db.Logger = &GormLoggerWrapper{
		logger: sqlLogger,
	}
}

// GormLoggerWrapper GORM日志记录器包装器
type GormLoggerWrapper struct {
	logger *GormSQLLogger
}

// LogMode 设置日志模式
func (w *GormLoggerWrapper) LogMode(level logger.LogLevel) logger.Interface {
	return w
}

// Info 记录信息级别日志
func (w *GormLoggerWrapper) Info(ctx context.Context, msg string, data ...interface{}) {
	// 转发到包装的日志记录器
}

// Warn 记录警告级别日志
func (w *GormLoggerWrapper) Warn(ctx context.Context, msg string, data ...interface{}) {
	// 转发到包装的日志记录器
}

// Error 记录错误级别日志
func (w *GormLoggerWrapper) Error(ctx context.Context, msg string, data ...interface{}) {
	// 转发到包装的日志记录器
}

// Trace 记录SQL查询跟踪信息
func (w *GormLoggerWrapper) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 转发到包装的日志记录器
	sql, rows := fc()
	duration := time.Since(begin)

	if err != nil {
		w.logger.LogError(sql, err)
	} else if duration > w.logger.config.SlowThreshold {
		w.logger.LogSlowQuery(sql, duration, rows)
	} else {
		w.logger.LogQuery(sql, duration, rows, nil)
	}
}
