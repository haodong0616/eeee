package simulator

import (
	"expchange-backend/database"
	"expchange-backend/matching"
	"expchange-backend/models"
	"expchange-backend/utils"
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DynamicOrderBookSimulator 动态订单簿模拟器 - 根据数据库配置决定模拟哪些交易对
type DynamicOrderBookSimulator struct {
	matchingManager *matching.Manager
	wsHub           interface {
		BroadcastOrderBook(data interface{})
		BroadcastTrade(data interface{})
	} // WebSocket Hub
	running          bool
	virtualUserID    string
	activePairs      map[string]bool                // 当前活跃的模拟交易对
	pairConfigs      map[string]*models.TradingPair // 缓存交易对配置
	configUpdateChan chan string                    // 配置更新通知通道
}

func NewDynamicOrderBookSimulator(matchingManager *matching.Manager, wsHub interface {
	BroadcastOrderBook(data interface{})
	BroadcastTrade(data interface{})
}) *DynamicOrderBookSimulator {
	// 创建或获取虚拟用户用于挂单
	var virtualUser models.User
	walletAddr := "0x0000000000000000000000000000000000000000"

	// 使用FirstOrCreate避免"record not found"日志
	database.DB.Where("wallet_address = ?", walletAddr).FirstOrCreate(&virtualUser, models.User{
		WalletAddress: walletAddr,
		Nonce:         "virtual_simulator",
		UserLevel:     "normal",
	})

	// 检查是否需要初始化余额
	var count int64
	database.DB.Model(&models.Balance{}).Where("user_id = ?", virtualUser.ID).Count(&count)

	if count == 0 {
		// 首次创建，为虚拟用户充值大量资金（所有可能的资产）
		assets := []string{
			"TITAN", "GENESIS", "LUNAR", "ORACLE", "QUANTUM", "NOVA",
			"ATLAS", "COSMOS", "NEXUS", "VERTEX", "AURORA", "ZEPHYR",
			"PRISM", "PULSE", "ARCANA", "BTC", "ETH", "BNB", "SOL", "XRP",
			"USDT",
		}
		for _, asset := range assets {
			balance := models.Balance{
				UserID:    virtualUser.ID,
				Asset:     asset,
				Available: decimal.NewFromFloat(100000000), // 1亿
				Frozen:    decimal.Zero,
			}
			database.DB.Create(&balance)
		}
		log.Println("✅ 创建虚拟模拟用户并充值")
	}

	return &DynamicOrderBookSimulator{
		matchingManager:  matchingManager,
		wsHub:            wsHub,
		running:          false,
		virtualUserID:    virtualUser.ID,
		activePairs:      make(map[string]bool),
		pairConfigs:      make(map[string]*models.TradingPair),
		configUpdateChan: make(chan string, 100),
	}
}

func (s *DynamicOrderBookSimulator) Start() {
	if s.running {
		return
	}

	s.running = true
	log.Println("🎯 动态订单簿模拟器已启动（根据配置动态启用交易对）")

	// 主循环：定期检查哪些交易对启用了模拟器
	go s.monitorPairs()

	// 做市商循环：定期吃掉真实用户的挂单
	go s.marketMakerLoop()
}

func (s *DynamicOrderBookSimulator) Stop() {
	s.running = false
	log.Println("🛑 动态订单簿模拟器已停止")
}

// monitorPairs 监控交易对配置，动态启停模拟器
func (s *DynamicOrderBookSimulator) monitorPairs() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次配置
	defer ticker.Stop()

	// 立即执行一次
	s.updateActivePairs()

	for range ticker.C {
		if !s.running {
			return
		}

		s.updateActivePairs()
	}
}

