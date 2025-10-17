'use client';

import { useState, useEffect } from 'react';
import { useAppSelector } from '@/lib/store/hooks';
import { useGetDepositRecordsQuery, useGetWithdrawRecordsQuery } from '@/lib/services/api';
import { useRouter } from 'next/navigation';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useChains } from '@/hooks/useChains';
import { formatQuantity } from '@/lib/utils/format';

// 禁用静态生成，因为此页面需要认证
export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';
export const revalidate = 0;

export default function RecordsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const [activeTab, setActiveTab] = useState<'deposit' | 'withdraw'>('deposit');
  const { getChainById } = useChains();

  // 获取充值提现记录
  const { data: depositRecords = [] } = useGetDepositRecordsQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 10000,
  });
  const { data: withdrawRecords = [] } = useGetWithdrawRecordsQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 10000,
  });

  // 格式化数字显示，根据币种类型使用不同精度
  const formatAmount = (amount: string | number, asset: string) => {
    const num = parseFloat(amount.toString());
    
    // USDT 显示2位小数
    if (asset === 'USDT') {
      return num.toFixed(2);
    }
    
    // 其他币种使用 formatQuantity，传入虚拟交易对格式
    return formatQuantity(num, `${asset}/USDT`);
  };

  // 使用 useEffect 来处理重定向，避免服务器端渲染时出错
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/');
    }
  }, [isAuthenticated, router]);

  // 如果未认证，不渲染内容
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="container mx-auto px-3 lg:px-4 py-4 lg:py-8">
      {/* 标题和返回按钮 */}
      <div className="flex items-center gap-3 mb-4 lg:mb-8">
        <button
          onClick={() => router.back()}
          className="p-2 hover:bg-[#151a35] rounded-lg transition"
        >
          <ArrowLeftIcon className="w-5 h-5" />
        </button>
        <h1 className="text-xl lg:text-3xl font-bold">交易记录</h1>
      </div>

      {/* Tabs */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
        <div className="flex border-b border-gray-800">
          <button
            onClick={() => setActiveTab('deposit')}
            className={`flex-1 px-4 py-2 lg:px-6 lg:py-3 text-sm lg:text-base transition ${
              activeTab === 'deposit'
                ? 'text-primary border-b-2 border-primary bg-primary/5'
                : 'text-gray-400 hover:bg-[#151a35]'
            }`}
          >
            充值记录 {depositRecords.length > 0 && `(${depositRecords.length})`}
          </button>
          <button
            onClick={() => setActiveTab('withdraw')}
            className={`flex-1 px-4 py-2 lg:px-6 lg:py-3 text-sm lg:text-base transition ${
              activeTab === 'withdraw'
                ? 'text-primary border-b-2 border-primary bg-primary/5'
                : 'text-gray-400 hover:bg-[#151a35]'
            }`}
          >
            提现记录 {withdrawRecords.length > 0 && `(${withdrawRecords.length})`}
          </button>
        </div>

        {/* 记录列表 */}
        <div className="p-3 lg:p-6">
          {activeTab === 'deposit' ? (
            depositRecords.length === 0 ? (
              <div className="text-center py-12 text-gray-400">暂无充值记录</div>
            ) : (
              <div className="space-y-2 lg:space-y-3">
                {depositRecords.map((record: any) => {
                  const chainConfig = getChainById(record.chain_id);
                  const explorerUrl = chainConfig?.block_explorer_url || 'https://bscscan.com';
                  
                  return (
                    <div
                      key={record.id}
                      className="bg-[#151a35] rounded-lg p-3 lg:p-4 border border-gray-800 hover:border-gray-700 transition"
                    >
                      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-2 lg:gap-3">
                        {/* 左侧：金额和资产 */}
                        <div className="flex items-center gap-3 lg:gap-4">
                          <div>
                            <div className="text-xs lg:text-sm text-gray-400 mb-0.5 lg:mb-1">充值金额</div>
                            <div className="text-lg lg:text-xl font-bold">{formatAmount(record.amount, record.asset)} {record.asset}</div>
                            {/* 链信息 */}
                            {chainConfig && (
                              <div className="text-xs text-gray-500 mt-1">
                                {chainConfig.chain_name}
                              </div>
                            )}
                          </div>
                        </div>

                        {/* 中间：交易信息 */}
                        <div className="flex-1 text-xs lg:text-sm space-y-0.5 lg:space-y-1">
                          <div className="flex items-center gap-2 text-gray-400">
                            <span>Hash:</span>
                            <a
                              href={`${explorerUrl}/tx/${record.tx_hash}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary hover:underline font-mono"
                            >
                              {record.tx_hash.slice(0, 8)}...{record.tx_hash.slice(-6)}
                            </a>
                          </div>
                          <div className="text-gray-400">
                            {new Date(record.created_at).toLocaleString('zh-CN', { 
                              month: '2-digit', 
                              day: '2-digit', 
                              hour: '2-digit', 
                              minute: '2-digit' 
                            })}
                          </div>
                        </div>

                        {/* 右侧：状态 */}
                        <div>
                          <span className={`px-2 py-1 lg:px-3 lg:py-1.5 rounded text-xs lg:text-sm ${
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
                    </div>
                  );
                })}
              </div>
            )
          ) : (
            withdrawRecords.length === 0 ? (
              <div className="text-center py-12 text-gray-400">暂无提现记录</div>
            ) : (
              <div className="space-y-2 lg:space-y-3">
                {withdrawRecords.map((record: any) => {
                  const chainConfig = getChainById(record.chain_id);
                  const explorerUrl = chainConfig?.block_explorer_url || 'https://bscscan.com';
                  
                  return (
                    <div
                      key={record.id}
                      className="bg-[#151a35] rounded-lg p-3 lg:p-4 border border-gray-800 hover:border-gray-700 transition"
                    >
                      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-2 lg:gap-3">
                        {/* 左侧：金额和资产 */}
                        <div className="flex items-center gap-3 lg:gap-4">
                          <div>
                            <div className="text-xs lg:text-sm text-gray-400 mb-0.5 lg:mb-1">提现金额</div>
                            <div className="text-lg lg:text-xl font-bold">{formatAmount(record.amount, record.asset)} {record.asset}</div>
                            {/* 链信息 */}
                            {chainConfig && (
                              <div className="text-xs text-gray-500 mt-1">
                                {chainConfig.chain_name}
                              </div>
                            )}
                          </div>
                        </div>

                        {/* 中间：交易信息 */}
                        <div className="flex-1 text-xs lg:text-sm space-y-0.5 lg:space-y-1">
                          <div className="flex items-center gap-2 text-gray-400">
                            <span>地址:</span>
                            <span className="font-mono">{record.address.slice(0, 8)}...{record.address.slice(-6)}</span>
                          </div>
                          {record.tx_hash && (
                            <div className="flex items-center gap-2 text-gray-400">
                              <span>Hash:</span>
                              <a
                                href={`${explorerUrl}/tx/${record.tx_hash}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-primary hover:underline font-mono"
                              >
                                {record.tx_hash.slice(0, 8)}...{record.tx_hash.slice(-6)}
                              </a>
                            </div>
                          )}
                          <div className="text-gray-400">
                            {new Date(record.created_at).toLocaleString('zh-CN', { 
                              month: '2-digit', 
                              day: '2-digit', 
                              hour: '2-digit', 
                              minute: '2-digit' 
                            })}
                          </div>
                        </div>

                        {/* 右侧：状态 */}
                        <div>
                          <span className={`px-2 py-1 lg:px-3 lg:py-1.5 rounded text-xs lg:text-sm whitespace-nowrap ${
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
                    </div>
                  );
                })}
              </div>
            )
          )}
        </div>
      </div>
    </div>
  );
}


