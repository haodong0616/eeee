# ExpChange 快速启动指南

本指南帮助你在 5 分钟内启动整个交易所系统。

## 前置要求

确保已安装以下软件：

- **Go** 1.21 或更高版本
- **Node.js** 18 或更高版本
- **Redis** 6 或更高版本（可选，用于缓存）

## 快速启动步骤

### 1. 启动Redis服务（可选）

#### Redis
```bash
# macOS
brew services start redis

# Linux
sudo systemctl start redis
```

### 2. 启动后端服务（自动初始化）

```bash
# 进入后端目录
cd backend

# 下载依赖
go mod download

# 启动服务（首次启动会自动初始化所有数据）
go run main.go
```

**首次启动时会自动**（智能模式）：
- ✅ 创建数据库文件 `expchange.db`
- ✅ 创建 6 个交易对
- ✅ 生成 6-12 个月历史数据（随机）
- ✅ 每个代币随机交易活跃度
- ✅ 生成常用周期的 K 线数据
- ✅ 初始化手续费配置
- ✅ 创建虚拟用户用于模拟交易

**初始化速度**：15-30 秒 ⚡（取决于随机月数）

后端将在 `http://localhost:8080` 启动。

### 3. 启动用户前端

打开新终端：

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:3000` 启动。

### 4. 启动管理后台

打开新终端：

```bash
# 进入管理后台目录
cd admin

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

管理后台将在 `http://localhost:3001` 启动。

## 访问系统

### 用户端
打开浏览器访问：`http://localhost:3000`

1. 点击右上角"连接钱包"
2. 在 MetaMask 中确认连接
3. 签名消息完成登录

### 管理后台
打开浏览器访问：`http://localhost:3001`

默认登录信息：
- 用户名：`admin`
- 密码：`admin`

## 数据说明

### 自动初始化的数据

首次启动后端时，系统已自动创建：

**交易对（6个）**：
- 🌙 LUNAR/USDT (~$8,500)
- ⭐ NOVA/USDT (~$1,250)
- 🔗 NEXUS/USDT (~$125)
- 🌬️ ZEPHYR/USDT (~$45)
- 💓 PULSE/USDT (~$2.85)
- 🔮 ARCANA/USDT (~$0.0085)

**历史数据（智能策略）**：
- 📅 6-12 个月交易记录（随机）
- 🎲 每个代币不同的交易活跃度
- 📊 10,000-30,000 条真实感交易
- 🌓 白天/夜间不同交易模式
- 📈 包含价格趋势和周期性波动
- ⏱️ K线数据（1m, 5m, 15m, 1h, 4h, 1d）

**初始化时间**：约 15-30 秒（自动优化批量插入）

### 充值测试资金

在用户端：
1. 连接钱包并登录
2. 进入"资产"页面
3. 点击"充值"
4. 选择资产：`USDT`
5. 输入金额：`10000`
6. 确认充值

### 3. 开始交易

在用户端：
1. 进入"交易"页面（如 LUNAR/USDT）
2. 查看实时K线图和盘口深度
3. 点击盘口价格自动填入
4. 点击余额自动填入最大数量
5. 点击"买入"或"卖出"

**提示**：启用市场模拟器后，盘口会有虚拟挂单，可以立即成交！

## 常见问题

### Redis 连接失败（可选功能）

检查 Redis 是否正在运行：
```bash
redis-cli ping
```

应该返回 `PONG`。如果未运行：
```bash
brew services start redis      # macOS
sudo systemctl start redis     # Linux
```

### 前端连接后端失败

确保后端正在 8080 端口运行：
```bash
curl http://localhost:8080/api/market/pairs
```

### MetaMask 未安装

访问 https://metamask.io 下载安装 MetaMask 浏览器扩展。

### 端口被占用

如果端口被占用，可以修改配置：

**后端** - 编辑 `backend/.env`:
```env
SERVER_PORT=8081
```

**前端** - 修改启动命令:
```bash
npm run dev -- -p 3002
```

**管理后台** - 修改启动命令:
```bash
npm run dev -- -p 3003
```

## 下一步

1. **浏览功能**
   - 探索首页、行情、交易和资产页面
   - 尝试创建不同类型的订单
   - 查看实时盘口和成交记录

2. **管理数据**
   - 在管理后台查看用户、订单和成交数据
   - 添加更多交易对
   - 查看系统统计信息

3. **深入开发**
   - 阅读各模块的 README 文件
   - 修改代码并测试
   - 集成 TradingView 图表
   - 添加更多功能

## 停止服务

在各个终端中按 `Ctrl + C` 停止服务。

如果启动了Redis，可以停止：
```bash
brew services stop redis      # macOS
sudo systemctl stop redis     # Linux
```

SQLite数据库文件 `expchange.db` 会保留在 `backend/` 目录中，下次启动会自动使用。

## 获取帮助

- 查看主 README.md 了解详细文档
- 查看各模块的 README 了解具体实现
- 提交 Issue 报告问题或建议

祝你使用愉快！🚀

