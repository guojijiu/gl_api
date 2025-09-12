package Utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceProfiler 性能分析器
type PerformanceProfiler struct {
	startTime time.Time
	endTime   time.Time
	metrics   map[string]interface{}
	mu        sync.RWMutex
}

// NewPerformanceProfiler 创建性能分析器
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		startTime: time.Now(),
		metrics:   make(map[string]interface{}),
	}
}

// Start 开始性能分析
func (p *PerformanceProfiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.startTime = time.Now()
}

// Stop 停止性能分析
func (p *PerformanceProfiler) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.endTime = time.Now()
}

// AddMetric 添加指标
func (p *PerformanceProfiler) AddMetric(name string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[name] = value
}

// GetDuration 获取执行时间
func (p *PerformanceProfiler) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.endTime.IsZero() {
		return time.Since(p.startTime)
	}
	return p.endTime.Sub(p.startTime)
}

// GetMetrics 获取所有指标
func (p *PerformanceProfiler) GetMetrics() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metrics := make(map[string]interface{})
	for k, v := range p.metrics {
		metrics[k] = v
	}

	// 添加执行时间
	metrics["duration"] = p.GetDuration()

	return metrics
}

// GetMemoryUsage 获取内存使用情况
func (p *PerformanceProfiler) GetMemoryUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// GetGoroutineCount 获取Goroutine数量
func (p *PerformanceProfiler) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// PerformanceOptimizer 性能优化器
type PerformanceOptimizer struct {
	profiler *PerformanceProfiler
	config   *OptimizationConfig
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	EnableMemoryOptimization      bool          `json:"enable_memory_optimization"`
	EnableGCOptimization          bool          `json:"enable_gc_optimization"`
	EnableConcurrencyOptimization bool          `json:"enable_concurrency_optimization"`
	MemoryThreshold               uint64        `json:"memory_threshold"`
	GCThreshold                   time.Duration `json:"gc_threshold"`
	MaxGoroutines                 int           `json:"max_goroutines"`
}

// NewPerformanceOptimizer 创建性能优化器
func NewPerformanceOptimizer(config *OptimizationConfig) *PerformanceOptimizer {
	if config == nil {
		config = &OptimizationConfig{
			EnableMemoryOptimization:      true,
			EnableGCOptimization:          true,
			EnableConcurrencyOptimization: true,
			MemoryThreshold:               100 * 1024 * 1024, // 100MB
			GCThreshold:                   5 * time.Second,
			MaxGoroutines:                 1000,
		}
	}

	return &PerformanceOptimizer{
		profiler: NewPerformanceProfiler(),
		config:   config,
	}
}

// OptimizeMemory 优化内存使用
func (po *PerformanceOptimizer) OptimizeMemory() error {
	if !po.config.EnableMemoryOptimization {
		return nil
	}

	// 获取当前内存使用情况
	memStats := po.profiler.GetMemoryUsage()

	// 如果内存使用超过阈值，触发GC
	if memStats.HeapAlloc > po.config.MemoryThreshold {
		runtime.GC()

		// 再次检查内存使用情况
		runtime.ReadMemStats(&memStats)
		if memStats.HeapAlloc > po.config.MemoryThreshold {
			return fmt.Errorf("内存使用仍然过高: %d bytes", memStats.HeapAlloc)
		}
	}

	return nil
}

// OptimizeGC 优化垃圾回收
func (po *PerformanceOptimizer) OptimizeGC() error {
	if !po.config.EnableGCOptimization {
		return nil
	}

	// 设置GC目标百分比
	runtime.GC()

	// 获取GC统计信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 如果GC频率过高，调整GC参数
	if m.NumGC > 0 {
		lastGC := time.Unix(0, int64(m.LastGC))
		if time.Since(lastGC) < po.config.GCThreshold {
			// 调整GC目标百分比
			runtime.GC()
		}
	}

	return nil
}

// OptimizeConcurrency 优化并发
func (po *PerformanceOptimizer) OptimizeConcurrency() error {
	if !po.config.EnableConcurrencyOptimization {
		return nil
	}

	// 检查Goroutine数量
	goroutineCount := po.profiler.GetGoroutineCount()

	if goroutineCount > po.config.MaxGoroutines {
		return fmt.Errorf("Goroutine数量过多: %d", goroutineCount)
	}

	return nil
}

// RunOptimization 运行所有优化
func (po *PerformanceOptimizer) RunOptimization() error {
	po.profiler.Start()
	defer po.profiler.Stop()

	var errors []error

	// 运行内存优化
	if err := po.OptimizeMemory(); err != nil {
		errors = append(errors, fmt.Errorf("内存优化失败: %v", err))
	}

	// 运行GC优化
	if err := po.OptimizeGC(); err != nil {
		errors = append(errors, fmt.Errorf("GC优化失败: %v", err))
	}

	// 运行并发优化
	if err := po.OptimizeConcurrency(); err != nil {
		errors = append(errors, fmt.Errorf("并发优化失败: %v", err))
	}

	// 添加性能指标
	po.profiler.AddMetric("memory_usage", po.profiler.GetMemoryUsage().HeapAlloc)
	po.profiler.AddMetric("goroutine_count", po.profiler.GetGoroutineCount())
	po.profiler.AddMetric("gc_count", po.profiler.GetMemoryUsage().NumGC)

	if len(errors) > 0 {
		return fmt.Errorf("优化过程中出现错误: %v", errors)
	}

	return nil
}

