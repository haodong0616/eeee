# 前端安装和运行指南

## 首次安装

```bash
cd frontend

# 1. 清理所有旧文件（如果存在）
rm -rf node_modules .next package-lock.json yarn.lock

# 2. 安装依赖
npm install

# 3. 启动开发服务器
npm run dev
```

## 访问地址

打开浏览器访问：http://localhost:3000

## 如果样式不显示

### 方法 1：完整重置
```bash
rm -rf node_modules .next
npm install
npm run dev
```

### 方法 2：只清理缓存
```bash
rm -rf .next
npm run dev
```

### 方法 3：强制重新构建
```bash
npm run dev -- --turbo
```

## 验证 Tailwind 是否工作

1. 打开浏览器开发者工具（F12）
2. 检查 Elements 标签中的元素
3. 应该看到类似这样的 Tailwind 类：
   - `min-h-screen`
   - `bg-[#0a0e27]`
   - `text-white`
   - `flex`

4. 查看 Styles 面板，应该有生成的 CSS

## 常见问题

### 页面是纯白色/黑色
**原因**：Tailwind CSS 没有正确加载

**解决**：
```bash
# 检查依赖
npm list tailwindcss
npm list postcss
npm list autoprefixer

# 如果缺失，重新安装
npm install -D tailwindcss@latest postcss@latest autoprefixer@latest
```

### 报错：Cannot find module 'tailwindcss'
**解决**：
```bash
npm install -D tailwindcss postcss autoprefixer
```

### 端口 3000 被占用
**解决**：
```bash
# 方法 1：杀死占用进程
lsof -ti:3000 | xargs kill -9

# 方法 2：使用其他端口
npm run dev -- -p 3002
```

### 样式只在刷新后显示
**原因**：热更新问题

**解决**：
```bash
# 重启开发服务器
# 按 Ctrl+C 停止
npm run dev
```

## 配置文件检查清单

确保以下文件存在且内容正确：

### ✅ `tailwind.config.ts`
```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "#3B82F6",
          dark: "#2563EB",
        },
        buy: "#10B981",
        sell: "#EF4444",
      },
    },
  },
  plugins: [],
};
export default config;
```

### ✅ `postcss.config.js`
```javascript
module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

### ✅ `app/globals.css`
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### ✅ `app/layout.tsx`
```typescript
import "./globals.css";
```

## 浏览器兼容性

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## 性能优化

开发模式下首次加载可能较慢，这是正常的。生产构建会更快：

```bash
npm run build
npm start
```

## 需要帮助？

1. 查看浏览器控制台错误
2. 查看终端错误输出
3. 检查 Node 版本：`node --version` (需要 v18+)
4. 查看 TROUBLESHOOTING.md

## 成功标志

当一切正常时，你应该看到：
- ✅ 深色背景（#0a0e27）
- ✅ 白色文字
- ✅ 蓝色按钮和链接
- ✅ 圆角卡片
- ✅ 导航栏样式正常

