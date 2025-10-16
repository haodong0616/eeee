'use client';

import Link from 'next/link';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { logout, setAuth } from '@/lib/store/slices/authSlice';
import { useGetNonceMutation, useLoginMutation } from '@/lib/services/api';
import { useState, useEffect } from 'react';
import { ConnectButton } from '@rainbow-me/rainbowkit';
import { useAccount, useSignMessage, useDisconnect } from 'wagmi';

export default function Header() {
  const dispatch = useAppDispatch();
  const { isAuthenticated, user } = useAppSelector((state) => state.auth);
  const [loggingIn, setLoggingIn] = useState(false);

  // 使用 wagmi 的钱包连接状态
  const { address: wagmiAddress, isConnected: wagmiIsConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const { disconnect } = useDisconnect();

  const [getNonce] = useGetNonceMutation();
  const [login] = useLoginMutation();

  // 当钱包连接且未认证时，自动触发登录
  useEffect(() => {
    if (wagmiIsConnected && wagmiAddress && !isAuthenticated && !loggingIn) {
      handleAutoLogin(wagmiAddress);
    }
  }, [wagmiIsConnected, wagmiAddress, isAuthenticated]);

  const handleAutoLogin = async (walletAddress: string) => {
    if (loggingIn) return;
    setLoggingIn(true);

    try {
      const lowerAddress = walletAddress.toLowerCase();
      console.log('🔐 开始登录流程，地址:', lowerAddress);
      
      // 获取nonce
      console.log('📡 获取 nonce...');
      const { nonce } = await getNonce(lowerAddress).unwrap();
      console.log('✅ 获取 nonce 成功:', nonce);

      // 签名
      const message = `登录到 Velocity Exchange\nNonce: ${nonce}`;
      console.log('✍️ 请求签名，消息:', message);
      const signature = await signMessageAsync({ message });
      console.log('✅ 签名成功:', signature?.slice(0, 10) + '...');

      if (!signature) {
        console.warn('⚠️ 签名为空');
        setLoggingIn(false);
        return;
      }

      // 登录
      console.log('🔑 提交登录请求...');
      const result = await login({ 
        walletAddress: lowerAddress, 
        signature 
      }).unwrap();
      
      dispatch(setAuth({ user: result.user, token: result.token }));
      
      console.log('✅ 登录成功！');
      alert('✅ 登录成功！');
    } catch (error: any) {
      console.error('❌ 登录失败详情:', error);
      
      let errorMsg = '登录失败';
      if (error.name === 'UserRejectedRequestError') {
        errorMsg = '用户取消了签名';
      } else if (error?.data?.error) {
        errorMsg = error.data.error;
      } else if (error?.message) {
        errorMsg = error.message;
      }
      
      alert(errorMsg);
    } finally {
      setLoggingIn(false);
    }
  };

  const handleLogout = () => {
    dispatch(logout());
    disconnect(); // 断开钱包连接
    window.location.href = '/';
  };

  return (
    <header className="bg-[#0f1429] border-b border-gray-800">
      <nav className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-8">
            <Link href="/" className="text-xl font-bold flex items-center gap-2 group">
              <span className="text-primary group-hover:scale-110 transition-transform">⚡</span>
              <span className="bg-gradient-to-r from-primary to-purple-400 bg-clip-text text-transparent">Velocity</span>
              {/* <span className="text-gray-400 text-base">Exchange</span> */}
            </Link>
            <div className="hidden md:flex space-x-6">
              <Link href="/" className="hover:text-primary transition">
                首页
              </Link>
              <Link href="/markets" className="hover:text-primary transition">
                行情
              </Link>
              <Link href="/trade/LUNAR-USDT" className="hover:text-primary transition">
                交易
              </Link>
              {isAuthenticated && (
                <Link href="/assets" className="hover:text-primary transition">
                  资产
                </Link>
              )}
            </div>
            
            {/* Mobile Menu */}
            <div className="md:hidden flex space-x-3 text-sm">
              <Link href="/markets" className="hover:text-primary transition">
                行情
              </Link>
              {isAuthenticated && (
                <Link href="/assets" className="hover:text-primary transition">
                  资产
                </Link>
              )}
            </div>
          </div>

          <div className="rainbow-wallet-btn">
            {isAuthenticated ? (
              <button
                onClick={handleLogout}
                className="px-3 md:px-4 py-1.5 md:py-2 text-sm md:text-base bg-red-600 hover:bg-red-700 rounded-lg transition"
              >
                退出
              </button>
            ) : (
              <ConnectButton.Custom>
                {({ account, chain, openConnectModal, mounted }) => {
                  return (
                    <div
                      {...(!mounted && {
                        'aria-hidden': true,
                        style: {
                          opacity: 0,
                          pointerEvents: 'none',
                          userSelect: 'none',
                        },
                      })}
                    >
                      <button
                        onClick={openConnectModal}
                        disabled={!mounted || loggingIn}
                        className="px-4 md:px-6 py-1.5 md:py-2 text-sm md:text-base bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white rounded-lg font-semibold transition shadow-lg hover:shadow-purple-500/50 disabled:opacity-50"
                      >
                        {loggingIn ? '登录中...' : '连接钱包'}
                      </button>
                    </div>
                  );
                }}
              </ConnectButton.Custom>
            )}
          </div>
        </div>
      </nav>
    </header>
  );
}
