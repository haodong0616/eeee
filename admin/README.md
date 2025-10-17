# ExpChange 管理后台

基于 Next.js 14 开发的交易所管理后台。

## 功能模块

### 1. 登录页面
- 用户名密码登录
- Token 认证
- 自动跳转

默认账号：
- 用户名：`admin`
- 密码：`admin`

### 2. 数据概览
- 用户总数
- 订单总数
- 成交总数
- 交易总量
- 系统信息

### 3. 用户管理
- 用户列表
- 钱包地址查询
- 注册时间显示

### 4. 订单管理
- 所有订单列表
- 状态筛选
- 用户订单关联
- 实时更新

### 5. 成交记录
- 成交历史
- 交易详情
- 成交金额统计

### 6. 交易对管理
- 交易对列表
- 创建新交易对
- 启用/禁用交易对
- 参数配置

## 页面路由

- `/` - 重定向
- `/login` - 登录页面
- `/dashboard` - 数据概览
- `/dashboard/users` - 用户管理
- `/dashboard/orders` - 订单管理
- `/dashboard/trades` - 成交记录
- `/dashboard/pairs` - 交易对管理

## API 接口

所有接口需要携带管理员 Token：

```typescript
Headers: {
  Authorization: Bearer {admin_token}
}
```

### 获取统计数据
```
GET /api/admin/stats
```

### 获取用户列表
```
GET /api/admin/users
```

### 获取订单列表
```
GET /api/admin/orders
```

### 获取成交记录
```
GET /api/admin/trades
```

### 创建交易对
```
POST /api/admin/pairs
Body: {
  symbol: string,
  base_asset: string,
  quote_asset: string,
  min_price?: string,
  max_price?: string,
  min_qty?: string,
  max_qty?: string
}
```

### 更新交易对状态
```
PUT /api/admin/pairs/:id/status
Body: status=active|inactive
```

## 布局结构

```
Layout
├── Sidebar
│   ├── Logo
│   ├── Navigation
│   └── Logout Button
└── Main Content
    └── Page Component
```

## 导航菜单

- 📊 概览 - `/dashboard`
- 👥 用户管理 - `/dashboard/users`
- 📋 订单管理 - `/dashboard/orders`
- 💱 成交记录 - `/dashboard/trades`
- ⚙️ 交易对管理 - `/dashboard/pairs`

## 开发指南

### 添加新页面

1. 在 `app/dashboard/` 下创建目录
2. 创建 `page.tsx` 文件
3. 在布局中添加导航链接

### 添加新 API

1. 在 `lib/api/admin.ts` 中定义接口
2. 创建类型定义
3. 在页面中调用

### 样式定制

使用 Tailwind CSS：

```tsx
<div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
  // 内容
</div>
```

## 数据刷新

大部分数据在页面加载时获取，可以添加定时刷新：

```typescript
useEffect(() => {
  loadData();
  const interval = setInterval(loadData, 5000);
  return () => clearInterval(interval);
}, []);
```

## 权限控制

当前实现简化版权限：
- 登录验证
- Token 过期自动跳转

生产环境建议添加：
- 角色权限
- 操作日志
- IP 白名单
- 二次验证

## 部署说明

### 开发环境
```bash
npm run dev
```

### 生产环境
```bash
npm run build
npm start
```

### Docker 部署
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3001
CMD ["npm", "start"]
```

## 环境变量

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 安全建议

1. 使用强密码
2. 启用 HTTPS
3. 添加 CSRF 保护
4. 实现操作日志
5. 设置会话超时
6. 限制登录尝试
7. 使用环境变量管理敏感信息

## 性能优化

- 使用服务端渲染
- 数据分页加载
- 图表懒加载
- 缓存静态资源

## 浏览器兼容性

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+


