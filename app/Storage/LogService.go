package Storage

import (
	"cloud-platform-api/app/Config"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogService 日志服务
type LogService struct {
	config   *Config.LogConfig
	basePath string
}

// NewLogService 创建日志服务
func NewLogService(config *Config.LogConfig) *LogService {
	return &LogService{
		config:   config,
		basePath: config.BasePath,
	}
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// WriteLog 写入日志
func (ls *LogService) WriteLog(level, message string, context map[string]interface{}) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   context,
	}

	// 确定日志文件路径
	logPath := ls.getLogPath(level)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 序列化日志条目
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化日志条目失败: %v", err)
	}

	// 追加到日志文件
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	// 写入日志条目，每行一个JSON对象
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}

	return nil
}

// ReadLogs 读取日志
func (ls *LogService) ReadLogs(level string, date string) ([]LogEntry, error) {
	logPath := ls.getLogPath(level)

	// 如果指定了日期，使用日期路径
	if date != "" {
		logPath = ls.getLogPathWithDate(level, date)
	}

	// 检查文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []LogEntry{}, nil
	}

	// 读取文件内容
	content, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %v", err)
	}

	// 解析日志条目
	var entries []LogEntry
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// 跳过无法解析的行
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetLogFiles 获取日志文件列表
func (ls *LogService) GetLogFiles(level string) ([]string, error) {
	logDir := filepath.Join(ls.basePath, "logs", level)

	// 检查目录是否存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// 读取目录内容
	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("读取日志目录失败: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			logFiles = append(logFiles, file.Name())
		}
	}

	return logFiles, nil
}

// CleanOldLogs 清理旧日志
func (ls *LogService) CleanOldLogs(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// 清理各级别的日志
	levels := []string{"debug", "info", "warning", "error", "fatal"}

	for _, level := range levels {
		logDir := filepath.Join(ls.basePath, "logs", level)

		if err := ls.cleanDirectory(logDir, cutoffTime); err != nil {
			return fmt.Errorf("清理日志目录失败 %s: %v", level, err)
		}
	}

	return nil
}

// cleanDirectory 清理目录中的旧文件
func (ls *LogService) cleanDirectory(dir string, cutoffTime time.Time) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	// 读取目录内容
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 获取文件信息以检查修改时间
		info, err := file.Info()
		if err != nil {
			continue
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(dir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("删除文件失败 %s: %v", filePath, err)
			}
		}
	}

	return nil
}

// getLogPath 获取日志文件路径
func (ls *LogService) getLogPath(level string) string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(ls.basePath, "logs", level, fmt.Sprintf("%s-%s.log", level, date))
}

// getLogPathWithDate 获取指定日期的日志文件路径
func (ls *LogService) getLogPathWithDate(level, date string) string {
	return filepath.Join(ls.basePath, "logs", level, fmt.Sprintf("%s-%s.log", level, date))
}

// GetLogs 获取日志列表
func (ls *LogService) GetLogs(level string, date string) ([]string, error) {
	logs, err := ls.ReadLogs(level, date)
	if err != nil {
		return nil, err
	}

	// 转换为字符串数组
	var logStrings []string
	for _, log := range logs {
		logBytes, err := json.Marshal(log)
		if err != nil {
			continue
		}
		logStrings = append(logStrings, string(logBytes))
	}

	return logStrings, nil
}

// GetLogStats 获取日志统计信息
func (ls *LogService) GetLogStats() map[string]interface{} {
	stats := make(map[string]interface{})

	levels := []string{"debug", "info", "warning", "error", "fatal"}

	for _, level := range levels {
		levelStats := make(map[string]interface{})

		// 获取今日日志数量
		todayLogs, _ := ls.ReadLogs(level, time.Now().Format("2006-01-02"))
		levelStats["today_count"] = len(todayLogs)

		// 获取日志文件数量
		files, _ := ls.GetLogFiles(level)
		levelStats["file_count"] = len(files)

		// 获取目录大小
		logDir := filepath.Join(ls.basePath, "logs", level)
		if size, err := ls.getDirectorySize(logDir); err == nil {
			levelStats["size_bytes"] = size
		}

		stats[level] = levelStats
	}

	return stats
}

// getDirectorySize 获取目录大小
func (ls *LogService) getDirectorySize(dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}
