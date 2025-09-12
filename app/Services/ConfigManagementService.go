package Services

import (
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigManagementService 配置管理服务
type ConfigManagementService struct {
	BaseService
	storageManager  *Storage.StorageManager
	configPath      string
	backupPath      string
	configHistory   []ConfigSnapshot
	mutex           sync.RWMutex
	validationRules map[string]ConfigValidationRule
	changeCallbacks []ConfigChangeCallback
}

// ConfigSnapshot 配置快照
type ConfigSnapshot struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Checksum    string                 `json:"checksum"`
}

// ConfigValidationRule 配置验证规则
type ConfigValidationRule struct {
	Field       string
	Required    bool
	Type        string
	Min         interface{}
	Max         interface{}
	Pattern     string
	CustomCheck func(interface{}) error
}

// ConfigChangeCallback 配置变更回调
type ConfigChangeCallback func(oldConfig, newConfig map[string]interface{}) error

// NewConfigManagementService 创建配置管理服务
func NewConfigManagementService(storageManager *Storage.StorageManager, configPath string) *ConfigManagementService {
	return &ConfigManagementService{
		storageManager:  storageManager,
		configPath:      configPath,
		backupPath:      filepath.Join(filepath.Dir(configPath), "backups"),
		configHistory:   make([]ConfigSnapshot, 0),
		validationRules: make(map[string]ConfigValidationRule),
		changeCallbacks: make([]ConfigChangeCallback, 0),
	}
}

// LoadConfig 加载配置
func (cms *ConfigManagementService) LoadConfig() (map[string]interface{}, error) {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(cms.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := cms.validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return config, nil
}

// SaveConfig 保存配置
func (cms *ConfigManagementService) SaveConfig(config map[string]interface{}, description string) error {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	// 验证配置
	if err := cms.validateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 创建备份
	if err := cms.createBackup(); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	// 保存配置
	if err := cms.saveConfigToFile(config); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	// 创建快照
	snapshot := ConfigSnapshot{
		ID:          fmt.Sprintf("config_%d", time.Now().Unix()),
		Timestamp:   time.Now(),
		Config:      config,
		Description: description,
		Version:     cms.getConfigVersion(config),
		Checksum:    cms.calculateChecksum(config),
	}

	cms.configHistory = append(cms.configHistory, snapshot)

	// 执行变更回调
	for _, callback := range cms.changeCallbacks {
		if err := callback(nil, config); err != nil {
			cms.storageManager.LogError("配置变更回调执行失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	cms.storageManager.LogInfo("配置已保存", map[string]interface{}{
		"description": description,
		"version":     snapshot.Version,
		"timestamp":   snapshot.Timestamp,
	})

	return nil
}

// validateConfig 验证配置
func (cms *ConfigManagementService) validateConfig(config map[string]interface{}) error {
	for field, rule := range cms.validationRules {
		value, exists := config[field]

		// 检查必填字段
		if rule.Required && !exists {
			return fmt.Errorf("必填字段 %s 缺失", field)
		}

		// 如果字段不存在且不是必填的，跳过验证
		if !exists {
			continue
		}

		// 类型验证
		if err := cms.validateFieldType(value, rule.Type); err != nil {
			return fmt.Errorf("字段 %s 类型验证失败: %v", field, err)
		}

		// 范围验证
		if err := cms.validateFieldRange(value, rule); err != nil {
			return fmt.Errorf("字段 %s 范围验证失败: %v", field, err)
		}

		// 自定义验证
		if rule.CustomCheck != nil {
			if err := rule.CustomCheck(value); err != nil {
				return fmt.Errorf("字段 %s 自定义验证失败: %v", field, err)
			}
		}
	}

	return nil
}

// validateFieldType 验证字段类型
func (cms *ConfigManagementService) validateFieldType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("期望字符串类型，实际类型: %T", value)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("期望整数类型，实际类型: %T", value)
		}
	case "float":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("期望浮点数类型，实际类型: %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("期望布尔类型，实际类型: %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("期望数组类型，实际类型: %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("期望对象类型，实际类型: %T", value)
		}
	}
	return nil
}

