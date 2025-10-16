# 订单簿模拟器

## 功能说明

订单簿模拟器会自动在盘口挂买卖单，让交易所的订单簿看起来有深度，像真实的交易所一样。

## 工作原理

### 1. 虚拟用户
- 创建一个虚拟用户（地址：`0x0000...0000`）
- 为虚拟用户充值大量资金
- 所有模拟订单都由这个用户发出

### 2. 订单分布

**买单（Bids）**：
- 在当前价格下方 0.1% - 3% 范围内
- 10个价格档位
- 距离当前价越远，数量越大

**卖单（Asks）**：
- 在当前价格上方 0.1% - 3% 范围内
- 10个价格档位
- 距离当前价越远，数量越大

### 3. 自动维护
- 每10秒检查订单簿
- 每15秒可能刷新部分订单
- 清理旧订单，创建新订单
- 保持盘口深度

## 启用方法

在 `backend/.env` 中设置：
```env
ENABLE_SIMULATOR=true
```

重启后端：
```bash
cd backend
go run main.go
```

你会看到：
```
🎮 启用市场模拟器
📈 趋势模拟器已启动
📚 订单簿模拟器已启动
✅ 创建虚拟用户并充值
📚 BTC/USDT 订单簿已更新 (买单x10, 卖单x10)
📚 ETH/USDT 订单簿已更新 (买单x10, 卖单x10)
```

## 效果展示

### 访问交易页面
http://localhost:3000/trade/BTC-USDT

你会看到：

**买单（绿色）**：
```
价格      数量      总计
64,870   0.05     3,243
64,740   0.08     5,179
64,610   0.12     7,753
64,480   0.15     9,672
...
```

**卖单（红色）**：
```
价格      数量      总计
65,130   0.05     3,256
65,260   0.08     5,220
65,390   0.12     7,846
65,520   0.15     9,828
...
```

## 订单特点

### 1. 价格分布
- **紧密挂单**：当前价 ±0.1% 到 ±0.5%
- **中等挂单**：当前价 ±0.5% 到 ±1.5%
- **远端挂单**：当前价 ±1.5% 到 ±3%

### 2. 数量分布
遵循真实市场规律：
- 靠近当前价：数量较小（容易成交）
- 远离当前价：数量较大（支撑/压力）

### 3. 动态调整
- 价格变化时，订单簿自动调整
- 订单定期刷新，看起来更真实
- 随机取消和创建，模拟真实用户行为

## 两种模式

### 基础模式（OrderBookSimulator）
```go
NewOrderBookSimulator(matchingManager)
```
- 固定间隔挂单
- 均匀分布
- 简单维护

### 高级模式（AdvancedOrderBookSimulator）
```go
NewAdvancedOrderBookSimulator(matchingManager, virtualUserID)
```
- 更自然的价格分布
- 动态数量调整
- 随机刷新订单
- 模拟真实用户行为

**当前使用**：基础模式（已启用）

## 配置调整

### 修改订单数量

编辑 `backend/simulator/orderbook.go`：

```go
// 改变档位数量
levels := 10  // 改为 20 可以有更多档位

// 改变基础数量
case "BTC/USDT":
    baseQty = 0.01 + float64(i)*0.02  // 调整系数
```

### 修改价格范围

```go
// 改变价格偏移范围
priceOffset := (0.005 + float64(i)*0.003)  // 0.5%-3.5%
// 改为
priceOffset := (0.001 + float64(i)*0.002)  // 0.1%-2.1%
```

### 修改刷新频率

```go
ticker := time.NewTicker(10 * time.Second)  // 改为 30秒
```

## 与真实订单的区别

### 识别虚拟订单
```sql
-- 虚拟订单
SELECT * FROM orders WHERE user_id = (
  SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000'
);

-- 真实订单
SELECT * FROM orders WHERE user_id != (
  SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000'
);
```

### 前端显示
虚拟订单和真实订单在前端显示完全一样，用户无法区分。

## 真实交易的影响

### 撮合优先级
真实用户订单和虚拟订单使用相同的撮合规则：
- 价格优先
- 时间优先

如果真实用户下单：
1. 价格更好的订单会优先成交
2. 虚拟订单可能被成交
3. 虚拟用户的余额会相应变化
4. 不影响真实用户的交易

### 示例场景

**当前盘口**：
```
卖: 65,200 USDT (虚拟)
卖: 65,100 USDT (虚拟)
-------------------
买: 64,900 USDT (虚拟)
买: 64,800 USDT (虚拟)
```

**真实用户下单**：
```
市价买单 0.1 BTC
↓
成交价: 65,100 USDT (与虚拟卖单成交)
↓
虚拟用户获得 USDT
真实用户获得 BTC（扣除手续费）
```

## 优势

✅ **盘口有深度** - 用户看到充足的流动性  
✅ **价格发现** - 帮助形成合理价格  
✅ **降低滑点** - 大额订单也能较好成交  
✅ **吸引用户** - 看起来像活跃的交易所  
✅ **真实体验** - 模拟真实市场环境  

## 风险控制

### 1. 资金隔离
虚拟用户的余额独立管理，不影响真实用户。

### 2. 风险对冲
虚拟订单成交后：
- 虚拟用户持有的资产可能增加或减少
- 长期运行需要平衡策略
- 建议定期重置虚拟用户余额

### 3. 监控
```sql
-- 查看虚拟用户余额
SELECT * FROM balances WHERE user_id = (
  SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000'
);

-- 查看虚拟订单数量
SELECT COUNT(*) FROM orders WHERE user_id = (
  SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000'
);
```

## 关闭模拟器

### 方法 1：环境变量
```env
ENABLE_SIMULATOR=false
```

### 方法 2：清理虚拟订单
```sql
-- 取消所有虚拟订单
UPDATE orders 
SET status = 'cancelled' 
WHERE user_id = (
  SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000'
) AND status = 'pending';
```

## 生产环境建议

### 开发/演示环境 ✅
完全可以使用，增强用户体验。

### 生产环境 ⚠️
需要谨慎考虑：
- 做市商策略
- 风险管理
- 监管合规
- 资金安全

建议使用专业的做市商系统，而不是简单的模拟器。

## 升级方向

### 1. 智能做市
- 根据真实订单簿调整
- 动态调整价差
- 保持价格稳定

### 2. 策略优化
- 网格策略
- 均值回归策略
- 趋势跟踪策略

### 3. 风险管理
- 持仓限制
- 价格偏离保护
- 自动平仓

## 监控命令

```bash
# 查看虚拟用户订单数量
sqlite3 backend/expchange.db "
SELECT symbol, side, COUNT(*) as count, SUM(CAST(quantity as REAL)) as total_qty
FROM orders 
WHERE user_id = (SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000')
AND status = 'pending'
GROUP BY symbol, side;
"

# 查看虚拟用户余额
sqlite3 backend/expchange.db "
SELECT asset, available, frozen
FROM balances 
WHERE user_id = (SELECT id FROM users WHERE wallet_address = '0x0000000000000000000000000000000000000000');
"
```

## 总结

订单簿模拟器让你的交易所：
- 🎯 有充足的流动性
- 📊 盘口看起来真实
- 💹 价格更稳定
- 👥 吸引更多用户

现在启用模拟器，你的交易所就有真实的盘口深度了！🚀

