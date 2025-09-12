#!/bin/bash

# 性能基准测试运行脚本

echo "🚀 开始运行性能基准测试..."

# 设置测试环境
export GIN_MODE=test
export GO_ENV=test

# 创建测试目录
mkdir -p ./test_storage
mkdir -p ./benchmark_results

# 运行基准测试
echo "📊 运行基准测试..."

# 运行所有基准测试
echo "🧪 运行基准测试..."
go test -bench=. -benchmem -run=^$ ./tests/... > ./benchmark_results/all_benchmarks.txt

# 运行性能测试脚本
echo "⚡ 运行性能测试脚本..."
go run ./scripts/performance_test.go > ./benchmark_results/performance_test.txt

# 生成性能报告
echo "📈 生成性能报告..."
cat > ./benchmark_results/performance_report.md << EOF
# 性能基准测试报告

生成时间: $(date)

## 测试环境
- Go版本: $(go version)
- 操作系统: $(uname -s)
- 架构: $(uname -m)
- CPU核心数: $(nproc)

## 测试结果

### 基准测试结果
EOF

# 添加基准测试结果
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/all_benchmarks.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加性能测试结果
echo "### 性能测试结果" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/performance_test.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加总结
cat >> ./benchmark_results/performance_report.md << EOF

## 性能总结

### 优化建议
1. 根据测试结果识别性能瓶颈
2. 优化数据库查询和索引
3. 实施缓存策略
4. 优化JWT处理
5. 监控内存使用情况

### 监控建议
1. 定期运行基准测试
2. 监控生产环境性能指标
3. 设置性能告警阈值
4. 持续优化关键路径

---
*报告生成时间: $(date)*
EOF

echo "✅ 性能基准测试完成！"
echo "📊 测试结果保存在 ./benchmark_results/ 目录中"
echo "📈 性能报告: ./benchmark_results/performance_report.md"

# 清理测试目录
rm -rf ./test_storage

echo "🧹 清理测试环境完成"
