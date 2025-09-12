# 故障排除指南

## 🚨 常见错误及解决方案

### 1. PowerShell执行策略错误

**错误信息**: `无法加载文件，因为在此系统上禁止运行脚本`

**解决方案**:
```powershell
# 临时允许执行（推荐）
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# 永久允许执行（需要管理员权限）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 2. 权限不足错误

**错误信息**: `拒绝访问` 或 `权限不足`

**解决方案**:
- 以管理员身份运行PowerShell
- 检查目录写入权限
- 确保在项目根目录运行脚本

### 3. 路径错误

**错误信息**: `找不到路径` 或 `系统找不到指定的路径`

**解决方案**:
```powershell
# 检查当前目录
Get-Location

# 确保在项目根目录（包含main.go的目录）
cd "D:\project\local\cloud_platform\api\back"

# 检查项目文件
Test-Path "main.go"
```

### 4. 依赖缺失

**错误信息**: `command not found` 或 `未安装`

**解决方案**:
```bash
# 安装Go工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# 安装golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

## 🔧 分步诊断

### 步骤1: 环境检查
```bash
# 检查Go版本
go version

# 检查当前目录
pwd

# 检查项目文件
ls -la main.go go.mod
```

### 步骤2: 权限检查
```bash
# 检查脚本权限
ls -la scripts/*.sh

# 添加执行权限
chmod +x scripts/*.sh
```

### 步骤3: 脚本测试
```bash
# 运行简单测试
./scripts/code_quality.sh --help

# 如果成功，再运行完整脚本
./scripts/code_quality.sh
```

## 📋 快速修复命令

### 一键修复（复制粘贴运行）
```bash
# 添加执行权限
chmod +x scripts/*.sh

# 切换到项目根目录
cd /path/to/your/project

# 运行测试脚本
./scripts/code_quality.sh
```

### 手动创建必要目录
```bash
# 如果脚本仍然失败，可以手动创建
mkdir -p storage/logs
mkdir -p storage/logs/{requests,sql,errors,audit,security,business,access,system}
```

## 🆘 获取帮助

### 1. 查看详细错误信息
```bash
# 添加详细输出参数
./scripts/code_quality.sh -v

# 或者查看错误详情
echo $?
```

### 2. 常见问题检查清单
- [ ] Go环境已安装并配置PATH
- [ ] 在项目根目录运行
- [ ] 有足够的文件读写权限
- [ ] 项目文件完整（main.go, go.mod等）
- [ ] 必要的工具已安装

## 🔄 替代方案

### 如果脚本仍然失败

#### 方案1: 使用Go代码创建
```go
// 在main.go中添加
package main

import (
    "os"
    "path/filepath"
)

func createLogDirectories() {
    basePath := "./storage/logs"
    dirs := []string{"requests", "sql", "errors", "audit", "security", "business", "access", "system"}
    
    for _, dir := range dirs {
        fullPath := filepath.Join(basePath, dir)
        os.MkdirAll(fullPath, 0755)
    }
}

func main() {
    createLogDirectories()
    // ... 其他代码
}
```

#### 方案2: 手动创建目录结构
```bash
mkdir -p storage/logs
mkdir -p storage/logs/requests
mkdir -p storage/logs/sql
mkdir -p storage/logs/errors
mkdir -p storage/logs/audit
mkdir -p storage/logs/security
mkdir -p storage/logs/business
mkdir -p storage/logs/access
mkdir -p storage/logs/system
```

## 📞 联系支持

如果以上方法都无法解决问题，请提供以下信息：

1. **错误信息**: 完整的错误文本
2. **系统信息**: 操作系统版本、Go版本
3. **运行环境**: 当前目录、项目路径
4. **执行步骤**: 具体执行了什么命令
5. **错误截图**: 如果有的话