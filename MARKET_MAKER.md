# 🤖 智能做市商系统

## 🎯 功能概述

虚拟用户（系统模拟器）现在不仅会挂单提供流动性，还会**智能地吃掉真实用户的订单**，扮演做市商角色，并记录每笔交易的盈亏。

## 🔄 工作原理

### 双重功能

```
1️⃣ 挂单提供流动性（原有功能）
   └─> 每15秒在市价周围挂买卖单（±0.1%-3%）

2️⃣ 智能做市吃单（新增功能）
   └─> 每20秒检查真实用户订单，慢慢吃掉
```

### 做市流程

```
┌─────────────────────────────────────────┐
│  1. 扫描真实用户的挂单                   │
│     └─> 排除虚拟用户自己的订单          │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  2. 随机选择一个订单（30%概率）         │
│     └─> 70%概率不操作（避免太快）       │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  3. 价格合理性检查                       │
│     └─> 偏离市价>5%不吃（避免亏损）     │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  4. 创建对手单                           │
│     ├─> 用户买单 → 虚拟用户卖          │
│     └─> 用户卖单 → 虚拟用户买          │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  5. 吃掉订单（50%-100%随机）            │
│     └─> 撮合引擎自动成交                │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  6. 计算盈亏并记录                       │
│     ├─> 执行价 vs 市价                  │
│     ├─> 计算盈亏金额（USDT）            │
│     └─> 保存到 market_maker_pnl 表      │
└─────────────────────────────────────────┘
```

## 💰 盈亏计算逻辑

### 买入盈亏

```
虚拟用户买入代币：
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
执行价格: $100
当前市价: $105
数量: 10 个
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
盈亏 = (市价 - 执行价) × 数量
     = ($105 - $100) × 10
     = $50 ✅ 盈利

盈亏% = (市价 - 执行价) / 执行价 × 100
      = ($105 - $100) / $100 × 100
      = 5% ✅
```

### 卖出盈亏

```
虚拟用户卖出代币：
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
执行价格: $105
当前市价: $100
数量: 10 个
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
盈亏 = (执行价 - 市价) × 数量
     = ($105 - $100) × 10
     = $50 ✅ 盈利

盈亏% = (执行价 - 市价) / 市价 × 100
      = ($105 - $100) / $100 × 100
      = 5% ✅
```

## 📊 数据库模型

### MarketMakerPnL 表结构

```sql
CREATE TABLE market_maker_pnls (
    id VARCHAR(24) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,           -- 交易对
    trade_id VARCHAR(24),                  -- 关联的交易ID（可选）
    side VARCHAR(10) NOT NULL,             -- buy/sell
    execute_price DECIMAL(20,8) NOT NULL,  -- 执行价格
    market_price DECIMAL(20,8) NOT NULL,   -- 当时市价
    quantity DECIMAL(20,8) NOT NULL,       -- 数量
    profit_loss DECIMAL(20,8),             -- 盈亏（USDT）
    profit_percent DECIMAL(10,4),          -- 盈亏百分比
    created_at TIMESTAMP
);

CREATE INDEX idx_mm_pnl_symbol ON market_maker_pnls(symbol);
CREATE INDEX idx_mm_pnl_created ON market_maker_pnls(created_at);
```

## 🎮 配置参数

### 可调节的参数

在 `simulator/dynamic_orderbook.go` 中：

```go
// 做市频率
ticker := time.NewTicker(20 * time.Second)  // 每20秒检查一次

// 执行概率
if rand.Float64() > 0.3 {  // 30%概率执行
    return
}

// 吃单数量
eatRatio := 0.5 + rand.Float64()*0.5  // 吃掉50%-100%

// 价格偏离阈值
if priceDeviation.GreaterThan(decimal.NewFromFloat(0.05)) {  // 5%
    return
}

// 每次查询订单数
Limit(5)  // 最多看5个订单
```

## 📡 管理后台 API

### 1. 获取盈亏记录

```
GET /api/admin/market-maker/pnl?symbol=BTC/USDT
```

