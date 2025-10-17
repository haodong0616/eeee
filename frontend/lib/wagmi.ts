import { connectorsForWallets } from '@rainbow-me/rainbowkit';
import {
  metaMaskWallet,
  walletConnectWallet,
  coinbaseWallet,
  trustWallet,
  okxWallet,
  injectedWallet,
  rabbyWallet,
  braveWallet,
  rainbowWallet,
} from '@rainbow-me/rainbowkit/wallets';
import { createConfig, http } from 'wagmi';
import { bsc, sepolia, mainnet, polygon, arbitrum } from 'wagmi/chains';
import type { Chain } from 'wagmi/chains';

// WalletConnect Project ID
const walletConnectProjectId = process.env.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID || 'YOUR_PROJECT_ID_HERE';

// 定义所有可能支持的链（从后端动态启用/禁用）
// RainbowKit会显示所有这些链，用户可以切换
// 后端会验证当前链是否启用
const chains = [
  mainnet,    // Ethereum Mainnet (1)
  bsc,        // BSC Mainnet (56)
  polygon,    // Polygon Mainnet (137)
  arbitrum,   // Arbitrum One (42161)
  sepolia,    // Sepolia Testnet (11155111)
] as const;

// 配置钱包列表
const connectors = connectorsForWallets(
  [
    {
      groupName: '推荐钱包',
      wallets: [
        injectedWallet,     // 浏览器内置钱包（自动检测）
        metaMaskWallet,     // MetaMask
        rabbyWallet,        // Rabby
        coinbaseWallet,     // Coinbase Wallet
        braveWallet,        // Brave Wallet
      ],
    },
    {
      groupName: '移动端钱包',
      wallets: [
        walletConnectWallet, // WalletConnect（支持所有移动钱包）
        trustWallet,         // Trust Wallet
        rainbowWallet,       // Rainbow
      ],
    },
    {
      groupName: '交易所钱包',
      wallets: [
        okxWallet,          // OKX Wallet
      ],
    },
  ],
  {
    appName: 'Velocity Exchange',
    projectId: walletConnectProjectId,
  }
);

// 创建 Wagmi 配置
// 🔧 配置优化的 RPC 节点（与后端配置保持一致）
export const wagmiConfig = createConfig({
  chains,
  connectors,
  transports: {
    // 使用优化的 RPC（可在后端数据库中配置相同的 URL）
    [mainnet.id]: http('https://eth.llamarpc.com'),           // Ethereum Mainnet
    [bsc.id]: http('https://bsc-dataseed1.binance.org'),      // BSC Mainnet
    [polygon.id]: http('https://polygon-rpc.com'),            // Polygon
    [arbitrum.id]: http('https://arb1.arbitrum.io/rpc'),      // Arbitrum
    [sepolia.id]: http('https://ethereum-sepolia.publicnode.com'),            // Sepolia Testnet
  },
  ssr: true,
});

