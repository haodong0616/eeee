'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useAppSelector } from '@/lib/store/hooks';
import {
  useGetTickerQuery,
  useGetOrderBookQuery,
  useGetRecentTradesQuery,
  useGetOrdersQuery,
  useCreateOrderMutation,
  useCancelOrderMutation,
} from '@/lib/services/api';
import { wsClient } from '@/lib/websocket';
import { formatPrice, formatPercent, formatQuantity, formatVolume } from '@/lib/utils/format';
import OrderBook from '@/components/OrderBook';
import TradeHistory from '@/components/TradeHistory';
import OrderForm from '@/components/OrderForm';
import MyOrders from '@/components/MyOrders';
import TradingChart from '@/components/TradingChart';

export default function TradePage() {
  const params = useParams();
  const symbol = (params.symbol as string).replace('-', '/');
  
  const { isAuthenticated } = useAppSelector((state) => state.auth);
  const [selectedTab, setSelectedTab] = useState<'open' | 'history'>('open');
  const [selectedPrice, setSelectedPrice] = useState<string>('');

  // 使用 RTK Query 自动刷新数据
  const { data: currentTicker } = useGetTickerQuery(symbol, {
    pollingInterval: 3000, // 3秒轮询
  });
  const { data: orderBook } = useGetOrderBookQuery(symbol, {
    pollingInterval: 2000, // 2秒轮询
  });
  const { data: recentTrades = [] } = useGetRecentTradesQuery(symbol, {
    pollingInterval: 3000,
  });

  // 获取用户订单（仅在登录时）
  const { data: ordersData } = useGetOrdersQuery(
    { symbol },
    { skip: !isAuthenticated, pollingInterval: 5000 }
  );

  // Mutations
  const [createOrder] = useCreateOrderMutation();
  const [cancelOrder] = useCancelOrderMutation();

  useEffect(() => {
    // 连接WebSocket
    wsClient.connect();

    return () => {
      // WebSocket 保持连接
    };
  }, [symbol]);

  const handleCreateOrder = async (orderData: any) => {
    if (!isAuthenticated) {
      alert('请先连接钱包');
      return;
    }

    try {
      await createOrder({ ...orderData, symbol }).unwrap();
      alert('订单创建成功');
    } catch (error: any) {
      alert(error?.data?.error || '订单创建失败');
    }
  };

  const handleCancelOrder = async (orderId: number) => {
    try {
      await cancelOrder(orderId).unwrap();
      alert('订单已取消');
    } catch (error: any) {
      alert(error?.data?.error || '取消订单失败');
    }
  };

  const handlePriceClick = (price: string) => {
    setSelectedPrice(price);
  };

  const changePercent = currentTicker?.change_24h
    ? parseFloat(currentTicker.change_24h)
    : 0;
  const isPositive = changePercent >= 0;

  return (
    <div className="container mx-auto px-4 py-4">
      {/* 顶部信息栏 */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 mb-4">
        {/* 移动端布局 - 紧凑 */}
        <div className="lg:hidden p-3">
          <div className="flex items-center justify-between mb-2">
            <h1 className="text-base font-bold">{symbol}</h1>
            <div className="text-right">
              <p className={`text-lg font-bold font-mono ${isPositive ? 'text-buy' : 'text-sell'}`}>
                ${currentTicker?.last_price ? formatPrice(currentTicker.last_price) : '-'}
              </p>
              <p className={`text-xs font-semibold ${isPositive ? 'text-buy' : 'text-sell'}`}>
                {currentTicker?.change_24h ? formatPercent(currentTicker.change_24h) : '-'}
              </p>
            </div>
          </div>
          <div className="grid grid-cols-3 gap-2 text-xs">
            <div className="text-center p-1.5 bg-gray-800/30 rounded">
              <p className="text-[10px] text-gray-500 mb-0.5">最高</p>
              <p className="font-mono text-[11px] text-gray-300">${currentTicker?.high_24h ? formatPrice(currentTicker.high_24h) : '-'}</p>
            </div>
            <div className="text-center p-1.5 bg-gray-800/30 rounded">
              <p className="text-[10px] text-gray-500 mb-0.5">最低</p>
              <p className="font-mono text-[11px] text-gray-300">${currentTicker?.low_24h ? formatPrice(currentTicker.low_24h) : '-'}</p>
            </div>
            <div className="text-center p-1.5 bg-gray-800/30 rounded">
              <p className="text-[10px] text-gray-500 mb-0.5">成交量</p>
              <p className="font-mono text-[11px] text-gray-300">{currentTicker?.volume_24h ? formatVolume(currentTicker.volume_24h) : '-'}</p>
            </div>
          </div>
        </div>

        {/* 桌面端布局 - 横向 */}
        <div className="hidden lg:block p-4">
          <div className="flex items-center space-x-8">
            <div>
              <h1 className="text-2xl font-bold">{symbol}</h1>
            </div>
            <div>
              <p className="text-sm text-gray-400">最新价格</p>
              <p className={`text-2xl font-bold font-mono ${isPositive ? 'text-buy' : 'text-sell'}`}>
                ${currentTicker?.last_price ? formatPrice(currentTicker.last_price) : '-'}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-400">24h涨跌</p>
              <p className={`text-lg font-semibold ${isPositive ? 'text-buy' : 'text-sell'}`}>
                {currentTicker?.change_24h ? formatPercent(currentTicker.change_24h) : '-'}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-400">24h最高</p>
              <p className="text-lg font-mono">${currentTicker?.high_24h ? formatPrice(currentTicker.high_24h) : '-'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">24h最低</p>
              <p className="text-lg font-mono">${currentTicker?.low_24h ? formatPrice(currentTicker.low_24h) : '-'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">24h成交量</p>
              <p className="text-lg font-mono">{currentTicker?.volume_24h ? formatVolume(currentTicker.volume_24h) : '-'}</p>
            </div>
          </div>
        </div>
      </div>

      {/* 主交易区域 */}
      {/* 移动端布局 */}
      <div className="lg:hidden space-y-3 mb-4">
        {/* 第一行：K线图 */}
        <div>
          <TradingChart symbol={symbol} />
        </div>

        {/* 第二行：盘口 + 买卖表单 */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <OrderBook orderBook={orderBook} onPriceClick={handlePriceClick} symbol={symbol} />
          </div>
          <div>
            <OrderForm
              symbol={symbol}
              currentPrice={currentTicker?.last_price}
              onSubmit={handleCreateOrder}
              isAuthenticated={isAuthenticated}
              initialPrice={selectedPrice}
            />
          </div>
        </div>

        {/* 第三行：最近成交 */}
        <div>
          <TradeHistory trades={recentTrades} symbol={symbol} />
        </div>
      </div>

      {/* 桌面端布局 */}
      <div className="hidden lg:grid grid-cols-12 gap-4 mb-4">
        {/* 左侧：盘口和最近成交 */}
        <div className="col-span-3">
          <OrderBook orderBook={orderBook} onPriceClick={handlePriceClick} symbol={symbol} />
          <div className="mt-4">
            <TradeHistory trades={recentTrades} symbol={symbol} />
          </div>
        </div>

        {/* 中间：图表区域 */}
        <div className="col-span-6">
          <TradingChart symbol={symbol} />
        </div>

        {/* 右侧：下单区域 */}
        <div className="col-span-3">
          <OrderForm
            symbol={symbol}
            currentPrice={currentTicker?.last_price}
            onSubmit={handleCreateOrder}
            isAuthenticated={isAuthenticated}
            initialPrice={selectedPrice}
          />
        </div>
      </div>

      {/* 底部：当前委托和历史委托 */}
      <div className="bg-[#0f1429] rounded-lg border border-gray-800">
        <div className="flex border-b border-gray-800">
          <button
            className={`px-6 py-3 ${
              selectedTab === 'open'
                ? 'text-primary border-b-2 border-primary'
                : 'text-gray-400'
            }`}
            onClick={() => setSelectedTab('open')}
          >
            当前委托
          </button>
          <button
            className={`px-6 py-3 ${
              selectedTab === 'history'
                ? 'text-primary border-b-2 border-primary'
                : 'text-gray-400'
            }`}
            onClick={() => setSelectedTab('history')}
          >
            历史委托
          </button>
        </div>
        <MyOrders
          type={selectedTab}
          symbol={symbol}
          onCancel={handleCancelOrder}
        />
      </div>
    </div>
  );
}

