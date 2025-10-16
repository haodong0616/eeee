import apiClient from './client';

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
}

export interface CreateOrderRequest {
  symbol: string;
  order_type: 'limit' | 'market';
  side: 'buy' | 'sell';
  price?: string;
  quantity: string;
}

export const orderApi = {
  createOrder: async (data: CreateOrderRequest): Promise<Order> => {
    // symbol 保持原样，后端会在内部处理
    const response = await apiClient.post('/orders', data);
    return response.data;
  },

  getOrders: async (symbol?: string, status?: string): Promise<Order[]> => {
    // symbol 作为查询参数，保持原样
    const response = await apiClient.get('/orders', {
      params: { symbol, status },
    });
    return response.data;
  },

  getOrder: async (orderId: number): Promise<Order> => {
    const response = await apiClient.get(`/orders/${orderId}`);
    return response.data;
  },

  cancelOrder: async (orderId: number): Promise<Order> => {
    const response = await apiClient.delete(`/orders/${orderId}`);
    return response.data;
  },
};

