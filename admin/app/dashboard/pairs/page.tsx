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
  const [showBatchActivityModal, setShowBatchActivityModal] = useState(false);
  const [showBatchInitModal, setShowBatchInitModal] = useState(false);
  const [showBatchKlineModal, setShowBatchKlineModal] = useState(false);
  const [selectedPair, setSelectedPair] = useState<TradingPair | null>(null);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  
  // 批量活跃度设置
  const [batchActivity, setBatchActivity] = useState({
    activity_level: 5,
    orderbook_depth: 15,
    trade_frequency: 20,
    price_volatility: '0.01',
  });
  
  // 批量初始化数据设置
  const [batchInitData, setBatchInitData] = useState({
    start_time: '',
    end_time: '',
    trade_count: 0,
    generate_klines: true,
  });
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
    activity_level: 5,
    orderbook_depth: 15,
    trade_frequency: 20,
    price_volatility: '0.01',
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
      activity_level: pair.activity_level || 5,
      orderbook_depth: pair.orderbook_depth || 15,
      trade_frequency: pair.trade_frequency || 20,
      price_volatility: pair.price_volatility || '0.01',
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

  // 批量设置活跃度
  const handleBatchActivitySubmit = async () => {
    try {
      const result = await adminApi.batchUpdatePairsActivity({
        activity_level: batchActivity.activity_level,
        orderbook_depth: batchActivity.orderbook_depth,
        trade_frequency: batchActivity.trade_frequency,
        price_volatility: batchActivity.price_volatility,
      });
      toast.success(`✅ 已更新 ${result.affected_count} 个交易对的活跃度配置\n配置将在10秒内生效`);
      setShowBatchActivityModal(false);
      mutate('/admin/pairs');
    } catch (error: any) {
      toast.error('批量设置失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  // 批量生成初始化数据
  const handleBatchInitSubmit = async () => {
    if (!batchInitData.start_time || !batchInitData.end_time) {
      toast.error('请选择开始和结束时间');
      return;
    }

    const days = Math.ceil((new Date(batchInitData.end_time).getTime() - new Date(batchInitData.start_time).getTime()) / (1000 * 60 * 60 * 24));
    const tradeCount = batchInitData.trade_count || days * 3;

    if (!confirm(
      `确定要为所有交易对批量生成初始化数据吗？\n\n` +
      `时间范围：${batchInitData.start_time} ~ ${batchInitData.end_time} (${days}天)\n` +
      `交易对数量：${pairs.length} 个\n` +
      `预计每个币种生成：${tradeCount} 笔交易\n` +
      `预计总交易数：${tradeCount * pairs.length} 笔\n` +
      `生成K线：${batchInitData.generate_klines ? '是' : '否'}\n\n` +
      `此操作可能需要较长时间，请耐心等待`
    )) {
      return;
    }

    try {
      const result = await adminApi.batchGenerateInitData({
        start_time: batchInitData.start_time,
        end_time: batchInitData.end_time,
        trade_count: batchInitData.trade_count,
        generate_klines: batchInitData.generate_klines,
      });
      toast.success(
        `🚀 批量初始化任务已创建\n` +
        `交易对数量：${result.pair_count}\n` +
        `任务数量：${result.task_count}\n` +
        `请在任务管理中查看进度`
      );
      setShowBatchInitModal(false);
    } catch (error: any) {
      toast.error('批量初始化失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  // 批量生成K线
  const handleBatchKlineSubmit = async () => {
    if (!confirm(
      `确定要为所有交易对批量生成K线数据吗？\n\n` +
      `交易对数量：${pairs.length} 个\n` +
      `将基于现有交易数据生成K线\n` +
      `包含周期：1m, 5m, 15m, 30m, 1h, 4h, 1d\n\n` +
      `预计耗时：${pairs.length} × 2-3秒 ≈ ${pairs.length * 2}-${pairs.length * 3}秒\n` +
      `任务会在后台队列中执行，请耐心等待`
    )) {
      return;
    }

    try {
      const result = await adminApi.batchGenerateKlines({});
      toast.success(
        `📊 批量K线生成任务已创建\n` +
        `交易对数量：${result.pair_count}\n` +
        `任务数量：${result.task_count}\n` +
        `请在任务管理中查看进度`
      );
      setShowBatchKlineModal(false);
    } catch (error: any) {
      toast.error('批量生成K线失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">交易对管理</h1>
        <div className="flex gap-2">
          <button
            onClick={() => setShowBatchActivityModal(true)}
            className="px-3 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition text-sm"
            title="一键设置所有交易对的活跃度参数"
          >
            🎮 批量活跃度
          </button>
          <button
            onClick={() => setShowBatchInitModal(true)}
            className="px-3 py-2 bg-green-600 hover:bg-green-700 rounded-lg transition text-sm"
            title="一键为所有交易对生成历史数据"
          >
            🚀 批量初始化
          </button>
          <button
            onClick={() => setShowBatchKlineModal(true)}
            className="px-3 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg transition text-sm"
            title="一键为所有交易对生成K线数据"
          >
            📊 批量K线
          </button>
          <button
            onClick={() => setShowModal(true)}
            className="px-6 py-2 bg-primary hover:bg-primary-dark rounded-lg transition"
          >
            添加交易对
          </button>
        </div>
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

              {/* 活跃度配置 */}
              <div className="border-t border-gray-800 pt-4 mt-4">
                <h3 className="text-lg font-bold mb-4 text-primary">🎮 活跃度配置</h3>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">
                      活跃度等级 (1-10)
                      <span className="text-xs ml-2">
                        {editFormData.activity_level <= 3 ? '🐢低' : 
                         editFormData.activity_level <= 6 ? '🚶中' : 
                         editFormData.activity_level <= 8 ? '🏃高' : '🚀极高'}
                      </span>
                    </label>
                    <input
                      type="range"
                      min="1"
                      max="10"
                      value={editFormData.activity_level}
                      onChange={(e) => setEditFormData({ ...editFormData, activity_level: parseInt(e.target.value) })}
                      className="w-full"
                    />
                    <div className="flex justify-between text-xs text-gray-500 mt-1">
                      <span>1</span>
                      <span className="font-bold text-primary">{editFormData.activity_level}</span>
                      <span>10</span>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm text-gray-400 mb-2">订单簿深度 (5-30档)</label>
                    <input
                      type="number"
                      min="5"
                      max="30"
                      value={editFormData.orderbook_depth}
                      onChange={(e) => setEditFormData({ ...editFormData, orderbook_depth: parseInt(e.target.value) })}
                      className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    />
                    <div className="text-xs text-gray-500 mt-1">买{editFormData.orderbook_depth}档 + 卖{editFormData.orderbook_depth}档 = {editFormData.orderbook_depth * 2}个订单</div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 mt-4">
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">成交频率 (5-60秒)</label>
                    <input
                      type="number"
                      min="5"
                      max="60"
                      value={editFormData.trade_frequency}
                      onChange={(e) => setEditFormData({ ...editFormData, trade_frequency: parseInt(e.target.value) })}
                      className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    />
                    <div className="text-xs text-gray-500 mt-1">实际: {Math.floor(editFormData.trade_frequency * 0.7)}-{Math.ceil(editFormData.trade_frequency * 1.3)}秒</div>
                  </div>

                  <div>
                    <label className="block text-sm text-gray-400 mb-2">价格波动率 (0.001-0.05)</label>
                    <input
                      type="number"
                      min="0.001"
                      max="0.05"
                      step="0.001"
                      value={editFormData.price_volatility}
                      onChange={(e) => setEditFormData({ ...editFormData, price_volatility: e.target.value })}
                      className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                    />
                    <div className="text-xs text-gray-500 mt-1">{(parseFloat(editFormData.price_volatility) * 100).toFixed(1)}% 基础波动</div>
                  </div>
                </div>

                <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-3 mt-4">
                  <div className="text-xs text-blue-400">
                    <div className="font-bold mb-1">💡 预计效果：</div>
                    <div>• 订单簿每 <strong>{Math.max(4, 24 - editFormData.activity_level * 2)}秒</strong> 更新一次</div>
                    <div>• 价格分布范围: ±<strong>{(parseFloat(editFormData.price_volatility) * editFormData.activity_level * 50).toFixed(1)}%</strong></div>
                    <div>• 成交量波动: <strong>{editFormData.activity_level <= 3 ? '小' : editFormData.activity_level <= 6 ? '中等' : '大'}</strong></div>
                  </div>
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

      {/* 批量设置活跃度弹窗 */}
      {showBatchActivityModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-2xl font-bold mb-6">🎮 批量设置所有交易对活跃度</h2>
            <div className="space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">
                    活跃度等级 (1-10)
                    <span className="text-xs ml-2">
                      {batchActivity.activity_level <= 3 ? '🐢低' : 
                       batchActivity.activity_level <= 6 ? '🚶中' : 
                       batchActivity.activity_level <= 8 ? '🏃高' : '🚀极高'}
                    </span>
                  </label>
                  <input
                    type="range"
                    min="1"
                    max="10"
                    value={batchActivity.activity_level}
                    onChange={(e) => setBatchActivity({ ...batchActivity, activity_level: parseInt(e.target.value) })}
                    className="w-full"
                  />
                  <div className="flex justify-between text-xs text-gray-500 mt-1">
                    <span>1</span>
                    <span className="font-bold text-primary">{batchActivity.activity_level}</span>
                    <span>10</span>
                  </div>
                </div>

                <div>
                  <label className="block text-sm text-gray-400 mb-2">订单簿深度 (5-30档)</label>
                  <input
                    type="number"
                    min="5"
                    max="30"
                    value={batchActivity.orderbook_depth}
                    onChange={(e) => setBatchActivity({ ...batchActivity, orderbook_depth: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  />
                  <div className="text-xs text-gray-500 mt-1">买{batchActivity.orderbook_depth}档 + 卖{batchActivity.orderbook_depth}档 = {batchActivity.orderbook_depth * 2}个订单</div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">成交频率 (5-60秒)</label>
                  <input
                    type="number"
                    min="5"
                    max="60"
                    value={batchActivity.trade_frequency}
                    onChange={(e) => setBatchActivity({ ...batchActivity, trade_frequency: parseInt(e.target.value) })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  />
                  <div className="text-xs text-gray-500 mt-1">实际: {Math.floor(batchActivity.trade_frequency * 0.7)}-{Math.ceil(batchActivity.trade_frequency * 1.3)}秒</div>
                </div>

                <div>
                  <label className="block text-sm text-gray-400 mb-2">价格波动率 (0.001-0.05)</label>
                  <input
                    type="number"
                    min="0.001"
                    max="0.05"
                    step="0.001"
                    value={batchActivity.price_volatility}
                    onChange={(e) => setBatchActivity({ ...batchActivity, price_volatility: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  />
                  <div className="text-xs text-gray-500 mt-1">{(parseFloat(batchActivity.price_volatility) * 100).toFixed(1)}% 基础波动</div>
                </div>
              </div>

              <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
                <div className="text-sm text-blue-400">
                  <div className="font-bold mb-2">💡 预计效果（所有交易对）：</div>
                  <div>• 订单簿每 <strong>{Math.max(4, 24 - batchActivity.activity_level * 2)}秒</strong> 更新一次</div>
                  <div>• 价格分布范围: ±<strong>{(parseFloat(batchActivity.price_volatility) * batchActivity.activity_level * 50).toFixed(1)}%</strong></div>
                  <div>• 成交量波动: <strong>{batchActivity.activity_level <= 3 ? '小' : batchActivity.activity_level <= 6 ? '中等' : '大'}</strong></div>
                  <div className="mt-2 text-xs text-gray-400">
                    将更新所有启用模拟器的交易对（共 {pairs.filter((p: TradingPair) => p.simulator_enabled).length} 个）
                  </div>
                </div>
              </div>

              <div className="flex space-x-4">
                <button
                  onClick={handleBatchActivitySubmit}
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition"
                >
                  批量设置
                </button>
                <button
                  onClick={() => setShowBatchActivityModal(false)}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 批量生成K线弹窗 */}
      {showBatchKlineModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6 w-full max-w-xl">
            <h2 className="text-2xl font-bold mb-6">📊 批量生成K线数据</h2>
            <div className="space-y-6">
              <div className="bg-purple-500/10 border border-purple-500/30 rounded-lg p-4">
                <div className="text-sm text-purple-400">
                  <div className="font-bold mb-2">📈 将为所有交易对生成K线：</div>
                  <div>• 交易对数量: <strong>{pairs.length} 个</strong></div>
                  <div>• K线周期: <strong>1m, 5m, 15m, 30m, 1h, 4h, 1d</strong></div>
                  <div>• 数据来源: <strong>基于现有交易记录</strong></div>
                  <div className="mt-2 text-xs text-yellow-400">
                    ⏱️ 预计耗时：{pairs.length * 2}-{pairs.length * 3}秒（{Math.ceil(pairs.length * 2 / 60)}-{Math.ceil(pairs.length * 3 / 60)}分钟）
                  </div>
                </div>
              </div>

              <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-3">
                <div className="text-xs text-blue-400">
                  <div className="font-bold mb-1">💡 适用场景：</div>
                  <div>• 已有交易数据，但缺少K线</div>
                  <div>• K线数据有问题，需要重新生成</div>
                  <div>• 新增了交易记录，需要更新K线</div>
                </div>
              </div>

              <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3">
                <div className="text-xs text-yellow-400">
                  <div className="font-bold mb-1">⚠️ 注意事项：</div>
                  <div>• 请确保已有交易数据，否则K线为空</div>
                  <div>• 任务会在后台队列中执行</div>
                  <div>• 在任务管理页面可以查看进度</div>
                  <div>• 重复生成会覆盖旧K线数据</div>
                </div>
              </div>

              <div className="flex space-x-4">
                <button
                  onClick={handleBatchKlineSubmit}
                  className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg transition"
                >
                  开始批量生成
                </button>
                <button
                  onClick={() => setShowBatchKlineModal(false)}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 批量初始化数据弹窗 */}
      {showBatchInitModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-2xl font-bold mb-6">🚀 批量生成所有交易对初始化数据</h2>
            <div className="space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-400 mb-2">开始时间</label>
                  <input
                    type="date"
                    value={batchInitData.start_time}
                    onChange={(e) => setBatchInitData({ ...batchInitData, start_time: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-2">结束时间</label>
                  <input
                    type="date"
                    value={batchInitData.end_time}
                    onChange={(e) => setBatchInitData({ ...batchInitData, end_time: e.target.value })}
                    className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm text-gray-400 mb-2">
                  建议总数（可选，留空自动计算）
                  <span className="text-xs ml-2 text-gray-500">每个币种生成的交易笔数</span>
                </label>
                <input
                  type="number"
                  min="0"
                  placeholder="默认：天数 × 3"
                  value={batchInitData.trade_count || ''}
                  onChange={(e) => setBatchInitData({ ...batchInitData, trade_count: parseInt(e.target.value) || 0 })}
                  className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
                />
                <div className="text-xs text-gray-500 mt-1">
                  {batchInitData.start_time && batchInitData.end_time ? (
                    <>
                      时间跨度：{Math.ceil((new Date(batchInitData.end_time).getTime() - new Date(batchInitData.start_time).getTime()) / (1000 * 60 * 60 * 24))} 天
                      {batchInitData.trade_count === 0 && ` → 约 ${Math.ceil((new Date(batchInitData.end_time).getTime() - new Date(batchInitData.start_time).getTime()) / (1000 * 60 * 60 * 24)) * 3} 笔/币种`}
                    </>
                  ) : '请先选择时间范围'}
                </div>
              </div>

              <div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={batchInitData.generate_klines}
                    onChange={(e) => setBatchInitData({ ...batchInitData, generate_klines: e.target.checked })}
                    className="w-4 h-4"
                  />
                  <span className="text-sm text-gray-300">同时生成K线数据</span>
                </label>
                <div className="text-xs text-gray-500 mt-1">
                  推荐开启，K线数据基于交易数据生成，前端需要K线才能显示图表
                </div>
              </div>

              <div className="bg-green-500/10 border border-green-500/30 rounded-lg p-4">
                <div className="text-sm text-green-400">
                  <div className="font-bold mb-2">📊 预计生成：</div>
                  <div>• 交易对数量: <strong>{pairs.length} 个</strong></div>
                  {batchInitData.start_time && batchInitData.end_time && (
                    <>
                      <div>• 时间范围: <strong>{batchInitData.start_time} ~ {batchInitData.end_time}</strong></div>
                      <div>• 每个币种: <strong>
                        {batchInitData.trade_count || Math.ceil((new Date(batchInitData.end_time).getTime() - new Date(batchInitData.start_time).getTime()) / (1000 * 60 * 60 * 24)) * 3} 笔交易
                      </strong></div>
                      <div>• 预计总交易数: <strong>
                        {(batchInitData.trade_count || Math.ceil((new Date(batchInitData.end_time).getTime() - new Date(batchInitData.start_time).getTime()) / (1000 * 60 * 60 * 24)) * 3) * pairs.length} 笔
                      </strong></div>
                      <div>• 生成K线: <strong>{batchInitData.generate_klines ? '是' : '否'}</strong></div>
                    </>
                  )}
                  <div className="mt-2 text-xs text-yellow-400">
                    ⏱️ 预计耗时：{pairs.length} × 2-5秒 ≈ {pairs.length * 3}-{pairs.length * 5}秒（{Math.ceil(pairs.length * 3 / 60)}-{Math.ceil(pairs.length * 5 / 60)}分钟）
                  </div>
                </div>
              </div>

              <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3">
                <div className="text-xs text-yellow-400">
                  <div className="font-bold mb-1">⚠️ 注意事项：</div>
                  <div>• 任务会在后台队列中依次执行，请耐心等待</div>
                  <div>• 在任务管理页面可以查看执行进度</div>
                  <div>• 建议：6个月数据，每个币种约500-1000笔交易</div>
                  <div>• 生成过多数据会影响查询性能，建议适度</div>
                </div>
              </div>

              <div className="flex space-x-4">
                <button
                  onClick={handleBatchInitSubmit}
                  className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg transition"
                  disabled={!batchInitData.start_time || !batchInitData.end_time}
                >
                  开始批量生成
                </button>
                <button
                  onClick={() => setShowBatchInitModal(false)}
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

