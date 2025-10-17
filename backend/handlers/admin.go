package handlers

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/queue"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// 获取所有用户（排除虚拟用户）
func (h *AdminHandler) GetUsers(c *gin.Context) {
	var users []models.User
	database.DB.
		Where("wallet_address != ?", "0x0000000000000000000000000000000000000000").
		Order("created_at DESC").
		Find(&users)

	c.JSON(http.StatusOK, users)
}

// 获取做市商盈亏记录
func (h *AdminHandler) GetMarketMakerPnL(c *gin.Context) {
	symbol := c.Query("symbol") // 可选：按交易对筛选
	limit := 100

	query := database.DB.Model(&models.MarketMakerPnL{})

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	var records []models.MarketMakerPnL
	query.Order("created_at DESC").Limit(limit).Find(&records)

	// 计算总盈亏
	var totalPnL decimal.Decimal
	for _, record := range records {
		totalPnL = totalPnL.Add(record.ProfitLoss)
	}

	c.JSON(http.StatusOK, gin.H{
		"records":   records,
		"total_pnl": totalPnL,
		"count":     len(records),
	})
}

// 批量生成交易对初始化数据
func (h *AdminHandler) BatchGenerateInitData(c *gin.Context) {
	var req struct {
		Symbols        []string `json:"symbols"`         // 要生成的交易对列表，为空则生成所有
		StartTime      string   `json:"start_time"`      // 开始时间，格式：2024-01-01
		EndTime        string   `json:"end_time"`        // 结束时间，格式：2024-12-31
		TradeCount     int      `json:"trade_count"`     // 建议生成的交易总数（可选）
		GenerateKlines bool     `json:"generate_klines"` // 是否同时生成K线，默认true
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析时间
	var startTime, endTime time.Time
	var err error

	if req.StartTime == "" {
		// 默认：6个月前
		startTime = time.Now().AddDate(0, -6, 0)
	} else {
		startTime, err = time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use YYYY-MM-DD"})
			return
		}
	}

	if req.EndTime == "" {
		// 默认：今天
		endTime = time.Now()
	} else {
		endTime, err = time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use YYYY-MM-DD"})
			return
		}
	}

	// 获取要生成的交易对列表
	var pairs []models.TradingPair
	query := database.DB.Where("status = ?", "active")

	if len(req.Symbols) > 0 {
		query = query.Where("symbol IN ?", req.Symbols)
	}

	query.Find(&pairs)

	if len(pairs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No trading pairs found"})
		return
	}

	// 计算建议的交易数量
	days := int(endTime.Sub(startTime).Hours() / 24)
	if req.TradeCount == 0 {
		// 默认：每天2-4笔交易 × 天数
		req.TradeCount = days * 3
	}

	// 为每个交易对创建生成任务
	taskQueue := queue.GetQueue()
	createdTasks := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		// 创建交易数据生成任务
		task, err := taskQueue.AddTask(queue.TaskGenerateTrades, pair.Symbol, &startTime, &endTime)
		if err != nil {
			log.Printf("❌ 创建任务失败 (%s): %v", pair.Symbol, err)
			continue
		}
		createdTasks = append(createdTasks, task.ID)

		// 如果需要生成K线，创建K线生成任务
		if req.GenerateKlines {
			klineTask, err := taskQueue.AddTask(queue.TaskGenerateKlines, pair.Symbol, nil, nil)
			if err == nil {
				createdTasks = append(createdTasks, klineTask.ID)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                   "Batch initialization tasks created",
		"pair_count":                len(pairs),
		"task_count":                len(createdTasks),
		"task_ids":                  createdTasks,
		"estimated_days":            days,
		"suggested_trades_per_pair": req.TradeCount,
		"start_time":                startTime.Format("2006-01-02"),
		"end_time":                  endTime.Format("2006-01-02"),
	})
}

