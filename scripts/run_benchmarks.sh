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

# 用户相关性能测试
echo "👤 测试用户注册性能..."
go test -bench=BenchmarkUserRegistration -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/user_registration.txt

echo "🔐 测试用户登录性能..."
go test -bench=BenchmarkUserLogin -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/user_login.txt

echo "👤 测试获取用户资料性能..."
go test -bench=BenchmarkGetUserProfile -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/get_user_profile.txt

# 文章相关性能测试
echo "📝 测试创建文章性能..."
go test -bench=BenchmarkCreatePost -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/create_post.txt

echo "📋 测试获取文章列表性能..."
go test -bench=BenchmarkGetPosts -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/get_posts.txt

# 系统性能测试
echo "🏥 测试健康检查性能..."
go test -bench=BenchmarkHealthCheck -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/health_check.txt

echo "⚡ 测试并发请求性能..."
go test -bench=BenchmarkConcurrentRequests -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/concurrent_requests.txt

# 数据库性能测试
echo "🗄️ 测试数据库操作性能..."
go test -bench=BenchmarkDatabaseOperations -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/database_operations.txt

echo "🔄 测试并发数据库写入性能..."
go test -bench=BenchmarkConcurrentDatabaseWrites -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/concurrent_db_writes.txt

# JWT性能测试
echo "🎫 测试JWT生成性能..."
go test -bench=BenchmarkJWTGeneration -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/jwt_generation.txt

echo "🔍 测试JWT验证性能..."
go test -bench=BenchmarkJWTValidation -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/jwt_validation.txt

# 密码性能测试
echo "🔒 测试密码哈希性能..."
go test -bench=BenchmarkPasswordHashing -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/password_hashing.txt

echo "✅ 测试密码验证性能..."
go test -bench=BenchmarkPasswordVerification -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/password_verification.txt

# 内存使用测试
echo "💾 测试内存使用性能..."
go test -bench=BenchmarkMemoryUsage -benchmem -run=^$ ./tests/benchmark/... > ./benchmark_results/memory_usage.txt

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

### 用户相关性能
EOF

# 添加用户注册性能结果
echo "### 用户注册性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/user_registration.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加用户登录性能结果
echo "### 用户登录性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/user_login.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加获取用户资料性能结果
echo "### 获取用户资料性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/get_user_profile.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加文章相关性能结果
echo "### 文章相关性能" >> ./benchmark_results/performance_report.md
echo "#### 创建文章性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/create_post.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

echo "#### 获取文章列表性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/get_posts.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加系统性能结果
echo "### 系统性能" >> ./benchmark_results/performance_report.md
echo "#### 健康检查性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/health_check.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

echo "#### 并发请求性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/concurrent_requests.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加数据库性能结果
echo "### 数据库性能" >> ./benchmark_results/performance_report.md
echo "#### 数据库操作性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/database_operations.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

echo "#### 并发数据库写入性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/concurrent_db_writes.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加JWT性能结果
echo "### JWT性能" >> ./benchmark_results/performance_report.md
echo "#### JWT生成性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/jwt_generation.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

echo "#### JWT验证性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/jwt_validation.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加密码性能结果
echo "### 密码性能" >> ./benchmark_results/performance_report.md
echo "#### 密码哈希性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/password_hashing.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

echo "#### 密码验证性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/password_verification.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加内存使用结果
echo "### 内存使用性能" >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
cat ./benchmark_results/memory_usage.txt >> ./benchmark_results/performance_report.md
echo '```' >> ./benchmark_results/performance_report.md
echo "" >> ./benchmark_results/performance_report.md

# 添加总结
cat >> ./benchmark_results/performance_report.md << EOF

## 性能总结

### 关键指标
- 用户注册: 查看用户注册性能结果
- 用户登录: 查看用户登录性能结果
- 文章创建: 查看文章创建性能结果
- 健康检查: 查看健康检查性能结果
- 并发处理: 查看并发请求性能结果

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
