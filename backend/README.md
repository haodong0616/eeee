# ExpChange 后端服务

## 项目结构

```
backend/
├── config/           # 配置管理
├── database/         # 数据库连接
├── handlers/         # HTTP 处理器
├── matching/         # 撮合引擎
├── middleware/       # 中间件
├── models/           # 数据模型
├── websocket/        # WebSocket 服务
├── main.go           # 程序入口
└── go.mod            # 依赖管理
```

## 撮合引擎说明

### 工作原理

1. **订单队列**
   - 买单队列：使用最大堆，价格高的优先
   - 卖单队列：使用最小堆，价格低的优先
   - 时间优先：相同价格按时间排序

2. **撮合流程**
   ```
   新订单 -> 添加到队列 -> 触发撮合
                           ↓
   检查价格匹配 -> 计算成交量 -> 更新订单状态
                           ↓
   生成成交记录 -> 更新余额 -> 推送 WebSocket
   ```

3. **订单状态**
   - `pending`: 待成交
   - `partial`: 部分成交
   - `filled`: 完全成交
   - `cancelled`: 已取消

### 性能优化

- 内存队列实现，毫秒级撮合
- 异步处理成交记录
- Redis 缓存盘口数据
- WebSocket 推送减少轮询

## API 路由

### 公开路由
- `POST /api/auth/nonce` - 获取签名随机数
- `POST /api/auth/login` - 钱包登录
- `GET /api/market/*` - 市场数据

### 认证路由
- `GET /api/profile` - 用户信息
- `POST /api/orders` - 创建订单
- `GET /api/orders` - 查询订单
- `DELETE /api/orders/:id` - 取消订单
- `GET /api/balances` - 查询余额
- `POST /api/balances/deposit` - 充值
- `POST /api/balances/withdraw` - 提现

### 管理路由
- `GET /api/admin/users` - 用户列表
- `GET /api/admin/orders` - 订单列表
- `GET /api/admin/trades` - 成交记录
- `GET /api/admin/stats` - 统计数据
- `POST /api/admin/pairs` - 创建交易对

## 数据库设计

### 主要表结构

**users 表**
```sql
id, wallet_address, nonce, created_at, updated_at
```

**trading_pairs 表**
```sql
id, symbol, base_asset, quote_asset, min_price, max_price, 
min_qty, max_qty, status, created_at, updated_at
```

**orders 表**
```sql
id, user_id, symbol, order_type, side, price, quantity, 
filled_qty, status, created_at, updated_at
```

**trades 表**
```sql
id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
```

**balances 表**
```sql
id, user_id, asset, available, frozen, created_at, updated_at
```

## 环境变量

```env
SERVER_PORT=8080
# SQLite数据库文件路径
DB_NAME=expchange.db
# Redis配置（可选）
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=your-secret-key
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

## 测试

### 创建测试交易对

```bash
curl -X POST http://localhost:8080/api/admin/pairs \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC/USDT",
    "base_asset": "BTC",
    "quote_asset": "USDT",
    "min_price": "0.01",
    "max_price": "1000000",
    "min_qty": "0.0001",
    "max_qty": "10000"
  }'
```

### 模拟充值

```bash
curl -X POST http://localhost:8080/api/balances/deposit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "asset": "USDT",
    "amount": "10000"
  }'
```

### 创建订单

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC/USDT",
    "order_type": "limit",
    "side": "buy",
    "price": "50000",
    "quantity": "0.1"
  }'
```

## 监控和日志

日志输出到标准输出，可以使用以下命令查看：

```bash
go run main.go 2>&1 | tee app.log
```

## 常见问题

**Q: 撮合引擎如何保证顺序？**
A: 使用优先队列和互斥锁，确保撮合操作原子性。

**Q: 如何处理并发订单？**
A: 每个交易对有独立的撮合引擎，使用锁机制保护。

**Q: 余额如何更新？**
A: 下单时冻结资产，成交时更新可用和冻结余额。

**Q: WebSocket 如何推送数据？**
A: Hub 管理所有连接，广播方式推送给所有客户端。

