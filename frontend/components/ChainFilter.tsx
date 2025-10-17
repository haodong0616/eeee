'use client';

import { useEffect, useRef, useMemo } from 'react';
import { useChainId } from 'wagmi';
import { useGetChainsQuery } from '@/lib/services/api';
import { showToast } from '@/hooks/useToast';

/**
 * 链过滤器 - 在后台检查用户当前链是否在启用列表中
 * 如果不在，提示用户切换
 */
export default function ChainFilter() {
  const chainId = useChainId();
  const { data: chains } = useGetChainsQuery();
  const lastCheckedChainId = useRef<number | undefined>();

  // 使用 useMemo 缓存链数据的序列化版本，作为稳定的依赖项
  const chainsKey = useMemo(() => {
    if (!chains || chains.length === 0) return '';
    return chains.map(c => `${c.chain_id}:${c.enabled ? '1' : '0'}`).join(',');
  }, [chains]);

  // 使用 useMemo 缓存启用的链信息
  const enabledChainsInfo = useMemo(() => {
    if (!chains || chains.length === 0) {
      return { ids: new Set<number>(), names: '' };
    }
    
    const enabledChains = chains.filter(c => c.enabled);
    return {
      ids: new Set(enabledChains.map(c => c.chain_id)),
      names: enabledChains.map(c => c.chain_name).join(', ')
    };
  }, [chainsKey]); // 只依赖序列化的 key

  useEffect(() => {
    // 只在 chainId 变化时检查
    if (!chainId || chainId === lastCheckedChainId.current) return;
    if (enabledChainsInfo.ids.size === 0) return; // 数据还未加载
    
    lastCheckedChainId.current = chainId;
    
    // 检查当前链是否在启用列表中
    if (!enabledChainsInfo.ids.has(chainId)) {
      showToast.warning(`当前链不支持，请切换到：${enabledChainsInfo.names}`);
    }
  }, [chainId, enabledChainsInfo]); // enabledChainsInfo 现在是稳定的

  return null;
}


