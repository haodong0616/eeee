'use client';

import { useEffect, useState } from 'react';
import { adminApi, Order } from '@/lib/api/admin';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    loadOrders();
  }, []);

  const loadOrders = async () => {
    try {
      const data = await adminApi.getOrders();
      setOrders(data);
    } catch (error) {
      console.error('Failed to load orders:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredOrders = filter === 'all' 
    ? orders 
    : orders.filter((order) => order.status === filter);

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">订单管理</h1>
        <div className="flex space-x-2">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-lg ${
              filter === 'all' ? 'bg-primary' : 'bg-gray-700'
            }`}
          >
            全部
          </button>
          <button
            onClick={() => setFilter('pending')}
            className={`px-4 py-2 rounded-lg ${
              filter === 'pending' ? 'bg-primary' : 'bg-gray-700'
            }`}
          >
            待成交
          </button>
          <button
            onClick={() => setFilter('filled')}
            className={`px-4 py-2 rounded-lg ${
              filter === 'filled' ? 'bg-primary' : 'bg-gray-700'
            }`}
          >
            已成交
          </button>
          <button
            onClick={() => setFilter('cancelled')}
            className={`px-4 py-2 rounded-lg ${
              filter === 'cancelled' ? 'bg-primary' : 'bg-gray-700'
            }`}
          >
            已取消
          </button>
        </div>
      </div>

      {loading ? (
        <div>加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-[#151a35]">
                <tr>
                  <th className="text-left p-4">ID</th>
                  <th className="text-left p-4">用户</th>
                  <th className="text-left p-4">交易对</th>
                  <th className="text-left p-4">类型</th>
                  <th className="text-left p-4">方向</th>
                  <th className="text-right p-4">价格</th>
                  <th className="text-right p-4">数量</th>
                  <th className="text-right p-4">已成交</th>
                  <th className="text-left p-4">状态</th>
                  <th className="text-left p-4">时间</th>
                </tr>
              </thead>
              <tbody>
                {filteredOrders.length === 0 ? (
                  <tr>
                    <td colSpan={10} className="text-center p-8 text-gray-400">
                      暂无数据
                    </td>
                  </tr>
                ) : (
                  filteredOrders.map((order) => (
                    <tr key={order.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                      <td className="p-4">{order.id}</td>
                      <td className="p-4 font-mono text-xs">
                        {order.user?.wallet_address.slice(0, 10)}...
                      </td>
                      <td className="p-4">{order.symbol}</td>
                      <td className="p-4">{order.order_type}</td>
                      <td className="p-4">
                        <span
                          className={
                            order.side === 'buy' ? 'text-green-500' : 'text-red-500'
                          }
                        >
                          {order.side}
                        </span>
                      </td>
                      <td className="p-4 text-right">{parseFloat(order.price).toFixed(2)}</td>
                      <td className="p-4 text-right">{parseFloat(order.quantity).toFixed(4)}</td>
                      <td className="p-4 text-right">
                        {parseFloat(order.filled_qty).toFixed(4)}
                      </td>
                      <td className="p-4">
                        <span
                          className={`px-2 py-1 rounded text-xs ${
                            order.status === 'filled'
                              ? 'bg-green-500/20 text-green-500'
                              : order.status === 'cancelled'
                              ? 'bg-gray-700 text-gray-400'
                              : 'bg-primary/20 text-primary'
                          }`}
                        >
                          {order.status}
                        </span>
                      </td>
                      <td className="p-4 text-sm">
                        {new Date(order.created_at).toLocaleString()}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}


