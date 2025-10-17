import type { Metadata, Viewport } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Providers from "./providers";
import { RainbowProvider } from "@/providers/RainbowProvider";
import Header from "@/components/Header";
import Footer from "@/components/Footer";
import { Toaster } from 'react-hot-toast';

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Velocity Exchange - 速度交易所",
  description: "高性能加密货币交易所 - 安全、快速、专业",
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh">
      <body className={inter.className}>
        <RainbowProvider>
          <Providers>
            <div className="flex flex-col min-h-screen bg-[#0a0e27] text-white">
              <Header />
              <main className="flex-1">{children}</main>
              <Footer />
            </div>
            <Toaster
              position="top-center"
              toastOptions={{
                duration: 3000,
                style: {
                  background: '#1a1f35',
                  color: '#fff',
                  border: '1px solid #374151',
                  borderRadius: '8px',
                  padding: '12px 20px',
                },
                success: {
                  iconTheme: {
                    primary: '#10b981',
                    secondary: '#fff',
                  },
                },
                error: {
                  iconTheme: {
                    primary: '#ef4444',
                    secondary: '#fff',
                  },
                },
              }}
            />
          </Providers>
        </RainbowProvider>
      </body>
    </html>
  );
}

