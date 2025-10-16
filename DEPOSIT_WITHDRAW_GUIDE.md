# 💰 充值提现功能完整指南

## 🎯 功能概述

实现了完整的区块链充值提现系统，包括：
- ✅ USDT 智能合约充值
- ✅ 后端自动验证交易
- ✅ 队列任务处理提现
- ✅ 自动转账到用户地址

## 📋 系统架构

```
充值流程:
用户钱包 → USDT合约转账 → 平台地址 → 提交txHash → 后端验证队列 → 确认到账

提现流程:
用户申请 → 冻结资金 → 提现队列 → 平台钱包转账 → 确认完成 → 解冻资金
```

## 🚀 启动步骤

### 1. 后端配置

在 `backend` 目录创建 `.env` 文件或设置环境变量：

```bash
# 平台私钥（用于提现转账）
export PLATFORM_PRIVATE_KEY=your_private_key_here

# 可选：启用市场模拟器
export ENABLE_SIMULATOR=true
```

⚠️ **重要**：私钥必须对应有足够 USDT 和 BNB（gas费）的地址！

### 2. 启动后端

```bash
cd backend
go run main.go
```

应该看到：
```
✅ 充值验证队列已启动
✅ 提现处理队列已启动
```

### 3. 安装前端依赖

```bash
cd frontend
npm install
```

### 4. 启动前端

```bash
npm run dev
```

## 💳 充值功能

### 前端流程

1. 用户连接钱包（BSC链）
2. 进入资产页面
3. 点击"充值"按钮
4. 输入金额
5. 调用 USDT 合约转账到平台地址
6. 提交交易hash到后端

### 关键代码

**前端合约调用** (`frontend/lib/contracts/depositService.ts`):
```typescript
// 执行 USDT 转账
const tx = await usdtContract.transfer(
  PLATFORM_DEPOSIT_ADDRESS, // 0x88888886757311de33778ce108fb312588e368db
  amountInWei
);

// 等待确认
await tx.wait();

// 提交到后端
await depositMutation({ 
  asset: 'USDT', 
  amount,
  txHash: tx.hash 
}).unwrap();
```

**后端验证** (`backend/services/deposit_verifier.go`):
- 每30秒检查pending状态的充值
- 验证交易是否成功
- 验证接收地址是否正确
- 等待至少1个区块确认
- 自动增加用户余额

### 配置参数

- **BSC RPC**: `https://bsc-dataseed1.binance.org`
- **USDT合约**: `0x55d398326f99059fF775485246999027B3197955`
- **平台地址**: `0x88888886757311de33778ce108fb312588e368db`
- **验证间隔**: 30秒
- **确认数**: 1个区块

## 💸 提现功能

### 前端流程

1. 用户进入资产页面
2. 点击"提现"按钮
3. 输入提现地址和金额
4. 提交申请（资金被冻结）
5. 等待后端处理

### 关键代码

**前端提交** (`frontend/app/assets/page.tsx`):
```typescript
await withdrawMutation({ 
  asset: 'USDT', 
  amount, 
  address: withdrawAddress 
}).unwrap();
```

**后端处理** (`backend/services/withdraw_processor.go`):
- 每1分钟检查pending状态的提现
- 验证余额是否足够
- 调用USDT合约转账
- 等待交易确认
- 解冻资金并标记完成

### 状态流转

```
pending → processing → completed
   ↓           ↓
 failed ← (转账失败时解冻资金)
```

## 🗄️ 数据模型

### DepositRecord (充值记录)

```go
type DepositRecord struct {
	ID        uint            // 记录ID
	UserID    uint            // 用户ID
	Asset     string          // 资产类型 (USDT)
	Amount    decimal.Decimal // 充值金额
	TxHash    string          // 交易hash (唯一)
	Status    string          // pending/confirmed/failed
	CreatedAt time.Time       
	UpdatedAt time.Time
}
```

### WithdrawRecord (提现记录)

