import { useEffect, useState } from 'react';
import { useGetChainsQuery, type ChainConfig } from '@/lib/services/api';

export function useChains() {
  const { data: chains = [], isLoading, error } = useGetChainsQuery();
  const [enabledChains, setEnabledChains] = useState<ChainConfig[]>([]);

  useEffect(() => {
    // 过滤出启用的链，并按Chain ID排序
    const enabled = chains
      .filter((chain: ChainConfig) => chain.enabled)
      .sort((a: ChainConfig, b: ChainConfig) => a.chain_id - b.chain_id);
    setEnabledChains(enabled);
  }, [chains]);

  const getChainById = (chainId: number): ChainConfig | undefined => {
    return chains.find((chain: ChainConfig) => chain.chain_id === chainId);
  };

  const getChainByName = (chainName: string): ChainConfig | undefined => {
    return chains.find((chain: ChainConfig) => chain.chain_name === chainName);
  };

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

