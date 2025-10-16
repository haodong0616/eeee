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

type OrderBookSimulator struct {
	matchingManager *matching.Manager
	running         bool
	symbols         []string
	virtualUserID   uint
}

func NewOrderBookSimulator(matchingManager *matching.Manager) *OrderBookSimulator {
	// 创建虚拟用户用于挂单
	var virtualUser models.User
	result := database.DB.Where("wallet_address = ?", "0x0000000000000000000000000000000000000000").First(&virtualUser)

	if result.Error != nil {
		virtualUser = models.User{
			WalletAddress: "0x0000000000000000000000000000000000000000",
			Nonce:         "virtual",
			UserLevel:     "normal",
		}
		database.DB.Create(&virtualUser)

		// 为虚拟用户充值大量资金
		assets := []string{
			"TITAN", "GENESIS", "LUNAR", "ORACLE", "QUANTUM", "NOVA",
			"ATLAS", "COSMOS", "NEXUS", "VERTEX", "AURORA", "ZEPHYR",
			"PRISM", "PULSE", "ARCANA", "USDT",
		}
		for _, asset := range assets {
			balance := models.Balance{
				UserID:    virtualUser.ID,
				Asset:     asset,
				Available: decimal.NewFromFloat(10000000),
				Frozen:    decimal.Zero,
			}
			database.DB.Create(&balance)
		}
		log.Println("✅ 创建虚拟用户并充值")
	}

	return &OrderBookSimulator{
		matchingManager: matchingManager,
		running:         false,
		symbols: []string{
			"TITAN/USDT", "GENESIS/USDT", "LUNAR/USDT",
			"ORACLE/USDT", "QUANTUM/USDT", "NOVA/USDT",
			"ATLAS/USDT", "COSMOS/USDT", "NEXUS/USDT",
			"VERTEX/USDT", "AURORA/USDT", "ZEPHYR/USDT",
			"PRISM/USDT", "PULSE/USDT", "ARCANA/USDT",
		},
		virtualUserID: virtualUser.ID,
	}
}

func (s *OrderBookSimulator) Start() {
	if s.running {
		return
	}

	s.running = true
	log.Println("📚 订单簿模拟器已启动")

	// 为每个交易对启动订单簿管理
	for _, symbol := range s.symbols {
		go s.maintainOrderBook(symbol)
	}
}

func (s *OrderBookSimulator) Stop() {
	s.running = false
	log.Println("🛑 订单簿模拟器已停止")
}

func (s *OrderBookSimulator) maintainOrderBook(symbol string) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查并维护订单簿
	defer ticker.Stop()

	// 首次启动时立即创建订单
	s.createOrders(symbol)

	for range ticker.C {
		if !s.running {
			return
		}

		s.createOrders(symbol)
	}
}

func (s *OrderBookSimulator) createOrders(symbol string) {
	// 获取当前价格（从最近成交记录）
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		// 如果没有成交记录，使用默认价格
		return
	}

	currentPrice, _ := lastTrade.Price.Float64()

	// 清理虚拟用户的旧挂单（避免订单堆积）
	database.DB.Where("user_id = ? AND symbol = ? AND status IN (?, ?)",
		s.virtualUserID, symbol, "pending", "partial").
		Delete(&models.Order{})

	// 在当前价格周围挂买卖单
	s.createBuyOrders(symbol, currentPrice)
	s.createSellOrders(symbol, currentPrice)
}

func (s *OrderBookSimulator) createBuyOrders(symbol string, currentPrice float64) {
	// 在当前价格下方0.5%-3%范围内挂买单
	levels := 10 // 10档买单

	for i := 1; i <= levels; i++ {
		// 价格递减
		priceOffset := (0.005 + float64(i)*0.003) // 0.5% 到 3.5%
		price := currentPrice * (1 - priceOffset)

		// 数量随机，越远离当前价格数量越大
		var baseQty float64
		switch symbol {
		case "BTC/USDT":
			baseQty = 0.01 + float64(i)*0.02
		case "ETH/USDT":
			baseQty = 0.1 + float64(i)*0.3
		case "BNB/USDT":
			baseQty = 1 + float64(i)*5
		case "SOL/USDT":
			baseQty = 5 + float64(i)*10
		case "XRP/USDT":
			baseQty = 100 + float64(i)*200
		}

		quantity := baseQty * (0.8 + rand.Float64()*0.4) // 随机波动 ±20%

		// 创建买单
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

		// 提交到撮合引擎
		s.matchingManager.AddOrder(&order)
	}
}

func (s *OrderBookSimulator) createSellOrders(symbol string, currentPrice float64) {
	// 在当前价格上方0.5%-3%范围内挂卖单
	levels := 10 // 10档卖单

	for i := 1; i <= levels; i++ {
		// 价格递增
		priceOffset := (0.005 + float64(i)*0.003) // 0.5% 到 3.5%
		price := currentPrice * (1 + priceOffset)

		// 数量随机，越远离当前价格数量越大
		var baseQty float64
		switch symbol {
		case "BTC/USDT":
			baseQty = 0.01 + float64(i)*0.02
		case "ETH/USDT":
			baseQty = 0.1 + float64(i)*0.3
		case "BNB/USDT":
			baseQty = 1 + float64(i)*5
		case "SOL/USDT":
			baseQty = 5 + float64(i)*10
		case "XRP/USDT":
			baseQty = 100 + float64(i)*200
		}

		quantity := baseQty * (0.8 + rand.Float64()*0.4)

		// 创建卖单
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

		// 提交到撮合引擎
		s.matchingManager.AddOrder(&order)
	}

	log.Printf("📚 %s 订单簿已更新 (买单x10, 卖单x10)", symbol)
}

