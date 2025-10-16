import axios from './client';

// ==================== 类型定义 ====================

// 用户接口
export interface User {
  id: string;
  wallet_address: string;
  created_at: string;
  updated_at: string;
}

// 订单接口
export interface Order {
  id: string;
  user_id: string;
  user?: User;
  symbol: string;
  side: string;
  order_type: string;
  price: string;
  quantity: string;
  filled_qty: string;
  status: string;
  created_at: string;
  updated_at: string;
}

// 交易接口
export interface Trade {
  id: string;
  symbol: string;
  price: string;
  quantity: string;
  buyer_id: string;
  seller_id: string;
  created_at: string;
}

// 交易对接口
export interface TradingPair {
  id: string;
  symbol: string;
  base_asset: string;
  quote_asset: string;
  enabled: boolean;
  price_decimals: number;
  quantity_decimals: number;
  min_price: string;
  max_price: string;
  min_quantity: string;
  max_quantity: string;
  status: string;
  simulator_enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 统计数据接口
export interface Stats {
  user_count: number;
  order_count: number;
  trade_count: number;
  total_volume: string;
}

// 链配置接口
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

// 任务接口
export interface Task {
  ID: string;
  Type: string;
  Status: string; // pending, running, completed, failed
  Symbol?: string;
  RecordID?: string;
  RecordType?: string; // deposit, withdraw
  StartTime?: string;
  EndTime?: string;
  Message: string;
  CreatedAt: string;
  StartedAt?: string;
  EndedAt?: string;
  Error?: string;
}

// 任务日志接口
export interface TaskLog {
  id: string;
  task_id: string;
  level: string; // info, warning, error
  stage: string;
  message: string;
  details?: string;
  created_at: string;
}

// ==================== 用户管理 ====================

export const getUsers = async () => {
  const response = await axios.get<User[]>('/admin/users');
  return response.data;
};

// ==================== 订单管理 ====================

export const getOrders = async () => {
  const response = await axios.get<Order[]>('/admin/orders');
  return response.data;
};

// ==================== 交易管理 ====================

export const getTrades = async () => {
  const response = await axios.get<Trade[]>('/admin/trades');
  return response.data;
};

// ==================== 充值提现管理 ====================

export const getDeposits = async () => {
  const response = await axios.get('/admin/deposits');
  return response.data;
};

export const getWithdrawals = async () => {
  const response = await axios.get('/admin/withdrawals');
  return response.data;
};

// ==================== 统计数据 ====================

export const getStats = async () => {
  const response = await axios.get<Stats>('/admin/stats');
  return response.data;
};

// ==================== 交易对管理 ====================

export const getTradingPairs = async () => {
  const response = await axios.get<TradingPair[]>('/admin/pairs');
  return response.data;
};

export const createTradingPair = async (data: Partial<TradingPair>) => {
  const response = await axios.post<TradingPair>('/admin/pairs', data);
  return response.data;
};

export const updateTradingPair = async (id: string, data: Partial<TradingPair>) => {
  const response = await axios.put<TradingPair>(`/admin/pairs/${id}`, data);
  return response.data;
};

export const updateTradingPairStatus = async (id: string, status: string) => {
  const response = await axios.put(`/admin/pairs/${id}/status`, { status });
  return response.data;
};

export const updateTradingPairSimulator = async (id: string, enabled: boolean) => {
  const response = await axios.put(`/admin/pairs/${id}/simulator`, { simulator_enabled: enabled });
  return response.data;
};

export const generateTradeDataForPair = async (symbol: string, startTime: string, endTime: string) => {
  const response = await axios.post('/admin/pairs/generate-trades', { symbol, start_time: startTime, end_time: endTime });
  return response.data;
};

export const generateKlineDataForPair = async (symbol: string) => {
  const response = await axios.post('/admin/pairs/generate-klines', { symbol });
  return response.data;
};

// ==================== 链配置管理 ====================

export const getChains = async () => {
  const response = await axios.get<ChainConfig[]>('/admin/chains');
  return response.data;
};

export const getChain = async (id: string) => {
  const response = await axios.get<ChainConfig>(`/admin/chains/${id}`);
  return response.data;
};

export const createChain = async (data: Partial<ChainConfig>) => {
  const response = await axios.post<ChainConfig>('/admin/chains', data);
  return response.data;
};

export const updateChain = async (id: string, data: Partial<ChainConfig>) => {
  const response = await axios.put<ChainConfig>(`/admin/chains/${id}`, data);
  return response.data;
};

export const updateChainStatus = async (id: string, enabled: boolean) => {
  const response = await axios.put(`/admin/chains/${id}/status`, { enabled });
  return response.data;
};

export const deleteChain = async (id: string) => {
  const response = await axios.delete(`/admin/chains/${id}`);
  return response.data;
};

// ==================== 任务管理 ====================

export const getAllTasks = async () => {
  const response = await axios.get<Task[]>('/admin/tasks');
  return response.data;
};

export const getTaskStatus = async (id: string) => {
  const response = await axios.get<Task>(`/admin/tasks/${id}`);
  return response.data;
};

export const getRunningTask = async () => {
  const response = await axios.get('/admin/tasks/running');
  return response.data;
};

export const getTaskLogs = async (taskId: string) => {
  const response = await axios.get<TaskLog[]>(`/admin/tasks/${taskId}/logs`);
  return response.data;
};

export const retryTask = async (taskId: string) => {
  const response = await axios.post(`/admin/tasks/${taskId}/retry`);
  return response.data;
};

// ==================== 统一导出 adminApi 对象 ====================

export const adminApi = {
  // 用户管理
  getUsers,
  
  // 订单管理
  getOrders,
  
  // 交易记录
  getTrades,
  
  // 充提记录
  getDeposits,
  getWithdrawals,
  
  // 统计数据
  getStats,
  
  // 交易对管理
  getTradingPairs,
  createTradingPair,
  updateTradingPair,
  updateTradingPairStatus,
  updateTradingPairSimulator,
  generateTradeDataForPair,
  generateKlineDataForPair,
  
  // 链配置管理
  getChains,
  getChain,
  createChain,
  updateChain,
  updateChainStatus,
  deleteChain,
  
  // 任务管理
  getAllTasks,
  getTaskStatus,
  getRunningTask,
  getTaskLogs,
  retryTask,
};
