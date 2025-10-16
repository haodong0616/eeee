# Velocity Exchange - 一键部署指南

⚡ **速度交易所** - 超简单启动流程

### 后端（自动初始化）

```bash
cd backend
go mod download
go run main.go
```

**就这么简单！** 首次启动会自动完成所有初始化。

你会看到：
```
🚀 首次启动，开始自动初始化数据...
📊 创建交易对...
✅ 创建了 6 个交易对
💱 生成30天交易数据...
✅ 生成了 4320 条交易记录 (30天)
📈 生成K线数据...
✅ 生成了 12580 根K线
🎉 数据初始化完成！系统已就绪
💡 访问 http://localhost:3000 查看前端
💡 访问 http://localhost:3001 查看管理后台
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

访问：http://localhost:3000

### 管理后台

```bash
cd admin
npm install
npm run dev
```

访问：http://localhost:3001  
登录：admin / admin

## 启用演示模式（可选）

如果想要价格自动波动和盘口自动挂单：

```bash
cd backend
echo "ENABLE_SIMULATOR=true" > .env
go run main.go
```

## 重置数据

如果想重新开始：

```bash
cd backend
rm expchange.db
go run main.go  # 自动重新初始化
```

## 生产部署

### 1. 后端编译

```bash
cd backend
go build -o expchange-server
./expchange-server
```

### 2. 前端构建

```bash
cd frontend
npm run build
npm start
```

### 3. 管理后台构建

```bash
cd admin
npm run build
npm start
```

### 4. 使用 Docker

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    volumes:
      - ./backend/expchange.db:/app/expchange.db
    environment:
      - ENABLE_SIMULATOR=true
      - JWT_SECRET=your-production-secret

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080

  admin:
    build: ./admin
    ports:
      - "3001:3001"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080
```

启动：
```bash
docker-compose up -d
```

## 环境变量

### 后端 (.env)
```env
SERVER_PORT=8080
DB_NAME=expchange.db
JWT_SECRET=your-secret-key
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
ENABLE_SIMULATOR=true
```

### 前端 (.env.local)
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 性能优化

### SQLite 优化
在 `database/database.go` 中添加：

```go
DB.Exec("PRAGMA journal_mode=WAL;")
DB.Exec("PRAGMA synchronous=NORMAL;")
DB.Exec("PRAGMA cache_size=-64000;")
```

### 定期清理
清理30天前的秒级K线：

```sql
DELETE FROM klines 
WHERE interval IN ('15s', '30s') 
AND created_at < datetime('now', '-30 days');
```

## 监控

### 查看数据库
```bash
sqlite3 backend/expchange.db
.tables
SELECT COUNT(*) FROM trades;
SELECT COUNT(*) FROM klines;
.exit
```

### 查看日志
```bash
cd backend
go run main.go 2>&1 | tee app.log
```

## 故障排查

### 问题：数据没有自动初始化
**原因**：数据库已存在  
**解决**：删除 `expchange.db` 重新启动

### 问题：端口被占用
**解决**：
```bash
# 修改端口
export SERVER_PORT=8081
go run main.go
```

### 问题：前端连接不上后端
**检查**：
1. 后端是否在运行
2. 端口是否正确
3. CORS 配置是否正确

## 成功标志

启动成功后：
- ✅ 后端日志显示"数据初始化完成"
- ✅ 访问前端能看到 6 个交易对
- ✅ 行情页面有价格数据
- ✅ 交易页面K线图正常显示
- ✅ 盘口有买卖单（如果启用模拟器）

## 零到生产只需三步

```bash
# 1. 启动后端（自动初始化）
cd backend && go run main.go

# 2. 启动前端
cd frontend && npm install && npm run dev

# 3. 访问系统
open http://localhost:3000
```

就这么简单！🚀

