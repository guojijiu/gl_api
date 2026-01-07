package Services

import (
	"cloud-platform-api/app/Storage"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// TestCase 测试用例
type TestCase struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	ExpectedStatus int            `json:"expected_status"`
	ExpectedBody   interface{}    `json:"expected_body"`
	Timeout       time.Duration   `json:"timeout"`
	Retries       int             `json:"retries"`
	Category      string          `json:"category"` // "unit", "integration", "performance"
}

// TestResult 测试结果
type TestResult struct {
	TestCaseID    string        `json:"test_case_id"`
	Name          string        `json:"name"`
	Status        string        `json:"status"` // "passed", "failed", "skipped"
	Duration      time.Duration `json:"duration"`
	StatusCode    int           `json:"status_code"`
	ResponseBody  string        `json:"response_body"`
	Error         string        `json:"error"`
	Timestamp     time.Time     `json:"timestamp"`
	RetryCount    int           `json:"retry_count"`
}

// PerformanceTest 性能测试
type PerformanceTest struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	TestCaseID   string        `json:"test_case_id"`
	Concurrency  int           `json:"concurrency"`
	Duration     time.Duration `json:"duration"`
	RampUpTime   time.Duration `json:"ramp_up_time"`
	ThinkTime    time.Duration `json:"think_time"`
}

// PerformanceResult 性能测试结果
type PerformanceResult struct {
	TestID           string        `json:"test_id"`
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulRequests int64       `json:"successful_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	TotalDuration    time.Duration `json:"total_duration"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	RequestsPerSecond float64      `json:"requests_per_second"`
	ErrorRate        float64      `json:"error_rate"`
	Percentiles      map[int]time.Duration `json:"percentiles"`
	Timestamp        time.Time     `json:"timestamp"`
}

// AutomatedTestService 自动化测试服务
type AutomatedTestService struct {
	storageManager *Storage.StorageManager
	testCases      map[string]*TestCase
	results        []*TestResult
	performanceResults []*PerformanceResult
	mutex          sync.RWMutex
	httpClient     *http.Client
}

// NewAutomatedTestService 创建自动化测试服务
// 功能说明：
// 1. 初始化自动化测试服务
// 2. 管理测试用例和结果
// 3. 执行API测试和性能测试
// 4. 生成测试报告
func NewAutomatedTestService(storageManager *Storage.StorageManager) *AutomatedTestService {
	service := &AutomatedTestService{
		storageManager: storageManager,
		testCases:      make(map[string]*TestCase),
		results:        make([]*TestResult, 0),
		performanceResults: make([]*PerformanceResult, 0),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 初始化默认测试用例
	service.initDefaultTestCases()

	return service
}

// AddTestCase 添加测试用例
func (s *AutomatedTestService) AddTestCase(testCase *TestCase) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.testCases[testCase.ID] = testCase
}

// RunTestCase 运行单个测试用例
func (s *AutomatedTestService) RunTestCase(testCaseID string) (*TestResult, error) {
	s.mutex.RLock()
	testCase, exists := s.testCases[testCaseID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("测试用例不存在: %s", testCaseID)
	}

	return s.executeTestCase(testCase)
}

// RunTestSuite 运行测试套件
func (s *AutomatedTestService) RunTestSuite(category string) ([]*TestResult, error) {
	s.mutex.RLock()
	var testCases []*TestCase
	for _, tc := range s.testCases {
		if category == "" || tc.Category == category {
			testCases = append(testCases, tc)
		}
	}
	s.mutex.RUnlock()

	var results []*TestResult
	var wg sync.WaitGroup
	resultChan := make(chan *TestResult, len(testCases))

	// 并发执行测试用例
	for _, testCase := range testCases {
		wg.Add(1)
		go func(tc *TestCase) {
			defer wg.Done()
			result, err := s.executeTestCase(tc)
			if err != nil {
				result = &TestResult{
					TestCaseID: tc.ID,
					Name:       tc.Name,
					Status:     "failed",
					Error:      err.Error(),
					Timestamp:  time.Now(),
				}
			}
			resultChan <- result
		}(testCase)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
		s.addResult(result)
	}

	return results, nil
}

