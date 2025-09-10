#!/bin/bash

# 测试运行脚本
# 用于执行不同类型的测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 打印帮助信息
print_help() {
    echo "测试运行脚本"
    echo ""
    echo "用法: $0 [选项] [测试模式]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -v, --verbose  详细输出"
    echo "  -c, --coverage 生成覆盖率报告"
    echo "  -p, --profile  生成性能分析报告"
    echo "  -r, --report   生成测试报告"
    echo "  -f, --fast     快速测试模式"
    echo "  -s, --skip     跳过特定测试"
    echo ""
    echo "测试模式:"
    echo "  unit           单元测试"
    echo "  integration    集成测试"
    echo "  performance    性能测试"
    echo "  all            所有测试"
    echo ""
    echo "示例:"
    echo "  $0 unit                    # 运行单元测试"
    echo "  $0 -c integration          # 运行集成测试并生成覆盖率报告"
    echo "  $0 -v -p all               # 运行所有测试，详细输出并生成性能分析"
    echo "  $0 -f unit                 # 快速运行单元测试"
}

# 检查依赖
check_dependencies() {
    print_message $BLUE "检查测试依赖..."
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        print_message $RED "错误: Go未安装或不在PATH中"
        exit 1
    fi
    
    # 检查Go版本
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    required_version="1.19"
    
    if [ "$(printf '%s\n' "$required_version" "$go_version" | sort -V | head -n1)" != "$required_version" ]; then
        print_message $RED "错误: 需要Go 1.19或更高版本，当前版本: $go_version"
        exit 1
    fi
    
    print_message $GREEN "✓ Go版本检查通过: $go_version"
    
    # 检查测试工具
    local tools=("testify" "gomock" "sqlmock")
    for tool in "${tools[@]}"; do
        if ! go list -m "$tool" &> /dev/null; then
            print_message $YELLOW "警告: $tool 未安装，正在安装..."
            go get "$tool"
        fi
    done
    
    print_message $GREEN "✓ 依赖检查完成"
}

# 设置测试环境
setup_test_env() {
    print_message $BLUE "设置测试环境..."
    
    # 创建必要的目录
    mkdir -p testdata
    mkdir -p coverage
    mkdir -p reports
    mkdir -p profiles
    
    # 设置环境变量
    export TEST_MODE=${TEST_MODE:-"all"}
    export TEST_VERBOSE=${TEST_VERBOSE:-"false"}
    export TEST_COVERAGE=${TEST_COVERAGE:-"true"}
    export TEST_PROFILE=${TEST_PROFILE:-"false"}
    export TEST_REPORT=${TEST_REPORT:-"true"}
    
    # 设置Go测试环境变量
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=1
    
    print_message $GREEN "✓ 测试环境设置完成"
}

# 运行单元测试
run_unit_tests() {
    print_message $BLUE "运行单元测试..."
    
    local test_args=""
    if [ "$VERBOSE" = true ]; then
        test_args="-v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args="$test_args -coverprofile=coverage/unit.out"
    fi
    
    if [ "$PROFILE" = true ]; then
        test_args="$test_args -cpuprofile=profiles/unit_cpu.prof -memprofile=profiles/unit_mem.prof"
    fi
    
    # 运行单元测试
    go test $test_args ./app/Testing/... -run "Test.*Unit" -timeout 30s
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ 单元测试通过"
    else
        print_message $RED "✗ 单元测试失败"
        exit 1
    fi
}

# 运行集成测试
run_integration_tests() {
    print_message $BLUE "运行集成测试..."
    
    local test_args=""
    if [ "$VERBOSE" = true ]; then
        test_args="-v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args="$test_args -coverprofile=coverage/integration.out"
    fi
    
    if [ "$PROFILE" = true ]; then
        test_args="$test_args -cpuprofile=profiles/integration_cpu.prof -memprofile=profiles/integration_mem.prof"
    fi
    
    # 运行集成测试
    go test $test_args ./app/Testing/... -run "Test.*Integration" -timeout 60s
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ 集成测试通过"
    else
        print_message $RED "✗ 集成测试失败"
        exit 1
    fi
}

# 运行性能测试
run_performance_tests() {
    print_message $BLUE "运行性能测试..."
    
    local test_args=""
    if [ "$VERBOSE" = true ]; then
        test_args="-v"
    fi
    
    if [ "$PROFILE" = true ]; then
        test_args="$test_args -cpuprofile=profiles/performance_cpu.prof -memprofile=profiles/performance_mem.prof"
    fi
    
    # 运行性能测试
    go test $test_args ./app/Testing/... -run "Test.*Performance" -timeout 300s
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ 性能测试通过"
    else
        print_message $RED "✗ 性能测试失败"
        exit 1
    fi
}

