# ExpChange 用户前端

基于 Next.js 14 开发的加密货币交易所用户端。

## 功能模块

### 1. 钱包连接
- 支持 MetaMask 钱包连接
- 签名登录验证
- 自动保持登录状态

### 2. 首页
- 热门交易对展示
- 24小时涨跌榜
- 平台特点介绍

### 3. 行情页面
- 所有交易对列表
- 实时价格更新
- 搜索过滤功能

### 4. 交易页面
- 实时盘口深度
- 最近成交记录
- K线图显示区域
- 限价/市价下单
- 当前委托管理
- 历史委托查看

### 5. 资产页面
- 资产总览
- 余额列表
- 充值/提现功能

## 技术实现

### Redux Store 结构

```
store/
├── slices/
│   └── authSlice.ts      # 认证状态
└── store.ts              # Store 配置 + RTK Query
```

### RTK Query API 封装

```
lib/services/
└── api.ts           # 统一的 RTK Query API
                     # 包含所有接口定义：
                     # - 认证接口
                     # - 市场数据接口
                     # - 订单接口
                     # - 余额接口
```

### WebSocket 连接

```typescript
import { wsClient } from '@/lib/websocket';

// 连接
wsClient.connect();

// 监听事件
wsClient.on('trade', (data) => {
  console.log('New trade:', data);
});

// 断开连接
wsClient.disconnect();
```

## 页面路由

- `/` - 首页
- `/markets` - 行情页面
- `/trade/[symbol]` - 交易页面
- `/assets` - 资产页面

## 组件说明

### Header
导航栏组件，包含钱包连接功能。

### OrderBook
显示盘口深度，实时更新买卖盘。

### TradeHistory
显示最近成交记录。

### OrderForm
下单表单，支持限价和市价单。

### MyOrders
显示用户的当前委托和历史委托。

## 样式设计

使用 Tailwind CSS 和 Flexbox 布局：
- 深色主题
- 响应式设计
- 输入框聚焦发光效果
- 平滑过渡动画

## 开发指南

### 添加新的交易对

1. 在后端创建交易对
2. 前端自动获取并显示

### 自定义主题颜色

编辑 `tailwind.config.ts`：

```typescript
colors: {
  primary: "#3B82F6",
  buy: "#10B981",
  sell: "#EF4444",
}
```

### 添加新的 API 接口

1. 在 `lib/api/` 下创建接口定义
2. 在 Redux slice 中创建 thunk
3. 在组件中使用 `useAppDispatch` 调用

## 环境变量

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 构建部署

```bash
# 开发环境
npm run dev

# 生产构建
npm run build
npm start
```

## 钱包集成

### MetaMask 连接流程

1. 检测 `window.ethereum`
2. 请求账户访问
3. 获取钱包地址
4. 获取 nonce
5. 签名消息
6. 提交登录

### 签名消息格式

```
登录到 ExpChange
Nonce: {random_nonce}
```

## 性能优化

- 使用 Next.js 14 App Router
- 服务端渲染（SSR）
- 图片优化
- 代码分割
- 懒加载组件

## 浏览器兼容性

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

需要支持 MetaMask 扩展。

