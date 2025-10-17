'use client';

import { useMemo } from 'react';
import useSWR from 'swr';
import { adminApi, type WithdrawRecord } from '@/lib/api/admin';
import { getChains } from '@/lib/api/admin';

export default function WithdrawalsPage() {
  const { data: withdrawals = [], isLoading, error, mutate } = useSWR(
    '/admin/withdrawals',
    () => adminApi.getWithdrawals(),
    {
      refreshInterval: 10000, // 每10秒自动刷新
    }
  );

  const { data: chains = [] } = useSWR('/admin/chains', getChains);

  // 创建链ID到链配置的映射
  const chainMap = useMemo(() => {
    const map = new Map<number, any>();
    chains.forEach((chain: any) => map.set(chain.chain_id, chain));
    return map;
  }, [chains]);

  const getStatusBadge = (status: string) => {
    const styles = {
      pending: 'bg-yellow-500/20 text-yellow-500',
      processing: 'bg-blue-500/20 text-blue-500',
      completed: 'bg-green-500/20 text-green-500',
      failed: 'bg-red-500/20 text-red-500',
    };
    return styles[status as keyof typeof styles] || 'bg-gray-700 text-gray-400';
  };

  const getStatusText = (status: string) => {
    const text = {
      pending: '待处理',
      processing: '处理中',
      completed: '已完成',
      failed: '失败',
    };
    return text[status as keyof typeof text] || status;
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">提现记录</h1>
        <div className="flex gap-3 items-center">
          {error && (
            <span className="text-red-500 text-sm">
              加载失败: {error.message || '未知错误'}
            </span>
          )}
          <span className="text-gray-400 text-sm">
            共 {withdrawals.length} 条记录
          </span>
          <button
            onClick={() => mutate()}
            className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg transition text-sm"
          >
            刷新
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-400">加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-[#151a35]">
                <tr>
                  <th className="text-left p-4">ID</th>
                  <th className="text-left p-4">用户地址</th>
                  <th className="text-left p-4">资产</th>
                  <th className="text-right p-4">金额</th>
                  <th className="text-left p-4">链</th>
                  <th className="text-left p-4">提现地址</th>
                  <th className="text-left p-4">交易哈希</th>
                  <th className="text-left p-4">任务ID</th>
                  <th className="text-left p-4">状态</th>
                  <th className="text-left p-4">时间</th>
                </tr>
              </thead>
              <tbody>
                {withdrawals.length === 0 ? (
                  <tr>
                    <td colSpan={10} className="text-center p-8 text-gray-400">
                      暂无提现记录
                    </td>
                  </tr>
                ) : (
                  withdrawals.map((withdrawal: WithdrawRecord) => {
                    const chain = chainMap.get(withdrawal.chain_id);
                    const explorerUrl = chain?.block_explorer_url || 'https://bscscan.com';

                    return (
                      <tr key={withdrawal.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                        <td className="p-4 text-xs font-mono text-gray-400">
                          {withdrawal.id.substring(0, 8)}...
                        </td>
                        <td className="p-4 text-xs font-mono">
                          {withdrawal.user?.wallet_address
                            ? `${withdrawal.user.wallet_address.substring(0, 6)}...${withdrawal.user.wallet_address.substring(38)}`
                            : '-'}
                        </td>
                        <td className="p-4 font-semibold">{withdrawal.asset}</td>
                        <td className="p-4 text-right font-mono">{parseFloat(withdrawal.amount).toFixed(8)}</td>
                        <td className="p-4 text-sm">
                          {chain ? chain.chain_name : `ID: ${withdrawal.chain_id}`}
                        </td>
                        <td className="p-4 text-xs font-mono text-gray-400">
                          {withdrawal.address.substring(0, 6)}...{withdrawal.address.substring(38)}
                        </td>
                        <td className="p-4">
                          {withdrawal.tx_hash ? (
                            <a
                              href={`${explorerUrl}/tx/${withdrawal.tx_hash}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary hover:underline text-xs font-mono"
                            >
                              {withdrawal.tx_hash.substring(0, 10)}...{withdrawal.tx_hash.substring(60)}
                            </a>
                          ) : (
                            <span className="text-gray-500 text-xs">-</span>
                          )}
                        </td>
                        <td className="p-4">
                          {withdrawal.task_id ? (
                            <a
                              href={`/dashboard/tasks`}
                              className="text-blue-400 hover:underline text-xs font-mono"
                              title="查看任务详情"
                            >
                              {withdrawal.task_id.substring(0, 10)}...
                            </a>
                          ) : (
                            <span className="text-gray-500 text-xs">-</span>
                          )}
                        </td>
                        <td className="p-4">
                          <span className={`px-2 py-1 rounded text-xs ${getStatusBadge(withdrawal.status)}`}>
                            {getStatusText(withdrawal.status)}
                          </span>
                        </td>
                        <td className="p-4 text-sm text-gray-400">
                          {new Date(withdrawal.created_at).toLocaleString('zh-CN')}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
