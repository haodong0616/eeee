import apiClient from './client';

export interface TradingPair {
  id: number;
  symbol: string;
  base_asset: string;
  quote_asset: string;
  min_price: string;
  max_price: string;
  min_qty: string;
  max_qty: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export const marketApi = {
  getTradingPairs: async (): Promise<TradingPair[]> => {
    const response = await apiClient.get('/market/pairs');
    return response.data;
  },
};

// 辅助函数：将 symbol 转换为 URL 安全格式
export function encodeSymbol(symbol: string): string {
  return symbol.replace('/', '-');
}

// 辅助函数：将 URL 格式转换回 symbol
export function decodeSymbol(encoded: string): string {
  return encoded.replace('-', '/');
}

