#!/bin/bash

# 云平台API性能测试脚本
# 功能：运行性能测试和基准测试

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

# 安装性能测试工具
install_tools() {
    log_info "安装性能测试工具..."
    
    # 安装hey (HTTP负载测试工具)
    if ! command -v hey &> /dev/null; then
        log_info "安装hey..."
        go install github.com/rakyll/hey@latest
    fi
    
    # 安装pprof工具
    if ! command -v pprof &> /dev/null; then
        log_info "安装pprof..."
        go install github.com/google/pprof@latest
    fi
    
    log_success "工具安装完成"
}

# 启动API服务
start_api_server() {
    log_info "启动API服务..."
    
    # 设置环境变量
    export PORT=8080
    export DATABASE_URL="postgres://test:test@localhost:5432/test_db?sslmode=disable"
    export REDIS_URL="redis://localhost:6379/1"
    export LOG_LEVEL="error"
    
    # 启动服务
    go run main.go &
    API_PID=$!
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务是否启动
    if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
        log_error "API服务启动失败"
        kill $API_PID 2>/dev/null || true
        exit 1
    fi
    
    log_success "API服务已启动 (PID: $API_PID)"
}

# 停止API服务
stop_api_server() {
    if [ ! -z "$API_PID" ]; then
        log_info "停止API服务..."
        kill $API_PID 2>/dev/null || true
        wait $API_PID 2>/dev/null || true
        log_success "API服务已停止"
    fi
}

# 运行基准测试
run_benchmarks() {
    log_info "运行基准测试..."
    
    # 创建结果目录
    mkdir -p benchmark-results
    
    # 运行基准测试
    go test -bench=. -benchmem -benchtime=30s ./tests/benchmark/... 2>&1 | tee benchmark-results/benchmark.log
    
    # 生成基准测试报告
    go test -bench=. -benchmem -benchtime=30s ./tests/benchmark/... -json 2>&1 | tee benchmark-results/benchmark.json
    
    log_success "基准测试完成"
}

# 运行负载测试
run_load_tests() {
    log_info "运行负载测试..."
    
    # 健康检查端点负载测试
    log_info "测试健康检查端点..."
    hey -n 10000 -c 100 -m GET http://localhost:8080/api/v1/health 2>&1 | tee benchmark-results/health-load.log
    
    # 详细健康检查端点负载测试
    log_info "测试详细健康检查端点..."
    hey -n 5000 -c 50 -m GET http://localhost:8080/api/v1/health/detailed 2>&1 | tee benchmark-results/detailed-health-load.log
    
    # API文档端点负载测试
    log_info "测试API文档端点..."
    hey -n 1000 -c 10 -m GET http://localhost:8080/api/v1/docs 2>&1 | tee benchmark-results/docs-load.log
    
    log_success "负载测试完成"
}

# 运行压力测试
run_stress_tests() {
    log_info "运行压力测试..."
    
    # 逐步增加并发数
    for concurrency in 10 50 100 200 500 1000; do
        log_info "测试并发数: $concurrency"
        hey -n 10000 -c $concurrency -m GET http://localhost:8080/api/v1/health 2>&1 | tee benchmark-results/stress-${concurrency}.log
        sleep 5
    done
    
    log_success "压力测试完成"
}

# 运行内存泄漏测试
run_memory_leak_tests() {
    log_info "运行内存泄漏测试..."
    
    # 运行长时间测试
    hey -n 100000 -c 100 -m GET http://localhost:8080/api/v1/health 2>&1 | tee benchmark-results/memory-leak.log
    
    log_success "内存泄漏测试完成"
}

