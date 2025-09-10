package Config

import (
	"fmt"
	"log"
)

// ExampleUsage 展示如何使用拆分后的配置
func ExampleUsage() {
	// 1. 加载配置
	LoadConfig()

	// 2. 验证配置
	if err := ValidateConfig(); err != nil {
		log.Fatal("配置验证失败:", err)
	}

	// 3. 获取全局配置
	config := GetConfig()
	fmt.Printf("全局配置: %+v\n", config)

	// 4. 使用服务器配置
	serverConfig := GetServerConfig()
	if serverConfig != nil {
		fmt.Printf("服务器端口: %s\n", serverConfig.Port)
		fmt.Printf("是否为调试模式: %t\n", serverConfig.IsDebugMode())
		fmt.Printf("完整URL: %s\n", serverConfig.GetFullURL("/api/v1"))
	}

	// 5. 使用数据库配置
	dbConfig := GetDatabaseConfig()
	if dbConfig != nil {
		fmt.Printf("数据库驱动: %s\n", dbConfig.Driver)
		fmt.Printf("数据库连接字符串: %s\n", dbConfig.GetDSN())
		fmt.Printf("是否为SQLite: %t\n", dbConfig.IsSQLite())
		
		if err := dbConfig.Validate(); err != nil {
			log.Printf("数据库配置验证失败: %v\n", err)
		}
	}

	// 6. 使用JWT配置
	jwtConfig := GetJWTConfig()
	if jwtConfig != nil {
		fmt.Printf("JWT过期时间: %d小时\n", jwtConfig.ExpireTime)
		fmt.Printf("JWT过期时间间隔: %v\n", jwtConfig.GetExpireDuration())
		fmt.Printf("是否为默认密钥: %t\n", jwtConfig.IsSecretDefault())
		
		if err := jwtConfig.Validate(); err != nil {
			log.Printf("JWT配置验证失败: %v\n", err)
		}
	}

	// 7. 使用Redis配置
	redisConfig := GetRedisConfig()
	if redisConfig != nil {
		fmt.Printf("Redis地址: %s\n", redisConfig.GetAddr())
		fmt.Printf("Redis连接字符串: %s\n", redisConfig.GetConnectionString())
		fmt.Printf("是否设置密码: %t\n", redisConfig.IsPasswordSet())
		
		if err := redisConfig.Validate(); err != nil {
			log.Printf("Redis配置验证失败: %v\n", err)
		}
	}

	// 8. 使用存储配置
	storageConfig := GetStorageConfig()
	if storageConfig != nil {
		fmt.Printf("上传路径: %s\n", storageConfig.UploadPath)
		fmt.Printf("最大文件大小: %dMB\n", storageConfig.MaxFileSize)
		fmt.Printf("允许的文件类型: %s\n", storageConfig.GetAllowedTypesString())
		fmt.Printf("公共文件路径: %s\n", storageConfig.GetPublicFilePath("example.jpg"))
		
		// 检查文件类型是否允许
		fmt.Printf("jpg文件是否允许: %t\n", storageConfig.IsFileTypeAllowed("jpg"))
		fmt.Printf("exe文件是否允许: %t\n", storageConfig.IsFileTypeAllowed("exe"))
		
		if err := storageConfig.Validate(); err != nil {
			log.Printf("存储配置验证失败: %v\n", err)
		}
	}
}

// ExampleServerConfig 服务器配置使用示例
func ExampleServerConfig() {
	serverConfig := GetServerConfig()
	if serverConfig == nil {
		log.Println("服务器配置未加载")
		return
	}

	// 检查运行模式
	if serverConfig.IsDebugMode() {
		log.Println("当前运行在调试模式")
	} else if serverConfig.IsProductionMode() {
		log.Println("当前运行在生产模式")
	} else {
		log.Println("当前运行在未知模式:", serverConfig.Mode)
	}

	// 获取完整URL
	apiURL := serverConfig.GetFullURL("/api/v1/users")
	log.Println("API完整URL:", apiURL)
}

// ExampleDatabaseConfig 数据库配置使用示例
func ExampleDatabaseConfig() {
	dbConfig := GetDatabaseConfig()
	if dbConfig == nil {
		log.Println("数据库配置未加载")
		return
	}

	// 根据数据库类型执行不同操作
	switch {
	case dbConfig.IsSQLite():
		log.Println("使用SQLite数据库:", dbConfig.Database)
	case dbConfig.IsMySQL():
		log.Println("使用MySQL数据库:", dbConfig.Database)
	case dbConfig.IsPostgreSQL():
		log.Println("使用PostgreSQL数据库:", dbConfig.Database)
	default:
		log.Println("未知数据库类型:", dbConfig.Driver)
	}

	// 获取连接字符串
	dsn := dbConfig.GetDSN()
	log.Println("数据库连接字符串:", dsn)
}

// ExampleStorageConfig 存储配置使用示例
func ExampleStorageConfig() {
	storageConfig := GetStorageConfig()
	if storageConfig == nil {
		log.Println("存储配置未加载")
		return
	}

	// 文件类型验证
	testFiles := []string{"document.pdf", "image.jpg", "script.exe", "data.txt"}
	for _, file := range testFiles {
		ext := file[len(file)-3:] // 简单获取扩展名
		if storageConfig.IsFileTypeAllowed(ext) {
			log.Printf("文件 %s 类型允许上传\n", file)
		} else {
			log.Printf("文件 %s 类型不允许上传\n", file)
		}
	}

	// 文件大小验证（假设文件大小为5MB）
	fileSizeMB := 5
	if fileSizeMB <= storageConfig.MaxFileSize {
		log.Printf("文件大小 %dMB 在允许范围内\n", fileSizeMB)
	} else {
		log.Printf("文件大小 %dMB 超出限制 %dMB\n", fileSizeMB, storageConfig.MaxFileSize)
	}
}