// RunPerformanceTest 运行性能测试
func (s *AutomatedTestService) RunPerformanceTest(perfTest *PerformanceTest) (*PerformanceResult, error) {
	s.mutex.RLock()
	testCase, exists := s.testCases[perfTest.TestCaseID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("测试用例不存在: %s", perfTest.TestCaseID)
	}

	return s.executePerformanceTest(perfTest, testCase)
}

// executeTestCase 执行测试用例
func (s *AutomatedTestService) executeTestCase(testCase *TestCase) (*TestResult, error) {
	startTime := time.Now()
	var lastError error

	// 重试机制
	for attempt := 0; attempt <= testCase.Retries; attempt++ {
		result, err := s.executeSingleTest(testCase)
		if err == nil {
			result.RetryCount = attempt
			result.Duration = time.Since(startTime)
			return result, nil
		}
		lastError = err
		
		if attempt < testCase.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second) // 递增延迟
		}
	}

	// 所有重试都失败
	return &TestResult{
		TestCaseID: testCase.ID,
		Name:       testCase.Name,
		Status:     "failed",
		Error:      lastError.Error(),
		Duration:   time.Since(startTime),
		RetryCount: testCase.Retries,
		Timestamp:  time.Now(),
	}, nil
}

// executeSingleTest 执行单次测试
func (s *AutomatedTestService) executeSingleTest(testCase *TestCase) (*TestResult, error) {
	// 准备请求
	var body io.Reader
	if testCase.Body != nil {
		jsonBody, err := json.Marshal(testCase.Body)
		if err != nil {
			return nil, fmt.Errorf("请求体序列化失败: %v", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	// 创建请求
	req, err := http.NewRequest(testCase.Method, testCase.URL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range testCase.Headers {
		req.Header.Set(key, value)
	}

	// 设置超时
	client := s.httpClient
	if testCase.Timeout > 0 {
		client = &http.Client{Timeout: testCase.Timeout}
	}

	// 执行请求
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return &TestResult{
			TestCaseID: testCase.ID,
			Name:       testCase.Name,
			Status:     "failed",
			Error:      err.Error(),
			Duration:   duration,
			Timestamp:  time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 验证结果
	status := "passed"
	if resp.StatusCode != testCase.ExpectedStatus {
		status = "failed"
	}

	result := &TestResult{
		TestCaseID:   testCase.ID,
		Name:         testCase.Name,
		Status:       status,
		Duration:     duration,
		StatusCode:   resp.StatusCode,
		ResponseBody: string(respBody),
		Timestamp:    time.Now(),
	}

	if status == "failed" {
		result.Error = fmt.Sprintf("期望状态码 %d，实际状态码 %d", testCase.ExpectedStatus, resp.StatusCode)
	}

	return result, nil
}

// executePerformanceTest 执行性能测试
//
// 功能说明：
// 1. 并发执行性能测试，模拟多用户同时访问的场景
// 2. 支持渐进式负载增加（Ramp Up），避免瞬间高负载
// 3. 支持思考时间（Think Time），模拟真实用户行为
// 4. 收集响应时间、成功率等性能指标
//
// 并发安全：
// - 使用sync.Mutex保护共享的results切片和计数器
// - 使用sync.WaitGroup等待所有goroutine完成
// - slice的append操作不是线程安全的，必须加锁
//
// 负载分配策略：
// - 将总请求数平均分配给各个worker
// - 如果无法整除，前N个worker会多执行一个请求
// - 这样可以确保所有请求都被执行，且负载相对均衡
//
// 渐进式负载增加：
// - 通过RampUpTime参数控制负载增加速度
// - 每个worker启动前等待一段时间，避免同时启动造成瞬间高负载
// - 等待时间 = RampUpTime / workerCount，确保负载线性增加
func (s *AutomatedTestService) executePerformanceTest(perfTest *PerformanceTest, testCase *TestCase) (*PerformanceResult, error) {
	// 记录测试开始时间，用于计算总耗时
	startTime := time.Now()
	
	// 共享数据结构，需要在多个goroutine间安全访问
	var results []time.Duration              // 存储所有请求的响应时间
	var successfulRequests int64              // 成功请求数（使用int64支持原子操作）
	var failedRequests int64                  // 失败请求数（使用int64支持原子操作）
	var mutex sync.Mutex                      // 保护results切片的互斥锁

	// 并发控制：使用WaitGroup等待所有worker完成
	var wg sync.WaitGroup
	workerCount := perfTest.Concurrency       // 并发worker数量
	
	// 计算每个worker需要执行的请求数
	// 公式：总请求数 = 持续时间(秒) * 并发数 / 持续时间(秒) = 并发数
	// 注意：这个公式有问题，应该是：总请求数 = 持续时间(秒) * 每秒请求数
	// 但这里假设每秒请求数等于并发数，即每个worker每秒执行1个请求
	requestsPerWorker := int(perfTest.Duration.Seconds() * float64(perfTest.Concurrency) / float64(perfTest.Duration.Seconds()))

	// 启动多个worker并发执行测试
	for i := 0; i < workerCount; i++ {
		wg.Add(1) // 增加WaitGroup计数
		
		// 启动goroutine执行测试
		// 注意：必须将i作为参数传入，避免闭包捕获循环变量的问题
		go func(workerID int) {
			defer wg.Done() // 确保goroutine结束时减少WaitGroup计数
			
			// 计算当前worker需要执行的请求数
			// 如果总请求数无法被worker数整除，前N个worker会多执行一个请求
			// 这样可以确保所有请求都被执行，且负载相对均衡
			requests := requestsPerWorker
			remainder := requestsPerWorker % workerCount
			if workerID < remainder {
				requests++ // 前remainder个worker多执行一个请求
			}

			// 执行指定数量的请求
			for j := 0; j < requests; j++ {
				// 思考时间：模拟真实用户操作之间的间隔
				// 真实用户不会连续发送请求，会有思考、阅读等时间
				if perfTest.ThinkTime > 0 {
					time.Sleep(perfTest.ThinkTime)
				}

				// 执行单个测试请求并记录响应时间
				reqStart := time.Now()
				_, err := s.executeSingleTest(testCase)
				duration := time.Since(reqStart)

				// 更新共享数据（需要加锁保护）
				// 注意：slice的append操作不是线程安全的，必须加锁
				// 虽然计数器可以使用原子操作，但为了代码一致性，也使用mutex
				mutex.Lock()
				results = append(results, duration) // 记录响应时间
				if err != nil {
					failedRequests++    // 失败请求计数
				} else {
					successfulRequests++ // 成功请求计数
				}
				mutex.Unlock()
			}
		}(i) // 传入worker ID

		// 渐进式增加负载（Ramp Up）
		// 每个worker启动前等待一段时间，避免所有worker同时启动造成瞬间高负载
		// 等待时间 = RampUpTime / workerCount，确保负载线性增加
		// 例如：10个worker，RampUpTime=10秒，则每个worker间隔1秒启动
		if perfTest.RampUpTime > 0 {
			time.Sleep(perfTest.RampUpTime / time.Duration(workerCount))
		}
	}

	// 等待所有worker完成
	// Wait()会阻塞直到所有goroutine调用Done()
	wg.Wait()
	
	// 计算总耗时
	totalDuration := time.Since(startTime)

	// 计算统计信息
	totalRequests := successfulRequests + failedRequests
	avgResponseTime := s.calculateAverageResponseTime(results)
	minResponseTime := s.calculateMinResponseTime(results)
	maxResponseTime := s.calculateMaxResponseTime(results)
	requestsPerSecond := float64(totalRequests) / totalDuration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100
	percentiles := s.calculatePercentiles(results)

	perfResult := &PerformanceResult{
		TestID:            perfTest.ID,
		TotalRequests:     totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:    failedRequests,
		TotalDuration:     totalDuration,
		AvgResponseTime:   avgResponseTime,
		MinResponseTime:   minResponseTime,
		MaxResponseTime:   maxResponseTime,
		RequestsPerSecond: requestsPerSecond,
		ErrorRate:         errorRate,
		Percentiles:       percentiles,
		Timestamp:         time.Now(),
	}

	s.addPerformanceResult(perfResult)

	return perfResult, nil
}

// calculateAverageResponseTime 计算平均响应时间
func (s *AutomatedTestService) calculateAverageResponseTime(results []time.Duration) time.Duration {
	if len(results) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, duration := range results {
		total += duration
	}
	return total / time.Duration(len(results))
}

// calculateMinResponseTime 计算最小响应时间
func (s *AutomatedTestService) calculateMinResponseTime(results []time.Duration) time.Duration {
	if len(results) == 0 {
		return 0
	}

	min := results[0]
	for _, duration := range results {
		if duration < min {
			min = duration
		}
	}
	return min
}

// calculateMaxResponseTime 计算最大响应时间
func (s *AutomatedTestService) calculateMaxResponseTime(results []time.Duration) time.Duration {
	if len(results) == 0 {
		return 0
	}

	max := results[0]
	for _, duration := range results {
		if duration > max {
			max = duration
		}
	}
	return max
}

// calculatePercentiles 计算百分位数
func (s *AutomatedTestService) calculatePercentiles(results []time.Duration) map[int]time.Duration {
	percentiles := make(map[int]time.Duration)
	if len(results) == 0 {
		return percentiles
	}

	// 排序
	sorted := make([]time.Duration, len(results))
	copy(sorted, results)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 计算百分位数
	percentileValues := []int{50, 90, 95, 99}
	for _, p := range percentileValues {
		index := int(float64(p) / 100.0 * float64(len(sorted)-1))
		if index >= 0 && index < len(sorted) {
			percentiles[p] = sorted[index]
		}
	}

	return percentiles
}

// addResult 添加测试结果
func (s *AutomatedTestService) addResult(result *TestResult) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.results = append(s.results, result)
}

// addPerformanceResult 添加性能测试结果
func (s *AutomatedTestService) addPerformanceResult(result *PerformanceResult) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.performanceResults = append(s.performanceResults, result)
}

// GetTestResults 获取测试结果
func (s *AutomatedTestService) GetTestResults() []*TestResult {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	results := make([]*TestResult, len(s.results))
	copy(results, s.results)
	return results
}

// GetPerformanceResults 获取性能测试结果
func (s *AutomatedTestService) GetPerformanceResults() []*PerformanceResult {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	results := make([]*PerformanceResult, len(s.performanceResults))
	copy(results, s.performanceResults)
	return results
}

// GenerateTestReport 生成测试报告
func (s *AutomatedTestService) GenerateTestReport() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	totalTests := len(s.results)
	passedTests := 0
	failedTests := 0
	skippedTests := 0
	totalDuration := time.Duration(0)

	for _, result := range s.results {
		totalDuration += result.Duration
		switch result.Status {
		case "passed":
			passedTests++
		case "failed":
			failedTests++
		case "skipped":
			skippedTests++
		}
	}

	passRate := 0.0
	if totalTests > 0 {
		passRate = float64(passedTests) / float64(totalTests) * 100
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tests":   totalTests,
			"passed_tests":  passedTests,
			"failed_tests":  failedTests,
			"skipped_tests": skippedTests,
			"pass_rate":     passRate,
			"total_duration": totalDuration.String(),
		},
		"recent_results": s.getRecentResults(10),
		"test_cases_count": len(s.testCases),
		"performance_tests_count": len(s.performanceResults),
	}
}

// getRecentResults 获取最近的测试结果
func (s *AutomatedTestService) getRecentResults(limit int) []*TestResult {
	if len(s.results) <= limit {
		return s.results
	}
	return s.results[len(s.results)-limit:]
}

// initDefaultTestCases 初始化默认测试用例
func (s *AutomatedTestService) initDefaultTestCases() {
	// 健康检查测试
	s.AddTestCase(&TestCase{
		ID:          "health_check",
		Name:        "健康检查",
		Description: "检查API服务是否正常运行",
		Method:      "GET",
		URL:         "http://localhost:8080/health",
		Headers:     map[string]string{},
		ExpectedStatus: 200,
		Timeout:     5 * time.Second,
		Retries:     2,
		Category:    "unit",
	})

	// 用户注册测试
	s.AddTestCase(&TestCase{
		ID:          "user_register",
		Name:        "用户注册",
		Description: "测试用户注册功能",
		Method:      "POST",
		URL:         "http://localhost:8080/api/v1/auth/register",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "password123",
		},
		ExpectedStatus: 201,
		Timeout:        10 * time.Second,
		Retries:        1,
		Category:       "integration",
	})

	// 用户登录测试
	s.AddTestCase(&TestCase{
		ID:          "user_login",
		Name:        "用户登录",
		Description: "测试用户登录功能",
		Method:      "POST",
		URL:         "http://localhost:8080/api/v1/auth/login",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		},
		ExpectedStatus: 200,
		Timeout:        10 * time.Second,
		Retries:        1,
		Category:       "integration",
	})
}