**响应**：
```json
{
  "records": [
    {
      "id": "xxx",
      "symbol": "BTC/USDT",
      "side": "buy",
      "execute_price": "42000.00",
      "market_price": "42100.00",
      "quantity": "0.1",
      "profit_loss": "10.00",
      "profit_percent": "0.24",
      "created_at": "2025-10-17T10:00:00Z"
    }
  ],
  "total_pnl": "1234.56",
  "count": 100
}
```

### 2. 获取盈亏统计

```
GET /api/admin/market-maker/stats
```

**响应**：
```json
{
  "total_pnl": "5678.90",
  "total_trades": 250,
  "profit_trades": 180,
  "loss_trades": 70,
  "win_rate": 72.0,
  "by_symbol": [
    {
      "symbol": "BTC/USDT",
      "total_pnl": "2345.67",
      "trade_count": 80
    },
    {
      "symbol": "ETH/USDT",
      "total_pnl": "1234.56",
      "trade_count": 60
    }
  ]
}
```

## 📈 日志示例

启动后会看到类似日志：

```
🤖 做市商吃单: BTC/USDT buy 0.0500 @ 42000.00000000 (市价: 42100.00000000, 📈 盈利: 5.00 USDT, 0.24%)

🤖 做市商吃单: ETH/USDT sell 1.2000 @ 2850.50000000 (市价: 2840.00000000, 📈 盈利: 12.60 USDT, 0.37%)

🤖 做市商吃单: PULSE/USDT buy 100.0000 @ 2.78000000 (市价: 2.80000000, 📈 盈利: 2.00 USDT, 0.72%)

🤖 做市商吃单: LUNAR/USDT sell 5.0000 @ 4520.00000000 (市价: 4500.00000000, 📈 盈利: 100.00 USDT, 0.44%)
```

## 🎯 优势特性

### 对真实用户友好

1. **价格优先** - 以真实用户的挂单价格成交
2. **慢速吃单** - 每20秒才检查一次，30%概率执行
3. **部分成交** - 随机吃掉50%-100%，避免一次吃完
4. **价格保护** - 偏离市价>5%不吃，保护用户

### 做市商智能

1. **优先吃老单** - FIFO 原则，先吃最早的订单
2. **风险控制** - 价格偏离检查
3. **盈亏记录** - 每笔交易都记录
4. **统计分析** - 可查看总盈亏、胜率等

## 🔧 启用/停用

在管理后台的**交易对管理**页面：

1. 找到要启用的交易对
2. 勾选"启用模拟器"
3. 自动开始做市

**或通过 API**：
```bash
curl -X PUT http://localhost:8080/api/admin/pairs/:id/simulator \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{"simulator_enabled": true}'
```

## 📊 监控建议

### 查看实时日志

```bash
tail -f /root/go/log/exchange_access.log | grep "做市商"
```

### 查看盈亏统计

访问管理后台（需要先实现前端页面）：
```
http://localhost:3001/dashboard/market-maker
```

或使用 API：
```bash
curl http://localhost:8080/api/admin/market-maker/stats \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## ⚙️ 高级配置

### 调整做市激进度

**更激进**（快速吃单）:
```go
ticker := time.NewTicker(10 * time.Second)  // 10秒检查一次
if rand.Float64() > 0.5 {  // 50%概率执行
```

**更保守**（慢速吃单）:
```go
ticker := time.NewTicker(60 * time.Second)  // 60秒检查一次
if rand.Float64() > 0.1 {  // 10%概率执行
```

### 调整风险容忍度

**更保守**：
```go
if priceDeviation.GreaterThan(decimal.NewFromFloat(0.02)) {  // 2%
```

**更激进**：
```go
if priceDeviation.GreaterThan(decimal.NewFromFloat(0.10)) {  // 10%
```

## ✨ 总结

✅ **真实用户** - 订单能被快速成交，获得流动性
✅ **虚拟用户** - 扮演做市商角色，赚取价差
✅ **盈亏透明** - 每笔交易都有详细记录
✅ **风险可控** - 价格偏离保护，避免大额亏损
✅ **灵活配置** - 可按交易对开关

这是一个完整的自动做市商（AMM）系统！🚀

