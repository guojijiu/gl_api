package Testing

import (
	"cloud-platform-api/app/Config"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceTest 性能测试结构
type PerformanceTest struct {
	// 测试配置
	Config *Config.PerformanceTestConfig
	// 测试结果
	Results *PerformanceResults
	// 测试状态
	Status string
	// 测试开始时间
	StartTime time.Time
	// 测试结束时间
	EndTime time.Time
	// 并发控制
	Wg sync.WaitGroup
	// 保护 ResponseTimes 的互斥锁
	responseTimesMu sync.Mutex
	// 上下文
	Ctx context.Context
	// 取消函数
	Cancel context.CancelFunc
}

// PerformanceResults 性能测试结果
type PerformanceResults struct {
	// 总请求数
	TotalRequests int64
	// 成功请求数
	SuccessfulRequests int64
	// 失败请求数
	FailedRequests int64
	// 响应时间统计
	ResponseTimes []time.Duration
	// 平均响应时间
	AverageResponseTime time.Duration
	// 95%响应时间
	P95ResponseTime time.Duration
	// 99%响应时间
	P99ResponseTime time.Duration
	// 最小响应时间
	MinResponseTime time.Duration
	// 最大响应时间
	MaxResponseTime time.Duration
	// 吞吐量 (请求/秒)
	Throughput float64
	// 错误率
	ErrorRate float64
	// 测试持续时间
	Duration time.Duration
	// 并发用户数
	ConcurrentUsers int
	// 请求间隔
	RequestInterval time.Duration
	// 性能指标
	Metrics map[string]interface{}
	// 测试时间戳
	Timestamp time.Time
}

// PerformanceTestRequest 性能测试请求
type PerformanceTestRequest struct {
	// 请求ID
	ID int64
	// 请求开始时间
	StartTime time.Time
	// 请求结束时间
	EndTime time.Time
	// 响应时间
	ResponseTime time.Duration
	// 请求状态
	Status string
	// 错误信息
	Error error
	// 请求数据
	Data map[string]interface{}
}

// NewPerformanceTest 创建新的性能测试
func NewPerformanceTest(config *Config.PerformanceTestConfig) *PerformanceTest {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)

	return &PerformanceTest{
		Config: config,
		Results: &PerformanceResults{
			ResponseTimes: make([]time.Duration, 0),
			Metrics:       make(map[string]interface{}),
		},
		Status: "ready",
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// RunLoadTest 运行负载测试
func (pt *PerformanceTest) RunLoadTest(testFunc func() error) *PerformanceResults {
	pt.Status = "running"
	pt.StartTime = time.Now()

	// 启动并发用户
	for i := 0; i < pt.Config.ConcurrentUsers; i++ {
		pt.Wg.Add(1)
		go pt.runUser(testFunc, i)
	}

	// 等待所有用户完成或超时
	pt.Wg.Wait()

	pt.EndTime = time.Now()
	pt.Status = "completed"

	// 计算测试结果
	pt.calculateResults()

	return pt.Results
}

// RunStressTest 运行压力测试
func (pt *PerformanceTest) RunStressTest(testFunc func() error, maxUsers int, step int) *PerformanceResults {
	pt.Status = "running"
	pt.StartTime = time.Now()

	var results []*PerformanceResults

	// 逐步增加并发用户数
	for users := step; users <= maxUsers; users += step {
		pt.Config.ConcurrentUsers = users
		result := pt.RunLoadTest(testFunc)
		results = append(results, result)

		// 检查是否达到性能阈值
		if pt.checkPerformanceThresholds(result) {
			break
		}

		// 等待一段时间再进行下一轮测试
		time.Sleep(5 * time.Second)
	}

	pt.EndTime = time.Now()
	pt.Status = "completed"

	// 合并所有结果
	pt.mergeResults(results)

	return pt.Results
}

// RunConcurrencyTest 运行并发测试
func (pt *PerformanceTest) RunConcurrencyTest(testFunc func() error, concurrencyLevels []int) *PerformanceResults {
	pt.Status = "running"
	pt.StartTime = time.Now()

	var results []*PerformanceResults

	// 测试不同并发级别
	for _, users := range concurrencyLevels {
		pt.Config.ConcurrentUsers = users
		result := pt.RunLoadTest(testFunc)
		results = append(results, result)

		// 等待一段时间再进行下一轮测试
		time.Sleep(2 * time.Second)
	}

	pt.EndTime = time.Now()
	pt.Status = "completed"

	// 合并所有结果
	pt.mergeResults(results)

	return pt.Results
}

// runUser 运行单个用户测试
func (pt *PerformanceTest) runUser(testFunc func() error, userID int) {
	defer pt.Wg.Done()

	ticker := time.NewTicker(pt.Config.RequestInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pt.Ctx.Done():
			return
		case <-ticker.C:
			// 执行测试函数
			startTime := time.Now()
			err := testFunc()
			endTime := time.Now()

			// 记录请求结果
			pt.recordRequest(startTime, endTime, err)
		}
	}
}

// recordRequest 记录请求结果
//
// 功能说明：
// 1. 记录单个请求的响应时间和执行结果
// 2. 更新性能测试的统计计数器
// 3. 线程安全地更新共享数据结构
//
// 并发安全设计：
// - 使用原子操作（atomic）更新计数器，性能更好且线程安全
// - 使用互斥锁保护slice的append操作，因为slice不是线程安全的
// - 锁的粒度尽可能小，只保护必要的操作
//
// 性能考虑：
// - 原子操作比互斥锁性能更好，适合简单的计数器操作
// - slice的append需要加锁，因为可能触发底层数组扩容和复制
// - 如果响应时间数据量很大，可以考虑使用channel收集，减少锁竞争
func (pt *PerformanceTest) recordRequest(startTime, endTime time.Time, err error) {
	// 计算响应时间（从请求开始到结束的耗时）
	responseTime := endTime.Sub(startTime)

	// 使用原子操作更新总请求数
	// 原子操作是线程安全的，不需要加锁，性能更好
	// AddInt64是原子性的：读取-修改-写入三个操作作为一个整体执行
	atomic.AddInt64(&pt.Results.TotalRequests, 1)

	// 根据请求结果更新成功或失败计数器
	// 同样使用原子操作，保证线程安全
	if err != nil {
		atomic.AddInt64(&pt.Results.FailedRequests, 1)
	} else {
		atomic.AddInt64(&pt.Results.SuccessfulRequests, 1)
	}

	// 记录响应时间到slice中
	// 注意：slice的append操作不是线程安全的，必须加锁保护
	// 原因：
	// 1. append可能触发底层数组扩容，需要重新分配内存和复制数据
	// 2. 多个goroutine同时append可能导致数据丢失或panic
	// 3. 即使不扩容，多个goroutine同时修改slice的len字段也会导致竞态条件
	pt.responseTimesMu.Lock()
	pt.Results.ResponseTimes = append(pt.Results.ResponseTimes, responseTime)
	pt.responseTimesMu.Unlock()
	// 注意：锁的粒度要尽可能小，只保护必要的操作
	// 这里只保护append操作，计数器使用原子操作，避免锁竞争
}

// calculateResults 计算测试结果
//
// 功能说明：
// 1. 计算性能测试的所有统计指标
// 2. 包括响应时间统计、吞吐量、错误率等
// 3. 在测试完成后调用，生成完整的测试报告
//
// 计算的指标：
// - 响应时间统计：最小值、最大值、平均值、P95、P99
// - 吞吐量：每秒处理的请求数（QPS）
// - 错误率：失败请求占总请求的百分比
// - 测试配置：并发用户数、请求间隔等
//
// 计算顺序：
// 1. 先计算响应时间统计（需要遍历所有响应时间数据）
// 2. 再计算吞吐量和错误率（基于总请求数和耗时）
// 3. 最后记录测试配置和时间戳
//
// 注意事项：
// - 如果响应时间数据为空，直接返回（避免除零错误）
// - 吞吐量计算需要确保Duration不为0
// - 错误率计算需要确保TotalRequests不为0
func (pt *PerformanceTest) calculateResults() {
	// 边界检查：如果响应时间数据为空，无法计算统计指标
	if len(pt.Results.ResponseTimes) == 0 {
		return
	}

	// 计算响应时间统计（最小值、最大值、平均值、百分位数）
	// 这会遍历所有响应时间数据，如果数据量很大可能较慢
	pt.calculateResponseTimeStats()

	// 计算测试总耗时
	pt.Results.Duration = pt.EndTime.Sub(pt.StartTime)
	
	// 计算吞吐量（每秒请求数，QPS）
	// 公式：吞吐量 = 总请求数 / 总耗时（秒）
	// 注意：如果Duration为0，会导致除零错误，但通常不会发生
	pt.Results.Throughput = float64(pt.Results.TotalRequests) / pt.Results.Duration.Seconds()

	// 计算错误率（失败请求占总请求的百分比）
	// 公式：错误率 = 失败请求数 / 总请求数 * 100%
	// 注意：需要检查TotalRequests不为0，避免除零错误
	if pt.Results.TotalRequests > 0 {
		pt.Results.ErrorRate = float64(pt.Results.FailedRequests) / float64(pt.Results.TotalRequests)
	}

	// 记录测试配置信息（用于报告和对比）
	pt.Results.ConcurrentUsers = pt.Config.ConcurrentUsers      // 并发用户数
	pt.Results.RequestInterval = pt.Config.RequestInterval      // 请求间隔
	pt.Results.Timestamp = time.Now()                            // 测试完成时间戳
}

// calculateResponseTimeStats 计算响应时间统计
//
// 功能说明：
// 1. 计算响应时间的各种统计指标（最小值、最大值、平均值、百分位数）
// 2. 这些指标用于评估系统性能的各个方面
//
// 统计指标说明：
// - MinResponseTime：最小响应时间，表示最快请求的耗时
// - MaxResponseTime：最大响应时间，表示最慢请求的耗时
// - AverageResponseTime：平均响应时间，所有请求的平均耗时
// - P95ResponseTime：95%的请求响应时间小于等于该值
// - P99ResponseTime：99%的请求响应时间小于等于该值
//
// 算法说明：
// - 先复制数据并排序（为了计算百分位数）
// - 遍历一次数据计算最小、最大和总和
// - 使用总和除以数量计算平均值
// - 使用排序后的数据计算百分位数
//
// 性能考虑：
// - 复制数据避免修改原始数据
// - 使用简单的冒泡排序（数据量通常不大）
// - 如果数据量很大，可以考虑使用更高效的排序算法
func (pt *PerformanceTest) calculateResponseTimeStats() {
	// 边界检查：如果数据为空，直接返回
	if len(pt.Results.ResponseTimes) == 0 {
		return
	}

	// 复制响应时间数据并排序
	// 注意：必须复制，因为排序会修改原数组
	// 排序是为了计算百分位数（P95、P99）
	times := make([]time.Duration, len(pt.Results.ResponseTimes))
	copy(times, pt.Results.ResponseTimes)
	
	// 对响应时间进行排序（升序）
	// 使用简单的冒泡排序，适合小数据量
	// 如果数据量很大（>1000），建议使用sort包的标准排序
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	// 计算最小和最大响应时间
	// 排序后，第一个元素是最小值，最后一个元素是最大值
	pt.Results.MinResponseTime = times[0]
	pt.Results.MaxResponseTime = times[0] // 初始化为第一个元素

	// 遍历所有响应时间，计算总和并更新最值
	// 虽然已经排序，但为了代码清晰，仍然遍历一次
	var total time.Duration
	for _, t := range times {
		// 更新最小值（虽然已排序，但为了代码通用性仍检查）
		if t < pt.Results.MinResponseTime {
			pt.Results.MinResponseTime = t
		}
		// 更新最大值
		if t > pt.Results.MaxResponseTime {
			pt.Results.MaxResponseTime = t
		}
		// 累加总和，用于计算平均值
		total += t
	}

	// 计算平均响应时间
	// 公式：平均值 = 总和 / 数量
	pt.Results.AverageResponseTime = total / time.Duration(len(times))

	// 计算百分位数（需要排序后的数据）
	// P95：95%的请求响应时间小于等于该值
	pt.Results.P95ResponseTime = pt.calculatePercentile(times, 95)
	// P99：99%的请求响应时间小于等于该值（识别极端情况）
	pt.Results.P99ResponseTime = pt.calculatePercentile(times, 99)
}

// calculatePercentile 计算百分位数
//
// 功能说明：
// 1. 计算响应时间分布的百分位数值（如P95、P99）
// 2. 百分位数表示有X%的请求响应时间小于等于该值
// 3. 用于评估系统性能的稳定性，识别长尾问题
//
// 算法说明：
// - 假设times已经按升序排序
// - 百分位数位置 = (数据总数 - 1) * 百分位 / 100
// - 例如：100个数据，P95位置 = (100-1) * 95 / 100 = 94.05 ≈ 94
// - 取第94个元素（0-based索引）作为P95值
//
// 边界处理：
// - 如果数据为空，返回0
// - 如果计算出的索引超出范围，使用最后一个元素
// - 这确保了即使数据量很小也能返回有效值
//
// 使用场景：
// - P50（中位数）：一半请求的响应时间
// - P95：95%的请求响应时间小于等于该值
// - P99：99%的请求响应时间小于等于该值（识别极端情况）
func (pt *PerformanceTest) calculatePercentile(times []time.Duration, percentile int) time.Duration {
	// 边界检查：如果数据为空，返回0
	if len(times) == 0 {
		return 0
	}

	// 计算百分位数的索引位置
	// 公式：index = (n-1) * percentile / 100
	// 例如：100个数据，P95的index = 99 * 95 / 100 = 94.05，取整为94
	// 注意：使用len(times)-1是因为索引从0开始，最后一个元素的索引是len-1
	index := int(float64(len(times)-1) * float64(percentile) / 100.0)
	
	// 边界检查：确保索引不超出范围
	// 虽然理论上不应该超出，但浮点数计算可能有精度问题
	if index >= len(times) {
		index = len(times) - 1 // 使用最后一个元素
	}

	// 返回对应索引位置的响应时间值
	// 注意：这里假设times已经按升序排序
	return times[index]
}

// checkPerformanceThresholds 检查性能阈值
func (pt *PerformanceTest) checkPerformanceThresholds(result *PerformanceResults) bool {
	thresholds := pt.Config.Thresholds

	// 检查平均响应时间
	if result.AverageResponseTime > thresholds.AvgResponseTime {
		return true
	}

	// 检查95%响应时间
	if result.P95ResponseTime > thresholds.P95ResponseTime {
		return true
	}

	// 检查99%响应时间
	if result.P99ResponseTime > thresholds.P99ResponseTime {
		return true
	}

	// 检查错误率
	if result.ErrorRate > thresholds.ErrorRate {
		return true
	}

	// 检查吞吐量
	if result.Throughput < float64(thresholds.Throughput) {
		return true
	}

	return false
}

// mergeResults 合并多个测试结果
//
// 功能说明：
// 1. 将多个测试阶段的结果合并为一个整体结果
// 2. 用于压力测试和并发测试，这些测试会分阶段执行
// 3. 合并后重新计算所有统计指标
//
// 合并策略：
// - 累加所有阶段的请求数（总数、成功数、失败数）
// - 合并所有阶段的响应时间数据
// - 重新计算平均值、百分位数等统计指标
//
// 并发安全：
// - 使用responseTimesMu锁保护ResponseTimes的append操作
// - 虽然通常只在测试完成后调用，但加锁保证线程安全
//
// 注意事项：
// - 合并后需要重新计算统计指标，因为数据量发生了变化
// - 响应时间数据会合并，可能导致内存占用增加
// - 如果数据量很大，可以考虑只保留统计值而不保留原始数据
func (pt *PerformanceTest) mergeResults(results []*PerformanceResults) {
	// 边界检查：如果结果列表为空，直接返回
	if len(results) == 0 {
		return
	}

	// 合并基础统计（需要加锁保护）
	// 虽然通常只在测试完成后调用，但加锁保证线程安全
	pt.responseTimesMu.Lock()
	defer pt.responseTimesMu.Unlock()
	
	// 遍历所有测试结果，累加统计数据
	for _, result := range results {
		// 累加总请求数
		pt.Results.TotalRequests += result.TotalRequests
		// 累加成功请求数
		pt.Results.SuccessfulRequests += result.SuccessfulRequests
		// 累加失败请求数
		pt.Results.FailedRequests += result.FailedRequests
		// 合并响应时间数据（使用append的展开语法）
		// 注意：这会增加内存占用，如果数据量很大可能需要优化
		pt.Results.ResponseTimes = append(pt.Results.ResponseTimes, result.ResponseTimes...)
	}

	// 重新计算统计结果
	// 因为数据量发生了变化，需要重新计算平均值、百分位数等指标
	// 这会遍历所有响应时间数据，如果数据量很大可能较慢
	pt.calculateResults()
}

// Stop 停止性能测试
func (pt *PerformanceTest) Stop() {
	if pt.Status == "running" {
		pt.Cancel()
		pt.Status = "stopped"
	}
}

// GetResults 获取测试结果
func (pt *PerformanceTest) GetResults() *PerformanceResults {
	return pt.Results
}

// PrintResults 打印测试结果
func (pt *PerformanceTest) PrintResults() {
	results := pt.Results

	fmt.Printf("\n=== 性能测试结果 ===\n")
	fmt.Printf("测试状态: %s\n", pt.Status)
	fmt.Printf("测试时间: %s - %s\n", pt.StartTime.Format("2006-01-02 15:04:05"), pt.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("测试持续时间: %v\n", results.Duration)
	fmt.Printf("并发用户数: %d\n", results.ConcurrentUsers)
	fmt.Printf("请求间隔: %v\n", results.RequestInterval)
	fmt.Printf("\n--- 请求统计 ---\n")
	fmt.Printf("总请求数: %d\n", results.TotalRequests)
	fmt.Printf("成功请求数: %d\n", results.SuccessfulRequests)
	fmt.Printf("失败请求数: %d\n", results.FailedRequests)
	fmt.Printf("错误率: %.2f%%\n", results.ErrorRate*100)
	fmt.Printf("\n--- 响应时间统计 ---\n")
	fmt.Printf("平均响应时间: %v\n", results.AverageResponseTime)
	fmt.Printf("95%%响应时间: %v\n", results.P95ResponseTime)
	fmt.Printf("99%%响应时间: %v\n", results.P99ResponseTime)
	fmt.Printf("最小响应时间: %v\n", results.MinResponseTime)
	fmt.Printf("最大响应时间: %v\n", results.MaxResponseTime)
	fmt.Printf("\n--- 性能指标 ---\n")
	fmt.Printf("吞吐量: %.2f 请求/秒\n", results.Throughput)
	fmt.Printf("==================\n\n")
}

// AssertPerformanceThresholds 断言性能阈值
func (pt *PerformanceTest) AssertPerformanceThresholds(t *testing.T) {
	results := pt.Results
	thresholds := pt.Config.Thresholds

	// 断言响应时间阈值
	assert.LessOrEqual(t, results.AverageResponseTime, thresholds.AvgResponseTime,
		"平均响应时间超过阈值")
	assert.LessOrEqual(t, results.P95ResponseTime, thresholds.P95ResponseTime,
		"95%%响应时间超过阈值")
	assert.LessOrEqual(t, results.P99ResponseTime, thresholds.P99ResponseTime,
		"99%%响应时间超过阈值")

	// 断言错误率阈值
	assert.LessOrEqual(t, results.ErrorRate, thresholds.ErrorRate,
		"错误率超过阈值")

	// 断言吞吐量阈值
	assert.GreaterOrEqual(t, results.Throughput, float64(thresholds.Throughput),
		"吞吐量低于阈值")
}

// SaveResultsToFile 保存结果到文件
func (pt *PerformanceTest) SaveResultsToFile(filename string) error {
	// 这里可以实现将结果保存为JSON、CSV或其他格式
	// 暂时返回nil，具体实现可以根据需要添加
	return nil
}

// LoadTestExample 负载测试示例
func LoadTestExample(t *testing.T) {
	// 创建性能测试配置
	config := &Config.PerformanceTestConfig{
		ConcurrentUsers:    10,
		Duration:           30 * time.Second,
		RequestInterval:    100 * time.Millisecond,
		RecordResponseTime: true,
		Thresholds: Config.PerformanceThresholds{
			AvgResponseTime: 100 * time.Millisecond,
			P95ResponseTime: 200 * time.Millisecond,
			P99ResponseTime: 500 * time.Millisecond,
			ErrorRate:       0.01,
			Throughput:      100,
		},
	}

	// 创建性能测试
	pt := NewPerformanceTest(config)

	// 定义测试函数
	testFunc := func() error {
		// 模拟API请求
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	// 运行负载测试
	results := pt.RunLoadTest(testFunc)

	// 打印结果
	pt.PrintResults()

	// 断言性能阈值
	pt.AssertPerformanceThresholds(t)

	// 验证结果
	require.NotNil(t, results)
	require.Greater(t, results.TotalRequests, int64(0))
	require.Less(t, results.ErrorRate, 0.1)
}
