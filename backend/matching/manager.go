package matching

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/services"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Manager struct {
	engines    map[string]*Engine
	mu         sync.RWMutex
	tradeChan  chan *models.Trade
	feeService *services.FeeService
}

func NewManager() *Manager {
	// 增大缓冲区，提高吞吐量
	tradeChan := make(chan *models.Trade, 10000)
	m := &Manager{
		engines:    make(map[string]*Engine),
		tradeChan:  tradeChan,
		feeService: services.NewFeeService(),
	}

	// 初始化手续费配置
	m.feeService.InitDefaultFeeConfig()

	// 启动批量成交处理协程（提升性能）
	go m.processTradesBatch()

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

func (m *Manager) CancelOrder(orderID string, symbol, side string) bool {
	engine := m.GetEngine(symbol)
	return engine.CancelOrder(orderID, side)
}

func (m *Manager) GetOrderBook(symbol string, depth int) *models.OrderBook {
	engine := m.GetEngine(symbol)
	return engine.GetOrderBook(depth)
}

// processTradesBatch 批量处理成交（性能优化版）
func (m *Manager) processTradesBatch() {
	batch := make([]*models.Trade, 0, 100)
	ticker := time.NewTicker(10 * time.Millisecond) // 每10ms或达到100条就处理一批
	defer ticker.Stop()

	for {
		select {
		case trade := <-m.tradeChan:
			batch = append(batch, trade)
			// 达到批量大小立即处理
			if len(batch) >= 100 {
				m.processBatch(batch)
				batch = batch[:0] // 清空batch
			}
		case <-ticker.C:
			// 定时处理剩余的成交
			if len(batch) > 0 {
				m.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// processBatch 批量处理一批成交
func (m *Manager) processBatch(trades []*models.Trade) {
	if len(trades) == 0 {
		return
	}

	// ⚠️ 过滤掉做市商手动创建的Trade（已经手动更新过余额了）
	// 做市商Trade的特征：BuyOrderID 或 SellOrderID 属于虚拟用户（通过wallet_address识别）
	realTrades := make([]*models.Trade, 0, len(trades))
	for _, trade := range trades {
		// 查询订单所属用户
		var buyOrder, sellOrder models.Order
		buyExists := database.DB.Where("id = ?", trade.BuyOrderID).First(&buyOrder).Error == nil
		sellExists := database.DB.Where("id = ?", trade.SellOrderID).First(&sellOrder).Error == nil

		// 查询用户的wallet_address来判断是否是虚拟用户
		isVirtualTrade := false
		if buyExists {
			var buyUser models.User
			if database.DB.Where("id = ?", buyOrder.UserID).First(&buyUser).Error == nil {
				if buyUser.WalletAddress == "0x0000000000000000000000000000000000000000" {
					isVirtualTrade = true
				}
			}
		}
		if sellExists && !isVirtualTrade {
			var sellUser models.User
			if database.DB.Where("id = ?", sellOrder.UserID).First(&sellUser).Error == nil {
				if sellUser.WalletAddress == "0x0000000000000000000000000000000000000000" {
					isVirtualTrade = true
				}
			}
		}

		if isVirtualTrade {
			log.Printf("⏭️ 跳过做市商Trade（已手动处理余额）: TradeID=%s", trade.ID)
			continue
		}

		realTrades = append(realTrades, trade)
	}

	if len(realTrades) == 0 {
		return // 没有需要处理的真实Trade
	}

	// 使用事务批量处理
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 批量保存成交记录
		if err := tx.CreateInBatches(realTrades, 100).Error; err != nil {
			return err
		}

		// 2. 收集所有涉及的订单ID
		orderIDs := make(map[string]bool)
		for _, trade := range realTrades {
			orderIDs[trade.BuyOrderID] = true
			orderIDs[trade.SellOrderID] = true
		}

		// 3. 批量查询订单
		ids := make([]string, 0, len(orderIDs))
		for id := range orderIDs {
			ids = append(ids, id)
		}

		var orders []models.Order
		if err := tx.Where("id IN ?", ids).Find(&orders).Error; err != nil {
			return err
		}

		// 4. 构建订单Map
		orderMap := make(map[string]*models.Order)
		for i := range orders {
			orderMap[orders[i].ID] = &orders[i]
		}

		// 5. 处理每个成交
		for _, trade := range realTrades {
			buyOrder := orderMap[trade.BuyOrderID]
			sellOrder := orderMap[trade.SellOrderID]

			if buyOrder == nil || sellOrder == nil {
				continue
			}

			// 更新订单状态
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

			// 更新用户余额（在事务中）
			m.updateBalancesInTx(tx, buyOrder, sellOrder, trade)
		}

		// 6. 批量更新订单
		for _, order := range orders {
			if err := tx.Save(order).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Failed to process trade batch: %v", err)
	} else {
		if len(realTrades) > 0 {
			log.Printf("✅ Processed %d trades in batch", len(realTrades))
		}
	}
}

// updateBalancesInTx 在事务中更新用户余额（性能优化版）
func (m *Manager) updateBalancesInTx(tx *gorm.DB, buyOrder, sellOrder *models.Order, trade *models.Trade) {
	cost := trade.Price.Mul(trade.Quantity)
	baseAsset := getBaseAsset(buyOrder.Symbol)
	quoteAsset := getQuoteAsset(buyOrder.Symbol)

	// 获取用户等级
	var buyer, seller models.User
	tx.Where("id = ?", buyOrder.UserID).First(&buyer)
	tx.Where("id = ?", sellOrder.UserID).First(&seller)

	// 判断谁是Maker，谁是Taker
	buyerIsMaker := buyOrder.CreatedAt.Before(sellOrder.CreatedAt)

	// 计算买方手续费
	buyerFee, buyerFeeRate, _ := m.feeService.CalculateFee(buyer.UserLevel, buyerIsMaker, trade.Quantity)
	buyerReceiveAmount := trade.Quantity.Sub(buyerFee)

	// 计算卖方手续费
	sellerFee, sellerFeeRate, _ := m.feeService.CalculateFee(seller.UserLevel, !buyerIsMaker, cost)
	sellerReceiveAmount := cost.Sub(sellerFee)

	// 更新买方余额
	var buyerQuoteBalance models.Balance
	tx.Where("user_id = ? AND asset = ?", buyOrder.UserID, quoteAsset).First(&buyerQuoteBalance)
	buyerQuoteBalance.Frozen = buyerQuoteBalance.Frozen.Sub(cost)
	tx.Save(&buyerQuoteBalance)

	var buyerBaseBalance models.Balance
	tx.Where("user_id = ? AND asset = ?", buyOrder.UserID, baseAsset).FirstOrCreate(&buyerBaseBalance, models.Balance{
		UserID: buyOrder.UserID,
		Asset:  baseAsset,
	})
	buyerBaseBalance.Available = buyerBaseBalance.Available.Add(buyerReceiveAmount)
	tx.Save(&buyerBaseBalance)

	// 更新卖方余额
	var sellerBaseBalance models.Balance
	tx.Where("user_id = ? AND asset = ?", sellOrder.UserID, baseAsset).First(&sellerBaseBalance)
	sellerBaseBalance.Frozen = sellerBaseBalance.Frozen.Sub(trade.Quantity)
	tx.Save(&sellerBaseBalance)

	var sellerQuoteBalance models.Balance
	tx.Where("user_id = ? AND asset = ?", sellOrder.UserID, quoteAsset).FirstOrCreate(&sellerQuoteBalance, models.Balance{
		UserID: sellOrder.UserID,
		Asset:  quoteAsset,
	})
	sellerQuoteBalance.Available = sellerQuoteBalance.Available.Add(sellerReceiveAmount)
	tx.Save(&sellerQuoteBalance)

	// 记录手续费
	buyerOrderSide := "maker"
	if !buyerIsMaker {
		buyerOrderSide = "taker"
	}
	m.feeService.RecordFeeInTx(tx, buyOrder.UserID, buyOrder.ID, trade.ID, baseAsset, buyerFee, buyerFeeRate, buyerOrderSide)

	sellerOrderSide := "maker"
	if buyerIsMaker {
		sellerOrderSide = "taker"
	}
	m.feeService.RecordFeeInTx(tx, sellOrder.UserID, sellOrder.ID, trade.ID, quoteAsset, sellerFee, sellerFeeRate, sellerOrderSide)
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
