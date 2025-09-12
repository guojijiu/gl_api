#!/bin/bash

# 代码质量检查脚本
# 用于自动化代码审查和质量检查

set -e

echo "🔍 开始代码质量检查..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查工具是否安装
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}❌ $1 未安装，请先安装 $1${NC}"
        exit 1
    fi
}

# 安装必要的工具
install_tools() {
    echo -e "${BLUE}📦 检查并安装必要的工具...${NC}"
    
    # 安装 golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "${YELLOW}安装 golangci-lint...${NC}"
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    fi
    
    # 安装 goimports
    if ! command -v goimports &> /dev/null; then
        echo -e "${YELLOW}安装 goimports...${NC}"
        go install golang.org/x/tools/cmd/goimports@latest
    fi
    
    # 安装 gosec
    if ! command -v gosec &> /dev/null; then
        echo -e "${YELLOW}安装 gosec...${NC}"
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
}

# 代码格式化检查
check_formatting() {
    echo -e "${BLUE}🎨 检查代码格式化...${NC}"
    
    # 检查 gofmt
    if ! gofmt -l . | grep -q .; then
        echo -e "${GREEN}✅ gofmt 检查通过${NC}"
    else
        echo -e "${RED}❌ gofmt 检查失败，发现格式问题：${NC}"
        gofmt -l .
        echo -e "${YELLOW}💡 运行 'gofmt -w .' 自动修复格式问题${NC}"
        return 1
    fi
    
    # 检查 goimports
    if ! goimports -l . | grep -q .; then
        echo -e "${GREEN}✅ goimports 检查通过${NC}"
    else
        echo -e "${RED}❌ goimports 检查失败，发现导入问题：${NC}"
        goimports -l .
        echo -e "${YELLOW}💡 运行 'goimports -w .' 自动修复导入问题${NC}"
        return 1
    fi
}

# 代码质量检查
check_quality() {
    echo -e "${BLUE}🔍 运行代码质量检查...${NC}"
    
    # 运行 golangci-lint
    if golangci-lint run --config .golangci.yml; then
        echo -e "${GREEN}✅ golangci-lint 检查通过${NC}"
    else
        echo -e "${RED}❌ golangci-lint 检查失败${NC}"
        return 1
    fi
}

# 安全检查
check_security() {
    echo -e "${BLUE}🔒 运行安全检查...${NC}"
    
    # 运行 gosec
    if gosec -fmt json -out gosec-report.json ./...; then
        echo -e "${GREEN}✅ gosec 安全检查通过${NC}"
    else
        echo -e "${YELLOW}⚠️  gosec 发现安全问题，请查看报告${NC}"
        if [ -f gosec-report.json ]; then
            echo -e "${BLUE}📊 安全检查报告已生成: gosec-report.json${NC}"
        fi
    fi
}

# 重复代码检查
check_duplicates() {
    echo -e "${BLUE}🔄 检查重复代码...${NC}"
    
    # 使用 gocloc 统计代码行数
    if command -v gocloc &> /dev/null; then
        gocloc --output-type=json . > gocloc-report.json
        echo -e "${GREEN}✅ 代码统计完成${NC}"
        echo -e "${BLUE}📊 代码统计报告已生成: gocloc-report.json${NC}"
    else
        echo -e "${YELLOW}⚠️  gocloc 未安装，跳过代码统计${NC}"
    fi
}

# 依赖检查
check_dependencies() {
    echo -e "${BLUE}📦 检查依赖...${NC}"
    
    # 检查是否有未使用的依赖
    if go mod tidy -v; then
        echo -e "${GREEN}✅ 依赖检查通过${NC}"
    else
        echo -e "${RED}❌ 依赖检查失败${NC}"
        return 1
    fi
    
    # 检查是否有安全漏洞
    if command -v govulncheck &> /dev/null; then
        echo -e "${BLUE}🔍 检查安全漏洞...${NC}"
        if govulncheck ./...; then
            echo -e "${GREEN}✅ 安全漏洞检查通过${NC}"
        else
            echo -e "${YELLOW}⚠️  发现安全漏洞，请及时修复${NC}"
        fi
    fi
}

# 测试覆盖率检查
check_coverage() {
    echo -e "${BLUE}🧪 检查测试覆盖率...${NC}"
    
    # 运行测试并生成覆盖率报告
    if go test -coverprofile=coverage.out -covermode=atomic ./...; then
        echo -e "${GREEN}✅ 测试通过${NC}"
        
        # 生成覆盖率报告
        go tool cover -html=coverage.out -o coverage.html
        echo -e "${BLUE}📊 覆盖率报告已生成: coverage.html${NC}"
        
        # 显示覆盖率统计
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo -e "${BLUE}📈 总覆盖率: ${coverage}${NC}"
        
        # 检查覆盖率是否达到要求
        coverage_num=$(echo $coverage | sed 's/%//')
        if (( $(echo "$coverage_num >= 70" | awk '{print ($1 >= 70)}') )); then
            echo -e "${GREEN}✅ 覆盖率达标 (≥70%)${NC}"
        else
            echo -e "${YELLOW}⚠️  覆盖率未达标 (<70%)${NC}"
        fi
    else
        echo -e "${RED}❌ 测试失败${NC}"
        return 1
    fi
}

# 生成质量报告
generate_report() {
    echo -e "${BLUE}📊 生成质量报告...${NC}"
    
    # 创建报告目录
    mkdir -p reports
    
    # 移动报告文件
    if [ -f gosec-report.json ]; then
        mv gosec-report.json reports/
    fi
    if [ -f gocloc-report.json ]; then
        mv gocloc-report.json reports/
    fi
    if [ -f coverage.out ]; then
        mv coverage.out reports/
    fi
    if [ -f coverage.html ]; then
        mv coverage.html reports/
    fi
    
    echo -e "${GREEN}✅ 质量报告已生成到 reports/ 目录${NC}"
}

# 主函数
main() {
    echo -e "${GREEN}🚀 开始代码质量检查流程${NC}"
    echo "=================================="
    
    # 安装工具
    install_tools
    
    # 检查工具
    check_tool golangci-lint
    check_tool goimports
    check_tool gosec
    
    # 执行检查
    local failed=0
    
    if ! check_formatting; then
        failed=1
    fi
    
    if ! check_quality; then
        failed=1
    fi
    
    check_security
    check_duplicates
    check_dependencies
    check_coverage
    
    # 生成报告
    generate_report
    
    echo "=================================="
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 所有检查通过！代码质量良好${NC}"
    else
        echo -e "${RED}❌ 部分检查失败，请修复问题后重新运行${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
