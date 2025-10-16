# MySQL 数据库配置说明

## 当前配置（使用共享 MySQL）

本项目使用与其他项目共享的 Docker MySQL 容器。

### MySQL 连接信息（已写死在代码中）

```go
// backend/config/config.go
DBType:     "mysql"
DBHost:     "localhost"
DBPort:     "3306"
DBUser:     "referral_user"
DBPassword: "referral123456"
DBName:     "expchange"
```

**数据库配置已直接写在 `config/config.go` 中，无需配置环境变量。**

### 可选环境变量配置

如需修改其他配置，在 `backend` 目录下创建 `.env` 文件：

```bash
# 服务器配置
SERVER_PORT=8080

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT 密钥
JWT_SECRET=your-secret-key-here-please-change-in-production

# CORS 允许的来源
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

### SQLite 配置（可选，用于开发测试）

```bash
# 服务器配置
SERVER_PORT=8080

# 数据库配置
DB_TYPE=sqlite
DB_NAME=expchange.db

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT 密钥
JWT_SECRET=your-secret-key-here-please-change-in-production

# CORS 允许的来源
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

## MySQL 数据库初始化

### 共享 MySQL 已自动初始化

`expchange` 数据库已在共享 MySQL 的初始化脚本中创建：

```bash
/Users/apple/work/haodong/youqianhua/back-end/mysql-docker/mysql-init.sql
```

数据库和用户权限已自动配置：
- 数据库名：`expchange`
- 用户名：`referral_user`
- 密码：`referral123456`
- 权限：对 `expchange` 数据库的完全访问权限

### 启动服务

首次启动时，GORM 会自动创建所有需要的表结构：

```bash
cd backend
go run main.go
```

## 共享 MySQL 容器信息

本项目使用共享的 Docker MySQL 容器：

- **位置**：`/Users/apple/work/haodong/youqianhua/back-end/mysql-docker/`
- **容器管理**：通过该目录的 `docker-compose.yml` 管理
- **共享数据库**：
  - `referral_system` - 推荐系统
  - `wallet_db` - 钱包生成器
  - `trading_bot` - 交易机器人
  - `expchange` - 交易所（本项目）

## 数据迁移说明

如果需要从 SQLite 迁移到 MySQL：

1. 导出 SQLite 数据
2. 使用工具转换为 MySQL 格式（如 `sqlite3-to-mysql`）
3. 或者重新执行数据初始化脚本

## 性能优化建议

1. **连接池配置**：MySQL 支持更大的连接池，可以提升并发性能
2. **索引优化**：GORM 会自动创建必要的索引
3. **批量操作**：已优化批量插入，每批 500 条记录
4. **事务支持**：MySQL 的 InnoDB 引擎提供更好的事务性能

## 注意事项

- MySQL 默认端口：3306
- 确保 MySQL 服务已启动
- 建议使用 MySQL 8.0 或以上版本
- 生产环境务必修改默认密码和 JWT 密钥

