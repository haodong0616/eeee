package handlers

import (
	"expchange-backend/database"
	"expchange-backend/matching"
	"expchange-backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type MarketHandler struct {
	matchingManager *matching.Manager
}

func NewMarketHandler(matchingManager *matching.Manager) *MarketHandler {
	return &MarketHandler{
		matchingManager: matchingManager,
	}
}

// 将 URL 中的 symbol (BTC-USDT) 转换为数据库格式 (BTC/USDT)
func normalizeSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "-", "/")
}

func (h *MarketHandler) GetTradingPairs(c *gin.Context) {
	var pairs []models.TradingPair
	if err := database.DB.Where("status = ?", "active").Find(&pairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trading pairs"})
		return
	}

	c.JSON(http.StatusOK, pairs)
}

func (h *MarketHandler) GetTicker(c *gin.Context) {
	symbol := normalizeSymbol(c.Param("symbol"))

	// 获取24小时数据
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	var trades []models.Trade
	database.DB.Where("symbol = ? AND created_at >= ?", symbol, dayAgo).
		Order("created_at DESC").
		Find(&trades)

	if len(trades) == 0 {
		c.JSON(http.StatusOK, models.Ticker{
			Symbol:    symbol,
			UpdatedAt: now,
		})
		return
	}

	lastPrice := trades[0].Price
	high := lastPrice
	low := lastPrice
	volume := decimal.Zero

	firstPrice := trades[len(trades)-1].Price

	for _, trade := range trades {
		if trade.Price.GreaterThan(high) {
			high = trade.Price
		}
		if trade.Price.LessThan(low) {
			low = trade.Price
		}
		volume = volume.Add(trade.Quantity)
	}

	change := decimal.Zero
	if !firstPrice.IsZero() {
		change = lastPrice.Sub(firstPrice).Div(firstPrice).Mul(decimal.NewFromInt(100))
	}

	ticker := models.Ticker{
		Symbol:    symbol,
		LastPrice: lastPrice,
		Change24h: change,
		High24h:   high,
		Low24h:    low,
		Volume24h: volume,
		UpdatedAt: now,
	}

	c.JSON(http.StatusOK, ticker)
}

func (h *MarketHandler) GetAllTickers(c *gin.Context) {
	var pairs []models.TradingPair
	database.DB.Where("status = ?", "active").Find(&pairs)

	tickers := make([]models.Ticker, 0, len(pairs))
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)

	for _, pair := range pairs {
		var trades []models.Trade
		database.DB.Where("symbol = ? AND created_at >= ?", pair.Symbol, dayAgo).
			Order("created_at DESC").
			Find(&trades)

		if len(trades) == 0 {
			tickers = append(tickers, models.Ticker{
				Symbol:    pair.Symbol,
				UpdatedAt: now,
			})
			continue
		}

		lastPrice := trades[0].Price
		high := lastPrice
		low := lastPrice
		volume := decimal.Zero
		firstPrice := trades[len(trades)-1].Price

		for _, trade := range trades {
			if trade.Price.GreaterThan(high) {
				high = trade.Price
			}
			if trade.Price.LessThan(low) {
				low = trade.Price
			}
			volume = volume.Add(trade.Quantity)
		}

		change := decimal.Zero
		if !firstPrice.IsZero() {
			change = lastPrice.Sub(firstPrice).Div(firstPrice).Mul(decimal.NewFromInt(100))
		}

		tickers = append(tickers, models.Ticker{
			Symbol:    pair.Symbol,
			LastPrice: lastPrice,
			Change24h: change,
			High24h:   high,
			Low24h:    low,
			Volume24h: volume,
			UpdatedAt: now,
		})
	}

	c.JSON(http.StatusOK, tickers)
}

func (h *MarketHandler) GetOrderBook(c *gin.Context) {
	symbol := normalizeSymbol(c.Param("symbol"))
	depth := 20 // 默认深度

	// ⚠️ 从数据库查询虚拟订单展示盘口（虚拟订单不在匹配引擎中）
	orderBook := h.getOrderBookFromDB(symbol, depth)
	c.JSON(http.StatusOK, orderBook)
}

// getOrderBookFromDB 从数据库查询虚拟订单展示盘口
func (h *MarketHandler) getOrderBookFromDB(symbol string, depth int) *models.OrderBook {
	var buyOrders []models.Order
	var sellOrders []models.Order

	// 查询虚拟用户的pending订单
	database.DB.Where("symbol = ? AND user_id = ? AND status = ? AND side = 'buy'",
		symbol, "virtual-simulator", "pending").
		Order("price DESC").
		Limit(depth).
		Find(&buyOrders)

	database.DB.Where("symbol = ? AND user_id = ? AND status = ? AND side = 'sell'",
		symbol, "virtual-simulator", "pending").
		Order("price ASC").
		Limit(depth).
		Find(&sellOrders)

	orderBook := &models.OrderBook{
		Symbol: symbol,
		Bids:   []models.OrderBookItem{},
		Asks:   []models.OrderBookItem{},
	}

	for _, order := range buyOrders {
		orderBook.Bids = append(orderBook.Bids, models.OrderBookItem{
			Price:    order.Price,
			Quantity: order.Quantity,
		})
	}

	for _, order := range sellOrders {
		orderBook.Asks = append(orderBook.Asks, models.OrderBookItem{
			Price:    order.Price,
			Quantity: order.Quantity,
		})
	}

	return orderBook
}

func (h *MarketHandler) GetRecentTrades(c *gin.Context) {
	symbol := normalizeSymbol(c.Param("symbol"))
	limit := 50 // 默认50条

	var trades []models.Trade
	database.DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		Limit(limit).
		Find(&trades)

	c.JSON(http.StatusOK, trades)
}

func (h *MarketHandler) GetKlines(c *gin.Context) {
	symbol := normalizeSymbol(c.Param("symbol"))
	interval := c.DefaultQuery("interval", "1m")
	limit := 100

	var klines []models.Kline
	database.DB.Where("symbol = ? AND `interval` = ?", symbol, interval).
		Order("open_time DESC").
		Limit(limit).
		Find(&klines)

	c.JSON(http.StatusOK, klines)
}
