'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    // 简化的登录逻辑，实际应该调用API
    // 这里假设admin/admin是正确的凭据
    if (username === 'admin' && password === 'admin') {
      localStorage.setItem('admin_token', 'admin-token-placeholder');
      router.push('/dashboard');
    } else {
      alert('用户名或密码错误');
    }

    setLoading(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="bg-[#0f1429] rounded-lg border border-gray-800 p-8 w-96">
        <h1 className="text-2xl font-bold text-center mb-8">管理后台登录</h1>
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-2">用户名</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
              placeholder="请输入用户名"
              required
            />
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-2">密码</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 bg-[#151a35] border border-gray-700 rounded-lg"
              placeholder="请输入密码"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 bg-primary hover:bg-primary-dark rounded-lg font-semibold transition disabled:opacity-50"
          >
            {loading ? '登录中...' : '登录'}
          </button>
        </form>
        <p className="text-xs text-gray-400 text-center mt-4">
          默认账号: admin / admin
        </p>
      </div>
    </div>
  );
}

