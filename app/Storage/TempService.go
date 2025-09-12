package Storage

import (
	"cloud-platform-api/app/Config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TempService 临时文件服务
type TempService struct {
	config    *Config.StorageConfig
	basePath  string
	tempFiles map[string]*TempFile
	mu        sync.RWMutex
}

// TempFile 临时文件信息
type TempFile struct {
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Size      int64     `json:"size"`
	File      *os.File  `json:"-"`
}

// NewTempService 创建临时文件服务
func NewTempService(config *Config.StorageConfig) *TempService {
	ts := &TempService{
		config:    config,
		basePath:  filepath.Join(config.BasePath, "temp"),
		tempFiles: make(map[string]*TempFile),
	}

	// 确保临时目录存在
	if err := os.MkdirAll(ts.basePath, 0755); err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
	}

	// 启动清理任务
	go ts.startCleanupTask()

	return ts
}

// CreateTempFile 创建临时文件
func (ts *TempService) CreateTempFile(prefix string) (*os.File, error) {
	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), os.Getpid())
	filePath := filepath.Join(ts.basePath, filename)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}

	// 记录临时文件信息
	tempFile := &TempFile{
		Path:      filePath,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时后过期
		File:      file,
	}

	ts.mu.Lock()
	ts.tempFiles[filePath] = tempFile
	ts.mu.Unlock()

	return file, nil
}

// CreateTempFileWithTTL 创建带TTL的临时文件
func (ts *TempService) CreateTempFileWithTTL(prefix string, ttl time.Duration) (*os.File, error) {
	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), os.Getpid())
	filePath := filepath.Join(ts.basePath, filename)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}

	// 记录临时文件信息
	tempFile := &TempFile{
		Path:      filePath,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		File:      file,
	}

	ts.mu.Lock()
	ts.tempFiles[filePath] = tempFile
	ts.mu.Unlock()

	return file, nil
}

// WriteTempFile 写入临时文件
func (ts *TempService) WriteTempFile(prefix string, data []byte) (string, error) {
	file, err := ts.CreateTempFile(prefix)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入数据
	if _, err := file.Write(data); err != nil {
		os.Remove(file.Name())
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 更新文件大小
	ts.mu.Lock()
	if tempFile, exists := ts.tempFiles[file.Name()]; exists {
		tempFile.Size = int64(len(data))
	}
	ts.mu.Unlock()

	return file.Name(), nil
}

// WriteTempFileFromReader 从Reader写入临时文件
func (ts *TempService) WriteTempFileFromReader(prefix string, reader io.Reader) (string, error) {
	file, err := ts.CreateTempFile(prefix)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 复制数据
	size, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(file.Name())
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 更新文件大小
	ts.mu.Lock()
	if tempFile, exists := ts.tempFiles[file.Name()]; exists {
		tempFile.Size = size
	}
	ts.mu.Unlock()

	return file.Name(), nil
}

// ReadTempFile 读取临时文件
func (ts *TempService) ReadTempFile(filePath string) ([]byte, error) {
	// 检查文件是否存在
	ts.mu.RLock()
	tempFile, exists := ts.tempFiles[filePath]
	ts.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("临时文件不存在: %s", filePath)
	}

	// 检查是否过期
	if time.Now().After(tempFile.ExpiresAt) {
		ts.DeleteTempFile(filePath)
		return nil, fmt.Errorf("临时文件已过期: %s", filePath)
	}

	// 读取文件内容
	return os.ReadFile(filePath)
}

// DeleteTempFile 删除临时文件
func (ts *TempService) DeleteTempFile(filePath string) error {
	// 关闭文件句柄
	ts.mu.Lock()
	if tempFile, exists := ts.tempFiles[filePath]; exists {
		if tempFile.File != nil {
			tempFile.File.Close()
		}
		delete(ts.tempFiles, filePath)
	}
	ts.mu.Unlock()

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除临时文件失败: %v", err)
	}

	return nil
}

// GetTempFileInfo 获取临时文件信息
func (ts *TempService) GetTempFileInfo(filePath string) (*TempFile, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tempFile, exists := ts.tempFiles[filePath]
	if !exists {
		return nil, fmt.Errorf("临时文件不存在: %s", filePath)
	}

	return tempFile, nil
}

// ListTempFiles 列出所有临时文件
func (ts *TempService) ListTempFiles() []*TempFile {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	var files []*TempFile
	for _, tempFile := range ts.tempFiles {
		files = append(files, tempFile)
	}

	return files
}

// CleanExpiredFiles 清理过期文件
func (ts *TempService) CleanExpiredFiles() error {
	now := time.Now()
	var expiredFiles []string

	ts.mu.RLock()
	for path, tempFile := range ts.tempFiles {
		if now.After(tempFile.ExpiresAt) {
			expiredFiles = append(expiredFiles, path)
		}
	}
	ts.mu.RUnlock()

	// 删除过期文件
	for _, path := range expiredFiles {
		if err := ts.DeleteTempFile(path); err != nil {
			fmt.Printf("删除过期临时文件失败 %s: %v\n", path, err)
		}
	}

	return nil
}

// CleanAllTempFiles 清理所有临时文件
func (ts *TempService) CleanAllTempFiles() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	var errors []error

	for path, tempFile := range ts.tempFiles {
		// 关闭文件句柄
		if tempFile.File != nil {
			tempFile.File.Close()
		}

		// 删除文件
		if err := os.Remove(path); err != nil {
			errors = append(errors, fmt.Errorf("删除临时文件失败 %s: %v", path, err))
		}
	}

	// 清空映射
	ts.tempFiles = make(map[string]*TempFile)

	if len(errors) > 0 {
		return fmt.Errorf("清理临时文件时发生错误: %v", errors)
	}

	return nil
}

// startCleanupTask 启动清理任务
func (ts *TempService) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ts.CleanExpiredFiles()
		}
	}
}

// GetTempStats 获取临时文件统计信息
func (ts *TempService) GetTempStats() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_files"] = len(ts.tempFiles)

	var totalSize int64
	var expiredCount int
	now := time.Now()

	for _, tempFile := range ts.tempFiles {
		totalSize += tempFile.Size
		if now.After(tempFile.ExpiresAt) {
			expiredCount++
		}
	}

	stats["total_size"] = totalSize
	stats["expired_files"] = expiredCount
	stats["base_path"] = ts.basePath

	return stats
}

// Close 关闭临时文件服务
func (ts *TempService) Close() error {
	return ts.CleanAllTempFiles()
}