// 批量生成K线数据
func (h *AdminHandler) BatchGenerateKlines(c *gin.Context) {
	var req struct {
		Symbols []string `json:"symbols"` // 要生成的交易对列表，为空则生成所有
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取要生成的交易对列表
	var pairs []models.TradingPair
	query := database.DB.Where("status = ?", "active")

	if len(req.Symbols) > 0 {
		query = query.Where("symbol IN ?", req.Symbols)
	}

	query.Find(&pairs)

	if len(pairs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No trading pairs found"})
		return
	}

	// 为每个交易对创建K线生成任务
	taskQueue := queue.GetQueue()
	createdTasks := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		task, err := taskQueue.AddTask(queue.TaskGenerateKlines, pair.Symbol, nil, nil)
		if err != nil {
			log.Printf("❌ 创建K线任务失败 (%s): %v", pair.Symbol, err)
			continue
		}
		createdTasks = append(createdTasks, task.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Batch kline generation tasks created",
		"pair_count": len(pairs),
		"task_count": len(createdTasks),
		"task_ids":   createdTasks,
	})
}

// 批量更新交易对活跃度配置
func (h *AdminHandler) BatchUpdatePairsActivity(c *gin.Context) {
	var req struct {
		Symbols         []string `json:"symbols"`          // 要更新的交易对列表，为空则更新所有
		ActivityLevel   *int     `json:"activity_level"`   // 活跃度等级 1-10
		OrderbookDepth  *int     `json:"orderbook_depth"`  // 订单簿深度 5-30
		TradeFrequency  *int     `json:"trade_frequency"`  // 成交频率 5-60秒
		PriceVolatility *string  `json:"price_volatility"` // 价格波动率 0.001-0.05
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.ActivityLevel != nil {
		if *req.ActivityLevel < 1 || *req.ActivityLevel > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "activity_level must be between 1 and 10"})
			return
		}
		updates["activity_level"] = *req.ActivityLevel
	}
	if req.OrderbookDepth != nil {
		if *req.OrderbookDepth < 5 || *req.OrderbookDepth > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderbook_depth must be between 5 and 30"})
			return
		}
		updates["orderbook_depth"] = *req.OrderbookDepth
	}
	if req.TradeFrequency != nil {
		if *req.TradeFrequency < 5 || *req.TradeFrequency > 60 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "trade_frequency must be between 5 and 60"})
			return
		}
		updates["trade_frequency"] = *req.TradeFrequency
	}
	if req.PriceVolatility != nil {
		volatility, err := decimal.NewFromString(*req.PriceVolatility)
		if err != nil || volatility.LessThan(decimal.NewFromFloat(0.001)) || volatility.GreaterThan(decimal.NewFromFloat(0.05)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price_volatility must be between 0.001 and 0.05"})
			return
		}
		updates["price_volatility"] = volatility
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// 执行批量更新
	query := database.DB.Model(&models.TradingPair{})

	if len(req.Symbols) > 0 {
		// 更新指定的交易对
		query = query.Where("symbol IN ?", req.Symbols)
	} else {
		// 更新所有启用模拟器的交易对
		query = query.Where("simulator_enabled = ?", true)
	}

	result := query.Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Batch update successful",
		"affected_count": result.RowsAffected,
		"updates":        updates,
	})
}

// 获取做市商盈亏统计
func (h *AdminHandler) GetMarketMakerStats(c *gin.Context) {
	// 总盈亏
	var totalPnL decimal.Decimal
	database.DB.Model(&models.MarketMakerPnL{}).Select("SUM(profit_loss)").Row().Scan(&totalPnL)

	// 总交易次数
	var totalTrades int64
	database.DB.Model(&models.MarketMakerPnL{}).Count(&totalTrades)

	// 盈利次数
	var profitTrades int64
	database.DB.Model(&models.MarketMakerPnL{}).Where("profit_loss > 0").Count(&profitTrades)

	// 亏损次数
	var lossTrades int64
	database.DB.Model(&models.MarketMakerPnL{}).Where("profit_loss < 0").Count(&lossTrades)

	// 按交易对统计
	type SymbolPnL struct {
		Symbol     string          `json:"symbol"`
		TotalPnL   decimal.Decimal `json:"total_pnl"`
		TradeCount int64           `json:"trade_count"`
	}

	var symbolStats []SymbolPnL
	database.DB.Model(&models.MarketMakerPnL{}).
		Select("symbol, SUM(profit_loss) as total_pnl, COUNT(*) as trade_count").
		Group("symbol").
		Order("total_pnl DESC").
		Scan(&symbolStats)

	c.JSON(http.StatusOK, gin.H{
		"total_pnl":     totalPnL,
		"total_trades":  totalTrades,
		"profit_trades": profitTrades,
		"loss_trades":   lossTrades,
		"win_rate":      float64(profitTrades) / float64(totalTrades) * 100,
		"by_symbol":     symbolStats,
	})
}

