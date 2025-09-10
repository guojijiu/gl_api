# Cloud Platform API 脚本管理

本目录包含了用于管理 Cloud Platform API 的各种脚本，支持环境配置、日志管理、部署自动化等功能。

## 📋 目录

- [脚本概览](#-脚本概览)
- [环境配置脚本](#-环境配置脚本)
- [日志管理脚本](#-日志管理脚本)
- [部署脚本](#-部署脚本)
- [测试脚本](#-测试脚本)
- [维护脚本](#-维护脚本)
- [使用示例](#-使用示例)
- [故障排除](#-故障排除)

## 📊 脚本概览

| 脚本类型 | 文件名 | 功能 | 支持平台 | 优先级 |
|----------|--------|------|----------|--------|
| 环境配置 | `setup_environment.ps1` | 环境配置生成 | Windows | ⭐⭐⭐⭐⭐ |
| 日志管理 | `setup_logging_directories.ps1` | 日志目录创建 | Windows | ⭐⭐⭐⭐⭐ |
| 快速启动 | `quick_start.ps1` | 一键启动 | Windows | ⭐⭐⭐⭐⭐ |
| 部署脚本 | `deploy.sh` | 自动部署 | Linux/Mac | ⭐⭐⭐⭐ |
| 测试脚本 | `run_tests.sh` | 测试执行 | Linux/Mac | ⭐⭐⭐⭐ |
| 维护脚本 | `maintenance.ps1` | 系统维护 | Windows | ⭐⭐⭐ |

## 📁 脚本文件说明

### 环境配置脚本

#### 1. `setup_environment.ps1` - 环境配置脚本
**功能**: 根据环境自动生成完整的环境配置文件
**支持环境**: development, testing, staging, production

**特性**:
- 自动生成环境特定的.env文件
- 根据环境调整所有配置参数
- 生成配置摘要和启动脚本
- 支持强制重新生成和详细输出

**使用方法**:
```powershell
# 基本使用（开发环境）
.\scripts\setup_environment.ps1 -Environment development

# 指定配置路径
.\scripts\setup_environment.ps1 -Environment testing -ConfigPath "./config"

# 强制重新生成
.\scripts\setup_environment.ps1 -Environment staging -Force

# 详细输出
.\scripts\setup_environment.ps1 -Environment production -Verbose
```

#### 2. `setup_logging_directories.ps1` - 日志目录设置脚本
**功能**: 根据环境自动创建完整的日志目录结构
**支持环境**: development, testing, staging, production

**特性**:
- 自动创建20+种不同类型的日志目录
- 根据环境调整日志配置参数
- 生成配置文件、README和示例日志
- 支持强制重新创建和详细输出

**使用方法**:
```powershell
# 基本使用（开发环境）
.\scripts\setup_logging_directories.ps1

# 指定环境
.\scripts\setup_logging_directories.ps1 -Environment production

# 强制重新创建
.\scripts\setup_logging_directories.ps1 -Environment testing -Force

# 详细输出
.\scripts\setup_logging_directories.ps1 -Environment staging -Verbose

# 自定义日志路径
.\scripts\setup_logging_directories.ps1 -Environment production -LogBasePath "./logs"
```

### 快速启动脚本

#### 3. `quick_start.ps1` - 快速启动脚本
**功能**: 一键完成环境配置、日志目录设置和应用启动
**支持环境**: development, testing, staging, production

**特性**:
- 自动检查Go环境和项目依赖
- 按顺序执行所有设置步骤
- 自动加载环境变量
- 启动应用并提供故障排除建议

**使用方法**:
```powershell
# 基本使用（开发环境，包含设置）
.\scripts\quick_start.ps1

# 指定环境
.\scripts\quick_start.ps1 -Environment production

# 跳过设置步骤（仅启动应用）
.\scripts\quick_start.ps1 -Environment testing -SkipSetup

# 强制重新设置
.\scripts\quick_start.ps1 -Environment staging -Force

# 详细输出
.\scripts\quick_start.ps1 -Environment development -Verbose
```

### 部署脚本

#### 4. `deploy.sh` - 自动部署脚本
**功能**: 自动化部署到不同环境
**支持环境**: development, staging, production

**特性**:
- 支持Docker和Kubernetes部署
- 自动构建和推送镜像
- 环境特定配置管理
- 部署前健康检查

**使用方法**:
```bash
# 部署到开发环境
./scripts/deploy.sh development

# 部署到生产环境
./scripts/deploy.sh production

# 使用Docker部署
./scripts/deploy.sh production --docker

# 使用Kubernetes部署
./scripts/deploy.sh production --k8s

# 强制重新部署
./scripts/deploy.sh production --force
```

#### 5. `deploy.bat` - Windows部署脚本
**功能**: Windows环境下的自动部署
**支持环境**: development, staging, production

**使用方法**:
```cmd
# 部署到开发环境
scripts\deploy.bat development

# 部署到生产环境
scripts\deploy.bat production

# 使用Docker部署
scripts\deploy.bat production --docker
```

### 测试脚本

#### 6. `run_tests.sh` - 测试执行脚本
**功能**: 运行各种类型的测试
**支持测试类型**: unit, integration, e2e, performance

**特性**:
- 支持多种测试类型
- 自动生成测试报告
- 测试覆盖率统计
- 性能基准测试

**使用方法**:
```bash
# 运行所有测试
./scripts/run_tests.sh

# 运行单元测试
./scripts/run_tests.sh unit

# 运行集成测试
./scripts/run_tests.sh integration

# 运行端到端测试
./scripts/run_tests.sh e2e

# 运行性能测试
./scripts/run_tests.sh performance

# 生成测试报告
./scripts/run_tests.sh --report

# 查看测试覆盖率
./scripts/run_tests.sh --coverage
```

#### 7. `test_setup.ps1` - 测试环境设置脚本
**功能**: 设置测试环境
**支持环境**: unit, integration, e2e

**使用方法**:
```powershell
# 设置单元测试环境
.\scripts\test_setup.ps1 -TestType unit

# 设置集成测试环境
.\scripts\test_setup.ps1 -TestType integration

# 设置端到端测试环境
.\scripts\test_setup.ps1 -TestType e2e
```

### 维护脚本

#### 8. `maintenance.ps1` - 系统维护脚本
**功能**: 系统维护和清理
**支持操作**: cleanup, backup, restore, update

**特性**:
- 日志文件清理
- 数据库备份和恢复
- 系统更新
- 性能优化

**使用方法**:
```powershell
# 清理日志文件
.\scripts\maintenance.ps1 -Action cleanup

# 备份数据库
.\scripts\maintenance.ps1 -Action backup

# 恢复数据库
.\scripts\maintenance.ps1 -Action restore

# 更新系统
.\scripts\maintenance.ps1 -Action update

# 性能优化
.\scripts\maintenance.ps1 -Action optimize
```

#### 9. `backup.sh` - 备份脚本
**功能**: 数据备份和恢复
**支持类型**: database, files, config

**使用方法**:
```bash
# 备份数据库
./scripts/backup.sh database

# 备份文件
./scripts/backup.sh files

# 备份配置
./scripts/backup.sh config

# 全量备份
./scripts/backup.sh all

# 恢复备份
./scripts/backup.sh restore backup_file.tar.gz
```

### 批处理脚本

#### 10. `setup_logging_directories.bat` - Windows批处理脚本
**功能**: Windows用户的简化日志目录设置脚本
**支持环境**: development, testing, staging, production

**使用方法**:
```cmd
# 基本使用
scripts\setup_logging_directories.bat

# 指定环境
scripts\setup_logging_directories.bat --env production

# 强制重新创建
scripts\setup_logging_directories.bat --force

# 详细输出
scripts\setup_logging_directories.bat --verbose

# 显示帮助
scripts\setup_logging_directories.bat --help
```

### 3. `quick_start.ps1` - 快速启动脚本
**功能**: 一键完成环境配置、日志目录设置和应用启动
**支持环境**: development, testing, staging, production

**特性**:
- 自动检查Go环境和项目依赖
- 按顺序执行所有设置步骤
- 自动加载环境变量
- 启动应用并提供故障排除建议

**使用方法**:
```powershell
# 基本使用（开发环境，包含设置）
.\scripts\quick_start.ps1

# 指定环境
.\scripts\quick_start.ps1 -Environment production

# 跳过设置步骤（仅启动应用）
.\scripts\quick_start.ps1 -Environment testing -SkipSetup

# 强制重新设置
.\scripts\quick_start.ps1 -Environment staging -Force

# 详细输出
.\scripts\quick_start.ps1 -Environment development -Verbose
```

### 4. `setup_logging_directories.bat` - Windows批处理脚本
**功能**: Windows用户的简化日志目录设置脚本
**支持环境**: development, testing, staging, production

**使用方法**:
```cmd
# 基本使用
scripts\setup_logging_directories.bat

# 指定环境
scripts\setup_logging_directories.bat --env production

# 强制重新创建
scripts\setup_logging_directories.bat --force

# 详细输出
scripts\setup_logging_directories.bat --verbose

# 显示帮助
scripts\setup_logging_directories.bat --help
```

## 🚀 快速开始

### 方法1: 一键启动（推荐）
```powershell
# 在项目根目录运行
.\scripts\quick_start.ps1 -Environment development
```

### 方法2: 分步执行
```powershell
# 步骤1: 设置环境配置
.\scripts\setup_environment.ps1 -Environment development

# 步骤2: 设置日志目录
.\scripts\setup_logging_directories.ps1 -Environment development

# 步骤3: 启动应用
go run main.go
```

### 方法3: 仅设置日志目录
```powershell
# 设置开发环境日志目录
.\scripts\setup_logging_directories.ps1 -Environment development

# 设置生产环境日志目录
.\scripts\setup_logging_directories.ps1 -Environment production
```

## 🌍 环境配置说明

### Development（开发环境）
- **日志级别**: debug
- **日志格式**: text（便于阅读）
- **数据库**: SQLite
- **Redis**: 禁用
- **监控**: 禁用
- **安全级别**: 低
- **文件上传限制**: 50MB
- **会话超时**: 24小时
- **速率限制**: 1000/分钟

### Testing（测试环境）
- **日志级别**: info
- **日志格式**: JSON
- **数据库**: SQLite
- **Redis**: 启用
- **监控**: 启用
- **安全级别**: 中等
- **文件上传限制**: 100MB
- **会话超时**: 8小时
- **速率限制**: 500/分钟

### Staging（预生产环境）
- **日志级别**: info
- **日志格式**: JSON
- **数据库**: MySQL
- **Redis**: 启用
- **监控**: 启用
- **安全级别**: 高
- **文件上传限制**: 200MB
- **会话超时**: 4小时
- **速率限制**: 200/分钟

### Production（生产环境）
- **日志级别**: warning
- **日志格式**: JSON
- **数据库**: MySQL
- **Redis**: 启用
- **监控**: 启用
- **安全级别**: 最高
- **文件上传限制**: 500MB
- **会话超时**: 2小时
- **速率限制**: 100/分钟

## 📊 日志目录结构

脚本会自动创建以下日志目录：

### 基础日志目录
- **requests**: 请求日志
- **sql**: SQL查询日志
- **errors**: 错误日志
- **audit**: 审计日志
- **security**: 安全日志
- **business**: 业务日志
- **access**: 访问日志
- **system**: 系统日志

### 监控日志目录
- **performance**: 性能监控日志
- **monitoring**: 系统监控日志
- **alerts**: 告警日志

### 服务日志目录
- **backup**: 备份服务日志
- **cron**: 定时任务日志
- **api**: API调用日志
- **websocket**: WebSocket连接日志
- **cache**: 缓存操作日志
- **queue**: 队列处理日志
- **email**: 邮件服务日志
- **sms**: 短信服务日志
- **third_party**: 第三方服务日志

### 环境特定目录
- **development**: debug, development, local
- **testing**: test, testing, qa
- **staging**: staging, preprod, uat
- **production**: production, prod, live

## ⚙️ 配置参数

### 日志轮转配置
- **最大文件大小**: 50MB - 500MB（根据环境）
- **最大保留时间**: 7天 - 1年（根据环境）
- **最大备份数量**: 5 - 50（根据环境）
- **是否压缩**: 开发环境不压缩，其他环境压缩

### 日志格式配置
- **开发环境**: 文本格式，便于调试
- **其他环境**: JSON格式，便于解析和分析

## 🔧 故障排除

### 常见问题

1. **PowerShell执行策略错误**
   ```powershell
   # 临时允许执行脚本
   Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
   ```

2. **权限不足**
   - 以管理员身份运行PowerShell
   - 检查目录写入权限

3. **Go环境问题**
   - 确保Go已安装并配置PATH
   - 运行 `go version` 验证

4. **端口占用**
   - 检查8080端口是否被占用
   - 使用 `netstat -ano | findstr :8080` 查看

### 日志检查
```powershell
# 查看日志目录结构
Get-ChildItem "./storage/logs" -Directory

# 查看最新日志
Get-Content "./storage/logs/system/system.log" -Tail 10

# 查看错误日志
Get-Content "./storage/logs/errors/errors.log" -Tail 10
```

## 📝 注意事项

1. **首次使用**: 建议先运行 `quick_start.ps1` 完成完整设置
2. **环境切换**: 切换环境时重新运行对应的设置脚本
3. **生产环境**: 生产环境请修改敏感配置（JWT密钥、数据库密码等）
4. **日志清理**: 定期清理旧日志文件，避免磁盘空间不足
5. **备份配置**: 重要环境请备份配置文件和日志数据

## 🆘 获取帮助

- 查看脚本帮助: 添加 `-Verbose` 参数
- 查看批处理帮助: 使用 `--help` 参数
- 检查日志文件: 查看 `./storage/logs/README.md`
- 环境配置摘要: 查看 `./env/environment_summary.md`

## 🔄 更新和维护

- 定期更新脚本以适应新的日志需求
- 根据实际使用情况调整日志配置参数
- 监控日志文件大小和性能影响
- 收集用户反馈改进脚本功能
