'use client';

import { useState } from 'react';
import useSWR, { mutate } from 'swr';
import { getChains, updateChain, updateChainStatus, createChain } from '@/lib/api/admin';
import type { ChainConfig } from '@/lib/api/admin';
import toast from 'react-hot-toast';

export default function ChainsPage() {
  const { data: chains = [], isLoading, error } = useSWR('/admin/chains', getChains, {
    refreshInterval: 5000,
  });

  const [editingChain, setEditingChain] = useState<ChainConfig | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [formData, setFormData] = useState<Partial<ChainConfig>>({});

  const handleEdit = (chain: ChainConfig) => {
    setEditingChain(chain);
    setFormData(chain);
    setShowEditModal(true);
  };

  const handleCreate = () => {
    setEditingChain(null);
    setFormData({
      chain_name: '',
      chain_id: 0,
      rpc_url: '',
      block_explorer_url: '',
      usdt_contract_address: '',
      usdt_decimals: 18,
      platform_deposit_address: '',
      platform_withdraw_private_key: '',
      enabled: true,
    });
    setShowEditModal(true);
  };

  const handleSave = async () => {
    try {
      if (editingChain) {
        await updateChain(editingChain.id, formData);
      } else {
        await createChain(formData);
      }
      mutate('/admin/chains');
      setShowEditModal(false);
      toast.success(editingChain ? '链配置已更新' : '链配置已创建');
    } catch (error: any) {
      toast.error(error.response?.data?.error || '操作失败');
    }
  };

  const handleToggleStatus = async (chain: ChainConfig) => {
    if (!confirm(`确定要${chain.enabled ? '禁用' : '启用'}${chain.chain_name}吗？`)) {
      return;
    }

    try {
      await updateChainStatus(chain.id, !chain.enabled);
      mutate('/admin/chains');
      toast.success(`${chain.chain_name}已${chain.enabled ? '禁用' : '启用'}`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || '操作失败');
    }
  };

  if (isLoading) {
    return <div className="p-6">加载中...</div>;
  }

  if (error) {
    return <div className="p-6 text-red-600">加载失败: {error.message}</div>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">链配置管理</h1>
        <button
          onClick={handleCreate}
          className="px-4 py-2 bg-primary hover:bg-primary-dark rounded-lg transition text-sm"
        >
          + 添加链
        </button>
      </div>

      <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4 text-gray-400 font-semibold">链名称</th>
                <th className="text-left p-4 text-gray-400 font-semibold">Chain ID</th>
                <th className="text-left p-4 text-gray-400 font-semibold">USDT合约</th>
                <th className="text-left p-4 text-gray-400 font-semibold">收款地址</th>
                <th className="text-left p-4 text-gray-400 font-semibold">状态</th>
                <th className="text-left p-4 text-gray-400 font-semibold">操作</th>
              </tr>
            </thead>
            <tbody>
              {chains.length === 0 ? (
                <tr>
                  <td colSpan={6} className="text-center p-8 text-gray-400">
                    暂无链配置
                  </td>
                </tr>
              ) : (
                chains.map((chain) => (
                  <tr key={chain.id} className="border-t border-gray-800 hover:bg-[#151a35] transition">
                    <td className="p-4">
                      <div className="font-medium">{chain.chain_name}</div>
                    </td>
                    <td className="p-4 text-sm text-gray-400">
                      {chain.chain_id}
                    </td>
                    <td className="p-4 text-sm text-gray-400 font-mono">
                      {chain.usdt_contract_address.slice(0, 6)}...{chain.usdt_contract_address.slice(-4)}
                    </td>
                    <td className="p-4 text-sm text-gray-400 font-mono">
                      {chain.platform_deposit_address.slice(0, 6)}...{chain.platform_deposit_address.slice(-4)}
                    </td>
                    <td className="p-4">
                      <span className={`px-2 py-1 text-xs rounded ${
                        chain.enabled 
                          ? 'bg-green-500/20 text-green-400' 
                          : 'bg-red-500/20 text-red-400'
                      }`}>
                        {chain.enabled ? '已启用' : '已禁用'}
                      </span>
                    </td>
                    <td className="p-4 text-sm space-x-2">
                      <button
                        onClick={() => handleEdit(chain)}
                        className="text-primary hover:text-primary-dark"
                      >
                        编辑
                      </button>
                      <button
                        onClick={() => handleToggleStatus(chain)}
                        className={chain.enabled ? 'text-red-400 hover:text-red-300' : 'text-green-400 hover:text-green-300'}
                      >
                        {chain.enabled ? '禁用' : '启用'}
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* 编辑/创建模态框 */}
      {showEditModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-5 w-full max-w-xl max-h-[90vh] overflow-y-auto border border-gray-800">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-bold text-white">
                {editingChain ? '编辑链配置' : '创建链配置'}
              </h2>
              <button
                onClick={() => setShowEditModal(false)}
                className="text-gray-400 hover:text-white text-2xl leading-none"
              >
                ×
              </button>
            </div>

            <div className="space-y-3">
              {/* 编辑时显示链信息，但不可修改 */}
              {editingChain && (
                <div className="p-3 bg-[#151a35] border border-gray-700 rounded">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <div className="text-xs text-gray-400 mb-1">链名称</div>
                      <div className="text-sm font-semibold">{editingChain.chain_name}</div>
                    </div>
                    <div>
                      <div className="text-xs text-gray-400 mb-1">Chain ID</div>
                      <div className="text-sm font-semibold">{editingChain.chain_id}</div>
                    </div>
                  </div>
                </div>
              )}

              {/* 创建时显示输入框 */}
              {!editingChain && (
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-400 mb-1.5">
                      链名称 *
                    </label>
                    <input
                      type="text"
                      value={formData.chain_name || ''}
                      onChange={(e) => setFormData({...formData, chain_name: e.target.value})}
                      className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white focus:ring-1 focus:ring-primary focus:border-transparent"
                      placeholder="BSC Mainnet"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-medium text-gray-400 mb-1.5">
                      Chain ID *
                    </label>
                    <input
                      type="number"
                      value={formData.chain_id || ''}
                      onChange={(e) => setFormData({...formData, chain_id: parseInt(e.target.value)})}
                      className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white focus:ring-1 focus:ring-primary focus:border-transparent"
                      placeholder="56"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">
                  RPC地址 *
                </label>
                <input
                  type="text"
                  value={formData.rpc_url || ''}
                  onChange={(e) => setFormData({...formData, rpc_url: e.target.value})}
                  className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white focus:ring-1 focus:ring-primary focus:border-transparent"
                  placeholder="https://..."
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">
                  区块浏览器
                </label>
                <input
                  type="text"
                  value={formData.block_explorer_url || ''}
                  onChange={(e) => setFormData({...formData, block_explorer_url: e.target.value})}
                  className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white focus:ring-1 focus:ring-primary focus:border-transparent"
                  placeholder="https://bscscan.com"
                />
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div className="col-span-2">
                  <label className="block text-xs font-medium text-gray-400 mb-1.5">
                    USDT合约 *
                  </label>
                  <input
                    type="text"
                    value={formData.usdt_contract_address || ''}
                    onChange={(e) => setFormData({...formData, usdt_contract_address: e.target.value})}
                    className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white font-mono focus:ring-1 focus:ring-primary focus:border-transparent"
                    placeholder="0x..."
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-gray-400 mb-1.5">
                    精度 *
                  </label>
                  <select
                    value={formData.usdt_decimals || 18}
                    onChange={(e) => setFormData({...formData, usdt_decimals: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white focus:ring-1 focus:ring-primary focus:border-transparent"
                  >
                    <option value={6}>6位</option>
                    <option value={18}>18位</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">
                  收款地址 *
                </label>
                <input
                  type="text"
                  value={formData.platform_deposit_address || ''}
                  onChange={(e) => setFormData({...formData, platform_deposit_address: e.target.value})}
                  className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white font-mono focus:ring-1 focus:ring-primary focus:border-transparent"
                  placeholder="0x..."
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">
                  提现私钥 <span className="text-red-400">(敏感)</span>
                </label>
                <input
                  type="password"
                  value={formData.platform_withdraw_private_key || ''}
                  onChange={(e) => setFormData({...formData, platform_withdraw_private_key: e.target.value})}
                  className="w-full px-3 py-2 text-sm bg-[#151a35] border border-gray-700 rounded text-white font-mono focus:ring-1 focus:ring-primary focus:border-transparent"
                  placeholder={editingChain ? "留空不修改" : "0x..."}
                />
                <p className="text-xs text-gray-500 mt-1">
                  {editingChain ? "留空则保持原私钥" : "私钥将存储在数据库"}
                </p>
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-5 pt-4 border-t border-gray-800">
              <button
                onClick={() => setShowEditModal(false)}
                className="px-4 py-1.5 text-sm border border-gray-700 rounded hover:bg-[#151a35] transition"
              >
                取消
              </button>
              <button
                onClick={handleSave}
                className="px-4 py-1.5 text-sm bg-primary text-white rounded hover:bg-primary-dark transition"
              >
                保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