```go
type WithdrawRecord struct {
	ID        uint            // 记录ID
	UserID    uint            // 用户ID
	Asset     string          // 资产类型 (USDT)
	Amount    decimal.Decimal // 提现金额
	Address   string          // 提现地址
	TxHash    string          // 转账hash (成功后填充)
	Status    string          // pending/processing/completed/failed
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

## 🔧 API 接口

### 充值接口

**POST** `/api/balances/deposit`

Request:
```json
{
  "asset": "USDT",
  "amount": "100.5",
  "txHash": "0x..."
}
```

Response:
```json
{
  "message": "Deposit submitted for verification",
  "deposit": {
    "id": 1,
    "user_id": 1,
    "asset": "USDT",
    "amount": "100.5",
    "tx_hash": "0x...",
    "status": "pending",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### 提现接口

**POST** `/api/balances/withdraw`

Request:
```json
{
  "asset": "USDT",
  "amount": "50.0",
  "address": "0x..."
}
```

Response:
```json
{
  "message": "Withdrawal request submitted",
  "withdrawal": {
    "id": 1,
    "user_id": 1,
    "asset": "USDT",
    "amount": "50.0",
    "address": "0x...",
    "status": "pending",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### 充值记录

**GET** `/api/balances/deposits`

Response:
```json
[
  {
    "id": 1,
    "amount": "100.5",
    "tx_hash": "0x...",
    "status": "confirmed",
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

### 提现记录

**GET** `/api/balances/withdraws`

Response:
```json
[
  {
    "id": 1,
    "amount": "50.0",
    "address": "0x...",
    "tx_hash": "0x...",
    "status": "completed",
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

## 🔍 验证流程详解

### 充值验证

1. **查询待验证记录**
   - 状态为 `pending`
   - 创建时间在24小时内

2. **获取交易收据**
   ```go
   receipt, err := client.TransactionReceipt(ctx, txHash)
   ```

3. **验证交易状态**
   - `receipt.Status == 1` (成功)

4. **验证接收地址**
   - 交易目标为 USDT 合约地址

5. **等待确认**
   - 至少1个区块确认

6. **增加余额**
   - 更新记录状态为 `confirmed`
   - 增加用户可用余额

### 提现处理

1. **查询待处理记录**
   - 状态为 `pending`

2. **标记为处理中**
   - 更新状态为 `processing`

3. **执行转账**
   ```go
   tx := types.NewTransaction(...)
   signedTx := types.SignTx(tx, signer, privateKey)
   client.SendTransaction(ctx, signedTx)
   ```

4. **更新记录**
   - 状态改为 `completed`
   - 填充 `tx_hash`
   - 减少冻结余额

## ⚠️ 安全注意事项

### 1. 私钥管理

```bash
# ❌ 不要这样做
PLATFORM_PRIVATE_KEY=0x1234...直接写在代码里

# ✅ 使用环境变量
export PLATFORM_PRIVATE_KEY=0x...

# ✅ 或使用密钥管理服务
# AWS Secrets Manager, Azure Key Vault, etc.
```

### 2. 地址验证

```go
// 验证地址格式
if !strings.HasPrefix(address, "0x") || len(address) != 42 {
    return errors.New("invalid address")
}
```

### 3. 金额验证

```go
// 检查最小金额
if amount.LessThan(decimal.NewFromFloat(0.01)) {
    return errors.New("amount too small")
}

// 检查余额
if balance.Available.LessThan(amount) {
    return errors.New("insufficient balance")
}
```

### 4. 重复检查

```go
// 防止重复充值
var existing DepositRecord
if err := db.Where("tx_hash = ?", txHash).First(&existing).Error; err == nil {
    return errors.New("duplicate transaction")
}
```

## 🐛 故障排查

### 问题1: 充值一直 pending

**可能原因**:
- 交易未上链
- 交易失败
- BSC RPC 连接问题

**解决方案**:
```bash
# 检查交易状态
curl https://bsc-dataseed1.binance.org \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x..."],"id":1}'

# 查看后端日志
tail -f backend.log
```

### 问题2: 提现失败

**可能原因**:
- 平台钱包余额不足
- Gas费不足
- 私钥错误

**解决方案**:
```bash
# 检查平台钱包余额
# 确保有足够的 USDT 和 BNB (gas)

# 验证私钥
go run check_balance.go
```

### 问题3: Gas费太高

**优化方案**:
```go
// 使用较低的 gas price
gasPrice := big.NewInt(5000000000) // 5 Gwei

// 或从网络获取推荐值
gasPrice, _ := client.SuggestGasPrice(ctx)
```

## 📊 监控建议

### 1. 充值监控

```sql
-- 查看pending充值
SELECT * FROM deposit_records 
WHERE status = 'pending' 
AND created_at > datetime('now', '-1 hour');

-- 统计充值
SELECT 
  status, 
  COUNT(*) as count, 
  SUM(amount) as total
FROM deposit_records
GROUP BY status;
```

### 2. 提现监控

```sql
-- 查看pending提现
SELECT * FROM withdraw_records 
WHERE status IN ('pending', 'processing');

-- 统计提现
SELECT 
  status, 
  COUNT(*) as count, 
  SUM(amount) as total
FROM withdraw_records
GROUP BY status;
```

### 3. 平台余额监控

定期检查平台钱包余额，确保有足够资金处理提现。

## 🎉 测试流程

### 1. 测试充值

```bash
# 1. 连接钱包（确保在BSC链）
# 2. 确保钱包有USDT
# 3. 进入资产页面
# 4. 点击充值，输入金额
# 5. 确认合约交易
# 6. 等待30秒-2分钟
# 7. 刷新页面查看余额
```

### 2. 测试提现

```bash
# 1. 确保账户有可用余额
# 2. 点击提现，输入地址和金额
# 3. 提交申请
# 4. 等待1-3分钟
# 5. 检查目标地址是否收到USDT
```

## 📝 TODO

- [ ] 添加提现手续费
- [ ] 实现提现审核机制
- [ ] 添加充值金额限制
- [ ] 实现批量提现
- [ ] 添加邮件通知
- [ ] 接入更多RPC节点（容错）
- [ ] 实现更详细的交易事件解析
- [ ] 添加管理后台查看充值提现记录

---

完整实现！现在您的交易所支持真实的区块链充值提现了！🚀

