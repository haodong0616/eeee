# 📱 手机局域网访问指南

## 🎯 智能代理模式（推荐）

现在使用 Next.js API Routes 代理，**完全自动化，零配置**！

## 快速开始

### 1. 启动后端

```bash
cd backend
go run main.go
```

### 2. 启动前端（已配置局域网访问）

```bash
cd frontend
npm run dev
```

服务会自动监听 `0.0.0.0:3000`，支持局域网访问。

### 3. 查看你的局域网地址

前端启动后会显示：
```
- Local:    http://localhost:3000
- Network:  http://192.168.1.100:3000  ← 用这个！
```

### 4. 手机访问

在手机浏览器输入 Network 地址：
```
http://192.168.1.100:3000
```

**就这么简单！** 🎉

## 注意事项

### ⚠️ 后端API地址

如果后端也在同一台电脑上运行，需要确保后端也支持局域网访问。

**检查后端配置**：
```bash
cd backend
# 确保监听 0.0.0.0:8080 或 :8080
```

**更新前端 API 地址**（如果需要）：

创建 `frontend/.env.local`：
```env
NEXT_PUBLIC_API_URL=http://192.168.1.100:8080
NEXT_PUBLIC_WS_URL=ws://192.168.1.100:8080
```

将 `192.168.1.100` 替换为你的实际IP地址。

### 🔥 防火墙设置

确保防火墙允许端口访问：

**Mac**：
```bash
# 检查防火墙状态
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# 如果开启，允许端口
# 系统偏好设置 -> 安全性与隐私 -> 防火墙 -> 防火墙选项
```

**Windows**：
```bash
# 允许端口 3000 和 8080
netsh advfirewall firewall add rule name="Next.js Dev" dir=in action=allow protocol=TCP localport=3000
netsh advfirewall firewall add rule name="Golang Backend" dir=in action=allow protocol=TCP localport=8080
```

### 📱 手机端要求

1. **同一WiFi网络**：手机和电脑必须在同一局域网
2. **输入完整URL**：包括 `http://`
3. **使用局域网IP**：不要用 `localhost`

## 完整启动流程

### 方式1：使用环境变量（推荐）

1. **获取你的IP地址**：
```bash
# Mac/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}'

# 假设输出: 192.168.1.100
```

2. **创建环境变量文件**：
```bash
cd frontend
cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://192.168.1.100:8080
NEXT_PUBLIC_WS_URL=ws://192.168.1.100:8080
EOF
```

3. **启动服务**：
```bash
# 启动后端
cd backend
go run main.go

# 启动前端（新终端）
cd frontend
npm run dev
```

4. **手机访问**：
```
http://192.168.1.100:3000
```

### 方式2：自动化脚本

创建 `start-mobile.sh`：
```bash
#!/bin/bash

# 获取局域网IP
IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)

echo "🌐 局域网IP: $IP"
echo ""
echo "📱 手机访问地址："
echo "   http://$IP:3000"
echo ""
echo "🔧 后端API地址："
echo "   http://$IP:8080"
echo ""

# 创建环境变量
cd frontend
cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://$IP:8080
NEXT_PUBLIC_WS_URL=ws://$IP:8080
EOF

echo "✅ 环境变量已配置"
echo ""
echo "🚀 启动中..."
npm run dev
```

使用：
```bash
chmod +x start-mobile.sh
./start-mobile.sh
```

## 常见问题

### Q: 手机无法访问？

**检查清单**：
1. ✅ 手机和电脑在同一WiFi？
2. ✅ 使用正确的IP地址？（不是127.0.0.1）
3. ✅ URL包含 `http://`？
4. ✅ 端口号正确？（3000）
5. ✅ 防火墙允许？

### Q: API请求失败？

确保前端配置了正确的后端地址：

```bash
cd frontend
cat .env.local
# 应该显示你的局域网IP，不是localhost
```

### Q: WebSocket连接失败？

检查 `.env.local` 中的 WebSocket 地址：
```env
NEXT_PUBLIC_WS_URL=ws://192.168.1.100:8080
```

### Q: 如何知道前端是否正确监听？

启动前端后，日志应该显示：
```
- Local:        http://localhost:3000
- Network:      http://192.168.1.100:3000
```

## 生产部署

生产环境建议使用域名：

```env
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_WS_URL=wss://api.yourdomain.com
```

## 安全提示

⚠️ 局域网访问仅适用于：
- 开发测试
- 内网演示
- 同网络设备

❌ 不要：
- 暴露到公网
- 在不安全的WiFi使用
- 处理真实资金

## 快速测试

在电脑浏览器输入：
```
http://192.168.1.100:3000
```

如果能访问，说明配置成功！然后在手机浏览器输入同样地址即可。

