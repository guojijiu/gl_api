package Services

import (
	"cloud-platform-api/app/Models"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"gorm.io/gorm"
)

// SystemMetricsCollector 系统指标收集器
type SystemMetricsCollector struct {
	service    *OptimizedMonitoringService
	mu         sync.RWMutex
	cache      map[string]interface{}
	lastUpdate time.Time
	cacheTTL   time.Duration
}

// NewSystemMetricsCollector 创建系统指标收集器
func NewSystemMetricsCollector(service *OptimizedMonitoringService) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		service:  service,
		cache:    make(map[string]interface{}),
		cacheTTL: 30 * time.Second, // 缓存30秒
	}
}

// Name 返回收集器名称
func (c *SystemMetricsCollector) Name() string {
	return "system_metrics"
}

// Type 返回收集器类型
func (c *SystemMetricsCollector) Type() string {
	return "system_metrics"
}

// Collect 收集系统指标
func (c *SystemMetricsCollector) Collect() (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存
	if time.Since(c.lastUpdate) < c.cacheTTL && len(c.cache) > 0 {
		return c.cache, nil
	}

	// 收集各种系统指标
	metrics := make(map[string]interface{})

	// CPU指标
	if cpuMetrics, err := c.collectCPUMetrics(); err == nil {
		metrics["cpu"] = cpuMetrics
	}

	// 内存指标
	if memMetrics, err := c.collectMemoryMetrics(); err == nil {
		metrics["memory"] = memMetrics
	}

	// 磁盘指标
	if diskMetrics, err := c.collectDiskMetrics(); err == nil {
		metrics["disk"] = diskMetrics
	}

	// 网络指标
	if netMetrics, err := c.collectNetworkMetrics(); err == nil {
		metrics["network"] = netMetrics
	}

	// Go运行时指标
	metrics["goruntime"] = c.collectGoRuntimeMetrics()

	// 进程指标
	if procMetrics, err := c.collectProcessMetrics(); err == nil {
		metrics["process"] = procMetrics
	}

	// 更新缓存
	c.cache = metrics
	c.lastUpdate = time.Now()

	return metrics, nil
}

// collectCPUMetrics 收集CPU指标
func (c *SystemMetricsCollector) collectCPUMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// CPU使用率
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		metrics["usage_percent"] = cpuPercent[0]
	}

	// CPU核心数
	if cpuCount, err := cpu.Counts(true); err == nil {
		metrics["cores"] = cpuCount
	}

	// CPU频率
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		metrics["frequency"] = cpuInfo[0].Mhz
		metrics["model"] = cpuInfo[0].ModelName
	}

	// CPU负载（Linux系统）
	// 注意：LoadAvg在Windows上不可用
	// if loadAvg, err := cpu.LoadAvg(); err == nil {
	// 	metrics["load_avg_1min"] = loadAvg.Load1
	// 	metrics["load_avg_5min"] = loadAvg.Load5
	// 	metrics["load_avg_15min"] = loadAvg.Load15
	// }

	return metrics, nil
}

// collectMemoryMetrics 收集内存指标
func (c *SystemMetricsCollector) collectMemoryMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 虚拟内存
	if vmStat, err := mem.VirtualMemory(); err == nil {
		metrics["total"] = vmStat.Total
		metrics["available"] = vmStat.Available
		metrics["used"] = vmStat.Used
		metrics["free"] = vmStat.Free
		metrics["usage_percent"] = vmStat.UsedPercent
		metrics["cached"] = vmStat.Cached
		metrics["buffers"] = vmStat.Buffers
	}

	// 交换内存
	if swapStat, err := mem.SwapMemory(); err == nil {
		metrics["swap_total"] = swapStat.Total
		metrics["swap_used"] = swapStat.Used
		metrics["swap_free"] = swapStat.Free
		metrics["swap_usage_percent"] = swapStat.UsedPercent
	}

	return metrics, nil
}

