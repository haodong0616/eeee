# 🔧 API 接口修复说明

## 问题描述

手机端登录失败的原因：
1. ✅ Next.js 14 的 `params` 需要 `await`（已修复）
2. ⚠️ 环境变量配置可能不正确

## 🚀 解决方案

### 1. 创建环境变量文件

在 `frontend` 目录创建 `.env.local` 文件：

```bash
cd frontend
touch .env.local
```

内容如下：

```env
# 后端 API 地址（服务端使用）
# 注意：必须包含 http:// 前缀
BACKEND_URL=http://localhost:8080

# WalletConnect Project ID（可选）
NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID=YOUR_PROJECT_ID_HERE
```

### 2. 重启前端服务

```bash
# 停止当前服务（Ctrl+C）
# 重新启动
npm run dev
```

## 📝 修复内容

### `/app/api/[...path]/route.ts`

✅ 已修复 `params` await 问题：

```typescript
// 修复前
{ params }: { params: { path: string[] } }
const path = params.path.join('/');

// 修复后
context: { params: Promise<{ path: string[] }> }
const params = await context.params;
const path = params.path.join('/');
```

✅ 添加详细日志，便于调试：

```typescript
console.log('🟢 POST API Proxy:', url);
console.log('📦 Body:', body.slice(0, 200));
console.log('✅ POST Response:', response.status);
```

## 🧪 测试

### 1. 检查后端是否运行

```bash
curl http://localhost:8080/api/market/tickers
```

应该返回交易对列表。

### 2. 测试前端 API 代理

打开浏览器控制台（F12），访问：
```
http://localhost:3000/api/market/tickers
```

应该返回相同的数据。

### 3. 测试登录流程

1. 打开 `http://localhost:3000`
2. 点击"连接钱包"
3. 选择钱包并连接
4. 查看浏览器控制台日志

应该看到：
```
🔐 开始登录流程，地址: 0x...
📡 获取 nonce...
✅ 获取 nonce 成功: abc123...
✍️ 请求签名，消息: ...
✅ 签名成功: 0x1234...
🔑 提交登录请求...
✅ 登录成功！
```

### 4. 测试手机端

在手机浏览器访问：
```
http://你的局域网IP:3000
```

例如：`http://192.168.1.100:3000`

## 🔍 查看日志

### 前端日志
- 浏览器控制台（F12 → Console）
- 查看前端终端输出

### 后端日志
- 查看后端终端输出
- 应该看到：
  ```
  [GIN] POST   /api/auth/nonce
  [GIN] POST   /api/auth/login
  ```

## ⚠️ 常见问题

### Q1: 环境变量不生效？

**A**: 修改 `.env.local` 后必须重启前端服务！

### Q2: 手机端还是连接失败？

**A**: 检查：
1. 后端是否在运行（`go run main.go`）
2. 前端是否在运行（`npm run dev`）
3. 手机和电脑是否在同一局域网
4. 打开手机浏览器的开发者工具查看错误

### Q3: 看到 "Failed to fetch from backend"？

**A**: 这说明 Next.js API 代理无法连接到后端，检查：
1. `BACKEND_URL` 是否正确（必须包含 `http://`）
2. 后端是否运行在 `8080` 端口
3. 查看前端终端的错误日志

### Q4: 签名成功但登录失败？

**A**: 查看后端日志，可能是：
1. 数据库连接问题
2. JWT Secret 配置问题
3. 钱包地址格式问题

## 📊 API 请求流程

```
手机浏览器
    ↓
http://192.168.1.100:3000/api/auth/login
    ↓
Next.js API Route (/app/api/[...path]/route.ts)
    ↓
http://localhost:8080/api/auth/login
    ↓
后端 (Golang)
    ↓
返回 { token, user }
```

## ✅ 验证修复

如果看到以下输出，说明修复成功：

### 前端终端
```
🟢 POST API Proxy: http://localhost:8080/api/auth/nonce
📦 Body: {"wallet_address":"0x..."}
✅ POST Response: 200

🟢 POST API Proxy: http://localhost:8080/api/auth/login
📦 Body: {"wallet_address":"0x...","signature":"0x..."}
✅ POST Response: 200
```

### 浏览器控制台
```
🔐 开始登录流程，地址: 0x...
✅ 获取 nonce 成功: abc123
✅ 签名成功: 0x1234...
✅ 登录成功！
```

现在应该可以正常登录了！🎉
