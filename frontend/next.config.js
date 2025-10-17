/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    // 不设置默认的 WS_URL，让客户端自动检测（通过 Next.js 服务器代理）
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL,
  },
  // 禁用静态导出模式，使用动态渲染
  output: undefined,
}

module.exports = nextConfig


