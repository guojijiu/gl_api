package main

import (
	"cloud-platform-api/app"
	"log"
)

// @title Cloud Platform API
// @version 1.0
// @description 基于Gin + Laravel设计理念的Web开发框架
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// main 应用程序入口点
// 功能说明：
// 1. 创建并启动Cloud Platform API应用
// 2. 处理应用启动过程中的错误
// 3. 支持优雅关闭和资源清理
// 4. 记录应用生命周期事件
//
// 启动流程：
// 1. 创建应用实例（NewApp）
//   - 加载和验证配置
//   - 初始化数据库连接
//   - 设置Redis缓存
//   - 注册路由和中间件
//   - 执行数据库迁移
//
// 2. 启动HTTP服务器（app.Run）
//   - 监听指定端口
//   - 处理系统信号
//   - 优雅关闭服务器
//   - 清理资源
//
// 错误处理：
// - 配置加载失败时立即退出
// - 数据库连接失败时重试后退出
// - 服务器启动失败时记录错误并退出
// - 优雅关闭失败时强制退出
//
// 信号处理：
// - SIGINT: 中断信号（Ctrl+C）
// - SIGTERM: 终止信号（kill命令）
// - 收到信号时启动优雅关闭流程
//
// 资源管理：
// - 自动创建必要的目录结构
// - 初始化数据库连接池
// - 设置Redis连接
// - 配置日志记录器
// - 清理临时文件和缓存
func main() {
	// 创建并启动应用
	app := app.NewApp()

	if err := app.Run(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
