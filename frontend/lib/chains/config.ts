/**
 * 多链配置
 */

export interface ChainConfig {
  id: number;
  name: string;
  displayName: string;
  rpcUrl: string;
  blockExplorer: string;
  nativeCurrency: {
    name: string;
    symbol: string;
    decimals: number;
  };
  usdtContract: string;
  platformDepositAddress: string;
  isTestnet: boolean;
}

export const SUPPORTED_CHAINS: Record<string, ChainConfig> = {
  bsc: {
    id: 56,
    name: 'bsc',
    displayName: 'BSC (币安智能链)',
    rpcUrl: 'https://bsc-dataseed1.binance.org',
    blockExplorer: 'https://bscscan.com',
    nativeCurrency: {
      name: 'BNB',
      symbol: 'BNB',
      decimals: 18,
    },
    usdtContract: '0x55d398326f99059fF775485246999027B3197955',
    platformDepositAddress: '0x88888886757311de33778ce108fb312588e368db',
    isTestnet: false,
  },
  sepolia: {
    id: 11155111,
    name: 'sepolia',
    displayName: 'Sepolia (测试网)',
    rpcUrl: 'https://rpc.sepolia.org',
    blockExplorer: 'https://sepolia.etherscan.io',
    nativeCurrency: {
      name: 'Sepolia ETH',
      symbol: 'ETH',
      decimals: 18,
    },
    // Sepolia测试网USDT合约地址（示例，需要替换为实际地址）
    usdtContract: '0x7169D38820dfd117C3FA1f22a697dBA58d90BA06',
    platformDepositAddress: '0x88888886757311de33778ce108fb312588e368db',
    isTestnet: true,
  },
};

export const DEFAULT_CHAIN = 'bsc';

export const USDT_DECIMALS = 18;

// USDT ABI（兼容多链）
export const USDT_ABI = [
  'function balanceOf(address owner) view returns (uint256)',
  'function transfer(address to, uint amount) returns (bool)',
  'function allowance(address owner, address spender) view returns (uint256)',
  'function approve(address spender, uint amount) returns (bool)',
  'function decimals() view returns (uint8)',
  'function symbol() view returns (string)',
  'function name() view returns (string)',
];

export function getChainConfig(chainKey: string): ChainConfig {
  return SUPPORTED_CHAINS[chainKey] || SUPPORTED_CHAINS[DEFAULT_CHAIN];
}

export function getChainById(chainId: number): ChainConfig | undefined {
  return Object.values(SUPPORTED_CHAINS).find(chain => chain.id === chainId);
}

export function getChainKeyById(chainId: number): string {
  const entry = Object.entries(SUPPORTED_CHAINS).find(([_, chain]) => chain.id === chainId);
  return entry ? entry[0] : DEFAULT_CHAIN;
}