// updateActivePairs 更新活跃的模拟交易对列表
func (s *DynamicOrderBookSimulator) updateActivePairs() {
	var pairs []models.TradingPair
	// 查询启用了模拟器的交易对
	database.DB.Where("simulator_enabled = ? AND status = ?", true, "active").Find(&pairs)

	newActivePairs := make(map[string]bool)
	for _, pair := range pairs {
		newActivePairs[pair.Symbol] = true

		// 检查配置是否发生变化
		oldConfig, exists := s.pairConfigs[pair.Symbol]
		configChanged := !exists ||
			oldConfig.ActivityLevel != pair.ActivityLevel ||
			oldConfig.OrderbookDepth != pair.OrderbookDepth ||
			oldConfig.TradeFrequency != pair.TradeFrequency ||
			oldConfig.VirtualTradePer10s != pair.VirtualTradePer10s ||
			!oldConfig.PriceVolatility.Equal(pair.PriceVolatility) ||
			!oldConfig.PriceSpreadRatio.Equal(pair.PriceSpreadRatio)

		// 更新配置缓存
		pairCopy := pair
		s.pairConfigs[pair.Symbol] = &pairCopy

		// 如果是新启用的交易对，启动模拟goroutine
		if !s.activePairs[pair.Symbol] {
			log.Printf("🟢 启用 %s 的订单簿模拟器 (活跃度:%d)", pair.Symbol, pair.ActivityLevel)
			go s.maintainOrderBook(pair.Symbol)
		} else if configChanged {
			// 配置改变，发送更新通知（goroutine会重新加载配置）
			log.Printf("🔄 %s 配置已更新 (活跃度:%d, 深度:%d档, 频率:%ds, 波动:%.3f%%, 虚拟成交:%d笔/10秒)",
				pair.Symbol, pair.ActivityLevel, pair.OrderbookDepth, pair.TradeFrequency,
				pair.PriceVolatility.InexactFloat64()*100, pair.VirtualTradePer10s)
			select {
			case s.configUpdateChan <- pair.Symbol:
			default:
			}
		}
	}

	// 更新活跃列表
	s.activePairs = newActivePairs
}

// maintainOrderBook 维护某个交易对的订单簿
func (s *DynamicOrderBookSimulator) maintainOrderBook(symbol string) {
	// 获取交易对配置
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		log.Printf("❌ 获取交易对配置失败: %s", symbol)
		return
	}

	// 首次启动时立即创建订单
	s.createOrdersWithConfig(symbol, &pair)

	// 根据配置动态调整更新频率（支持极速模式）
	// ActivityLevel: 1-10 → 订单簿更新间隔: 20秒-1秒
	orderbookInterval := 22 - (pair.ActivityLevel * 2) // 1→20s, 5→12s, 8→6s, 10→2s
	if pair.ActivityLevel >= 9 {
		orderbookInterval = 1 // 活跃度9-10: 1秒极速更新 🚀
	} else if orderbookInterval < 2 {
		orderbookInterval = 2
	}
	orderbookTicker := time.NewTicker(time.Duration(orderbookInterval) * time.Second)
	defer orderbookTicker.Stop()

	// 根据 TradeFrequency 配置成交频率（默认20秒，范围5-60秒）
	tradeFrequency := pair.TradeFrequency
	if tradeFrequency < 5 {
		tradeFrequency = 5
	}
	if tradeFrequency > 60 {
		tradeFrequency = 60
	}
	// 添加±30%的随机波动
	minInterval := int(float64(tradeFrequency) * 0.7)
	maxInterval := int(float64(tradeFrequency) * 1.3)
	tradeTicker := time.NewTicker(time.Duration(minInterval+rand.Intn(maxInterval-minInterval)) * time.Second)
	defer tradeTicker.Stop()

	for {
		select {
		case <-orderbookTicker.C:
			if !s.running || !s.activePairs[symbol] {
				log.Printf("🔴 停止 %s 的订单簿模拟器", symbol)
				s.cleanupOrders(symbol)
				return
			}
			// 重新获取配置并更新订单簿
			if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err == nil {
				s.pairConfigs[symbol] = &pair
				s.createOrdersWithConfig(symbol, &pair)
			}

		case <-tradeTicker.C:
			if !s.running || !s.activePairs[symbol] {
				return
			}
			// 重新获取配置
			if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err == nil {
				s.pairConfigs[symbol] = &pair
				// 模拟市价成交
				s.simulateMarketTradeWithConfig(symbol, &pair)
				// 根据配置重置下一次成交时间
				minInterval := int(float64(pair.TradeFrequency) * 0.7)
				maxInterval := int(float64(pair.TradeFrequency) * 1.3)
				tradeTicker.Reset(time.Duration(minInterval+rand.Intn(maxInterval-minInterval)) * time.Second)
			}

		case <-s.configUpdateChan:
			// 收到配置更新通知，立即重新加载配置
			if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err == nil {
				s.pairConfigs[symbol] = &pair

				// 重新计算并重置订单簿更新间隔（支持极速模式）
				newOrderbookInterval := 22 - (pair.ActivityLevel * 2)
				if newOrderbookInterval < 1 {
					newOrderbookInterval = 1 // 极速：1秒更新
				}
				orderbookTicker.Reset(time.Duration(newOrderbookInterval) * time.Second)

				// 立即更新一次订单簿
				s.createOrdersWithConfig(symbol, &pair)

				log.Printf("✅ %s 配置已热更新生效", symbol)
			}
		}
	}
}

