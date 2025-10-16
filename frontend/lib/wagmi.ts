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
import { bsc } from 'wagmi/chains';

// WalletConnect Project ID
const walletConnectProjectId = process.env.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID || 'YOUR_PROJECT_ID_HERE';

// 只支持 BSC 主网
const chains = [bsc] as const;

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
export const wagmiConfig = createConfig({
  chains,
  connectors,
  transports: {
    [bsc.id]: http(),
  },
  ssr: true,
});

