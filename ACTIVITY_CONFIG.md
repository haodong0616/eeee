# 🎮 交易对活跃度配置

## 🎯 功能说明

现在可以为**每个交易对单独设置**活跃度参数，让不同的代币有不同的市场表现！

## 📊 配置参数

### 1. **活跃度等级** (ActivityLevel)

**范围**: 1-10  
**默认**: 5

| 等级 | 说明 | 订单簿更新 | 效果 |
|------|------|-----------|------|
| 1-2 | 🐢 低活跃 | 20-22秒 | 适合稳定币、小币种 |
| 3-5 | 🚶 中活跃 | 14-18秒 | 适合主流币 |
| 6-8 | 🏃 高活跃 | 8-12秒 | 适合热门币 |
| 9-10 | 🚀 极活跃 | 4-6秒 | 适合明星币、爆款 |

**影响**：
- 订单簿刷新频率
- 价格分布范围
- 数量波动幅度
- 整体市场活跃感

### 2. **订单簿深度** (OrderbookDepth)

**范围**: 5-30  
**默认**: 15

```
5档:  适合低流动性币种
10档: 适合小市值币
15档: 适合中等币种（默认）
20档: 适合大市值币
30档: 适合顶级币种（BTC/ETH）
```

**示例**：
```
深度=5:  买5档 + 卖5档 = 10个订单
深度=15: 买15档 + 卖15档 = 30个订单
深度=30: 买30档 + 卖30档 = 60个订单
```

### 3. **成交频率** (TradeFrequency)

**范围**: 5-60 秒  
**默认**: 20 秒

| 频率 | 说明 | 适用场景 |
|------|------|----------|
| 5-10秒 | ⚡ 极快 | 高频交易币 |
| 15-25秒 | 🔥 快速 | 活跃交易币 |
| 30-45秒 | 📊 正常 | 普通币种 |
| 50-60秒 | 🐌 慢速 | 冷门币种 |

**实际间隔**: 设定值 ±30% 随机波动  
例如: 20秒 → 实际14-26秒

### 4. **价格波动率** (PriceVolatility)

**范围**: 0.001-0.05 (0.1%-5%)  
**默认**: 0.01 (1%)

```
0.001 (0.1%): 稳定币（USDT等）
0.005 (0.5%): 低波动（主流大币）
0.01  (1.0%): 中等波动（默认）
0.02  (2.0%): 高波动（中小币）
0.05  (5.0%): 极高波动（山寨币）
```

**影响**：
- 订单簿价格分布范围
- 市价成交的价格偏离
- 整体价格波动幅度

## 🎨 配置示例

### 示例 1：稳定的大盘币（BTC/USDT）

```sql
UPDATE trading_pairs SET
  activity_level = 8,        -- 高活跃度
  orderbook_depth = 25,      -- 深度订单簿
  trade_frequency = 12,      -- 12秒一次成交
  price_volatility = 0.008   -- 0.8% 波动
WHERE symbol = 'BTC/USDT';
```

**效果**：
- 订单簿每8秒更新（活跃度8）
- 25档买卖盘（深度大）
- 每8-16秒一次成交（快）
- 价格波动小（0.8%）

### 示例 2：活跃的热门币（PULSE/USDT）

```sql
UPDATE trading_pairs SET
  activity_level = 10,       -- 极活跃
  orderbook_depth = 20,      -- 20档
  trade_frequency = 8,       -- 8秒一次成交
  price_volatility = 0.015   -- 1.5% 波动
WHERE symbol = 'PULSE/USDT';
```

**效果**：
- 订单簿每4秒更新（极活跃）
- 20档买卖盘
- 每5-10秒一次成交（非常快）
- 价格波动较大（1.5%）

### 示例 3：冷门小币（ARCANA/USDT）

```sql
UPDATE trading_pairs SET
  activity_level = 2,        -- 低活跃
  orderbook_depth = 8,       -- 8档
  trade_frequency = 45,      -- 45秒一次成交
  price_volatility = 0.03    -- 3% 波动
WHERE symbol = 'ARCANA/USDT';
```

**效果**：
- 订单簿每20秒更新（慢）
- 8档买卖盘（浅）
- 每31-58秒一次成交（慢）
- 价格波动大（3%）

## 🔄 参数关系

### 活跃度 → 订单簿更新间隔

```
ActivityLevel = 1  → 22秒更新一次
ActivityLevel = 5  → 14秒更新一次
ActivityLevel = 10 → 4秒更新一次

公式: 24 - (ActivityLevel × 2)
```

### 活跃度 → 价格分布范围

