'use client';

import { useEffect } from 'react';
import { useChainId } from 'wagmi';
import { useChains } from '@/hooks/useChains';
import { useToast } from '@/hooks/useToast';

/**
 * 链过滤器 - 在后台检查用户当前链是否在启用列表中
 * 如果不在，提示用户切换
 */
export default function ChainFilter() {
  const chainId = useChainId();
  const { enabledChains, getChainById } = useChains();
  const toast = useToast();

  useEffect(() => {
    if (chainId && enabledChains.length > 0) {
      const currentChain = getChainById(chainId);
      
      // 如果当前链不在启用列表中
      if (!currentChain) {
        const enabledChainNames = enabledChains.map(c => c.chain_name).join(', ');
        toast.warning(`当前链不支持，请切换到：${enabledChainNames}`, { duration: 5000 });
      }
    }
  }, [chainId, enabledChains, getChainById]);

  return null; // 这是一个后台组件，不渲染UI
}

