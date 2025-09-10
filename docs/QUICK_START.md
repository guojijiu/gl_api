# 快速开始指南

## 🚀 5分钟快速体验

### 1. 环境准备

```bash
# 克隆项目
git clone <repository-url>
cd cloud-platform-api

# 安装依赖
go mod download
```

### 2. 配置环境

```bash
# 复制配置文件
cp env.example .env

# 编辑配置（使用默认值即可）
vim .env
```

**最小配置示例：**
```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DB_DRIVER=sqlite
DB_DATABASE=:memory:

# JWT配置
JWT_SECRET=your-secret-key-at-least-32-characters-long

# 日志配置
LOG_LEVEL=info
```

### 3. 启动服务

```bash
# 启动数据库（使用Docker）
docker-compose up -d mysql redis

# 运行迁移
go run scripts/migrate.go

# 启动应用
go run main.go
```

### 4. 验证服务

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 查看API文档
open http://localhost:8080/swagger/index.html
```

## 📚 下一步

- [开发指南](DEVELOPMENT.md) - 深入了解开发
- [API文档](API.md) - 查看所有接口
- [部署指南](DEPLOYMENT.md) - 学习部署方法

## 🔧 常见问题

### Q: 启动失败怎么办？
A: 检查端口是否被占用，查看日志文件获取详细错误信息。

### Q: 数据库连接失败？
A: 确保数据库服务正在运行，检查配置文件中的数据库连接信息。

### Q: 如何查看API文档？
A: 启动服务后访问 `http://localhost:8080/swagger/index.html`。

## 📞 获取帮助

- 查看 [故障排除](../scripts/TROUBLESHOOTING.md)
- 提交 Issue 到项目仓库
- 联系技术支持团队
