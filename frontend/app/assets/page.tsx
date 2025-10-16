'use client';

import { useState, useEffect } from 'react';
import { useAppSelector } from '@/lib/store/hooks';
import { useGetBalancesQuery, useDepositMutation, useWithdrawMutation, useGetDepositRecordsQuery, useGetWithdrawRecordsQuery } from '@/lib/services/api';
import { useRouter } from 'next/navigation';
import { useAccount, useChainId } from 'wagmi';
import { useWalletClient } from 'wagmi';
import { DepositService } from '@/lib/contracts/depositService';
import { ClockIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import { useToast } from '@/hooks/useToast';
import { useChains } from '@/hooks/useChains';

export default function AssetsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const { address } = useAccount();
  const { data: walletClient } = useWalletClient();
  const chainId = useChainId();
  const toast = useToast();
  const { getChainById } = useChains();
  
  // 使用 RTK Query 自动刷新余额
  const { data: balances = [], isLoading } = useGetBalancesQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 5000,
  });
  
  const [depositMutation] = useDepositMutation();
  const [withdrawMutation] = useWithdrawMutation();
  
  const [showDepositModal, setShowDepositModal] = useState(false);
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState('');
  const [amount, setAmount] = useState('');
  const [withdrawAddress, setWithdrawAddress] = useState('');
  const [processing, setProcessing] = useState(false);
  const [usdtBalance, setUsdtBalance] = useState('0');

  // 获取充值提现记录
  const { data: depositRecords = [] } = useGetDepositRecordsQuery(undefined, {
    skip: !isAuthenticated,
  });
  const { data: withdrawRecords = [] } = useGetWithdrawRecordsQuery(undefined, {
    skip: !isAuthenticated,
  });

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/');
    }
  }, [isAuthenticated, router]);

  // 获取钱包 USDT 余额
  useEffect(() => {
    if (walletClient && address && showDepositModal && selectedAsset === 'USDT' && chainId) {
      const chainConfig = getChainById(chainId);
      if (chainConfig) {
        DepositService.getUSDTBalance(walletClient, address, chainConfig).then(setUsdtBalance);
      }
    }
  }, [walletClient, address, showDepositModal, selectedAsset, chainId, getChainById]);

  const handleDeposit = async () => {
    if (!selectedAsset || !amount) {
      toast.error('请输入充值金额');
      return;
    }

    if (selectedAsset !== 'USDT') {
      toast.error('目前仅支持 USDT 充值');
      return;
    }

    if (!walletClient) {
      toast.error('请先连接钱包');
      return;
    }

    if (!chainId) {
      toast.error('请先选择网络');
      return;
    }
    
    // 获取链信息
    const chainConfig = getChainById(chainId);
    if (!chainConfig) {
      toast.error('不支持的链');
      return;
    }
    
    setProcessing(true);
    
    try {
      await toast.promise(
        (async () => {
          // 1. 调用合约转账
          console.log('📤 开始 USDT 转账...');
          console.log('链:', chainConfig.chain_name, 'ChainID:', chainId);
          const txHash = await DepositService.depositUSDT(walletClient, amount, chainConfig);
          console.log('✅ 转账成功，hash:', txHash);

          // 2. 提交到后端验证
          console.log('📡 提交充值记录到后端...');
          await depositMutation({ 
            asset: selectedAsset, 
            amount,
            txHash,
            chain: chainConfig.chain_name,
            chainId: chainConfig.chain_id
          }).unwrap();

          return txHash;
        })(),
        {
          loading: `正在 ${chainConfig.chain_name} 上处理充值交易...`,
          success: (txHash) => `充值交易已提交！\n链: ${chainConfig.chain_name}\n交易hash: ${txHash.slice(0, 10)}...\n后端正在验证，预计1-3分钟到账`,
          error: (err) => err?.message || err?.data?.error || '充值失败',
        }
      );
      
      setShowDepositModal(false);
      setAmount('');
    } catch (error: any) {
      console.error('❌ 充值失败:', error);
    } finally {
      setProcessing(false);
    }
  };

  const handleWithdraw = async () => {
    if (!selectedAsset || !amount || !withdrawAddress) {
      toast.error('请填写完整信息');
      return;
    }

    if (selectedAsset !== 'USDT') {
      toast.error('目前仅支持 USDT 提现');
      return;
    }

    // 验证地址格式
    if (!/^0x[a-fA-F0-9]{40}$/.test(withdrawAddress)) {
      toast.error('请输入正确的钱包地址');
      return;
    }

    if (!chainId) {
      toast.error('请先选择网络');
      return;
    }
    
    // 获取链信息
    const chainConfig = getChainById(chainId);
    if (!chainConfig) {
      toast.error('不支持的链');
      return;
    }
    
    setProcessing(true);
    
    try {
      await toast.promise(
        withdrawMutation({ 
          asset: selectedAsset, 
          amount, 
          address: withdrawAddress,
          chain: chainConfig.chain_name,
          chainId: chainConfig.chain_id
        }).unwrap(),
        {
          loading: `正在提交 ${chainConfig.chain_name} 提现申请...`,
          success: `提现申请已提交！\n链: ${chainConfig.chain_name}\n预计10-30分钟内到账，请注意查收`,
          error: (err) => err?.message || err?.data?.error || '提现失败',
        }
      );
      
      setShowWithdrawModal(false);
      setAmount('');
      setWithdrawAddress('');
    } catch (error: any) {
      console.error('❌ 提现失败:', error);
    } finally {
      setProcessing(false);
    }
  };

  const totalValueUSDT = balances.reduce((sum, balance) => {
    // 简化计算，实际需要根据当前价格计算
    return sum + parseFloat(balance.available || '0') + parseFloat(balance.frozen || '0');
  }, 0);

  // 格式化数字显示，USDT显示2位小数，其他显示8位
  const formatAmount = (amount: string | number, asset: string) => {
    const num = parseFloat(amount.toString());
    return asset === 'USDT' ? num.toFixed(2) : num.toFixed(8);
  };

  return (
    <div className="container mx-auto px-3 lg:px-4 py-4 lg:py-8">
      <h1 className="text-xl lg:text-3xl font-bold mb-4 lg:mb-8">我的资产</h1>

      {/* 总览 */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4 lg:p-6 mb-4 lg:mb-8">
        {/* 第一行：总资产估值 + 记录图标 */}
        <div className="flex items-center justify-between mb-4">
          <div>
            <p className="text-gray-400 text-sm lg:text-base mb-1 lg:mb-2">总资产估值 (USDT)</p>
            <p className="text-2xl lg:text-3xl font-bold">{totalValueUSDT.toFixed(2)}</p>
          </div>
          
          {/* 记录图标按钮 */}
          <Link
            href="/assets/records"
            className="p-2 lg:p-3 bg-[#151a35] hover:bg-[#1a1f3a] border border-gray-700 rounded-lg transition relative"
            title="充值/提现记录"
          >
            <ClockIcon className="w-5 h-5 lg:w-6 lg:h-6 text-gray-400" />
            {(depositRecords.length + withdrawRecords.length) > 0 && (
              <span className="absolute -top-1 -right-1 bg-primary text-white text-xs w-5 h-5 rounded-full flex items-center justify-center">
                {depositRecords.length + withdrawRecords.length}
              </span>
            )}
          </Link>
        </div>

        {/* 第二行：充值和提现按钮 */}
        <div className="flex gap-3 lg:gap-4">
          <button
            onClick={() => {
              setSelectedAsset('USDT');
              setShowDepositModal(true);
            }}
            className="flex-1 px-4 lg:px-6 py-2.5 text-sm lg:text-base bg-primary hover:bg-primary-dark rounded-lg transition font-semibold"
          >
            充值
          </button>
          <button
            onClick={() => {
              setSelectedAsset('USDT');
              setShowWithdrawModal(true);
            }}
            className="flex-1 px-4 lg:px-6 py-2.5 text-sm lg:text-base bg-gray-700 hover:bg-gray-600 rounded-lg transition font-semibold"
          >
            提现
          </button>
        </div>
      </div>

      {/* 余额列表 - 桌面端表格 */}
      <div className="hidden lg:block bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
        <div className="grid grid-cols-4 gap-4 p-4 bg-[#151a35] text-gray-400 text-sm font-semibold">
          <div>币种</div>
          <div className="text-right">总计</div>
          <div className="text-right">可用</div>
          <div className="text-right">冻结</div>
        </div>
        {isLoading ? (
          <div className="p-8 text-center text-gray-400">加载中...</div>
        ) : balances.length === 0 ? (
          <div className="p-8 text-center text-gray-400">暂无资产</div>
        ) : (
          balances.map((balance) => {
            const total = parseFloat(balance.available) + parseFloat(balance.frozen);
            return (
              <div
                key={balance.id}
                className="grid grid-cols-4 gap-4 p-4 border-t border-gray-800 hover:bg-[#151a35] transition"
              >
                <div className="font-semibold">{balance.asset}</div>
                <div className="text-right font-semibold">
                  {formatAmount(total, balance.asset)}
                </div>
                <div className="text-right text-gray-400">{formatAmount(balance.available, balance.asset)}</div>
                <div className="text-right text-gray-400">{formatAmount(balance.frozen, balance.asset)}</div>
              </div>
            );
          })
        )}
      </div>

      {/* 余额列表 - 移动端卡片 */}
      <div className="lg:hidden space-y-3">
        {isLoading ? (
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-8 text-center text-gray-400">
            加载中...
          </div>
        ) : balances.length === 0 ? (
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-8 text-center text-gray-400">
            暂无资产
          </div>
        ) : (
          balances.map((balance) => {
            const total = parseFloat(balance.available) + parseFloat(balance.frozen);
            return (
              <div
                key={balance.id}
                className="bg-[#0f1429] rounded-lg border border-gray-800 p-4"
              >
                {/* 币种和总计 */}
                <div className="flex items-center justify-between mb-3">
                  <div className="text-lg font-bold">{balance.asset}</div>
                  <div className="text-right">
                    <div className="text-xs text-gray-400 mb-1">总计</div>
                    <div className="text-lg font-semibold">
                      {formatAmount(total, balance.asset)}
                    </div>
                  </div>
                </div>

                {/* 可用和冻结 */}
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div className="bg-[#151a35] rounded-lg p-3">
                    <div className="text-gray-400 mb-1">可用</div>
                    <div className="font-mono">{formatAmount(balance.available, balance.asset)}</div>
                  </div>
                  <div className="bg-[#151a35] rounded-lg p-3">
                    <div className="text-gray-400 mb-1">冻结</div>
                    <div className="font-mono">{formatAmount(balance.frozen, balance.asset)}</div>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* 充值模态框 */}
      {showDepositModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-4 lg:p-6 w-full max-w-md border border-gray-800">
            <h2 className="text-lg lg:text-xl font-bold mb-4">充值 {selectedAsset}</h2>
            
            {selectedAsset === 'USDT' && (
              <div className="mb-4 p-3 bg-[#151a35] rounded-lg border border-gray-700">
                <div className="text-xs text-gray-400 mb-1">钱包余额</div>
                <div className="text-sm font-semibold">{parseFloat(usdtBalance).toFixed(4)} USDT</div>
              </div>
            )}
            
            <div className="mb-4">
              <label className="block text-sm text-gray-400 mb-2">充值金额</label>
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full px-3 lg:px-4 py-2 text-sm lg:text-base bg-[#151a35] border border-gray-700 rounded-lg focus:outline-none focus:border-primary"
                placeholder="输入充值金额"
                disabled={processing}
                min="0"
                step="0.01"
              />
              <div className="text-xs text-gray-400 mt-2">
                💡 将通过智能合约转账到平台地址
              </div>
            </div>
            
            <div className="flex gap-3 lg:gap-4">
              <button
                onClick={handleDeposit}
                disabled={processing || !amount}
                className="flex-1 px-4 py-2 text-sm lg:text-base bg-primary hover:bg-primary-dark rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {processing ? '处理中...' : '确认充值'}
              </button>
              <button
                onClick={() => {
                  setShowDepositModal(false);
                  setAmount('');
                }}
                disabled={processing}
                className="flex-1 px-4 py-2 text-sm lg:text-base bg-gray-700 hover:bg-gray-600 rounded-lg transition disabled:opacity-50"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 提现模态框 */}
      {showWithdrawModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-4 lg:p-6 w-full max-w-md border border-gray-800">
            <h2 className="text-lg lg:text-xl font-bold mb-4">提现 {selectedAsset}</h2>
            <div className="mb-4">
              <label className="block text-sm text-gray-400 mb-2">提现地址</label>
              <input
                type="text"
                value={withdrawAddress}
                onChange={(e) => setWithdrawAddress(e.target.value)}
                className="w-full px-3 lg:px-4 py-2 text-sm lg:text-base bg-[#151a35] border border-gray-700 rounded-lg focus:outline-none focus:border-primary font-mono"
                placeholder="0x..."
                disabled={processing}
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm text-gray-400 mb-2">提现金额</label>
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full px-3 lg:px-4 py-2 text-sm lg:text-base bg-[#151a35] border border-gray-700 rounded-lg focus:outline-none focus:border-primary"
                placeholder="输入提现金额"
                disabled={processing}
                min="0"
                step="0.01"
              />
             
            </div>
            <div className="flex gap-3 lg:gap-4">
              <button
                onClick={handleWithdraw}
                disabled={processing || !amount || !withdrawAddress}
                className="flex-1 px-4 py-2 text-sm lg:text-base bg-primary hover:bg-primary-dark rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {processing ? '处理中...' : '确认提现'}
              </button>
              <button
                onClick={() => {
                  setShowWithdrawModal(false);
                  setAmount('');
                  setWithdrawAddress('');
                }}
                disabled={processing}
                className="flex-1 px-4 py-2 text-sm lg:text-base bg-gray-700 hover:bg-gray-600 rounded-lg transition disabled:opacity-50"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