// createOrders 创建模拟订单（兼容旧调用）
func (s *DynamicOrderBookSimulator) createOrders(symbol string) {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		return
	}
	s.createOrdersWithConfig(symbol, &pair)
}

// createOrdersWithConfig 使用配置创建模拟订单
func (s *DynamicOrderBookSimulator) createOrdersWithConfig(symbol string, pair *models.TradingPair) {
	// 获取当前价格
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		// 如果没有成交记录，跳过
		return
	}

	currentPrice, _ := lastTrade.Price.Float64()
	if currentPrice <= 0 {
		return
	}

	// 清理虚拟用户的旧挂单
	database.DB.Where("user_id = ? AND symbol = ? AND status IN (?, ?)",
		s.virtualUserID, symbol, "pending", "partial").
		Delete(&models.Order{})

	// 使用配置的档位数和波动率创建买卖单
	s.createBuyOrdersWithConfig(symbol, currentPrice, pair)
	s.createSellOrdersWithConfig(symbol, currentPrice, pair)

	log.Printf("📚 %s 订单簿已更新 (市价:%.8f, 买档:%d, 卖档:%d)",
		symbol, currentPrice, pair.OrderbookDepth, pair.OrderbookDepth)

	// 推送订单簿更新到WebSocket（实时推送）
	if s.wsHub != nil {
		orderbook := s.matchingManager.GetOrderBook(symbol, 50) // 获取50档深度

		// 转换为JSON友好格式（确保价格是字符串）
		bids := make([]map[string]string, 0, len(orderbook.Bids))
		for _, bid := range orderbook.Bids {
			bids = append(bids, map[string]string{
				"price":    bid.Price.String(),
				"quantity": bid.Quantity.String(),
			})
		}

		asks := make([]map[string]string, 0, len(orderbook.Asks))
		for _, ask := range orderbook.Asks {
			asks = append(asks, map[string]string{
				"price":    ask.Price.String(),
				"quantity": ask.Quantity.String(),
			})
		}

		// 调试日志：显示前3档价格，确认排序
		if len(bids) >= 3 && len(asks) >= 3 {
			log.Printf("📊 %s 推送盘口 - 买盘[%s, %s, %s] 卖盘[%s, %s, %s]",
				symbol,
				bids[0]["price"], bids[1]["price"], bids[2]["price"],
				asks[0]["price"], asks[1]["price"], asks[2]["price"])
		}

		s.wsHub.BroadcastOrderBook(map[string]interface{}{
			"symbol": symbol,
			"bids":   bids, // 已排序：价格从高到低
			"asks":   asks, // 已排序：价格从低到高
		})
	}
}

// createBuyOrders 创建买单（兼容旧调用）
func (s *DynamicOrderBookSimulator) createBuyOrders(symbol string, currentPrice float64) {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		return
	}
	s.createBuyOrdersWithConfig(symbol, currentPrice, &pair)
}

