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
	stopChan  chan struct{}
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
		stopChan:  make(chan struct{}),
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
//
// 功能说明：
// 1. 定期清理过期的临时文件，释放磁盘空间
// 2. 在独立的goroutine中运行，不阻塞主流程
// 3. 支持优雅停止，通过stopChan通道控制
//
// 实现原理：
// - 使用time.Ticker定时触发清理任务（默认每小时一次）
// - 使用select同时监听定时器和停止信号
// - 当收到停止信号时，立即退出goroutine，避免资源泄漏
//
// 优雅停止机制：
// - Close()方法会关闭stopChan通道
// - 关闭的通道会立即返回零值，select会立即选择该case
// - goroutine可以立即退出，不会无限阻塞
//
// 注意事项：
// - defer ticker.Stop()确保ticker资源被释放
// - 清理任务在后台运行，不会影响服务的正常使用
// - 如果服务异常退出，goroutine也会自动结束（进程退出）
func (ts *TempService) startCleanupTask() {
	// 创建定时器，每小时触发一次清理任务
	// Ticker会在指定时间间隔重复发送时间到通道
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop() // 确保ticker资源被释放，避免泄漏

	// 无限循环，持续监听定时器和停止信号
	for {
		select {
		case <-ticker.C:
			// 定时器触发，执行清理过期文件的操作
			// CleanExpiredFiles会遍历所有临时文件，删除已过期的文件
			ts.CleanExpiredFiles()
		case <-ts.stopChan:
			// 收到停止信号（stopChan被关闭）
			// 关闭的通道会立即返回零值，select会立即选择该case
			// 这样可以立即退出goroutine，实现优雅停止
			return
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
//
// 功能说明：
// 1. 停止清理goroutine，避免goroutine泄漏
// 2. 清理所有临时文件，释放磁盘空间
// 3. 释放相关资源
//
// 优雅关闭流程：
// 1. 关闭stopChan通道，通知清理goroutine停止
// 2. 清理所有临时文件（包括未过期的）
// 3. 返回清理过程中的错误（如果有）
//
// 注意事项：
// - close(chan)操作是幂等的，多次关闭会panic
// - 应该在服务关闭时调用，确保资源正确释放
// - 清理所有文件可能会删除正在使用的临时文件，需要谨慎
func (ts *TempService) Close() error {
	// 关闭stopChan通道，发送停止信号给清理goroutine
	// 注意：关闭通道后，所有等待该通道的goroutine都会立即收到零值
	// 这会导致startCleanupTask中的select立即选择stopChan case并退出
	// 如果通道已经关闭，再次关闭会panic，所以这里假设只调用一次
	close(ts.stopChan)

	// 清理所有临时文件（包括未过期的）
	// 这会在服务关闭时释放所有临时文件占用的磁盘空间
	// 注意：可能会删除正在使用的临时文件，需要确保调用时机正确
	return ts.CleanAllTempFiles()
}
