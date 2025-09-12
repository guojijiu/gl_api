# Cloud Platform API 脚本管理

本目录包含了用于管理 Cloud Platform API 的各种脚本，支持环境配置、日志管理、部署自动化等功能。

## 📋 脚本概览

| 脚本类型 | 文件名 | 功能 | 支持平台 | 优先级 |
|----------|--------|------|----------|--------|
| 代码质量 | `code_quality.sh` | 代码质量检查 | Linux/Mac | ⭐⭐⭐⭐⭐ |
| 代码质量 | `code_quality.ps1` | 代码质量检查 | Windows | ⭐⭐⭐⭐⭐ |
| 部署脚本 | `deploy.sh` | 自动部署 | Linux/Mac | ⭐⭐⭐⭐ |
| 部署脚本 | `deploy.bat` | 自动部署 | Windows | ⭐⭐⭐⭐ |
| 测试脚本 | `run_tests.sh` | 测试执行 | Linux/Mac | ⭐⭐⭐⭐ |
| 性能测试 | `run_benchmarks.sh` | 性能基准测试 | Linux/Mac | ⭐⭐⭐ |
| 文档维护 | `docs_maintenance.sh` | 文档检查 | Linux/Mac | ⭐⭐⭐ |
| 数据库 | `init-db.sql` | 数据库初始化 | 通用 | ⭐⭐⭐⭐⭐ |
| 数据库 | `migrate.go` | 数据库迁移 | 通用 | ⭐⭐⭐⭐⭐ |
| 工具 | `generate-jwt-secret.go` | JWT密钥生成 | 通用 | ⭐⭐⭐ |

## 🚀 快速开始

### 代码质量检查
```bash
# Linux/Mac
./scripts/code_quality.sh

# Windows
.\scripts\code_quality.ps1
```

### 运行测试
```bash
# 运行所有测试
./scripts/run_tests.sh

# 运行性能测试
./scripts/run_benchmarks.sh
```

### 部署应用
```bash
# Linux/Mac
./scripts/deploy.sh production

# Windows
scripts\deploy.bat production
```

### 数据库操作
```bash
# 生成JWT密钥
go run scripts/generate-jwt-secret.go

# 运行数据库迁移
go run scripts/migrate.go -action migrate
```

## 📁 脚本文件说明

### 代码质量脚本
- **`code_quality.sh`**: Linux/Mac环境的代码质量检查脚本
- **`code_quality.ps1`**: Windows环境的代码质量检查脚本

**功能**:
- 代码格式化检查 (gofmt, goimports)
- 代码质量检查 (golangci-lint)
- 安全检查 (gosec)
- 依赖检查 (go mod tidy)
- 测试覆盖率检查

### 部署脚本
- **`deploy.sh`**: Linux/Mac环境的自动部署脚本
- **`deploy.bat`**: Windows环境的自动部署脚本

**功能**:
- 支持多种部署模式 (local, docker, production)
- 自动构建和部署应用
- 环境特定配置管理
- 部署前健康检查

### 测试脚本
- **`run_tests.sh`**: 测试执行脚本
- **`run_benchmarks.sh`**: 性能基准测试脚本

**功能**:
- 支持多种测试类型 (unit, integration, benchmark)
- 自动生成测试报告
- 测试覆盖率统计
- 性能基准测试

### 数据库脚本
- **`init-db.sql`**: 数据库初始化脚本
- **`migrate.go`**: 数据库迁移工具
- **`optimize_database.sql`**: 数据库优化脚本

**功能**:
- 创建数据库表结构
- 插入初始数据
- 创建索引和触发器
- 数据库迁移管理

### 工具脚本
- **`generate-jwt-secret.go`**: JWT密钥生成工具
- **`docs_maintenance.sh`**: 文档维护脚本

**功能**:
- 生成安全的JWT密钥
- 检查文档完整性
- 生成文档统计

## 🔧 故障排除

### 常见问题

1. **权限问题**
   ```bash
   # 给脚本添加执行权限
   chmod +x scripts/*.sh
   ```

2. **依赖缺失**
   ```bash
   # 安装必要的工具
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
   ```

3. **PowerShell执行策略**
   ```powershell
   # 临时允许执行脚本
   Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
   ```

## 📝 注意事项

1. **首次使用**: 确保已安装Go环境和必要的依赖
2. **环境配置**: 根据实际环境修改配置参数
3. **权限管理**: 确保有足够的文件读写权限
4. **备份数据**: 重要操作前请备份数据

## 🆘 获取帮助

- 查看脚本帮助: 使用 `-h` 或 `--help` 参数
- 查看详细输出: 使用 `-v` 或 `--verbose` 参数
- 检查日志文件: 查看生成的日志和报告文件