// createBuyOrdersWithConfig 使用配置创建买单
func (s *DynamicOrderBookSimulator) createBuyOrdersWithConfig(symbol string, currentPrice float64, pair *models.TradingPair) {
	// 1. 查询真实用户的最低卖单价格（虚拟买单不能高于这个价格）
	var lowestRealSellOrder models.Order
	err := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
		Where("symbol = ? AND user_id != ? AND status = ? AND side = 'sell'",
			symbol, s.virtualUserID, "pending").
		Order("price ASC").
		First(&lowestRealSellOrder).Error

	// 2. 确定虚拟买单的最高价格上限
	var maxVirtualBuyPrice float64
	if err == nil {
		// 如果有真实卖单，虚拟买单价格必须低于真实卖单（避免埋掉真实卖单）
		lowestRealPrice, _ := lowestRealSellOrder.Price.Float64()
		maxVirtualBuyPrice = lowestRealPrice * 0.999 // 留0.1%的价差
		log.Printf("🔍 %s 发现真实卖单 @ %.8f，虚拟买单上限: %.8f", symbol, lowestRealPrice, maxVirtualBuyPrice)
	} else {
		// 没有真实卖单，可以接近市价
		maxVirtualBuyPrice = currentPrice * 0.999
	}

	// 使用配置的档位数（默认15，范围5-30）
	levels := pair.OrderbookDepth
	if levels < 5 {
		levels = 5
	}
	if levels > 30 {
		levels = 30
	}

	// 使用配置的波动率（默认0.01=1%）
	volatility, _ := pair.PriceVolatility.Float64()
	if volatility <= 0 {
		volatility = 0.01
	}

	// 根据价格分布范围倍数调整价格分布
	spreadRatio, _ := pair.PriceSpreadRatio.Float64()
	if spreadRatio <= 0 {
		spreadRatio = 1.0
	}
	maxSpread := volatility * spreadRatio

	for i := 1; i <= levels; i++ {
		// 价格从maxVirtualBuyPrice向下递减
		priceOffset := (maxSpread / float64(levels)) * float64(i)
		price := maxVirtualBuyPrice * (1 - priceOffset)

		// 调试：打印前3档价格
		if i <= 3 {
			log.Printf("  买%d档: %.8f (上限:%.8f, 偏移:%.2f%%)", i, price, maxVirtualBuyPrice, priceOffset*100)
		}

		// 数量随距离增加
		quantity := s.getQuantityForSymbol(symbol, currentPrice) * (1 + float64(i)*0.6)

		// 根据活跃度调整数量波动（活跃度越高，波动越大）
		volatilityFactor := 0.2 + (float64(pair.ActivityLevel) * 0.06) // 1→0.26, 5→0.5, 10→0.8
		quantity = quantity * (1 - volatilityFactor + rand.Float64()*volatilityFactor*2)

		order := models.Order{
			UserID:    s.virtualUserID,
			Symbol:    symbol,
			OrderType: "limit",
			Side:      "buy",
			Price:     decimal.NewFromFloat(price),
			Quantity:  decimal.NewFromFloat(quantity),
			FilledQty: decimal.Zero,
			Status:    "pending",
		}

		database.DB.Create(&order)
		s.matchingManager.AddOrder(&order) // 进入匹配引擎用于展示盘口
	}

	log.Printf("📚 %s 虚拟买单已创建 (x%d档)", symbol, levels)
}

// createSellOrders 创建卖单（兼容旧调用）
func (s *DynamicOrderBookSimulator) createSellOrders(symbol string, currentPrice float64) {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		return
	}
	s.createSellOrdersWithConfig(symbol, currentPrice, &pair)
}

// createSellOrdersWithConfig 使用配置创建卖单
func (s *DynamicOrderBookSimulator) createSellOrdersWithConfig(symbol string, currentPrice float64, pair *models.TradingPair) {
	// 1. 查询真实用户的最高买单价格（虚拟卖单不能低于这个价格）
	var highestRealBuyOrder models.Order
	err := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
		Where("symbol = ? AND user_id != ? AND status = ? AND side = 'buy'",
			symbol, s.virtualUserID, "pending").
		Order("price DESC").
		First(&highestRealBuyOrder).Error

	// 2. 确定虚拟卖单的最低价格下限
	var minVirtualSellPrice float64
	if err == nil {
		// 如果有真实买单，虚拟卖单价格必须高于真实买单（避免埋掉真实买单）
		highestRealPrice, _ := highestRealBuyOrder.Price.Float64()
		minVirtualSellPrice = highestRealPrice * 1.001 // 留0.1%的价差
		log.Printf("🔍 %s 发现真实买单 @ %.8f，虚拟卖单下限: %.8f", symbol, highestRealPrice, minVirtualSellPrice)
	} else {
		// 没有真实买单，可以接近市价
		minVirtualSellPrice = currentPrice * 1.001
	}

	// 使用配置的档位数
	levels := pair.OrderbookDepth
	if levels < 5 {
		levels = 5
	}
	if levels > 30 {
		levels = 30
	}

	// 使用配置的波动率
	volatility, _ := pair.PriceVolatility.Float64()
	if volatility <= 0 {
		volatility = 0.01
	}

	// 根据价格分布范围倍数调整价格分布
	spreadRatio, _ := pair.PriceSpreadRatio.Float64()
	if spreadRatio <= 0 {
		spreadRatio = 1.0
	}
	maxSpread := volatility * spreadRatio

	for i := 1; i <= levels; i++ {
		// 价格从minVirtualSellPrice向上递增
		priceOffset := (maxSpread / float64(levels)) * float64(i)
		price := minVirtualSellPrice * (1 + priceOffset)

		// 调试：打印前3档价格
		if i <= 3 {
			log.Printf("  卖%d档: %.8f (下限:%.8f, 偏移:%.2f%%)", i, price, minVirtualSellPrice, priceOffset*100)
		}

		// 数量随距离增加
		quantity := s.getQuantityForSymbol(symbol, currentPrice) * (1 + float64(i)*0.6)

		// 根据活跃度调整数量波动
		volatilityFactor := 0.2 + (float64(pair.ActivityLevel) * 0.06)
		quantity = quantity * (1 - volatilityFactor + rand.Float64()*volatilityFactor*2)

		order := models.Order{
			UserID:    s.virtualUserID,
			Symbol:    symbol,
			OrderType: "limit",
			Side:      "sell",
			Price:     decimal.NewFromFloat(price),
			Quantity:  decimal.NewFromFloat(quantity),
			FilledQty: decimal.Zero,
			Status:    "pending",
		}

		database.DB.Create(&order)
		s.matchingManager.AddOrder(&order) // 进入匹配引擎用于展示盘口
	}

	log.Printf("📚 %s 虚拟卖单已创建 (x%d档)", symbol, levels)
}

