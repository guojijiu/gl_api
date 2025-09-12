package Container

import (
	"cloud-platform-api/app/Services"
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

// 确保Container实现了Services.ContainerInterface接口
var _ Services.ContainerInterface = (*Container)(nil)

// NewContainer 创建新的容器实例
func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

// Register 注册服务到容器
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// RegisterSingleton 注册单例服务
func (c *Container) RegisterSingleton(name string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 使用懒加载，第一次获取时才创建实例
	c.services[name] = &singletonWrapper{
		factory:  factory,
		instance: nil,
		mu:       sync.Mutex{},
	}
}

// RegisterTransient 注册瞬态服务（每次获取都创建新实例）
func (c *Container) RegisterTransient(name string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = &transientWrapper{
		factory: factory,
	}
}

// Get 获取服务
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	service, exists := c.services[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	// 处理不同类型的服务包装器
	switch wrapper := service.(type) {
	case *singletonWrapper:
		return wrapper.getInstance(), nil
	case *transientWrapper:
		return wrapper.createInstance(), nil
	default:
		return service, nil
	}
}

// GetWithContext 使用上下文获取服务
func (c *Container) GetWithContext(ctx context.Context, name string) (interface{}, error) {
	// 检查上下文是否被取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c.Get(name)
	}
}

// Resolve 解析服务并自动注入依赖
func (c *Container) Resolve(target interface{}) error {
	return c.resolveValue(reflect.ValueOf(target))
}

// resolveValue 解析值并注入依赖
func (c *Container) resolveValue(value reflect.Value) error {
	if !value.IsValid() || !value.CanSet() {
		return fmt.Errorf("invalid or unsettable value")
	}

	// 获取类型信息
	targetType := value.Type()

	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(targetType.Elem()))
		}
		return c.resolveValue(value.Elem())
	}

	// 处理结构体类型
	if targetType.Kind() == reflect.Struct {
		return c.resolveStruct(value)
	}

	return nil
}

// resolveStruct 解析结构体并注入依赖
func (c *Container) resolveStruct(value reflect.Value) error {
	targetType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := targetType.Field(i)

		// 检查是否有依赖注入标签
		if tag, ok := fieldType.Tag.Lookup("inject"); ok {
			if tag == "" {
				continue
			}

			// 获取依赖服务
			service, err := c.Get(tag)
			if err != nil {
				return fmt.Errorf("failed to resolve dependency '%s' for field '%s': %v", tag, fieldType.Name, err)
			}

			// 设置字段值
			serviceValue := reflect.ValueOf(service)
			if serviceValue.Type().AssignableTo(field.Type()) {
				field.Set(serviceValue)
			} else {
				return fmt.Errorf("type mismatch for field '%s': expected %s, got %s",
					fieldType.Name, field.Type(), serviceValue.Type())
			}
		}
	}

	return nil
}

// Has 检查服务是否已注册
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.services[name]
	return exists
}

// Remove 移除服务
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.services, name)
}

// Clear 清空所有服务
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[string]interface{})
}

// List 列出所有已注册的服务
func (c *Container) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]string, 0, len(c.services))
	for name := range c.services {
		services = append(services, name)
	}
	return services
}

// singletonWrapper 单例服务包装器
type singletonWrapper struct {
	factory  func() interface{}
	instance interface{}
	mu       sync.Mutex
}

func (s *singletonWrapper) getInstance() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.instance == nil {
		s.instance = s.factory()
	}
	return s.instance
}

// transientWrapper 瞬态服务包装器
type transientWrapper struct {
	factory func() interface{}
}

func (t *transientWrapper) createInstance() interface{} {
	return t.factory()
}

// ServiceProvider 服务提供者接口
type ServiceProvider interface {
	Register(container *Container) error
}

// ServiceProviderFunc 服务提供者函数类型
type ServiceProviderFunc func(container *Container) error

func (f ServiceProviderFunc) Register(container *Container) error {
	return f(container)
}

// RegisterProvider 注册服务提供者
func (c *Container) RegisterProvider(provider ServiceProvider) error {
	return provider.Register(c)
}

// 全局容器实例
var globalContainer *Container
var once sync.Once

// GetGlobalContainer 获取全局容器实例
func GetGlobalContainer() *Container {
	once.Do(func() {
		globalContainer = NewContainer()
	})
	return globalContainer
}

// SetGlobalContainer 设置全局容器实例
func SetGlobalContainer(container *Container) {
	globalContainer = container
}
