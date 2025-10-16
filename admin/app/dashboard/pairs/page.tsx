'use client';

import { useState } from 'react';
import useSWR, { mutate } from 'swr';
import { adminApi, TradingPair } from '@/lib/api/admin';
import toast from 'react-hot-toast';

export default function PairsPage() {
  const { data: pairs = [], isLoading } = useSWR('/admin/pairs', () => adminApi.getTradingPairs(), {
    refreshInterval: 5000, // 每5秒自动刷新
  });

  const [showModal, setShowModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showInitModal, setShowInitModal] = useState(false);
  const [selectedPair, setSelectedPair] = useState<TradingPair | null>(null);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [formData, setFormData] = useState({
    symbol: '',
    base_asset: '',
    quote_asset: '',
    min_price: '',
    max_price: '',
    min_qty: '',
    max_qty: '',
  });
  const [editFormData, setEditFormData] = useState({
    symbol: '',
    base_asset: '',
    quote_asset: '',
    min_price: '',
    max_price: '',
    min_qty: '',
    max_qty: '',
  });

  // 初始化默认时间范围
  useState(() => {
    const now = new Date();
    const sixMonthsAgo = new Date();
    sixMonthsAgo.setMonth(now.getMonth() - 6);
    
    setStartTime(sixMonthsAgo.toISOString().split('T')[0]);
    setEndTime(now.toISOString().split('T')[0]);
  });

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
      mutate('/admin/pairs'); // 刷新数据
      toast.success('交易对创建成功');
    } catch (error) {
      toast.error('创建失败');
    }
  };

  const handleStatusChange = async (id: string, newStatus: string) => {
    try {
      await adminApi.updateTradingPairStatus(id, newStatus);
      mutate('/admin/pairs'); // 立即刷新数据
      toast.success('状态更新成功');
    } catch (error: any) {
      console.error('更新失败:', error);
      toast.error('更新失败: ' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleSimulatorToggle = async (id: string, enabled: boolean) => {
    try {
      await adminApi.updateTradingPairSimulator(id, enabled);
      mutate('/admin/pairs'); // 立即刷新数据
      toast.success(enabled ? '🤖 模拟器已启用，盘口将自动活跃' : '模拟器已停用');
    } catch (error: any) {
      console.error('更新失败:', error);
      toast.error('更新失败: ' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleEdit = (pair: TradingPair) => {
    setSelectedPair(pair);
    setEditFormData({
      symbol: pair.symbol,
      base_asset: pair.base_asset,
      quote_asset: pair.quote_asset,
      min_price: pair.min_price,
      max_price: pair.max_price,
      min_qty: pair.min_qty,
      max_qty: pair.max_qty,
    });
    setShowEditModal(true);
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPair) return;

    try {
      await adminApi.updateTradingPair(selectedPair.id, editFormData);
      setShowEditModal(false);
      mutate('/admin/pairs');
      toast.success('交易对更新成功');
    } catch (error: any) {
      toast.error('更新失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleInitData = (pair: TradingPair) => {
    setSelectedPair(pair);
    setShowInitModal(true);
  };

  const handleGenerateTradeData = async () => {
    if (!selectedPair) return;

    try {
      const result = await adminApi.generateTradeDataForPair(
        selectedPair.symbol,
        startTime,
        endTime
      );
      toast.success(`交易数据生成任务已创建\n任务ID: ${result.task_id}\n任务会在后台队列中执行`);
      setShowInitModal(false);
    } catch (error: any) {
      toast.error('生成失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleGenerateKlineData = async (pair: TradingPair) => {
    if (!confirm(`确定要为 ${pair.symbol} 生成K线数据吗？\n请确保已经生成交易数据`)) {
      return;
    }

    try {
      const result = await adminApi.generateKlineDataForPair(pair.symbol);
      toast.success(`K线数据生成任务已创建\n任务ID: ${result.task_id}\n任务会在后台队列中执行`);
    } catch (error: any) {
      toast.error('生成失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">交易对管理</h1>
        <button
          onClick={() => setShowModal(true)}
          className="px-6 py-2 bg-primary hover:bg-primary-dark rounded-lg transition"
        >
          添加交易对
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-400">加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4">ID</th>
                <th className="text-left p-4">交易对</th>
                <th className="text-left p-4">基础资产</th>
                <th className="text-left p-4">报价资产</th>
                <th className="text-left p-4">状态</th>
                <th className="text-left p-4">模拟器</th>
                <th className="text-right p-4">操作</th>
              </tr>
            </thead>
            <tbody>
              {pairs.length === 0 ? (
                <tr>
                  <td colSpan={7} className="text-center p-8 text-gray-400">
                    暂无数据
                  </td>
                </tr>
              ) : (
                pairs.map((pair) => (
                  <tr key={pair.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                    <td className="p-4 text-xs font-mono text-gray-400">
                      {pair.id.substring(0, 8)}...
                    </td>
                    <td className="p-4 font-semibold">{pair.symbol}</td>
                    <td className="p-4">{pair.base_asset}</td>
                    <td className="p-4">{pair.quote_asset}</td>
                    <td className="p-4">
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          pair.status === 'active'
                            ? 'bg-green-500/20 text-green-500'
                            : 'bg-gray-700 text-gray-400'
                        }`}
                      >
                        {pair.status === 'active' ? '启用' : '禁用'}
                      </span>
                    </td>
                    <td className="p-4">
                      <button
                        onClick={() => handleSimulatorToggle(pair.id, !pair.simulator_enabled)}
                        className={`px-2 py-1 rounded text-xs transition ${
                          pair.simulator_enabled
                            ? 'bg-blue-500/20 text-blue-400 hover:bg-blue-500/30'
                            : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
                        }`}
                      >
                        {pair.simulator_enabled ? '🤖 已开启' : '关闭'}
                      </button>
                    </td>
                    <td className="p-4 text-right">
                      <div className="flex items-center justify-end gap-2 flex-wrap">
                        <button
                          onClick={() => handleEdit(pair)}
                          className="px-3 py-1 bg-gray-600 hover:bg-gray-700 rounded text-xs transition"
                        >
                          编辑
                        </button>
                        <button
                          onClick={() => handleInitData(pair)}
                          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 rounded text-xs transition"
                        >
                          初始化数据
                        </button>
                        <button
                          onClick={() => handleGenerateKlineData(pair)}
                          className="px-3 py-1 bg-purple-600 hover:bg-purple-700 rounded text-xs transition"
                        >
                          生成K线
                        </button>
                        <button
                          onClick={() =>
                            handleStatusChange(
                              pair.id,
                              pair.status === 'active' ? 'inactive' : 'active'
                            )
                          }
                          className={`px-3 py-1 rounded text-xs transition ${
                            pair.status === 'active'
                              ? 'bg-red-600 hover:bg-red-700'
                              : 'bg-green-600 hover:bg-green-700'
                          }`}
                        >
                          {pair.status === 'active' ? '禁用' : '启用'}
                        </button>
                      </div>
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

      {/* 编辑交易对模态框 */}
      {showEditModal && selectedPair && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-[#0f1429] rounded-lg p-6 w-[500px] border border-gray-800">
            <h2 className="text-xl font-bold mb-4">编辑交易对</h2>
            <form onSubmit={handleEditSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">交易对符号</label>
                  <input
                    type="text"
                    value={editFormData.symbol}
                    onChange={(e) => setEditFormData({ ...editFormData, symbol: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="BTC/USDT"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">基础资产</label>
                  <input
                    type="text"
                    value={editFormData.base_asset}
                    onChange={(e) => setEditFormData({ ...editFormData, base_asset: e.target.value })}
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
                  value={editFormData.quote_asset}
                  onChange={(e) => setEditFormData({ ...editFormData, quote_asset: e.target.value })}
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
                    value={editFormData.min_price}
                    onChange={(e) => setEditFormData({ ...editFormData, min_price: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="0.01"
                    step="0.01"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最大价格</label>
                  <input
                    type="number"
                    value={editFormData.max_price}
                    onChange={(e) => setEditFormData({ ...editFormData, max_price: e.target.value })}
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
                    value={editFormData.min_qty}
                    onChange={(e) => setEditFormData({ ...editFormData, min_qty: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    placeholder="0.0001"
                    step="0.0001"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">最大数量</label>
                  <input
                    type="number"
                    value={editFormData.max_qty}
                    onChange={(e) => setEditFormData({ ...editFormData, max_qty: e.target.value })}
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
                  更新
                </button>
                <button
                  type="button"
                  onClick={() => setShowEditModal(false)}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 初始化数据模态框 */}
      {showInitModal && selectedPair && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-[#0f1429] rounded-lg p-6 w-[500px] border border-gray-800">
            <h2 className="text-xl font-bold mb-4">初始化数据 - {selectedPair.symbol}</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-gray-400 mb-2">开始时间</label>
                <input
                  type="date"
                  value={startTime}
                  onChange={(e) => setStartTime(e.target.value)}
                  className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-2">结束时间</label>
                <input
                  type="date"
                  value={endTime}
                  onChange={(e) => setEndTime(e.target.value)}
                  className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                />
              </div>
              <div className="p-3 bg-yellow-500/10 border border-yellow-500/30 rounded text-sm text-yellow-400">
                ⚠️ 将生成 {selectedPair.symbol} 从 {startTime} 到 {endTime} 的历史交易数据，任务会在后台队列中执行
              </div>
              <div className="flex space-x-4">
                <button
                  onClick={handleGenerateTradeData}
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition"
                >
                  开始生成
                </button>
                <button
                  onClick={() => setShowInitModal(false)}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
