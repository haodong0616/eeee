# 数据库说明

## 当前配置

项目使用 **SQLite** 作为数据库，无需额外安装和配置。

### 为什么选择 SQLite？

✅ **零配置** - 无需安装数据库服务器  
✅ **轻量级** - 单文件存储，易于备份  
✅ **快速** - 对于中小型应用性能优秀  
✅ **跨平台** - 在任何系统上都能运行  
✅ **开发友好** - 快速启动和测试  

## 数据库文件

- **位置**: `backend/expchange.db`
- **类型**: SQLite 3.x
- **自动创建**: 首次启动时自动生成
- **自动迁移**: GORM 自动创建表结构

## 数据表结构

### users - 用户表
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    nonce TEXT,
    user_level VARCHAR(20) DEFAULT 'normal',
    created_at DATETIME,
    updated_at DATETIME
);
```

### trading_pairs - 交易对表
```sql
CREATE TABLE trading_pairs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    base_asset VARCHAR(10) NOT NULL,
    quote_asset VARCHAR(10) NOT NULL,
    min_price DECIMAL(20,8),
    max_price DECIMAL(20,8),
    min_qty DECIMAL(20,8),
    max_qty DECIMAL(20,8),
    status VARCHAR(20) DEFAULT 'active',
    created_at DATETIME,
    updated_at DATETIME
);
```

### balances - 余额表
```sql
CREATE TABLE balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    asset VARCHAR(10) NOT NULL,
    available DECIMAL(30,8) DEFAULT 0,
    frozen DECIMAL(30,8) DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX idx_balances_user_id ON balances(user_id);
```

### orders - 订单表
```sql
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    price DECIMAL(20,8),
    quantity DECIMAL(20,8) NOT NULL,
    filled_qty DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);
```

### trades - 成交记录表
```sql
CREATE TABLE trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id INTEGER NOT NULL,
    sell_order_id INTEGER NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    created_at DATETIME
);
CREATE INDEX idx_trades_symbol ON trades(symbol);
```

### klines - K线数据表
```sql
CREATE TABLE klines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open_time INTEGER NOT NULL,
    open DECIMAL(20,8),
    high DECIMAL(20,8),
    low DECIMAL(20,8),
    close DECIMAL(20,8),
    volume DECIMAL(20,8),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX idx_klines_symbol ON klines(symbol);
CREATE INDEX idx_klines_open_time ON klines(open_time);
```

### fee_configs - 手续费配置表
```sql
CREATE TABLE fee_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_level VARCHAR(20) NOT NULL,
    maker_fee_rate DECIMAL(10,6) NOT NULL,
    taker_fee_rate DECIMAL(10,6) NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
```

### fee_records - 手续费记录表
```sql
CREATE TABLE fee_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    trade_id INTEGER NOT NULL,
    asset VARCHAR(10) NOT NULL,
    amount DECIMAL(30,8) NOT NULL,
    fee_rate DECIMAL(10,6) NOT NULL,
    order_side VARCHAR(10) NOT NULL,
    created_at DATETIME
);
CREATE INDEX idx_fee_records_user_id ON fee_records(user_id);
CREATE INDEX idx_fee_records_order_id ON fee_records(order_id);
CREATE INDEX idx_fee_records_trade_id ON fee_records(trade_id);
```

## 数据库操作

### 查看数据库
```bash
cd backend
sqlite3 expchange.db
```

### 常用SQL命令
```sql
-- 查看所有表
.tables

-- 查看表结构
.schema users

-- 查询用户
SELECT * FROM users;

-- 查询交易对
SELECT * FROM trading_pairs;

-- 查询订单
SELECT * FROM orders ORDER BY created_at DESC LIMIT 10;

-- 查询成交记录
SELECT * FROM trades ORDER BY created_at DESC LIMIT 10;

-- 退出
.exit
```

### 备份数据库
```bash
# 简单备份
cp expchange.db expchange_backup.db

# 带时间戳的备份
cp expchange.db expchange_backup_$(date +%Y%m%d_%H%M%S).db

# 导出SQL
sqlite3 expchange.db .dump > backup.sql
```

### 恢复数据库
```bash
# 从备份恢复
cp expchange_backup.db expchange.db

# 从SQL恢复
sqlite3 expchange.db < backup.sql
```

### 清空数据库
```bash
# 删除数据库文件重新开始
rm expchange.db

# 或者清空所有数据
sqlite3 expchange.db "DELETE FROM users;"
sqlite3 expchange.db "DELETE FROM orders;"
sqlite3 expchange.db "DELETE FROM trades;"
# ... 等等
```

## Redis（可选）

Redis目前**已配置但未实际使用**，可以完全不启动Redis也能正常运行。

未来可能用于：
- 盘口数据缓存
- 会话管理
- 实时数据缓存
- 消息队列

如果不需要Redis功能，可以忽略Redis相关的错误提示。

## 性能优化建议

### SQLite配置优化
在生产环境可以考虑以下优化：

```go
// 在 database.go 中添加
DB.Exec("PRAGMA journal_mode=WAL;")  // 写前日志模式
DB.Exec("PRAGMA synchronous=NORMAL;") // 同步模式
DB.Exec("PRAGMA cache_size=-64000;")  // 缓存大小64MB
DB.Exec("PRAGMA temp_store=MEMORY;")  // 临时表存在内存
```

### 索引优化
确保关键查询字段都有索引：
- `orders(user_id, symbol, status)`
- `trades(symbol, created_at)`
- `balances(user_id, asset)`
- `klines(symbol, interval, open_time)`

## 迁移到其他数据库

如果需要迁移到PostgreSQL或MySQL：

1. 修改 `go.mod` 依赖
2. 修改 `database/database.go` 的驱动
3. 调整 `config/config.go` 的配置
4. 数据迁移（导出/导入）

SQLite适合：
- 开发和测试环境
- 中小型应用（< 10万用户）
- 单服务器部署
- 每秒 < 1000 次写入

如需更高性能，建议迁移到PostgreSQL。

## 常见问题

**Q: 数据库文件在哪里？**  
A: `backend/expchange.db`

**Q: 如何重置数据库？**  
A: 删除 `expchange.db` 文件，重启服务会自动创建新的。

**Q: SQLite性能够用吗？**  
A: 对于开发和中小型应用完全够用，支持数万并发读取。

**Q: 需要安装SQLite吗？**  
A: 不需要，Go的SQLite驱动是纯Go实现的CGO版本。

**Q: 可以在生产环境使用吗？**  
A: 可以，但建议日交易量大时迁移到PostgreSQL。

**Q: 如何查看数据库大小？**  
A: `ls -lh backend/expchange.db`

**Q: 并发安全吗？**  
A: 是的，SQLite支持多读单写，GORM会处理连接池。

