'use client';

import { useEffect } from 'react';
import { useChainId } from 'wagmi';
import { ConnectButton } from '@rainbow-me/rainbowkit';

export default function ChainSwitcher() {
  const chainId = useChainId();

  // 同步钱包的当前链到 localStorage 和触发事件
  useEffect(() => {
    if (chainId) {
      localStorage.setItem('selected_chain_id', chainId.toString());
      
      // 触发自定义事件，通知其他组件链已切换
      window.dispatchEvent(new CustomEvent('chainChanged', { 
        detail: chainId 
      }));
    }
  }, [chainId]);

  return (
    <ConnectButton.Custom>
      {({ account, chain, openChainModal, mounted }) => {
        // 如果没有连接钱包或未挂载，不显示
        if (!mounted || !account || !chain) {
          return null;
        }

        return (
          <button
            onClick={openChainModal}
            type="button"
            className="flex items-center gap-2 px-3 py-2 text-sm border border-gray-700 rounded-lg bg-[#151a35] hover:bg-[#1a1f3a] text-white transition-colors"
          >
            {chain.hasIcon && chain.iconUrl && (
              <img
                alt={chain.name ?? 'Chain icon'}
                src={chain.iconUrl}
                className="w-4 h-4 rounded-full"
              />
            )}
            <span className="font-medium">{chain.name}</span>
            <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        );
      }}
    </ConnectButton.Custom>
  );
}

