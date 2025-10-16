'use client';

import { useState, useEffect } from 'react';
import { useGetBalancesQuery } from '@/lib/services/api';
import { getPriceStep, getQuantityStep, formatPrice } from '@/lib/utils/format';

interface OrderFormProps {
  symbol: string;
  currentPrice?: string;
  onSubmit: (data: any) => void;
  isAuthenticated: boolean;
  initialPrice?: string;
}

export default function OrderForm({ symbol, currentPrice, onSubmit, isAuthenticated, initialPrice }: OrderFormProps) {
  const [side, setSide] = useState<'buy' | 'sell'>('buy');
  const [orderType, setOrderType] = useState<'limit' | 'market'>('limit');
  const [price, setPrice] = useState('');
  const [quantity, setQuantity] = useState('');

  // 获取用户余额
  const { data: balances = [] } = useGetBalancesQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: 5000,
  });

  // 当外部传入价格时，自动填充
  useEffect(() => {
    if (initialPrice && orderType === 'limit') {
      const priceNum = parseFloat(initialPrice);
      // 根据价格大小决定小数位
      let decimals = 2;
      if (priceNum >= 1000) decimals = 2;
      else if (priceNum >= 100) decimals = 2;
      else if (priceNum >= 1) decimals = 3;
      else if (priceNum >= 0.01) decimals = 4;
      else if (priceNum >= 0.0001) decimals = 6;
      else decimals = 8;
      
      setPrice(priceNum.toFixed(decimals));
    }
  }, [initialPrice, orderType]);

  // 获取交易对的基础资产和报价资产
  const [baseAsset, quoteAsset] = symbol.split('/');

  // 根据买卖方向获取可用余额
  const getAvailableBalance = () => {
    if (!isAuthenticated || balances.length === 0) {
      return '-';
    }

    const asset = side === 'buy' ? quoteAsset : baseAsset;
    const balance = balances.find((b) => b.asset === asset);
    
    if (!balance) {
      return '0';
    }

    return parseFloat(balance.available).toFixed(8);
  };

  const availableAsset = side === 'buy' ? quoteAsset : baseAsset;
  const availableBalance = getAvailableBalance();

  // 点击余额，填充最大可用数量
  const handleMaxClick = () => {
    if (!isAuthenticated || availableBalance === '-' || availableBalance === '0') {
      return;
    }

    const available = parseFloat(availableBalance);
    
    if (side === 'buy') {
      // 买入：根据可用USDT和价格计算最大可买数量
      const currentPriceValue = parseFloat(price || currentPrice || '0');
      if (currentPriceValue > 0) {
        const maxQty = available / currentPriceValue;
        setQuantity(maxQty.toFixed(8));
      }
    } else {
      // 卖出：直接使用可用BTC数量
      setQuantity(available.toFixed(8));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!isAuthenticated) {
      alert('请先连接钱包');
      return;
    }

    if (orderType === 'limit' && !price) {
      alert('请输入价格');
      return;
    }

    if (!quantity) {
      alert('请输入数量');
      return;
    }
    
    const priceValue = orderType === 'limit' ? parseFloat(price) : parseFloat(currentPrice || '0');
    const quantityValue = parseFloat(quantity);
    
    // 验证价格和数量为正数
    if (priceValue <= 0) {
      alert('价格必须大于0');
      return;
    }
    
    if (quantityValue <= 0) {
      alert('数量必须大于0');
      return;
    }

    // 检查余额是否足够
    const asset = side === 'buy' ? quoteAsset : baseAsset;
    const balance = balances.find((b) => b.asset === asset);
    
    if (balance) {
      const available = parseFloat(balance.available);
      const required = side === 'buy' 
        ? parseFloat(price || currentPrice || '0') * parseFloat(quantity)
        : parseFloat(quantity);
      
      if (available < required) {
        alert(`余额不足，可用 ${available.toFixed(8)} ${asset}`);
        return;
      }
    }

    onSubmit({
      order_type: orderType,
      side,
      price: orderType === 'limit' ? price : currentPrice || '0',
      quantity,
    });

    // 重置表单
    setQuantity('');
  };

  return (
    <div className="bg-[#0f1429] rounded-lg border border-gray-800 overflow-hidden">
      <div className="flex border-b border-gray-800">
        <button
          className={`flex-1 py-2 text-sm font-medium transition-colors ${
            side === 'buy' ? 'text-buy border-b-2 border-buy bg-buy/5' : 'text-gray-400 hover:bg-[#151a35]'
          }`}
          onClick={() => setSide('buy')}
        >
          买入
        </button>
        <button
          className={`flex-1 py-2 text-sm font-medium transition-colors ${
            side === 'sell' ? 'text-sell border-b-2 border-sell bg-sell/5' : 'text-gray-400 hover:bg-[#151a35]'
          }`}
          onClick={() => setSide('sell')}
        >
          卖出
        </button>
      </div>

      <div className="p-3">
        <div className="flex mb-3 space-x-2">
          <button
            className={`flex-1 py-1.5 rounded text-xs transition-colors ${
              orderType === 'limit' ? 'bg-primary text-white' : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
            onClick={() => setOrderType('limit')}
          >
            限价
          </button>
          <button
            className={`flex-1 py-1.5 rounded text-xs transition-colors ${
              orderType === 'market' ? 'bg-primary text-white' : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
            onClick={() => setOrderType('market')}
          >
            市价
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          {orderType === 'limit' && (
            <div>
              <label className="block text-[11px] text-gray-400 mb-1">价格</label>
              <input
                type="number"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="w-full px-3 py-1.5 text-sm bg-[#151a35] border border-gray-700 rounded-lg focus:border-primary transition-colors font-mono"
                placeholder={currentPrice ? formatPrice(currentPrice) : '0.00'}
                step={getPriceStep(parseFloat(currentPrice || '0'))}
                min="0"
              />
            </div>
          )}

          <div>
            <label className="block text-[11px] text-gray-400 mb-1">数量</label>
            <input
              type="number"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              className="w-full px-3 py-1.5 text-sm bg-[#151a35] border border-gray-700 rounded-lg focus:border-primary transition-colors font-mono"
              placeholder="0.00"
              step={getQuantityStep(symbol)}
              min="0"
            />
          </div>

          <div className="flex items-center justify-between text-[11px] text-gray-400 py-1">
            <span>可用 {availableAsset}</span>
            <button
              type="button"
              onClick={handleMaxClick}
              className="text-white font-mono hover:text-primary transition-colors cursor-pointer underline decoration-dotted"
              title="点击填入最大可用数量"
            >
              {availableBalance}
            </button>
          </div>

          {orderType === 'limit' && price && quantity && (
            <div className="flex items-center justify-between text-[11px] text-gray-400 py-1">
              <span>预计{side === 'buy' ? '花费' : '获得'}</span>
              <span className="text-white font-mono">
                {side === 'buy' 
                  ? `${(parseFloat(price) * parseFloat(quantity)).toFixed(2)} ${quoteAsset}`
                  : `${(parseFloat(price) * parseFloat(quantity)).toFixed(2)} ${quoteAsset}`
                }
              </span>
            </div>
          )}

          <button
            type="submit"
            disabled={!isAuthenticated}
            className={`w-full py-2.5 rounded-lg text-sm font-semibold transition-all ${
              side === 'buy'
                ? 'bg-buy hover:bg-green-600 hover:shadow-lg hover:shadow-buy/20'
                : 'bg-sell hover:bg-red-600 hover:shadow-lg hover:shadow-sell/20'
            } disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {!isAuthenticated ? '请先连接钱包' : side === 'buy' ? '买入' : '卖出'}
          </button>
        </form>
      </div>
    </div>
  );
}

