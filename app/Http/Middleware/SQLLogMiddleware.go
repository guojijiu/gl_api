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
//
// 功能说明：
// 1. 记录SQL查询的详细信息（SQL语句、执行时间、影响行数等）
// 2. 检测慢查询（超过阈值的查询）
// 3. 记录错误信息（如果查询失败）
// 4. 可选地记录调用栈（便于调试）
// 5. 根据查询结果选择日志级别（错误、警告、信息）
//
// 参数说明：
// - ctx: 上下文（用于传递请求信息）
// - begin: SQL查询开始时间（用于计算执行时间）
// - fc: 回调函数（执行SQL查询并返回SQL语句和影响行数）
// - err: 查询错误（如果查询失败）
//
// 执行流程：
// 1. 检查是否应该记录日志（根据日志级别配置）
// 2. 执行SQL查询（通过回调函数）
// 3. 计算执行时间（从开始时间到现在）
// 4. 检查是否是慢查询（超过配置的阈值）
// 5. 构建日志字段（SQL、执行时间、影响行数、错误等）
// 6. 可选地添加调用栈（如果配置启用）
// 7. 根据查询结果选择日志级别
// 8. 记录日志（通过LogManagerService）
// 9. 如果是慢查询，发送告警
//
// 日志级别：
// - Error: SQL执行错误
// - Warning: 慢查询（超过阈值）
// - Info: 正常查询
//
// 性能考虑：
// - 日志记录是异步的，不会阻塞SQL查询
// - SQL语句会被截断（如果超过最大长度）
// - 调用栈记录是可选的，默认关闭以提高性能
//
// 注意事项：
// - 慢查询阈值由配置决定
// - SQL语句会被截断以防止日志过大
// - 调用栈记录会影响性能，建议只在调试时启用
func (l *SQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 检查是否应该记录日志
	// 根据配置的日志级别决定是否记录
	if !l.shouldLog(Config.LogLevelInfo) {
		return
	}

	// 执行SQL查询
	// 通过回调函数执行SQL，获取SQL语句和影响行数
	sql, rows := fc()
	duration := time.Since(begin)

	// 检查是否是慢查询
	// 如果执行时间超过配置的阈值，标记为慢查询
	isSlowQuery := duration > l.config.SlowThreshold

	// 构建日志字段
	// 包含SQL语句、执行时间、影响行数、是否慢查询等信息
	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),             // SQL语句（可能被截断）
		"duration_ms":   duration.Milliseconds(),        // 执行时间（毫秒）
		"duration_ns":   duration.Nanoseconds(),         // 执行时间（纳秒）
		"rows_affected": rows,                           // 影响行数
		"is_slow_query": isSlowQuery,                    // 是否慢查询
		"timestamp":     begin.Format(time.RFC3339Nano), // 查询开始时间
	}

	// 添加错误信息
	// 如果查询失败，记录错误信息
	if err != nil {
		fields["error"] = err.Error()
	}

	// 添加调用栈信息
	// 如果配置启用，记录调用栈（便于调试，但会影响性能）
	if l.config.IncludeStack {
		fields["stack_trace"] = l.getStackTrace()
	}

	// 选择日志级别
	// 根据查询结果（错误、慢查询、正常）选择不同的日志级别
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
	// 通过LogManagerService异步记录日志
	l.logManager.Log("sql", level, message, fields)

	// 慢查询告警
	// 如果是慢查询，发送额外的告警日志
	if isSlowQuery {
		l.logSlowQueryAlert(sql, duration, rows)
	}
}

