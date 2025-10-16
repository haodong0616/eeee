'use client';

import Link from 'next/link';

export default function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="bg-[#0f1429] border-t border-gray-800 mt-auto">
      <div className="container mx-auto px-4 py-8 md:py-12">
        {/* Main Footer Content */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 md:gap-8 mb-6 md:mb-8">
          {/* Brand */}
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2 mb-3 md:mb-4">
              <div className="w-8 h-8 md:w-10 md:h-10 bg-gradient-to-br from-primary to-purple-500 rounded-lg flex items-center justify-center">
                <span className="text-lg md:text-xl">⚡</span>
              </div>
              <div>
                <div className="font-bold text-base md:text-lg bg-gradient-to-r from-primary to-purple-400 bg-clip-text text-transparent">
                  Velocity
                </div>
                <div className="text-[10px] md:text-xs text-gray-500">Exchange</div>
              </div>
            </div>
            <p className="text-gray-400 text-xs md:text-sm mb-3 md:mb-4">
              极速交易，流畅体验。专业的数字资产交易平台。
            </p>
            <div className="flex gap-2 md:gap-3">
              <a href="#" className="w-7 h-7 md:w-8 md:h-8 bg-gray-800 hover:bg-primary rounded-lg flex items-center justify-center transition-colors">
                <span className="text-xs md:text-sm">𝕏</span>
              </a>
              <a href="#" className="w-7 h-7 md:w-8 md:h-8 bg-gray-800 hover:bg-primary rounded-lg flex items-center justify-center transition-colors">
                <span className="text-xs md:text-sm">📱</span>
              </a>
              <a href="#" className="w-7 h-7 md:w-8 md:h-8 bg-gray-800 hover:bg-primary rounded-lg flex items-center justify-center transition-colors">
                <span className="text-xs md:text-sm">📧</span>
              </a>
            </div>
          </div>

          {/* Products */}
          <div>
            <h3 className="font-semibold mb-3 md:mb-4 text-xs md:text-sm">产品</h3>
            <ul className="space-y-1.5 md:space-y-2 text-xs md:text-sm">
              <li>
                <Link href="/markets" className="text-gray-400 hover:text-primary transition-colors">
                  现货交易
                </Link>
              </li>
              <li>
                <Link href="/markets" className="text-gray-400 hover:text-primary transition-colors">
                  行情中心
                </Link>
              </li>
              <li>
                <Link href="/assets" className="text-gray-400 hover:text-primary transition-colors">
                  资产管理
                </Link>
              </li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h3 className="font-semibold mb-3 md:mb-4 text-xs md:text-sm">支持</h3>
            <ul className="space-y-1.5 md:space-y-2 text-xs md:text-sm">
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  帮助中心
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  API文档
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  手续费率
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  联系我们
                </a>
              </li>
            </ul>
          </div>

          {/* About */}
          <div>
            <h3 className="font-semibold mb-3 md:mb-4 text-xs md:text-sm">关于</h3>
            <ul className="space-y-1.5 md:space-y-2 text-xs md:text-sm">
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  关于我们
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  服务条款
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  隐私政策
                </a>
              </li>
              <li>
                <a href="#" className="text-gray-400 hover:text-primary transition-colors">
                  风险提示
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* Copyright */}
        <div className="pt-6 md:pt-8 border-t border-gray-800">
          <div className="flex flex-col md:flex-row justify-between items-center gap-2 md:gap-4 text-xs md:text-sm text-gray-400">
            <div className="text-center md:text-left">
              © {currentYear} Velocity Exchange. All rights reserved.
            </div>
            <div className="flex items-center gap-3 md:gap-4 text-[10px] md:text-xs">
              <span className="flex items-center gap-1">
                <span className="w-1.5 h-1.5 md:w-2 md:h-2 bg-green-400 rounded-full animate-pulse"></span>
                系统正常
              </span>
              <span>|</span>
              <span>v1.0.0</span>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}

