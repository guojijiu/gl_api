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

// ConfigReloader 配置热重载器
type ConfigReloader struct {
	watcher    *fsnotify.Watcher
	configPath string
	callbacks  []ConfigChangeCallback
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// ConfigChangeCallback 配置变更回调函数类型
type ConfigChangeCallback func(config *Config, changeType ConfigChangeType) error

// ConfigChangeType 配置变更类型
type ConfigChangeType int

const (
	ConfigChangeReload ConfigChangeType = iota
	ConfigChangeError
	ConfigChangeFileDeleted
	ConfigChangeFileCreated
)

// NewConfigReloader 创建配置热重载器
func NewConfigReloader(configPath string) (*ConfigReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	reloader := &ConfigReloader{
		watcher:    watcher,
		configPath: configPath,
		callbacks:  make([]ConfigChangeCallback, 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	return reloader, nil
}

// AddCallback 添加配置变更回调
func (cr *ConfigReloader) AddCallback(callback ConfigChangeCallback) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.callbacks = append(cr.callbacks, callback)
}

// Start 开始监听配置变更
func (cr *ConfigReloader) Start() error {
	// 添加配置文件监听
	if err := cr.watcher.Add(cr.configPath); err != nil {
		return fmt.Errorf("failed to add config file to watcher: %v", err)
	}

	// 添加配置文件所在目录监听
	configDir := filepath.Dir(cr.configPath)
	if err := cr.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to add config directory to watcher: %v", err)
	}

	// 启动监听协程
	go cr.watch()

	log.Printf("配置热重载已启动，监听路径: %s", cr.configPath)
	return nil
}

// Stop 停止监听
func (cr *ConfigReloader) Stop() {
	cr.cancel()
	if cr.watcher != nil {
		cr.watcher.Close()
	}
	log.Println("配置热重载已停止")
}

// watch 监听文件变更
func (cr *ConfigReloader) watch() {
	defer cr.watcher.Close()

	for {
		select {
		case event, ok := <-cr.watcher.Events:
			if !ok {
				return
			}
			cr.handleEvent(event)

		case err, ok := <-cr.watcher.Errors:
			if !ok {
				return
			}
			cr.handleError(err)

		case <-cr.ctx.Done():
			return
		}
	}
}

// handleEvent 处理文件事件
func (cr *ConfigReloader) handleEvent(event fsnotify.Event) {
	// 只处理配置文件相关的事件
	if event.Name != cr.configPath {
		return
	}

	log.Printf("检测到配置文件变更: %s, 事件: %s", event.Name, event.Op)

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		cr.reloadConfig(ConfigChangeReload)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		cr.reloadConfig(ConfigChangeFileDeleted)
	case event.Op&fsnotify.Create == fsnotify.Create:
		cr.reloadConfig(ConfigChangeFileCreated)
	}
}

// handleError 处理监听错误
func (cr *ConfigReloader) handleError(err error) {
	log.Printf("配置监听错误: %v", err)
	cr.notifyCallbacks(nil, ConfigChangeError)
}

// reloadConfig 重新加载配置
func (cr *ConfigReloader) reloadConfig(changeType ConfigChangeType) {
	// 检查文件是否存在
	if _, err := os.Stat(cr.configPath); os.IsNotExist(err) {
		log.Printf("配置文件不存在: %s", cr.configPath)
		cr.notifyCallbacks(nil, ConfigChangeFileDeleted)
		return
	}

	// 重新加载配置
	newConfig, err := cr.loadConfig()
	if err != nil {
		log.Printf("配置重新加载失败: %v", err)
		cr.notifyCallbacks(nil, ConfigChangeError)
		return
	}

	log.Printf("配置重新加载成功")
	cr.notifyCallbacks(newConfig, changeType)
}

// loadConfig 加载配置
func (cr *ConfigReloader) loadConfig() (*Config, error) {
	// 重新读取配置文件
	viper.Reset()

	// 设置配置文件
	viper.SetConfigFile(cr.configPath)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 创建新的配置实例
	newConfig := &Config{}

	// 设置默认值
	newConfig.SetDefaults()

	// 绑定环境变量
	newConfig.BindEnvs()

	// 解析配置
	if err := viper.Unmarshal(newConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 验证配置
	if err := ValidateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return newConfig, nil
}

// notifyCallbacks 通知所有回调函数
func (cr *ConfigReloader) notifyCallbacks(config *Config, changeType ConfigChangeType) {
	cr.mu.RLock()
	callbacks := make([]ConfigChangeCallback, len(cr.callbacks))
	copy(callbacks, cr.callbacks)
	cr.mu.RUnlock()

	for _, callback := range callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(config, changeType); err != nil {
				log.Printf("配置变更回调执行失败: %v", err)
			}
		}(callback)
	}
}

// ConfigManager 配置管理器
type ConfigManager struct {
	reloader   *ConfigReloader
	config     *Config
	mu         sync.RWMutex
	lastUpdate time.Time
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	reloader, err := NewConfigReloader(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config reloader: %v", err)
	}

	manager := &ConfigManager{
		reloader: reloader,
	}

	// 添加默认回调
	manager.reloader.AddCallback(manager.onConfigChange)

	return manager, nil
}

// Start 启动配置管理器
func (cm *ConfigManager) Start() error {
	// 初始加载配置
	config, err := cm.reloader.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load initial config: %v", err)
	}

	cm.mu.Lock()
	cm.config = config
	cm.lastUpdate = time.Now()
	cm.mu.Unlock()

	// 启动热重载
	return cm.reloader.Start()
}

// Stop 停止配置管理器
func (cm *ConfigManager) Stop() {
	cm.reloader.Stop()
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// GetLastUpdate 获取最后更新时间
func (cm *ConfigManager) GetLastUpdate() time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastUpdate
}

// AddCallback 添加配置变更回调
func (cm *ConfigManager) AddCallback(callback ConfigChangeCallback) {
	cm.reloader.AddCallback(callback)
}

// onConfigChange 配置变更处理
func (cm *ConfigManager) onConfigChange(config *Config, changeType ConfigChangeType) error {
	if config == nil {
		return nil
	}

	cm.mu.Lock()
	cm.config = config
	cm.lastUpdate = time.Now()
	cm.mu.Unlock()

	// 更新全局配置
	globalConfig = config

	log.Printf("配置已更新，变更类型: %d", changeType)
	return nil
}

// 全局配置管理器
var globalConfigManager *ConfigManager

// InitializeConfigManager 初始化全局配置管理器
func InitializeConfigManager(configPath string) error {
	manager, err := NewConfigManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config manager: %v", err)
	}

	if err := manager.Start(); err != nil {
		return fmt.Errorf("failed to start config manager: %v", err)
	}

	globalConfigManager = manager
	return nil
}

// GetConfigManager 获取全局配置管理器
func GetConfigManager() *ConfigManager {
	return globalConfigManager
}

// StopConfigManager 停止全局配置管理器
func StopConfigManager() {
	if globalConfigManager != nil {
		globalConfigManager.Stop()
	}
}
