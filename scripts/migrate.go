package main

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Database"
	"cloud-platform-api/app/Database/Migrations"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// 定义命令行参数
	action := flag.String("action", "migrate", "迁移操作: migrate, rollback, reset, status")
	steps := flag.Int("steps", 1, "回滚步数")
	flag.Parse()

	// 加载配置
	Config.LoadConfig()

	// 初始化数据库
	Database.InitDB()

	// 创建迁移管理器
	migrationManager := Migrations.NewMigrationManager(Database.GetDB())

	switch *action {
	case "migrate":
		fmt.Println("开始执行数据库迁移...")
		if err := migrationManager.RunMigrations(); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		fmt.Println("✅ 数据库迁移完成")

	case "rollback":
		fmt.Printf("开始回滚 %d 个批次的迁移...\n", *steps)
		if err := migrationManager.RollbackMigrations(*steps); err != nil {
			log.Fatalf("回滚失败: %v", err)
		}
		fmt.Println("✅ 数据库回滚完成")

	case "reset":
		fmt.Println("开始重置所有迁移...")
		if err := migrationManager.ResetMigrations(); err != nil {
			log.Fatalf("重置失败: %v", err)
		}
		fmt.Println("✅ 数据库重置完成")

	case "status":
		fmt.Println("获取迁移状态...")
		status, err := migrationManager.GetMigrationStatus()
		if err != nil {
			log.Fatalf("获取状态失败: %v", err)
		}
		
		fmt.Printf("迁移统计:\n")
		fmt.Printf("  总迁移数: %d\n", status["total_migrations"])
		fmt.Printf("  已执行: %d\n", status["ran_migrations"])
		fmt.Printf("  待执行: %d\n", status["pending_migrations"])
		fmt.Printf("  最后批次: %v\n", status["last_batch"])
		
		fmt.Printf("\n迁移详情:\n")
		migrations := status["migrations"].([]map[string]interface{})
		for _, migration := range migrations {
			name := migration["name"].(string)
			status := migration["status"].(string)
			batch := migration["batch"]
			
			if status == "ran" {
				fmt.Printf("  ✅ %s (批次: %v)\n", name, batch)
			} else {
				fmt.Printf("  ⏳ %s (待执行)\n", name)
			}
		}

	default:
		fmt.Printf("未知操作: %s\n", *action)
		fmt.Println("支持的操作: migrate, rollback, reset, status")
		os.Exit(1)
	}
}