```
ActivityLevel = 1  → 最大价差 0.5% × Volatility
ActivityLevel = 5  → 最大价差 2.5% × Volatility
ActivityLevel = 10 → 最大价差 5.0% × Volatility

公式: Volatility × ActivityLevel × 0.5
```

### 活跃度 → 数量波动

```
ActivityLevel = 1  → 波动范围 74%-126%
ActivityLevel = 5  → 波动范围 50%-150%
ActivityLevel = 10 → 波动范围 20%-180%

公式: VolatilityFactor = 0.2 + (ActivityLevel × 0.06)
```

### 活跃度 → 成交量

```
ActivityLevel = 1  → 基础量的 0.47-0.94倍
ActivityLevel = 5  → 基础量的 1.15-2.30倍
ActivityLevel = 10 → 基础量的 2.00-4.00倍

公式: VolumeFactor = 0.3 + (ActivityLevel × 0.17)
```

## ⚙️ 在管理后台配置

### 方法 1：直接编辑数据库

```sql
UPDATE trading_pairs SET
  activity_level = 8,
  orderbook_depth = 20,
  trade_frequency = 15,
  price_volatility = 0.012
WHERE symbol = 'PULSE/USDT';
```

### 方法 2：通过管理后台（需要先添加UI）

访问：`http://localhost:3001/dashboard/pairs`

在交易对编辑界面添加：
- 活跃度滑块（1-10）
- 订单簿深度输入框（5-30）
- 成交频率输入框（5-60秒）
- 价格波动率输入框（0.001-0.05）

### 方法 3：通过 API

```bash
curl -X PUT http://localhost:8080/api/admin/pairs/:id \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "activity_level": 8,
    "orderbook_depth": 20,
    "trade_frequency": 15,
    "price_volatility": "0.012"
  }'
```

## 🎯 推荐配置方案

### 方案 A：差异化市场

```sql
-- BTC: 稳定大币
UPDATE trading_pairs SET activity_level=8, orderbook_depth=25, 
  trade_frequency=15, price_volatility=0.008 
WHERE symbol='BTC/USDT';

-- ETH: 活跃大币
UPDATE trading_pairs SET activity_level=7, orderbook_depth=20, 
  trade_frequency=18, price_volatility=0.01 
WHERE symbol='ETH/USDT';

-- PULSE: 热门中币
UPDATE trading_pairs SET activity_level=9, orderbook_depth=18, 
  trade_frequency=10, price_volatility=0.015 
WHERE symbol='PULSE/USDT';

-- ARCANA: 冷门小币
UPDATE trading_pairs SET activity_level=3, orderbook_depth=10, 
  trade_frequency=40, price_volatility=0.025 
WHERE symbol='ARCANA/USDT';
```

### 方案 B：全高活跃

```sql
-- 所有币种高活跃度
UPDATE trading_pairs SET 
  activity_level=8, 
  orderbook_depth=20, 
  trade_frequency=12, 
  price_volatility=0.015
WHERE simulator_enabled=true;
```

### 方案 C：默认配置

```sql
-- 恢复默认值
UPDATE trading_pairs SET 
  activity_level=5, 
  orderbook_depth=15, 
  trade_frequency=20, 
  price_volatility=0.01
WHERE simulator_enabled=true;
```

## 📈 实时效果

### 配置前（默认）

```
PULSE/USDT:
- 每14秒更新订单簿
- 15档买卖盘
- 每14-26秒成交一次
- 价格波动 ±2.5%
```

### 配置后（activity_level=10）

```
PULSE/USDT:
- 每4秒更新订单簿  ⚡ 快3.5倍
- 20档买卖盘       📊 深度+33%
- 每5-10秒成交一次  🔥 快2倍
- 价格波动 ±5%      📈 波动+2倍
```

## 🔍 日志观察

启动后端后，观察日志输出：

```
📚 PULSE/USDT 订单簿已更新 (买单x20, 卖单x20, 活跃度:10)
💹 PULSE/USDT 模拟市价成交: sell 215.75 @ 2.809
📚 BTC/USDT 订单簿已更新 (买单x25, 卖单x25, 活跃度:8)
📚 ARCANA/USDT 订单簿已更新 (买单x8, 卖单x8, 活跃度:2)
```

## 🚀 立即应用

修改后**无需重启**！模拟器每次循环都会重新读取配置：
1. 修改数据库配置
2. 等待最多20秒
3. 新配置自动生效

## ✨ 总结

✅ **每个交易对独立配置** - 差异化市场表现
✅ **实时热更新** - 无需重启后端
✅ **灵活调节** - 4个维度精细控制
✅ **智能适配** - 自动计算最优参数

现在你可以让 BTC 像顶级交易所一样活跃，同时让小币种保持平稳！🎉

