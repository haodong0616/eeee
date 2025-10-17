/**
 * USDT 合约 ABI 定义
 * 
 * 注意：本文件只提供 ABI 定义，其他链配置（RPC、合约地址等）都从后端获取
 * 
 * 配置来源：
 * - RPC URL：从 lib/wagmi.ts 配置（与后端保持一致）
 * - 链配置：从后端 API /api/chains 获取（通过 useChains hook）
 * - 合约地址、平台地址等：后端数据库管理
 */

// USDT ERC20 标准 ABI（兼容所有 EVM 链）
export const USDT_ABI = [
  'function balanceOf(address owner) view returns (uint256)',
  'function transfer(address to, uint amount) returns (bool)',
  'function allowance(address owner, address spender) view returns (uint256)',
  'function approve(address spender, uint amount) returns (bool)',
  'function decimals() view returns (uint8)',
  'function symbol() view returns (string)',
  'function name() view returns (string)',
];




