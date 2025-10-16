import apiClient from './client';

export interface User {
  id: number;
  wallet_address: string;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: number;
  user_id: number;
  symbol: string;
  order_type: string;
  side: string;
  price: string;
  quantity: string;
  filled_qty: string;
  status: string;
  created_at: string;
  updated_at: string;
  user?: User;
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

export interface Stats {
  user_count: number;
  order_count: number;
  trade_count: number;
  total_volume: string;
}

export const adminApi = {
  getUsers: async (): Promise<User[]> => {
    const response = await apiClient.get('/admin/users');
    return response.data;
  },

  getOrders: async (): Promise<Order[]> => {
    const response = await apiClient.get('/admin/orders');
    return response.data;
  },

  getTrades: async (): Promise<Trade[]> => {
    const response = await apiClient.get('/admin/trades');
    return response.data;
  },

  getStats: async (): Promise<Stats> => {
    const response = await apiClient.get('/admin/stats');
    return response.data;
  },

  createTradingPair: async (data: {
    symbol: string;
    base_asset: string;
    quote_asset: string;
    min_price?: string;
    max_price?: string;
    min_qty?: string;
    max_qty?: string;
  }): Promise<TradingPair> => {
    const response = await apiClient.post('/admin/pairs', data);
    return response.data;
  },

  updateTradingPairStatus: async (id: number, status: string): Promise<TradingPair> => {
    const formData = new FormData();
    formData.append('status', status);
    const response = await apiClient.put(`/admin/pairs/${id}/status`, formData);
    return response.data;
  },
};

