package main

import (
	"cloud-platform-api/app/Storage"
	"fmt"
	"log"
	"time"
)

func main() {
	// 创建Storage管理器
	storageManager := Storage.NewStorageManager("./storage")

	fmt.Println("=== Storage 功能演示 ===")

	// 1. 日志功能演示
	fmt.Println("\n1. 记录日志...")
	err := storageManager.LogInfo("应用程序启动", map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0.0",
		"user":      "admin",
	})
	if err != nil {
		log.Printf("记录日志失败: %v", err)
	}

	err = storageManager.LogWarning("系统资源使用率较高", map[string]interface{}{
		"cpu_usage":    75.5,
		"memory_usage": 80.2,
		"disk_usage":   65.8,
	})
	if err != nil {
		log.Printf("记录警告日志失败: %v", err)
	}

	// 2. 缓存功能演示
	fmt.Println("2. 缓存操作...")
	
	// 设置缓存
	err = storageManager.Cache("user:123", map[string]interface{}{
		"id":       123,
		"name":     "张三",
		"email":    "zhangsan@example.com",
		"role":     "user",
		"last_login": time.Now().Format("2006-01-02 15:04:05"),
	}, 30*time.Minute)
	if err != nil {
		log.Printf("设置缓存失败: %v", err)
	}

	// 获取缓存
	if cached, err := storageManager.GetCache("user:123"); err == nil {
		fmt.Printf("从缓存获取用户信息: %+v\n", cached)
	} else {
		fmt.Printf("获取缓存失败: %v\n", err)
	}

	// 3. 临时文件功能演示
	fmt.Println("3. 临时文件操作...")
	
	// 创建临时文件
	tempFile, err := storageManager.CreateTempFile("example")
	if err != nil {
		log.Printf("创建临时文件失败: %v", err)
	} else {
		defer tempFile.Close()
		
		// 写入一些数据
		_, err = tempFile.WriteString("这是临时文件的内容\n")
		if err != nil {
			log.Printf("写入临时文件失败: %v", err)
		}
		
		fmt.Printf("临时文件已创建: %s\n", tempFile.Name())
	}

	// 4. 获取存储信息
	fmt.Println("4. 存储信息...")
	info := storageManager.GetStorageInfo()
	fmt.Printf("存储基础路径: %s\n", info["base_path"])
	fmt.Printf("公共文件路径: %s\n", info["public_path"])
	fmt.Printf("私有文件路径: %s\n", info["private_path"])
	fmt.Printf("日志路径: %s\n", info["log_path"])
	fmt.Printf("缓存路径: %s\n", info["cache_path"])
	fmt.Printf("临时文件路径: %s\n", info["temp_path"])
	fmt.Printf("临时文件数量: %d\n", info["temp_files"])
	fmt.Printf("临时文件总大小: %d 字节\n", info["temp_size"])

	// 5. 清理操作演示
	fmt.Println("5. 清理操作...")
	
	// 清理临时文件
	err = storageManager.CleanTempFiles()
	if err != nil {
		log.Printf("清理临时文件失败: %v", err)
	} else {
		fmt.Println("临时文件已清理")
	}

	// 清空缓存
	err = storageManager.ClearCache()
	if err != nil {
		log.Printf("清空缓存失败: %v", err)
	} else {
		fmt.Println("缓存已清空")
	}

	fmt.Println("\n=== 演示完成 ===")
}