// GetOptimizationReport 获取优化报告
func (po *PerformanceOptimizer) GetOptimizationReport() map[string]interface{} {
	report := make(map[string]interface{})

	// 添加性能指标
	report["metrics"] = po.profiler.GetMetrics()

	// 添加配置信息
	report["config"] = po.config

	// 添加系统信息
	report["system"] = map[string]interface{}{
		"go_version":      runtime.Version(),
		"num_cpu":         runtime.NumCPU(),
		"goroutine_count": po.profiler.GetGoroutineCount(),
		"memory_usage":    po.profiler.GetMemoryUsage(),
	}

	return report
}

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
	config *CacheOptimizationConfig
}

// CacheOptimizationConfig 缓存优化配置
type CacheOptimizationConfig struct {
	EnableLRUOptimization  bool          `json:"enable_lru_optimization"`
	EnableTTLOptimization  bool          `json:"enable_ttl_optimization"`
	EnableSizeOptimization bool          `json:"enable_size_optimization"`
	LRUThreshold           int           `json:"lru_threshold"`
	TTLThreshold           time.Duration `json:"ttl_threshold"`
	SizeThreshold          int           `json:"size_threshold"`
}

// NewCacheOptimizer 创建缓存优化器
func NewCacheOptimizer(config *CacheOptimizationConfig) *CacheOptimizer {
	if config == nil {
		config = &CacheOptimizationConfig{
			EnableLRUOptimization:  true,
			EnableTTLOptimization:  true,
			EnableSizeOptimization: true,
			LRUThreshold:           1000,
			TTLThreshold:           1 * time.Hour,
			SizeThreshold:          10000,
		}
	}

	return &CacheOptimizer{
		config: config,
	}
}

// OptimizeLRU 优化LRU策略
func (co *CacheOptimizer) OptimizeLRU() error {
	if !co.config.EnableLRUOptimization {
		return nil
	}

	// 这里可以实现LRU优化逻辑
	// 例如：调整LRU阈值、优化LRU算法等

	return nil
}

// OptimizeTTL 优化TTL策略
func (co *CacheOptimizer) OptimizeTTL() error {
	if !co.config.EnableTTLOptimization {
		return nil
	}

	// 这里可以实现TTL优化逻辑
	// 例如：动态调整TTL、优化过期策略等

	return nil
}

// OptimizeSize 优化缓存大小
func (co *CacheOptimizer) OptimizeSize() error {
	if !co.config.EnableSizeOptimization {
		return nil
	}

	// 这里可以实现缓存大小优化逻辑
	// 例如：动态调整缓存大小、优化内存使用等

	return nil
}

// RunCacheOptimization 运行缓存优化
func (co *CacheOptimizer) RunCacheOptimization() error {
	var errors []error

	// 运行LRU优化
	if err := co.OptimizeLRU(); err != nil {
		errors = append(errors, fmt.Errorf("LRU优化失败: %v", err))
	}

	// 运行TTL优化
	if err := co.OptimizeTTL(); err != nil {
		errors = append(errors, fmt.Errorf("TTL优化失败: %v", err))
	}

	// 运行大小优化
	if err := co.OptimizeSize(); err != nil {
		errors = append(errors, fmt.Errorf("大小优化失败: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("缓存优化过程中出现错误: %v", errors)
	}

	return nil
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics   map[string]interface{}
	mu        sync.RWMutex
	startTime time.Time
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]interface{}),
		startTime: time.Now(),
	}
}

// RecordMetric 记录指标
func (pm *PerformanceMonitor) RecordMetric(name string, value interface{}) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.metrics[name] = value
}

// GetMetrics 获取指标
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make(map[string]interface{})
	for k, v := range pm.metrics {
		metrics[k] = v
	}

	// 添加运行时间
	metrics["uptime"] = time.Since(pm.startTime)

	return metrics
}

// GetSystemMetrics 获取系统指标
func (pm *PerformanceMonitor) GetSystemMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"memory_alloc":       m.Alloc,
		"memory_total_alloc": m.TotalAlloc,
		"memory_sys":         m.Sys,
		"memory_heap_alloc":  m.HeapAlloc,
		"memory_heap_sys":    m.HeapSys,
		"gc_count":           m.NumGC,
		"goroutine_count":    runtime.NumGoroutine(),
		"cpu_count":          runtime.NumCPU(),
	}
}

// 全局性能监控器
var globalPerformanceMonitor *PerformanceMonitor
var once sync.Once

// GetGlobalPerformanceMonitor 获取全局性能监控器
func GetGlobalPerformanceMonitor() *PerformanceMonitor {
	once.Do(func() {
		globalPerformanceMonitor = NewPerformanceMonitor()
	})
	return globalPerformanceMonitor
}

// RecordGlobalMetric 记录全局指标
func RecordGlobalMetric(name string, value interface{}) {
	monitor := GetGlobalPerformanceMonitor()
	monitor.RecordMetric(name, value)
}

// GetGlobalMetrics 获取全局指标
func GetGlobalMetrics() map[string]interface{} {
	monitor := GetGlobalPerformanceMonitor()
	return monitor.GetMetrics()
}
