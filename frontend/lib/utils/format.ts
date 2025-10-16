// 根据价格自动调整小数位精度
export function formatPrice(price: string | number): string {
  const numPrice = typeof price === 'string' ? parseFloat(price) : price;
  
  if (isNaN(numPrice) || numPrice === 0) {
    return '-';
  }

  // 根据价格区间决定小数位
  if (numPrice >= 1000) {
    // >= $1,000: 2位小数 (如 8,500.00)
    return numPrice.toFixed(2);
  } else if (numPrice >= 100) {
    // >= $100: 2位小数 (如 125.50)
    return numPrice.toFixed(2);
  } else if (numPrice >= 1) {
    // >= $1: 3位小数 (如 45.500)
    return numPrice.toFixed(3);
  } else if (numPrice >= 0.01) {
    // >= $0.01: 4位小数 (如 0.0850)
    return numPrice.toFixed(4);
  } else if (numPrice >= 0.0001) {
    // >= $0.0001: 6位小数 (如 0.008500)
    return numPrice.toFixed(6);
  } else {
    // < $0.0001: 8位小数 (如 0.00008500)
    return numPrice.toFixed(8);
  }
}

// 格式化数量（根据交易对调整）
export function formatQuantity(quantity: string | number, symbol?: string): string {
  const numQty = typeof quantity === 'string' ? parseFloat(quantity) : quantity;
  
  if (isNaN(numQty)) {
    return '-';
  }

  // 根据代币类型决定精度
  if (symbol) {
    const baseAsset = symbol.split('/')[0];
    
    // 高价币 (>$5000): 4位小数
    if (['TITAN', 'GENESIS', 'LUNAR'].includes(baseAsset)) {
      return numQty.toFixed(4);
    }
    // 中高价币 ($1000-$5000): 4位小数
    if (['ORACLE', 'QUANTUM', 'NOVA'].includes(baseAsset)) {
      return numQty.toFixed(4);
    }
    // 中价币 ($100-$1000): 3位小数
    if (['ATLAS', 'COSMOS', 'NEXUS'].includes(baseAsset)) {
      return numQty.toFixed(3);
    }
    // 中低价币 ($10-$100): 2位小数
    if (['VERTEX', 'AURORA', 'ZEPHYR'].includes(baseAsset)) {
      return numQty.toFixed(2);
    }
    // 低价币 ($1-$10): 2位小数
    if (['PRISM', 'PULSE'].includes(baseAsset)) {
      return numQty.toFixed(2);
    }
    // 超低价币 (<$1): 整数
    if (baseAsset === 'ARCANA') {
      return numQty.toFixed(0);
    }
    
    return numQty.toFixed(4);
  }

  // 默认根据数量大小决定
  if (numQty >= 1000) {
    return numQty.toFixed(0);
  } else if (numQty >= 1) {
    return numQty.toFixed(2);
  } else if (numQty >= 0.01) {
    return numQty.toFixed(4);
  } else {
    return numQty.toFixed(8);
  }
}

// 格式化金额（带千分位）
export function formatAmount(amount: string | number, decimals: number = 2): string {
  const numAmount = typeof amount === 'string' ? parseFloat(amount) : amount;
  
  if (isNaN(numAmount)) {
    return '-';
  }

  return numAmount.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

// 格式化百分比
export function formatPercent(percent: string | number): string {
  const numPercent = typeof percent === 'string' ? parseFloat(percent) : percent;
  
  if (isNaN(numPercent)) {
    return '-';
  }

  const sign = numPercent >= 0 ? '+' : '';
  return `${sign}${numPercent.toFixed(2)}%`;
}

// 获取价格的step值（用于input的step属性）
export function getPriceStep(price: number): string {
  if (price >= 1000) return '1';
  if (price >= 100) return '0.1';
  if (price >= 1) return '0.01';
  if (price >= 0.01) return '0.001';
  if (price >= 0.0001) return '0.00001';
  return '0.00000001';
}

// 获取数量的step值
export function getQuantityStep(symbol: string): string {
  const baseAsset = symbol.split('/')[0];
  
  // 高价和中高价币
  if (['TITAN', 'GENESIS', 'LUNAR', 'ORACLE', 'QUANTUM', 'NOVA'].includes(baseAsset)) {
    return '0.0001';
  }
  // 中价币
  if (['ATLAS', 'COSMOS', 'NEXUS'].includes(baseAsset)) {
    return '0.001';
  }
  // 中低价和低价币
  if (['VERTEX', 'AURORA', 'ZEPHYR', 'PRISM', 'PULSE'].includes(baseAsset)) {
    return '0.01';
  }
  // 超低价币
  if (baseAsset === 'ARCANA') {
    return '1';
  }
  
  return '0.0001';
}

// 格式化成交量（千位转换）
export function formatVolume(volume: string | number): string {
  const numVolume = typeof volume === 'string' ? parseFloat(volume) : volume;
  
  if (isNaN(numVolume) || numVolume === 0) {
    return '-';
  }

  if (numVolume >= 1_000_000_000) {
    return (numVolume / 1_000_000_000).toFixed(2) + 'B';
  } else if (numVolume >= 1_000_000) {
    return (numVolume / 1_000_000).toFixed(2) + 'M';
  } else if (numVolume >= 1_000) {
    return (numVolume / 1_000).toFixed(2) + 'K';
  } else {
    return numVolume.toFixed(2);
  }
}

