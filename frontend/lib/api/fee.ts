import apiClient from './client';

export interface FeeConfig {
  id: number;
  user_level: string;
  maker_fee_rate: string;
  taker_fee_rate: string;
  created_at: string;
  updated_at: string;
}

export interface FeeRecord {
  id: number;
  user_id: number;
  order_id: number;
  trade_id: number;
  asset: string;
  amount: string;
  fee_rate: string;
  order_side: string;
  created_at: string;
}

export const feeApi = {
  getFeeConfigs: async (): Promise<FeeConfig[]> => {
    const response = await apiClient.get('/fees/configs');
    return response.data;
  },

  getUserFeeStats: async (): Promise<Record<string, string>> => {
    const response = await apiClient.get('/fees/stats');
    return response.data;
  },

  getUserFeeRecords: async (): Promise<FeeRecord[]> => {
    const response = await apiClient.get('/fees/records');
    return response.data;
  },
};