// 高级订单簿模拟器 - 更自然的订单分布
type AdvancedOrderBookSimulator struct {
	matchingManager *matching.Manager
	running         bool
	symbols         []string
	virtualUserID   uint
}

func NewAdvancedOrderBookSimulator(matchingManager *matching.Manager, virtualUserID uint) *AdvancedOrderBookSimulator {
	return &AdvancedOrderBookSimulator{
		matchingManager: matchingManager,
		running:         false,
		symbols:         []string{"LUNAR/USDT", "NOVA/USDT", "ZEPHYR/USDT", "PULSE/USDT", "ARCANA/USDT", "NEXUS/USDT"},
		virtualUserID:   virtualUserID,
	}
}

func (s *AdvancedOrderBookSimulator) Start() {
	if s.running {
		return
	}

	s.running = true
	log.Println("🎯 高级订单簿模拟器已启动")

	for _, symbol := range s.symbols {
		go s.maintainAdvancedOrderBook(symbol)
	}
}

func (s *AdvancedOrderBookSimulator) Stop() {
	s.running = false
}

func (s *AdvancedOrderBookSimulator) maintainAdvancedOrderBook(symbol string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	s.createAdvancedOrders(symbol)

	for range ticker.C {
		if !s.running {
			return
		}

		// 随机取消一些订单，创建新的
		if rand.Float64() < 0.3 { // 30%概率刷新订单簿
			s.refreshOrderBook(symbol)
		}
	}
}

func (s *AdvancedOrderBookSimulator) createAdvancedOrders(symbol string) {
	var lastTrade models.Trade
	result := database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&lastTrade)

	if result.Error != nil {
		return
	}

	currentPrice, _ := lastTrade.Price.Float64()

	// 清理旧订单
	database.DB.Where("user_id = ? AND symbol = ? AND status = ?",
		s.virtualUserID, symbol, "pending").
		Delete(&models.Order{})

	// 创建更自然的订单分布
	// 买单：集中在当前价下方0.1%-1%
	// 卖单：集中在当前价上方0.1%-1%

	buyLevels := []float64{0.001, 0.002, 0.003, 0.005, 0.007, 0.01, 0.015, 0.02, 0.025, 0.03}
	sellLevels := []float64{0.001, 0.002, 0.003, 0.005, 0.007, 0.01, 0.015, 0.02, 0.025, 0.03}

	for _, offset := range buyLevels {
		price := currentPrice * (1 - offset)
		qty := s.getQuantityForLevel(symbol, offset)

		order := models.Order{
			UserID:    s.virtualUserID,
			Symbol:    symbol,
			OrderType: "limit",
			Side:      "buy",
			Price:     decimal.NewFromFloat(price),
			Quantity:  decimal.NewFromFloat(qty),
			FilledQty: decimal.Zero,
			Status:    "pending",
		}
		database.DB.Create(&order)
		s.matchingManager.AddOrder(&order)
	}

	for _, offset := range sellLevels {
		price := currentPrice * (1 + offset)
		qty := s.getQuantityForLevel(symbol, offset)

		order := models.Order{
			UserID:    s.virtualUserID,
			Symbol:    symbol,
			OrderType: "limit",
			Side:      "sell",
			Price:     decimal.NewFromFloat(price),
			Quantity:  decimal.NewFromFloat(qty),
			FilledQty: decimal.Zero,
			Status:    "pending",
		}
		database.DB.Create(&order)
		s.matchingManager.AddOrder(&order)
	}
}

func (s *AdvancedOrderBookSimulator) getQuantityForLevel(symbol string, priceOffset float64) float64 {
	// 距离当前价格越远，数量越大（更真实的订单簿）
	multiplier := 1.0 + priceOffset*50

	var baseQty float64
	switch symbol {
	case "LUNAR/USDT":
		baseQty = 0.1 * multiplier
	case "NOVA/USDT":
		baseQty = 1.5 * multiplier
	case "ZEPHYR/USDT":
		baseQty = 15 * multiplier
	case "PULSE/USDT":
		baseQty = 150 * multiplier
	case "ARCANA/USDT":
		baseQty = 1500 * multiplier
	case "NEXUS/USDT":
		baseQty = 10 * multiplier
	}

	return baseQty * (0.7 + rand.Float64()*0.6)
}

func (s *AdvancedOrderBookSimulator) refreshOrderBook(symbol string) {
	// 随机取消一些订单
	var orders []models.Order
	database.DB.Where("user_id = ? AND symbol = ? AND status = ?",
		s.virtualUserID, symbol, "pending").
		Limit(5).
		Find(&orders)

	for _, order := range orders {
		if rand.Float64() < 0.5 { // 50%概率取消
			order.Status = "cancelled"
			database.DB.Save(&order)
			s.matchingManager.CancelOrder(order.ID, order.Symbol, order.Side)
		}
	}

	// 创建新订单补充
	s.createAdvancedOrders(symbol)
}
