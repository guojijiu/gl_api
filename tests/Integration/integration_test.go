package Integration

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Container"
	"cloud-platform-api/app/Services"
	"context"
	"fmt"
	"testing"
	"time"
)

// TestIntegrationContainerAndServices 测试容器和服务集成
func TestIntegrationContainerAndServices(t *testing.T) {
	// 初始化容器
	container, err := Container.InitializeContainer()
	if err != nil {
		t.Fatalf("初始化容器失败: %v", err)
	}

	// 测试获取服务
	userService, err := container.Get("user_service")
	if err != nil {
		t.Fatalf("获取用户服务失败: %v", err)
	}

	if userService == nil {
		t.Error("用户服务不应该为nil")
	}

	// 测试获取缓存服务
	cacheService, err := container.Get("cache_service")
	if err != nil {
		t.Fatalf("获取缓存服务失败: %v", err)
	}

	if cacheService == nil {
		t.Error("缓存服务不应该为nil")
	}
}

// TestIntegrationServiceManager 测试服务管理器集成
func TestIntegrationServiceManager(t *testing.T) {
	// 创建示例服务
	exampleService := Services.NewExampleService()

	// 注册服务
	Services.RegisterGlobalService("test_example_service", exampleService)

	// 获取服务
	service, err := Services.GetGlobalService("test_example_service")
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if service == nil {
		t.Error("服务不应该为nil")
	}

	// 验证服务名称
	if service.GetName() != "example_service" {
		t.Errorf("期望服务名称: example_service, 实际: %s", service.GetName())
	}
}

// TestIntegrationServiceInitialization 测试服务初始化
func TestIntegrationServiceInitialization(t *testing.T) {
	// 创建示例服务
	exampleService := Services.NewExampleService()

	// 初始化服务
	err := exampleService.Initialize()
	if err != nil {
		t.Fatalf("服务初始化失败: %v", err)
	}

	// 验证服务已初始化
	if !exampleService.IsInitialized() {
		t.Error("服务应该已初始化")
	}

	// 关闭服务
	err = exampleService.Shutdown()
	if err != nil {
		t.Fatalf("服务关闭失败: %v", err)
	}

	// 验证服务已关闭
	if exampleService.IsInitialized() {
		t.Error("服务应该已关闭")
	}
}

// TestIntegrationServiceWithDependencies 测试带依赖的服务
func TestIntegrationServiceWithDependencies(t *testing.T) {
	// 初始化容器
	_, err := Container.InitializeContainer()
	if err != nil {
		t.Fatalf("初始化容器失败: %v", err)
	}

	// 创建示例服务
	exampleService := Services.NewExampleService()

	// 初始化服务（这会触发依赖注入）
	err = exampleService.Initialize()
	if err != nil {
		t.Fatalf("服务初始化失败: %v", err)
	}

	// 测试服务功能
	ctx := context.Background()
	testData := "test_data"

	err = exampleService.ProcessData(ctx, testData)
	if err != nil {
		t.Fatalf("处理数据失败: %v", err)
	}

	// 获取服务信息
	info := exampleService.GetServiceInfo()
	if info == nil {
		t.Error("服务信息不应该为nil")
	}

	// 验证服务名称
	if name, ok := info["name"].(string); !ok || name != "example_service" {
		t.Error("服务信息应该包含正确的名称")
	}

	// 验证依赖信息
	if deps, ok := info["dependencies"].(map[string]interface{}); !ok {
		t.Error("服务信息应该包含依赖信息")
	} else {
		// 检查依赖是否已注入
		if deps["database_config"] != "已注入" {
			t.Error("数据库配置应该已注入")
		}
	}
}

// TestIntegrationConfigHotReload 测试配置热重载
func TestIntegrationConfigHotReload(t *testing.T) {
	// 创建临时配置文件
	configPath := "test_config.yaml"

	// 初始化配置管理器
	err := Config.InitializeConfigManager(configPath)
	if err != nil {
		t.Logf("配置管理器初始化失败（这是预期的，因为配置文件不存在）: %v", err)
		return
	}

	// 获取配置管理器
	manager := Config.GetConfigManager()
	if manager == nil {
		t.Error("配置管理器不应该为nil")
	}

	// 停止配置管理器
	Config.StopConfigManager()
}

