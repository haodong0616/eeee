'use client';

import { OrderBook as OrderBookType } from '@/lib/services/api';
import { formatPrice, formatQuantity } from '@/lib/utils/format';

interface OrderBookProps {
  orderBook: OrderBookType | undefined;
  onPriceClick?: (price: string) => void;
  symbol?: string;
}

export default function OrderBook({ orderBook, onPriceClick, symbol }: OrderBookProps) {
  const asks = orderBook?.asks || [];
  const bids = orderBook?.bids || [];
  
  // 移动端显示6条，桌面端显示10条
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 1024;
  const displayLimit = isMobile ? 6 : 10;

  const handlePriceClick = (price: string) => {
    if (onPriceClick) {
      onPriceClick(price);
    }
  };

  return (
    <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
      <div className="px-2 lg:px-3 py-1.5 lg:py-2 bg-[#151a35] border-b border-gray-800">
        <h3 className="text-xs lg:text-sm font-semibold">盘口深度</h3>
      </div>

      <div className="p-1 lg:p-2">
        {/* 卖单（从低到高，倒序显示） */}
        <div className="mb-1 lg:mb-2">
          <div className="grid grid-cols-3 gap-1 text-[9px] lg:text-[10px] text-gray-500 mb-1 px-1">
            <div className="text-right">价格</div>
            <div className="text-right">数量</div>
            <div className="text-right">总计</div>
          </div>
          <div className="space-y-[1px] lg:space-y-[2px]">
            {[...asks].reverse().slice(0, displayLimit).map((ask, index) => (
              <div
                key={index}
                className="grid grid-cols-3 gap-1 text-[10px] lg:text-xs hover:bg-[#1a1f3a] px-1 py-[1px] lg:py-[2px] rounded cursor-pointer transition-all hover:scale-[1.02]"
                onClick={() => handlePriceClick(ask.price)}
                title="点击填入价格"
              >
              <div className="text-right text-sell font-medium hover:text-red-400 transition-colors">
                {formatPrice(ask.price)}
              </div>
              <div className="text-right text-gray-300">{formatQuantity(ask.quantity, symbol)}</div>
              <div className="text-right text-gray-400 text-[9px] lg:text-[10px]">
                {formatPrice(parseFloat(ask.price) * parseFloat(ask.quantity))}
              </div>
              </div>
            ))}
          </div>
        </div>

        {/* 最新价格 */}
        <div className="my-1 lg:my-2 py-1 lg:py-1.5 text-center border-y border-gray-700">
          <span className="text-sm lg:text-base font-bold text-primary">
            {bids[0] ? formatPrice(bids[0].price) : '-'}
          </span>
        </div>

        {/* 买单（从高到低） */}
        <div>
          <div className="space-y-[1px] lg:space-y-[2px]">
            {bids.slice(0, displayLimit).map((bid, index) => (
              <div
                key={index}
                className="grid grid-cols-3 gap-1 text-[10px] lg:text-xs hover:bg-[#1a1f3a] px-1 py-[1px] lg:py-[2px] rounded cursor-pointer transition-all hover:scale-[1.02]"
                onClick={() => handlePriceClick(bid.price)}
                title="点击填入价格"
              >
              <div className="text-right text-buy font-medium hover:text-green-400 transition-colors">
                {formatPrice(bid.price)}
              </div>
              <div className="text-right text-gray-300">{formatQuantity(bid.quantity, symbol)}</div>
              <div className="text-right text-gray-400 text-[9px] lg:text-[10px]">
                {formatPrice(parseFloat(bid.price) * parseFloat(bid.quantity))}
              </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

