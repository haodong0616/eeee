package simulator

import (
	"expchange-backend/database"
	"expchange-backend/matching"
	"expchange-backend/models"
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

// DynamicOrderBookSimulator 动态订单簿模拟器 - 根据数据库配置决定模拟哪些交易对
type DynamicOrderBookSimulator struct {
	matchingManager  *matching.Manager
	wsHub            interface{ BroadcastOrderBook(data interface{}) } // WebSocket Hub
	running          bool
	virtualUserID    string
	activePairs      map[string]bool                // 当前活跃的模拟交易对
	pairConfigs      map[string]*models.TradingPair // 缓存交易对配置
	configUpdateChan chan string                    // 配置更新通知通道
}

func NewDynamicOrderBookSimulator(matchingManager *matching.Manager, wsHub interface{ BroadcastOrderBook(data interface{}) }) *DynamicOrderBookSimulator {
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
			!oldConfig.PriceVolatility.Equal(pair.PriceVolatility)

		// 更新配置缓存
		pairCopy := pair
		s.pairConfigs[pair.Symbol] = &pairCopy

		// 如果是新启用的交易对，启动模拟goroutine
		if !s.activePairs[pair.Symbol] {
			log.Printf("🟢 启用 %s 的订单簿模拟器 (活跃度:%d)", pair.Symbol, pair.ActivityLevel)
			go s.maintainOrderBook(pair.Symbol)
		} else if configChanged {
			// 配置改变，发送更新通知（goroutine会重新加载配置）
			log.Printf("🔄 %s 配置已更新 (活跃度:%d, 深度:%d档, 频率:%ds, 波动:%.3f%%)",
				pair.Symbol, pair.ActivityLevel, pair.OrderbookDepth, pair.TradeFrequency,
				pair.PriceVolatility.InexactFloat64()*100)
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

	// 推送订单簿更新到WebSocket（实时推送）
	if s.wsHub != nil {
		orderbook := s.matchingManager.GetOrderBook(symbol, 50) // 获取50档深度
		s.wsHub.BroadcastOrderBook(map[string]interface{}{
			"symbol": symbol,
			"bids":   orderbook.Bids,
			"asks":   orderbook.Asks,
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

	// 根据活跃度调整价格分布范围
	// ActivityLevel: 1→范围小, 10→范围大
	maxSpread := volatility * float64(pair.ActivityLevel) * 0.5 // 1→0.5%, 5→2.5%, 10→5%

	for i := 1; i <= levels; i++ {
		priceOffset := volatility*0.1 + (maxSpread/float64(levels))*float64(i)
		price := currentPrice * (1 - priceOffset)

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
		s.matchingManager.AddOrder(&order)
	}
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

	// 根据活跃度调整价格分布范围
	maxSpread := volatility * float64(pair.ActivityLevel) * 0.5

	for i := 1; i <= levels; i++ {
		priceOffset := volatility*0.1 + (maxSpread/float64(levels))*float64(i)
		price := currentPrice * (1 + priceOffset)

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
		s.matchingManager.AddOrder(&order)
	}

	log.Printf("📚 %s 订单簿已更新 (买单x%d, 卖单x%d, 活跃度:%d)", symbol, levels, levels, pair.ActivityLevel)
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
	s.matchingManager.AddOrder(&order)

	log.Printf("💹 %s 模拟市价成交: %s %.8f @ %.8f", symbol, side, quantity, price)
}

// marketMakerLoop 做市商循环 - 极速吃单模式 🚀
func (s *DynamicOrderBookSimulator) marketMakerLoop() {
	// 极速模式：每200毫秒检查一次（每秒5次）
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	log.Println("🤖 做市商极速模式已启动，嘎嘎快速吃单中...")

	for range ticker.C {
		if !s.running {
			return
		}

		// 为每个活跃交易对执行做市
		for symbol := range s.activePairs {
			s.makeMarketForSymbol(symbol)
		}
	}
}

// makeMarketForSymbol 为指定交易对做市
func (s *DynamicOrderBookSimulator) makeMarketForSymbol(symbol string) {
	// 1. 查找真实用户的挂单（排除虚拟用户）
	var realOrders []models.Order
	database.DB.Where("symbol = ? AND user_id != ? AND status = ?",
		symbol, s.virtualUserID, "pending").
		Order("created_at ASC"). // 优先吃掉最早的订单
		Limit(5).                // 每次最多吃5个订单
		Find(&realOrders)

	if len(realOrders) == 0 {
		return // 没有真实用户订单，跳过
	}

	// 2. 优先吃买一/卖一，制造价格上下波动效果 🎯
	// 80%概率操作（高频率）
	if rand.Float64() > 0.8 {
		return
	}

	// 按价格排序：优先吃最优价格（买一/卖一）
	var buyOrders, sellOrders []models.Order
	for _, order := range realOrders {
		if order.Side == "buy" {
			buyOrders = append(buyOrders, order)
		} else {
			sellOrders = append(sellOrders, order)
		}
	}

	// 交替吃买单和卖单，产生上下波动
	var targetOrder models.Order
	shouldEatBuy := rand.Float64() < 0.5

	if shouldEatBuy && len(buyOrders) > 0 {
		// 吃买单 → 价格可能下跌
		// 找最高买价
		maxBuyPrice := buyOrders[0].Price
		targetOrder = buyOrders[0]
		for _, order := range buyOrders {
			if order.Price.GreaterThan(maxBuyPrice) {
				maxBuyPrice = order.Price
				targetOrder = order
			}
		}
	} else if len(sellOrders) > 0 {
		// 吃卖单 → 价格可能上涨
		// 找最低卖价
		minSellPrice := sellOrders[0].Price
		targetOrder = sellOrders[0]
		for _, order := range sellOrders {
			if order.Price.LessThan(minSellPrice) {
				minSellPrice = order.Price
				targetOrder = order
			}
		}
	} else if len(buyOrders) > 0 {
		// 回退：吃买单
		targetOrder = buyOrders[0]
	} else {
		return
	}

	// 3. 获取当前市场价格
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		return
	}

	currentPrice := lastTrade.Price

	// 4. 判断是否值得吃掉这个订单（价格合理性检查）
	orderPrice := targetOrder.Price
	priceDeviation := orderPrice.Sub(currentPrice).Div(currentPrice).Abs()

	// 如果价格偏离超过5%，暂不吃（避免亏损过大）
	if priceDeviation.GreaterThan(decimal.NewFromFloat(0.05)) {
		return
	}

	// 5. 创建对手单吃掉目标订单
	var matchingSide string
	var matchingPrice decimal.Decimal

	if targetOrder.Side == "buy" {
		// 真实用户想买，虚拟用户卖给他
		matchingSide = "sell"
		matchingPrice = targetOrder.Price // 以买单价格成交（对真实用户有利）
	} else {
		// 真实用户想卖，虚拟用户买入
		matchingSide = "buy"
		matchingPrice = targetOrder.Price // 以卖单价格成交（对真实用户有利）
	}

	// 吃掉较大数量（70%-100%），制造明显的价格波动
	remainingQty := targetOrder.Quantity.Sub(targetOrder.FilledQty)
	eatRatio := 0.7 + rand.Float64()*0.3 // 70%-100%，更激进
	eatQty := remainingQty.Mul(decimal.NewFromFloat(eatRatio))

	// 创建匹配订单（市价单，立即成交）
	matchingOrder := models.Order{
		UserID:    s.virtualUserID,
		Symbol:    symbol,
		OrderType: "market", // 改为市价单，确保立即成交
		Side:      matchingSide,
		Price:     matchingPrice,
		Quantity:  eatQty,
		FilledQty: decimal.Zero,
		Status:    "pending",
	}

	database.DB.Create(&matchingOrder)
	s.matchingManager.AddOrder(&matchingOrder)

	// 立即生成一笔模拟成交，更新最新价（让价格真正波动）
	trade := models.Trade{
		Symbol:      symbol,
		BuyOrderID:  matchingOrder.ID,
		SellOrderID: targetOrder.ID,
		Price:       matchingPrice,
		Quantity:    eatQty,
	}
	if matchingSide == "buy" {
		trade.BuyOrderID = matchingOrder.ID
		trade.SellOrderID = targetOrder.ID
	} else {
		trade.BuyOrderID = targetOrder.ID
		trade.SellOrderID = matchingOrder.ID
	}
	database.DB.Create(&trade)

	// 6. 计算盈亏（与当前市价对比）
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

	// 7. 保存盈亏记录
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
