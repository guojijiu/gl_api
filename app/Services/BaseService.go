package Services

import (
	"context"
	"fmt"
	"sync"
)

// ContainerInterface 容器接口，用于解耦依赖
type ContainerInterface interface {
	Get(name string) (interface{}, error)
	GetWithContext(ctx context.Context, name string) (interface{}, error)
	Resolve(target interface{}) error
}

// BaseService 服务基类
type BaseService struct {
	container ContainerInterface
	ctx       context.Context
	DB        interface{} // 添加 DB 字段，用于数据库连接
}

// NewBaseService 创建基础服务
func NewBaseService() *BaseService {
	return &BaseService{
		container: nil, // 将在运行时设置
		ctx:       context.Background(),
		DB:        nil,
	}
}

// NewBaseServiceWithContainer 使用容器创建基础服务
func NewBaseServiceWithContainer(container ContainerInterface) *BaseService {
	return &BaseService{
		container: container,
		ctx:       context.Background(),
		DB:        nil,
	}
}

// NewBaseServiceWithContext 使用上下文创建基础服务
func NewBaseServiceWithContext(ctx context.Context) *BaseService {
	return &BaseService{
		container: nil, // 将在运行时设置
		ctx:       ctx,
		DB:        nil,
	}
}

// SetContainer 设置容器
func (s *BaseService) SetContainer(container ContainerInterface) {
	s.container = container
}

// GetService 获取服务
func (s *BaseService) GetService(name string) (interface{}, error) {
	if s.container == nil {
		return nil, fmt.Errorf("container not initialized")
	}
	return s.container.GetWithContext(s.ctx, name)
}

// ResolveService 解析服务并注入依赖
func (s *BaseService) ResolveService(target interface{}) error {
	if s.container == nil {
		return fmt.Errorf("container not initialized")
	}
	return s.container.Resolve(target)
}

// GetContext 获取上下文
func (s *BaseService) GetContext() context.Context {
	return s.ctx
}

// WithContext 使用新的上下文
func (s *BaseService) WithContext(ctx context.Context) *BaseService {
	return &BaseService{
		container: s.container,
		ctx:       ctx,
	}
}

// ServiceInterface 服务接口
type ServiceInterface interface {
	Initialize() error
	Shutdown() error
	GetName() string
	IsInitialized() bool
}

// ServiceBase 服务基础结构
type ServiceBase struct {
	*BaseService
	name        string
	initialized bool
}

// NewServiceBase 创建服务基础实例
func NewServiceBase(name string) *ServiceBase {
	return &ServiceBase{
		BaseService: NewBaseService(),
		name:        name,
		initialized: false,
	}
}

// NewServiceBaseWithContainer 使用容器创建服务基础实例
func NewServiceBaseWithContainer(name string, container ContainerInterface) *ServiceBase {
	return &ServiceBase{
		BaseService: NewBaseServiceWithContainer(container),
		name:        name,
		initialized: false,
	}
}

// Initialize 初始化服务
func (s *ServiceBase) Initialize() error {
	if s.initialized {
		return nil
	}

	// 自动注入依赖
	if s.container != nil {
		if err := s.ResolveService(s); err != nil {
			return fmt.Errorf("failed to resolve dependencies for service %s: %v", s.name, err)
		}
	}

	s.initialized = true
	return nil
}

// Shutdown 关闭服务
func (s *ServiceBase) Shutdown() error {
	s.initialized = false
	return nil
}

// GetName 获取服务名称
func (s *ServiceBase) GetName() string {
	return s.name
}

// IsInitialized 检查服务是否已初始化
func (s *ServiceBase) IsInitialized() bool {
	return s.initialized
}

// ServiceManager 服务管理器
type ServiceManager struct {
	services map[string]ServiceInterface
	mu       sync.RWMutex
}

// NewServiceManager 创建服务管理器
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]ServiceInterface),
	}
}

// Register 注册服务
func (sm *ServiceManager) Register(name string, service ServiceInterface) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[name] = service
}

// Get 获取服务
func (sm *ServiceManager) Get(name string) (ServiceInterface, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	service, exists := sm.services[name]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}
	return service, nil
}

// InitializeAll 初始化所有服务
func (sm *ServiceManager) InitializeAll() error {
	sm.mu.RLock()
	services := make([]ServiceInterface, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mu.RUnlock()

	for _, service := range services {
		if err := service.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize service %s: %v", service.GetName(), err)
		}
	}

	return nil
}

// ShutdownAll 关闭所有服务
func (sm *ServiceManager) ShutdownAll() error {
	sm.mu.RLock()
	services := make([]ServiceInterface, 0, len(sm.services))
	for _, service := range sm.services {
		services = append(services, service)
	}
	sm.mu.RUnlock()

	var lastErr error
	for _, service := range services {
		if err := service.Shutdown(); err != nil {
			lastErr = fmt.Errorf("failed to shutdown service %s: %v", service.GetName(), err)
		}
	}

	return lastErr
}

// List 列出所有服务
func (sm *ServiceManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.services))
	for name := range sm.services {
		names = append(names, name)
	}
	return names
}

// 全局服务管理器
var globalServiceManager *ServiceManager
var once sync.Once

// GetGlobalServiceManager 获取全局服务管理器
func GetGlobalServiceManager() *ServiceManager {
	once.Do(func() {
		globalServiceManager = NewServiceManager()
	})
	return globalServiceManager
}

// RegisterGlobalService 注册全局服务
func RegisterGlobalService(name string, service ServiceInterface) {
	manager := GetGlobalServiceManager()
	manager.Register(name, service)
}

// GetGlobalService 获取全局服务
func GetGlobalService(name string) (ServiceInterface, error) {
	manager := GetGlobalServiceManager()
	return manager.Get(name)
}

// InitializeGlobalServices 初始化所有全局服务
func InitializeGlobalServices() error {
	manager := GetGlobalServiceManager()
	return manager.InitializeAll()
}

// ShutdownGlobalServices 关闭所有全局服务
func ShutdownGlobalServices() error {
	manager := GetGlobalServiceManager()
	return manager.ShutdownAll()
}