// getQuantityForSymbol 根据交易对和价格计算合适的数量
func (s *DynamicOrderBookSimulator) getQuantityForSymbol(symbol string, price float64) float64 {
	// 根据价格区间动态计算数量
	if price > 10000 { // 高价币（如BTC）
		return 0.01 + rand.Float64()*0.05
	} else if price > 1000 { // 中高价币（如ETH）
		return 0.1 + rand.Float64()*0.5
	} else if price > 100 { // 中价币（如BNB）
		return 1 + rand.Float64()*10
	} else if price > 10 { // 中低价币
		return 10 + rand.Float64()*50
	} else if price > 1 { // 低价币
		return 50 + rand.Float64()*200
	} else if price > 0.1 { // 极低价币
		return 500 + rand.Float64()*2000
	} else { // 超低价币
		return 5000 + rand.Float64()*20000
	}
}

// cleanupOrders 清理某个交易对的模拟订单
func (s *DynamicOrderBookSimulator) cleanupOrders(symbol string) {
	result := database.DB.Where("user_id = ? AND symbol = ? AND status IN (?, ?)",
		s.virtualUserID, symbol, "pending", "partial").
		Delete(&models.Order{})

	if result.RowsAffected > 0 {
		log.Printf("🧹 清理了 %s 的 %d 个模拟订单", symbol, result.RowsAffected)
	}
}

// simulateMarketTrade 模拟市价成交（兼容旧调用）
func (s *DynamicOrderBookSimulator) simulateMarketTrade(symbol string) {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		return
	}
	s.simulateMarketTradeWithConfig(symbol, &pair)
}

// simulateMarketTradeWithConfig 使用配置模拟市价成交
func (s *DynamicOrderBookSimulator) simulateMarketTradeWithConfig(symbol string, pair *models.TradingPair) {
	// 获取当前价格
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		return
	}

	currentPrice, _ := lastTrade.Price.Float64()
	if currentPrice <= 0 {
		return
	}

	// 随机决定买还是卖（50%概率）
	side := "buy"
	if rand.Float64() < 0.5 {
		side = "sell"
	}

	// 根据活跃度调整成交量范围
	// ActivityLevel: 1→小量, 10→大量
	volumeFactor := 0.3 + (float64(pair.ActivityLevel) * 0.17) // 1→0.47, 5→1.15, 10→2.0
	quantity := s.getQuantityForSymbol(symbol, currentPrice) * (volumeFactor + rand.Float64()*volumeFactor)

	// 使用配置的波动率计算价格
	volatility, _ := pair.PriceVolatility.Float64()
	if volatility <= 0 {
		volatility = 0.01
	}

	// 根据活跃度调整价格波动范围
	// ActivityLevel: 1→小波动, 10→大波动
	maxPriceMove := volatility * float64(pair.ActivityLevel) * 0.1 // 1→0.1%, 5→0.5%, 10→1%

	var price float64
	if side == "buy" {
		// 买单价格略高于当前价（吃卖单）
		price = currentPrice * (1 + volatility*0.1 + rand.Float64()*maxPriceMove)
	} else {
		// 卖单价格略低于当前价（吃买单）
		price = currentPrice * (1 - volatility*0.1 - rand.Float64()*maxPriceMove)
	}

	// 创建限价单（实际上是模拟市价单的效果）
	order := models.Order{
		UserID:    s.virtualUserID,
		Symbol:    symbol,
		OrderType: "limit",
		Side:      side,
		Price:     decimal.NewFromFloat(price),
		Quantity:  decimal.NewFromFloat(quantity),
		FilledQty: decimal.Zero,
		Status:    "pending",
	}

	database.DB.Create(&order)
	s.matchingManager.AddOrder(&order) // 进入匹配引擎

	log.Printf("💹 %s 模拟市价成交: %s %.8f @ %.8f", symbol, side, quantity, price)
}