// 获取所有订单（排除系统模拟器订单）
func (h *AdminHandler) GetAllOrders(c *gin.Context) {
	var orders []models.Order

	// 子查询：排除虚拟用户的订单
	database.DB.
		Preload("User").
		Where("user_id NOT IN (?)",
			database.DB.Table("users").Select("id").Where("wallet_address = ?", "0x0000000000000000000000000000000000000000")).
		Order("created_at DESC").
		Limit(500).
		Find(&orders)

	c.JSON(http.StatusOK, orders)
}

// 获取所有交易（排除系统模拟器交易）
func (h *AdminHandler) GetAllTrades(c *gin.Context) {
	var trades []models.Trade

	// 排除虚拟用户（系统模拟器）的订单产生的交易
	database.DB.
		Where("buy_order_id NOT IN (?) AND sell_order_id NOT IN (?)",
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", "0x0000000000000000000000000000000000000000").
				Select("orders.id"),
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", "0x0000000000000000000000000000000000000000").
				Select("orders.id")).
		Order("created_at DESC").
		Limit(500).
		Find(&trades)

	c.JSON(http.StatusOK, trades)
}

type CreateTradingPairRequest struct {
	Symbol             string  `json:"symbol" binding:"required"`
	BaseAsset          string  `json:"base_asset" binding:"required"`
	QuoteAsset         string  `json:"quote_asset" binding:"required"`
	MinPrice           string  `json:"min_price"`
	MaxPrice           string  `json:"max_price"`
	MinQty             string  `json:"min_qty"`
	MaxQty             string  `json:"max_qty"`
	ActivityLevel      *int    `json:"activity_level"`        // 活跃度等级
	OrderbookDepth     *int    `json:"orderbook_depth"`       // 订单簿深度
	TradeFrequency     *int    `json:"trade_frequency"`       // 成交频率
	PriceVolatility    *string `json:"price_volatility"`      // 价格波动率
	VirtualTradePer10s *int    `json:"virtual_trade_per_10s"` // 虚拟成交频率（每10秒N笔）
	PriceSpreadRatio   *string `json:"price_spread_ratio"`    // 盘口价格分布范围倍数
}

