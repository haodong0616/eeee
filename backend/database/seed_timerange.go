package database

import (
	"expchange-backend/models"
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

// SeedTradesForSymbolWithTimeRange 为单个交易对生成指定时间范围的历史交易数据
func SeedTradesForSymbolWithTimeRange(symbol string, startTime, endTime time.Time) {
	log.Printf("💱 开始为 %s 生成历史交易数据 [%s ~ %s]...\n",
		symbol, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	// 获取交易对信息
	var pair models.TradingPair
	if err := DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		log.Printf("❌ 交易对不存在: %s", symbol)
		return
	}

	// 根据交易对设置基础价格
	basePriceMap := map[string]float64{
		"TITAN/USDT":   12000,
		"GENESIS/USDT": 9500,
		"LUNAR/USDT":   8500,
		"ORACLE/USDT":  3200,
		"QUANTUM/USDT": 2800,
		"ATLAS/USDT":   450,
		"NEXUS/USDT":   125,
		"AURORA/USDT":  52,
		"ZEPHYR/USDT":  45,
		"PULSE/USDT":   2.85,
	}

	basePrice, exists := basePriceMap[symbol]
	if !exists {
		// 如果不在预设列表中，使用随机价格
		basePrice = 100 + rand.Float64()*900 // 100-1000之间
		log.Printf("⚠️  使用随机基础价格: %.2f", basePrice)
	}

	currentPrice := basePrice
	currentTime := startTime

	// 随机交易频率 (1-4笔/小时)
	tradeFreq := 1 + rand.Intn(4)
	log.Printf("📈 交易频率: %d 笔/小时", tradeFreq)

	allTrades := make([]models.Trade, 0, 10000)

	// 生成交易数据
	for currentTime.Before(endTime) {
		hour := currentTime.Hour()

		// 白天(8-22点)交易更频繁
		var tradesThisHour int
		if hour >= 8 && hour <= 22 {
			tradesThisHour = tradeFreq
		} else {
			tradesThisHour = rand.Intn(tradeFreq + 1)
		}

		for i := 0; i < tradesThisHour; i++ {
			// 价格波动
			daysSinceStart := currentTime.Sub(startTime).Hours() / 24
			trendFactor := 1.0

			// 每30天可能出现趋势转变
			if int(daysSinceStart)%30 == 0 && rand.Float64() > 0.5 {
				trendFactor = 0.95 + rand.Float64()*0.1
			}

			// 日内随机波动 ±2%
			priceChange := (rand.Float64() - 0.5) * 0.04
			currentPrice = currentPrice * (1 + priceChange) * trendFactor

			// 保持在合理范围 (基准价的 70%-150%)
			if currentPrice < basePrice*0.7 {
				currentPrice = basePrice * 0.7
			}
			if currentPrice > basePrice*1.5 {
				currentPrice = basePrice * 1.5
			}

			// 根据价格决定数量范围
			var quantity float64
			if basePrice >= 5000 {
				quantity = 0.01 + rand.Float64()*0.5
			} else if basePrice >= 1000 {
				quantity = 0.1 + rand.Float64()*2
			} else if basePrice >= 100 {
				quantity = 1 + rand.Float64()*20
			} else if basePrice >= 10 {
				quantity = 5 + rand.Float64()*100
			} else if basePrice >= 1 {
				quantity = 50 + rand.Float64()*500
			} else {
				quantity = 500 + rand.Float64()*5000
			}

			tradeTime := currentTime.Add(time.Duration(rand.Intn(3600)) * time.Second)
			// 确保不超过结束时间
			if tradeTime.After(endTime) {
				tradeTime = endTime.Add(-time.Second)
			}

			trade := models.Trade{
				Symbol:      symbol,
				BuyOrderID:  "",
				SellOrderID: "",
				Price:       decimal.NewFromFloat(currentPrice),
				Quantity:    decimal.NewFromFloat(quantity),
			}
			trade.CreatedAt = tradeTime

			allTrades = append(allTrades, trade)
		}

		currentTime = currentTime.Add(1 * time.Hour)
	}

	// 批量插入
	log.Printf("📝 批量插入 %d 条交易记录...", len(allTrades))
	batchSize := 1000
	for i := 0; i < len(allTrades); i += batchSize {
		end := i + batchSize
		if end > len(allTrades) {
			end = len(allTrades)
		}

		batch := allTrades[i:end]
		for j := range batch {
			DB.Create(&batch[j])
		}

		if (i/batchSize+1)%10 == 0 {
			log.Printf("   已插入 %d/%d 条...", end, len(allTrades))
		}
	}

	log.Printf("✅ %s 生成了 %d 条交易记录\n", symbol, len(allTrades))
}
