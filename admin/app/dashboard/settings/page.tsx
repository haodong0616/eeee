'use client';

import { useState } from 'react';
import useSWR, { mutate } from 'swr';
import { adminApi, SystemConfig } from '@/lib/api/admin';
import toast from 'react-hot-toast';

export default function SettingsPage() {
  const { data: configs = [], isLoading } = useSWR(
    '/admin/configs',
    () => adminApi.getSystemConfigs(),
    {
      refreshInterval: 30000, // 每30秒自动刷新
    }
  );

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const categories = [
    { value: 'all', label: '全部' },
    { value: 'platform', label: '平台配置' },
    { value: 'fee', label: '手续费' },
    { value: 'task', label: '任务队列' },
    { value: 'blockchain', label: '区块链' },
    { value: 'kline', label: 'K线' },
    { value: 'websocket', label: 'WebSocket' },
    { value: 'simulator', label: '模拟器' },
  ];

  const handleEdit = (config: SystemConfig) => {
    setEditingId(config.id);
    setEditValue(config.value);
  };

  const handleSave = async (id: string) => {
    try {
      await adminApi.updateSystemConfig(id, editValue);
      setEditingId(null);
      mutate('/admin/configs');
      toast.success('配置已更新并热更新到系统');
    } catch (error: any) {
      toast.error('更新失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleCancel = () => {
    setEditingId(null);
    setEditValue('');
  };

  const handleReloadAll = async () => {
    if (!confirm('确定要重新加载所有配置吗？')) {
      return;
    }

    try {
      await adminApi.reloadSystemConfigs();
      mutate('/admin/configs');
      toast.success('所有配置已重新加载');
    } catch (error: any) {
      toast.error('重新加载失败：' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const filteredConfigs =
    selectedCategory === 'all'
      ? configs
      : configs.filter((c) => c.category === selectedCategory);

  const getCategoryLabel = (category: string) => {
    const cat = categories.find((c) => c.value === category);
    return cat?.label || category;
  };

  const getValueTypeIcon = (type: string) => {
    const icons = {
      string: '📝',
      number: '🔢',
      boolean: '✅',
    };
    return icons[type as keyof typeof icons] || '📋';
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">系统设置</h1>
        <button
          onClick={handleReloadAll}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition text-sm"
        >
          重新加载所有配置
        </button>
      </div>

      {/* 分类筛选 */}
      <div className="mb-6 flex gap-2 flex-wrap">
        {categories.map((cat) => (
          <button
            key={cat.value}
            onClick={() => setSelectedCategory(cat.value)}
            className={`px-4 py-2 rounded-lg text-sm transition ${
              selectedCategory === cat.value
                ? 'bg-primary text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* 提示信息 */}
      <div className="mb-6 p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
        <p className="text-sm text-green-400">
          💡 <strong>热更新说明：</strong>
          <br />
          修改配置后会立即生效，无需重启服务。部分配置（如任务队列worker数）可能需要新任务生效。
        </p>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-400">加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4">分类</th>
                <th className="text-left p-4">配置项</th>
                <th className="text-left p-4">当前值</th>
                <th className="text-left p-4">说明</th>
                <th className="text-right p-4">操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredConfigs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="text-center p-8 text-gray-400">
                    暂无配置数据
                  </td>
                </tr>
              ) : (
                filteredConfigs.map((config) => (
                  <tr key={config.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                    <td className="p-4">
                      <span className="px-2 py-1 bg-gray-700 rounded text-xs">
                        {getCategoryLabel(config.category)}
                      </span>
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <span>{getValueTypeIcon(config.value_type)}</span>
                        <span className="font-mono text-sm text-gray-300">{config.key}</span>
                      </div>
                    </td>
                    <td className="p-4">
                      {editingId === config.id ? (
                        config.value_type === 'boolean' ? (
                          <select
                            value={editValue}
                            onChange={(e) => setEditValue(e.target.value)}
                            className="w-full px-3 py-1.5 bg-[#151a35] border border-primary rounded text-sm"
                            autoFocus
                          >
                            <option value="true">启用</option>
                            <option value="false">禁用</option>
                          </select>
                        ) : (
                          <input
                            type={config.value_type === 'number' ? 'number' : config.key.includes('key') ? 'password' : 'text'}
                            value={editValue}
                            onChange={(e) => setEditValue(e.target.value)}
                            className="w-full px-3 py-1.5 bg-[#151a35] border border-primary rounded text-sm font-mono"
                            step={config.value_type === 'number' ? '0.0001' : undefined}
                            autoFocus
                          />
                        )
                      ) : (
                        <span className="font-mono text-sm">
                          {config.value_type === 'boolean' ? (
                            <span
                              className={`px-2 py-1 rounded text-xs ${
                                config.value === 'true'
                                  ? 'bg-green-500/20 text-green-500'
                                  : 'bg-red-500/20 text-red-500'
                              }`}
                            >
                              {config.value === 'true' ? '启用' : '禁用'}
                            </span>
                          ) : config.key.includes('key') && config.value ? (
                            <span className="text-gray-500">••••••••</span>
                          ) : (
                            config.value || <span className="text-gray-500">未设置</span>
                          )}
                        </span>
                      )}
                    </td>
                    <td className="p-4 text-sm text-gray-400">{config.description}</td>
                    <td className="p-4 text-right">
                      {editingId === config.id ? (
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleSave(config.id)}
                            className="px-3 py-1 bg-green-600 hover:bg-green-700 rounded text-xs transition"
                          >
                            保存
                          </button>
                          <button
                            onClick={handleCancel}
                            className="px-3 py-1 bg-gray-600 hover:bg-gray-700 rounded text-xs transition"
                          >
                            取消
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => handleEdit(config)}
                          className="px-3 py-1 bg-primary hover:bg-primary-dark rounded text-xs transition"
                        >
                          编辑
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

