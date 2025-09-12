#!/bin/bash

# 测试运行脚本
# 用于运行各种类型的测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认参数
TEST_TYPE="all"
COVERAGE=false
VERBOSE=false
BENCHMARK=false
INTEGRATION=false

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -t, --type TYPE     测试类型 (unit|integration|all) [默认: all]"
    echo "  -c, --coverage      生成覆盖率报告"
    echo "  -v, --verbose       详细输出"
    echo "  -b, --benchmark     运行性能测试"
    echo "  -i, --integration   运行集成测试"
    echo "  -h, --help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -t unit -c       运行单元测试并生成覆盖率报告"
    echo "  $0 -t integration   运行集成测试"
    echo "  $0 -b              运行性能测试"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -b|--benchmark)
            BENCHMARK=true
            shift
            ;;
        -i|--integration)
            INTEGRATION=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 设置测试参数
TEST_ARGS=""
if [ "$VERBOSE" = true ]; then
    TEST_ARGS="$TEST_ARGS -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_ARGS="$TEST_ARGS -coverprofile=coverage.out -covermode=atomic"
fi

# 运行单元测试
run_unit_tests() {
    echo -e "${BLUE}🧪 运行单元测试...${NC}"
    
    if [ "$COVERAGE" = true ]; then
        go test $TEST_ARGS ./tests/Container/... ./tests/Utils/... ./tests/Models/...
    else
        go test $TEST_ARGS ./tests/Container/... ./tests/Utils/... ./tests/Models/...
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 单元测试通过${NC}"
    else
        echo -e "${RED}❌ 单元测试失败${NC}"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    echo -e "${BLUE}🔗 运行集成测试...${NC}"
    
    # 设置集成测试环境变量
    export TEST_ENV=true
    export DB_DRIVER=sqlite
    export DB_DATABASE=:memory:
    
    if [ "$COVERAGE" = true ]; then
        go test $TEST_ARGS ./tests/Integration/...
    else
        go test $TEST_ARGS ./tests/Integration/...
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 集成测试通过${NC}"
    else
        echo -e "${RED}❌ 集成测试失败${NC}"
        return 1
    fi
}

# 运行性能测试
run_benchmark_tests() {
    echo -e "${BLUE}⚡ 运行性能测试...${NC}"
    
    go test -bench=. -benchmem ./tests/... | tee benchmark_results.txt
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 性能测试完成${NC}"
        echo -e "${BLUE}📊 性能测试结果已保存到 benchmark_results.txt${NC}"
    else
        echo -e "${RED}❌ 性能测试失败${NC}"
        return 1
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    if [ "$COVERAGE" = true ]; then
        echo -e "${BLUE}📊 生成覆盖率报告...${NC}"
        
        # 生成HTML报告
        go tool cover -html=coverage.out -o coverage.html
    
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
        
        echo -e "${BLUE}📊 覆盖率报告已生成: coverage.html${NC}"
    fi
}

# 清理测试文件
cleanup() {
    echo -e "${BLUE}🧹 清理测试文件...${NC}"
    
    # 删除测试生成的临时文件
    rm -f test.db
    rm -f test_*.log
    rm -f temp_*
    
    echo -e "${GREEN}✅ 清理完成${NC}"
}

# 主函数
main() {
    echo -e "${GREEN}🚀 开始测试流程${NC}"
    echo "=================================="
    
    local failed=0
    
    # 根据测试类型运行相应的测试
    case $TEST_TYPE in
        "unit")
            if ! run_unit_tests; then
                failed=1
            fi
            ;;
        "integration")
            if ! run_integration_tests; then
                failed=1
            fi
            ;;
        "all")
            if ! run_unit_tests; then
                failed=1
            fi
            
            if ! run_integration_tests; then
                failed=1
            fi
            ;;
        *)
            echo -e "${RED}❌ 未知的测试类型: $TEST_TYPE${NC}"
            show_help
            exit 1
            ;;
    esac
    
    # 运行性能测试
    if [ "$BENCHMARK" = true ]; then
        if ! run_benchmark_tests; then
            failed=1
        fi
    fi
    
    # 生成覆盖率报告
    generate_coverage_report
    
    # 清理测试文件
    cleanup
    
    echo "=================================="
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 所有测试通过！${NC}"
    else
        echo -e "${RED}❌ 部分测试失败，请检查错误信息${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"