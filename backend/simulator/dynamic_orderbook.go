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
	matchingManager *matching.Manager
	running         bool
	virtualUserID   string
	activePairs     map[string]bool // 当前活跃的模拟交易对
}

func NewDynamicOrderBookSimulator(matchingManager *matching.Manager) *DynamicOrderBookSimulator {
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
		matchingManager: matchingManager,
		running:         false,
		virtualUserID:   virtualUser.ID,
		activePairs:     make(map[string]bool),
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

		// 如果是新启用的交易对，启动模拟goroutine
		if !s.activePairs[pair.Symbol] {
			log.Printf("🟢 启用 %s 的订单簿模拟器", pair.Symbol)
			go s.maintainOrderBook(pair.Symbol)
		}
	}

	// 更新活跃列表
	s.activePairs = newActivePairs
}

// maintainOrderBook 维护某个交易对的订单簿
func (s *DynamicOrderBookSimulator) maintainOrderBook(symbol string) {
	// 首次启动时立即创建订单
	s.createOrders(symbol)

	// 订单簿更新ticker（每15秒）
	orderbookTicker := time.NewTicker(15 * time.Second)
	defer orderbookTicker.Stop()

	// 市价成交模拟ticker（每30-60秒随机）
	tradeTicker := time.NewTicker(time.Duration(30+rand.Intn(30)) * time.Second)
	defer tradeTicker.Stop()

	for {
		select {
		case <-orderbookTicker.C:
			if !s.running || !s.activePairs[symbol] {
				log.Printf("🔴 停止 %s 的订单簿模拟器", symbol)
				s.cleanupOrders(symbol)
				return
			}
			// 更新订单簿
			s.createOrders(symbol)

		case <-tradeTicker.C:
			if !s.running || !s.activePairs[symbol] {
				return
			}
			// 模拟市价成交
			s.simulateMarketTrade(symbol)
			// 重置下一次成交时间（30-60秒随机）
			tradeTicker.Reset(time.Duration(30+rand.Intn(30)) * time.Second)
		}
	}
}

// createOrders 创建模拟订单
func (s *DynamicOrderBookSimulator) createOrders(symbol string) {
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

	// 在当前价格周围创建买卖单
	s.createBuyOrders(symbol, currentPrice)
	s.createSellOrders(symbol, currentPrice)
}

// createBuyOrders 创建买单
func (s *DynamicOrderBookSimulator) createBuyOrders(symbol string, currentPrice float64) {
	// 10档买单，分布在当前价格下方 0.1% - 3%
	levels := 10

	for i := 1; i <= levels; i++ {
		priceOffset := 0.001 + float64(i)*0.003 // 0.1% 到 3.1%
		price := currentPrice * (1 - priceOffset)

		// 数量随距离增加（越远离当前价，数量越大）
		quantity := s.getQuantityForSymbol(symbol, currentPrice) * (1 + float64(i)*0.5)

		// 添加随机波动
		quantity = quantity * (0.8 + rand.Float64()*0.4)

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

// createSellOrders 创建卖单
func (s *DynamicOrderBookSimulator) createSellOrders(symbol string, currentPrice float64) {
	// 10档卖单，分布在当前价格上方 0.1% - 3%
	levels := 10

	for i := 1; i <= levels; i++ {
		priceOffset := 0.001 + float64(i)*0.003 // 0.1% 到 3.1%
		price := currentPrice * (1 + priceOffset)

		// 数量随距离增加
		quantity := s.getQuantityForSymbol(symbol, currentPrice) * (1 + float64(i)*0.5)

		// 添加随机波动
		quantity = quantity * (0.8 + rand.Float64()*0.4)

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

	log.Printf("📚 %s 订单簿已更新 (买单x%d, 卖单x%d)", symbol, levels, levels)
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

// simulateMarketTrade 模拟市价成交，让盘口更活跃
func (s *DynamicOrderBookSimulator) simulateMarketTrade(symbol string) {
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

	// 计算成交量（相对较小，不要太影响价格）
	quantity := s.getQuantityForSymbol(symbol, currentPrice) * (0.5 + rand.Float64()*1.5)

	// 计算市价单的价格（稍微偏离当前价，确保能成交）
	var price float64
	if side == "buy" {
		// 买单价格略高于当前价（吃卖单）
		price = currentPrice * (1 + 0.001 + rand.Float64()*0.005) // 0.1%-0.6%
	} else {
		// 卖单价格略低于当前价（吃买单）
		price = currentPrice * (1 - 0.001 - rand.Float64()*0.005) // -0.1%--0.6%
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
