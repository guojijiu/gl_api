# 简单的日志目录设置测试脚本
# 用于快速验证基本功能

Write-Host "=== 日志目录设置测试 ===" -ForegroundColor Green
Write-Host ""

# 检查当前目录
$currentDir = Get-Location
Write-Host "当前目录: $currentDir" -ForegroundColor Yellow

# 检查项目文件
if (-not (Test-Path "main.go")) {
    Write-Host "❌ 未找到main.go文件，请确保在项目根目录运行" -ForegroundColor Red
    Write-Host "请切换到项目根目录后重试" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ 项目文件检查通过" -ForegroundColor Green

# 检查PowerShell版本
$psVersion = $PSVersionTable.PSVersion
Write-Host "PowerShell版本: $psVersion" -ForegroundColor Yellow

if ($psVersion.Major -lt 5) {
    Write-Host "⚠️  PowerShell版本较低，建议使用PowerShell 5.0或更高版本" -ForegroundColor Yellow
}

# 检查执行策略
$executionPolicy = Get-ExecutionPolicy
Write-Host "执行策略: $executionPolicy" -ForegroundColor Yellow

if ($executionPolicy -eq "Restricted") {
    Write-Host "⚠️  执行策略受限，正在临时设置为Bypass..." -ForegroundColor Yellow
    Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force
    Write-Host "✅ 执行策略已临时设置为Bypass" -ForegroundColor Green
}

# 创建简单的日志目录结构
$logBasePath = "./storage/logs"
Write-Host ""
Write-Host "正在创建基础日志目录..." -ForegroundColor Cyan

# 创建基础目录
if (-not (Test-Path $logBasePath)) {
    New-Item -ItemType Directory -Path $logBasePath -Force | Out-Null
    Write-Host "✅ 基础日志目录创建成功: $logBasePath" -ForegroundColor Green
} else {
    Write-Host "✅ 基础日志目录已存在: $logBasePath" -ForegroundColor Green
}

# 创建基本日志目录
$basicDirs = @("requests", "sql", "errors", "audit", "security", "business", "access", "system")

foreach ($dir in $basicDirs) {
    $fullPath = Join-Path $logBasePath $dir
    if (-not (Test-Path $fullPath)) {
        New-Item -ItemType Directory -Path $fullPath -Force | Out-Null
        Write-Host "✅ 创建目录: $dir" -ForegroundColor Green
    } else {
        Write-Host "✅ 目录已存在: $dir" -ForegroundColor Gray
    }
}

# 创建示例日志文件
$sampleLogFile = Join-Path $logBasePath "system" "test.log"
$sampleContent = @"
[$(Get-Date -Format "yyyy-MM-dd HH:mm:ss")] [INFO] 测试日志文件创建成功
[$(Get-Date -Format "yyyy-MM-dd HH:mm:ss")] [INFO] 环境: development
[$(Get-Date -Format "yyyy-MM-dd HH:mm:ss")] [INFO] 基础路径: $logBasePath
"@

$sampleContent | Set-Content $sampleLogFile -Encoding UTF8
Write-Host "✅ 示例日志文件创建成功: $sampleLogFile" -ForegroundColor Green

# 显示目录结构
Write-Host ""
Write-Host "=== 创建的目录结构 ===" -ForegroundColor Cyan
Get-ChildItem $logBasePath -Directory | Sort-Object Name | ForEach-Object {
    $dirName = $_.Name
    $fileCount = (Get-ChildItem $_.FullName -File | Measure-Object).Count
    Write-Host "📁 $dirName ($fileCount 个文件)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Green
Write-Host "✅ 基础日志目录结构已创建" -ForegroundColor Green
Write-Host "✅ 现在可以运行 'go run main.go' 启动应用" -ForegroundColor Yellow
Write-Host ""

# 提供下一步建议
Write-Host "下一步操作建议:" -ForegroundColor Cyan
Write-Host "1. 运行完整设置: .\scripts\setup_logging_directories.ps1 -Environment development" -ForegroundColor Yellow
Write-Host "2. 或者直接启动应用: go run main.go" -ForegroundColor Yellow
Write-Host "3. 查看详细帮助: .\scripts\setup_logging_directories.ps1 -Verbose" -ForegroundColor Yellow
