import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface TradeState {
  lastTradingPair: string; // 最后访问的交易对
}

const initialState: TradeState = {
  lastTradingPair: 'LUNAR/USDT', // 默认交易对
};

const tradeSlice = createSlice({
  name: 'trade',
  initialState,
  reducers: {
    setLastTradingPair: (state, action: PayloadAction<string>) => {
      state.lastTradingPair = action.payload;
    },
  },
});

export const { setLastTradingPair } = tradeSlice.actions;
export default tradeSlice.reducer;

