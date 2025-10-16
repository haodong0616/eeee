'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useGetAllTickersQuery } from '@/lib/services/api';
import { formatPrice, formatPercent, formatQuantity, formatVolume } from '@/lib/utils/format';

export default function MarketsPage() {
  const { data: tickers = [], isLoading } = useGetAllTickersQuery(undefined, {
    pollingInterval: 5000,
  });
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState<'volume' | 'change' | 'price'>('volume');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  const filteredTickers = tickers.filter((ticker) =>
    ticker.symbol.toLowerCase().includes(search.toLowerCase())
  );

  const sortedTickers = [...filteredTickers].sort((a, b) => {
    let aValue = 0, bValue = 0;
    
    switch (sortBy) {
      case 'volume':
        aValue = parseFloat(a.volume_24h || '0');
        bValue = parseFloat(b.volume_24h || '0');
        break;
      case 'change':
        aValue = parseFloat(a.change_24h || '0');
        bValue = parseFloat(b.change_24h || '0');
        break;
      case 'price':
        aValue = parseFloat(a.last_price || '0');
        bValue = parseFloat(b.last_price || '0');
        break;
    }
    
    return sortOrder === 'desc' ? bValue - aValue : aValue - bValue;
  });

  const handleSort = (field: 'volume' | 'change' | 'price') => {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc');
    } else {
      setSortBy(field);
      setSortOrder('desc');
    }
  };


  return (
    <div className="min-h-screen">
      {/* Header Section - Compact for Mobile */}
      <section className="sticky top-0 z-10 bg-[#0a0e27] border-b border-gray-800/50">
        <div className="container mx-auto px-4 py-3 md:py-4">
          <div className="flex items-center gap-3 mb-3">
            <h1 className="text-xl md:text-3xl font-bold flex items-center gap-2">
              <span className="text-xl md:text-3xl">📈</span>
              <span className="bg-gradient-to-r from-primary to-purple-400 bg-clip-text text-transparent">
                实时行情
              </span>
            </h1>
            
            {/* Stats Inline - Mobile */}
            <div className="md:hidden flex gap-2 ml-auto text-[10px]">
              <span className="px-2 py-1 bg-primary/10 rounded text-primary font-bold">
                {tickers.length}币
              </span>
              <span className="px-2 py-1 bg-green-500/10 rounded text-green-400 font-bold">
                {tickers.filter(t => parseFloat(t.change_24h || '0') > 0).length}↗
              </span>
              <span className="px-2 py-1 bg-red-500/10 rounded text-red-400 font-bold">
                {tickers.filter(t => parseFloat(t.change_24h || '0') < 0).length}↘
              </span>
            </div>
          </div>

          {/* Search Bar */}
          <div className="relative">
            <input
              type="text"
              placeholder="🔍 搜索交易对..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full px-4 py-2 md:py-2.5 bg-gray-800/50 backdrop-blur-sm border border-gray-700 hover:border-primary focus:border-primary rounded-lg transition-all outline-none text-sm placeholder:text-gray-500"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white transition-colors text-sm"
              >
                ✕
              </button>
            )}
          </div>

          {/* Sort Buttons - Mobile Only */}
          <div className="md:hidden flex gap-2 mt-2 overflow-x-auto pb-1">
            <button
              onClick={() => handleSort('change')}
              className={`flex-shrink-0 px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                sortBy === 'change' 
                  ? 'bg-primary text-white' 
                  : 'bg-gray-800/50 text-gray-400'
              }`}
            >
              涨跌排序 {sortBy === 'change' && (sortOrder === 'desc' ? '↓' : '↑')}
            </button>
            <button
              onClick={() => handleSort('price')}
              className={`flex-shrink-0 px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                sortBy === 'price' 
                  ? 'bg-primary text-white' 
                  : 'bg-gray-800/50 text-gray-400'
              }`}
            >
              价格排序 {sortBy === 'price' && (sortOrder === 'desc' ? '↓' : '↑')}
            </button>
            <button
              onClick={() => handleSort('volume')}
              className={`flex-shrink-0 px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                sortBy === 'volume' 
                  ? 'bg-primary text-white' 
                  : 'bg-gray-800/50 text-gray-400'
              }`}
            >
              成交量排序 {sortBy === 'volume' && (sortOrder === 'desc' ? '↓' : '↑')}
            </button>
          </div>
        </div>
      </section>

      {/* Markets Table */}
      <section className="container mx-auto px-2 md:px-4 py-2 md:py-4">
        {isLoading ? (
          <div className="flex items-center justify-center py-12 md:py-20">
            <div className="flex flex-col items-center gap-4">
              <div className="w-12 h-12 md:w-16 md:h-16 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
              <p className="text-gray-400 text-sm md:text-base">加载中...</p>
            </div>
          </div>
        ) : (
          <div className="bg-gradient-to-br from-gray-800/40 to-gray-900/40 backdrop-blur-sm rounded-lg md:rounded-2xl border border-gray-700/50 overflow-hidden shadow-xl md:shadow-2xl">
            {/* Table Header - Desktop */}
            <div className="hidden md:grid grid-cols-6 gap-4 p-4 bg-gradient-to-r from-primary/10 to-purple-500/10 border-b border-gray-700/50 font-semibold text-sm">
              <div>交易对</div>
              <button
                onClick={() => handleSort('price')}
                className="text-right hover:text-primary transition-colors flex items-center justify-end gap-1"
              >
                最新价格
                {sortBy === 'price' && (
                  <span className="text-xs">{sortOrder === 'desc' ? '↓' : '↑'}</span>
                )}
              </button>
              <button
                onClick={() => handleSort('change')}
                className="text-right hover:text-primary transition-colors flex items-center justify-end gap-1"
              >
                24h涨跌
                {sortBy === 'change' && (
                  <span className="text-xs">{sortOrder === 'desc' ? '↓' : '↑'}</span>
                )}
              </button>
              <div className="text-right">24h最高</div>
              <div className="text-right">24h最低</div>
              <button
                onClick={() => handleSort('volume')}
                className="text-right hover:text-primary transition-colors flex items-center justify-end gap-1"
              >
                24h成交量
                {sortBy === 'volume' && (
                  <span className="text-xs">{sortOrder === 'desc' ? '↓' : '↑'}</span>
                )}
              </button>
            </div>

            {/* Table Header - Mobile (Hidden, using top sort buttons instead) */}
            <div className="md:hidden bg-gradient-to-r from-primary/10 to-purple-500/10 border-b border-gray-700/50 px-3 py-2">
              <div className="grid grid-cols-3 gap-2 text-[10px] font-semibold text-gray-400">
                <div>交易对</div>
                <div className="text-right">最新价</div>
                <div className="text-right">24h涨跌</div>
              </div>
            </div>

            {/* Table Rows */}
            {sortedTickers.length === 0 ? (
              <div className="text-center py-20 text-gray-400">
                <div className="text-6xl mb-4">🔍</div>
                <p className="text-xl">未找到匹配的交易对</p>
                <p className="text-sm mt-2">试试其他搜索词</p>
              </div>
            ) : (
              sortedTickers.map((ticker, index) => {
                const isPositive = parseFloat(ticker.change_24h || '0') >= 0;
                
                return (
                  <Link
                    key={ticker.symbol}
                    href={`/trade/${ticker.symbol.replace('/', '-')}`}
                    className={`group hover:bg-gradient-to-r hover:from-primary/5 hover:to-purple-500/5 transition-all border-t border-gray-800/50 hover:border-primary/30 animate-fade-in ${
                      index % 2 === 0 ? 'bg-gray-900/20' : 'bg-transparent'
                    }`}
                    style={{ animationDelay: `${index * 30}ms` }}
                  >
                    {/* Desktop Layout */}
                    <div className="hidden md:grid grid-cols-6 gap-4 p-4">
                      {/* Token Name */}
                      <div className="font-semibold">
                        <div className="group-hover:text-primary transition-colors text-base">
                          {ticker.symbol}
                        </div>
                        <div className="text-xs text-gray-500">
                          {ticker.symbol.split('/')[0]}
                        </div>
                      </div>

                      {/* Price */}
                      <div className="text-right">
                        <div className="font-mono font-bold text-base group-hover:text-primary transition-colors">
                          ${ticker.last_price ? formatPrice(ticker.last_price) : '-'}
                        </div>
                        <div className="text-xs text-gray-500">USDT</div>
                      </div>

                      {/* 24h Change */}
                      <div className="text-right flex flex-col items-end justify-center">
                        <span className={`inline-flex items-center gap-1 px-3 py-1.5 rounded-lg font-bold text-sm ${
                          isPositive 
                            ? 'bg-gradient-to-r from-green-500/20 to-green-600/20 text-green-400 border border-green-500/30' 
                            : 'bg-gradient-to-r from-red-500/20 to-red-600/20 text-red-400 border border-red-500/30'
                        }`}>
                          {isPositive ? '↗' : '↘'}
                          {ticker.change_24h ? formatPercent(ticker.change_24h) : '-'}
                        </span>
                      </div>

                      {/* 24h High */}
                      <div className="text-right">
                        <div className="font-mono text-gray-300 text-sm">
                          ${ticker.high_24h ? formatPrice(ticker.high_24h) : '-'}
                        </div>
                        <div className="text-xs text-gray-500">最高</div>
                      </div>

                      {/* 24h Low */}
                      <div className="text-right">
                        <div className="font-mono text-gray-300 text-sm">
                          ${ticker.low_24h ? formatPrice(ticker.low_24h) : '-'}
                        </div>
                        <div className="text-xs text-gray-500">最低</div>
                      </div>

                      {/* Volume */}
                      <div className="text-right">
                        <div className="font-mono text-gray-300 font-semibold text-sm">
                          {ticker.volume_24h ? formatVolume(ticker.volume_24h) : '-'}
                        </div>
                        <div className="text-xs text-gray-500">成交量</div>
                      </div>
                    </div>

                    {/* Mobile Layout */}
                    <div className="md:hidden px-3 py-2.5 space-y-1">
                      {/* Row 1: Token Name | Price | Change */}
                      <div className="grid grid-cols-3 gap-2 items-center">
                        <div className="font-semibold text-sm group-hover:text-primary transition-colors">
                          {ticker.symbol.split('/')[0]}
                        </div>
                        <div className="font-mono font-bold text-sm text-right group-hover:text-primary transition-colors">
                          ${ticker.last_price ? formatPrice(ticker.last_price) : '-'}
                        </div>
                        <div className="text-right">
                          <span className={`inline-flex items-center justify-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold min-w-[60px] ${
                            isPositive 
                              ? 'bg-green-500/20 text-green-400' 
                              : 'bg-red-500/20 text-red-400'
                          }`}>
                            {isPositive ? '↗' : '↘'}
                            {ticker.change_24h ? formatPercent(ticker.change_24h) : '-'}
                          </span>
                        </div>
                      </div>
                      
                      {/* Row 2: Volume */}
                      <div className="text-[10px] text-gray-500 flex items-center justify-between">
                        <span>24h量</span>
                        <span className="font-mono text-gray-400 font-semibold">
                          {ticker.volume_24h ? formatVolume(ticker.volume_24h) : '-'}
                        </span>
                      </div>
                    </div>
                  </Link>
                );
              })
            )}
          </div>
        )}

      </section>
    </div>
  );
}
