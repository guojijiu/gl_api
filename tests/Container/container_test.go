package Container

import (
	"cloud-platform-api/app/Container"
	"context"
	"testing"
	"time"
)

// TestContainer 测试容器基本功能
func TestContainer(t *testing.T) {
	container := Container.NewContainer()

	// 测试注册和获取服务
	container.Register("test_service", "test_value")

	service, err := container.Get("test_service")
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if service != "test_value" {
		t.Errorf("期望值: test_value, 实际值: %v", service)
	}
}

// TestContainerSingleton 测试单例服务
func TestContainerSingleton(t *testing.T) {
	container := Container.NewContainer()

	// 注册单例服务
	container.RegisterSingleton("singleton_service", func() interface{} {
		return time.Now()
	})

	// 获取服务两次
	service1, err := container.Get("singleton_service")
	if err != nil {
		t.Fatalf("获取单例服务失败: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	service2, err := container.Get("singleton_service")
	if err != nil {
		t.Fatalf("获取单例服务失败: %v", err)
	}

	// 验证是同一个实例
	if service1 != service2 {
		t.Error("单例服务应该返回同一个实例")
	}
}

// TestContainerTransient 测试瞬态服务
func TestContainerTransient(t *testing.T) {
	container := Container.NewContainer()

	// 注册瞬态服务
	container.RegisterTransient("transient_service", func() interface{} {
		return time.Now()
	})

	// 获取服务两次
	service1, err := container.Get("transient_service")
	if err != nil {
		t.Fatalf("获取瞬态服务失败: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	service2, err := container.Get("transient_service")
	if err != nil {
		t.Fatalf("获取瞬态服务失败: %v", err)
	}

	// 验证是不同的实例
	if service1 == service2 {
		t.Error("瞬态服务应该返回不同的实例")
	}
}

// TestContainerWithContext 测试带上下文的服务获取
func TestContainerWithContext(t *testing.T) {
	container := Container.NewContainer()
	container.Register("test_service", "test_value")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	service, err := container.GetWithContext(ctx, "test_service")
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if service != "test_value" {
		t.Errorf("期望值: test_value, 实际值: %v", service)
	}
}

// TestContainerWithCancelledContext 测试取消的上下文
func TestContainerWithCancelledContext(t *testing.T) {
	container := Container.NewContainer()
	container.Register("test_service", "test_value")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	_, err := container.GetWithContext(ctx, "test_service")
	if err == nil {
		t.Error("期望在取消的上下文中获取服务失败")
	}
}

// TestContainerHas 测试服务存在检查
func TestContainerHas(t *testing.T) {
	container := Container.NewContainer()

	// 测试不存在的服务
	if container.Has("non_existent_service") {
		t.Error("不存在的服务应该返回false")
	}

	// 注册服务
	container.Register("test_service", "test_value")

	// 测试存在的服务
	if !container.Has("test_service") {
		t.Error("存在的服务应该返回true")
	}
}

// TestContainerRemove 测试服务移除
func TestContainerRemove(t *testing.T) {
	container := Container.NewContainer()
	container.Register("test_service", "test_value")

	// 验证服务存在
	if !container.Has("test_service") {
		t.Error("服务应该存在")
	}

	// 移除服务
	container.Remove("test_service")

	// 验证服务不存在
	if container.Has("test_service") {
		t.Error("服务应该不存在")
	}
}

// TestContainerClear 测试清空容器
func TestContainerClear(t *testing.T) {
	container := Container.NewContainer()
	container.Register("service1", "value1")
	container.Register("service2", "value2")

	// 验证服务存在
	if !container.Has("service1") || !container.Has("service2") {
		t.Error("服务应该存在")
	}

	// 清空容器
	container.Clear()

	// 验证服务不存在
	if container.Has("service1") || container.Has("service2") {
		t.Error("服务应该不存在")
	}
}

// TestContainerList 测试服务列表
func TestContainerList(t *testing.T) {
	container := Container.NewContainer()
	container.Register("service1", "value1")
	container.Register("service2", "value2")

	services := container.List()

	if len(services) != 2 {
		t.Errorf("期望2个服务，实际: %d", len(services))
	}

	// 验证服务名称
	serviceMap := make(map[string]bool)
	for _, service := range services {
		serviceMap[service] = true
	}

	if !serviceMap["service1"] || !serviceMap["service2"] {
		t.Error("服务列表应该包含所有注册的服务")
	}
}

// TestServiceProvider 测试服务提供者
func TestServiceProvider(t *testing.T) {
	container := Container.NewContainer()

	// 创建测试服务提供者
	provider := Container.ServiceProviderFunc(func(container *Container.Container) error {
		container.Register("provider_service", "provider_value")
		return nil
	})

	// 注册服务提供者
	err := container.RegisterProvider(provider)
	if err != nil {
		t.Fatalf("注册服务提供者失败: %v", err)
	}

	// 验证服务被注册
	if !container.Has("provider_service") {
		t.Error("服务提供者应该注册服务")
	}

	// 验证服务值
	service, err := container.Get("provider_service")
	if err != nil {
		t.Fatalf("获取服务失败: %v", err)
	}

	if service != "provider_value" {
		t.Errorf("期望值: provider_value, 实际值: %v", service)
	}
}

// TestGlobalContainer 测试全局容器
func TestGlobalContainer(t *testing.T) {
	// 获取全局容器
	container := Container.GetGlobalContainer()
	if container == nil {
		t.Error("全局容器不应该为nil")
	}

	// 设置新的全局容器
	newContainer := Container.NewContainer()
	Container.SetGlobalContainer(newContainer)

	// 验证全局容器已更新
	globalContainer := Container.GetGlobalContainer()
	if globalContainer != newContainer {
		t.Error("全局容器应该被更新")
	}
}

// BenchmarkContainerGet 性能测试：获取服务
func BenchmarkContainerGet(b *testing.B) {
	container := Container.NewContainer()
	container.Register("test_service", "test_value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := container.Get("test_service")
		if err != nil {
			b.Fatalf("获取服务失败: %v", err)
		}
	}
}

// BenchmarkContainerSingleton 性能测试：单例服务
func BenchmarkContainerSingleton(b *testing.B) {
	container := Container.NewContainer()
	container.RegisterSingleton("singleton_service", func() interface{} {
		return "singleton_value"
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := container.Get("singleton_service")
		if err != nil {
			b.Fatalf("获取单例服务失败: %v", err)
		}
	}
}

// BenchmarkContainerTransient 性能测试：瞬态服务
func BenchmarkContainerTransient(b *testing.B) {
	container := Container.NewContainer()
	container.RegisterTransient("transient_service", func() interface{} {
		return "transient_value"
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := container.Get("transient_service")
		if err != nil {
			b.Fatalf("获取瞬态服务失败: %v", err)
		}
	}
}
