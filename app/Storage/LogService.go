package Storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LogService 日志服务
// 功能说明：
// 1. 提供基于日期的日志记录功能
// 2. 支持不同级别的日志（INFO、WARNING、ERROR、DEBUG）
// 3. 使用JSON格式存储日志条目
// 4. 支持日志查询和过滤
// 5. 提供日志清理和统计功能
type LogService struct {
	logsPath string
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// LogFileInfo 日志文件信息
type LogFileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// NewLogService 创建新的日志服务实例
func NewLogService(logsPath string) *LogService {
	return &LogService{
		logsPath: logsPath,
	}
}

// Log 记录日志
// 功能说明：
// 1. 创建包含时间戳、级别、消息和上下文的日志条目
// 2. 根据上下文中的category字段路由到对应的日志子目录
// 3. 按日期生成日志文件（YYYY-MM-DD.log格式）
// 4. 使用JSON格式序列化日志条目
// 5. 追加到对应的日志文件中
// 6. 自动创建日志目录（如果不存在）
func (ls *LogService) Log(level string, message string, context map[string]interface{}) error {
	// 创建日志条目
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
		Context:   context,
	}

	// 确定日志子目录
	logSubDir := ls.determineLogSubDir(context)

	// 构建完整的日志路径
	fullLogPath := filepath.Join(ls.logsPath, logSubDir)

	// 确保日志目录存在
	if err := os.MkdirAll(fullLogPath, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成日志文件名（按日期）
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(fullLogPath, fmt.Sprintf("%s-%s.log", logSubDir, today))

	// 序列化日志条目
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化日志失败: %v", err)
	}

	// 追加到日志文件
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	// 写入日志条目，每条日志一行
	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}

	return nil
}

// determineLogSubDir 根据上下文信息确定日志子目录
// 功能说明：
// 1. 检查上下文中的category字段
// 2. 根据category映射到对应的日志子目录
// 3. 如果没有匹配的category或category为空，则使用system目录
// 4. 支持预定义的日志分类：access, audit, business, errors, requests, security, sql, system
func (ls *LogService) determineLogSubDir(context map[string]interface{}) string {
	// 预定义的日志分类映射
	categoryMap := map[string]string{
		"access":   "access",
		"audit":    "audit",
		"business": "business",
		"error":    "errors",
		"errors":   "errors",
		"request":  "requests",
		"requests": "requests",
		"security": "security",
		"sql":      "sql",
		"system":   "system",
	}

	// 从上下文中获取category
	if category, exists := context["category"]; exists {
		if categoryStr, ok := category.(string); ok {
			// 转换为小写进行匹配
			categoryLower := strings.ToLower(categoryStr)
			if subDir, found := categoryMap[categoryLower]; found {
				return subDir
			}
		}
	}

	// 默认使用system目录
	return "system"
}

// GetLogs 获取指定级别和日期的日志
// 功能说明：
// 1. 读取指定日期的日志文件
// 2. 支持从指定子目录读取日志
// 3. 解析JSON格式的日志条目
// 4. 按级别过滤日志（如果指定了级别）
// 5. 返回过滤后的日志列表
// 6. 如果文件不存在，返回空列表
func (ls *LogService) GetLogs(level string, date string) ([]string, error) {
	return ls.GetLogsFromSubDir(level, date, "")
}

// GetLogsFromSubDir 从指定子目录获取日志
// 功能说明：
// 1. 从指定的日志子目录读取日志文件
// 2. 如果subDir为空，则从所有子目录读取
// 3. 解析JSON格式的日志条目
// 4. 按级别过滤日志（如果指定了级别）
// 5. 返回过滤后的日志列表
func (ls *LogService) GetLogsFromSubDir(level string, date string, subDir string) ([]string, error) {
	var logs []string

	// 如果指定了子目录，只从该目录读取
	if subDir != "" {
		return ls.readLogsFromDir(level, date, subDir)
	}

	// 如果没有指定子目录，从所有子目录读取
	subDirs := []string{"access", "audit", "business", "errors", "requests", "security", "sql", "system"}

	for _, dir := range subDirs {
		dirLogs, err := ls.readLogsFromDir(level, date, dir)
		if err != nil {
			continue // 跳过无法读取的目录
		}
		logs = append(logs, dirLogs...)
	}

	return logs, nil
}

// readLogsFromDir 从指定目录读取日志
func (ls *LogService) readLogsFromDir(level string, date string, subDir string) ([]string, error) {
	var logs []string

	// 构建日志文件路径
	logFile := filepath.Join(ls.logsPath, subDir, fmt.Sprintf("%s.log", date))

	// 检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return logs, nil // 文件不存在返回空数组
	}

	// 读取日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %v", err)
	}

	// 按行分割
	lines := strings.Split(string(content), "\n")

	// 过滤指定级别的日志
	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // 跳过无效的日志行
		}

		// 如果指定了级别，则只返回该级别的日志
		if level == "" || entry.Level == level {
			logs = append(logs, line)
		}
	}

	return logs, nil
}

