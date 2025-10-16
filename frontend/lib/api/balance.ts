import apiClient from './client';

export interface Balance {
  id: number;
  user_id: number;
  asset: string;
  available: string;
  frozen: string;
  created_at: string;
  updated_at: string;
}

export interface DepositRequest {
  asset: string;
  amount: string;
}

export interface WithdrawRequest {
  asset: string;
  amount: string;
  address: string;
}

export const balanceApi = {
  getBalances: async (): Promise<Balance[]> => {
    const response = await apiClient.get('/balances');
    return response.data;
  },

  getBalance: async (asset: string): Promise<Balance> => {
    const response = await apiClient.get(`/balances/${asset}`);
    return response.data;
  },

  deposit: async (data: DepositRequest): Promise<Balance> => {
    const response = await apiClient.post('/balances/deposit', data);
    return response.data;
  },

  withdraw: async (data: WithdrawRequest): Promise<any> => {
    const response = await apiClient.post('/balances/withdraw', data);
    return response.data;
  },
};

