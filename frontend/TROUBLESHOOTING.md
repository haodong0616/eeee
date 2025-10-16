# 前端问题排查指南

## Tailwind CSS 样式不生效

如果页面显示混乱，Tailwind 样式没有生效，请按以下步骤操作：

### 方法 1：完整重置（推荐）

```bash
# 1. 删除 node_modules 和锁文件
rm -rf node_modules
rm -rf .next
rm package-lock.json yarn.lock

# 2. 重新安装依赖
npm install

# 3. 启动开发服务器
npm run dev
```

### 方法 2：快速修复

```bash
# 1. 清理缓存
rm -rf .next

# 2. 重启开发服务器
npm run dev
```

### 方法 3：检查端口冲突

如果 3000 端口被占用：

```bash
# 查找占用 3000 端口的进程
lsof -ti:3000

# 杀死进程
kill -9 $(lsof -ti:3000)

# 或者使用其他端口
npm run dev -- -p 3002
```

## 常见问题

### 1. 样式完全不显示

**症状**：页面是纯白色或黑色，没有任何样式

**解决方案**：
```bash
# 确保所有依赖都已安装
npm install tailwindcss postcss autoprefixer --save-dev

# 重启服务
npm run dev
```

### 2. 部分组件样式错误

**症状**：某些组件有样式，某些没有

**解决方案**：检查组件文件路径是否在 `tailwind.config.ts` 的 `content` 配置中。

当前配置：
```typescript
content: [
  "./pages/**/*.{js,ts,jsx,tsx,mdx}",
  "./components/**/*.{js,ts,jsx,tsx,mdx}",
  "./app/**/*.{js,ts,jsx,tsx,mdx}",
]
```

### 3. 自定义颜色不生效

**症状**：`text-primary`、`bg-buy` 等自定义颜色不显示

**解决方案**：这些颜色已在 `tailwind.config.ts` 中定义，重启开发服务器即可。

### 4. 热更新不工作

**症状**：修改代码后页面不自动刷新

**解决方案**：
```bash
# 重启开发服务器
# 按 Ctrl+C 停止，然后重新运行
npm run dev
```

## 验证 Tailwind 是否正常工作

访问 http://localhost:3000 后，打开浏览器开发者工具：

1. **检查元素** - 右键点击任意元素 → 检查
2. **查看样式** - 应该能看到 Tailwind 生成的 CSS 类
3. **Console** - 不应该有 CSS 相关的错误

正常情况下，你应该看到类似这样的类名：
- `min-h-screen`
- `bg-[#0a0e27]`
- `text-white`
- `flex`
- `items-center`

## 还是不行？

### 完全重新开始

```bash
# 1. 停止所有 Node 进程
pkill -f node

# 2. 清理所有缓存
rm -rf node_modules
rm -rf .next
rm -rf .npm
rm package-lock.json
rm yarn.lock

# 3. 清理 npm 缓存
npm cache clean --force

# 4. 重新安装
npm install

# 5. 启动
npm run dev
```

### 检查 Node 版本

```bash
node --version  # 应该是 v18 或更高
npm --version   # 应该是 v9 或更高
```

如果版本太低：
```bash
# 使用 nvm 升级
nvm install 18
nvm use 18
```

## 配置文件检查清单

确保以下文件存在且内容正确：

- ✅ `tailwind.config.ts` - Tailwind 配置
- ✅ `postcss.config.mjs` - PostCSS 配置
- ✅ `app/globals.css` - 全局样式
- ✅ `app/layout.tsx` - 根布局（导入 globals.css）
- ✅ `package.json` - 依赖配置

## 开发技巧

### 1. 使用浏览器扩展

安装 [Tailwind CSS IntelliSense](https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss) VS Code 扩展，获得更好的开发体验。

### 2. 实时查看生成的 CSS

```bash
# 查看 Tailwind 生成的完整 CSS
npx tailwindcss -o output.css --watch
```

### 3. 调试模式

在 `next.config.js` 中添加：
```javascript
module.exports = {
  // ... 其他配置
  compiler: {
    removeConsole: false,
  },
}
```

## 联系支持

如果以上方法都不行，请提供：
1. Node.js 版本 (`node --version`)
2. npm 版本 (`npm --version`)
3. 操作系统
4. 浏览器控制台的错误信息
5. 终端的错误信息

