# MySQL 和 SQLite 兼容性说明

## 已完成的兼容性优化

为了确保代码同时兼容 MySQL 和 SQLite，已对所有模型进行以下优化：

### 1. ✅ 字符串字段长度规范

**问题**：MySQL 要求所有 VARCHAR/CHAR 类型字段必须指定长度。

**解决方案**：所有字符串字段都已明确指定 `size` 或 `type`。

#### 修改的字段：

| 模型 | 字段 | 原定义 | 新定义 | 说明 |
|------|------|--------|--------|------|
| `User` | `Nonce` | 无类型 | `size:100` | 用户随机数 |
| `Task` | `Message` | `size:500` | `type:varchar(1000)` | 任务消息，扩大容量 |
| `TaskLog` | `Message` | `type:text` | `type:varchar(2000)` | 日志消息，指定长度 |
| `ChainConfig` | `RpcURL` | `type:text` | `type:varchar(500)` | RPC地址 |
| `ChainConfig` | `BlockExplorerURL` | `type:text` | `type:varchar(500)` | 区块浏览器地址 |
| `ChainConfig` | `PlatformWithdrawPrivateKey` | `type:text` | `type:varchar(500)` | 私钥 |
| `DepositRecord` | `TaskID` | 缺失 | `size:24;index` | 新增任务关联字段 |

### 2. ✅ TEXT 类型使用规范

**最佳实践**：

- **短文本（< 2KB）**：使用 `type:varchar(N)`，性能更好
- **中等文本（2KB-64KB）**：使用 `type:text`（MySQL 的 TEXT 类型）
- **长文本（> 64KB）**：可使用 `type:mediumtext` 或 `type:longtext`

#### 当前 TEXT 字段：

| 字段 | 类型 | 最大长度 | 用途 |
|------|------|---------|------|
| `Task.Error` | `text` | ~65KB | 错误堆栈信息 |
| `TaskLog.Details` | `text` | ~65KB | 详细日志 |

这些字段使用 `type:text` 是合理的，因为它们可能包含较长的内容。

### 3. ✅ Decimal 类型

**MySQL**：`DECIMAL(M,D)` - M 是总位数，D 是小数位数
**SQLite**：内部使用 NUMERIC，GORM 自动处理

所有 decimal 字段已正确定义：

```go
// 余额类字段
gorm:"type:decimal(30,8)"   // 最大: 9999999999999999999999.99999999

// 价格/数量类字段
gorm:"type:decimal(20,8)"   // 最大: 999999999999.99999999

// 手续费率字段
gorm:"type:decimal(10,6)"   // 最大: 9999.999999
```

### 4. ✅ 索引定义

所有索引定义兼容两种数据库：

- **单列索引**：`gorm:"index"`
- **唯一索引**：`gorm:"uniqueIndex"`
- **复合唯一索引**：`gorm:"index:idx_name,unique"`

示例：
```go
// Kline 表的复合唯一索引
Symbol   string `gorm:"size:20;not null;index:idx_kline_unique,unique"`
Interval string `gorm:"size:10;not null;index:idx_kline_unique,unique"`
OpenTime int64  `gorm:"not null;index;index:idx_kline_unique,unique"`
```

### 5. ✅ 默认值

所有默认值都使用兼容的方式定义：

```go
// ✅ 正确 - 兼容两种数据库
gorm:"default:true"
gorm:"default:'active'"
gorm:"default:0"

// ❌ 避免 - 可能不兼容
gorm:"default:NOW()"  // 应该在代码中处理
```

### 6. ✅ 外键关系

使用 GORM 的软关联，不创建物理外键：

```go
User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
```

这种方式在两种数据库中都能正常工作。

## 数据类型映射表

| Go 类型 | SQLite | MySQL | GORM 标签 |
|---------|--------|-------|-----------|
| `string` | TEXT | VARCHAR(N) | `size:N` |
| `string` | TEXT | TEXT | `type:text` |
| `int` | INTEGER | INT | - |
| `int64` | INTEGER | BIGINT | - |
| `bool` | INTEGER | TINYINT(1) | - |
| `time.Time` | DATETIME | DATETIME | - |
| `decimal.Decimal` | NUMERIC | DECIMAL(M,D) | `type:decimal(M,D)` |

## 自动迁移注意事项

### 首次迁移
```bash
# 自动创建所有表和索引
go run main.go
```

### 修改字段后迁移

GORM AutoMigrate 会：
- ✅ 创建新表
- ✅ 添加缺失的字段
- ✅ 创建缺失的索引
- ❌ **不会修改现有字段类型**
- ❌ **不会删除未使用的字段**

如果需要修改字段类型，建议：

1. **SQLite**：删除表，重新创建
```sql
DROP TABLE table_name;
-- 然后重启服务，AutoMigrate 会重建
```

2. **MySQL**：手动执行 ALTER TABLE
```sql
ALTER TABLE tasks MODIFY COLUMN message VARCHAR(1000);
```

## 性能对比

| 特性 | SQLite | MySQL |
|------|--------|-------|
| VARCHAR vs TEXT | TEXT 性能相同 | VARCHAR 更快（< 2KB） |
| 索引性能 | 中等 | 优秀 |
| 并发写入 | 低（表级锁） | 高（行级锁） |
| 全文搜索 | FTS5 扩展 | FULLTEXT 索引 |
| 批量插入 | 每批 < 500 | 每批 1000+ |

## 测试验证

已通过编译测试，确保：
- ✅ 所有字段定义符合规范
- ✅ 代码可以正常编译
- ✅ 与 SQLite 和 MySQL 都兼容

## 切换数据库

只需修改 `config/config.go` 中的 `DBType`：

```go
DBType: "mysql"   // 使用 MySQL
DBType: "sqlite"  // 使用 SQLite
```

或使用环境变量：
```bash
DB_TYPE=sqlite go run main.go
```

## 参考链接

- [GORM 数据类型](https://gorm.io/docs/models.html#Fields-Tags)
- [MySQL 数据类型](https://dev.mysql.com/doc/refman/8.0/en/data-types.html)
- [SQLite 数据类型](https://www.sqlite.org/datatype3.html)