// marketMakerLoop 做市商循环 - 极速吃单模式 🚀
func (s *DynamicOrderBookSimulator) marketMakerLoop() {
	// 极速模式：每200毫秒检查一次（每秒5次）
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	log.Println("🤖 做市商极速模式已启动，嘎嘎快速成交中...")

	for range ticker.C {
		if !s.running {
			return
		}

		// 为每个活跃交易对执行做市
		for symbol := range s.activePairs {
			// 1. 尝试吃真实用户订单
			s.makeMarketForSymbol(symbol)

			// 2. 如果没有真实订单，生成虚拟成交（保持活跃度）
			s.createVirtualTrade(symbol)
		}
	}
}

// makeMarketForSymbol 为指定交易对做市
func (s *DynamicOrderBookSimulator) makeMarketForSymbol(symbol string) {
	// 1. 分别查询真实用户的买单和卖单（排除虚拟用户）
	var buyOrders []models.Order
	var sellOrders []models.Order

	database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
		Where("symbol = ? AND user_id != ? AND status = ? AND side = 'buy'",
					symbol, s.virtualUserID, "pending").
		Order("price DESC"). // 买单按价格从高到低排序
		Find(&buyOrders)

	database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
		Where("symbol = ? AND user_id != ? AND status = ? AND side = 'sell'",
					symbol, s.virtualUserID, "pending").
		Order("price ASC"). // 卖单按价格从低到高排序
		Find(&sellOrders)

	if len(buyOrders) == 0 && len(sellOrders) == 0 {
		return // 没有真实用户订单，跳过
	}

	log.Printf("🔍 %s 发现 %d 个买单, %d 个卖单（真实用户）", symbol, len(buyOrders), len(sellOrders))

	// 2. 获取当前市场价格
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		return
	}

	currentPrice := lastTrade.Price

	// 3. 极致激进策略：吃掉所有真实订单（不限价格范围）
	// 理由：用户挂单了就应该立即成交，不管价格偏离多少

	// 3.1 吃掉所有真实买单
	for i := range buyOrders {
		s.eatSingleOrder(&buyOrders[i], currentPrice, symbol)
	}

	// 3.2 吃掉所有真实卖单
	for i := range sellOrders {
		s.eatSingleOrder(&sellOrders[i], currentPrice, symbol)
	}
}

