# MySQL 快速启动指南

## 🚀 使用共享 Docker MySQL（推荐）

本项目使用与其他项目共享的 Docker MySQL 容器。

### 1. 启动共享 MySQL

```bash
# 进入共享 MySQL 目录
cd /Users/apple/work/haodong/youqianhua/back-end/mysql-docker

# 启动 MySQL（如果还没启动）
docker-compose up -d
```

这将启动共享的 MySQL 8.0 容器，已自动创建 `expchange` 数据库。

### 2. MySQL 配置（已写死在代码中）

配置已直接写在 `config/config.go` 中：

```go
DBType:     "mysql"
DBHost:     "localhost"
DBPort:     "3306"
DBUser:     "referral_user"
DBPassword: "referral123456"
DBName:     "expchange"
```

**无需配置 .env 文件中的数据库相关参数！**

### 3. 启动后端服务

```bash
cd backend
go run main.go
```

首次启动时，GORM 会自动在 `expchange` 数据库中创建所有表结构。

### 4. 可选配置

如果需要修改其他配置，可以创建 `.env` 文件：

```bash
SERVER_PORT=8080
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key
CORS_ORIGINS=http://localhost:3000
```

## 📋 方式二：使用本地 MySQL

### 1. 安装 MySQL

```bash
# macOS
brew install mysql@8.0
brew services start mysql@8.0

# Ubuntu/Debian
sudo apt-get install mysql-server
sudo systemctl start mysql
```

### 2. 创建数据库

```bash
mysql -u root -p
```

```sql
CREATE DATABASE expchange CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 配置环境变量

创建 `backend/.env` 文件：

```bash
SERVER_PORT=8080

DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=expchange

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_SECRET=your-secret-key-please-change-in-production
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

### 4. 启动后端服务

```bash
cd backend
go run main.go
```

## 🔄 从 SQLite 切换到 MySQL

如果你之前使用 SQLite，只需修改 `.env` 文件：

```bash
# 从这个
DB_TYPE=sqlite
DB_NAME=expchange.db

# 改为这个
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=expchange
```

然后重启服务，GORM 会自动在 MySQL 中创建表结构。

## 🛠 管理命令

### 查看 MySQL 状态
```bash
docker ps | grep mysql
```

### 连接到 MySQL
```bash
mysql -h localhost -P 3306 -u referral_user -preferral123456 expchange
```

### 备份数据库
```bash
mysqldump -h localhost -P 3306 -u referral_user -preferral123456 expchange > expchange_backup.sql
```

### 恢复数据库
```bash
mysql -h localhost -P 3306 -u referral_user -preferral123456 expchange < expchange_backup.sql
```

### 查看数据库列表
```bash
mysql -h localhost -P 3306 -u referral_user -preferral123456 -e "SHOW DATABASES;"
```

## 📊 性能对比

| 特性 | SQLite | MySQL |
|------|--------|-------|
| 并发写入 | 较低 | 高 |
| 并发读取 | 中等 | 高 |
| 数据量支持 | < 1GB | > 100GB |
| 事务性能 | 中等 | 高 |
| 网络访问 | ❌ | ✅ |
| 适用场景 | 开发测试 | 生产环境 |

## ⚠️ 注意事项

1. **生产环境**：务必修改默认密码
2. **连接池**：MySQL 支持更大的并发连接
3. **备份**：定期备份生产数据
4. **监控**：建议使用 MySQL 监控工具
5. **索引**：GORM 会自动创建必要的索引

## 🔍 故障排查

### 连接失败
```bash
# 检查 MySQL 是否运行
docker ps | grep mysql

# 检查端口是否被占用
lsof -i :3306

# 查看 MySQL 日志
docker logs expchange-mysql
```

### 权限问题
```sql
-- 重置用户权限
GRANT ALL PRIVILEGES ON expchange.* TO 'expchange'@'%';
FLUSH PRIVILEGES;
```

### 字符集问题
```sql
-- 检查字符集
SHOW VARIABLES LIKE 'character%';
SHOW VARIABLES LIKE 'collation%';
```

更多详细配置，请参考 [MYSQL_CONFIG.md](./MYSQL_CONFIG.md)

