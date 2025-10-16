import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

// API 配置 - 使用 Next.js API Routes 代理
const API_URL = '';

// 定义所有接口的类型
export interface User {
  id: string;
  wallet_address: string;
  user_level: string;
  created_at: string;
  updated_at: string;
}

export interface TradingPair {
  id: string;
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
  id: string;
  symbol: string;
  buy_order_id: string;
  sell_order_id: string;
  price: string;
  quantity: string;
  created_at: string;
}

export interface Kline {
  id: string;
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

export interface Order {
  id: string;
  user_id: string;
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

export interface Balance {
  id: string;
  user_id: string;
  asset: string;
  available: string;
  frozen: string;
  created_at: string;
  updated_at: string;
}

export interface ChainConfig {
  id: string;
  chain_name: string;
  chain_id: number;
  rpc_url: string;
  block_explorer_url: string;
  usdt_contract_address: string;
  usdt_decimals: number;
  platform_deposit_address: string;
  platform_withdraw_private_key?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 创建 API
export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: `${API_URL}/api`,
    prepareHeaders: (headers) => {
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['TradingPairs', 'Tickers', 'Orders', 'Balances', 'Trades', 'OrderBook', 'Klines'],
  endpoints: (builder) => ({
    // ========== 认证接口 ==========
    getNonce: builder.mutation<{ nonce: string }, string>({
      query: (walletAddress) => ({
        url: '/auth/nonce',
        method: 'POST',
        body: { wallet_address: walletAddress },
      }),
    }),
    login: builder.mutation<{ token: string; user: User }, { walletAddress: string; signature: string }>({
      query: ({ walletAddress, signature }) => ({
        url: '/auth/login',
        method: 'POST',
        body: {
          wallet_address: walletAddress,
          signature: signature,
        },
      }),
    }),
    getProfile: builder.query<User, void>({
      query: () => '/profile',
    }),

    // ========== 链配置接口 ==========
    getChains: builder.query<ChainConfig[], void>({
      query: () => '/chains',
    }),

    // ========== 市场数据接口 ==========
    getTradingPairs: builder.query<TradingPair[], void>({
      query: () => '/market/pairs',
      providesTags: ['TradingPairs'],
    }),
    getTicker: builder.query<Ticker, string>({
      query: (symbol) => `/market/ticker/${symbol.replace('/', '-')}`,
      providesTags: (result, error, symbol) => [{ type: 'Tickers', id: symbol }],
    }),
    getAllTickers: builder.query<Ticker[], void>({
      query: () => '/market/tickers',
      providesTags: ['Tickers'],
    }),
    getOrderBook: builder.query<OrderBook, string>({
      query: (symbol) => `/market/orderbook/${symbol.replace('/', '-')}`,
      providesTags: (result, error, symbol) => [{ type: 'OrderBook', id: symbol }],
    }),
    getRecentTrades: builder.query<Trade[], string>({
      query: (symbol) => `/market/trades/${symbol.replace('/', '-')}`,
      providesTags: (result, error, symbol) => [{ type: 'Trades', id: symbol }],
    }),
    getKlines: builder.query<Kline[], { symbol: string; interval: string }>({
      query: ({ symbol, interval }) => ({
        url: `/market/klines/${symbol.replace('/', '-')}`,
        params: { interval },
      }),
      providesTags: (result, error, { symbol, interval }) => [
        { type: 'Klines', id: `${symbol}-${interval}` },
      ],
    }),

    // ========== 订单接口 ==========
    createOrder: builder.mutation<Order, {
      symbol: string;
      order_type: 'limit' | 'market';
      side: 'buy' | 'sell';
      price?: string;
      quantity: string;
    }>({
      query: (data) => ({
        url: '/orders',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Orders', 'Balances'],
    }),
    getOrders: builder.query<Order[], { symbol?: string; status?: string }>({
      query: ({ symbol, status }) => ({
        url: '/orders',
        params: { symbol, status },
      }),
      providesTags: ['Orders'],
    }),
    getOrder: builder.query<Order, string>({
      query: (orderId) => `/orders/${orderId}`,
      providesTags: (result, error, orderId) => [{ type: 'Orders', id: orderId }],
    }),
    cancelOrder: builder.mutation<Order, string>({
      query: (orderId) => ({
        url: `/orders/${orderId}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Orders', 'Balances'],
    }),

    // ========== 余额接口 ==========
    getBalances: builder.query<Balance[], void>({
      query: () => '/balances',
      providesTags: ['Balances'],
    }),
    getBalance: builder.query<Balance, string>({
      query: (asset) => `/balances/${asset}`,
      providesTags: (result, error, asset) => [{ type: 'Balances', id: asset }],
    }),
    deposit: builder.mutation<Balance, { asset: string; amount: string; txHash: string; chain?: string; chainId?: number }>({
      query: (data) => ({
        url: '/balances/deposit',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Balances'],
    }),
    withdraw: builder.mutation<any, { asset: string; amount: string; address: string; chain?: string; chainId?: number }>({
      query: (data) => ({
        url: '/balances/withdraw',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Balances'],
    }),
    
    // 充值记录
    getDepositRecords: builder.query<any[], void>({
      query: () => '/balances/deposits',
    }),
    
    // 提现记录
    getWithdrawRecords: builder.query<any[], void>({
      query: () => '/balances/withdraws',
    }),
  }),
});

// 导出 hooks
export const {
  // 认证
  useGetNonceMutation,
  useLoginMutation,
  useGetProfileQuery,
  
  // 链配置
  useGetChainsQuery,
  
  // 市场数据
  useGetTradingPairsQuery,
  useGetTickerQuery,
  useGetAllTickersQuery,
  useGetOrderBookQuery,
  useGetRecentTradesQuery,
  useGetKlinesQuery,
  
  // 订单
  useCreateOrderMutation,
  useGetOrdersQuery,
  useGetOrderQuery,
  useCancelOrderMutation,
  
  // 余额
  useGetBalancesQuery,
  useGetBalanceQuery,
  useDepositMutation,
  useWithdrawMutation,
  useGetDepositRecordsQuery,
  useGetWithdrawRecordsQuery,
} = api;

