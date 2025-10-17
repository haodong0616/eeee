'use client';

import { useState } from 'react';
import useSWR from 'swr';
import axios from '@/lib/api/client';

interface PnLRecord {
  id: string;
  symbol: string;
  trade_id?: string;
  side: string;
  execute_price: string;
  market_price: string;
  quantity: string;
  profit_loss: string;
  profit_percent: string;
  created_at: string;
}

interface SymbolPnL {
  symbol: string;
  total_pnl: string;
  trade_count: number;
}

interface MarketMakerStats {
  total_pnl: string;
  total_trades: number;
  profit_trades: number;
  loss_trades: number;
  win_rate: number;
  by_symbol: SymbolPnL[];
}

export default function MarketMakerPage() {
  const [selectedSymbol, setSelectedSymbol] = useState<string>('');

  // 获取盈亏统计
  const { data: stats, isLoading: statsLoading } = useSWR<MarketMakerStats>(
    '/admin/market-maker/stats',
    () => axios.get('/admin/market-maker/stats').then(res => res.data),
    { refreshInterval: 10000 }
  );

  // 获取盈亏记录
  const { data: pnlData, isLoading: recordsLoading, mutate } = useSWR(
    `/admin/market-maker/pnl${selectedSymbol ? `?symbol=${selectedSymbol}` : ''}`,
    () => axios.get(`/admin/market-maker/pnl${selectedSymbol ? `?symbol=${selectedSymbol}` : ''}`).then(res => res.data),
    { refreshInterval: 10000 }
  );

  const records: PnLRecord[] = pnlData?.records || [];
  const totalPnL = parseFloat(pnlData?.total_pnl || '0');

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">🤖 做市商盈亏</h1>
        <button
          onClick={() => mutate()}
          className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg transition text-sm"
        >
          刷新
        </button>
      </div>

      {/* 统计卡片 */}
      {statsLoading ? (
        <div className="text-center py-8 text-gray-400">加载统计中...</div>
      ) : stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          {/* 总盈亏 */}
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
            <div className="text-gray-400 text-sm mb-2">总盈亏</div>
            <div className={`text-2xl font-bold ${
              parseFloat(stats.total_pnl) > 0 ? 'text-green-400' :
              parseFloat(stats.total_pnl) < 0 ? 'text-red-400' : 'text-gray-400'
            }`}>
              {parseFloat(stats.total_pnl) > 0 ? '+' : ''}{parseFloat(stats.total_pnl).toFixed(2)} USDT
            </div>
          </div>

          {/* 总交易次数 */}
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
            <div className="text-gray-400 text-sm mb-2">总交易次数</div>
            <div className="text-2xl font-bold">{stats.total_trades}</div>
          </div>

          {/* 胜率 */}
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
            <div className="text-gray-400 text-sm mb-2">胜率</div>
            <div className="text-2xl font-bold text-primary">
              {stats.win_rate.toFixed(1)}%
            </div>
            <div className="text-xs text-gray-500 mt-1">
              盈利 {stats.profit_trades} / 亏损 {stats.loss_trades}
            </div>
          </div>

          {/* 平均盈亏 */}
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
            <div className="text-gray-400 text-sm mb-2">平均盈亏</div>
            <div className={`text-2xl font-bold ${
              parseFloat(stats.total_pnl) / stats.total_trades > 0 ? 'text-green-400' : 'text-red-400'
            }`}>
              {stats.total_trades > 0 
                ? (parseFloat(stats.total_pnl) / stats.total_trades).toFixed(2)
                : '0.00'} USDT
            </div>
          </div>
        </div>
      )}

      {/* 按交易对统计 */}
      {stats && stats.by_symbol.length > 0 && (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6 mb-6">
          <h2 className="text-xl font-bold mb-4">各交易对盈亏</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
            {stats.by_symbol.map((item) => (
              <button
                key={item.symbol}
                onClick={() => setSelectedSymbol(selectedSymbol === item.symbol ? '' : item.symbol)}
                className={`p-4 rounded-lg border transition ${
                  selectedSymbol === item.symbol
                    ? 'border-primary bg-primary/10'
                    : 'border-gray-800 hover:border-gray-700'
                }`}
              >
                <div className="text-sm text-gray-400 mb-1">{item.symbol}</div>
                <div className={`text-lg font-bold ${
                  parseFloat(item.total_pnl) > 0 ? 'text-green-400' :
                  parseFloat(item.total_pnl) < 0 ? 'text-red-400' : 'text-gray-400'
                }`}>
                  {parseFloat(item.total_pnl) > 0 ? '+' : ''}{parseFloat(item.total_pnl).toFixed(2)}
                </div>
                <div className="text-xs text-gray-500 mt-1">{item.trade_count} 笔</div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* 交易记录 */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
        <div className="flex items-center justify-between p-4 border-b border-gray-800">
          <h2 className="text-xl font-bold">
            {selectedSymbol ? `${selectedSymbol} 盈亏记录` : '所有盈亏记录'}
          </h2>
          {selectedSymbol && (
            <button
              onClick={() => setSelectedSymbol('')}
              className="text-sm text-gray-400 hover:text-white transition"
            >
              查看全部
            </button>
          )}
        </div>

        {recordsLoading ? (
          <div className="text-center py-12 text-gray-400">加载中...</div>
        ) : records.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            暂无记录
          </div>
        ) : (
          <>
            {/* 当前筛选的总盈亏 */}
            <div className="p-4 bg-[#151a35] border-b border-gray-800">
              <div className="flex items-center justify-between">
                <span className="text-gray-400">当前筛选盈亏：</span>
                <span className={`text-xl font-bold ${
                  totalPnL > 0 ? 'text-green-400' :
                  totalPnL < 0 ? 'text-red-400' : 'text-gray-400'
                }`}>
                  {totalPnL > 0 ? '+' : ''}{totalPnL.toFixed(2)} USDT
                </span>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-[#151a35]">
                  <tr>
                    <th className="text-left p-4">时间</th>
                    <th className="text-left p-4">交易对</th>
                    <th className="text-left p-4">方向</th>
                    <th className="text-right p-4">数量</th>
                    <th className="text-right p-4">执行价</th>
                    <th className="text-right p-4">市价</th>
                    <th className="text-right p-4">盈亏(USDT)</th>
                    <th className="text-right p-4">盈亏%</th>
                  </tr>
                </thead>
                <tbody>
                  {records.map((record) => {
                    const pnl = parseFloat(record.profit_loss);
                    const pnlPercent = parseFloat(record.profit_percent);
                    
                    return (
                      <tr key={record.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                        <td className="p-4 text-sm text-gray-400">
                          {new Date(record.created_at).toLocaleString('zh-CN', {
                            month: '2-digit',
                            day: '2-digit',
                            hour: '2-digit',
                            minute: '2-digit',
                            second: '2-digit'
                          })}
                        </td>
                        <td className="p-4 font-semibold">{record.symbol}</td>
                        <td className="p-4">
                          <span className={`px-2 py-1 rounded text-xs ${
                            record.side === 'buy' 
                              ? 'bg-green-500/20 text-green-400'
                              : 'bg-red-500/20 text-red-400'
                          }`}>
                            {record.side === 'buy' ? '买入' : '卖出'}
                          </span>
                        </td>
                        <td className="p-4 text-right font-mono">{parseFloat(record.quantity).toFixed(4)}</td>
                        <td className="p-4 text-right font-mono text-sm">${parseFloat(record.execute_price).toFixed(2)}</td>
                        <td className="p-4 text-right font-mono text-sm">${parseFloat(record.market_price).toFixed(2)}</td>
                        <td className={`p-4 text-right font-mono font-bold ${
                          pnl > 0 ? 'text-green-400' :
                          pnl < 0 ? 'text-red-400' : 'text-gray-400'
                        }`}>
                          {pnl > 0 ? '+' : ''}{pnl.toFixed(2)}
                        </td>
                        <td className={`p-4 text-right font-mono text-sm ${
                          pnlPercent > 0 ? 'text-green-400' :
                          pnlPercent < 0 ? 'text-red-400' : 'text-gray-400'
                        }`}>
                          {pnlPercent > 0 ? '+' : ''}{pnlPercent.toFixed(2)}%
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

