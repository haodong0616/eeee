# 🌈 RainbowKit 钱包集成指南

## 概述

Velocity Exchange 现已集成 **RainbowKit**，提供美观、易用的 Web3 钱包连接体验。

## 🎯 支持的链

目前配置支持：
- ✅ **BSC Mainnet** (Chain ID: 56)
- ✅ **BSC Testnet** (Chain ID: 97)

可以轻松扩展到其他 EVM 链。

## 🔌 支持的钱包

### 推荐钱包
- 浏览器内置钱包（自动检测）
- MetaMask
- Rabby
- Coinbase Wallet
- Brave Wallet

### 移动端钱包
- WalletConnect（支持所有移动钱包）
- Trust Wallet
- Rainbow

### 交易所钱包
- OKX Wallet

## 📦 技术栈

- **@rainbow-me/rainbowkit** - UI 组件库
- **wagmi** - React Hooks for Ethereum
- **viem** - TypeScript 以太坊库
- **@tanstack/react-query** - 数据获取和缓存

## 🚀 快速开始

### 1. 安装依赖

```bash
cd frontend
npm install
```

### 2. 配置环境变量

创建 `frontend/.env.local`：

```env
# WalletConnect Project ID（可选，推荐配置）
NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID=your_project_id_here

# 后端地址
BACKEND_URL=localhost:8080
```

**获取 WalletConnect Project ID**：
1. 访问 https://cloud.walletconnect.com/
2. 注册/登录
3. 创建新项目
4. 复制 Project ID

### 3. 启动服务

```bash
# 启动后端
cd backend
go run main.go

# 启动前端（新终端）
cd frontend
npm run dev
```

## 📁 核心文件

### `lib/wagmi.ts`
Wagmi 配置文件，定义支持的链和钱包。

```typescript
import { bsc, bscTestnet } from 'wagmi/chains';

const chains = [bsc, bscTestnet] as const;
```

### `providers/RainbowProvider.tsx`
RainbowKit Provider，包裹应用。

```typescript
<RainbowKitProvider
  modalSize="compact"
  showRecentTransactions={true}
  coolMode={true}
>
  {children}
</RainbowKitProvider>
```

### `components/Header.tsx`
使用 `ConnectButton` 组件。

```typescript
import { ConnectButton } from '@rainbow-me/rainbowkit';

<ConnectButton 
  chainStatus="icon"
  showBalance={{
    smallScreen: false,
    largeScreen: true,
  }}
/>
```

## 🔄 状态同步

钱包地址会自动同步到 Redux：

```typescript
// wagmi → Redux
const { address, isConnected } = useAccount();

useEffect(() => {
  if (isConnected && address) {
    dispatch(setWalletAddress(address.toLowerCase()));
  }
}, [isConnected, address]);
```

## 🎨 自定义样式

在 `globals.css` 中自定义 RainbowKit 按钮：

```css
.rainbow-wallet-btn button {
  background: linear-gradient(to right, rgb(168 85 247), rgb(236 72 153)) !important;
  border-radius: 0.5rem !important;
}
```

## 📱 移动端支持

- ✅ 响应式设计
- ✅ WalletConnect 支持所有移动钱包
- ✅ 扫码连接
- ✅ 深度链接

## 🔧 常见问题

### Q: 钱包连接后页面不刷新？

**A**: 已实现自动同步，钱包连接状态会自动更新到 Redux。

### Q: 手机端无法连接？

**A**: 确保：
1. 使用 WalletConnect 选项
2. 手机已安装钱包 APP
3. 扫描二维码或通过深度链接打开

### Q: 为什么需要 WalletConnect Project ID？

**A**: 
- WalletConnect 用于移动端连接
- 免费注册即可获取
- 不配置也能用，但推荐配置以获得更好体验

### Q: 如何添加其他链？

**A**: 修改 `lib/wagmi.ts`：

```typescript
import { bsc, ethereum, polygon } from 'wagmi/chains';

const chains = [bsc, ethereum, polygon] as const;

// 添加 transports
transports: {
  [bsc.id]: http(),
  [ethereum.id]: http(),
  [polygon.id]: http(),
}
```

## 🔐 安全说明

- ✅ 钱包地址统一转换为小写存储
- ✅ 不存储私钥
- ✅ 签名在客户端完成
- ✅ 使用标准 EIP-712 签名

## 📚 相关资源

- [RainbowKit 文档](https://www.rainbowkit.com/)
- [wagmi 文档](https://wagmi.sh/)
- [WalletConnect](https://walletconnect.com/)

## ✨ 特性

- 🎨 美观的钱包选择 UI
- 🚀 一键连接多种钱包
- 🌐 自动检测已安装的钱包
- 📱 完美支持移动端
- 🔄 实时同步链切换
- 💎 显示余额和最近交易
- 🌈 Cool Mode 动画效果

---

集成完成！现在用户可以通过 RainbowKit 连接他们喜欢的钱包了 🎉

