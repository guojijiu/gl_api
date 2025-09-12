package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("🚀 开始性能测试...")

	// 测试并发性能
	testConcurrencyPerformance()

	// 测试内存使用
	testMemoryUsage()

	fmt.Println("✅ 性能测试完成")
}

// testConcurrencyPerformance 测试并发性能
func testConcurrencyPerformance() {
	fmt.Println("\n⚡ 测试并发性能...")

	// 测试并发计算
	concurrency := 100
	requestsPerGoroutine := 100

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				// 简单的计算任务
				_ = j * j
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	totalRequests := concurrency * requestsPerGoroutine

	fmt.Printf("并发性能测试结果:\n")
	fmt.Printf("  并发数: %d\n", concurrency)
	fmt.Printf("  每协程请求数: %d\n", requestsPerGoroutine)
	fmt.Printf("  总请求数: %d\n", totalRequests)
	fmt.Printf("  总耗时: %v\n", duration)
	fmt.Printf("  平均耗时: %v/次\n", duration/time.Duration(totalRequests))
	fmt.Printf("  吞吐量: %.2f 请求/秒\n", float64(totalRequests)/duration.Seconds())
}

// testMemoryUsage 测试内存使用
func testMemoryUsage() {
	fmt.Println("\n🧠 测试内存使用...")

	// 获取初始内存使用
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// 创建大量对象
	objects := make([]string, 100000)
	for i := 0; i < 100000; i++ {
		objects[i] = fmt.Sprintf("object_%d", i)
	}

	// 获取创建对象后的内存使用
	var afterMem runtime.MemStats
	runtime.ReadMemStats(&afterMem)

	// 强制垃圾回收
	runtime.GC()

	// 获取GC后的内存使用
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)

	fmt.Printf("内存使用测试结果:\n")
	fmt.Printf("  初始内存: %d bytes (%.2f MB)\n", initialMem.HeapAlloc, float64(initialMem.HeapAlloc)/1024/1024)
	fmt.Printf("  创建对象后: %d bytes (%.2f MB)\n", afterMem.HeapAlloc, float64(afterMem.HeapAlloc)/1024/1024)
	fmt.Printf("  GC后内存: %d bytes (%.2f MB)\n", finalMem.HeapAlloc, float64(finalMem.HeapAlloc)/1024/1024)
	fmt.Printf("  内存增长: %d bytes (%.2f MB)\n", afterMem.HeapAlloc-initialMem.HeapAlloc, float64(afterMem.HeapAlloc-initialMem.HeapAlloc)/1024/1024)
	fmt.Printf("  GC次数: %d\n", finalMem.NumGC-initialMem.NumGC)
}
