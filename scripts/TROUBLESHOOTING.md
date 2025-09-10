# 故障排除指南

## 🚨 常见错误及解决方案

### 1. PowerShell执行策略错误

**错误信息**: `无法加载文件，因为在此系统上禁止运行脚本`

**解决方案**:
```powershell
# 方法1: 临时允许执行（推荐）
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# 方法2: 永久允许执行（需要管理员权限）
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

### 4. PowerShell版本过低

**错误信息**: 脚本语法错误或功能不支持

**解决方案**:
```powershell
# 检查PowerShell版本
$PSVersionTable.PSVersion

# 如果版本低于5.0，建议升级到PowerShell 7
# 下载地址: https://github.com/PowerShell/PowerShell/releases
```

### 5. 编码问题

**错误信息**: 中文显示乱码

**解决方案**:
```powershell
# 设置控制台编码
chcp 65001

# 或者在PowerShell中设置
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

## 🔧 分步诊断

### 步骤1: 环境检查
```powershell
# 检查PowerShell版本
$PSVersionTable.PSVersion

# 检查执行策略
Get-ExecutionPolicy

# 检查当前目录
Get-Location

# 检查项目文件
Test-Path "main.go"
Test-Path "go.mod"
```

### 步骤2: 权限检查
```powershell
# 检查目录权限
Get-Acl ".\storage" | Format-List

# 尝试创建测试目录
New-Item -ItemType Directory -Path ".\test" -Force
Remove-Item ".\test" -Force
```

### 步骤3: 脚本测试
```powershell
# 运行简单测试脚本
.\scripts\test_setup.ps1

# 如果成功，再运行完整脚本
.\scripts\setup_logging_directories.ps1 -Environment development
```

## 📋 快速修复命令

### 一键修复（复制粘贴运行）
```powershell
# 设置执行策略
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

# 切换到项目根目录（根据你的实际路径修改）
cd "D:\project\local\cloud_platform\api\back"

# 运行测试脚本
.\scripts\test_setup.ps1
```

### 手动创建日志目录
```powershell
# 如果脚本仍然失败，可以手动创建
$logPath = ".\storage\logs"
New-Item -ItemType Directory -Path $logPath -Force

# 创建基本目录
@("requests", "sql", "errors", "audit", "security", "business", "access", "system") | ForEach-Object {
    New-Item -ItemType Directory -Path "$logPath\$_" -Force
}
```

## 🆘 获取帮助

### 1. 查看详细错误信息
```powershell
# 添加详细输出参数
.\scripts\setup_logging_directories.ps1 -Environment development -Verbose

# 或者查看PowerShell错误详情
$Error[0] | Format-List -Force
```

### 2. 检查日志文件
```powershell
# 查看Windows事件日志
Get-EventLog -LogName Application -Newest 10 | Where-Object {$_.Source -like "*PowerShell*"}

# 查看PowerShell错误历史
Get-History | Select-Object -Last 10
```

### 3. 常见问题检查清单
- [ ] PowerShell版本 >= 5.0
- [ ] 执行策略允许运行脚本
- [ ] 在项目根目录运行
- [ ] 有足够的目录写入权限
- [ ] 项目文件完整（main.go, go.mod等）

## 🔄 替代方案

### 如果PowerShell脚本仍然失败

#### 方案1: 使用批处理文件
```cmd
# 在cmd中运行
scripts\setup_logging_directories.bat --env development
```

#### 方案2: 手动创建目录结构
```cmd
# 在cmd中运行
mkdir storage\logs
mkdir storage\logs\requests
mkdir storage\logs\sql
mkdir storage\logs\errors
mkdir storage\logs\audit
mkdir storage\logs\security
mkdir storage\logs\business
mkdir storage\logs\access
mkdir storage\logs\system
```

#### 方案3: 使用Go代码创建
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

## 📞 联系支持

如果以上方法都无法解决问题，请提供以下信息：

1. **错误信息**: 完整的错误文本
2. **系统信息**: Windows版本、PowerShell版本
3. **运行环境**: 当前目录、项目路径
4. **执行步骤**: 具体执行了什么命令
5. **错误截图**: 如果有的话

## 🎯 预防措施

1. **定期更新**: 保持PowerShell和系统更新
2. **权限管理**: 使用适当的用户权限运行脚本
3. **路径规范**: 使用相对路径，避免硬编码绝对路径
4. **错误处理**: 在脚本中添加适当的错误处理
5. **备份配置**: 定期备份重要的配置文件