// shouldLog 检查是否应该记录日志
//
// 功能说明：
// 1. 根据日志级别配置决定是否记录日志
// 2. 只记录级别大于等于配置级别的日志
// 3. 过滤掉低级别的日志，减少日志量
//
// 日志级别优先级：
// - Debug: 0（最低，用于详细调试信息）
// - Info: 1（一般信息）
// - Warning: 2（警告信息）
// - Error: 3（错误信息）
// - Fatal: 4（最高，致命错误）
//
// 过滤规则：
// - 如果配置级别是Info，则只记录Info、Warning、Error、Fatal
// - 如果配置级别是Warning，则只记录Warning、Error、Fatal
// - 如果配置级别是Error，则只记录Error、Fatal
//
// 使用场景：
// - 生产环境：通常设置为Warning或Error，减少日志量
// - 开发环境：可以设置为Debug，记录所有日志
// - 测试环境：可以设置为Info，记录关键信息
//
// 注意事项：
// - 日志级别配置影响性能和日志量
// - 过低的日志级别会产生大量日志，影响性能
// - 过高的日志级别可能丢失重要信息
func (l *SQLLogger) shouldLog(level Config.LogLevel) bool {
	// 定义日志级别优先级映射
	// 数字越大，优先级越高
	levels := map[Config.LogLevel]int{
		Config.LogLevelDebug:   0,
		Config.LogLevelInfo:    1,
		Config.LogLevelWarning: 2,
		Config.LogLevelError:   3,
		Config.LogLevelFatal:   4,
	}

	// 配置的日志级别优先级映射
	configLevel := map[Config.LogLevel]int{
		Config.LogLevelDebug:   0,
		Config.LogLevelInfo:    1,
		Config.LogLevelWarning: 2,
		Config.LogLevelError:   3,
		Config.LogLevelFatal:   4,
	}

	// 只有当请求的日志级别大于等于配置的日志级别时才记录
	// 例如：如果配置是Warning，则只记录Warning、Error、Fatal
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
//
// 功能说明：
// 1. 记录慢查询的详细信息（SQL、执行时间、阈值等）
// 2. 使用Warning级别记录告警
// 3. 包含告警类型和严重程度信息
//
// 参数说明：
// - sql: SQL查询语句
// - duration: SQL执行时间
// - rows: 影响行数
//
// 告警信息：
// - SQL语句（可能被截断）
// - 执行时间（毫秒和纳秒）
// - 影响行数
// - 慢查询阈值（用于对比）
// - 告警类型（slow_query）
// - 严重程度（warning）
//
// 使用场景：
// - 性能监控：识别慢查询
// - 性能优化：分析慢查询原因
// - 告警通知：通知开发人员优化慢查询
//
// 注意事项：
// - 慢查询阈值由配置决定
// - 告警会记录到日志系统，可以通过监控系统告警
// - 建议定期分析慢查询日志，优化数据库性能
func (l *SQLLogger) logSlowQueryAlert(sql string, duration time.Duration, rows int64) {
	// 构建告警字段
	// 包含SQL、执行时间、阈值等关键信息
	fields := map[string]interface{}{
		"sql":           l.truncateSQL(sql),                    // SQL语句（可能被截断）
		"duration_ms":   duration.Milliseconds(),               // 执行时间（毫秒）
		"duration_ns":   duration.Nanoseconds(),                // 执行时间（纳秒）
		"rows_affected": rows,                                  // 影响行数
		"threshold_ms":  l.config.SlowThreshold.Milliseconds(), // 慢查询阈值（毫秒）
		"alert_type":    "slow_query",                          // 告警类型
		"severity":      "warning",                             // 严重程度
	}

	// 记录告警日志
	// 使用Warning级别，便于监控系统识别和告警
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
	if w.logger == nil {
		return
	}
	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}
	w.logger.logManager.Log("sql", Config.LogLevelInfo, msg, fields)
}

// Warn 记录警告级别日志
func (w *GormLoggerWrapper) Warn(ctx context.Context, msg string, data ...interface{}) {
	if w.logger == nil {
		return
	}
	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}
	w.logger.logManager.Log("sql", Config.LogLevelWarning, msg, fields)
}

// Error 记录错误级别日志
func (w *GormLoggerWrapper) Error(ctx context.Context, msg string, data ...interface{}) {
	if w.logger == nil {
		return
	}
	fields := map[string]interface{}{
		"message": msg,
		"data":    data,
	}
	w.logger.logManager.Log("sql", Config.LogLevelError, msg, fields)
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
