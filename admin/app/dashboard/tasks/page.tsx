'use client';

import { useState } from 'react';
import useSWR from 'swr';
import { adminApi, Task, TaskLog } from '@/lib/api/admin';
import toast from 'react-hot-toast';

export default function TasksPage() {
  const { data: tasks = [], isLoading, mutate } = useSWR('/admin/tasks', () => adminApi.getAllTasks(), {
    refreshInterval: 3000, // 每3秒自动刷新
  });

  const [filter, setFilter] = useState('all');
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [taskLogs, setTaskLogs] = useState<TaskLog[]>([]);

  const filteredTasks = filter === 'all' 
    ? tasks 
    : tasks.filter((task) => task.Status === filter);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-500/20 text-green-400';
      case 'running':
        return 'bg-blue-500/20 text-blue-400';
      case 'failed':
        return 'bg-red-500/20 text-red-400';
      case 'pending':
        return 'bg-yellow-500/20 text-yellow-400';
      default:
        return 'bg-gray-500/20 text-gray-400';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed':
        return '✅ 已完成';
      case 'running':
        return '⚙️ 运行中';
      case 'failed':
        return '❌ 失败';
      case 'pending':
        return '⏳ 等待中';
      default:
        return status;
    }
  };

  const getTaskTypeText = (type: string) => {
    switch (type) {
      case 'generate_trades':
        return '📊 生成交易数据';
      case 'generate_klines':
        return '📈 生成K线数据';
      case 'verify_deposit':
        return '💰 验证充值';
      case 'process_withdraw':
        return '💸 处理提现';
      default:
        return type;
    }
  };

  const formatDuration = (startTime?: string, endTime?: string) => {
    if (!startTime || !endTime) return '-';
    const start = new Date(startTime).getTime();
    const end = new Date(endTime).getTime();
    const duration = Math.floor((end - start) / 1000);
    
    if (duration < 60) return `${duration}秒`;
    if (duration < 3600) return `${Math.floor(duration / 60)}分${duration % 60}秒`;
    return `${Math.floor(duration / 3600)}小时${Math.floor((duration % 3600) / 60)}分`;
  };

  const handleViewLogs = async (task: Task) => {
    setSelectedTask(task);
    setShowLogsModal(true);
    try {
      const logs = await adminApi.getTaskLogs(task.ID);
      setTaskLogs(logs);
    } catch (error: any) {
      toast.error('加载日志失败: ' + (error?.response?.data?.error || '未知错误'));
    }
  };

  const handleRetryTask = async (taskId: string) => {
    try {
      await toast.promise(
        adminApi.retryTask(taskId),
        {
          loading: '正在重试任务...',
          success: '任务已重新加入队列',
          error: (err) => err?.response?.data?.error || '重试失败',
        }
      );
      mutate(); // 刷新任务列表
    } catch (error) {
      // toast.promise 已经处理了错误
    }
  };

  const getLevelBadge = (level: string) => {
    switch (level) {
      case 'error':
        return 'bg-red-500/20 text-red-400';
      case 'warning':
        return 'bg-yellow-500/20 text-yellow-400';
      default:
        return 'bg-blue-500/20 text-blue-400';
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">队列任务</h1>
        <div className="flex gap-2">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-lg transition ${
              filter === 'all' ? 'bg-primary' : 'bg-gray-700 hover:bg-gray-600'
            }`}
          >
            全部 ({tasks.length})
          </button>
          <button
            onClick={() => setFilter('pending')}
            className={`px-4 py-2 rounded-lg transition ${
              filter === 'pending' ? 'bg-primary' : 'bg-gray-700 hover:bg-gray-600'
            }`}
          >
            等待中 ({tasks.filter(t => t.Status === 'pending').length})
          </button>
          <button
            onClick={() => setFilter('running')}
            className={`px-4 py-2 rounded-lg transition ${
              filter === 'running' ? 'bg-primary' : 'bg-gray-700 hover:bg-gray-600'
            }`}
          >
            运行中 ({tasks.filter(t => t.Status === 'running').length})
          </button>
          <button
            onClick={() => setFilter('completed')}
            className={`px-4 py-2 rounded-lg transition ${
              filter === 'completed' ? 'bg-primary' : 'bg-gray-700 hover:bg-gray-600'
            }`}
          >
            已完成 ({tasks.filter(t => t.Status === 'completed').length})
          </button>
          <button
            onClick={() => setFilter('failed')}
            className={`px-4 py-2 rounded-lg transition ${
              filter === 'failed' ? 'bg-primary' : 'bg-gray-700 hover:bg-gray-600'
            }`}
          >
            失败 ({tasks.filter(t => t.Status === 'failed').length})
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-400">加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-[#151a35]">
                <tr>
                  <th className="text-left p-4">任务ID</th>
                  <th className="text-left p-4">类型</th>
                  <th className="text-left p-4">关联对象</th>
                  <th className="text-left p-4">详情</th>
                  <th className="text-left p-4">状态</th>
                  <th className="text-left p-4">消息</th>
                  <th className="text-left p-4">耗时</th>
                  <th className="text-left p-4">创建时间</th>
                  <th className="text-right p-4">操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredTasks.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="text-center p-8 text-gray-400">
                      暂无任务
                    </td>
                  </tr>
                ) : (
                  filteredTasks.map((task) => (
                    <tr key={task.ID} className="border-t border-gray-800 hover:bg-[#151a35] transition">
                      <td className="p-4 font-mono text-xs">{task.ID}</td>
                      <td className="p-4">{getTaskTypeText(task.Type)}</td>
                      <td className="p-4">
                        {task.Symbol ? (
                          <span className="font-semibold">{task.Symbol}</span>
                        ) : task.RecordID ? (
                          <span className="text-xs text-gray-500 font-mono">
                            {task.RecordID.substring(0, 12)}...
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td className="p-4 text-sm text-gray-400">
                        {task.StartTime && task.EndTime ? (
                          <div className="space-y-1">
                            <div className="text-xs">开始: {new Date(task.StartTime).toLocaleDateString('zh-CN')}</div>
                            <div className="text-xs">结束: {new Date(task.EndTime).toLocaleDateString('zh-CN')}</div>
                          </div>
                        ) : task.RecordType ? (
                          <div className="text-xs">
                            <span className="px-2 py-0.5 rounded bg-gray-700 text-gray-300">
                              {task.RecordType === 'deposit' ? '充值记录' : '提现记录'}
                            </span>
                          </div>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td className="p-4">
                        <span className={`px-3 py-1 rounded text-sm ${getStatusBadge(task.Status)}`}>
                          {getStatusText(task.Status)}
                        </span>
                      </td>
                      <td className="p-4">
                        <div className="max-w-xs">
                          <div className="text-sm">{task.Message}</div>
                          {task.Error && (
                            <div className="text-xs text-red-400 mt-1 truncate" title={task.Error}>
                              错误: {task.Error}
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="p-4 text-sm">
                        {formatDuration(task.StartedAt, task.EndedAt)}
                      </td>
                      <td className="p-4 text-sm text-gray-400">
                        {new Date(task.CreatedAt).toLocaleString('zh-CN')}
                      </td>
                      <td className="p-4">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => handleViewLogs(task)}
                            className="px-3 py-1 bg-blue-600 hover:bg-blue-700 rounded text-xs transition"
                            title="查看日志"
                          >
                            📋 日志
                          </button>
                          {task.Status === 'failed' && (
                            <button
                              onClick={() => handleRetryTask(task.ID)}
                              className="px-3 py-1 bg-green-600 hover:bg-green-700 rounded text-xs transition"
                              title="重试任务"
                            >
                              🔄 重试
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 任务日志模态框 */}
      {showLogsModal && selectedTask && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0f1429] rounded-lg p-6 w-full max-w-4xl max-h-[80vh] border border-gray-800">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-bold">任务执行日志</h2>
                <p className="text-sm text-gray-400 mt-1">
                  任务ID: {selectedTask.ID}
                  {selectedTask.Symbol && ` | 交易对: ${selectedTask.Symbol}`}
                  {selectedTask.RecordID && ` | 记录ID: ${selectedTask.RecordID}`}
                </p>
              </div>
              <button
                onClick={() => setShowLogsModal(false)}
                className="text-gray-400 hover:text-white transition"
              >
                ✕
              </button>
            </div>

            <div className="overflow-y-auto max-h-[60vh]">
              {taskLogs.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                  暂无日志记录
                </div>
              ) : (
                <div className="space-y-2">
                  {taskLogs.map((log) => (
                    <div
                      key={log.id}
                      className="bg-[#151a35] rounded-lg p-3 border border-gray-800"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`px-2 py-0.5 rounded text-xs ${getLevelBadge(log.level)}`}>
                              {log.level.toUpperCase()}
                            </span>
                            <span className="text-xs text-gray-500">{log.stage}</span>
                            <span className="text-xs text-gray-500">
                              {new Date(log.created_at).toLocaleTimeString('zh-CN')}
                            </span>
                          </div>
                          <div className="text-sm">{log.message}</div>
                          {log.details && (
                            <div className="text-xs text-gray-400 mt-1">
                              {log.details}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="mt-4 flex justify-end">
              <button
                onClick={() => setShowLogsModal(false)}
                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 统计信息 */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mt-6">
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
          <div className="text-gray-400 text-sm mb-1">总任务数</div>
          <div className="text-2xl font-bold">{tasks.length}</div>
        </div>
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
          <div className="text-gray-400 text-sm mb-1">等待中</div>
          <div className="text-2xl font-bold text-yellow-400">
            {tasks.filter(t => t.Status === 'pending').length}
          </div>
        </div>
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
          <div className="text-gray-400 text-sm mb-1">运行中</div>
          <div className="text-2xl font-bold text-blue-400">
            {tasks.filter(t => t.Status === 'running').length}
          </div>
        </div>
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
          <div className="text-gray-400 text-sm mb-1">已完成</div>
          <div className="text-2xl font-bold text-green-400">
            {tasks.filter(t => t.Status === 'completed').length}
          </div>
        </div>
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-4">
          <div className="text-gray-400 text-sm mb-1">失败</div>
          <div className="text-2xl font-bold text-red-400">
            {tasks.filter(t => t.Status === 'failed').length}
          </div>
        </div>
      </div>
    </div>
  );
}

