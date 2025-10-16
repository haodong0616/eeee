# 🚀 快速启动充值提现功能

## ⚡ 快速开始

### 1. 后端配置（重要！）

```bash
cd backend

# 设置平台私钥（用于提现转账）
export PLATFORM_PRIVATE_KEY=your_private_key_here

# 可选：启用市场模拟器
export ENABLE_SIMULATOR=true

# 安装 Go 依赖
go mod tidy

# 启动后端
go run main.go
```

应该看到：
```
✅ 充值验证队列已启动
✅ 提现处理队列已启动
🚀 Server starting on port 8080
```

### 2. 前端配置

```bash
cd frontend

# 安装依赖（首次）
npm install

# 启动前端
npm run dev
```

访问 `http://localhost:3000`

## 💡 测试流程

### 测试充值

1. **连接钱包**
   - 确保钱包在 BSC 主网
   - 确保有 USDT 和 BNB（gas费）

2. **进入资产页面**
   - 点击右上角"资产"

3. **开始充值**
   - 点击"充值"按钮
   - 输入金额（例如：10）
   - 确认合约交易
   - 等待钱包弹出确认

4. **等待到账**
   - 交易确认后约30秒-2分钟到账
   - 刷新页面查看余额

### 测试提现

1. **确保有余额**
   - 必须先完成充值

2. **申请提现**
   - 点击"提现"按钮
   - 输入提现地址（0x...）
   - 输入金额
   - 提交申请

3. **等待处理**
   - 资金会被冻结
   - 后端每1分钟处理一次
   - 约1-3分钟到账

## 📋 重要配置

### 合约地址

- **USDT合约**: `0x55d398326f99059fF775485246999027B3197955`
- **平台收款地址**: `0x88888886757311de33778ce108fb312588e368db`
- **BSC RPC**: `https://bsc-dataseed1.binance.org`

### 队列配置

- **充值验证**: 每30秒检查一次
- **提现处理**: 每1分钟处理一次
- **确认数**: 1个区块

## ⚠️ 注意事项

### 1. 私钥安全

```bash
# ❌ 不要提交到 git
echo "PLATFORM_PRIVATE_KEY=0x..." >> backend/.env
git add backend/.env  # 不要这样做！

# ✅ 使用环境变量
export PLATFORM_PRIVATE_KEY=0x...
```

### 2. 平台钱包准备

平台钱包（对应私钥的地址）需要：
- ✅ 足够的 USDT（用于提现）
- ✅ 足够的 BNB（用于 gas 费）
- ✅ 在 BSC 主网

### 3. 首次充值

首次充值前，用户需要：
- 钱包切换到 BSC 主网
- 有足够的 USDT 和 BNB
- 确认智能合约授权

## 🔍 查看日志

### 后端日志

```bash
# 查看充值验证日志
# 应该看到：
📋 发现 X 条待验证充值
🔍 验证充值: ID=1, Hash=0x..., Amount=100
✅ 充值验证成功: 0x...
🎉 充值已到账: 用户ID=1, 资产=USDT, 金额=100
```

### 提现日志

```bash
# 应该看到：
📋 发现 X 条待处理提现
💸 处理提现: ID=1, Amount=50 USDT, Address=0x...
✅ 转账成功: 0x...
🎉 提现已完成: 用户ID=1, 资产=USDT, 金额=50
```

## 🐛 常见问题

### Q1: 充值一直 pending？

**检查**：
1. 交易是否成功？（在 BSCScan 查询）
2. 接收地址是否正确？
3. 后端验证队列是否运行？

### Q2: 提现失败？

**检查**：
1. 平台钱包 USDT 余额是否足够？
2. 平台钱包 BNB（gas）余额是否足够？
3. 私钥是否正确配置？
4. 后端处理队列是否运行？

### Q3: 如何查看交易状态？

访问 BSCScan：
- 主网: https://bscscan.com/tx/0x...
- 测试网: https://testnet.bscscan.com/tx/0x...

## 📚 完整文档

查看 `DEPOSIT_WITHDRAW_GUIDE.md` 获取：
- 完整架构说明
- API接口文档
- 安全最佳实践
- 故障排查指南

## 🎉 快速验证

```bash
# 1. 启动后端
cd backend && go run main.go

# 2. 启动前端（新终端）
cd frontend && npm run dev

# 3. 访问
open http://localhost:3000

# 4. 测试充值提现
```

---

**重要提醒**：
- ⚠️ 测试时使用小金额
- ⚠️ 保护好私钥
- ⚠️ 确认网络是 BSC 主网
- ⚠️ 平台钱包保持足够余额

祝您使用愉快！🚀

