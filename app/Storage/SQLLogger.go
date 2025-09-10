package Storage

import (
	"database/sql"
	"time"
)

// SQLLogger SQL日志记录器接口
// 功能说明：
// 1. 定义SQL日志记录的标准接口
// 2. 记录SQL查询语句、参数、执行时间和错误信息
// 3. 支持事务操作的日志记录
type SQLLogger interface {
	LogQuery(query string, args []interface{}, duration time.Duration, error error)
}

// SQLLoggerImpl SQL日志记录器实现
// 功能说明：
// 1. 实现SQLLogger接口
// 2. 使用StorageManager记录SQL日志
// 3. 支持不同日志级别的自动选择
type SQLLoggerImpl struct {
	StorageManager *StorageManager
}

// NewSQLLogger 创建新的SQL日志记录器
func NewSQLLogger(storageManager *StorageManager) *SQLLoggerImpl {
	return &SQLLoggerImpl{
		StorageManager: storageManager,
	}
}

// LogQuery 记录SQL查询日志
// 功能说明：
// 1. 根据是否有错误自动选择日志级别（ERROR/INFO）
// 2. 记录SQL查询语句、参数、执行时间
// 3. 记录错误信息（如果有）
// 4. 使用JSON格式存储，便于后续分析
func (sl *SQLLoggerImpl) LogQuery(query string, args []interface{}, duration time.Duration, error error) {
	// 确定日志级别
	var level string
	if error != nil {
		level = "ERROR"
	} else {
		level = "INFO"
	}

	// 构建SQL日志数据
	sqlLog := map[string]interface{}{
		"level":       level,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"query":       query,
		"args":        args,
		"duration_ms": duration.Milliseconds(),
		"duration_ns": duration.Nanoseconds(),
		"has_error":   error != nil,
	}

	if error != nil {
		sqlLog["error"] = error.Error()
		sl.StorageManager.LogError("SQL查询执行失败", sqlLog)
	} else {
		sl.StorageManager.LogInfo("SQL查询执行成功", map[string]interface{}{
			"category": "sql",
			"query":    sqlLog["query"],
			"duration": sqlLog["duration"],
			"rows":     sqlLog["rows"],
		})
	}
}

// WrapDB 包装数据库连接，添加SQL日志记录
// 功能说明：
// 1. 包装原始的*sql.DB对象
// 2. 为所有数据库操作添加日志记录
// 3. 保持原有接口不变，透明地添加日志功能
func WrapDB(db *sql.DB, logger *SQLLoggerImpl) *LoggedDB {
	return &LoggedDB{
		DB:     db,
		logger: logger,
	}
}

// LoggedDB 带日志记录的数据库连接
// 功能说明：
// 1. 继承*sql.DB的所有功能
// 2. 为Query、QueryRow、Exec、Begin等方法添加日志记录
// 3. 记录每个操作的执行时间和结果
type LoggedDB struct {
	*sql.DB
	logger *SQLLoggerImpl
}

// Query 执行查询并记录日志
// 功能说明：
// 1. 记录查询开始时间
// 2. 执行原始查询操作
// 3. 计算执行时间并记录日志
// 4. 返回查询结果和错误信息
func (ldb *LoggedDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := ldb.DB.Query(query, args...)
	duration := time.Since(startTime)

	ldb.logger.LogQuery(query, args, duration, err)
	return rows, err
}

// QueryRow 执行单行查询并记录日志
// 功能说明：
// 1. 记录查询开始时间
// 2. 执行原始查询操作
// 3. 计算执行时间并记录日志
// 4. 注意：QueryRow不返回错误，错误会在Scan时出现
func (ldb *LoggedDB) QueryRow(query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := ldb.DB.QueryRow(query, args...)
	duration := time.Since(startTime)

	// 注意：QueryRow不返回错误，错误会在Scan时出现
	// 我们记录查询开始，但无法在这里检测到错误
	ldb.logger.LogQuery(query, args, duration, nil)
	return row
}

