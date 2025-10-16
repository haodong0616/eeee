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

export interface Ticker {
  symbol: string;
  last_price: string;
  change_24h: string;
  high_24h: string;
  low_24h: string;
  volume_24h: string;
  updated_at: string;
}

export interface OrderBookItem {
  price: string;
  quantity: string;
}

export interface OrderBook {
  symbol: string;
  bids: OrderBookItem[];
  asks: OrderBookItem[];
}

export interface Trade {
  id: number;
  symbol: string;
  buy_order_id: number;
  sell_order_id: number;
  price: string;
  quantity: string;
  created_at: string;
}

export interface Kline {
  id: number;
  symbol: string;
  interval: string;
  open_time: number;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  created_at: string;
  updated_at: string;
}

export const marketApi = {
  getTradingPairs: async (): Promise<TradingPair[]> => {
    const response = await apiClient.get('/market/pairs');
    return response.data;
  },

  getTicker: async (symbol: string): Promise<Ticker> => {
    // 将 / 替换为 - 以避免 URL 路径问题
    const encodedSymbol = symbol.replace('/', '-');
    const response = await apiClient.get(`/market/ticker/${encodedSymbol}`);
    return response.data;
  },

  getAllTickers: async (): Promise<Ticker[]> => {
    const response = await apiClient.get('/market/tickers');
    return response.data;
  },

  getOrderBook: async (symbol: string): Promise<OrderBook> => {
    const encodedSymbol = symbol.replace('/', '-');
    const response = await apiClient.get(`/market/orderbook/${encodedSymbol}`);
    return response.data;
  },

  getRecentTrades: async (symbol: string): Promise<Trade[]> => {
    const encodedSymbol = symbol.replace('/', '-');
    const response = await apiClient.get(`/market/trades/${encodedSymbol}`);
    return response.data;
  },

  getKlines: async (symbol: string, interval: string): Promise<Kline[]> => {
    const encodedSymbol = symbol.replace('/', '-');
    const response = await apiClient.get(`/market/klines/${encodedSymbol}`, {
      params: { interval },
    });
    return response.data;
  },
};