// collectDiskMetrics 收集磁盘指标
func (c *SystemMetricsCollector) collectDiskMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 磁盘使用情况
	if diskUsage, err := disk.Usage("/"); err == nil {
		metrics["total"] = diskUsage.Total
		metrics["used"] = diskUsage.Used
		metrics["free"] = diskUsage.Free
		metrics["usage_percent"] = diskUsage.UsedPercent
		metrics["inodes_total"] = diskUsage.InodesTotal
		metrics["inodes_used"] = diskUsage.InodesUsed
		metrics["inodes_free"] = diskUsage.InodesFree
		metrics["inodes_usage_percent"] = diskUsage.InodesUsedPercent
	}

	// 磁盘IO
	if diskIO, err := disk.IOCounters(); err == nil {
		var totalReadBytes, totalWriteBytes uint64
		var totalReadCount, totalWriteCount uint64

		for _, io := range diskIO {
			totalReadBytes += io.ReadBytes
			totalWriteBytes += io.WriteBytes
			totalReadCount += io.ReadCount
			totalWriteCount += io.WriteCount
		}

		metrics["read_bytes"] = totalReadBytes
		metrics["write_bytes"] = totalWriteBytes
		metrics["read_count"] = totalReadCount
		metrics["write_count"] = totalWriteCount
	}

	return metrics, nil
}

// collectNetworkMetrics 收集网络指标
func (c *SystemMetricsCollector) collectNetworkMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 网络IO统计
	if netIO, err := net.IOCounters(true); err == nil {
		var totalBytesSent, totalBytesRecv uint64
		var totalPacketsSent, totalPacketsRecv uint64
		var totalErrin, totalErrout uint64

		for _, io := range netIO {
			totalBytesSent += io.BytesSent
			totalBytesRecv += io.BytesRecv
			totalPacketsSent += io.PacketsSent
			totalPacketsRecv += io.PacketsRecv
			totalErrin += io.Errin
			totalErrout += io.Errout
		}

		metrics["bytes_sent"] = totalBytesSent
		metrics["bytes_recv"] = totalBytesRecv
		metrics["packets_sent"] = totalPacketsSent
		metrics["packets_recv"] = totalPacketsRecv
		metrics["errors_in"] = totalErrin
		metrics["errors_out"] = totalErrout
	}

	// 网络连接统计
	if connections, err := net.Connections("all"); err == nil {
		metrics["connections_total"] = len(connections)

		// 按状态统计连接
		statusCount := make(map[string]int)
		for _, conn := range connections {
			statusCount[conn.Status]++
		}
		metrics["connections_by_status"] = statusCount
	}

	return metrics, nil
}

// collectGoRuntimeMetrics 收集Go运行时指标
func (c *SystemMetricsCollector) collectGoRuntimeMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 内存统计
	metrics["alloc_bytes"] = m.Alloc
	metrics["total_alloc_bytes"] = m.TotalAlloc
	metrics["sys_bytes"] = m.Sys
	metrics["lookups"] = m.Lookups
	metrics["mallocs"] = m.Mallocs
	metrics["frees"] = m.Frees

	// 堆内存
	metrics["heap_alloc_bytes"] = m.HeapAlloc
	metrics["heap_sys_bytes"] = m.HeapSys
	metrics["heap_idle_bytes"] = m.HeapIdle
	metrics["heap_inuse_bytes"] = m.HeapInuse
	metrics["heap_released_bytes"] = m.HeapReleased
	metrics["heap_objects"] = m.HeapObjects

	// GC统计
	metrics["gc_cycles"] = m.NumGC
	metrics["gc_pause_total_ns"] = m.PauseTotalNs
	metrics["gc_pause_ns"] = m.PauseNs[(m.NumGC+255)%256]
	metrics["gc_cpu_fraction"] = m.GCCPUFraction

	// Goroutine统计
	metrics["goroutines"] = runtime.NumGoroutine()
	metrics["cgo_calls"] = runtime.NumCgoCall()

	return metrics
}

