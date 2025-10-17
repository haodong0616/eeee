# useMemo 解决 React 无限循环问题

## ✅ 最终解决方案

使用 **两层 useMemo** 来稳定依赖项：

### 核心思路

```typescript
// 第一层：将数组序列化为字符串（稳定的基本类型）
const chainsKey = useMemo(() => {
  if (!chains) return '';
  return chains.map(c => `${c.chain_id}:${c.enabled ? '1' : '0'}`).join(',');
}, [chains]); // 依赖原始数组

// 第二层：基于稳定的字符串 key 计算实际需要的数据
const enabledChainsInfo = useMemo(() => {
  if (!chains) return { ids: new Set<number>(), names: '' };
  
  const enabledChains = chains.filter(c => c.enabled);
  return {
    ids: new Set(enabledChains.map(c => c.chain_id)),
    names: enabledChains.map(c => c.chain_name).join(', ')
  };
}, [chainsKey]); // 依赖序列化的 key，而不是原始数组

// useEffect 现在有稳定的依赖项
useEffect(() => {
  // ...逻辑
}, [chainId, enabledChainsInfo]); // enabledChainsInfo 只在内容真正变化时更新
```

## 🔍 为什么这样有效？

### 问题根源
```typescript
// ❌ 问题代码
const { data: chains } = useGetChainsQuery();

useEffect(() => {
  // chains 每次可能都是新的数组引用
  // 即使内容完全相同，[] !== [] （引用不同）
}, [chains]); // 每次都触发！
```

### 解决方案原理

**第一层 useMemo** (`chainsKey`):
- 输入：数组（引用类型，可能每次都是新的）
- 输出：字符串（基本类型，内容相同则相等）
- 作用：`"56:1,97:1"` 这样的字符串，内容相同时 `===` 比较为 `true`

**第二层 useMemo** (`enabledChainsInfo`):
- 输入：稳定的字符串 `chainsKey`
- 输出：对象（但只在 key 变化时重新创建）
- 作用：确保下游只在数据真正变化时才更新

## 🎯 完整实现

### frontend/components/ChainFilter.tsx

```typescript
'use client';

import { useEffect, useRef, useMemo } from 'react';
import { useChainId } from 'wagmi';
import { useGetChainsQuery } from '@/lib/services/api';
import { showToast } from '@/hooks/useToast';

export default function ChainFilter() {
  const chainId = useChainId();
  const { data: chains } = useGetChainsQuery();
  const lastCheckedChainId = useRef<number | undefined>();

  // 🔑 关键：两层 useMemo
  const chainsKey = useMemo(() => {
    if (!chains || chains.length === 0) return '';
    return chains.map(c => `${c.chain_id}:${c.enabled ? '1' : '0'}`).join(',');
  }, [chains]);

  const enabledChainsInfo = useMemo(() => {
    if (!chains || chains.length === 0) {
      return { ids: new Set<number>(), names: '' };
    }
    
    const enabledChains = chains.filter(c => c.enabled);
    return {
      ids: new Set(enabledChains.map(c => c.chain_id)),
      names: enabledChains.map(c => c.chain_name).join(', ')
    };
  }, [chainsKey]); // ✅ 只依赖字符串 key

  useEffect(() => {
    if (!chainId || chainId === lastCheckedChainId.current) return;
    if (enabledChainsInfo.ids.size === 0) return;
    
    lastCheckedChainId.current = chainId;
    
    if (!enabledChainsInfo.ids.has(chainId)) {
      showToast.warning(`当前链不支持，请切换到：${enabledChainsInfo.names}`);
    }
  }, [chainId, enabledChainsInfo]); // ✅ 稳定的依赖项

  return null;
}
```

### frontend/hooks/useChains.ts

```typescript
import { useMemo, useCallback } from 'react';
import { useGetChainsQuery, type ChainConfig } from '@/lib/services/api';

export function useChains() {
  const { data: chains = [], isLoading, error } = useGetChainsQuery();
  
  // 使用 useMemo 代替 useState + useEffect
  const enabledChains = useMemo(() => {
    return chains
      .filter((chain: ChainConfig) => chain.enabled)
      .sort((a: ChainConfig, b: ChainConfig) => a.chain_id - b.chain_id);
  }, [chains]);

  // 使用 useCallback 稳定函数引用
  const getChainById = useCallback((chainId: number) => {
    return chains.find(chain => chain.chain_id === chainId);
  }, [chains]);

  const getChainByName = useCallback((chainName: string) => {
    return chains.find(chain => chain.chain_name === chainName);
  }, [chains]);

  return {
    chains,
    enabledChains,
    isLoading,
    error,
    getChainById,
    getChainByName,
    hasMultipleChains: enabledChains.length > 1,
    hasSingleChain: enabledChains.length === 1,
    singleChain: enabledChains.length === 1 ? enabledChains[0] : null,
  };
}
```

