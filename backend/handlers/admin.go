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

// 获取所有用户
func (h *AdminHandler) GetUsers(c *gin.Context) {
	var users []models.User
	database.DB.Order("created_at DESC").Find(&users)

	c.JSON(http.StatusOK, users)
}

// 获取所有订单
func (h *AdminHandler) GetAllOrders(c *gin.Context) {
	var orders []models.Order
	database.DB.Preload("User").Order("created_at DESC").Limit(500).Find(&orders)

	c.JSON(http.StatusOK, orders)
}

// 获取所有交易
func (h *AdminHandler) GetAllTrades(c *gin.Context) {
	var trades []models.Trade
	database.DB.Order("created_at DESC").Limit(500).Find(&trades)

	c.JSON(http.StatusOK, trades)
}

type CreateTradingPairRequest struct {
	Symbol     string `json:"symbol" binding:"required"`
	BaseAsset  string `json:"base_asset" binding:"required"`
	QuoteAsset string `json:"quote_asset" binding:"required"`
	MinPrice   string `json:"min_price"`
	MaxPrice   string `json:"max_price"`
	MinQty     string `json:"min_qty"`
	MaxQty     string `json:"max_qty"`
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

	if err := database.DB.Save(&pair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trading pair"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

// 获取统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	var userCount int64
	var orderCount int64
	var tradeCount int64

	database.DB.Model(&models.User{}).Count(&userCount)
	database.DB.Model(&models.Order{}).Count(&orderCount)
	database.DB.Model(&models.Trade{}).Count(&tradeCount)

	var totalVolume decimal.Decimal
	var trades []models.Trade
	database.DB.Find(&trades)
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

	query := database.DB.Order("category, key")
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
