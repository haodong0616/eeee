package handlers

import (
	"expchange-backend/database"
	"expchange-backend/matching"
	"expchange-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type OrderHandler struct {
	matchingManager *matching.Manager
}

func NewOrderHandler(matchingManager *matching.Manager) *OrderHandler {
	return &OrderHandler{
		matchingManager: matchingManager,
	}
}

type CreateOrderRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	OrderType string `json:"order_type" binding:"required"` // limit, market
	Side      string `json:"side" binding:"required"`       // buy, sell
	Price     string `json:"price"`
	Quantity  string `json:"quantity" binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证交易对
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ? AND status = ?", req.Symbol, "active").First(&pair).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trading pair"})
		return
	}

	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
		return
	}

	var price decimal.Decimal
	if req.OrderType == "limit" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price"})
			return
		}
	}

	// 检查余额并冻结资产
	if req.Side == "buy" {
		// 买单需要冻结报价资产
		quoteAsset := getQuoteAsset(req.Symbol)
		requiredAmount := price.Mul(quantity)

		var balance models.Balance
		if err := database.DB.Where("user_id = ? AND asset = ?", userID, quoteAsset).First(&balance).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}

		if balance.Available.LessThan(requiredAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}

		balance.Available = balance.Available.Sub(requiredAmount)
		balance.Frozen = balance.Frozen.Add(requiredAmount)
		database.DB.Save(&balance)

	} else {
		// 卖单需要冻结基础资产
		baseAsset := getBaseAsset(req.Symbol)

		var balance models.Balance
		if err := database.DB.Where("user_id = ? AND asset = ?", userID, baseAsset).First(&balance).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}

		if balance.Available.LessThan(quantity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return
		}

		balance.Available = balance.Available.Sub(quantity)
		balance.Frozen = balance.Frozen.Add(quantity)
		database.DB.Save(&balance)
	}

	// 创建订单
	order := models.Order{
		UserID:    userID,
		Symbol:    req.Symbol,
		OrderType: req.OrderType,
		Side:      req.Side,
		Price:     price,
		Quantity:  quantity,
		FilledQty: decimal.Zero,
		Status:    "pending",
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// 提交到撮合引擎
	h.matchingManager.AddOrder(&order)

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status != "pending" && order.Status != "partial" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be cancelled"})
		return
	}

	// 从撮合引擎移除
	h.matchingManager.CancelOrder(order.ID, order.Symbol, order.Side)

	// 解冻资产
	remaining := order.Quantity.Sub(order.FilledQty)
	if order.Side == "buy" {
		quoteAsset := getQuoteAsset(order.Symbol)
		amount := order.Price.Mul(remaining)

		var balance models.Balance
		database.DB.Where("user_id = ? AND asset = ?", userID, quoteAsset).First(&balance)
		balance.Available = balance.Available.Add(amount)
		balance.Frozen = balance.Frozen.Sub(amount)
		database.DB.Save(&balance)
	} else {
		baseAsset := getBaseAsset(order.Symbol)

		var balance models.Balance
		database.DB.Where("user_id = ? AND asset = ?", userID, baseAsset).First(&balance)
		balance.Available = balance.Available.Add(remaining)
		balance.Frozen = balance.Frozen.Sub(remaining)
		database.DB.Save(&balance)
	}

	// 更新订单状态
	// 如果已经有成交，状态改为 partial_cancelled，否则为 cancelled
	if order.FilledQty.GreaterThan(decimal.Zero) {
		order.Status = "partial_cancelled"
	} else {
		order.Status = "cancelled"
	}
	database.DB.Save(&order)

	c.JSON(http.StatusOK, gin.H{
		"order":         order,
		"message":       "Order cancelled successfully",
		"filled_qty":    order.FilledQty.String(),
		"cancelled_qty": remaining.String(),
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	symbol := c.Query("symbol")
	status := c.Query("status")

	query := database.DB.Where("user_id = ?", userID)

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []models.Order
	query.Order("created_at DESC").Limit(100).Find(&orders)

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func getBaseAsset(symbol string) string {
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			return symbol[:i]
		}
	}
	return ""
}

func getQuoteAsset(symbol string) string {
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			return symbol[i+1:]
		}
	}
	return ""
}
