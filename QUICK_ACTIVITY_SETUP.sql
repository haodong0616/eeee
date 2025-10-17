-- =========================================
-- 快速设置交易对活跃度
-- =========================================

-- 方案 1：为所有交易对设置默认中等活跃度
UPDATE trading_pairs SET
  activity_level = 5,
  orderbook_depth = 15,
  trade_frequency = 20,
  price_volatility = 0.01
WHERE simulator_enabled = true;

-- 方案 2：差异化配置（推荐）

-- 顶级币：BTC、ETH（高活跃，深盘口）
UPDATE trading_pairs SET
  activity_level = 8,
  orderbook_depth = 25,
  trade_frequency = 15,
  price_volatility = 0.008
WHERE base_asset IN ('BTC', 'ETH');

-- 热门币：PULSE、LUNAR（极活跃）
UPDATE trading_pairs SET
  activity_level = 10,
  orderbook_depth = 20,
  trade_frequency = 8,
  price_volatility = 0.015
WHERE base_asset IN ('PULSE', 'LUNAR', 'TITAN', 'GENESIS');

-- 主流币：中等活跃
UPDATE trading_pairs SET
  activity_level = 6,
  orderbook_depth = 15,
  trade_frequency = 18,
  price_volatility = 0.012
WHERE base_asset IN ('ORACLE', 'QUANTUM', 'NOVA', 'ATLAS');

-- 普通币：正常活跃
UPDATE trading_pairs SET
  activity_level = 4,
  orderbook_depth = 12,
  trade_frequency = 25,
  price_volatility = 0.018
WHERE base_asset IN ('COSMOS', 'NEXUS', 'VERTEX', 'AURORA');

-- 小币：低活跃
UPDATE trading_pairs SET
  activity_level = 2,
  orderbook_depth = 8,
  trade_frequency = 40,
  price_volatility = 0.025
WHERE base_asset IN ('ZEPHYR', 'PRISM', 'ARCANA');

-- 方案 3：全高活跃（让所有盘口都非常活跃）
UPDATE trading_pairs SET
  activity_level = 9,
  orderbook_depth = 20,
  trade_frequency = 10,
  price_volatility = 0.015
WHERE simulator_enabled = true;

-- 查看当前配置
SELECT 
  symbol,
  activity_level as 活跃度,
  orderbook_depth as 档位数,
  trade_frequency as 成交间隔秒,
  price_volatility as 波动率,
  simulator_enabled as 已启用
FROM trading_pairs
WHERE simulator_enabled = true
ORDER BY activity_level DESC;

-- 查看预计效果
SELECT 
  symbol,
  activity_level,
  CONCAT(24 - (activity_level * 2), '秒') as 订单簿更新间隔,
  CONCAT(trade_frequency * 0.7, '-', trade_frequency * 1.3, '秒') as 成交间隔,
  CONCAT(orderbook_depth * 2, '个订单') as 总订单数,
  CONCAT(price_volatility * activity_level * 50, '%') as 最大价差
FROM trading_pairs
WHERE simulator_enabled = true
ORDER BY activity_level DESC;

