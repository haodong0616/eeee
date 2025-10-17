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
            title={`切换链 (当前: ${chain.name})`}
            className="relative flex items-center justify-center w-8 h-8 md:w-9 md:h-9 border border-gray-700 rounded-full bg-[#151a35] hover:bg-[#1a1f3a] hover:border-primary transition-all group"
          >
            {chain.hasIcon && chain.iconUrl ? (
              <img
                alt={chain.name ?? 'Chain icon'}
                src={chain.iconUrl}
                className="w-5 h-5 rounded-full"
              />
            ) : (
              <svg className="w-5 h-5 text-gray-400 group-hover:text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            )}
            {/* 小指示器 - 桌面端显示 */}
            <span className="hidden md:block absolute -bottom-0.5 -right-0.5 w-2 h-2 bg-green-500 rounded-full border border-[#0f1429]"></span>
          </button>
        );
      }}
    </ConnectButton.Custom>
  );
}

