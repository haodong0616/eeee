'use client';

import { useEffect, useState } from 'react';
import { adminApi, User } from '@/lib/api/admin';

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const data = await adminApi.getUsers();
      setUsers(data);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredUsers = users.filter((user) =>
    user.wallet_address.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">用户管理</h1>
        <input
          type="text"
          placeholder="搜索钱包地址..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="px-4 py-2 bg-[#0f1429] border border-gray-700 rounded-lg"
        />
      </div>

      {loading ? (
        <div>加载中...</div>
      ) : (
        <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
          <table className="w-full">
            <thead className="bg-[#151a35]">
              <tr>
                <th className="text-left p-4">ID</th>
                <th className="text-left p-4">钱包地址</th>
                <th className="text-left p-4">注册时间</th>
                <th className="text-left p-4">最后更新</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.length === 0 ? (
                <tr>
                  <td colSpan={4} className="text-center p-8 text-gray-400">
                    暂无数据
                  </td>
                </tr>
              ) : (
                filteredUsers.map((user) => (
                  <tr key={user.id} className="border-t border-gray-800 hover:bg-[#151a35]">
                    <td className="p-4">{user.id}</td>
                    <td className="p-4 font-mono">{user.wallet_address}</td>
                    <td className="p-4">{new Date(user.created_at).toLocaleString()}</td>
                    <td className="p-4">{new Date(user.updated_at).toLocaleString()}</td>
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

