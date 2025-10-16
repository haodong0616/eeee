#!/bin/bash

echo "🌐 Velocity Exchange - 移动端访问配置"
echo "=========================================="
echo ""

# 获取局域网IP
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null)
else
    # Linux
    IP=$(hostname -I | awk '{print $1}')
fi

if [ -z "$IP" ]; then
    echo "⚠️  无法自动获取IP地址"
    echo "请手动查看："
    echo "  Mac: 系统偏好设置 -> 网络"
    echo "  Linux: ip addr show"
    echo ""
    read -p "请输入你的局域网IP: " IP
fi

echo "📱 局域网IP: $IP"
echo ""
echo "✅ 手机访问地址（确保在同一WiFi）："
echo "   http://$IP:3000"
echo ""
echo "🔧 后端API地址："
echo "   http://$IP:8080"
echo ""

# 提示
echo "📝 说明："
echo "  1. 前端会自动使用当前访问地址的主机名连接后端"
echo "  2. 手机访问 http://$IP:3000 即可"
echo "  3. 无需配置 .env.local 文件"
echo ""

# 询问是否继续
read -p "是否启动前端服务？(y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    cd "$(dirname "$0")/frontend"
    echo ""
    echo "🚀 启动前端服务..."
    echo "================================================"
    npm run dev
fi

