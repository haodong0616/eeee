package matching

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/services"
	"log"
	"sync"
)

type Manager struct {
	engines    map[string]*Engine
	mu         sync.RWMutex
	tradeChan  chan *models.Trade
	feeService *services.FeeService
}

func NewManager() *Manager {
	tradeChan := make(chan *models.Trade, 1000)
	m := &Manager{
		engines:    make(map[string]*Engine),
		tradeChan:  tradeChan,
		feeService: services.NewFeeService(),
	}

	// 初始化手续费配置
	m.feeService.InitDefaultFeeConfig()

	// 启动成交处理协程
	go m.processTrades()

	return m
}

func (m *Manager) GetEngine(symbol string) *Engine {
	m.mu.RLock()
	engine, exists := m.engines[symbol]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		engine = NewEngine(symbol, m.tradeChan)
		m.engines[symbol] = engine
		m.mu.Unlock()
	}

	return engine
}

func (m *Manager) AddOrder(order *models.Order) {
	engine := m.GetEngine(order.Symbol)
	engine.AddOrder(order)
}

func (m *Manager) CancelOrder(orderID uint, symbol, side string) bool {
	engine := m.GetEngine(symbol)
	return engine.CancelOrder(orderID, side)
}

func (m *Manager) GetOrderBook(symbol string, depth int) *models.OrderBook {
	engine := m.GetEngine(symbol)
	return engine.GetOrderBook(depth)
}

func (m *Manager) processTrades() {
	for trade := range m.tradeChan {
		// 保存成交记录到数据库
		if err := database.DB.Create(trade).Error; err != nil {
			log.Printf("Failed to save trade: %v", err)
			continue
		}

		// 更新订单状态
		var buyOrder, sellOrder models.Order
		database.DB.First(&buyOrder, trade.BuyOrderID)
		database.DB.First(&sellOrder, trade.SellOrderID)

		if buyOrder.FilledQty.Equal(buyOrder.Quantity) {
			buyOrder.Status = "filled"
		} else {
			buyOrder.Status = "partial"
		}

		if sellOrder.FilledQty.Equal(sellOrder.Quantity) {
			sellOrder.Status = "filled"
		} else {
			sellOrder.Status = "partial"
		}

		database.DB.Save(&buyOrder)
		database.DB.Save(&sellOrder)

		// 更新用户余额
		m.updateBalances(&buyOrder, &sellOrder, trade)

		log.Printf("Trade executed: %s %s @ %s", trade.Symbol, trade.Quantity, trade.Price)
	}
}

func (m *Manager) updateBalances(buyOrder, sellOrder *models.Order, trade *models.Trade) {
	cost := trade.Price.Mul(trade.Quantity)
	baseAsset := getBaseAsset(buyOrder.Symbol)
	quoteAsset := getQuoteAsset(buyOrder.Symbol)

	// 获取用户等级
	var buyer, seller models.User
	database.DB.First(&buyer, buyOrder.UserID)
	database.DB.First(&seller, sellOrder.UserID)

	// 判断谁是Maker，谁是Taker（简化处理：先创建的是Maker）
	buyerIsMaker := buyOrder.CreatedAt.Before(sellOrder.CreatedAt)

	// 计算买方手续费（从获得的base资产中扣除）
	buyerFee, buyerFeeRate, _ := m.feeService.CalculateFee(buyer.UserLevel, buyerIsMaker, trade.Quantity)
	buyerReceiveAmount := trade.Quantity.Sub(buyerFee)

	// 计算卖方手续费（从获得的quote资产中扣除）
	sellerFee, sellerFeeRate, _ := m.feeService.CalculateFee(seller.UserLevel, !buyerIsMaker, cost)
	sellerReceiveAmount := cost.Sub(sellerFee)

	// 买方：减少quote资产frozen，增加base资产available（扣除手续费）
	var buyerQuoteBalance models.Balance
	database.DB.Where("user_id = ? AND asset = ?", buyOrder.UserID, quoteAsset).First(&buyerQuoteBalance)
	buyerQuoteBalance.Frozen = buyerQuoteBalance.Frozen.Sub(cost)
	database.DB.Save(&buyerQuoteBalance)

	var buyerBaseBalance models.Balance
	database.DB.Where("user_id = ? AND asset = ?", buyOrder.UserID, baseAsset).FirstOrCreate(&buyerBaseBalance, models.Balance{
		UserID: buyOrder.UserID,
		Asset:  baseAsset,
	})
	buyerBaseBalance.Available = buyerBaseBalance.Available.Add(buyerReceiveAmount)
	database.DB.Save(&buyerBaseBalance)

	// 卖方：减少base资产frozen，增加quote资产available（扣除手续费）
	var sellerBaseBalance models.Balance
	database.DB.Where("user_id = ? AND asset = ?", sellOrder.UserID, baseAsset).First(&sellerBaseBalance)
	sellerBaseBalance.Frozen = sellerBaseBalance.Frozen.Sub(trade.Quantity)
	database.DB.Save(&sellerBaseBalance)

	var sellerQuoteBalance models.Balance
	database.DB.Where("user_id = ? AND asset = ?", sellOrder.UserID, quoteAsset).FirstOrCreate(&sellerQuoteBalance, models.Balance{
		UserID: sellOrder.UserID,
		Asset:  quoteAsset,
	})
	sellerQuoteBalance.Available = sellerQuoteBalance.Available.Add(sellerReceiveAmount)
	database.DB.Save(&sellerQuoteBalance)

	// 记录手续费
	buyerOrderSide := "maker"
	if !buyerIsMaker {
		buyerOrderSide = "taker"
	}
	m.feeService.RecordFee(buyOrder.UserID, buyOrder.ID, trade.ID, baseAsset, buyerFee, buyerFeeRate, buyerOrderSide)

	sellerOrderSide := "maker"
	if buyerIsMaker {
		sellerOrderSide = "taker"
	}
	m.feeService.RecordFee(sellOrder.UserID, sellOrder.ID, trade.ID, quoteAsset, sellerFee, sellerFeeRate, sellerOrderSide)

	log.Printf("Trade fees - Buyer: %s %s (%s), Seller: %s %s (%s)",
		buyerFee, baseAsset, buyerOrderSide, sellerFee, quoteAsset, sellerOrderSide)
}

func getBaseAsset(symbol string) string {
	// 简单解析，实际应该从TradingPair表查询
	// symbol format: BTC/USDT
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			return symbol[:i]
		}
	}
	return ""
}

func getQuoteAsset(symbol string) string {
	// symbol format: BTC/USDT
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			return symbol[i+1:]
		}
	}
	return ""
}
