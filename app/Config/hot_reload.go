package Config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// HotReloadManager 配置热重载管理器
type HotReloadManager struct {
	configPath      string
	watcher         *fsnotify.Watcher
	reloadCallbacks []func(*Config)
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	isWatching      bool
}

// NewHotReloadManager 创建配置热重载管理器
func NewHotReloadManager(configPath string) *HotReloadManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &HotReloadManager{
		configPath:      configPath,
		reloadCallbacks: make([]func(*Config), 0),
		ctx:             ctx,
		cancel:          cancel,
		isWatching:      false,
	}
}

// StartWatching 开始监控配置文件变化
func (hrm *HotReloadManager) StartWatching() error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	if hrm.isWatching {
		return fmt.Errorf("配置热重载已经在运行")
	}

	// 创建文件监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %v", err)
	}
	hrm.watcher = watcher

	// 监控配置文件目录
	configDir := filepath.Dir(hrm.configPath)
	if err := watcher.Add(configDir); err != nil {
		watcher.Close()
		return fmt.Errorf("添加监控目录失败: %v", err)
	}

	// 监控配置文件
	if err := watcher.Add(hrm.configPath); err != nil {
		watcher.Close()
		return fmt.Errorf("添加监控文件失败: %v", err)
	}

	hrm.isWatching = true

	// 启动监控goroutine
	go hrm.watchFiles()

	log.Println("配置热重载已启动")
	return nil
}

// StopWatching 停止监控配置文件变化
func (hrm *HotReloadManager) StopWatching() error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	if !hrm.isWatching {
		return fmt.Errorf("配置热重载未在运行")
	}

	// 取消上下文
	hrm.cancel()

	// 关闭文件监控器
	if hrm.watcher != nil {
		if err := hrm.watcher.Close(); err != nil {
			return fmt.Errorf("关闭文件监控器失败: %v", err)
		}
	}

	hrm.isWatching = false
	log.Println("配置热重载已停止")
	return nil
}

// watchFiles 监控文件变化
func (hrm *HotReloadManager) watchFiles() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-hrm.watcher.Events:
			if !ok {
				return
			}

			// 检查是否是配置文件变化
			if hrm.isConfigFile(event.Name) {
				hrm.handleConfigChange(event)
			}

		case err, ok := <-hrm.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("配置文件监控错误: %v", err)

		case <-hrm.ctx.Done():
			return

		case <-ticker.C:
			// 定期检查配置文件是否存在
			if _, err := os.Stat(hrm.configPath); os.IsNotExist(err) {
				log.Printf("配置文件不存在: %s", hrm.configPath)
			}
		}
	}
}

// isConfigFile 检查是否是配置文件
func (hrm *HotReloadManager) isConfigFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml" || ext == ".env"
}

// handleConfigChange 处理配置文件变化
func (hrm *HotReloadManager) handleConfigChange(event fsnotify.Event) {
	// 防抖处理，避免频繁重载
	time.Sleep(100 * time.Millisecond)

	// 检查文件是否仍然存在
	if _, err := os.Stat(event.Name); os.IsNotExist(err) {
		return
	}

	log.Printf("检测到配置文件变化: %s", event.Name)

	// 重新加载配置
	if err := hrm.reloadConfig(); err != nil {
		log.Printf("配置重载失败: %v", err)
		return
	}

	log.Println("配置重载成功")
}

// reloadConfig 重新加载配置
func (hrm *HotReloadManager) reloadConfig() error {
	// 重新读取配置文件
	viper.SetConfigFile(hrm.configPath)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 重新解析配置
	var newConfig Config
	if err := viper.Unmarshal(&newConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 验证配置（这里可以添加配置验证逻辑）
	// if err := newConfig.Validate(); err != nil {
	//     return fmt.Errorf("配置验证失败: %v", err)
	// }

	// 更新全局配置
	globalConfig = &newConfig

	// 执行重载回调
	hrm.mutex.RLock()
	callbacks := make([]func(*Config), len(hrm.reloadCallbacks))
	copy(callbacks, hrm.reloadCallbacks)
	hrm.mutex.RUnlock()

	for _, callback := range callbacks {
		go func(cb func(*Config)) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("配置重载回调执行失败: %v", r)
				}
			}()
			cb(&newConfig)
		}(callback)
	}

	return nil
}

// AddReloadCallback 添加配置重载回调
func (hrm *HotReloadManager) AddReloadCallback(callback func(*Config)) {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	hrm.reloadCallbacks = append(hrm.reloadCallbacks, callback)
}

// RemoveReloadCallback 移除配置重载回调
func (hrm *HotReloadManager) RemoveReloadCallback(callback func(*Config)) {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	// 由于Go中函数不能直接比较，我们需要使用不同的方法
	// 这里我们通过重新构建切片来移除回调
	newCallbacks := make([]func(*Config), 0, len(hrm.reloadCallbacks))
	for _, cb := range hrm.reloadCallbacks {
		// 使用反射或其他方法来比较函数，但这里我们简化处理
		// 在实际应用中，可能需要使用回调ID或其他标识符
		if fmt.Sprintf("%p", cb) != fmt.Sprintf("%p", callback) {
			newCallbacks = append(newCallbacks, cb)
		}
	}
	hrm.reloadCallbacks = newCallbacks
}

// IsWatching 检查是否正在监控
func (hrm *HotReloadManager) IsWatching() bool {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()
	return hrm.isWatching
}

// GetConfigPath 获取配置文件路径
func (hrm *HotReloadManager) GetConfigPath() string {
	return hrm.configPath
}

// SetConfigPath 设置配置文件路径
func (hrm *HotReloadManager) SetConfigPath(path string) error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	if hrm.isWatching {
		return fmt.Errorf("配置热重载正在运行，无法更改配置文件路径")
	}

	hrm.configPath = path
	return nil
}

// GetReloadCallbacksCount 获取重载回调数量
func (hrm *HotReloadManager) GetReloadCallbacksCount() int {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()
	return len(hrm.reloadCallbacks)
}

// 全局热重载管理器
var globalHotReloadManager *HotReloadManager

// InitHotReload 初始化配置热重载
func InitHotReload(configPath string) error {
	if globalHotReloadManager != nil {
		return fmt.Errorf("配置热重载已经初始化")
	}

	globalHotReloadManager = NewHotReloadManager(configPath)
	return globalHotReloadManager.StartWatching()
}

// StopHotReload 停止配置热重载
func StopHotReload() error {
	if globalHotReloadManager == nil {
		return fmt.Errorf("配置热重载未初始化")
	}

	err := globalHotReloadManager.StopWatching()
	globalHotReloadManager = nil
	return err
}

// GetHotReloadManager 获取全局热重载管理器
func GetHotReloadManager() *HotReloadManager {
	return globalHotReloadManager
}

// AddGlobalReloadCallback 添加全局配置重载回调
func AddGlobalReloadCallback(callback func(*Config)) {
	if globalHotReloadManager != nil {
		globalHotReloadManager.AddReloadCallback(callback)
	}
}

// RemoveGlobalReloadCallback 移除全局配置重载回调
func RemoveGlobalReloadCallback(callback func(*Config)) {
	if globalHotReloadManager != nil {
		globalHotReloadManager.RemoveReloadCallback(callback)
	}
}
