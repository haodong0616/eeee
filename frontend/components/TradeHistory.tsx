'use client';

import { Trade } from '@/lib/services/api';
import { formatPrice, formatQuantity } from '@/lib/utils/format';

interface TradeHistoryProps {
  trades: Trade[];
  symbol?: string;
}

export default function TradeHistory({ trades, symbol }: TradeHistoryProps) {
  // 移动端只显示8条，桌面端显示20条
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 1024;
  const displayLimit = isMobile ? 8 : 20;

  return (
    <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
      <div className="px-2 lg:px-3 py-1.5 lg:py-2 bg-[#151a35] border-b border-gray-800">
        <h3 className="text-xs lg:text-sm font-semibold">最近成交</h3>
      </div>

      <div className="p-1 lg:p-2">
        <div className="grid grid-cols-3 gap-1 text-[9px] lg:text-[10px] text-gray-500 mb-1 px-1">
          <div>价格</div>
          <div className="text-right">数量</div>
          <div className="text-right">时间</div>
        </div>

        <div className="space-y-[1px] lg:space-y-[2px] max-h-[200px] lg:max-h-80 overflow-y-auto">
          {trades.length === 0 ? (
            <div className="text-center text-gray-400 py-6 lg:py-8 text-xs">暂无成交</div>
          ) : (
            trades.slice(0, displayLimit).map((trade) => {
              const time = new Date(trade.created_at);
              return (
                <div key={trade.id} className="grid grid-cols-3 gap-1 text-[10px] lg:text-xs hover:bg-[#1a1f3a] px-1 py-[1px] lg:py-[2px] rounded">
                  <div className="text-buy font-medium">
                    {formatPrice(trade.price)}
                  </div>
                  <div className="text-right text-gray-300">{formatQuantity(trade.quantity, symbol)}</div>
                  <div className="text-right text-[9px] lg:text-[10px] text-gray-500">
                    {time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}

