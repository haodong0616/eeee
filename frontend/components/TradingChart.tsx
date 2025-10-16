'use client';

import { useEffect, useRef, useState } from 'react';
import { createChart, IChartApi, ISeriesApi, CandlestickData } from 'lightweight-charts';
import { useGetKlinesQuery } from '@/lib/services/api';

interface TradingChartProps {
  symbol: string;
}

export default function TradingChart({ symbol }: TradingChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const [interval, setInterval] = useState('1m');

  // 使用 RTK Query 自动刷新 K线数据
  const { data: klines } = useGetKlinesQuery(
    { symbol, interval },
    {
      pollingInterval: 10000, // 10秒自动轮询
    }
  );

  // 初始化图表
  useEffect(() => {
    if (!chartContainerRef.current) return;

    // 响应式高度：移动端300px，桌面端600px
    const isMobile = window.innerWidth < 1024;
    const chartHeight = isMobile ? 300 : 600;
    
    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: chartHeight,
      layout: {
        background: { color: '#0f1429' },
        textColor: '#d1d5db',
      },
      grid: {
        vertLines: { color: '#1f2937' },
        horzLines: { color: '#1f2937' },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: '#374151',
      },
      rightPriceScale: {
        borderColor: '#374151',
      },
    });

    chartRef.current = chart;

    const candlestickSeries = chart.addCandlestickSeries({
      upColor: '#10B981',
      downColor: '#EF4444',
      borderUpColor: '#10B981',
      borderDownColor: '#EF4444',
      wickUpColor: '#10B981',
      wickDownColor: '#EF4444',
    });

    candlestickSeriesRef.current = candlestickSeries;

    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
        });
      }
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, []);

  // 当 K线数据更新时，更新图表
  useEffect(() => {
    if (candlestickSeriesRef.current && klines && klines.length > 0) {
      const data: CandlestickData[] = klines.map((kline) => ({
        time: kline.open_time as any,
        open: parseFloat(kline.open),
        high: parseFloat(kline.high),
        low: parseFloat(kline.low),
        close: parseFloat(kline.close),
      }));

      candlestickSeriesRef.current.setData(data);
    }
  }, [klines]);

  const handleIntervalChange = (newInterval: string) => {
    setInterval(newInterval);
  };

  return (
    <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
      <div className="px-3 py-2 bg-[#151a35] border-b border-gray-800">
        <div className="flex items-center justify-between mb-2 lg:mb-0">
          <h3 className="text-sm font-semibold">K线图</h3>
        </div>

        {/* Desktop Intervals */}
        <div className="hidden lg:flex flex-wrap gap-1.5 mt-2">
          {/* 秒级 */}
          <div className="flex items-center space-x-1">
            <span className="text-[10px] text-gray-500 mr-0.5">秒</span>
            {['15s', '30s'].map((int) => (
              <button
                key={int}
                onClick={() => handleIntervalChange(int)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors ${
                  interval === int ? 'bg-primary text-white' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'
                }`}
              >
                {int}
              </button>
            ))}
          </div>
          {/* 分钟级 */}
          <div className="flex items-center space-x-1">
            <span className="text-[10px] text-gray-500 mr-0.5">分</span>
            {['1m', '3m', '5m', '15m', '30m'].map((int) => (
              <button
                key={int}
                onClick={() => handleIntervalChange(int)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors ${
                  interval === int ? 'bg-primary text-white' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'
                }`}
              >
                {int}
              </button>
            ))}
          </div>
          {/* 小时和日级 */}
          <div className="flex items-center space-x-1">
            <span className="text-[10px] text-gray-500 mr-0.5">时</span>
            {['1h', '4h', '1d'].map((int) => (
              <button
                key={int}
                onClick={() => handleIntervalChange(int)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors ${
                  interval === int ? 'bg-primary text-white' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'
                }`}
              >
                {int}
              </button>
            ))}
          </div>
        </div>

        {/* Mobile Intervals - Compact */}
        <div className="lg:hidden flex items-center gap-1.5 mt-2">
          {/* Quick Buttons */}
          {['1m', '5m', '15m'].map((int) => (
            <button
              key={int}
              onClick={() => handleIntervalChange(int)}
              className={`flex-1 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                interval === int 
                  ? 'bg-primary text-white shadow-md' 
                  : 'bg-gray-700/50 text-gray-300 active:bg-gray-600'
              }`}
            >
              {int}
            </button>
          ))}

          {/* More Intervals Dropdown */}
          <div className="relative flex-shrink-0">
            <select
              value={['1m', '5m', '15m'].includes(interval) ? '' : interval}
              onChange={(e) => e.target.value && handleIntervalChange(e.target.value)}
              className={`appearance-none text-xs font-semibold px-2.5 py-1.5 pr-7 rounded-lg border transition-all outline-none cursor-pointer ${
                ['1m', '5m', '15m'].includes(interval)
                  ? 'bg-gray-700/50 text-gray-300 border-gray-600'
                  : 'bg-primary text-white border-primary shadow-md'
              }`}
              style={{
                minWidth: '60px',
                WebkitAppearance: 'none',
                MozAppearance: 'none',
              }}
            >
              <option value="" disabled className="bg-gray-800 text-white">更多</option>
              <option value="" disabled className="bg-gray-700 text-gray-400 font-bold">━ 秒级 ━</option>
              <option value="15s" className="bg-gray-800 text-white py-2">15秒</option>
              <option value="30s" className="bg-gray-800 text-white py-2">30秒</option>
              <option value="" disabled className="bg-gray-700 text-gray-400 font-bold">━ 分钟 ━</option>
              <option value="3m" className="bg-gray-800 text-white py-2">3分钟</option>
              <option value="30m" className="bg-gray-800 text-white py-2">30分钟</option>
              <option value="" disabled className="bg-gray-700 text-gray-400 font-bold">━ 小时 ━</option>
              <option value="1h" className="bg-gray-800 text-white py-2">1小时</option>
              <option value="4h" className="bg-gray-800 text-white py-2">4小时</option>
              <option value="" disabled className="bg-gray-700 text-gray-400 font-bold">━ 日级 ━</option>
              <option value="1d" className="bg-gray-800 text-white py-2">1天</option>
            </select>
            <div className="absolute right-1.5 top-1/2 -translate-y-1/2 pointer-events-none text-gray-300">
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>
        </div>
      </div>
      
      <div className="p-1 lg:p-2">
        <div ref={chartContainerRef} />
      </div>
    </div>
  );
}