// eatSingleOrder 吃掉单个订单（智能价格匹配策略）
func (s *DynamicOrderBookSimulator) eatSingleOrder(targetOrder *models.Order, currentPrice decimal.Decimal, symbol string) {
	// 1. 从数据库重新查询订单状态（避免使用过期数据）
	var freshOrder models.Order
	if err := database.DB.Where("id = ?", targetOrder.ID).First(&freshOrder).Error; err != nil {
		return // 订单不存在或已删除
	}

	// 2. 检查订单状态和剩余数量
	if freshOrder.Status != "pending" && freshOrder.Status != "partial" {
		return // 订单已完成或取消
	}

	remainingQty := freshOrder.Quantity.Sub(freshOrder.FilledQty)
	if remainingQty.LessThanOrEqual(decimal.Zero) {
		return // 已完全成交
	}

	orderPrice := freshOrder.Price
	log.Printf("✅ %s 准备吃单: %s %s @ %s (剩余:%s)",
		symbol, freshOrder.Side, freshOrder.Quantity.String(), orderPrice.String(), remainingQty.String())

	// 3. 获取当前盘口价格（从虚拟订单中获取）
	var bestOppositePrice decimal.Decimal
	var hasBestPrice bool

	if freshOrder.Side == "buy" {
		// 用户想买，查询虚拟卖单的最低价
		var lowestSellOrder models.Order
		err := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
			Where("symbol = ? AND user_id = ? AND status = ? AND side = 'sell'",
				symbol, s.virtualUserID, "pending").
			Order("price ASC").
			First(&lowestSellOrder).Error
		if err == nil {
			bestOppositePrice = lowestSellOrder.Price
			hasBestPrice = true
			log.Printf("  📊 盘口最低卖价: %s", bestOppositePrice.String())
		}
	} else {
		// 用户想卖，查询虚拟买单的最高价
		var highestBuyOrder models.Order
		err := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
			Where("symbol = ? AND user_id = ? AND status = ? AND side = 'buy'",
				symbol, s.virtualUserID, "pending").
			Order("price DESC").
			First(&highestBuyOrder).Error
		if err == nil {
			bestOppositePrice = highestBuyOrder.Price
			hasBestPrice = true
			log.Printf("  📊 盘口最高买价: %s", bestOppositePrice.String())
		}
	}

	// 4. 智能价格策略：如果用户价格更优，先按盘口价格成交
	var matchingSide string
	var matchingPrice decimal.Decimal

	if freshOrder.Side == "buy" {
		matchingSide = "sell"
		// 如果用户买价 > 盘口卖价，先以盘口价成交（用户得到更好的价格）
		if hasBestPrice && orderPrice.GreaterThan(bestOppositePrice) {
			matchingPrice = bestOppositePrice
			log.Printf("  💡 用户买价(%s) > 盘口卖价(%s)，以盘口价成交", orderPrice.String(), bestOppositePrice.String())
		} else {
			matchingPrice = orderPrice
		}
	} else {
		matchingSide = "buy"
		// 如果用户卖价 < 盘口买价，先以盘口价成交（用户得到更好的价格）
		if hasBestPrice && orderPrice.LessThan(bestOppositePrice) {
			matchingPrice = bestOppositePrice
			log.Printf("  💡 用户卖价(%s) < 盘口买价(%s)，以盘口价成交", orderPrice.String(), bestOppositePrice.String())
		} else {
			matchingPrice = orderPrice
		}
	}

	// 5. 吃掉全部剩余数量
	eatQty := remainingQty

	log.Printf("  💰 决定成交价: %s, 数量: %s", matchingPrice.String(), eatQty.String())

	// 创建对手单（限价单）
	matchingOrder := models.Order{
		UserID:    s.virtualUserID,
		Symbol:    symbol,
		OrderType: "limit",
		Side:      matchingSide,
		Price:     matchingPrice,
		Quantity:  eatQty,
		FilledQty: decimal.Zero,
		Status:    "pending",
	}

	if err := database.DB.Create(&matchingOrder).Error; err != nil {
		log.Printf("❌ %s 创建对手单失败: %v", symbol, err)
		return
	}

	log.Printf("🎯 %s 对手单已创建: %s %s @ %s (ID:%s)",
		symbol, matchingSide, eatQty.String(), matchingPrice.String(), matchingOrder.ID)

	// ⚠️ 关键：直接通过匹配引擎提交，让引擎处理
	s.matchingManager.AddOrder(&matchingOrder)

	log.Printf("  ✅ %s 已提交到匹配引擎，等待自动匹配...", symbol)

	// 计算盈亏（与当前市价对比）
	var profitLossUSDT decimal.Decimal
	var profitPercent decimal.Decimal

	if matchingSide == "buy" {
		// 虚拟用户买入，如果买价低于市价就是赚的
		profitLossUSDT = currentPrice.Sub(matchingPrice).Mul(eatQty).Mul(currentPrice)
		if !matchingPrice.IsZero() {
			profitPercent = currentPrice.Sub(matchingPrice).Div(matchingPrice).Mul(decimal.NewFromInt(100))
		}
	} else {
		// 虚拟用户卖出，如果卖价高于市价就是赚的
		profitLossUSDT = matchingPrice.Sub(currentPrice).Mul(eatQty).Mul(currentPrice)
		if !currentPrice.IsZero() {
			profitPercent = matchingPrice.Sub(currentPrice).Div(currentPrice).Mul(decimal.NewFromInt(100))
		}
	}

	// 保存盈亏记录
	pnlRecord := models.MarketMakerPnL{
		Symbol:        symbol,
		Side:          matchingSide,
		ExecutePrice:  matchingPrice,
		MarketPrice:   currentPrice,
		Quantity:      eatQty,
		ProfitLoss:    profitLossUSDT,
		ProfitPercent: profitPercent,
	}
	database.DB.Create(&pnlRecord)

	// 输出日志
	profitSign := ""
	if profitLossUSDT.GreaterThan(decimal.Zero) {
		profitSign = "📈 盈利"
	} else if profitLossUSDT.LessThan(decimal.Zero) {
		profitSign = "📉 亏损"
	} else {
		profitSign = "➖ 持平"
	}

	log.Printf("🤖 做市商吃单: %s %s %s @ %s (市价: %s, %s: %s USDT, %.2f%%)",
		symbol, matchingSide, eatQty.StringFixed(4), matchingPrice.StringFixed(8), currentPrice.StringFixed(8),
		profitSign, profitLossUSDT.StringFixed(2), profitPercent.InexactFloat64())
}

