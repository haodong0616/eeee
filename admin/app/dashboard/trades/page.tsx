'use client';

import { useEffect, useState } from 'react';
import { adminApi, Trade } from '@/lib/api/admin';

export default function TradesPage() {
  const [trades, setTrades] = useState<Trade[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTrades();
  }, []);

  const loadTrades = async () => {
    try {
      const data = await adminApi.getTrades();
      setTrades(data);
    } catch (error) {
      console.error('Failed to load trades:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">成交记录</h1>

      {loading ? (
        <div>加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4">ID</th>
                <th className="text-left p-4">交易对</th>
                <th className="text-left p-4">买单ID</th>
                <th className="text-left p-4">卖单ID</th>
                <th className="text-right p-4">价格</th>
                <th className="text-right p-4">数量</th>
                <th className="text-right p-4">成交额</th>
                <th className="text-left p-4">时间</th>
              </tr>
            </thead>
            <tbody>
              {trades.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-center p-8 text-gray-400">
                    暂无数据
                  </td>
                </tr>
              ) : (
                trades.map((trade) => {
                  const amount = parseFloat(trade.price) * parseFloat(trade.quantity);
                  return (
                    <tr key={trade.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                      <td className="p-4">{trade.id}</td>
                      <td className="p-4">{trade.symbol}</td>
                      <td className="p-4">{trade.buy_order_id}</td>
                      <td className="p-4">{trade.sell_order_id}</td>
                      <td className="p-4 text-right">{parseFloat(trade.price).toFixed(2)}</td>
                      <td className="p-4 text-right">{parseFloat(trade.quantity).toFixed(4)}</td>
                      <td className="p-4 text-right font-semibold">{amount.toFixed(2)}</td>
                      <td className="p-4 text-sm">
                        {new Date(trade.created_at).toLocaleString()}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}


