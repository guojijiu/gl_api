package Config

import (
	"cloud-platform-api/app/Config"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotReloadManager(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 创建测试配置文件
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		},
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
	}

	configJSON, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, configJSON, 0644)
	require.NoError(t, err)

	t.Run("创建热重载管理器", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)
		assert.NotNil(t, manager)
		assert.Equal(t, configFile, manager.GetConfigPath())
		assert.False(t, manager.IsWatching())
	})

	t.Run("添加重载回调", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		callbackCount := 0
		callback := func(config *Config.Config) {
			callbackCount++
		}

		manager.AddReloadCallback(callback)
		assert.Equal(t, 1, manager.GetReloadCallbacksCount())

		// 添加重复回调
		manager.AddReloadCallback(callback)
		assert.Equal(t, 2, manager.GetReloadCallbacksCount())
	})

	t.Run("移除重载回调", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		callback := func(config *Config.Config) {}
		manager.AddReloadCallback(callback)
		assert.Equal(t, 1, manager.GetReloadCallbacksCount())

		manager.RemoveReloadCallback(callback)
		assert.Equal(t, 0, manager.GetReloadCallbacksCount())
	})

	t.Run("设置配置文件路径", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		newPath := filepath.Join(tempDir, "new_config.yaml")
		err := manager.SetConfigPath(newPath)
		assert.NoError(t, err)
		assert.Equal(t, newPath, manager.GetConfigPath())
	})
}

func TestHotReloadManagerIntegration(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 创建初始配置文件
	initialConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		},
	}

	initialJSON, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, initialJSON, 0644)
	require.NoError(t, err)

	t.Run("启动和停止监控", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		// 启动监控
		err := manager.StartWatching()
		assert.NoError(t, err)
		assert.True(t, manager.IsWatching())

		// 停止监控
		err = manager.StopWatching()
		assert.NoError(t, err)
		assert.False(t, manager.IsWatching())
	})

	t.Run("重复启动监控", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		err := manager.StartWatching()
		assert.NoError(t, err)

		// 尝试重复启动
		err = manager.StartWatching()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已经在运行")

		// 清理
		manager.StopWatching()
	})

	t.Run("停止未启动的监控", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		err := manager.StopWatching()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "未在运行")
	})
}

func TestHotReloadManagerFileWatching(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 创建初始配置文件
	initialConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
			"host": "localhost",
		},
	}

	initialJSON, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, initialJSON, 0644)
	require.NoError(t, err)

	t.Run("文件变化监控", func(t *testing.T) {
		manager := Config.NewHotReloadManager(configFile)

		callbackCount := 0
		manager.AddReloadCallback(func(config *Config.Config) {
			callbackCount++
		})

		// 启动监控
		err := manager.StartWatching()
		require.NoError(t, err)
		defer manager.StopWatching()

		// 等待监控启动
		time.Sleep(100 * time.Millisecond)

		// 修改配置文件
		updatedConfig := map[string]interface{}{
			"server": map[string]interface{}{
				"port": 9090,
				"host": "localhost",
			},
		}

		updatedJSON, err := json.MarshalIndent(updatedConfig, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(configFile, updatedJSON, 0644)
		require.NoError(t, err)

		// 等待回调执行
		time.Sleep(200 * time.Millisecond)

		// 检查回调是否被调用
		assert.Greater(t, callbackCount, 0)
	})
}

func TestGlobalHotReloadFunctions(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 创建测试配置文件
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
		},
	}

	configJSON, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, configJSON, 0644)
	require.NoError(t, err)

	t.Run("全局热重载初始化", func(t *testing.T) {
		// 确保全局管理器为空
		Config.StopHotReload()

		// 初始化全局热重载
		err := Config.InitHotReload(configFile)
		assert.NoError(t, err)

		// 检查全局管理器
		manager := Config.GetHotReloadManager()
		assert.NotNil(t, manager)
		assert.True(t, manager.IsWatching())

		// 清理
		Config.StopHotReload()
	})

	t.Run("重复初始化全局热重载", func(t *testing.T) {
		// 确保全局管理器为空
		Config.StopHotReload()

		// 第一次初始化
		err := Config.InitHotReload(configFile)
		assert.NoError(t, err)

		// 第二次初始化应该失败
		err = Config.InitHotReload(configFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已经初始化")

		// 清理
		Config.StopHotReload()
	})

	t.Run("全局回调函数", func(t *testing.T) {
		// 确保全局管理器为空
		Config.StopHotReload()

		// 初始化全局热重载
		err := Config.InitHotReload(configFile)
		require.NoError(t, err)
		defer Config.StopHotReload()

		callbackCount := 0
		callback := func(config *Config.Config) {
			callbackCount++
		}

		// 添加全局回调
		Config.AddGlobalReloadCallback(callback)

		// 检查回调数量
		manager := Config.GetHotReloadManager()
		assert.Equal(t, 1, manager.GetReloadCallbacksCount())

		// 移除全局回调
		Config.RemoveGlobalReloadCallback(callback)
		assert.Equal(t, 0, manager.GetReloadCallbacksCount())
	})
}

func BenchmarkHotReloadManager(b *testing.B) {
	// 创建临时目录
	tempDir := b.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 创建测试配置文件
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
		},
	}

	configJSON, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(b, err)

	err = os.WriteFile(configFile, configJSON, 0644)
	require.NoError(b, err)

	manager := Config.NewHotReloadManager(configFile)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 测试添加和移除回调
			callback := func(config *Config.Config) {}
			manager.AddReloadCallback(callback)
			manager.RemoveReloadCallback(callback)
		}
	})
}