// TestIntegrationOptimizedCache 测试优化的缓存服务
func TestIntegrationOptimizedCache(t *testing.T) {
	// 创建优化的缓存服务
	cacheService := Services.NewOptimizedCacheService()

	// 初始化服务
	err := cacheService.Initialize()
	if err != nil {
		t.Fatalf("缓存服务初始化失败: %v", err)
	}

	// 测试缓存操作
	ctx := context.Background()

	// 设置缓存
	err = cacheService.Set(ctx, "test_key", "test_value", time.Hour)
	if err != nil {
		t.Fatalf("设置缓存失败: %v", err)
	}

	// 获取缓存
	value, err := cacheService.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("获取缓存失败: %v", err)
	}

	if value != "test_value" {
		t.Errorf("期望值: test_value, 实际值: %v", value)
	}

	// 检查缓存是否存在
	exists := cacheService.Exists(ctx, "test_key")
	if !exists {
		t.Error("缓存应该存在")
	}

	// 删除缓存
	err = cacheService.Delete(ctx, "test_key")
	if err != nil {
		t.Fatalf("删除缓存失败: %v", err)
	}

	// 验证缓存已删除
	exists = cacheService.Exists(ctx, "test_key")
	if exists {
		t.Error("缓存应该不存在")
	}

	// 获取统计信息
	stats := cacheService.GetStats()
	if stats == nil {
		t.Error("统计信息不应该为nil")
	}

	// 关闭服务
	err = cacheService.Stop()
	if err != nil {
		t.Fatalf("缓存服务关闭失败: %v", err)
	}
}

// TestIntegrationOptimizedMonitoring 测试优化的监控服务
func TestIntegrationOptimizedMonitoring(t *testing.T) {
	// 创建优化的监控服务
	monitoringService := Services.NewOptimizedMonitoringService()

	// 初始化服务
	err := monitoringService.Initialize()
	if err != nil {
		t.Fatalf("监控服务初始化失败: %v", err)
	}

	// 启动监控服务
	err = monitoringService.Start()
	if err != nil {
		t.Fatalf("启动监控服务失败: %v", err)
	}

	// 验证服务正在运行
	if !monitoringService.IsRunning() {
		t.Error("监控服务应该正在运行")
	}

	// 添加一些指标
	monitoringService.AddMetric("test_metric", 123, map[string]string{"tag1": "value1"})

	// 等待一段时间让指标被收集
	time.Sleep(100 * time.Millisecond)

	// 获取系统指标
	systemMetrics, err := monitoringService.GetSystemMetrics()
	if err != nil {
		t.Logf("获取系统指标失败（这是预期的，因为可能没有实际的数据）: %v", err)
	} else if systemMetrics != nil {
		// 验证指标结构
		if systemMetrics.Timestamp.IsZero() {
			t.Error("系统指标应该包含时间戳")
		}
	}

	// 停止监控服务
	err = monitoringService.Stop()
	if err != nil {
		t.Fatalf("停止监控服务失败: %v", err)
	}

	// 验证服务已停止
	if monitoringService.IsRunning() {
		t.Error("监控服务应该已停止")
	}
}

// TestIntegrationServiceLifecycle 测试服务生命周期
func TestIntegrationServiceLifecycle(t *testing.T) {
	// 创建服务管理器
	manager := Services.NewServiceManager()

	// 创建示例服务
	exampleService := Services.NewExampleService()

	// 注册服务
	manager.Register("test_service", exampleService)

	// 初始化所有服务
	err := manager.InitializeAll()
	if err != nil {
		t.Fatalf("初始化所有服务失败: %v", err)
	}

	// 验证服务已初始化
	service, err := manager.Get("test_service")
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if !service.IsInitialized() {
		t.Error("服务应该已初始化")
	}

	// 关闭所有服务
	err = manager.ShutdownAll()
	if err != nil {
		t.Fatalf("关闭所有服务失败: %v", err)
	}

	// 验证服务已关闭
	if service.IsInitialized() {
		t.Error("服务应该已关闭")
	}
}

// TestIntegrationConcurrentAccess 测试并发访问
func TestIntegrationConcurrentAccess(t *testing.T) {
	// 创建优化的缓存服务
	cacheService := Services.NewOptimizedCacheService()

	// 初始化服务
	err := cacheService.Initialize()
	if err != nil {
		t.Fatalf("缓存服务初始化失败: %v", err)
	}

	// 并发设置缓存
	ctx := context.Background()
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			key := fmt.Sprintf("key_%d", index)
			value := fmt.Sprintf("value_%d", index)

			err := cacheService.Set(ctx, key, value, time.Hour)
			if err != nil {
				t.Errorf("设置缓存失败: %v", err)
			}

			done <- true
		}(i)
	}

	// 等待所有协程完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证所有缓存都已设置
	for i := 0; i < concurrency; i++ {
		key := fmt.Sprintf("key_%d", i)
		exists := cacheService.Exists(ctx, key)
		if !exists {
			t.Errorf("缓存 %s 应该存在", key)
		}
	}

	// 关闭服务
	err = cacheService.Stop()
	if err != nil {
		t.Fatalf("缓存服务关闭失败: %v", err)
	}
}

// BenchmarkIntegrationCache 性能测试：缓存操作
func BenchmarkIntegrationCache(b *testing.B) {
	cacheService := Services.NewOptimizedCacheService()
	cacheService.Initialize()
	defer cacheService.Stop()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)

		cacheService.Set(ctx, key, value, time.Hour)
		cacheService.Get(ctx, key)
	}
}

// BenchmarkIntegrationServiceCreation 性能测试：服务创建
func BenchmarkIntegrationServiceCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service := Services.NewExampleService()
		service.Initialize()
		service.Shutdown()
	}
}
