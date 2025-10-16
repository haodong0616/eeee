'use client';

import { useAppSelector } from '@/lib/store/hooks';
import { useGetOrdersQuery } from '@/lib/services/api';

interface MyOrdersProps {
  type: 'open' | 'history';
  symbol: string;
  onCancel: (orderId: number) => void;
}

export default function MyOrders({ type, symbol, onCancel }: MyOrdersProps) {
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  const { data: allOrders = [] } = useGetOrdersQuery(
    { symbol },
    { skip: !isAuthenticated, pollingInterval: 5000 }
  );

  const currentOrders = allOrders.filter(
    (order) => order.status === 'pending' || order.status === 'partial'
  );
  const historyOrders = allOrders.filter(
    (order) => order.status === 'filled' || order.status === 'cancelled' || order.status === 'partial_cancelled'
  );

  const orders = type === 'open' ? currentOrders : historyOrders;

  if (!isAuthenticated) {
    return (
      <div className="p-8 text-center text-gray-400">
        请先连接钱包
      </div>
    );
  }

  return (
    <div className="p-4">
      {orders.length === 0 ? (
        <div className="py-8 text-center text-gray-400">
          暂无{type === 'open' ? '当前' : '历史'}委托
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="text-sm text-gray-400 border-b border-gray-800">
                <th className="text-left pb-2">时间</th>
                <th className="text-left pb-2">交易对</th>
                <th className="text-left pb-2">类型</th>
                <th className="text-left pb-2">方向</th>
                <th className="text-right pb-2">价格</th>
                <th className="text-right pb-2">数量</th>
                <th className="text-right pb-2">已成交</th>
                <th className="text-left pb-2">状态</th>
                {type === 'open' && <th className="text-right pb-2">操作</th>}
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id} className="border-b border-gray-800">
                  <td className="py-3 text-sm">
                    {new Date(order.created_at).toLocaleString()}
                  </td>
                  <td className="py-3">{order.symbol}</td>
                  <td className="py-3 text-sm">{order.order_type}</td>
                  <td className="py-3">
                    <span
                      className={`${
                        order.side === 'buy' ? 'text-buy' : 'text-sell'
                      }`}
                    >
                      {order.side === 'buy' ? '买入' : '卖出'}
                    </span>
                  </td>
                  <td className="py-3 text-right">{parseFloat(order.price).toFixed(2)}</td>
                  <td className="py-3 text-right">
                    {parseFloat(order.quantity).toFixed(4)}
                  </td>
                  <td className="py-3 text-right">
                    {parseFloat(order.filled_qty).toFixed(4)}
                  </td>
                  <td className="py-3">
                    <span
                      className={`text-xs px-2 py-1 rounded ${
                        order.status === 'filled'
                          ? 'bg-buy/20 text-buy'
                          : order.status === 'cancelled'
                          ? 'bg-gray-700 text-gray-400'
                          : order.status === 'partial_cancelled'
                          ? 'bg-orange-500/20 text-orange-400'
                          : order.status === 'partial'
                          ? 'bg-yellow-500/20 text-yellow-400'
                          : 'bg-primary/20 text-primary'
                      }`}
                    >
                      {order.status === 'filled'
                        ? '已成交'
                        : order.status === 'cancelled'
                        ? '已取消'
                        : order.status === 'partial_cancelled'
                        ? '部分成交已取消'
                        : order.status === 'partial'
                        ? '部分成交'
                        : '待成交'}
                    </span>
                  </td>
                  {type === 'open' && (
                    <td className="py-3 text-right">
                      <button
                        onClick={() => onCancel(order.id)}
                        className="text-sell hover:underline text-sm"
                      >
                        取消
                      </button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