// LogInfo 记录信息日志
func (ls *LogService) LogInfo(message string, context map[string]interface{}) error {
	return ls.Log("INFO", message, context)
}

// LogWarning 记录警告日志
func (ls *LogService) LogWarning(message string, context map[string]interface{}) error {
	return ls.Log("WARNING", message, context)
}

// LogError 记录错误日志
func (ls *LogService) LogError(message string, context map[string]interface{}) error {
	return ls.Log("ERROR", message, context)
}

// LogDebug 记录调试日志
func (ls *LogService) LogDebug(message string, context map[string]interface{}) error {
	return ls.Log("DEBUG", message, context)
}

// CleanupLogs 清理过期和过多的日志文件
// 功能说明：
// 1. 清理超过指定天数的日志文件
// 2. 限制日志文件的总大小
// 3. 保留最近的日志文件
// 4. 返回清理的文件数量和大小
func (ls *LogService) CleanupLogs(maxDays int, maxSizeMB int64) (int, int64, error) {
	cleanedCount := 0
	cleanedSize := int64(0)

	// 获取所有日志文件
	files, err := ls.GetAllLogFiles()
	if err != nil {
		return 0, 0, err
	}

	// 计算总大小
	totalSize := int64(0)
	for _, file := range files {
		totalSize += file.Size
	}

	// 按修改时间排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	// 清理过期的文件
	cutoffTime := time.Now().AddDate(0, 0, -maxDays)
	for _, file := range files {
		if file.ModTime.Before(cutoffTime) {
			err := os.Remove(filepath.Join(ls.logsPath, file.Name))
			if err == nil {
				cleanedCount++
				cleanedSize += file.Size
			}
		}
	}

	// 如果总大小仍然超过限制，删除最旧的文件
	if totalSize-cleanedSize > maxSizeMB*1024*1024 {
		for _, file := range files {
			if totalSize-cleanedSize <= maxSizeMB*1024*1024 {
				break
			}
			if file.ModTime.After(cutoffTime) { // 只删除未过期的文件
				err := os.Remove(filepath.Join(ls.logsPath, file.Name))
				if err == nil {
					cleanedCount++
					cleanedSize += file.Size
					totalSize -= file.Size
				}
			}
		}
	}

	return cleanedCount, cleanedSize, nil
}

// GetLogStats 获取日志统计信息
// 功能说明：
// 1. 统计日志文件数量和总大小
// 2. 按级别统计日志数量
// 3. 获取最近的日志文件信息
// 4. 返回详细的统计信息
func (ls *LogService) GetLogStats() (map[string]interface{}, error) {
	files, err := ls.GetAllLogFiles()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_files":  len(files),
		"total_size":   int64(0),
		"levels":       make(map[string]int),
		"recent_files": []map[string]interface{}{},
	}

	// 按级别统计
	levelStats := make(map[string]int)

	for _, file := range files {
		stats["total_size"] = stats["total_size"].(int64) + file.Size

		// 从文件名提取级别信息
		if strings.Contains(file.Name, "INFO") {
			levelStats["INFO"]++
		} else if strings.Contains(file.Name, "WARNING") {
			levelStats["WARNING"]++
		} else if strings.Contains(file.Name, "ERROR") {
			levelStats["ERROR"]++
		} else if strings.Contains(file.Name, "DEBUG") {
			levelStats["DEBUG"]++
		}
	}

	stats["levels"] = levelStats

	// 获取最近的5个文件
	recentFiles := []map[string]interface{}{}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	for i, file := range files {
		if i >= 5 {
			break
		}
		recentFiles = append(recentFiles, map[string]interface{}{
			"name":     file.Name,
			"size":     file.Size,
			"mod_time": file.ModTime.Format("2006-01-02 15:04:05"),
		})
	}

	stats["recent_files"] = recentFiles

	return stats, nil
}

// GetAllLogFiles 获取所有日志文件信息
// 功能说明：
// 1. 扫描日志目录及其所有子目录下的文件
// 2. 获取文件大小和修改时间
// 3. 返回文件信息列表，包含子目录信息
func (ls *LogService) GetAllLogFiles() ([]LogFileInfo, error) {
	var files []LogFileInfo

	// 扫描所有子目录
	subDirs := []string{"access", "audit", "business", "errors", "requests", "security", "sql", "system"}

	for _, subDir := range subDirs {
		dirPath := filepath.Join(ls.logsPath, subDir)

		// 检查子目录是否存在
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filePath := filepath.Join(dirPath, entry.Name())
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}

			// 在文件名中包含子目录信息
			relativeName := filepath.Join(subDir, entry.Name())

			files = append(files, LogFileInfo{
				Name:    relativeName,
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
	}

	return files, nil
}
