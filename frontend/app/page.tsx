'use client';

import Link from 'next/link';
import { useGetAllTickersQuery } from '@/lib/services/api';
import { formatPrice, formatPercent, formatVolume } from '@/lib/utils/format';

export default function Home() {
  const { data: tickers = [] } = useGetAllTickersQuery(undefined, {
    pollingInterval: 5000,
  });

  return (
    <div className="min-h-screen">
      {/* Hero Section with Gradient Background */}
      <section className="relative overflow-hidden">
        {/* Animated Background */}
        <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-purple-500/10">
          <div className="absolute inset-0 opacity-20"></div>
        </div>
        
        {/* Floating Elements - Hidden on mobile */}
        <div className="hidden md:block absolute top-20 left-10 w-72 h-72 bg-primary/20 rounded-full blur-3xl animate-pulse"></div>
        <div className="hidden md:block absolute bottom-20 right-10 w-96 h-96 bg-purple-500/20 rounded-full blur-3xl animate-pulse delay-1000"></div>
        
        <div className="relative container mx-auto px-4 py-12 md:py-20">
          <div className="text-center max-w-4xl mx-auto">
            {/* Logo & Title */}
            <div className="mb-4 md:mb-6 animate-fade-in">
              <div className="inline-flex items-center gap-2 md:gap-3 mb-3">
                <div className="w-12 h-12 md:w-16 md:h-16 bg-gradient-to-br from-primary to-purple-500 rounded-xl md:rounded-2xl flex items-center justify-center shadow-lg shadow-primary/50 animate-float">
                  <span className="text-2xl md:text-3xl">⚡</span>
                </div>
                <h1 className="text-4xl md:text-6xl font-bold bg-gradient-to-r from-primary via-purple-400 to-primary bg-clip-text text-transparent animate-gradient">
                  Velocity
                </h1>
              </div>
              <p className="text-xl md:text-2xl text-gray-400 mb-1">Exchange</p>
              <p className="text-sm md:text-lg text-gray-500">速度交易所</p>
            </div>

            {/* Tagline */}
            <p className="text-lg md:text-2xl text-gray-300 mb-2 md:mb-4 animate-fade-in-delay-1">
              ⚡ 极速交易，流畅体验
            </p>
            <p className="text-sm md:text-lg text-gray-400 mb-6 md:mb-12 animate-fade-in-delay-2">
              安全 · 快速 · 专业的数字资产交易平台
            </p>

            {/* CTA Buttons */}
            <div className="flex flex-col sm:flex-row justify-center gap-3 md:gap-4 mb-8 md:mb-12 animate-fade-in-delay-3 px-4">
              <Link
                href="/markets"
                className="group px-6 md:px-8 py-3 md:py-4 bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90 rounded-lg md:rounded-xl font-semibold transition-all shadow-lg shadow-primary/50 hover:shadow-primary/70 hover:scale-105 flex items-center justify-center gap-2 text-sm md:text-base"
              >
                <span>🚀</span>
                开始交易
                <svg className="w-4 h-4 md:w-5 md:h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </Link>
              <Link
                href="/markets"
                className="px-6 md:px-8 py-3 md:py-4 bg-gray-800/50 hover:bg-gray-700/50 border-2 border-gray-700 hover:border-primary rounded-lg md:rounded-xl font-semibold transition-all backdrop-blur-sm hover:scale-105 flex items-center justify-center gap-2 text-sm md:text-base"
              >
                <span>📈</span>
                查看行情
              </Link>
            </div>

            {/* Features Grid */}
            <div className="grid grid-cols-3 gap-2 md:gap-6 max-w-4xl mx-auto animate-fade-in-delay-4">
              <div className="p-2 md:p-6 bg-gray-800/30 backdrop-blur-sm rounded-lg md:rounded-xl border border-gray-700/50 hover:border-primary/50 transition-all hover:scale-105 text-center">
                <div className="text-xl md:text-4xl mb-1 md:mb-3">🔒</div>
                <h3 className="text-xs md:text-xl font-semibold mb-0.5 md:mb-2">安全可靠</h3>
                <p className="text-gray-400 text-[10px] md:text-sm hidden md:block">多重安全防护机制</p>
              </div>
              <div className="p-2 md:p-6 bg-gray-800/30 backdrop-blur-sm rounded-lg md:rounded-xl border border-gray-700/50 hover:border-primary/50 transition-all hover:scale-105 text-center">
                <div className="text-xl md:text-4xl mb-1 md:mb-3">⚡</div>
                <h3 className="text-xs md:text-xl font-semibold mb-0.5 md:mb-2">极速撮合</h3>
                <p className="text-gray-400 text-[10px] md:text-sm hidden md:block">毫秒级订单处理</p>
              </div>
              <div className="p-2 md:p-6 bg-gray-800/30 backdrop-blur-sm rounded-lg md:rounded-xl border border-gray-700/50 hover:border-primary/50 transition-all hover:scale-105 text-center">
                <div className="text-xl md:text-4xl mb-1 md:mb-3">💎</div>
                <h3 className="text-xs md:text-xl font-semibold mb-0.5 md:mb-2">专业工具</h3>
                <p className="text-gray-400 text-[10px] md:text-sm hidden md:block">完善的交易工具</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Hot Trading Pairs */}
      <section className="container mx-auto px-4 pb-12 md:pb-20">
        <div className="flex items-center justify-between mb-6 md:mb-8">
          <div>
            <h2 className="text-2xl md:text-3xl font-bold mb-1 md:mb-2 flex items-center gap-2 md:gap-3">
              <span className="text-2xl md:text-3xl">🔥</span>
              热门交易对
            </h2>
            <p className="text-sm md:text-base text-gray-400 hidden sm:block">实时行情数据，每5秒更新</p>
          </div>
          <Link 
            href="/markets"
            className="px-4 md:px-6 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 hover:border-primary rounded-lg transition-all flex items-center gap-2 text-sm md:text-base"
          >
            <span className="hidden sm:inline">查看全部</span>
            <span className="sm:hidden">全部</span>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        </div>

        <div className="bg-gradient-to-br from-gray-800/40 to-gray-900/40 backdrop-blur-sm rounded-xl md:rounded-2xl border border-gray-700/50 overflow-hidden shadow-2xl">
          {/* Table Header - Desktop */}
          <div className="hidden md:grid grid-cols-5 gap-4 p-4 md:p-5 bg-gradient-to-r from-primary/10 to-purple-500/10 border-b border-gray-700/50 font-semibold text-sm">
            <div>交易对</div>
            <div className="text-right">最新价格</div>
            <div className="text-right">24h涨跌</div>
            <div className="text-right">24h最高</div>
            <div className="text-right">24h成交量</div>
          </div>

          {/* Table Header - Mobile */}
          <div className="md:hidden grid grid-cols-3 gap-2 p-3 bg-gradient-to-r from-primary/10 to-purple-500/10 border-b border-gray-700/50 font-semibold text-xs">
            <div>交易对</div>
            <div className="text-right">价格</div>
            <div className="text-right">24h涨跌</div>
          </div>

          {/* Table Rows */}
          {tickers.slice(0, 10).map((ticker, index) => {
            const isPositive = parseFloat(ticker.change_24h || '0') >= 0;
            return (
              <Link
                key={ticker.symbol}
                href={`/trade/${ticker.symbol.replace('/', '-')}`}
                className={`group hover:bg-gradient-to-r hover:from-primary/5 hover:to-purple-500/5 transition-all border-t border-gray-800/50 hover:border-primary/30 animate-fade-in ${
                  index % 2 === 0 ? 'bg-gray-900/20' : 'bg-transparent'
                }`}
                style={{ animationDelay: `${index * 50}ms` }}
              >
                {/* Desktop Layout */}
                <div className="hidden md:grid grid-cols-5 gap-4 p-4 md:p-5">
                  <div className="font-semibold">
                    <div className="group-hover:text-primary transition-colors">{ticker.symbol}</div>
                    <div className="text-xs text-gray-500">{ticker.symbol.split('/')[0]}</div>
                  </div>
                  <div className="text-right font-mono font-semibold group-hover:text-primary transition-colors">
                    ${ticker.last_price ? formatPrice(ticker.last_price) : '-'}
                  </div>
                  <div className="text-right">
                    <span className={`inline-flex items-center gap-1 px-3 py-1 rounded-lg font-semibold ${
                      isPositive 
                        ? 'bg-green-500/20 text-green-400' 
                        : 'bg-red-500/20 text-red-400'
                    }`}>
                      {isPositive ? '↗' : '↘'}
                      {ticker.change_24h ? formatPercent(ticker.change_24h) : '-'}
                    </span>
                  </div>
                  <div className="text-right font-mono text-gray-300 text-sm">
                    ${ticker.high_24h ? formatPrice(ticker.high_24h) : '-'}
                  </div>
                  <div className="text-right font-mono text-gray-400 text-sm">
                    {ticker.volume_24h ? formatVolume(ticker.volume_24h) : '-'}
                  </div>
                </div>

                {/* Mobile Layout */}
                <div className="md:hidden grid grid-cols-3 gap-2 p-3">
                  <div className="text-sm font-semibold group-hover:text-primary transition-colors">
                    {ticker.symbol.split('/')[0]}
                  </div>
                  <div className="text-right">
                    <div className="font-mono font-semibold text-sm">
                      ${ticker.last_price ? formatPrice(ticker.last_price) : '-'}
                    </div>
                  </div>
                  <div className="text-right">
                    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-semibold ${
                      isPositive 
                        ? 'bg-green-500/20 text-green-400' 
                        : 'bg-red-500/20 text-red-400'
                    }`}>
                      {isPositive ? '↗' : '↘'}
                      {ticker.change_24h ? formatPercent(ticker.change_24h) : '-'}
                    </span>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>

      </section>
    </div>
  );
}
