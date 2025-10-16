import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Providers from "./providers";
import { RainbowProvider } from "@/providers/RainbowProvider";
import Header from "@/components/Header";
import Footer from "@/components/Footer";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Velocity Exchange - 速度交易所",
  description: "高性能加密货币交易所 - 安全、快速、专业",
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
          </Providers>
        </RainbowProvider>
      </body>
    </html>
  );
}

