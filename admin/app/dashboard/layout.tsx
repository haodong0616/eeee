'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    if (!token) {
      router.push('/login');
    }
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    router.push('/login');
  };

  const navItems = [
    { href: '/dashboard', label: '概览', icon: '📊' },
    { href: '/dashboard/users', label: '用户管理', icon: '👥' },
    { href: '/dashboard/orders', label: '订单管理', icon: '📋' },
    { href: '/dashboard/trades', label: '成交记录', icon: '💱' },
    { href: '/dashboard/pairs', label: '交易对管理', icon: '⚙️' },
  ];

  return (
    <div className="flex min-h-screen">
      {/* 侧边栏 */}
      <aside className="w-64 bg-[#0f1429] border-r border-gray-800">
        <div className="p-6">
          <h1 className="text-2xl font-bold text-primary">ExpChange</h1>
          <p className="text-sm text-gray-400">管理后台</p>
        </div>

        <nav className="px-4 space-y-2">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center space-x-3 px-4 py-3 rounded-lg transition ${
                pathname === item.href
                  ? 'bg-primary text-white'
                  : 'text-gray-400 hover:bg-[#151a35]'
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        <div className="absolute bottom-4 left-4 right-4">
          <button
            onClick={handleLogout}
            className="w-full px-4 py-3 bg-red-600 hover:bg-red-700 rounded-lg transition"
          >
            退出登录
          </button>
        </div>
      </aside>

      {/* 主内容区 */}
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}

