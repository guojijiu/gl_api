# 代码质量检查脚本 (Windows PowerShell版本)
# 用于自动化代码审查和质量检查

param(
    [switch]$Install,
    [switch]$Format,
    [switch]$Quality,
    [switch]$Security,
    [switch]$Coverage,
    [switch]$All
)

# 颜色定义
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"
$Cyan = "Cyan"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-ToolInstalled {
    param([string]$ToolName)
    
    try {
        $null = Get-Command $ToolName -ErrorAction Stop
        return $true
    }
    catch {
        return $false
    }
}

function Install-Tools {
    Write-ColorOutput "📦 检查并安装必要的工具..." $Blue
    
    # 安装 golangci-lint
    if (-not (Test-ToolInstalled "golangci-lint")) {
        Write-ColorOutput "安装 golangci-lint..." $Yellow
        $env:PATH += ";$env:GOPATH\bin"
        Invoke-WebRequest -Uri "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" -OutFile "install-golangci-lint.sh"
        bash install-golangci-lint.sh -b "$env:GOPATH\bin" v1.54.2
        Remove-Item "install-golangci-lint.sh" -Force
    }
    
    # 安装 goimports
    if (-not (Test-ToolInstalled "goimports")) {
        Write-ColorOutput "安装 goimports..." $Yellow
        go install golang.org/x/tools/cmd/goimports@latest
    }
    
    # 安装 gosec
    if (-not (Test-ToolInstalled "gosec")) {
        Write-ColorOutput "安装 gosec..." $Yellow
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    }
}

function Test-Formatting {
    Write-ColorOutput "🎨 检查代码格式化..." $Blue
    
    # 检查 gofmt
    $gofmtResult = gofmt -l .
    if ($gofmtResult) {
        Write-ColorOutput "❌ gofmt 检查失败，发现格式问题：" $Red
        Write-Host $gofmtResult
        Write-ColorOutput "💡 运行 'gofmt -w .' 自动修复格式问题" $Yellow
        return $false
    } else {
        Write-ColorOutput "✅ gofmt 检查通过" $Green
    }
    
    # 检查 goimports
    $goimportsResult = goimports -l .
    if ($goimportsResult) {
        Write-ColorOutput "❌ goimports 检查失败，发现导入问题：" $Red
        Write-Host $goimportsResult
        Write-ColorOutput "💡 运行 'goimports -w .' 自动修复导入问题" $Yellow
        return $false
    } else {
        Write-ColorOutput "✅ goimports 检查通过" $Green
    }
    
    return $true
}

function Test-Quality {
    Write-ColorOutput "🔍 运行代码质量检查..." $Blue
    
    # 运行 golangci-lint
    $env:PATH += ";$env:GOPATH\bin"
    $golangciResult = golangci-lint run --config .golangci.yml
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✅ golangci-lint 检查通过" $Green
        return $true
    } else {
        Write-ColorOutput "❌ golangci-lint 检查失败" $Red
        return $false
    }
}

function Test-Security {
    Write-ColorOutput "🔒 运行安全检查..." $Blue
    
    # 运行 gosec
    $env:PATH += ";$env:GOPATH\bin"
    $gosecResult = gosec -fmt json -out gosec-report.json ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✅ gosec 安全检查通过" $Green
    } else {
        Write-ColorOutput "⚠️  gosec 发现安全问题，请查看报告" $Yellow
        if (Test-Path "gosec-report.json") {
            Write-ColorOutput "📊 安全检查报告已生成: gosec-report.json" $Blue
        }
    }
}

function Test-Dependencies {
    Write-ColorOutput "📦 检查依赖..." $Blue
    
    # 检查是否有未使用的依赖
    $modTidyResult = go mod tidy -v
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✅ 依赖检查通过" $Green
    } else {
        Write-ColorOutput "❌ 依赖检查失败" $Red
        return $false
    }
    
    # 检查是否有安全漏洞
    if (Test-ToolInstalled "govulncheck") {
        Write-ColorOutput "🔍 检查安全漏洞..." $Blue
        $vulnResult = govulncheck ./...
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✅ 安全漏洞检查通过" $Green
        } else {
            Write-ColorOutput "⚠️  发现安全漏洞，请及时修复" $Yellow
        }
    }
    
    return $true
}

function Test-Coverage {
    Write-ColorOutput "🧪 检查测试覆盖率..." $Blue
    
    # 运行测试并生成覆盖率报告
    $testResult = go test -coverprofile=coverage.out -covermode=atomic ./...
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✅ 测试通过" $Green
        
        # 生成覆盖率报告
        go tool cover -html=coverage.out -o coverage.html
        Write-ColorOutput "📊 覆盖率报告已生成: coverage.html" $Blue
        
        # 显示覆盖率统计
        $coverageOutput = go tool cover -func=coverage.out | Select-String "total"
        $coverage = ($coverageOutput -split '\s+')[2]
        Write-ColorOutput "📈 总覆盖率: $coverage" $Blue
        
        # 检查覆盖率是否达到要求
        $coverageNum = [double]($coverage -replace '%', '')
        if ($coverageNum -ge 70) {
            Write-ColorOutput "✅ 覆盖率达标 (≥70%)" $Green
        } else {
            Write-ColorOutput "⚠️  覆盖率未达标 (<70%)" $Yellow
        }
        
        return $true
    } else {
        Write-ColorOutput "❌ 测试失败" $Red
        return $false
    }
}

function New-QualityReport {
    Write-ColorOutput "📊 生成质量报告..." $Blue
    
    # 创建报告目录
    if (-not (Test-Path "reports")) {
        New-Item -ItemType Directory -Path "reports" -Force
    }
    
    # 移动报告文件
    $reportFiles = @("gosec-report.json", "gocloc-report.json", "coverage.out", "coverage.html")
    foreach ($file in $reportFiles) {
        if (Test-Path $file) {
            Move-Item $file "reports\" -Force
        }
    }
    
    Write-ColorOutput "✅ 质量报告已生成到 reports/ 目录" $Green
}

function Main {
    Write-ColorOutput "🚀 开始代码质量检查流程" $Green
    Write-ColorOutput "==================================" $Cyan
    
    # 安装工具
    if ($Install -or $All) {
        Install-Tools
    }
    
    # 检查工具
    if (-not (Test-ToolInstalled "golangci-lint")) {
        Write-ColorOutput "❌ golangci-lint 未安装，请先运行 -Install" $Red
        exit 1
    }
    
    $failed = $false
    
    # 执行检查
    if ($Format -or $All) {
        if (-not (Test-Formatting)) {
            $failed = $true
        }
    }
    
    if ($Quality -or $All) {
        if (-not (Test-Quality)) {
            $failed = $true
        }
    }
    
    if ($Security -or $All) {
        Test-Security
    }
    
    if ($All) {
        Test-Dependencies
        if (-not (Test-Coverage)) {
            $failed = $true
        }
    }
    
    # 生成报告
    if ($All) {
        New-QualityReport
    }
    
    Write-ColorOutput "==================================" $Cyan
    if (-not $failed) {
        Write-ColorOutput "🎉 所有检查通过！代码质量良好" $Green
    } else {
        Write-ColorOutput "❌ 部分检查失败，请修复问题后重新运行" $Red
        exit 1
    }
}

# 运行主函数
Main