// validateFieldRange 验证字段范围
func (cms *ConfigManagementService) validateFieldRange(value interface{}, rule ConfigValidationRule) error {
	if rule.Min != nil {
		if err := cms.compareValue(value, rule.Min, ">="); err != nil {
			return fmt.Errorf("值小于最小值: %v", err)
		}
	}

	if rule.Max != nil {
		if err := cms.compareValue(value, rule.Max, "<="); err != nil {
			return fmt.Errorf("值大于最大值: %v", err)
		}
	}

	return nil
}

// compareValue 比较值
func (cms *ConfigManagementService) compareValue(value, threshold interface{}, operator string) error {
	switch v := value.(type) {
	case int:
		t, ok := threshold.(int)
		if !ok {
			return fmt.Errorf("阈值类型不匹配")
		}
		switch operator {
		case ">=":
			if v < t {
				return fmt.Errorf("%d < %d", v, t)
			}
		case "<=":
			if v > t {
				return fmt.Errorf("%d > %d", v, t)
			}
		}
	case float64:
		t, ok := threshold.(float64)
		if !ok {
			return fmt.Errorf("阈值类型不匹配")
		}
		switch operator {
		case ">=":
			if v < t {
				return fmt.Errorf("%.2f < %.2f", v, t)
			}
		case "<=":
			if v > t {
				return fmt.Errorf("%.2f > %.2f", v, t)
			}
		}
	case string:
		t, ok := threshold.(string)
		if !ok {
			return fmt.Errorf("阈值类型不匹配")
		}
		switch operator {
		case ">=":
			if v < t {
				return fmt.Errorf("%s < %s", v, t)
			}
		case "<=":
			if v > t {
				return fmt.Errorf("%s > %s", v, t)
			}
		}
	}
	return nil
}

