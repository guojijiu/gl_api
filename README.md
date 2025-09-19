# Cloud Platform API

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-org/cloud-platform-api)
[![Coverage](https://img.shields.io/badge/Coverage-85%25-brightgreen.svg)](https://github.com/your-org/cloud-platform-api)

一个基于 Go 和 Gin 框架构建的高性能云平台后端 API 服务，提供用户管理、认证、监控等核心功能。

## ✨ 特性

- 🚀 **高性能**: 基于 Go 和 Gin 框架，支持高并发处理
- 🔐 **安全可靠**: JWT 认证、密码加密、SQL 注入防护、XSS 防护
- 📊 **监控完善**: 集成 Prometheus 和 Grafana 监控系统
- 🗄️ **数据存储**: 支持 MySQL 数据库和 Redis 缓存
- 📝 **日志系统**: 结构化日志记录，支持多种日志级别
- 🐳 **容器化**: 支持 Docker 和 Kubernetes 部署
- 🧪 **测试覆盖**: 完整的单元测试和集成测试
- 📚 **文档完善**: 详细的 API 文档和开发指南

## 🏗️ 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **缓存**: Redis 6.0+
- **监控**: Prometheus + Grafana
- **容器**: Docker + Docker Compose
- **测试**: Go Testing + Testify

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+ (可选)

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/your-org/cloud-platform-api.git
cd cloud-platform-api
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境**
```bash
cp env.example .env
# 编辑 .env 文件，配置数据库等信息
```

4. **启动数据库**
```bash
# 使用 Docker Compose 启动 MySQL 和 Redis
docker-compose up -d mysql redis
```

5. **运行应用**
```bash
go run main.go
```

6. **验证安装**
```bash
curl http://localhost:8080/api/v1/health
```

## 📖 文档

- [API 文档](docs/API_Documentation.md) - 详细的 API 接口文档
- [部署指南](docs/Deployment_Guide.md) - 生产环境部署指南
- [开发指南](docs/Development_Guide.md) - 开发者指南和最佳实践

## 🏃‍♂️ 快速使用

### 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePass123!"
  }'
```

### 获取用户信息
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ 项目结构

```
cloud-platform-api/
├── app/                    # 应用核心代码
│   ├── Config/            # 配置管理
│   ├── Database/          # 数据库相关
│   ├── Http/              # HTTP 层
│   │   ├── Controllers/   # 控制器
│   │   ├── Middleware/    # 中间件
│   │   └── Routes/        # 路由
│   ├── Models/            # 数据模型
│   ├── Services/          # 业务逻辑层
│   └── Utils/             # 工具函数
├── docs/                  # 文档
├── monitoring/            # 监控配置
├── tests/                 # 测试文件
├── storage/               # 存储目录
└── main.go               # 应用入口
```

## 🔧 配置

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `DB_HOST` | 数据库主机 | localhost |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_USERNAME` | 数据库用户名 | root |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_DATABASE` | 数据库名称 | cloud_platform |
| `REDIS_HOST` | Redis 主机 | localhost |
| `REDIS_PORT` | Redis 端口 | 6379 |
| `JWT_SECRET` | JWT 密钥 | - |
| `SERVER_PORT` | 服务器端口 | 8080 |

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 使用 Docker

```bash
# 构建镜像
docker build -t cloud-platform-api .

# 运行容器
docker run -d \
  --name cloud-platform-api \
  -p 8080:8080 \
  -e DB_HOST=mysql \
  -e DB_USERNAME=root \
  -e DB_PASSWORD=password \
  -e DB_DATABASE=cloud_platform \
  cloud-platform-api
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./app/Services/...

# 运行测试并显示覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 测试覆盖率

项目目标测试覆盖率达到 85% 以上，包括：
- 单元测试：业务逻辑测试
- 集成测试：API 接口测试
- 安全测试：安全漏洞测试

## 📊 监控

### 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

### 系统监控

```bash
curl http://localhost:8080/api/v1/monitor/system
```

### Prometheus 指标

访问 `http://localhost:9090` 查看 Prometheus 指标

### Grafana 仪表板

访问 `http://localhost:3000` 查看 Grafana 仪表板

## 🔒 安全特性

- **密码安全**: 使用 bcrypt 加密存储
- **JWT 认证**: 安全的 token 生成和验证
- **输入验证**: 严格的输入参数验证
- **SQL 注入防护**: 参数化查询
- **XSS 防护**: 输入过滤和输出编码
- **CSRF 防护**: CSRF token 验证
- **速率限制**: API 调用频率限制

## 🚀 性能优化

- **数据库优化**: 连接池配置、查询优化、索引优化
- **缓存策略**: Redis 缓存、内存缓存
- **并发处理**: Goroutine 池、并发安全
- **响应压缩**: Gzip 压缩
- **静态资源优化**: CDN 支持

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 提交规范

- `feat`: 新功能
- `fix`: 修复问题
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

## 📝 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础用户管理功能
- JWT 认证系统
- 监控和日志系统
- Docker 支持

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 联系方式

- **项目地址**: https://github.com/your-org/cloud-platform-api
- **问题反馈**: https://github.com/your-org/cloud-platform-api/issues
- **邮箱**: support@yourdomain.com

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

⭐ 如果这个项目对你有帮助，请给它一个星标！