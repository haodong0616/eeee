# MySQL 保留关键字问题修复

## 🐛 问题描述

在使用 MySQL 数据库时，出现以下错误：

```
Error 1064 (42000): You have an error in your SQL syntax; 
check the manual that corresponds to your MySQL server version 
for the right syntax to use near '= ? ORDER BY open_time DESC LIMIT ?' at line 1
```

生成的 SQL：
```sql
SELECT * FROM `klines` WHERE symbol = 'LUNAR/USDT' AND interval = '1m' ORDER BY open_time DESC LIMIT 100
```

## 🔍 问题原因

**`interval` 是 MySQL 的保留关键字**！

在 MySQL 中，`INTERVAL` 关键字用于日期时间计算，例如：
```sql
SELECT NOW() + INTERVAL 1 DAY;
```

当数据库字段名与保留关键字冲突时，必须使用**反引号** `` ` `` 将其包裹起来。

## ✅ 解决方案

在所有使用 `interval` 字段的 GORM 查询中，将字段名用反引号包裹：

### 修改前（错误）：
```go
database.DB.Where("symbol = ? AND interval = ?", symbol, interval)
```

### 修改后（正确）：
```go
database.DB.Where("symbol = ? AND `interval` = ?", symbol, interval)
```

## 📝 修改的文件

1. **backend/handlers/kline.go** (2处修改)
   - `GetKlines()` 方法 - 第36行
   - `GetKlinesForTradingView()` 方法 - 第65行

2. **backend/kline/generator.go** (2处修改)
   - `generateKline()` 方法 - 第79行
   - `generateKline()` 方法 - 第119行

3. **backend/handlers/market.go** (1处修改)
   - `GetKlines()` 方法 - 第179行

## 📚 MySQL 常见保留关键字

以下是一些常见的 MySQL 保留关键字，在命名字段时应避免使用：

- `interval` - 时间间隔
- `order` - 排序
- `group` - 分组
- `select`, `from`, `where`, `join` - SQL 语句关键字
- `table`, `index`, `key` - 数据库对象
- `status`, `type`, `value` - 有时也可能冲突

**最佳实践**：
- 总是检查字段名是否为保留关键字
- 如果必须使用，在 SQL 中用反引号包裹
- 或者在 GORM 的 struct tag 中使用 `column` 显式指定列名

## 🔗 参考资料

- [MySQL 保留关键字列表](https://dev.mysql.com/doc/refman/8.0/en/keywords.html)
- [GORM 字段标签](https://gorm.io/docs/models.html#Fields-Tags)

## ✨ 验证

修复后重启后端服务，K线数据查询应该正常工作，不再出现 SQL 语法错误。

测试命令：
```bash
cd backend
go run main.go
```

然后访问前端，查看 K线图表是否正常显示。

