# RTK Query 使用指南

## 为什么选择 RTK Query？

✅ **与 Redux 完美集成** - 无需额外的状态管理  
✅ **自动缓存管理** - 智能缓存和失效  
✅ **自动轮询** - 支持定时刷新  
✅ **乐观更新** - 提升用户体验  
✅ **类型安全** - 完整的 TypeScript 支持  
✅ **无需 axios** - 内置 fetch  
✅ **统一的错误处理**  

## 项目架构

### API 定义 (`lib/services/api.ts`)

所有接口都在一个文件中定义：

```typescript
export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: 'http://localhost:8080/api',
    prepareHeaders: (headers) => {
      const token = localStorage.getItem('token');
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['TradingPairs', 'Tickers', 'Orders', 'Balances', ...],
  endpoints: (builder) => ({
    // 定义所有接口
  }),
});
```

### Store 配置 (`lib/store/store.ts`)

```typescript
export const store = configureStore({
  reducer: {
    [api.reducerPath]: api.reducer, // RTK Query
    auth: authReducer,               // 认证状态
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(api.middleware),
});

setupListeners(store.dispatch); // 启用自动刷新
```

## 接口使用

### 查询（Query）

#### 基本用法
```typescript
const { data, isLoading, error } = useGetAllTickersQuery();
```

#### 带轮询
```typescript
const { data: tickers } = useGetAllTickersQuery(undefined, {
  pollingInterval: 5000, // 每5秒自动刷新
});
```

#### 带参数
```typescript
const { data: ticker } = useGetTickerQuery('BTC/USDT', {
  pollingInterval: 3000,
});
```

#### 条件查询
```typescript
const { data: balances } = useGetBalancesQuery(undefined, {
  skip: !isAuthenticated, // 未登录时不请求
  pollingInterval: 5000,
});
```

### 变更（Mutation）

#### 基本用法
```typescript
const [createOrder, { isLoading }] = useCreateOrderMutation();

const handleSubmit = async () => {
  try {
    await createOrder(orderData).unwrap();
    alert('成功');
  } catch (error) {
    alert('失败');
  }
};
```

#### 认证流程
```typescript
const [getNonce] = useGetNonceMutation();
const [login] = useLoginMutation();

// 1. 获取 nonce
const { nonce } = await getNonce(walletAddress).unwrap();

// 2. 签名
const signature = await signMessage(message);

// 3. 登录
const result = await login({ walletAddress, signature }).unwrap();
dispatch(setAuth(result));
```

## 所有可用的 Hooks

### 认证相关
```typescript
useGetNonceMutation()      // 获取 nonce
useLoginMutation()         // 登录
useGetProfileQuery()       // 获取用户信息
```

### 市场数据
```typescript
useGetTradingPairsQuery()           // 获取交易对列表
useGetTickerQuery(symbol)           // 获取单个行情
useGetAllTickersQuery()             // 获取所有行情
useGetOrderBookQuery(symbol)        // 获取盘口数据
useGetRecentTradesQuery(symbol)     // 获取最近成交
useGetKlinesQuery({ symbol, interval }) // 获取K线数据
```

### 订单相关
```typescript
useCreateOrderMutation()            // 创建订单
useGetOrdersQuery({ symbol, status }) // 获取订单列表
useGetOrderQuery(orderId)           // 获取单个订单
useCancelOrderMutation()            // 取消订单
```

### 余额相关
```typescript
useGetBalancesQuery()               // 获取所有余额
useGetBalanceQuery(asset)           // 获取单个资产余额
useDepositMutation()                // 充值
useWithdrawMutation()               // 提现
```

## 轮询间隔配置

| 数据类型 | 轮询间隔 | 原因 |
|---------|---------|------|
| 盘口数据 | 2秒 | 最实时 |
| 行情数据 | 3秒 | 较实时 |
| 成交记录 | 3秒 | 较实时 |
| 订单列表 | 5秒 | 中等频率 |
| 余额数据 | 5秒 | 中等频率 |
| K线数据 | 10秒 | 较慢 |
| 交易对列表 | 不轮询 | 静态数据 |

## 缓存和失效

### 自动缓存
RTK Query 会自动缓存所有请求结果：

```typescript
// 第一次请求 - 从服务器获取
useGetTickerQuery('BTC/USDT');

// 第二次请求（其他组件）- 从缓存获取
useGetTickerQuery('BTC/USDT');
```

### 自动失效

Mutation 会自动失效相关缓存：

```typescript
// 创建订单后，自动重新获取订单列表和余额
createOrder: builder.mutation({
  invalidatesTags: ['Orders', 'Balances'],
});
```

### 手动刷新

```typescript
const { refetch } = useGetBalancesQuery();

// 手动刷新
const handleRefresh = () => {
  refetch();
};
```

## 组件示例

### 首页
```typescript
export default function Home() {
  const { data: tickers = [] } = useGetAllTickersQuery(undefined, {
    pollingInterval: 5000,
  });

  return (
    <div>
      {tickers.map(ticker => (
        <div key={ticker.symbol}>{ticker.last_price}</div>
      ))}
    </div>
  );
}
```

