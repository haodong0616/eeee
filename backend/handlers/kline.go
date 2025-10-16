package handlers

import (
	"expchange-backend/database"
	"expchange-backend/kline"
	"expchange-backend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type KlineHandler struct {
	generator *kline.Generator
}

func NewKlineHandler(generator *kline.Generator) *KlineHandler {
	return &KlineHandler{
		generator: generator,
	}
}

func (h *KlineHandler) GetKlines(c *gin.Context) {
	symbol := strings.ReplaceAll(c.Param("symbol"), "-", "/")
	interval := c.DefaultQuery("interval", "1m")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	var klines []models.Kline
	database.DB.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("open_time DESC").
		Limit(limit).
		Find(&klines)

	// 反转顺序，从旧到新
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}

	c.JSON(http.StatusOK, klines)
}

// 为TradingView格式化K线数据
func (h *KlineHandler) GetKlinesForTradingView(c *gin.Context) {
	symbol := strings.ReplaceAll(c.Param("symbol"), "-", "/")
	interval := c.DefaultQuery("interval", "1m")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to parameters required"})
		return
	}

	fromTimestamp, _ := strconv.ParseInt(fromStr, 10, 64)
	toTimestamp, _ := strconv.ParseInt(toStr, 10, 64)

	var klines []models.Kline
	database.DB.Where("symbol = ? AND interval = ? AND open_time >= ? AND open_time <= ?",
		symbol, interval, fromTimestamp, toTimestamp).
		Order("open_time ASC").
		Find(&klines)

	// TradingView格式
	response := gin.H{
		"s": "ok",
		"t": []int64{},
		"o": []string{},
		"h": []string{},
		"l": []string{},
		"c": []string{},
		"v": []string{},
	}

	for _, kline := range klines {
		response["t"] = append(response["t"].([]int64), kline.OpenTime)
		response["o"] = append(response["o"].([]string), kline.Open.String())
		response["h"] = append(response["h"].([]string), kline.High.String())
		response["l"] = append(response["l"].([]string), kline.Low.String())
		response["c"] = append(response["c"].([]string), kline.Close.String())
		response["v"] = append(response["v"].([]string), kline.Volume.String())
	}

	c.JSON(http.StatusOK, response)
}

// 生成历史K线数据（管理员接口）
func (h *KlineHandler) GenerateHistoricalKlines(c *gin.Context) {
	symbol := c.PostForm("symbol")
	interval := c.PostForm("interval")
	fromStr := c.PostForm("from")
	toStr := c.PostForm("to")

	if symbol == "" || interval == "" || fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from time format"})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to time format"})
		return
	}

	go h.generator.GenerateHistoricalKlines(symbol, interval, from, to)

	c.JSON(http.StatusOK, gin.H{"message": "Historical klines generation started"})
}

