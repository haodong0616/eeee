// USDT 合约 ABI（只包含需要的方法）
export const USDT_ABI = [
  // 转账方法
  {
    constant: false,
    inputs: [
      { name: '_to', type: 'address' },
      { name: '_value', type: 'uint256' }
    ],
    name: 'transfer',
    outputs: [{ name: '', type: 'bool' }],
    type: 'function'
  },
  // 查询余额
  {
    constant: true,
    inputs: [{ name: 'who', type: 'address' }],
    name: 'balanceOf',
    outputs: [{ name: '', type: 'uint256' }],
    type: 'function'
  },
  // 授权
  {
    constant: false,
    inputs: [
      { name: '_spender', type: 'address' },
      { name: '_value', type: 'uint256' }
    ],
    name: 'approve',
    outputs: [{ name: '', type: 'bool' }],
    type: 'function'
  },
  // 查询授权额度
  {
    constant: true,
    inputs: [
      { name: '_owner', type: 'address' },
      { name: '_spender', type: 'address' }
    ],
    name: 'allowance',
    outputs: [{ name: '', type: 'uint256' }],
    type: 'function'
  },
  // decimals
  {
    constant: true,
    inputs: [],
    name: 'decimals',
    outputs: [{ name: '', type: 'uint8' }],
    type: 'function'
  }
] as const;

// BSC 主网 USDT 合约地址
export const USDT_CONTRACT_ADDRESS = '0x55d398326f99059fF775485246999027B3197955';

// 平台收款地址
export const PLATFORM_DEPOSIT_ADDRESS = '0x88888886757311de33778ce108fb312588e368db';

// USDT 精度
export const USDT_DECIMALS = 18;

