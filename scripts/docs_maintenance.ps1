# 文档维护脚本
# 用于检查文档链接、格式和完整性

Write-Host "检查文档完整性..." -ForegroundColor Green

# 检查所有Markdown文件
Get-ChildItem -Path . -Name "*.md" -Recurse | ForEach-Object {
    $file = $_
    Write-Host "检查文件: $file" -ForegroundColor Yellow
    
    # 检查文件是否为空
    if ((Get-Item $file).Length -eq 0) {
        Write-Host "警告: $file 是空文件" -ForegroundColor Red
    }
    
    # 检查是否有标题
    $content = Get-Content $file -Raw
    if ($content -notmatch "^#") {
        Write-Host "警告: $file 没有标题" -ForegroundColor Red
    }
    
    # 检查链接格式
    if ($content -match "\[.*\]\(.*\)") {
        Write-Host "包含链接: $file" -ForegroundColor Green
    }
}

Write-Host "生成文档统计..." -ForegroundColor Green

# 统计文档数量
$mdFiles = Get-ChildItem -Path . -Name "*.md" -Recurse
$mdCount = $mdFiles.Count
Write-Host "总文档数量: $mdCount" -ForegroundColor Cyan

# 统计总行数
$totalLines = 0
$mdFiles | ForEach-Object {
    $lines = (Get-Content $_).Count
    $totalLines += $lines
}
Write-Host "总行数: $totalLines" -ForegroundColor Cyan

Write-Host "文档检查完成" -ForegroundColor Green