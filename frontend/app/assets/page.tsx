'use client';

import { useState, useEffect } from 'react';
import { useAppSelector } from '@/lib/store/hooks';
import { useGetBalancesQuery, useDepositMutation, useWithdrawMutation, useGetDepositRecordsQuery, useGetWithdrawRecordsQuery } from '@/lib/services/api';
import { useRouter } from 'next/navigation';
import { useAccount } from 'wagmi';
import { useWalletClient } from 'wagmi';
import { DepositService } from '@/lib/contracts/depositService';
import { ClockIcon } from '@heroicons/react/24/outline';

export default function AssetsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const { address } = useAccount();
  const { data: walletClient } = useWalletClient();
  
  // 使用 RTK Query 自动刷新余额
  const { data: balances = [], isLoading } = useGetBalancesQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 5000,
  });
  
  const [depositMutation] = useDepositMutation();
  const [withdrawMutation] = useWithdrawMutation();
  
  const [showDepositModal, setShowDepositModal] = useState(false);
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);
  const [showDepositRecords, setShowDepositRecords] = useState(false);
  const [showWithdrawRecords, setShowWithdrawRecords] = useState(false);
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
    if (walletClient && address && showDepositModal && selectedAsset === 'USDT') {
      DepositService.getUSDTBalance(walletClient, address).then(setUsdtBalance);
    }
  }, [walletClient, address, showDepositModal, selectedAsset]);

  const handleDeposit = async () => {
    if (!selectedAsset || !amount) {
      alert('请输入充值金额');
      return;
    }

    if (selectedAsset !== 'USDT') {
      alert('目前仅支持 USDT 充值');
      return;
    }

    if (!walletClient) {
      alert('请先连接钱包');
      return;
    }
    
    setProcessing(true);
    try {
      // 1. 调用合约转账
      console.log('📤 开始 USDT 转账...');
      const txHash = await DepositService.depositUSDT(walletClient, amount);
      console.log('✅ 转账成功，hash:', txHash);

      // 2. 提交到后端验证
      console.log('📡 提交充值记录到后端...');
      await depositMutation({ 
        asset: selectedAsset, 
        amount,
        txHash 
      }).unwrap();

      console.log('✅ 充值记录已提交，等待确认...');
      alert('充值交易已提交！\n交易hash: ' + txHash.slice(0, 10) + '...\n\n后端正在验证交易，预计1-3分钟后到账');
      
      setShowDepositModal(false);
      setAmount('');
    } catch (error: any) {
      console.error('❌ 充值失败:', error);
      alert(error?.message || error?.data?.error || '充值失败');
    } finally {
      setProcessing(false);
    }
  };

  const handleWithdraw = async () => {
    if (!selectedAsset || !amount || !withdrawAddress) {
      alert('请填写完整信息');
      return;
    }

    if (selectedAsset !== 'USDT') {
      alert('目前仅支持 USDT 提现');
      return;
    }

    // 验证地址格式
    if (!/^0x[a-fA-F0-9]{40}$/.test(withdrawAddress)) {
      alert('请输入正确的钱包地址');
      return;
    }
    
    setProcessing(true);
    try {
      await withdrawMutation({ 
        asset: selectedAsset, 
        amount, 
        address: withdrawAddress 
      }).unwrap();
      
      alert('提现申请已提交！\n预计10-30分钟内到账，请注意查收');
      setShowWithdrawModal(false);
      setAmount('');
      setWithdrawAddress('');
    } catch (error: any) {
      console.error('❌ 提现失败:', error);
      alert(error?.message || error?.data?.error || '提现失败');
    } finally {
      setProcessing(false);
    }
  };

  const totalValueUSDT = balances.reduce((sum, balance) => {
    // 简化计算，实际需要根据当前价格计算
    return sum + parseFloat(balance.available || '0') + parseFloat(balance.frozen || '0');
  }, 0);

  return (
    <div className="container mx-auto px-3 lg:px-4 py-4 lg:py-8">
      <h1 className="text-xl lg:text-3xl font-bold mb-4 lg:mb-8">我的资产</h1>

      {/* 总览 */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4 lg:p-6 mb-4 lg:mb-8">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <p className="text-gray-400 text-sm lg:text-base mb-1 lg:mb-2">总资产估值 (USDT)</p>
            <p className="text-2xl lg:text-3xl font-bold">{totalValueUSDT.toFixed(2)}</p>
          </div>
          <div className="flex gap-2 lg:gap-4">
            <button
              onClick={() => {
                setSelectedAsset('USDT');
                setShowDepositModal(true);
              }}
              className="flex-1 lg:flex-none px-4 lg:px-6 py-2 text-sm lg:text-base bg-primary hover:bg-primary-dark rounded-lg transition"
            >
              充值
            </button>
            <button
              onClick={() => {
                setSelectedAsset('USDT');
                setShowWithdrawModal(true);
              }}
              className="flex-1 lg:flex-none px-4 lg:px-6 py-2 text-sm lg:text-base bg-gray-700 hover:bg-gray-600 rounded-lg transition"
            >
              提现
            </button>
          </div>
        </div>

        {/* 充值提现记录入口 */}
        <div className="flex gap-3 lg:gap-4 mt-4 pt-4 border-t border-gray-800">
          <button
            onClick={() => setShowDepositRecords(true)}
            className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-sm bg-[#151a35] hover:bg-[#1a1f3a] border border-gray-700 rounded-lg transition"
          >
            <ClockIcon className="w-4 h-4" />
            <span>充值记录</span>
            {depositRecords.length > 0 && (
              <span className="bg-primary/20 text-primary text-xs px-2 py-0.5 rounded-full">
                {depositRecords.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setShowWithdrawRecords(true)}
            className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-sm bg-[#151a35] hover:bg-[#1a1f3a] border border-gray-700 rounded-lg transition"
          >
            <ClockIcon className="w-4 h-4" />
            <span>提现记录</span>
            {withdrawRecords.length > 0 && (
              <span className="bg-primary/20 text-primary text-xs px-2 py-0.5 rounded-full">
                {withdrawRecords.length}
              </span>
            )}
          </button>
        </div>
      </div>

      {/* 余额列表 - 桌面端表格 */}
      <div className="hidden lg:block bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
        <div className="grid grid-cols-5 gap-4 p-4 bg-[#151a35] text-gray-400 text-sm font-semibold">
          <div>币种</div>
          <div className="text-right">可用</div>
          <div className="text-right">冻结</div>
          <div className="text-right">总计</div>
          <div className="text-right">操作</div>
        </div>
        {isLoading ? (
          <div className="p-8 text-center text-gray-400">加载中...</div>
        ) : balances.length === 0 ? (
          <div className="p-8 text-center text-gray-400">暂无资产</div>
        ) : (
          balances.map((balance) => (
            <div
              key={balance.id}
              className="grid grid-cols-5 gap-4 p-4 border-t border-gray-800 hover:bg-[#151a35] transition"
            >
              <div className="font-semibold">{balance.asset}</div>
              <div className="text-right">{parseFloat(balance.available).toFixed(8)}</div>
              <div className="text-right">{parseFloat(balance.frozen).toFixed(8)}</div>
              <div className="text-right font-semibold">
                {(parseFloat(balance.available) + parseFloat(balance.frozen)).toFixed(8)}
              </div>
              <div className="text-right space-x-2">
                <button
                  onClick={() => {
                    setSelectedAsset(balance.asset);
                    setShowDepositModal(true);
                  }}
                  className="text-primary hover:underline"
                >
                  充值
                </button>
                <button
                  onClick={() => {
                    setSelectedAsset(balance.asset);
                    setShowWithdrawModal(true);
                  }}
                  className="text-gray-400 hover:underline"
                >
                  提现
                </button>
              </div>
            </div>
          ))
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
          balances.map((balance) => (
            <div
              key={balance.id}
              className="bg-[#0f1429] rounded-lg border border-gray-800 p-4"
            >
              {/* 币种和总计 */}
              <div className="flex items-center justify-between mb-3 pb-3 border-b border-gray-800">
                <div className="text-lg font-bold">{balance.asset}</div>
                <div className="text-right">
                  <div className="text-xs text-gray-400 mb-1">总计</div>
                  <div className="text-base font-semibold">
                    {(parseFloat(balance.available) + parseFloat(balance.frozen)).toFixed(8)}
                  </div>
                </div>
              </div>

              {/* 可用和冻结 */}
              <div className="grid grid-cols-2 gap-3 mb-3 text-sm">
                <div>
                  <div className="text-gray-400 mb-1">可用</div>
                  <div className="font-mono">{parseFloat(balance.available).toFixed(8)}</div>
                </div>
                <div>
                  <div className="text-gray-400 mb-1">冻结</div>
                  <div className="font-mono">{parseFloat(balance.frozen).toFixed(8)}</div>
                </div>
              </div>

              {/* 操作按钮 */}
              <div className="flex gap-2">
                <button
                  onClick={() => {
                    setSelectedAsset(balance.asset);
                    setShowDepositModal(true);
                  }}
                  className="flex-1 py-2 text-sm bg-primary hover:bg-primary-dark rounded-lg transition"
                >
                  充值
                </button>
                <button
                  onClick={() => {
                    setSelectedAsset(balance.asset);
                    setShowWithdrawModal(true);
                  }}
                  className="flex-1 py-2 text-sm bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  提现
                </button>
              </div>
            </div>
          ))
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
              <div className="text-xs text-gray-400 mt-2">
                ⚠️ 提现将冻结资金，审核通过后自动转账
              </div>
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

      {/* 充值记录模态框 */}
      {showDepositRecords && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-4 lg:p-6 w-full max-w-2xl border border-gray-800 max-h-[80vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg lg:text-xl font-bold">充值记录</h2>
              <button
                onClick={() => setShowDepositRecords(false)}
                className="text-gray-400 hover:text-white text-2xl"
              >
                ×
              </button>
            </div>

            {depositRecords.length === 0 ? (
              <div className="text-center py-8 text-gray-400">暂无充值记录</div>
            ) : (
              <div className="space-y-3">
                {depositRecords.map((record: any) => (
                  <div
                    key={record.id}
                    className="bg-[#151a35] rounded-lg p-3 lg:p-4 border border-gray-800"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <div className="text-sm text-gray-400">充值金额</div>
                        <div className="text-lg font-semibold">{parseFloat(record.amount).toFixed(4)} {record.asset}</div>
                      </div>
                      <div>
                        <span className={`px-2 py-1 rounded text-xs ${
                          record.status === 'confirmed' 
                            ? 'bg-green-500/20 text-green-400'
                            : record.status === 'pending'
                            ? 'bg-yellow-500/20 text-yellow-400'
                            : 'bg-red-500/20 text-red-400'
                        }`}>
                          {record.status === 'confirmed' ? '已确认' : record.status === 'pending' ? '待确认' : '失败'}
                        </span>
                      </div>
                    </div>
                    
                    <div className="text-xs text-gray-400 space-y-1">
                      <div className="flex items-center gap-2">
                        <span>交易Hash:</span>
                        <a
                          href={`https://bscscan.com/tx/${record.tx_hash}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline font-mono"
                        >
                          {record.tx_hash.slice(0, 10)}...{record.tx_hash.slice(-8)}
                        </a>
                      </div>
                      <div>时间: {new Date(record.created_at).toLocaleString('zh-CN')}</div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* 提现记录模态框 */}
      {showWithdrawRecords && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-4 lg:p-6 w-full max-w-2xl border border-gray-800 max-h-[80vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg lg:text-xl font-bold">提现记录</h2>
              <button
                onClick={() => setShowWithdrawRecords(false)}
                className="text-gray-400 hover:text-white text-2xl"
              >
                ×
              </button>
            </div>

            {withdrawRecords.length === 0 ? (
              <div className="text-center py-8 text-gray-400">暂无提现记录</div>
            ) : (
              <div className="space-y-3">
                {withdrawRecords.map((record: any) => (
                  <div
                    key={record.id}
                    className="bg-[#151a35] rounded-lg p-3 lg:p-4 border border-gray-800"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <div className="text-sm text-gray-400">提现金额</div>
                        <div className="text-lg font-semibold">{parseFloat(record.amount).toFixed(4)} {record.asset}</div>
                      </div>
                      <div>
                        <span className={`px-2 py-1 rounded text-xs ${
                          record.status === 'completed' 
                            ? 'bg-green-500/20 text-green-400'
                            : record.status === 'pending'
                            ? 'bg-yellow-500/20 text-yellow-400'
                            : record.status === 'processing'
                            ? 'bg-blue-500/20 text-blue-400'
                            : 'bg-red-500/20 text-red-400'
                        }`}>
                          {record.status === 'completed' 
                            ? '已完成' 
                            : record.status === 'pending' 
                            ? '待处理' 
                            : record.status === 'processing'
                            ? '处理中'
                            : '失败'}
                        </span>
                      </div>
                    </div>
                    
                    <div className="text-xs text-gray-400 space-y-1">
                      <div className="flex items-center gap-2">
                        <span>提现地址:</span>
                        <span className="font-mono">{record.address.slice(0, 10)}...{record.address.slice(-8)}</span>
                      </div>
                      {record.tx_hash && (
                        <div className="flex items-center gap-2">
                          <span>交易Hash:</span>
                          <a
                            href={`https://bscscan.com/tx/${record.tx_hash}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary hover:underline font-mono"
                          >
                            {record.tx_hash.slice(0, 10)}...{record.tx_hash.slice(-8)}
                          </a>
                        </div>
                      )}
                      <div>时间: {new Date(record.created_at).toLocaleString('zh-CN')}</div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
