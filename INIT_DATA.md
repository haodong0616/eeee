# 数据初始化说明

## 自动初始化（开箱即用）

**无需手动操作！** 首次启动后端时会自动初始化所有数据。

```bash
cd backend
go run main.go
```

系统会自动检测数据库是否为空，如果是首次启动，会自动：
1. ✅ 创建 6 个交易对
2. ✅ 生成 30 天历史交易数据
3. ✅ 生成所有周期的 K 线数据
4. ✅ 初始化手续费配置
5. ✅ 创建虚拟用户用于模拟交易

## 初始化内容

系统首次启动会自动创建：

### 📊 交易对（6个）
- 🌙 LUNAR/USDT (~$8,500)
- ⭐ NOVA/USDT (~$1,250)
- 🔗 NEXUS/USDT (~$125)
- 🌬️ ZEPHYR/USDT (~$45)
- 💓 PULSE/USDT (~$2.85)
- 🔮 ARCANA/USDT (~$0.0085)

### 💱 历史成交记录（智能策略）
系统会随机生成 6-12 个月的交易数据：
- 📅 时间跨度：随机 6-12 个月
- 🎲 交易频率：每个代币 1-5 笔/小时（随机）
- 🌓 时段模式：白天(8-22点)更活跃
- 📈 价格趋势：有周期性波动和趋势
- 💾 数据量：约 10,000-30,000 条交易
- ⚡ 初始化速度：15-30 秒

### 📈 K线数据
为每个交易对生成 6 个常用时间周期的K线：
- 分钟级：1m, 5m, 15m
- 小时级：1h, 4h
- 日级：1d

**注意**：秒级K线（15s, 30s）和3m, 30m会在启动后由K线生成器实时生成

## 验证数据

首次启动后端时，会在日志中看到：

```
🚀 首次启动，开始自动初始化数据（智能模式）...
📊 策略：生成 6-12 个月随机数据，每个代币活跃度随机
📊 创建交易对...
✅ 创建了 6 个交易对
💱 生成 9 个月交易数据（智能模式）...
📊 各代币交易频率：
   LUNAR/USDT: 2 笔/小时
   NOVA/USDT: 3 笔/小时
   ZEPHYR/USDT: 2 笔/小时
   PULSE/USDT: 4 笔/小时
   ARCANA/USDT: 3 笔/小时
   NEXUS/USDT: 2 笔/小时
📝 批量插入 18,542 条交易记录...
   已插入 10000/18542 条...
✅ 生成了 18542 条交易记录 (9 个月)
📈 生成K线数据（智能模式）...
   处理 LUNAR/USDT K线...
   处理 NOVA/USDT K线...
   ...
📝 批量插入 22,148 根K线...
   已插入 10000/22148 根...
✅ 生成了 22148 根K线
🎉 数据初始化完成！耗时: 22.35秒
```

### 访问前端验证

1. **首页** - http://localhost:3000
   - 应该看到 6 个交易对
   - 有价格和涨跌幅

2. **行情页面** - http://localhost:3000/markets
   - 6 个交易对及其详细数据

3. **交易页面** - http://localhost:3000/trade/LUNAR-USDT
   - 完整的 K 线图（30天数据）
   - 最近成交记录
   - 盘口数据（如果启用模拟器）

## 检查数据

### 查看数据库
```bash
cd backend
sqlite3 expchange.db

# 查看交易对
SELECT * FROM trading_pairs;

# 查看成交数量
SELECT symbol, COUNT(*) FROM trades GROUP BY symbol;

# 查看K线数据
SELECT symbol, interval, COUNT(*) FROM klines GROUP BY symbol, interval;

# 退出
.exit
```

### 查看数据统计
```bash
sqlite3 expchange.db "
SELECT 
  'Trading Pairs' as item, COUNT(*) as count FROM trading_pairs
UNION ALL
SELECT 'Trades', COUNT(*) FROM trades
UNION ALL
SELECT 'Klines', COUNT(*) FROM klines;
"
```

## 测试交易

### 1. 连接钱包
访问前端：http://localhost:3000  
点击右上角"连接钱包"

### 2. 充值测试资金
进入"资产"页面，使用充值功能：
- 充值 USDT: 100,000
- 充值 LUNAR: 10

### 3. 开始交易
进入交易页面，体验完整功能！

## 重置数据

如果需要重新开始：

```bash
cd backend
rm expchange.db    # 删除数据库
go run main.go     # 重新启动，会自动重新初始化
```

## 常见问题

### Q: 运行初始化脚本后还是没有行情？
A: 
1. 确保后端正在运行
2. 刷新前端页面（Ctrl+R 或 Cmd+R）
3. 检查浏览器控制台是否有错误
4. 检查后端日志是否有 404 错误

### Q: 价格显示为 0 或 "-"？
A: 需要有成交记录才会显示价格。运行初始化脚本会创建测试成交记录。

### Q: K线图是空的？
A: 需要 K线数据。初始化脚本会生成基础K线数据。

### Q: 能否自动生成实时价格？
A: 可以，但需要有实际的交易。初始化脚本提供的是静态测试数据。要有动态价格，需要：
1. 创建买卖订单
2. 订单撮合成交
3. K线生成器会自动聚合成交数据

## 测试交易流程

1. **初始化数据**：`./scripts/init.sh`
2. **启动后端**：`go run main.go`
3. **启动前端**：`cd frontend && npm run dev`
4. **连接钱包**
5. **充值资金**（资产页面）
6. **进入交易页面**
7. **下单测试**

## 生产环境注意

⚠️ **测试数据仅用于开发环境**

生产环境应该：
- 只创建真实的交易对
- 不创建虚假的成交记录
- 价格来自真实交易
- K线数据由实际成交生成

## 更多数据生成

如果需要更多历史数据，可以使用管理后台的 K线生成功能：

```bash
curl -X POST http://localhost:8080/api/admin/klines/generate \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d "symbol=BTC/USDT" \
  -d "interval=1m" \
  -d "from=2024-01-01T00:00:00Z" \
  -d "to=2024-01-31T23:59:59Z"
```

