package handlers

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"net/http"

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
	status := c.PostForm("status")

	var pair models.TradingPair
	if err := database.DB.First(&pair, pairID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	pair.Status = status
	database.DB.Save(&pair)

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
