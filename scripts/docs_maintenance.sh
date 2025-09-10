#!/bin/bash

# 文档维护脚本
# 用于检查文档链接、格式和完整性

echo "🔍 检查文档完整性..."

# 检查所有Markdown文件
find . -name "*.md" -type f | while read file; do
    echo "检查文件: $file"
    
    # 检查是否有空行
    if [ ! -s "$file" ]; then
        echo "❌ 警告: $file 是空文件"
    fi
    
    # 检查是否有标题
    if ! grep -q "^#" "$file"; then
        echo "❌ 警告: $file 没有标题"
    fi
    
    # 检查链接格式
    if grep -q "\[.*\](.*)" "$file"; then
        echo "✅ $file 包含链接"
    fi
done

echo "📊 生成文档统计..."

# 统计文档数量
md_count=$(find . -name "*.md" -type f | wc -l)
echo "总文档数量: $md_count"

# 统计总行数
total_lines=$(find . -name "*.md" -type f -exec wc -l {} + | tail -1 | awk '{print $1}')
echo "总行数: $total_lines"

# 统计各目录文档数量
echo "📁 各目录文档数量:"
find . -name "*.md" -type f | cut -d'/' -f2 | sort | uniq -c | sort -nr

echo "✅ 文档检查完成"
