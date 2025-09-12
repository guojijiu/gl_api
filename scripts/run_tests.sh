#!/bin/bash

# 云平台API测试运行脚本
# 功能：运行所有测试，包括单元测试、集成测试、性能测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go_env() {
    log_info "检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go未安装或未在PATH中"
        exit 1
    fi
    
    go_version=$(go version | awk '{print $3}')
    log_success "Go版本: $go_version"
}

# 安装测试依赖
install_dependencies() {
    log_info "安装测试依赖..."
    
    # 安装测试框架
    go mod tidy
    
    # 安装测试工具
    if ! command -v go-junit-report &> /dev/null; then
        log_info "安装go-junit-report..."
        go install github.com/jstemmer/go-junit-report@latest
    fi
    
    if ! command -v gocov &> /dev/null; then
        log_info "安装gocov..."
        go install github.com/axw/gocov/gocov@latest
    fi
    
    if ! command -v gocov-xml &> /dev/null; then
        log_info "安装gocov-xml..."
        go install github.com/AlekSi/gocov-xml@latest
    fi
    
    log_success "依赖安装完成"
}

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    
    # 创建测试结果目录
    mkdir -p test-results
    
    # 运行单元测试
    go test -v -race -coverprofile=test-results/coverage.out -covermode=atomic ./... 2>&1 | tee test-results/unit-test.log
    
    # 生成测试报告
    go test -v ./... 2>&1 | go-junit-report > test-results/unit-test.xml
    
    # 生成覆盖率报告
    gocov convert test-results/coverage.out | gocov-xml > test-results/coverage.xml
    
    log_success "单元测试完成"
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    
    # 设置测试环境变量
    export TEST_ENV=integration
    export TEST_DB_URL="postgres://test:test@localhost:5432/test_db?sslmode=disable"
    export TEST_REDIS_URL="redis://localhost:6379/1"
    
    # 运行集成测试
    go test -v -tags=integration ./tests/Integration/... 2>&1 | tee test-results/integration-test.log
    
    log_success "集成测试完成"
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    # 运行基准测试
    go test -bench=. -benchmem ./tests/benchmark/... 2>&1 | tee test-results/benchmark.log
    
    # 运行负载测试
    go test -v -tags=load ./tests/benchmark/... 2>&1 | tee test-results/load-test.log
    
    log_success "性能测试完成"
}

# 运行安全测试
run_security_tests() {
    log_info "运行安全测试..."
    
    # 运行安全测试
    go test -v -tags=security ./tests/... 2>&1 | tee test-results/security-test.log
    
    log_success "安全测试完成"
}

# 生成测试报告
generate_report() {
    log_info "生成测试报告..."
    
    # 创建HTML报告
    cat > test-results/index.html << EOF
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
        <p>生成时间: $(date)</p>
    </div>
    
    <div class="section">
        <h2>测试概览</h2>
        <p>总测试数: $(grep -c "PASS\|FAIL" test-results/unit-test.log || echo "0")</p>
        <p>通过数: $(grep -c "PASS" test-results/unit-test.log || echo "0")</p>
        <p>失败数: $(grep -c "FAIL" test-results/unit-test.log || echo "0")</p>
    </div>
    
    <div class="section">
        <h2>单元测试结果</h2>
        <pre>$(cat test-results/unit-test.log)</pre>
    </div>
    
    <div class="section">
        <h2>集成测试结果</h2>
        <pre>$(cat test-results/integration-test.log)</pre>
    </div>
    
    <div class="section">
        <h2>性能测试结果</h2>
        <pre>$(cat test-results/benchmark.log)</pre>
    </div>
    
    <div class="section">
        <h2>安全测试结果</h2>
        <pre>$(cat test-results/security-test.log)</pre>
    </div>
</body>
</html>
EOF
    
    log_success "测试报告已生成: test-results/index.html"
}

# 清理测试环境
cleanup() {
    log_info "清理测试环境..."
    
    # 清理临时文件
    rm -f test-results/*.log
    rm -f test-results/*.xml
    rm -f test-results/coverage.out
    
    log_success "清理完成"
}

# 主函数
main() {
    log_info "开始运行云平台API测试套件..."
    
    # 检查参数
    case "${1:-all}" in
        "unit")
            check_go_env
            install_dependencies
            run_unit_tests
            generate_report
            ;;
        "integration")
            check_go_env
            install_dependencies
            run_integration_tests
            generate_report
            ;;
        "performance")
            check_go_env
            install_dependencies
            run_performance_tests
            generate_report
            ;;
        "security")
            check_go_env
            install_dependencies
            run_security_tests
            generate_report
            ;;
        "all")
            check_go_env
            install_dependencies
            run_unit_tests
            run_integration_tests
            run_performance_tests
            run_security_tests
            generate_report
            ;;
        "clean")
            cleanup
            ;;
        *)
            echo "用法: $0 [unit|integration|performance|security|all|clean]"
            echo "  unit        - 运行单元测试"
            echo "  integration - 运行集成测试"
            echo "  performance - 运行性能测试"
            echo "  security    - 运行安全测试"
            echo "  all         - 运行所有测试（默认）"
            echo "  clean       - 清理测试环境"
            exit 1
            ;;
    esac
    
    log_success "测试套件运行完成！"
}

# 运行主函数
main "$@"