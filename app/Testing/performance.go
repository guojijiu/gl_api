package Testing

import (
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
		Status:    "ready",
		Ctx:       ctx,
		Cancel:    cancel,
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
func (pt *PerformanceTest) recordRequest(startTime, endTime time.Time, err error) {
	responseTime := endTime.Sub(startTime)
	
	// 原子操作更新计数器
	atomic.AddInt64(&pt.Results.TotalRequests, 1)
	
	if err != nil {
		atomic.AddInt64(&pt.Results.FailedRequests, 1)
	} else {
		atomic.AddInt64(&pt.Results.SuccessfulRequests, 1)
	}
	
	// 记录响应时间
	pt.Results.ResponseTimes = append(pt.Results.ResponseTimes, responseTime)
}

// calculateResults 计算测试结果
func (pt *PerformanceTest) calculateResults() {
	if len(pt.Results.ResponseTimes) == 0 {
		return
	}
	
	// 计算响应时间统计
	pt.calculateResponseTimeStats()
	
	// 计算吞吐量
	pt.Results.Duration = pt.EndTime.Sub(pt.StartTime)
	pt.Results.Throughput = float64(pt.Results.TotalRequests) / pt.Results.Duration.Seconds()
	
	// 计算错误率
	if pt.Results.TotalRequests > 0 {
		pt.Results.ErrorRate = float64(pt.Results.FailedRequests) / float64(pt.Results.TotalRequests)
	}
	
	// 计算其他指标
	pt.Results.ConcurrentUsers = pt.Config.ConcurrentUsers
	pt.Results.RequestInterval = pt.Config.RequestInterval
	pt.Results.Timestamp = time.Now()
}

// calculateResponseTimeStats 计算响应时间统计
func (pt *PerformanceTest) calculateResponseTimeStats() {
	if len(pt.Results.ResponseTimes) == 0 {
		return
	}
	
	// 排序响应时间
	times := make([]time.Duration, len(pt.Results.ResponseTimes))
	copy(times, pt.Results.ResponseTimes)
	
	// 计算最小和最大响应时间
	pt.Results.MinResponseTime = times[0]
	pt.Results.MaxResponseTime = times[0]
	
	var total time.Duration
	for _, t := range times {
		if t < pt.Results.MinResponseTime {
			pt.Results.MinResponseTime = t
		}
		if t > pt.Results.MaxResponseTime {
			pt.Results.MaxResponseTime = t
		}
		total += t
	}
	
	// 计算平均响应时间
	pt.Results.AverageResponseTime = total / time.Duration(len(times))
	
	// 计算百分位数
	pt.Results.P95ResponseTime = pt.calculatePercentile(times, 95)
	pt.Results.P99ResponseTime = pt.calculatePercentile(times, 99)
}

// calculatePercentile 计算百分位数
func (pt *PerformanceTest) calculatePercentile(times []time.Duration, percentile int) time.Duration {
	if len(times) == 0 {
		return 0
	}
	
	index := int(float64(len(times)-1) * float64(percentile) / 100.0)
	if index >= len(times) {
		index = len(times) - 1
	}
	
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
func (pt *PerformanceTest) mergeResults(results []*PerformanceResults) {
	if len(results) == 0 {
		return
	}
	
	// 合并基础统计
	for _, result := range results {
		pt.Results.TotalRequests += result.TotalRequests
		pt.Results.SuccessfulRequests += result.SuccessfulRequests
		pt.Results.FailedRequests += result.FailedRequests
		pt.Results.ResponseTimes = append(pt.Results.ResponseTimes, result.ResponseTimes...)
	}
	
	// 重新计算统计结果
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
		ConcurrentUsers: 10,
		Duration:        30 * time.Second,
		RequestInterval: 100 * time.Millisecond,
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
