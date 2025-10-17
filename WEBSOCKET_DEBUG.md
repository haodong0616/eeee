# 🔍 WebSocket调试指南

## 📋 诊断步骤

### 步骤1：重启后端（必须！）

```bash
cd /Users/hh/work/haodong/expchange/backend
go run main.go

# 等待启动完成，看到：
✅ 动态订单簿模拟器已启动（支持WebSocket实时推送）
🤖 做市商极速模式已启动，嘎嘎快速吃单中...
```

### 步骤2：刷新前端

```bash
# 浏览器访问
http://localhost:3000/trade/PULSE-USDT

# 硬刷新（清除缓存）
Mac: Cmd + Shift + R
Windows: Ctrl + Shift + R
```

### 步骤3：检查Console

```bash
F12 → Console

应该看到：
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WebSocket connected ✅
📡 WebSocket已监听: PULSE/USDT ✅
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

如果没看到 "WebSocket connected"：
❌ WebSocket连接失败
```

### 步骤4：检查Network → WS

```bash
F12 → Network → WS标签

应该看到：
- 一个WebSocket连接（绿色）
- Status: 101 Switching Protocols

点击连接 → Messages标签

应该看到实时消息：
↓ {"type":"orderbook","data":{"symbol":"PULSE/USDT",...}}
↓ {"type":"trade","data":{"symbol":"PULSE/USDT",...}}
...

如果Messages是空的：
❌ WebSocket连接了但没有推送消息
```

### 步骤5：检查后端日志

```bash
# 后端运行的终端

应该看到：
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Client connected, total: 1  ← 前端连接成功
📚 PULSE/USDT 订单簿已更新 (买单x25, 卖单x25, 活跃度:10)
📡 推送orderbook消息（客户端数: 1）  ← 推送了！
🤖 做市商吃单: PULSE/USDT buy 210.0000 @ ...
📡 推送trade消息（客户端数: 1）  ← 推送了！
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

如果看不到 "📡 推送" 日志：
❌ 后端没有推送消息
```

## 🔧 常见问题排查

### 问题1：Console没有 "WebSocket connected"

**原因**：WebSocket连接失败

**检查**：
```bash
# 检查后端是否启动
后端日志是否显示：Server is running on :8383

# 检查前端是否运行
前端是否运行在 http://localhost:3000

# 检查WebSocket URL
F12 → Console 输入：
> localStorage.getItem('ws_url')
应该返回：ws://localhost:3000
```

**解决**：
```bash
# 确保后端和前端都在运行
cd backend && go run main.go
cd frontend && npm run dev
```

### 问题2：Connected但没有消息

**原因**：后端没有推送或前端没有监听

**检查后端**：
```bash
# 后端日志应该有：
📡 推送orderbook消息（客户端数: 1）
📡 推送trade消息（客户端数: 1）

如果没有：
- 检查是否有真实用户订单（做市商需要吃单）
- 检查活跃度是否>0
- 检查模拟器是否启用
```

**检查前端**：
```javascript
// F12 → Console 输入：
wsClient.on('orderbook', (data) => {
  console.log('🎯 手动监听orderbook:', data);
});

wsClient.on('trade', (data) => {
  console.log('🎯 手动监听trade:', data);
});

// 然后等待，应该会有输出
```

### 问题3：收到消息但界面不更新

**原因**：React状态更新问题

**检查**：
```javascript
// F12 → Console 应该看到：
📩 WebSocket收到消息 [orderbook]: {symbol: "PULSE/USDT", ...}
✅ 执行1个处理器

// 如果看到：
⚠️ 没有orderbook类型的处理器
// 说明监听器没注册成功
```

**解决**：
```bash
# 刷新页面（清除状态）
Cmd/Ctrl + Shift + R
```

## 🧪 快速测试脚本

### 前端Console执行

```javascript
// 1. 检查WebSocket状态
console.log('WebSocket状态:', wsClient.ws?.readyState);
// 0=连接中, 1=已连接, 2=关闭中, 3=已关闭

// 2. 手动连接
wsClient.connect();

// 3. 添加临时监听器
wsClient.on('orderbook', (data) => {
  console.log('📊 盘口更新:', data);
});

wsClient.on('trade', (data) => {
  console.log('💱 成交记录:', data);
});

// 4. 等待消息
// 应该每秒看到多条消息输出
```

## 📊 正常的Console输出

```
WebSocket connected
📡 WebSocket已监听: PULSE/USDT
📩 WebSocket收到消息 [orderbook]: {symbol: 'PULSE/USDT', bids: Array(25), asks: Array(25)}
✅ 执行1个处理器
📩 WebSocket收到消息 [trade]: {symbol: 'PULSE/USDT', price: '2.851', quantity: '100', side: 'buy'}
✅ 执行1个处理器
📩 WebSocket收到消息 [orderbook]: {symbol: 'PULSE/USDT', bids: Array(25), asks: Array(25)}
✅ 执行1个处理器
📩 WebSocket收到消息 [trade]: {symbol: 'PULSE/USDT', price: '2.849', quantity: '50', side: 'sell'}
✅ 执行1个处理器
...

每秒多条消息持续输出 = 成功！🎉
```

## ⚡ 快速验证命令

### 1行命令检查WebSocket

```javascript
// F12 → Console 粘贴执行：
setInterval(() => console.log('🔍 监听器数量:', {orderbook: wsClient.handlers?.get?.('orderbook')?.length || 0, trade: wsClient.handlers?.get?.('trade')?.length || 0}), 2000);

// 应该每2秒输出：
🔍 监听器数量: {orderbook: 1, trade: 1}
```

## 🎯 确认清单

做完以下所有步骤：

```
□ 后端重启（最重要！）
  cd backend && go run main.go
  
□ 前端刷新（硬刷新）
  Cmd/Ctrl + Shift + R
  
□ Console看到 "WebSocket connected"
  
□ Console看到 "📡 WebSocket已监听: PULSE/USDT"
  
□ Console持续输出 "📩 WebSocket收到消息"
  
□ Network → WS → Messages 有实时消息流
  
□ 盘口数字实时跳动
  
□ 成交记录实时滚动
```

全部打勾 = WebSocket工作正常！✅

## 💡 最可能的原因

如果看不到推送数据，**99%是因为后端没重启**！

```bash
# 必须重启后端，因为：
1. 新增了WebSocket推送代码
2. 修改了模拟器构造函数（传入wsHub）
3. 旧进程还没有这些代码

解决：
Ctrl+C 停止后端
go run main.go 重新启动

等待看到：
✅ 动态订单簿模拟器已启动（支持WebSocket实时推送）
                           ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑
                           这行说明新代码生效了！
```

重启后端后，刷新前端，应该就能看到满屏的WebSocket消息了！🚀