// createVirtualTrade 创建虚拟成交，确保市场始终活跃（按配置频率）
func (s *DynamicOrderBookSimulator) createVirtualTrade(symbol string) {
	// 获取交易对配置
	config, exists := s.pairConfigs[symbol]
	if !exists {
		return
	}

	// 根据配置的虚拟成交频率决定是否成交
	// VirtualTradePer10s: 每10秒N笔（1-30）
	// 当前每200ms调用一次（每10秒50次调用）
	tradesPer10s := config.VirtualTradePer10s
	if tradesPer10s < 1 {
		tradesPer10s = 10 // 默认每10秒10笔
	}
	if tradesPer10s > 30 {
		tradesPer10s = 30 // 最多每10秒30笔
	}

	// 每200ms调用一次 = 每10秒50次调用机会
	// 概率 = 目标笔数 / 50次机会
	probability := float64(tradesPer10s) / 50.0

	// 根据概率决定是否成交
	if rand.Float64() > probability {
		return // 这次不成交
	}

	// 生成一笔虚拟成交
	s.doCreateVirtualTrade(symbol, config)
}

// doCreateVirtualTrade 执行虚拟成交创建
func (s *DynamicOrderBookSimulator) doCreateVirtualTrade(symbol string, pair *models.TradingPair) {
	// 从盘口获取买一/卖一价格（确保和盘口一致）
	orderbook := s.matchingManager.GetOrderBook(symbol, 5)

	if len(orderbook.Bids) == 0 || len(orderbook.Asks) == 0 {
		return // 没有盘口数据，跳过
	}

	// 随机在买一/卖一之间成交（更真实）
	buyPrice := orderbook.Bids[0].Price  // 买一价格
	sellPrice := orderbook.Asks[0].Price // 卖一价格

	// 50%概率在买一成交，50%在卖一成交（价格上下波动）
	var newPrice decimal.Decimal
	var side string
	if rand.Float64() < 0.5 {
		newPrice = buyPrice // 在买一价格成交（价格下跌）
		side = "sell"
	} else {
		newPrice = sellPrice // 在卖一价格成交（价格上涨）
		side = "buy"
	}

	// 小额成交（使用统一的数量规则）
	minQty, maxQty := utils.GetQuantityByPrice(newPrice)
	qtyRange := maxQty.Sub(minQty)
	randomFactor := decimal.NewFromFloat(rand.Float64())
	quantity := minQty.Add(qtyRange.Mul(randomFactor))
	quantity = utils.RoundQuantity(quantity, newPrice) // 按精度舍入

	// 创建虚拟成交记录
	trade := models.Trade{
		Symbol:      symbol,
		BuyOrderID:  "virtual-buy-" + symbol,
		SellOrderID: "virtual-sell-" + symbol,
		Price:       newPrice,
		Quantity:    quantity,
	}
	database.DB.Create(&trade)

	// 立即推送到WebSocket
	if s.wsHub != nil {
		s.wsHub.BroadcastTrade(map[string]interface{}{
			"symbol":     symbol,
			"price":      newPrice.String(),
			"quantity":   quantity.String(),
			"side":       side,
			"created_at": trade.CreatedAt,
		})
	}
}
