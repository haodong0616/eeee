#!/bin/bash

echo "🔧 开始修复前端样式问题..."
echo ""

# 检查是否在正确的目录
if [ ! -f "package.json" ]; then
    echo "❌ 错误：请在 frontend 目录下运行此脚本"
    exit 1
fi

echo "1️⃣  停止现有的开发服务器..."
pkill -f "next dev" 2>/dev/null || true
sleep 2

echo "2️⃣  清理缓存和旧文件..."
rm -rf .next
rm -rf node_modules
rm -f package-lock.json
rm -f yarn.lock
echo "   ✅ 清理完成"

echo ""
echo "3️⃣  重新安装依赖..."
npm install
if [ $? -ne 0 ]; then
    echo "   ❌ 安装失败，请检查网络连接"
    exit 1
fi
echo "   ✅ 依赖安装完成"

echo ""
echo "4️⃣  验证配置文件..."

# 检查必要的文件
files=("tailwind.config.ts" "postcss.config.mjs" "app/globals.css" "app/layout.tsx")
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "   ✅ $file 存在"
    else
        echo "   ❌ $file 缺失"
    fi
done

echo ""
echo "✨ 修复完成！"
echo ""
echo "现在请运行以下命令启动开发服务器："
echo "   npm run dev"
echo ""
echo "然后访问: http://localhost:3000"
echo ""