// 创建交易对
func (h *AdminHandler) CreateTradingPair(c *gin.Context) {
	var req CreateTradingPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	minPrice, _ := decimal.NewFromString(req.MinPrice)
	maxPrice, _ := decimal.NewFromString(req.MaxPrice)
	minQty, _ := decimal.NewFromString(req.MinQty)
	maxQty, _ := decimal.NewFromString(req.MaxQty)

	pair := models.TradingPair{
		Symbol:     req.Symbol,
		BaseAsset:  req.BaseAsset,
		QuoteAsset: req.QuoteAsset,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		MinQty:     minQty,
		MaxQty:     maxQty,
		Status:     "active",
	}

	if err := database.DB.Create(&pair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trading pair"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// 更新交易对状态
func (h *AdminHandler) UpdateTradingPairStatus(c *gin.Context) {
	pairID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pair models.TradingPair
	if err := database.DB.Where("id = ?", pairID).First(&pair).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	pair.Status = req.Status
	database.DB.Save(&pair)

	c.JSON(http.StatusOK, pair)
}

// 更新交易对模拟器状态
func (h *AdminHandler) UpdateTradingPairSimulator(c *gin.Context) {
	pairID := c.Param("id")

	var req struct {
		SimulatorEnabled bool `json:"simulator_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pair models.TradingPair
	if err := database.DB.Where("id = ?", pairID).First(&pair).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	pair.SimulatorEnabled = req.SimulatorEnabled
	database.DB.Save(&pair)

	c.JSON(http.StatusOK, pair)
}

// 更新交易对信息
func (h *AdminHandler) UpdateTradingPair(c *gin.Context) {
	pairID := c.Param("id")

	var req CreateTradingPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pair models.TradingPair
	if err := database.DB.Where("id = ?", pairID).First(&pair).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	// 更新字段
	pair.Symbol = req.Symbol
	pair.BaseAsset = req.BaseAsset
	pair.QuoteAsset = req.QuoteAsset

	if req.MinPrice != "" {
		minPrice, _ := decimal.NewFromString(req.MinPrice)
		pair.MinPrice = minPrice
	}
	if req.MaxPrice != "" {
		maxPrice, _ := decimal.NewFromString(req.MaxPrice)
		pair.MaxPrice = maxPrice
	}
	if req.MinQty != "" {
		minQty, _ := decimal.NewFromString(req.MinQty)
		pair.MinQty = minQty
	}
	if req.MaxQty != "" {
		maxQty, _ := decimal.NewFromString(req.MaxQty)
		pair.MaxQty = maxQty
	}

	// 更新活跃度配置
	if req.ActivityLevel != nil {
		pair.ActivityLevel = *req.ActivityLevel
	}
	if req.OrderbookDepth != nil {
		pair.OrderbookDepth = *req.OrderbookDepth
	}
	if req.TradeFrequency != nil {
		pair.TradeFrequency = *req.TradeFrequency
	}
	if req.PriceVolatility != nil {
		volatility, _ := decimal.NewFromString(*req.PriceVolatility)
		pair.PriceVolatility = volatility
	}
	if req.VirtualTradePer10s != nil {
		pair.VirtualTradePer10s = *req.VirtualTradePer10s
	}
	if req.PriceSpreadRatio != nil {
		spreadRatio, _ := decimal.NewFromString(*req.PriceSpreadRatio)
		pair.PriceSpreadRatio = spreadRatio
	}

	if err := database.DB.Save(&pair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trading pair"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// 获取统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	// 排除虚拟模拟用户（钱包地址为全0的用户）
	simulatorWallet := "0x0000000000000000000000000000000000000000"

	// 统计真实用户数（排除模拟器用户）
	var userCount int64
	database.DB.Model(&models.User{}).
		Where("wallet_address != ?", simulatorWallet).
		Count(&userCount)

	// 统计真实订单数（排除模拟器用户的订单）
	var orderCount int64
	database.DB.Model(&models.Order{}).
		Where("user_id NOT IN (?)",
			database.DB.Table("users").Select("id").Where("wallet_address = ?", simulatorWallet)).
		Count(&orderCount)

	// 统计真实交易数（排除模拟器用户参与的交易）
	var tradeCount int64
	database.DB.Model(&models.Trade{}).
		Where("buy_order_id NOT IN (?) AND sell_order_id NOT IN (?)",
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", simulatorWallet).
				Select("orders.id"),
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", simulatorWallet).
				Select("orders.id")).
		Count(&tradeCount)

	// 统计真实交易量（排除模拟器交易）
	var totalVolume decimal.Decimal
	var trades []models.Trade
	database.DB.
		Where("buy_order_id NOT IN (?) AND sell_order_id NOT IN (?)",
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", simulatorWallet).
				Select("orders.id"),
			database.DB.Table("orders").
				Joins("JOIN users ON users.id = orders.user_id").
				Where("users.wallet_address = ?", simulatorWallet).
				Select("orders.id")).
		Find(&trades)

	for _, trade := range trades {
		totalVolume = totalVolume.Add(trade.Price.Mul(trade.Quantity))
	}

	c.JSON(http.StatusOK, gin.H{
		"user_count":   userCount,
		"order_count":  orderCount,
		"trade_count":  tradeCount,
		"total_volume": totalVolume,
	})
}

// 获取所有交易对
func (h *AdminHandler) GetTradingPairs(c *gin.Context) {
	var pairs []models.TradingPair
	database.DB.Order("created_at DESC").Find(&pairs)

	c.JSON(http.StatusOK, pairs)
}

// 为单个交易对生成历史交易数据
func (h *AdminHandler) GenerateTradeDataForPair(c *gin.Context) {
	var req struct {
		Symbol    string `json:"symbol" binding:"required"`
		StartTime string `json:"start_time"` // 可选，格式：2024-01-01
		EndTime   string `json:"end_time"`   // 可选，格式：2024-12-31
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析时间
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		t, err := time.Parse("2006-01-02", req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format, use YYYY-MM-DD"})
			return
		}
		startTime = &t
	}
	if req.EndTime != "" {
		t, err := time.Parse("2006-01-02", req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format, use YYYY-MM-DD"})
			return
		}
		endTime = &t
	}

	taskQueue := queue.GetQueue()
	task, err := taskQueue.AddTask(queue.TaskGenerateTrades, req.Symbol, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trade data generation task added to queue",
		"task_id": task.ID,
		"symbol":  task.Symbol,
		"status":  task.Status,
	})
}

// 为单个交易对生成K线数据
func (h *AdminHandler) GenerateKlineDataForPair(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskQueue := queue.GetQueue()
	task, err := taskQueue.AddTask(queue.TaskGenerateKlines, req.Symbol, nil, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kline data generation task added to queue",
		"task_id": task.ID,
		"symbol":  task.Symbol,
		"status":  task.Status,
	})
}

// 获取任务状态
func (h *AdminHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	taskQueue := queue.GetQueue()

	task, exists := taskQueue.GetTask(taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// 获取所有任务
func (h *AdminHandler) GetAllTasks(c *gin.Context) {
	taskQueue := queue.GetQueue()
	tasks := taskQueue.GetAllTasks()

	c.JSON(http.StatusOK, tasks)
}

// 获取当前运行的任务
func (h *AdminHandler) GetRunningTask(c *gin.Context) {
	taskQueue := queue.GetQueue()
	task := taskQueue.GetRunningTask()

	if task == nil {
		c.JSON(http.StatusOK, gin.H{"running": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"running": true,
		"task":    task,
	})
}

// 获取任务日志
func (h *AdminHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")

	var logs []models.TaskLog
	database.DB.Where("task_id = ?", taskID).
		Order("created_at ASC").
		Find(&logs)

	c.JSON(http.StatusOK, logs)
}

// 重试任务
func (h *AdminHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	taskQueue := queue.GetQueue()

	if err := taskQueue.RetryTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task has been retried successfully"})
}

// 获取所有充值记录
func (h *AdminHandler) GetAllDeposits(c *gin.Context) {
	var deposits []models.DepositRecord
	database.DB.Preload("User").Order("created_at DESC").Limit(500).Find(&deposits)

	c.JSON(http.StatusOK, deposits)
}

// 获取所有提现记录
func (h *AdminHandler) GetAllWithdrawals(c *gin.Context) {
	var withdrawals []models.WithdrawRecord
	database.DB.Preload("User").Order("created_at DESC").Limit(500).Find(&withdrawals)

	c.JSON(http.StatusOK, withdrawals)
}

// 获取所有系统配置
func (h *AdminHandler) GetSystemConfigs(c *gin.Context) {
	category := c.Query("category")

	query := database.DB.Order("category, `key`")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var configs []models.SystemConfig
	query.Find(&configs)

	c.JSON(http.StatusOK, configs)
}

// 获取单个系统配置
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	configID := c.Param("id")

	var config models.SystemConfig
	if err := database.DB.Where("id = ?", configID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 更新系统配置
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	configID := c.Param("id")

	var req struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sysConfig models.SystemConfig
	if err := database.DB.Where("id = ?", configID).First(&sysConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	oldValue := sysConfig.Value
	sysConfig.Value = req.Value
	database.DB.Save(&sysConfig)

	// 热更新到内存配置
	configManager := database.GetSystemConfigManager()
	configManager.Set(sysConfig.Key, sysConfig.Value)

	log.Printf("🔄 系统配置已更新: %s = %s (旧值: %s)", sysConfig.Key, sysConfig.Value, oldValue)

	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated and reloaded",
		"config":  sysConfig,
	})
}

// 重新加载所有系统配置
func (h *AdminHandler) ReloadSystemConfigs(c *gin.Context) {
	configManager := database.GetSystemConfigManager()
	configManager.Reload()

	c.JSON(http.StatusOK, gin.H{
		"message": "System configs reloaded successfully",
	})
}