## 📊 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **两层 useMemo** | ✅ 完全稳定<br>✅ 性能好<br>✅ 优雅 | 需要理解原理 | ⭐ **推荐** |
| useCallback | ✅ 简单 | 只能稳定函数 | 函数依赖 |
| useRef | ✅ 完全控制 | 代码复杂 | 复杂场景 |
| JSON.stringify | ✅ 简单直观 | 性能差 | 小数据量 |
| 禁用组件 | ✅ 最快修复 | 失去功能 | 临时方案 |

## 🧠 核心知识点

### 1. JavaScript 引用比较

```javascript
// 基本类型：按值比较
"abc" === "abc"  // ✅ true
123 === 123      // ✅ true

// 引用类型：按引用比较
[] === []                    // ❌ false
{} === {}                    // ❌ false
[1,2] === [1,2]             // ❌ false
new Set([1]) === new Set([1]) // ❌ false
```

### 2. useMemo 的作用

```typescript
// 没有 useMemo（每次渲染都创建新数组）
const filtered = data.filter(x => x.enabled); // 新数组引用

// 有 useMemo（引用稳定）
const filtered = useMemo(
  () => data.filter(x => x.enabled),
  [data] // 只在 data 变化时重新计算
);
```

### 3. 序列化技巧

```typescript
// 方法 1：自定义序列化（推荐，性能好）
const key = useMemo(() => 
  items.map(x => `${x.id}:${x.name}`).join(','),
  [items]
);

// 方法 2：JSON.stringify（简单，但慢）
const key = useMemo(() => 
  JSON.stringify(items),
  [items]
);

// 方法 3：哈希值（复杂数据）
const key = useMemo(() => 
  hashCode(JSON.stringify(items)),
  [items]
);
```

## 🎓 最佳实践

### ✅ DO（推荐做法）

1. **优先使用 useMemo 缓存计算结果**
   ```typescript
   const sorted = useMemo(() => data.sort(), [data]);
   ```

2. **用 useCallback 包裹事件处理函数**
   ```typescript
   const onClick = useCallback(() => {}, [deps]);
   ```

3. **将数组/对象转换为稳定的基本类型**
   ```typescript
   const key = useMemo(() => items.map(x => x.id).join(','), [items]);
   ```

4. **使用 useRef 存储不影响渲染的值**
   ```typescript
   const lastValue = useRef();
   ```

### ❌ DON'T（避免的做法）

1. **不要忘记依赖项**
   ```typescript
   // ❌ 错误
   useMemo(() => expensive(value), []); // value 变化不会更新
   
   // ✅ 正确
   useMemo(() => expensive(value), [value]);
   ```

2. **不要过度使用 useMemo**
   ```typescript
   // ❌ 不需要
   const simple = useMemo(() => a + b, [a, b]);
   
   // ✅ 直接计算
   const simple = a + b;
   ```

3. **不要在 useMemo 中修改外部状态**
   ```typescript
   // ❌ 错误
   useMemo(() => {
     setCount(count + 1); // 副作用！
     return data;
   }, [data]);
   
   // ✅ 使用 useEffect
   useEffect(() => {
     setCount(count + 1);
   }, [data]);
   ```

## 🔧 调试技巧

### 1. 添加日志查看更新频率

```typescript
const value = useMemo(() => {
  console.log('🔄 Recomputing value');
  return expensiveCalculation();
}, [deps]);
```

### 2. 使用 React DevTools Profiler

- 打开 Chrome DevTools → React → Profiler
- 记录一次交互
- 查看组件重渲染次数

### 3. 使用 why-did-you-render

```bash
npm install @welldone-software/why-did-you-render
```

```typescript
import whyDidYouRender from '@welldone-software/why-did-you-render';

if (process.env.NODE_ENV === 'development') {
  whyDidYouRender(React, {
    trackAllPureComponents: true,
  });
}
```

## 📚 延伸阅读

- [React useMemo 文档](https://react.dev/reference/react/useMemo)
- [React 性能优化指南](https://react.dev/learn/render-and-commit)
- [useCallback vs useMemo](https://kentcdodds.com/blog/usememo-and-usecallback)
- [React 渲染行为](https://blog.isquaredsoftware.com/2020/05/blogged-answers-a-mostly-complete-guide-to-react-rendering-behavior/)

## ✨ 总结

通过 **两层 useMemo** 模式，我们成功地：
1. ✅ 消除了无限循环
2. ✅ 保持了组件功能
3. ✅ 提升了性能
4. ✅ 写出了优雅的代码

关键是理解 **引用稳定性** 的重要性，并善用 React Hooks 来管理它！

