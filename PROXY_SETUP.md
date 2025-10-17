# Next.js 代理配置说明

## 架构设计

```
手机浏览器
    ↓
http://192.168.1.100:3000
    ↓
┌─────────────────────────┐
│  Next.js 服务器 (3000)   │
│  ├─ API Routes 代理      │ → http://localhost:8080/api/*
│  └─ WebSocket 代理       │ → ws://localhost:8080/ws
└─────────────────────────┘
    ↓
后端服务器 (localhost:8080)
```

## 文件说明

### 1. `server.js` - 自定义服务器
- 提供 Next.js 页面服务
- 代理 WebSocket 连接
- 监听 `0.0.0.0:3000`

### 2. `app/api/[...path]/route.ts` - API 代理
- 代理所有 `/api/*` 请求到后端
- 支持 GET, POST, PUT, DELETE
- 自动转发 Authorization 头

### 3. `lib/services/api.ts` - API 配置
```typescript
const API_URL = '';  // 空字符串 = 使用相对路径
// 请求会发送到: /api/market/tickers
// Next.js 代理到: http://localhost:8080/api/market/tickers
```

### 4. `lib/websocket.ts` - WebSocket 配置
```typescript
// WebSocket 连接到 Next.js 服务器
ws://192.168.1.100:3000/ws
// Next.js 转发到后端
ws://localhost:8080/ws
```

## 启动方式

### 开发环境

```bash
# 1. 启动后端
cd backend
go run main.go

# 2. 安装 ws 依赖（首次）
cd frontend
npm install ws

# 3. 启动前端（使用自定义服务器）
npm run dev
```

### 生产环境

```bash
cd frontend
npm run build
npm start
```

## 环境变量

创建 `frontend/.env.local`（可选）：

```env
# 后端地址（服务端使用）
BACKEND_URL=localhost:8080

# 如果后端在其他服务器
# BACKEND_URL=192.168.1.200:8080
```

## 优势

1. **零配置**：手机访问时完全自动
2. **统一入口**：所有请求都通过前端
3. **避免跨域**：浏览器只访问一个域名
4. **WebSocket 支持**：实时数据正常工作
5. **安全**：后端不对外暴露

## 测试

### 1. 电脑测试
```bash
curl http://localhost:3000/api/market/tickers
```

### 2. 手机测试
打开浏览器访问：
```
http://192.168.1.100:3000
```

### 3. WebSocket 测试
在浏览器控制台：
```javascript
const ws = new WebSocket('ws://192.168.1.100:3000/ws');
ws.onopen = () => console.log('✅ WebSocket 连接成功');
ws.onmessage = (e) => console.log('📨', e.data);
```

## 故障排查

### 问题1: Module not found: ws

**解决**：
```bash
cd frontend
npm install ws
npm install -D @types/ws
```

### 问题2: WebSocket 连接失败

**检查**：
1. 后端是否运行？
2. 后端端口是否是 8080？
3. 浏览器控制台有什么错误？

### 问题3: API 请求 404

**检查**：
1. `app/api/[...path]/route.ts` 文件是否存在？
2. 重启前端服务

## npm scripts

- `npm run dev` - 启动自定义服务器（支持代理）
- `npm run dev:next` - 原生 Next.js 开发模式（无代理）
- `npm start` - 生产模式自定义服务器
- `npm run start:next` - 原生 Next.js 启动

推荐使用 `npm run dev` 获得完整功能！


