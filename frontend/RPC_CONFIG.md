# RPC 配置说明

## 📡 RPC 配置架构

### 当前配置方式

前端使用 **静态配置** + **后端动态配置** 的混合方案：

```
┌─────────────────────────────────────────────────────┐
│                  前端应用                            │
├─────────────────────────────────────────────────────┤
│                                                      │
│  1️⃣ Wagmi Config (静态)                             │
│     ├── RPC URLs (lib/wagmi.ts)                     │
│     │   ├── BSC: https://bsc-dataseed1.binance.org │
│     │   ├── Ethereum: https://eth.llamarpc.com     │
│     │   └── Polygon: https://polygon-rpc.com       │
│     └── 用于：钱包连接、余额查询、交易签名          │
│                                                      │
│  2️⃣ 后端配置 (动态)                                 │
│     ├── API: GET /api/chains                        │
│     │   ├── USDT 合约地址                           │
│     │   ├── 平台收款地址                            │
│     │   ├── USDT 精度                               │
│     │   └── 区块浏览器 URL                          │
│     └── 用于：合约交互参数                          │
│                                                      │
└─────────────────────────────────────────────────────┘
```

## 🔑 核心文件

### 1. `lib/wagmi.ts` - Wagmi/RainbowKit 配置

```typescript
transports: {
  [bsc.id]: http('https://bsc-dataseed1.binance.org'),  // ✅ 现在使用指定的 RPC
  [mainnet.id]: http('https://eth.llamarpc.com'),
  // ...
}
```

**作用**：
- 钱包连接时使用的 RPC
- 查询余额、发送交易时使用的 RPC
- **静态配置**，编译时确定

### 2. `hooks/useChains.ts` - 后端链配置

```typescript
const { data: chains } = useGetChainsQuery();  // 从后端获取
```

**返回数据**（示例）：
```json
{
  "chain_id": 56,
  "chain_name": "BSC",
  "rpc_url": "https://bsc-dataseed1.binance.org",      // 后端配置
  "usdt_contract_address": "0x55d3...B3197955",        // 后端配置
  "platform_deposit_address": "0x88888886...8e368db",  // 后端配置
  "usdt_decimals": 18,                                 // 后端配置
  "enabled": true
}
```

**作用**：
- 提供合约地址
- 提供平台收款地址
- 提供 USDT 精度
- 提供区块浏览器 URL
- **动态配置**，运行时从后端获取

### 3. `lib/chains/config.ts` - 已废弃（仅保留 ABI）

**作用**：
- ✅ 提供 `USDT_ABI` 定义
- ❌ 其他配置已不再使用

## 🎯 配置使用流程

### 场景 1：查询钱包 USDT 余额

```typescript
// 1. 获取后端链配置
const chainConfig = getChainById(chainId);  // 从后端 API

// 2. 创建合约实例
const ethersProvider = new ethers.BrowserProvider(walletClient);
//                                                  ^^^^^^^^^^^
//                                    这个 walletClient 使用 wagmi.ts 中配置的 RPC

const usdtContract = new ethers.Contract(
  chainConfig.usdt_contract_address,  // ✅ 后端配置
  USDT_ABI,                            // ✅ 静态 ABI
  ethersProvider                       // ⚠️ 使用 wagmi 配置的 RPC
);

const balance = await usdtContract.balanceOf(address);
```

**使用的 RPC**：
- ✅ **Wagmi 配置的 RPC**（`lib/wagmi.ts` 中指定）
- 现在是：`https://bsc-dataseed1.binance.org`（BSC）

### 场景 2：充值转账

```typescript
// 转账交易
const tx = await usdtContract.transfer(
  chainConfig.platform_deposit_address,  // ✅ 后端配置的收款地址
  amountInWei
);
```

**使用的 RPC**：
- ✅ **Wagmi 配置的 RPC**（用于广播交易）

### 场景 3：后端验证充值

```go
// backend/services/deposit_verifier.go
client, err := ethclient.Dial(chainConfig.RpcURL)
//                            ^^^^^^^^^^^^^^^^^^^
//                            ✅ 后端配置的 RPC
```

**使用的 RPC**：
- ✅ **后端数据库配置的 RPC**

## 🔄 RPC 配置同步

为了确保前后端使用相同的 RPC，建议：

### 方案 A：后端配置相同的 RPC（推荐）

在后端数据库的 `chain_configs` 表中，配置与前端相同的 RPC：

```sql
UPDATE chain_configs 
SET rpc_url = 'https://bsc-dataseed1.binance.org' 
WHERE chain_id = 56;
```

### 方案 B：环境变量传递

在 `frontend/.env.local` 中：
```bash
NEXT_PUBLIC_BSC_RPC=https://bsc-dataseed1.binance.org
NEXT_PUBLIC_ETH_RPC=https://eth.llamarpc.com
```

在 `lib/wagmi.ts` 中：
```typescript
transports: {
  [bsc.id]: http(process.env.NEXT_PUBLIC_BSC_RPC || 'https://bsc-dataseed1.binance.org'),
}
```

## 🚀 优化建议

### 1. 使用私有 RPC（推荐）

**公共 RPC 限制**：
- 速率限制（每秒请求数）
- 不稳定（可能宕机）
- 较慢（共享带宽）

**私有 RPC 提供商**：
- [Alchemy](https://www.alchemy.com/) - 免费额度 3M 请求/月
- [Infura](https://www.infura.io/) - 免费额度 100k 请求/天
- [QuickNode](https://www.quicknode.com/) - 多链支持
- [Ankr](https://www.ankr.com/) - 免费层

**配置示例**：
```typescript
transports: {
  [bsc.id]: http('https://your-project-id.bsc.rpc.thirdweb.com'),
  [mainnet.id]: http('https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY'),
}
```

### 2. 备用 RPC（容错）

Wagmi 支持配置多个 RPC 作为备用：

```typescript
import { http, fallback } from 'wagmi';

transports: {
  [bsc.id]: fallback([
    http('https://bsc-dataseed1.binance.org'),  // 主 RPC
    http('https://bsc-dataseed2.binance.org'),  // 备用 1
    http('https://bsc-dataseed3.binance.org'),  // 备用 2
  ]),
}
```

### 3. 缓存优化

配置请求缓存：
```typescript
transports: {
  [bsc.id]: http('https://bsc-dataseed1.binance.org', {
    batch: true,        // 批量请求
    timeout: 30_000,    // 30秒超时
  }),
}
```

## 📊 当前 RPC 使用情况

| 链 | 前端 RPC (Wagmi) | 后端 RPC (验证用) | 说明 |
|-----|------------------|-------------------|------|
| **BSC** | `bsc-dataseed1.binance.org` | 后端数据库配置 | 公共节点 |
| **Ethereum** | `eth.llamarpc.com` | 后端数据库配置 | 免费聚合 |
| **Sepolia** | `rpc.sepolia.org` | 后端数据库配置 | 测试网 |

## ✅ 总结

**现在的配置**：
- ✅ 前端使用 **优化的公共 RPC**（已在 `wagmi.ts` 中配置）
- ✅ 后端使用 **数据库配置的 RPC**（可在管理后台修改）
- ✅ 合约地址、平台地址等都从**后端动态获取**
- ✅ RPC 可以在前后端独立优化

**建议**：
1. 保持前端 RPC 静态配置（性能更好）
2. 后端数据库配置相同的 RPC URL（保持一致性）
3. 考虑升级到私有 RPC 节点（提高稳定性）

