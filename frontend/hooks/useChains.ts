import { useMemo, useCallback } from 'react';
import { useGetChainsQuery, type ChainConfig } from '@/lib/services/api';

export function useChains() {
  const { data: chains = [], isLoading, error } = useGetChainsQuery();
  
  // 使用 useMemo 缓存计算结果，避免每次渲染都创建新数组
  const enabledChains = useMemo(() => {
    return chains
      .filter((chain: ChainConfig) => chain.enabled)
      .sort((a: ChainConfig, b: ChainConfig) => a.chain_id - b.chain_id);
  }, [chains]);

  const getChainById = useCallback((chainId: number): ChainConfig | undefined => {
    return chains.find((chain: ChainConfig) => chain.chain_id === chainId);
  }, [chains]);

  const getChainByName = useCallback((chainName: string): ChainConfig | undefined => {
    return chains.find((chain: ChainConfig) => chain.chain_name === chainName);
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

