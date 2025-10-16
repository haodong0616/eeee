package kline

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Generator struct {
	intervals map[string]time.Duration
}

func NewGenerator() *Generator {
	return &Generator{
		intervals: map[string]time.Duration{
			"15s": 15 * time.Second,
			"30s": 30 * time.Second,
			"1m":  1 * time.Minute,
			"3m":  3 * time.Minute,
			"5m":  5 * time.Minute,
			"15m": 15 * time.Minute,
			"30m": 30 * time.Minute,
			"1h":  1 * time.Hour,
			"4h":  4 * time.Hour,
			"1d":  24 * time.Hour,
		},
	}
}

func (g *Generator) Start() {
	// 启动各个时间周期的K线生成器
	for interval, duration := range g.intervals {
		go g.generateKlines(interval, duration)
	}
}

func (g *Generator) generateKlines(interval string, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		g.generateKlineForAllPairs(interval, duration)
	}
}

func (g *Generator) generateKlineForAllPairs(interval string, duration time.Duration) {
	// 获取所有交易对
	var pairs []models.TradingPair
	database.DB.Where("status = ?", "active").Find(&pairs)

	now := time.Now()
	openTime := now.Truncate(duration).Unix()

	for _, pair := range pairs {
		g.generateKlineForPair(pair.Symbol, interval, openTime, now)
	}
}

func (g *Generator) generateKlineForPair(symbol string, interval string, openTime int64, endTime time.Time) {
	// 获取该时间段内的所有成交记录
	startTime := time.Unix(openTime, 0)

	var trades []models.Trade
	database.DB.Where("symbol = ? AND created_at >= ? AND created_at < ?",
		symbol, startTime, endTime).
		Order("created_at ASC").
		Find(&trades)

	if len(trades) == 0 {
		// 如果没有成交记录，使用上一根K线的收盘价
		var lastKline models.Kline
		// 静默查询，不记录"record not found"错误
		result := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
			Where("symbol = ? AND interval = ?", symbol, interval).
			Order("open_time DESC").
			First(&lastKline)

		if result.Error == nil {
			// 创建新K线，OHLC都使用上一根K线的收盘价
			kline := models.Kline{
				Symbol:   symbol,
				Interval: interval,
				OpenTime: openTime,
				Open:     lastKline.Close,
				High:     lastKline.Close,
				Low:      lastKline.Close,
				Close:    lastKline.Close,
				Volume:   decimal.Zero,
			}
			database.DB.Create(&kline)
		}
		return
	}

	// 计算OHLC
	open := trades[0].Price
	close := trades[len(trades)-1].Price
	high := open
	low := open
	volume := decimal.Zero

	for _, trade := range trades {
		if trade.Price.GreaterThan(high) {
			high = trade.Price
		}
		if trade.Price.LessThan(low) {
			low = trade.Price
		}
		volume = volume.Add(trade.Quantity)
	}

	// 检查是否已存在该K线
	var existingKline models.Kline
	result := database.DB.Where("symbol = ? AND interval = ? AND open_time = ?",
		symbol, interval, openTime).First(&existingKline)

	if result.Error != nil {
		// 创建新K线
		kline := models.Kline{
			Symbol:   symbol,
			Interval: interval,
			OpenTime: openTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Volume:   volume,
		}
		database.DB.Create(&kline)
		// 静默创建，不输出日志
	} else {
		// 更新现有K线
		existingKline.High = high
		existingKline.Low = low
		existingKline.Close = close
		existingKline.Volume = volume
		database.DB.Save(&existingKline)
	}
}

// 手动生成历史K线数据
func (g *Generator) GenerateHistoricalKlines(symbol string, interval string, from time.Time, to time.Time) {
	duration := g.intervals[interval]
	if duration == 0 {
		log.Printf("Invalid interval: %s", interval)
		return
	}

	current := from.Truncate(duration)
	for current.Before(to) {
		next := current.Add(duration)
		g.generateKlineForPair(symbol, interval, current.Unix(), next)
		current = next
	}

	log.Printf("Generated historical klines for %s %s from %s to %s",
		symbol, interval, from, to)
}
