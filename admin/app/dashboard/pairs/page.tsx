'use client';

import { useEffect, useState } from 'react';
import { adminApi, TradingPair } from '@/lib/api/admin';
import { marketApi } from '@/lib/api/market';

export default function PairsPage() {
  const [pairs, setPairs] = useState<TradingPair[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    symbol: '',
    base_asset: '',
    quote_asset: '',
    min_price: '',
    max_price: '',
    min_qty: '',
    max_qty: '',
  });

  useEffect(() => {
    loadPairs();
  }, []);

  const loadPairs = async () => {
    try {
      const data = await marketApi.getTradingPairs();
      setPairs(data);
    } catch (error) {
      console.error('Failed to load pairs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await adminApi.createTradingPair(formData);
      setShowModal(false);
      setFormData({
        symbol: '',
        base_asset: '',
        quote_asset: '',
        min_price: '',
        max_price: '',
        min_qty: '',
        max_qty: '',
      });
      loadPairs();
      alert('交易对创建成功');
    } catch (error) {
      alert('创建失败');
    }
  };

  const handleStatusChange = async (id: number, status: string) => {
    try {
      await adminApi.updateTradingPairStatus(id, status);
      loadPairs();
      alert('状态更新成功');
    } catch (error) {
      alert('更新失败');
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">交易对管理</h1>
        <button
          onClick={() => setShowModal(true)}
          className="px-6 py-2 bg-primary hover:bg-primary-dark rounded-lg transition"
        >
          添加交易对
        </button>
      </div>

      {loading ? (
        <div>加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4">ID</th>
                <th className="text-left p-4">交易对</th>
                <th className="text-left p-4">基础资产</th>
                <th className="text-left p-4">报价资产</th>
                <th className="text-right p-4">最小价格</th>
                <th className="text-right p-4">最大价格</th>
                <th className="text-left p-4">状态</th>
                <th className="text-right p-4">操作</th>
              </tr>
            </thead>
            <tbody>
              {pairs.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-center p-8 text-gray-400">
                    暂无数据
                  </td>
                </tr>
              ) : (
                pairs.map((pair) => (
                  <tr key={pair.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                    <td className="p-4">{pair.id}</td>
                    <td className="p-4 font-semibold">{pair.symbol}</td>
                    <td className="p-4">{pair.base_asset}</td>
                    <td className="p-4">{pair.quote_asset}</td>
                    <td className="p-4 text-right">{pair.min_price || '-'}</td>
                    <td className="p-4 text-right">{pair.max_price || '-'}</td>
                    <td className="p-4">
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          pair.status === 'active'
                            ? 'bg-green-500/20 text-green-500'
                            : 'bg-gray-700 text-gray-400'
                        }`}
                      >
                        {pair.status}
                      </span>
                    </td>
                    <td className="p-4 text-right">
                      <button
                        onClick={() =>
                          handleStatusChange(
                            pair.id,
                            pair.status === 'active' ? 'inactive' : 'active'
                          )
                        }
                        className="text-primary hover:underline text-sm"
                      >
                        {pair.status === 'active' ? '禁用' : '启用'}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* 添加交易对模态框 */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-[#0f1429] rounded-lg p-6 w-[500px] border border-gray-800">
            <h2 className="text-xl font-bold mb-4">添加交易对</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">交易对符号</label>
                  <input
                    type="text"
                    value={formData.symbol}
                    onChange={(e) => setFormData({ ...formData, symbol: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="BTC/USDT"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">基础资产</label>
                  <input
                    type="text"
                    value={formData.base_asset}
                    onChange={(e) => setFormData({ ...formData, base_asset: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="BTC"
                    required
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-2">报价资产</label>
                <input
                  type="text"
                  value={formData.quote_asset}
                  onChange={(e) => setFormData({ ...formData, quote_asset: e.target.value })}
                  className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  placeholder="USDT"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最小价格</label>
                  <input
                    type="number"
                    value={formData.min_price}
                    onChange={(e) => setFormData({ ...formData, min_price: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="0.01"
                    step="0.01"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最大价格</label>
                  <input
                    type="number"
                    value={formData.max_price}
                    onChange={(e) => setFormData({ ...formData, max_price: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="1000000"
                    step="0.01"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最小数量</label>
                  <input
                    type="number"
                    value={formData.min_qty}
                    onChange={(e) => setFormData({ ...formData, min_qty: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="0.0001"
                    step="0.0001"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最大数量</label>
                  <input
                    type="number"
                    value={formData.max_qty}
                    onChange={(e) => setFormData({ ...formData, max_qty: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="10000"
                    step="0.0001"
                  />
                </div>
              </div>
              <div className="flex space-x-4">
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg transition"
                >
                  创建
                </button>
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