// createBackup 创建备份
func (cms *ConfigManagementService) createBackup() error {
	// 确保备份目录存在
	if err := os.MkdirAll(cms.backupPath, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 读取当前配置
	data, err := os.ReadFile(cms.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 创建备份文件
	backupFile := filepath.Join(cms.backupPath, fmt.Sprintf("config_backup_%d.json", time.Now().Unix()))
	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}

	return nil
}

// saveConfigToFile 保存配置到文件
func (cms *ConfigManagementService) saveConfigToFile(config map[string]interface{}) error {
	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(cms.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// getConfigVersion 获取配置版本
func (cms *ConfigManagementService) getConfigVersion(config map[string]interface{}) string {
	if version, exists := config["version"]; exists {
		if v, ok := version.(string); ok {
			return v
		}
	}
	return "1.0.0"
}

// calculateChecksum 计算配置校验和
func (cms *ConfigManagementService) calculateChecksum(config map[string]interface{}) string {
	data, _ := json.Marshal(config)
	return fmt.Sprintf("%x", data)
}

// AddValidationRule 添加验证规则
func (cms *ConfigManagementService) AddValidationRule(field string, rule ConfigValidationRule) {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	cms.validationRules[field] = rule
	cms.storageManager.LogInfo("配置验证规则已添加", map[string]interface{}{
		"field": field,
		"rule":  rule,
	})
}

// RemoveValidationRule 移除验证规则
func (cms *ConfigManagementService) RemoveValidationRule(field string) {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	delete(cms.validationRules, field)
	cms.storageManager.LogInfo("配置验证规则已移除", map[string]interface{}{
		"field": field,
	})
}

// AddChangeCallback 添加变更回调
func (cms *ConfigManagementService) AddChangeCallback(callback ConfigChangeCallback) {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	cms.changeCallbacks = append(cms.changeCallbacks, callback)
	cms.storageManager.LogInfo("配置变更回调已添加", map[string]interface{}{
		"callback_count": len(cms.changeCallbacks),
	})
}

// RemoveChangeCallback 移除变更回调
func (cms *ConfigManagementService) RemoveChangeCallback(callback ConfigChangeCallback) {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	for i, cb := range cms.changeCallbacks {
		if &cb == &callback {
			cms.changeCallbacks = append(cms.changeCallbacks[:i], cms.changeCallbacks[i+1:]...)
			break
		}
	}
}

// GetConfigHistory 获取配置历史
func (cms *ConfigManagementService) GetConfigHistory(limit int) []ConfigSnapshot {
	cms.mutex.RLock()
	defer cms.mutex.RUnlock()

	if limit <= 0 || limit > len(cms.configHistory) {
		limit = len(cms.configHistory)
	}

	start := len(cms.configHistory) - limit
	if start < 0 {
		start = 0
	}

	return cms.configHistory[start:]
}

// RestoreConfig 恢复配置
func (cms *ConfigManagementService) RestoreConfig(snapshotID string) error {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	// 查找快照
	var snapshot *ConfigSnapshot
	for _, s := range cms.configHistory {
		if s.ID == snapshotID {
			snapshot = &s
			break
		}
	}

	if snapshot == nil {
		return fmt.Errorf("配置快照不存在: %s", snapshotID)
	}

	// 创建备份
	if err := cms.createBackup(); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	// 恢复配置
	if err := cms.saveConfigToFile(snapshot.Config); err != nil {
		return fmt.Errorf("恢复配置失败: %v", err)
	}

	cms.storageManager.LogInfo("配置已恢复", map[string]interface{}{
		"snapshot_id": snapshotID,
		"version":     snapshot.Version,
		"timestamp":   snapshot.Timestamp,
	})

	return nil
}

// GetConfigStatus 获取配置状态
func (cms *ConfigManagementService) GetConfigStatus() map[string]interface{} {
	cms.mutex.RLock()
	defer cms.mutex.RUnlock()

	return map[string]interface{}{
		"config_path":      cms.configPath,
		"backup_path":      cms.backupPath,
		"history_count":    len(cms.configHistory),
		"validation_rules": len(cms.validationRules),
		"change_callbacks": len(cms.changeCallbacks),
		"last_modified":    cms.getLastModifiedTime(),
		"current_checksum": cms.getCurrentChecksum(),
	}
}

// getLastModifiedTime 获取最后修改时间
func (cms *ConfigManagementService) getLastModifiedTime() time.Time {
	info, err := os.Stat(cms.configPath)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// getCurrentChecksum 获取当前校验和
func (cms *ConfigManagementService) getCurrentChecksum() string {
	data, err := os.ReadFile(cms.configPath)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", data)
}

// CleanupOldBackups 清理旧备份
func (cms *ConfigManagementService) CleanupOldBackups(keepDays int) error {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)

	// 读取备份目录
	entries, err := os.ReadDir(cms.backupPath)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %v", err)
	}

	removedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			backupFile := filepath.Join(cms.backupPath, entry.Name())
			if err := os.Remove(backupFile); err != nil {
				cms.storageManager.LogError("删除旧备份失败", map[string]interface{}{
					"file":  backupFile,
					"error": err.Error(),
				})
			} else {
				removedCount++
			}
		}
	}

	cms.storageManager.LogInfo("旧备份清理完成", map[string]interface{}{
		"removed_count": removedCount,
		"keep_days":     keepDays,
	})

	return nil
}

// 全局配置管理服务
var globalConfigManagementService *ConfigManagementService

// InitConfigManagement 初始化配置管理
func InitConfigManagement(storageManager *Storage.StorageManager, configPath string) {
	globalConfigManagementService = NewConfigManagementService(storageManager, configPath)

	// 添加默认验证规则
	globalConfigManagementService.AddValidationRule("server.port", ConfigValidationRule{
		Field:    "server.port",
		Required: true,
		Type:     "int",
		Min:      1,
		Max:      65535,
	})

	globalConfigManagementService.AddValidationRule("server.host", ConfigValidationRule{
		Field:    "server.host",
		Required: true,
		Type:     "string",
	})

	globalConfigManagementService.AddValidationRule("database.host", ConfigValidationRule{
		Field:    "database.host",
		Required: true,
		Type:     "string",
	})
}

// GetConfigManagementService 获取全局配置管理服务
func GetConfigManagementService() *ConfigManagementService {
	return globalConfigManagementService
}
