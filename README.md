# 🚀 Velocity Exchange - 高性能加密货币交易所

<div align="center">

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)
![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)

一个功能完整的加密货币交易所，支持实时交易、K线图表、充值提现等功能。

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [技术栈](#️-技术栈) • [文档](#-文档)

</div>

---

## ✨ 功能特性

### 核心功能
- 🔐 **钱包登录** - 支持 MetaMask、Trust Wallet 等（RainbowKit 集成）
- 📊 **实时行情** - WebSocket 推送，K线图表（15s-1d 多种周期）
- 💱 **限价交易** - 高性能撮合引擎，毫秒级成交
- 📈 **盘口展示** - 实时买卖盘，深度数据
- 💰 **资产管理** - 充值提现、余额查询、交易记录
- 📱 **响应式设计** - 完美适配桌面和移动端

### 高级功能
- 🔄 **区块链充值** - USDT 智能合约充值（BSC）
- 💸 **自动提现** - 后台队列自动处理提现
- 🤖 **市场模拟器** - 自动生成市场数据（演示模式）
- 📉 **历史数据** - 自动生成6-12个月模拟交易数据
- 🎨 **美观UI** - 暗色主题，渐变效果，动画交互

## 🚀 快速开始

### 前置要求

- **Go** 1.21+
- **Node.js** 18+
- **npm** 或 **yarn**

### 安装步骤

```bash
# 1. 克隆仓库
git clone https://github.com/your-username/expchange.git
cd expchange

# 2. 启动后端
cd backend
go mod tidy
go run main.go

# 3. 启动前端（新终端）
cd frontend
npm install
npm run dev

# 4. 访问应用
# 前端: http://localhost:3000
# 后端: http://localhost:8080
```

### 环境变量配置

**后端** (`backend/.env`):
```bash
# 数据库
DB_NAME=expchange.db

# JWT密钥
JWT_SECRET=your_jwt_secret_here

# 服务端口
SERVER_PORT=8080

# 提现功能（可选）
PLATFORM_PRIVATE_KEY=your_private_key_here

# 市场模拟器（可选）
ENABLE_SIMULATOR=true
```

**前端** (`frontend/.env.local`):
```bash
# WalletConnect Project ID（可选）
NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID=your_project_id_here

# 后端地址（用于服务端代理）
BACKEND_URL=http://localhost:8080
```

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin (Web框架)
- **数据库**: SQLite + GORM
- **WebSocket**: Gorilla WebSocket
- **区块链**: go-ethereum (BSC 交互)

### 前端
- **框架**: Next.js 14 (React 18)
- **语言**: TypeScript
- **样式**: Tailwind CSS
- **状态管理**: Redux Toolkit (RTK Query)
- **图表**: Lightweight Charts
- **钱包**: RainbowKit + Wagmi + Viem

### 管理后台
- **框架**: Next.js 14
- **语言**: TypeScript
- **样式**: Tailwind CSS

## 📁 项目结构

```
expchange/
├── backend/              # Go 后端
│   ├── config/          # 配置管理
│   ├── database/        # 数据库初始化和迁移
│   ├── handlers/        # HTTP 处理器
│   ├── matching/        # 撮合引擎
│   ├── middleware/      # 中间件
│   ├── models/          # 数据模型
│   ├── services/        # 业务服务（充值验证、提现处理）
│   ├── simulator/       # 市场模拟器
│   ├── websocket/       # WebSocket 服务
│   └── main.go          # 入口文件
│
├── frontend/            # Next.js 前端
│   ├── app/             # App Router 页面
│   ├── components/      # React 组件
│   ├── lib/             # 工具库
│   │   ├── contracts/   # 智能合约交互
│   │   ├── services/    # API 服务（RTK Query）
│   │   └── store/       # Redux Store
│   ├── providers/       # Context Providers
│   └── server.js        # 自定义服务器（WebSocket 代理）
│
├── admin/               # 管理后台
│   └── ...
│
└── docs/                # 文档
    ├── QUICKSTART.md
    ├── DEPOSIT_WITHDRAW_GUIDE.md
    └── ...
```

## 📖 文档

- [快速开始指南](QUICKSTART.md) - 详细的安装和配置说明
- [充值提现指南](DEPOSIT_WITHDRAW_GUIDE.md) - 区块链充值提现功能说明
- [移动端访问](MOBILE_ACCESS.md) - 手机端访问配置
- [API 代理配置](PROXY_SETUP.md) - Next.js API 代理说明

## 🌟 核心特性详解

### 1. 撮合引擎

高性能内存撮合引擎：
- ✅ 价格-时间优先算法
- ✅ 毫秒级撮合速度
- ✅ 部分成交支持
- ✅ 实时 WebSocket 推送

### 2. K线生成器

智能 K线数据生成：
- ✅ 支持 15s-1d 多种周期
- ✅ 自动聚合交易数据
- ✅ 历史数据回填
- ✅ 实时更新推送

### 3. 充值提现系统

完整的区块链集成：
- ✅ USDT 智能合约充值（BSC）
- ✅ 自动交易验证队列
- ✅ 后台提现处理队列
- ✅ 交易记录追踪

### 4. 市场模拟器

演示模式数据生成：
- ✅ 智能价格趋势模拟
- ✅ 自动订单簿深度生成
- ✅ 6-12个月历史数据
- ✅ 可配置开关

## 🎨 截图

### 首页
![首页](docs/screenshots/home.png)

### 交易页面
![交易](docs/screenshots/trading.png)

### 资产页面
![资产](docs/screenshots/assets.png)

## 🔧 开发

### 后端开发

```bash
cd backend

# 运行
go run main.go

# 构建
go build -o expchange-backend

# 测试
go test ./...
```

### 前端开发

```bash
cd frontend

# 开发模式
npm run dev

# 构建
npm run build

# 生产模式
npm start
```

## 🚢 部署

### Docker 部署（推荐）

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 传统部署

参考 [部署指南](docs/DEPLOYMENT.md)

## ⚙️ 配置说明

### 代币配置

在 `backend/database/seed.go` 中修改交易对：

```go
pairs := []models.TradingPair{
    {Symbol: "BTC/USDT", BaseAsset: "BTC", QuoteAsset: "USDT", ...},
    // 添加更多交易对
}
```

### 手续费配置

在数据库中修改 `fee_configs` 表。

## 🤝 贡献

欢迎贡献！请查看 [贡献指南](CONTRIBUTING.md)。

## 📄 许可证

[MIT License](LICENSE)

## 🙏 致谢

- [RainbowKit](https://www.rainbowkit.com/) - 钱包连接
- [TradingView Lightweight Charts](https://www.tradingview.com/lightweight-charts/) - K线图表
- [Gin](https://gin-gonic.com/) - Go Web框架
- [Next.js](https://nextjs.org/) - React框架

## 📞 联系方式

- 📧 Email: your-email@example.com
- 🐦 Twitter: [@yourhandle](https://twitter.com/yourhandle)
- 💬 Discord: [Join our server](https://discord.gg/yourserver)

---

<div align="center">
Made with ❤️ by Your Team
</div>
