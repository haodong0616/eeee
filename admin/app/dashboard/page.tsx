'use client';

import { useEffect, useState } from 'react';
import { adminApi, Stats } from '@/lib/api/admin';

export default function DashboardPage() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const data = await adminApi.getStats();
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">数据概览</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">总用户数</p>
              <p className="text-3xl font-bold">{stats?.user_count || 0}</p>
            </div>
            <div className="text-4xl">👥</div>
          </div>
        </div>

        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">总订单数</p>
              <p className="text-3xl font-bold">{stats?.order_count || 0}</p>
            </div>
            <div className="text-4xl">📋</div>
          </div>
        </div>

        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">总成交数</p>
              <p className="text-3xl font-bold">{stats?.trade_count || 0}</p>
            </div>
            <div className="text-4xl">💱</div>
          </div>
        </div>

        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">总交易量</p>
              <p className="text-3xl font-bold">
                {stats?.total_volume ? parseFloat(stats.total_volume).toFixed(2) : '0.00'}
              </p>
            </div>
            <div className="text-4xl">💰</div>
          </div>
        </div>
      </div>

      <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6">
        <h2 className="text-xl font-bold mb-4">系统信息</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-400">系统版本</span>
            <span>v1.0.0</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-400">运行状态</span>
            <span className="text-green-500">正常</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-400">最后更新</span>
            <span>{new Date().toLocaleString()}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