// collectProcessMetrics 收集进程指标
func (c *SystemMetricsCollector) collectProcessMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 获取当前进程
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	// 进程基本信息
	if name, err := proc.Name(); err == nil {
		metrics["name"] = name
	}
	if cmdline, err := proc.Cmdline(); err == nil {
		metrics["cmdline"] = cmdline
	}
	if createTime, err := proc.CreateTime(); err == nil {
		metrics["create_time"] = createTime
	}

	// CPU使用率
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		metrics["cpu_percent"] = cpuPercent
	}

	// 内存使用
	if memInfo, err := proc.MemoryInfo(); err == nil {
		metrics["memory_rss"] = memInfo.RSS
		metrics["memory_vms"] = memInfo.VMS
		metrics["memory_swap"] = memInfo.Swap
	}

	// 文件描述符
	if numFDs, err := proc.NumFDs(); err == nil {
		metrics["file_descriptors"] = numFDs
	}

	// 线程数
	if numThreads, err := proc.NumThreads(); err == nil {
		metrics["threads"] = numThreads
	}

	// 打开的文件
	if openFiles, err := proc.OpenFiles(); err == nil {
		metrics["open_files"] = len(openFiles)
	}

	return metrics, nil
}

// SaveMetrics 保存指标到数据库
func (c *SystemMetricsCollector) SaveMetrics(ctx context.Context, metrics map[string]interface{}) error {
	// 保存CPU指标
	if cpuMetrics, ok := metrics["cpu"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "cpu", cpuMetrics); err != nil {
			return fmt.Errorf("failed to save CPU metrics: %w", err)
		}
	}

	// 保存内存指标
	if memMetrics, ok := metrics["memory"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "memory", memMetrics); err != nil {
			return fmt.Errorf("failed to save memory metrics: %w", err)
		}
	}

	// 保存磁盘指标
	if diskMetrics, ok := metrics["disk"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "disk", diskMetrics); err != nil {
			return fmt.Errorf("failed to save disk metrics: %w", err)
		}
	}

	// 保存网络指标
	if netMetrics, ok := metrics["network"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "network", netMetrics); err != nil {
			return fmt.Errorf("failed to save network metrics: %w", err)
		}
	}

	// 保存Go运行时指标
	if goMetrics, ok := metrics["goruntime"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "goruntime", goMetrics); err != nil {
			return fmt.Errorf("failed to save Go runtime metrics: %w", err)
		}
	}

	// 保存进程指标
	if procMetrics, ok := metrics["process"].(map[string]interface{}); ok {
		if err := c.saveMetricToDB(ctx, "system", "process", procMetrics); err != nil {
			return fmt.Errorf("failed to save process metrics: %w", err)
		}
	}

	return nil
}

// saveMetricToDB 保存指标到数据库
func (c *SystemMetricsCollector) saveMetricToDB(ctx context.Context, metricType, metricName string, data map[string]interface{}) error {
	for key, value := range data {
		metric := &Models.MonitoringMetric{
			Type:        metricType,
			Name:        fmt.Sprintf("%s_%s", metricName, key),
			Value:       c.convertToFloat64(value),
			Unit:        c.getUnit(key),
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("System metric: %s %s", metricName, key),
			Status:      "normal",
			Severity:    "info",
		}

		if err := c.service.DB.(*gorm.DB).WithContext(ctx).Create(metric).Error; err != nil {
			return err
		}
	}

	return nil
}

// convertToFloat64 转换为float64
func (c *SystemMetricsCollector) convertToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		return 0.0
	}
}

// getUnit 获取单位
func (c *SystemMetricsCollector) getUnit(key string) string {
	switch {
	case contains(key, "percent") || contains(key, "usage"):
		return "%"
	case contains(key, "bytes") || contains(key, "memory"):
		return "bytes"
	case contains(key, "count") || contains(key, "threads") || contains(key, "goroutines"):
		return "count"
	case contains(key, "frequency") || contains(key, "mhz"):
		return "MHz"
	case contains(key, "time") || contains(key, "ns"):
		return "ns"
	default:
		return ""
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

// containsSubstring 检查字符串是否包含子字符串
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
