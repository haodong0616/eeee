# React 无限循环问题修复

## 🐛 问题描述

控制台出现大量错误：

```
Warning: Maximum update depth exceeded. This can happen when a component 
calls setState inside useEffect, but useEffect either doesn't have a 
dependency array, or one of the dependencies changes on every render.
```

错误来源：`useChains.ts:13` 和 `ChainFilter.tsx:20`

## 🔍 问题原因

### 根本原因：useEffect 依赖项引起的无限循环

在 `ChainFilter.tsx` 中：

```typescript
// ❌ 问题代码
useEffect(() => {
  if (chainId && enabledChains.length > 0) {
    const currentChain = getChainById(chainId);
    // ...
  }
}, [chainId, enabledChains, getChainById]); // 🔥 getChainById 每次都是新函数！
```

**问题链路**：
1. `getChainById` 在 `useChains` hook 中每次渲染都会重新创建
2. `ChainFilter` 的 `useEffect` 依赖 `getChainById`
3. `getChainById` 变化 → `useEffect` 触发 → 可能导致状态更新
4. 状态更新 → 重新渲染 → `getChainById` 又是新函数 → 无限循环 ♻️

### 触发条件

- 组件挂载时
- 任何导致 `useChains` 重新执行的状态变化
- Toast 显示也可能触发重新渲染

## ✅ 解决方案

### 1. 使用 `useCallback` 包裹函数（推荐）

在 `useChains.ts` 中：

```typescript
// ✅ 修复后
import { useCallback } from 'react';

const getChainById = useCallback((chainId: number): ChainConfig | undefined => {
  return chains.find((chain: ChainConfig) => chain.chain_id === chainId);
}, [chains]); // 只在 chains 变化时重新创建

const getChainByName = useCallback((chainName: string): ChainConfig | undefined => {
  return chains.find((chain: ChainConfig) => chain.chain_name === chainName);
}, [chains]);
```

**好处**：
- 函数引用稳定，只在依赖项（`chains`）变化时更新
- 可以安全地放在 `useEffect` 依赖数组中
- 不影响其他使用该 hook 的组件

### 2. 优化 Toast 使用

在 `ChainFilter.tsx` 中：

```typescript
// ✅ 修复后
import { showToast } from '@/hooks/useToast'; // 使用静态方法而非 hook
import { useRef } from 'react';

const hasShownWarning = useRef(false); // 防止重复显示

useEffect(() => {
  if (chainId && enabledChains.length > 0) {
    const currentChain = getChainById(chainId);
    
    // 只在未显示过警告时显示
    if (!currentChain && !hasShownWarning.current) {
      showToast.warning(`当前链不支持，请切换到：${enabledChainNames}`);
      hasShownWarning.current = true;
    }
    
    // 切换到支持的链时重置
    if (currentChain) {
      hasShownWarning.current = false;
    }
  }
}, [chainId, enabledChains, getChainById]); // 现在 getChainById 是稳定的了
```

**好处**：
- 使用 `showToast` 静态方法，不需要在依赖数组中
- 使用 `useRef` 避免重复显示 toast
- 更好的用户体验

## 📝 修改的文件

### 1. `frontend/hooks/useChains.ts`
- ✅ 导入 `useCallback`
- ✅ 用 `useCallback` 包裹 `getChainById`
- ✅ 用 `useCallback` 包裹 `getChainByName`

### 2. `frontend/components/ChainFilter.tsx`
- ✅ 导入 `useRef`
- ✅ 改用 `showToast` 静态方法
- ✅ 添加 `hasShownWarning` ref 防止重复显示
- ✅ 添加重置逻辑

## 🎯 React 性能优化最佳实践

### 1. 何时使用 `useCallback`

**应该使用**：
- 函数作为 `useEffect`、`useMemo` 的依赖项时
- 函数作为 props 传递给使用 `React.memo` 的子组件时
- 函数在每次渲染时重新创建会影响性能时

**不需要使用**：
- 简单的事件处理函数（如 `onClick`）
- 不作为依赖项使用的函数
- 性能影响可忽略的场景

### 2. 何时使用 `useMemo`

```typescript
// 适合使用 useMemo 的场景
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(a, b);
}, [a, b]);

// 不需要 useMemo 的场景
const simpleValue = a + b; // 简单计算，直接使用
```

### 3. 何时使用 `useRef`

**适用场景**：
- 存储不影响渲染的值（如标志位、定时器 ID）
- 访问 DOM 元素
- 保存上一次的值

```typescript
// ✅ 好用法：防止重复操作
const hasExecuted = useRef(false);

// ❌ 不要用 useState：会触发重新渲染
const [hasExecuted, setHasExecuted] = useState(false);
```

### 4. useEffect 依赖数组规则

```typescript
// ✅ 正确：所有使用的变量都在依赖数组中
useEffect(() => {
  doSomething(value);
}, [value]);

// ❌ 错误：遗漏依赖项（会有警告）
useEffect(() => {
  doSomething(value);
}, []); // ESLint 会警告

// ✅ 正确：函数用 useCallback 包裹
const doSomething = useCallback(() => {
  // ...
}, [dependencies]);

useEffect(() => {
  doSomething();
}, [doSomething]); // 安全
```

## 🧪 验证修复

修复后，应该看到：
- ✅ 控制台没有 "Maximum update depth exceeded" 错误
- ✅ 页面正常渲染，没有卡顿
- ✅ Toast 只显示一次
- ✅ 链切换功能正常工作

## 📚 相关资源

- [React useCallback 文档](https://react.dev/reference/react/useCallback)
- [React useEffect 文档](https://react.dev/reference/react/useEffect)
- [React Performance Optimization](https://react.dev/learn/render-and-commit)

## 💡 预防措施

1. **ESLint 规则**：启用 `react-hooks/exhaustive-deps` 规则
2. **代码审查**：注意 `useEffect` 依赖项中的函数和对象
3. **性能监控**：使用 React DevTools Profiler 检查重复渲染
4. **组件设计**：尽量让组件依赖少量稳定的 props

