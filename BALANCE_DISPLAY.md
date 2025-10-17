# 余额显示和验证

## 下单表单的余额显示

### 买入时
显示**报价资产**（Quote Asset）的可用余额：
```
交易对：BTC/USDT
买入方向：显示 USDT 余额
可用 USDT: 10,000.00000000
```

### 卖出时
显示**基础资产**（Base Asset）的可用余额：
```
交易对：BTC/USDT
卖出方向：显示 BTC 余额
可用 BTC: 1.50000000
```

## 预计花费/获得

### 限价单
输入价格和数量后，自动计算：

**买入**：
```
价格：65,000 USDT
数量：0.5 BTC
---
预计花费：32,500.00 USDT
```

**卖出**：
```
价格：65,000 USDT
数量：0.5 BTC
---
预计获得：32,500.00 USDT
```

### 市价单
使用当前市场价格估算。

## 余额验证

下单前会自动检查余额：

### 买入验证
```typescript
required = 价格 × 数量
if (USDT余额 < required) {
  alert('余额不足，可用 XXX USDT')
  return  // 阻止下单
}
```

### 卖出验证
```typescript
required = 数量
if (BTC余额 < required) {
  alert('余额不足，可用 XXX BTC')
  return  // 阻止下单
}
```

## 实时更新

余额数据通过 RTK Query 自动刷新：
```typescript
const { data: balances } = useGetBalancesQuery(undefined, {
  skip: !isAuthenticated,
  pollingInterval: 5000, // 每5秒刷新
});
```

### 何时更新
- ✅ 页面加载时
- ✅ 每5秒自动刷新
- ✅ 窗口获得焦点时
- ✅ 下单成功后（自动失效缓存）
- ✅ 充值/提现后（自动失效缓存）

## 数字格式

### 精度显示
```
BTC: 8位小数  1.50000000
ETH: 8位小数  10.12345678
USDT: 2位小数  10,000.00
```

### 自适应精度
```typescript
parseFloat(balance.available).toFixed(8)
```

## 用户体验优化

### 1. 快捷填充
点击"25%"、"50%"、"75%"、"100%"按钮（可扩展）：
```typescript
const handlePercentClick = (percent: number) => {
  const balance = getAvailableBalance();
  if (balance === '-') return;
  
  const available = parseFloat(balance);
  if (side === 'buy') {
    // 买入：根据余额和价格计算数量
    const qty = (available * percent) / parseFloat(price || '0');
    setQuantity(qty.toFixed(8));
  } else {
    // 卖出：直接使用余额百分比
    setQuantity((available * percent).toFixed(8));
  }
};
```

### 2. 余额不足提示
```typescript
if (余额不足) {
  显示红色提示：'余额不足，需要 XXX，可用 YYY'
  禁用下单按钮
}
```

### 3. 实时计算
输入价格或数量时，实时显示：
- 预计花费/获得
- 手续费估算
- 最终到账

## 示例场景

### 场景1：买入 BTC

```
用户余额：
  USDT: 50,000
  BTC: 0

下单界面：
[买入] [卖出]
限价 | 市价

价格：65,000
数量：0.5
---
可用 USDT: 50,000.00000000
预计花费: 32,500.00 USDT
---
[买入] ✅ 可以下单（余额充足）
```

### 场景2：卖出 BTC（余额不足）

```
用户余额：
  USDT: 50,000
  BTC: 0.3

下单界面：
[买入] [卖出]
限价 | 市价

价格：65,000
数量：0.5
---
可用 BTC: 0.30000000
预计获得: 32,500.00 USDT
---
❌ 余额不足，可用 0.3 BTC
[卖出] 按钮禁用或显示警告
```

## 代码位置

### 前端
- `components/OrderForm.tsx` - 余额显示和验证
- `hooks/useBalances.ts` - 余额数据获取（已删除，改用RTK Query）
- `lib/services/api.ts` - 余额API定义

### 后端
- `handlers/balance.go` - 余额查询接口
- `handlers/order.go` - 下单时的余额验证

## 测试步骤

1. **连接钱包并登录**
2. **充值测试资金**：
   - 进入资产页面
   - 充值 USDT: 100,000
   - 充值 BTC: 10
3. **查看下单表单**：
   - 买入时应显示 USDT 余额
   - 卖出时应显示 BTC 余额
4. **测试验证**：
   - 输入超过余额的数量
   - 应提示"余额不足"

现在余额会正确显示了！如果还没充值，显示会是 0 或 -。