// Exec 执行SQL语句并记录日志
// 功能说明：
// 1. 记录执行开始时间
// 2. 执行原始SQL语句
// 3. 计算执行时间并记录日志
// 4. 返回执行结果和错误信息
func (ldb *LoggedDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := ldb.DB.Exec(query, args...)
	duration := time.Since(startTime)

	ldb.logger.LogQuery(query, args, duration, err)
	return result, err
}

// Begin 开始事务并记录日志
// 功能说明：
// 1. 记录事务开始时间
// 2. 开始数据库事务
// 3. 计算执行时间并记录日志
// 4. 返回带日志记录的事务对象
func (ldb *LoggedDB) Begin() (*LoggedTx, error) {
	startTime := time.Now()
	tx, err := ldb.DB.Begin()
	duration := time.Since(startTime)

	if err != nil {
		ldb.logger.LogQuery("BEGIN TRANSACTION", nil, duration, err)
		return nil, err
	}

	ldb.logger.LogQuery("BEGIN TRANSACTION", nil, duration, nil)
	return &LoggedTx{
		Tx:     tx,
		logger: ldb.logger,
	}, nil
}

// LoggedTx 带日志记录的事务
// 功能说明：
// 1. 继承*sql.Tx的所有功能
// 2. 为事务内的所有操作添加日志记录
// 3. 记录事务的提交和回滚操作
type LoggedTx struct {
	*sql.Tx
	logger *SQLLoggerImpl
}

// Query 事务查询并记录日志
// 功能说明：
// 1. 在查询前添加"TX:"前缀以标识事务操作
// 2. 记录查询开始时间和执行时间
// 3. 记录查询结果和错误信息
func (ltx *LoggedTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := ltx.Tx.Query(query, args...)
	duration := time.Since(startTime)

	ltx.logger.LogQuery("TX: "+query, args, duration, err)
	return rows, err
}

// QueryRow 事务单行查询并记录日志
// 功能说明：
// 1. 在查询前添加"TX:"前缀以标识事务操作
// 2. 记录查询开始时间和执行时间
// 3. 注意：QueryRow不返回错误，错误会在Scan时出现
func (ltx *LoggedTx) QueryRow(query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := ltx.Tx.QueryRow(query, args...)
	duration := time.Since(startTime)

	ltx.logger.LogQuery("TX: "+query, args, duration, nil)
	return row
}

// Exec 事务执行SQL语句并记录日志
// 功能说明：
// 1. 在SQL语句前添加"TX:"前缀以标识事务操作
// 2. 记录执行开始时间和执行时间
// 3. 记录执行结果和错误信息
func (ltx *LoggedTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := ltx.Tx.Exec(query, args...)
	duration := time.Since(startTime)

	ltx.logger.LogQuery("TX: "+query, args, duration, err)
	return result, err
}

// Commit 提交事务并记录日志
// 功能说明：
// 1. 记录提交开始时间
// 2. 提交数据库事务
// 3. 计算执行时间并记录日志
// 4. 返回提交结果和错误信息
func (ltx *LoggedTx) Commit() error {
	startTime := time.Now()
	err := ltx.Tx.Commit()
	duration := time.Since(startTime)

	ltx.logger.LogQuery("COMMIT TRANSACTION", nil, duration, err)
	return err
}

// Rollback 回滚事务并记录日志
// 功能说明：
// 1. 记录回滚开始时间
// 2. 回滚数据库事务
// 3. 计算执行时间并记录日志
// 4. 返回回滚结果和错误信息
func (ltx *LoggedTx) Rollback() error {
	startTime := time.Now()
	err := ltx.Tx.Rollback()
	duration := time.Since(startTime)

	ltx.logger.LogQuery("ROLLBACK TRANSACTION", nil, duration, err)
	return err
}