# 生成性能报告
generate_performance_report() {
    log_info "生成性能报告..."
    
    # 创建HTML报告
    cat > benchmark-results/index.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>云平台API性能测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
        pre { background-color: #f5f5f5; padding: 10px; border-radius: 3px; overflow-x: auto; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>云平台API性能测试报告</h1>
        <p>生成时间: $(date)</p>
    </div>
    
    <div class="section">
        <h2>基准测试结果</h2>
        <pre>$(cat benchmark-results/benchmark.log)</pre>
    </div>
    
    <div class="section">
        <h2>负载测试结果</h2>
        <h3>健康检查端点</h3>
        <pre>$(cat benchmark-results/health-load.log)</pre>
        
        <h3>详细健康检查端点</h3>
        <pre>$(cat benchmark-results/detailed-health-load.log)</pre>
        
        <h3>API文档端点</h3>
        <pre>$(cat benchmark-results/docs-load.log)</pre>
    </div>
    
    <div class="section">
        <h2>压力测试结果</h2>
        <table>
            <tr>
                <th>并发数</th>
                <th>请求数</th>
                <th>平均响应时间</th>
                <th>QPS</th>
                <th>错误率</th>
            </tr>
EOF

    # 解析压力测试结果
    for concurrency in 10 50 100 200 500 1000; do
        if [ -f "benchmark-results/stress-${concurrency}.log" ]; then
            # 提取关键指标
            requests=$(grep "Total:" benchmark-results/stress-${concurrency}.log | awk '{print $2}' || echo "0")
            avg_time=$(grep "Average:" benchmark-results/stress-${concurrency}.log | awk '{print $2}' || echo "0")
            qps=$(grep "Requests/sec:" benchmark-results/stress-${concurrency}.log | awk '{print $2}' || echo "0")
            error_rate=$(grep "Error distribution:" benchmark-results/stress-${concurrency}.log | awk '{print $3}' || echo "0%")
            
            cat >> benchmark-results/index.html << EOF
            <tr>
                <td>$concurrency</td>
                <td>$requests</td>
                <td>$avg_time</td>
                <td>$qps</td>
                <td>$error_rate</td>
            </tr>
EOF
        fi
    done
    
    cat >> benchmark-results/index.html << EOF
        </table>
    </div>
    
    <div class="section">
        <h2>内存泄漏测试结果</h2>
        <pre>$(cat benchmark-results/memory-leak.log)</pre>
    </div>
</body>
</html>
EOF
    
    log_success "性能报告已生成: benchmark-results/index.html"
}

# 清理测试环境
cleanup() {
    log_info "清理测试环境..."
    
    # 停止API服务
    stop_api_server
    
    # 清理临时文件
    rm -f benchmark-results/*.log
    rm -f benchmark-results/*.json
    
    log_success "清理完成"
}

# 主函数
main() {
    log_info "开始运行云平台API性能测试..."
    
    # 检查参数
    case "${1:-all}" in
        "benchmark")
            check_go_env
            install_tools
            run_benchmarks
            generate_performance_report
            ;;
        "load")
            check_go_env
            install_tools
            start_api_server
            run_load_tests
            stop_api_server
            generate_performance_report
            ;;
        "stress")
            check_go_env
            install_tools
            start_api_server
            run_stress_tests
            stop_api_server
            generate_performance_report
            ;;
        "memory")
            check_go_env
            install_tools
            start_api_server
            run_memory_leak_tests
            stop_api_server
            generate_performance_report
            ;;
        "all")
            check_go_env
            install_tools
            run_benchmarks
            start_api_server
            run_load_tests
            run_stress_tests
            run_memory_leak_tests
            stop_api_server
            generate_performance_report
            ;;
        "clean")
            cleanup
            ;;
        *)
            echo "用法: $0 [benchmark|load|stress|memory|all|clean]"
            echo "  benchmark - 运行基准测试"
            echo "  load      - 运行负载测试"
            echo "  stress    - 运行压力测试"
            echo "  memory    - 运行内存泄漏测试"
            echo "  all       - 运行所有测试（默认）"
            echo "  clean     - 清理测试环境"
            exit 1
            ;;
    esac
    
    log_success "性能测试完成！"
}

# 设置退出时清理
trap cleanup EXIT

# 运行主函数
main "$@"