# 运行所有测试
run_all_tests() {
    print_message $BLUE "运行所有测试..."
    
    local test_args=""
    if [ "$VERBOSE" = true ]; then
        test_args="-v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args="$test_args -coverprofile=coverage/all.out"
    fi
    
    if [ "$PROFILE" = true ]; then
        test_args="$test_args -cpuprofile=profiles/all_cpu.prof -memprofile=profiles/all_mem.prof"
    fi
    
    # 运行所有测试
    go test $test_args ./app/Testing/... -timeout 600s
    
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ 所有测试通过"
    else
        print_message $RED "✗ 部分测试失败"
        exit 1
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    if [ "$COVERAGE" != true ]; then
        return
    fi
    
    print_message $BLUE "生成覆盖率报告..."
    
    # 合并覆盖率文件
    if [ -f "coverage/unit.out" ] && [ -f "coverage/integration.out" ]; then
        echo "mode: set" > coverage/merged.out
        tail -n +2 coverage/unit.out >> coverage/merged.out
        tail -n +2 coverage/integration.out >> coverage/merged.out
        go tool cover -html=coverage/merged.out -o coverage/report.html
    elif [ -f "coverage/unit.out" ]; then
        go tool cover -html=coverage/unit.out -o coverage/report.html
    elif [ -f "coverage/integration.out" ]; then
        go tool cover -html=coverage/integration.out -o coverage/report.html
    elif [ -f "coverage/all.out" ]; then
        go tool cover -html=coverage/all.out -o coverage/report.html
    fi
    
    # 显示覆盖率统计
    if [ -f "coverage/merged.out" ]; then
        go tool cover -func=coverage/merged.out
    elif [ -f "coverage/all.out" ]; then
        go tool cover -func=coverage/all.out
    fi
    
    print_message $GREEN "✓ 覆盖率报告生成完成: coverage/report.html"
}

# 生成测试报告
generate_test_report() {
    if [ "$REPORT" != true ]; then
        return
    fi
    
    print_message $BLUE "生成测试报告..."
    
    # 创建测试报告目录
    local report_dir="reports/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$report_dir"
    
    # 生成JUnit XML报告
    if command -v go-junit-report &> /dev/null; then
        go test -v ./app/Testing/... 2>&1 | go-junit-report > "$report_dir/junit.xml"
    fi
    
    # 生成HTML报告
    if [ -f "coverage/report.html" ]; then
        cp coverage/report.html "$report_dir/"
    fi
    
    # 生成性能分析报告
    if [ "$PROFILE" = true ] && [ -d "profiles" ]; then
        cp -r profiles "$report_dir/"
    fi
    
    # 生成测试摘要
    {
        echo "# 测试报告"
        echo "生成时间: $(date)"
        echo "测试模式: $TEST_MODE"
        echo ""
        echo "## 测试结果"
        echo "- 单元测试: $(if [ -f "coverage/unit.out" ]; then echo "通过"; else echo "未运行"; fi)"
        echo "- 集成测试: $(if [ -f "coverage/integration.out" ]; then echo "通过"; else echo "未运行"; fi)"
        echo "- 性能测试: $(if [ -d "profiles" ]; then echo "完成"; else echo "未运行"; fi)"
        echo ""
        echo "## 覆盖率"
        if [ -f "coverage/merged.out" ]; then
            go tool cover -func=coverage/merged.out | tail -1
        fi
        echo ""
        echo "## 详细报告"
        echo "- 覆盖率报告: report.html"
        echo "- JUnit报告: junit.xml"
        if [ "$PROFILE" = true ]; then
            echo "- 性能分析: profiles/"
        fi
    } > "$report_dir/README.md"
    
    print_message $GREEN "✓ 测试报告生成完成: $report_dir"
}

# 清理测试文件
cleanup_test_files() {
    print_message $BLUE "清理测试文件..."
    
    # 清理临时文件
    rm -rf testdata/*.db
    rm -rf testdata/*.sqlite
    rm -rf testdata/*.tmp
    
    # 清理覆盖率文件（保留报告）
    if [ "$KEEP_COVERAGE" != true ]; then
        rm -f coverage/*.out
    fi
    
    # 清理性能分析文件（保留报告）
    if [ "$KEEP_PROFILES" != true ]; then
        rm -f profiles/*.prof
    fi
    
    print_message $GREEN "✓ 测试文件清理完成"
}

# 主函数
main() {
    # 解析命令行参数
    VERBOSE=false
    COVERAGE=false
    PROFILE=false
    REPORT=false
    FAST=false
    SKIP=""
    TEST_MODE="all"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_help
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -p|--profile)
                PROFILE=true
                shift
                ;;
            -r|--report)
                REPORT=true
                shift
                ;;
            -f|--fast)
                FAST=true
                shift
                ;;
            -s|--skip)
                SKIP="$2"
                shift 2
                ;;
            unit|integration|performance|all)
                TEST_MODE="$1"
                shift
                ;;
            *)
                print_message $RED "未知选项: $1"
                print_help
                exit 1
                ;;
        esac
    done
    
    # 设置环境变量
    export TEST_VERBOSE=$VERBOSE
    export TEST_COVERAGE=$COVERAGE
    export TEST_PROFILE=$PROFILE
    export TEST_REPORT=$REPORT
    
    print_message $BLUE "开始执行测试..."
    print_message $BLUE "测试模式: $TEST_MODE"
    print_message $BLUE "详细输出: $VERBOSE"
    print_message $BLUE "覆盖率: $COVERAGE"
    print_message $BLUE "性能分析: $PROFILE"
    print_message $BLUE "生成报告: $REPORT"
    echo ""
    
    # 检查依赖
    check_dependencies
    
    # 设置测试环境
    setup_test_env
    
    # 根据测试模式运行测试
    case $TEST_MODE in
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        performance)
            run_performance_tests
            ;;
        all)
            run_all_tests
            ;;
        *)
            print_message $RED "未知的测试模式: $TEST_MODE"
            exit 1
            ;;
    esac
    
    # 生成覆盖率报告
    generate_coverage_report
    
    # 生成测试报告
    generate_test_report
    
    # 清理测试文件
    cleanup_test_files
    
    print_message $GREEN "所有测试完成！"
}

# 执行主函数
main "$@"
