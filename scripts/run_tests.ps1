# 云平台API测试运行脚本 (PowerShell版本)
# 功能：运行所有测试，包括单元测试、集成测试、性能测试

param(
    [Parameter(Position=0)]
    [ValidateSet("unit", "integration", "performance", "security", "all", "clean")]
    [string]$TestType = "all"
)

# 颜色定义
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# 检查Go环境
function Test-GoEnvironment {
    Write-Info "检查Go环境..."
    
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go未安装或未在PATH中"
        exit 1
    }
    
    $goVersion = (go version).Split(' ')[2]
    Write-Success "Go版本: $goVersion"
}

# 安装测试依赖
function Install-Dependencies {
    Write-Info "安装测试依赖..."
    
    # 安装测试框架
    go mod tidy
    
    # 安装测试工具
    if (-not (Get-Command go-junit-report -ErrorAction SilentlyContinue)) {
        Write-Info "安装go-junit-report..."
        go install github.com/jstemmer/go-junit-report@latest
    }
    
    if (-not (Get-Command gocov -ErrorAction SilentlyContinue)) {
        Write-Info "安装gocov..."
        go install github.com/axw/gocov/gocov@latest
    }
    
    if (-not (Get-Command gocov-xml -ErrorAction SilentlyContinue)) {
        Write-Info "安装gocov-xml..."
        go install github.com/AlekSi/gocov-xml@latest
    }
    
    Write-Success "依赖安装完成"
}

# 运行单元测试
function Invoke-UnitTests {
    Write-Info "运行单元测试..."
    
    # 创建测试结果目录
    if (-not (Test-Path "test-results")) {
        New-Item -ItemType Directory -Path "test-results" | Out-Null
    }
    
    # 运行单元测试
    go test -v -race -coverprofile=test-results/coverage.out -covermode=atomic ./... 2>&1 | Tee-Object -FilePath "test-results/unit-test.log"
    
    # 生成测试报告
    go test -v ./... 2>&1 | go-junit-report > test-results/unit-test.xml
    
    # 生成覆盖率报告
    gocov convert test-results/coverage.out | gocov-xml > test-results/coverage.xml
    
    Write-Success "单元测试完成"
}

# 运行集成测试
function Invoke-IntegrationTests {
    Write-Info "运行集成测试..."
    
    # 设置测试环境变量
    $env:TEST_ENV = "integration"
    $env:TEST_DB_URL = "postgres://test:test@localhost:5432/test_db?sslmode=disable"
    $env:TEST_REDIS_URL = "redis://localhost:6379/1"
    
    # 运行集成测试
    go test -v -tags=integration ./tests/Integration/... 2>&1 | Tee-Object -FilePath "test-results/integration-test.log"
    
    Write-Success "集成测试完成"
}

# 运行性能测试
function Invoke-PerformanceTests {
    Write-Info "运行性能测试..."
    
    # 运行基准测试
    go test -bench=. -benchmem ./tests/benchmark/... 2>&1 | Tee-Object -FilePath "test-results/benchmark.log"
    
    # 运行负载测试
    go test -v -tags=load ./tests/benchmark/... 2>&1 | Tee-Object -FilePath "test-results/load-test.log"
    
    Write-Success "性能测试完成"
}

# 运行安全测试
function Invoke-SecurityTests {
    Write-Info "运行安全测试..."
    
    # 运行安全测试
    go test -v -tags=security ./tests/... 2>&1 | Tee-Object -FilePath "test-results/security-test.log"
    
    Write-Success "安全测试完成"
}

# 生成测试报告
function New-TestReport {
    Write-Info "生成测试报告..."
    
    # 创建HTML报告
    $htmlContent = @"
<!DOCTYPE html>
<html>
<head>
    <title>云平台API测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
        pre { background-color: #f5f5f5; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>云平台API测试报告</h1>
        <p>生成时间: $(Get-Date)</p>
    </div>
    
    <div class="section">
        <h2>测试概览</h2>
        <p>总测试数: $((Get-Content "test-results/unit-test.log" | Select-String "PASS|FAIL").Count)</p>
        <p>通过数: $((Get-Content "test-results/unit-test.log" | Select-String "PASS").Count)</p>
        <p>失败数: $((Get-Content "test-results/unit-test.log" | Select-String "FAIL").Count)</p>
    </div>
    
    <div class="section">
        <h2>单元测试结果</h2>
        <pre>$(Get-Content "test-results/unit-test.log" -Raw)</pre>
    </div>
    
    <div class="section">
        <h2>集成测试结果</h2>
        <pre>$(Get-Content "test-results/integration-test.log" -Raw)</pre>
    </div>
    
    <div class="section">
        <h2>性能测试结果</h2>
        <pre>$(Get-Content "test-results/benchmark.log" -Raw)</pre>
    </div>
    
    <div class="section">
        <h2>安全测试结果</h2>
        <pre>$(Get-Content "test-results/security-test.log" -Raw)</pre>
    </div>
</body>
</html>
"@
    
    $htmlContent | Out-File -FilePath "test-results/index.html" -Encoding UTF8
    
    Write-Success "测试报告已生成: test-results/index.html"
}

# 清理测试环境
function Clear-TestEnvironment {
    Write-Info "清理测试环境..."
    
    # 清理临时文件
    Remove-Item "test-results/*.log" -ErrorAction SilentlyContinue
    Remove-Item "test-results/*.xml" -ErrorAction SilentlyContinue
    Remove-Item "test-results/coverage.out" -ErrorAction SilentlyContinue
    
    Write-Success "清理完成"
}

# 主函数
function Main {
    Write-Info "开始运行云平台API测试套件..."
    
    switch ($TestType) {
        "unit" {
            Test-GoEnvironment
            Install-Dependencies
            Invoke-UnitTests
            New-TestReport
        }
        "integration" {
            Test-GoEnvironment
            Install-Dependencies
            Invoke-IntegrationTests
            New-TestReport
        }
        "performance" {
            Test-GoEnvironment
            Install-Dependencies
            Invoke-PerformanceTests
            New-TestReport
        }
        "security" {
            Test-GoEnvironment
            Install-Dependencies
            Invoke-SecurityTests
            New-TestReport
        }
        "all" {
            Test-GoEnvironment
            Install-Dependencies
            Invoke-UnitTests
            Invoke-IntegrationTests
            Invoke-PerformanceTests
            Invoke-SecurityTests
            New-TestReport
        }
        "clean" {
            Clear-TestEnvironment
        }
        default {
            Write-Host "用法: .\run_tests.ps1 [unit|integration|performance|security|all|clean]"
            Write-Host "  unit        - 运行单元测试"
            Write-Host "  integration - 运行集成测试"
            Write-Host "  performance - 运行性能测试"
            Write-Host "  security    - 运行安全测试"
            Write-Host "  all         - 运行所有测试（默认）"
            Write-Host "  clean       - 清理测试环境"
            exit 1
        }
    }
    
    Write-Success "测试套件运行完成！"
}

# 运行主函数
Main