### 交易页面
```typescript
export default function TradePage() {
  const { data: ticker } = useGetTickerQuery(symbol, {
    pollingInterval: 3000,
  });
  const { data: orderBook } = useGetOrderBookQuery(symbol, {
    pollingInterval: 2000,
  });
  const [createOrder] = useCreateOrderMutation();

  const handleOrder = async (orderData) => {
    await createOrder({ ...orderData, symbol }).unwrap();
  };

  return <div>...</div>;
}
```

### 资产页面
```typescript
export default function AssetsPage() {
  const { data: balances = [] } = useGetBalancesQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 5000,
  });
  const [deposit] = useDepositMutation();

  const handleDeposit = async () => {
    await deposit({ asset, amount }).unwrap();
  };

  return <div>...</div>;
}
```

## 错误处理

### 统一错误处理
```typescript
try {
  await createOrder(data).unwrap();
  alert('成功');
} catch (error: any) {
  // RTK Query 错误格式
  alert(error?.data?.error || '操作失败');
}
```

### Loading 状态
```typescript
const { data, isLoading, isFetching, error } = useGetBalancesQuery();

if (isLoading) return <div>首次加载...</div>;
if (error) return <div>加载失败</div>;
if (isFetching) return <div>刷新中...（后台）</div>;
```

## 性能优化

### 1. 标签系统
```typescript
providesTags: ['Balances'],        // 提供标签
invalidatesTags: ['Balances'],     // 失效标签
```

### 2. 自动去重
相同参数的请求会自动合并：
```typescript
// 这两个请求只会发送一次
useGetTickerQuery('BTC/USDT');
useGetTickerQuery('BTC/USDT');
```

### 3. 选择性订阅
```typescript
const { data } = useGetBalancesQuery(undefined, {
  skip: !isAuthenticated, // 条件查询
  selectFromResult: ({ data }) => ({
    data: data?.filter(b => b.available > 0), // 过滤数据
  }),
});
```

## 对比其他方案

### vs Axios + useEffect
```typescript
// ❌ Axios + useEffect
useEffect(() => {
  fetchData();
  const interval = setInterval(fetchData, 5000);
  return () => clearInterval(interval);
}, []);

// ✅ RTK Query
const { data } = useGetDataQuery(undefined, {
  pollingInterval: 5000,
});
```

### vs SWR
```typescript
// SWR - 需要额外依赖
const { data } = useSWR(key, fetcher, { refreshInterval: 5000 });

// RTK Query - 已包含在 Redux Toolkit
const { data } = useGetDataQuery(undefined, { pollingInterval: 5000 });
```

### 优势
- ✅ 无需额外依赖（axios, swr）
- ✅ 与 Redux 完美集成
- ✅ 更好的类型推导
- ✅ 统一的状态管理
- ✅ 自动请求去重
- ✅ 乐观更新支持

## 迁移完成清单

- ✅ 删除 axios 依赖
- ✅ 删除 swr 依赖
- ✅ 删除所有 API 文件 (`lib/api/`)
- ✅ 删除旧的 Redux slices (market, order, balance)
- ✅ 创建统一的 RTK Query API
- ✅ 更新所有组件使用新 hooks
- ✅ 配置 Store 和 middleware
- ✅ 添加自动轮询
- ✅ 添加标签系统

## 调试技巧

### Redux DevTools
安装 Redux DevTools 扩展，可以看到：
- 所有 API 请求
- 缓存状态
- 标签失效记录
- 查询状态

### 查看缓存
```typescript
import { api } from '@/lib/services/api';

// 在组件中
const cacheState = api.endpoints.getBalances.select()(state);
console.log('Cache:', cacheState);
```

## 常见问题

### Q: 如何停止轮询？
A: 设置 `pollingInterval: 0` 或不设置

### Q: 如何手动刷新？
A: 使用 `refetch()` 方法

### Q: 如何清除缓存？
A: 使用 `dispatch(api.util.resetApiState())`

### Q: 数据更新太频繁？
A: 调整 `pollingInterval` 或使用 `skip`

## 最佳实践

1. **合理设置轮询间隔** - 根据数据实时性需求
2. **使用条件查询** - `skip` 避免不必要的请求
3. **统一错误处理** - 捕获 `.unwrap()` 异常
4. **标签管理** - 合理设计标签系统
5. **类型定义** - 充分利用 TypeScript

## 性能对比

| 指标 | Axios + useEffect | SWR | RTK Query |
|------|------------------|-----|-----------|
| 依赖大小 | ~1MB | ~50KB | 0 (已包含) |
| 缓存 | ❌ | ✅ | ✅ |
| 自动刷新 | ❌ | ✅ | ✅ |
| Redux 集成 | ❌ | ❌ | ✅ |
| 请求去重 | ❌ | ✅ | ✅ |
| 类型推导 | 一般 | 好 | 优秀 |

RTK Query 是最适合这个项目的方案！🎉

