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
//
// 功能说明：
// 1. 提供所有服务的通用基础功能
// 2. 管理依赖注入容器（用于服务解耦）
// 3. 管理上下文（用于请求追踪和超时控制）
// 4. 管理数据库连接（用于数据访问）
//
// 设计模式：
// - 组合模式：所有服务都嵌入BaseService
// - 依赖注入：通过容器管理服务依赖
// - 上下文传递：支持请求级别的上下文管理
//
// 字段说明：
// - container: 依赖注入容器（可选，用于服务解耦）
// - ctx: 上下文（用于请求追踪、超时控制、取消操作）
// - DB: 数据库连接（可以是*gorm.DB或其他数据库接口）
//
// 使用场景：
// - 所有业务服务都应该嵌入BaseService
// - 需要依赖注入的服务
// - 需要上下文管理的服务
// - 需要数据库访问的服务
type BaseService struct {
	container ContainerInterface // 依赖注入容器
	ctx       context.Context    // 上下文
	DB        interface{}         // 数据库连接（可以是*gorm.DB或其他数据库接口）
}

// NewBaseService 创建基础服务
//
// 功能说明：
// 1. 创建新的BaseService实例
// 2. 初始化默认值（容器为nil，上下文为Background）
// 3. 返回可用的服务基类实例
//
// 初始化值：
// - container: nil（将在运行时通过SetContainer设置）
// - ctx: context.Background()（默认上下文）
// - DB: nil（将在运行时设置）
//
// 使用场景：
// - 创建新的服务实例
// - 不需要容器的简单服务
// - 测试环境中的服务创建
//
// 注意事项：
// - 容器和DB可以在运行时通过方法设置
// - 上下文可以通过WithContext方法创建新实例
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
//
// 功能说明：
// 1. 管理所有服务的注册和获取
// 2. 提供服务的统一生命周期管理
// 3. 支持服务的初始化和关闭
// 4. 线程安全的服务访问
//
// 设计模式：
// - 注册表模式：通过名称管理服务
// - 单例模式：全局服务管理器
// - 工厂模式：统一创建和管理服务
//
// 并发安全：
// - 使用sync.RWMutex保护服务映射
// - 支持并发读取和互斥写入
// - 线程安全的服务注册和获取
//
// 使用场景：
// - 应用启动时注册所有服务
// - 运行时通过名称获取服务
// - 应用关闭时统一关闭所有服务
type ServiceManager struct {
	services map[string]ServiceInterface // 服务映射表（名称->服务实例）
	mu       sync.RWMutex                // 读写锁（保护并发访问）
}

// NewServiceManager 创建服务管理器
//
// 功能说明：
// 1. 创建新的ServiceManager实例
// 2. 初始化服务映射表
// 3. 返回可用的服务管理器
//
// 使用场景：
// - 应用启动时创建服务管理器
// - 测试环境中创建独立的服务管理器
//
// 注意事项：
// - 每个ServiceManager实例独立管理自己的服务
// - 建议使用全局服务管理器（GetGlobalServiceManager）
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]ServiceInterface),
	}
}

// Register 注册服务
//
// 功能说明：
// 1. 将服务注册到管理器中
// 2. 使用服务名称作为key
// 3. 支持服务覆盖（同名服务会覆盖）
//
// 参数说明：
// - name: 服务名称（唯一标识）
// - service: 服务实例（必须实现ServiceInterface）
//
// 并发安全：
// - 使用写锁保护注册操作
// - 支持并发注册（会串行化）
//
// 注意事项：
// - 同名服务会覆盖之前的服务
// - 服务名称应该唯一且有意义
// - 注册后服务不会自动初始化
func (sm *ServiceManager) Register(name string, service ServiceInterface) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[name] = service
}

// Get 获取服务
//
// 功能说明：
// 1. 根据服务名称获取服务实例
// 2. 如果服务不存在返回错误
// 3. 返回服务接口（便于使用）
//
// 参数说明：
// - name: 服务名称（注册时使用的名称）
//
// 返回信息：
// - ServiceInterface: 服务实例（如果存在）
// - error: 错误信息（如果服务不存在）
//
// 并发安全：
// - 使用读锁保护读取操作
// - 支持并发读取
//
// 注意事项：
// - 服务必须已注册才能获取
// - 返回的服务可能未初始化
// - 建议在获取后调用Initialize确保服务可用
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
