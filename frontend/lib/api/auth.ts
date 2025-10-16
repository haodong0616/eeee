import apiClient from './client';

export interface NonceResponse {
  nonce: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: number;
    wallet_address: string;
    created_at: string;
  };
}

export interface User {
  id: number;
  wallet_address: string;
  created_at: string;
  updated_at: string;
}

export const authApi = {
  getNonce: async (walletAddress: string): Promise<NonceResponse> => {
    const response = await apiClient.post('/auth/nonce', {
      wallet_address: walletAddress,
    });
    return response.data;
  },

  login: async (walletAddress: string, signature: string): Promise<LoginResponse> => {
    const response = await apiClient.post('/auth/login', {
      wallet_address: walletAddress,
      signature: signature,
    });
    return response.data;
  },

  getProfile: async (): Promise<User> => {
    const response = await apiClient.get('/profile');
    return response.data;
  },
};

