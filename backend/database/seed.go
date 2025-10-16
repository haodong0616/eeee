package database

import (
	"expchange-backend/models"
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

// 自动检测并初始化数据
func AutoSeed() {
	// 检查是否已有交易对
	var count int64
	DB.Model(&models.TradingPair{}).Count(&count)

	if count > 0 {
		log.Println("✅ 数据已存在，跳过初始化")
		return
	}

	log.Println("🚀 首次启动，开始自动初始化数据（智能模式）...")
	log.Println("📊 策略：生成 6-12 个月随机数据，每个代币活跃度随机")
	startTime := time.Now()

	// 使用事务提升性能
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 临时替换DB为事务
	originalDB := DB
	DB = tx

	// 1. 创建交易对
	seedTradingPairs()

	// 2. 创建虚拟用户
	virtualUser := createVirtualUser()

	// 3. 生成历史交易数据
	seedTrades()

	// 4. 生成K线数据
	seedKlines()

	// 5. 初始化手续费配置
	seedFeeConfig()

	// 恢复原始DB
	DB = originalDB

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 初始化失败: %v", err)
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("🎉 数据初始化完成！耗时: %.2f秒\n", elapsed.Seconds())
	log.Println("💡 访问 http://localhost:3000 查看前端")
	log.Println("💡 访问 http://localhost:3001 查看管理后台")

	_ = virtualUser
}

func seedTradingPairs() {
	log.Println("📊 创建交易对...")

	pairs := []models.TradingPair{
		// 高价区 (>$5,000)
		{Symbol: "TITAN/USDT", BaseAsset: "TITAN", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},
		{Symbol: "GENESIS/USDT", BaseAsset: "GENESIS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},
		{Symbol: "LUNAR/USDT", BaseAsset: "LUNAR", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},

		// 中高价区 ($1,000-$5,000)
		{Symbol: "ORACLE/USDT", BaseAsset: "ORACLE", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.1), MaxPrice: decimal.NewFromFloat(10000), MinQty: decimal.NewFromFloat(0.01), MaxQty: decimal.NewFromFloat(100000), Status: "active"},
		{Symbol: "QUANTUM/USDT", BaseAsset: "QUANTUM", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.1), MaxPrice: decimal.NewFromFloat(10000), MinQty: decimal.NewFromFloat(0.01), MaxQty: decimal.NewFromFloat(100000), Status: "active"},
		{Symbol: "NOVA/USDT", BaseAsset: "NOVA", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.1), MaxPrice: decimal.NewFromFloat(10000), MinQty: decimal.NewFromFloat(0.01), MaxQty: decimal.NewFromFloat(100000), Status: "active"},

		// 中价区 ($100-$1,000)
		{Symbol: "ATLAS/USDT", BaseAsset: "ATLAS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(5000), MinQty: decimal.NewFromFloat(0.1), MaxQty: decimal.NewFromFloat(200000), Status: "active"},
		{Symbol: "COSMOS/USDT", BaseAsset: "COSMOS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(5000), MinQty: decimal.NewFromFloat(0.1), MaxQty: decimal.NewFromFloat(200000), Status: "active"},
		{Symbol: "NEXUS/USDT", BaseAsset: "NEXUS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(5000), MinQty: decimal.NewFromFloat(0.5), MaxQty: decimal.NewFromFloat(200000), Status: "active"},

		// 中低价区 ($10-$100)
		{Symbol: "VERTEX/USDT", BaseAsset: "VERTEX", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(1000), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(500000), Status: "active"},
		{Symbol: "AURORA/USDT", BaseAsset: "AURORA", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(1000), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(500000), Status: "active"},
		{Symbol: "ZEPHYR/USDT", BaseAsset: "ZEPHYR", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(1000), MinQty: decimal.NewFromFloat(0.1), MaxQty: decimal.NewFromFloat(500000), Status: "active"},

		// 低价区 ($1-$10)
		{Symbol: "PRISM/USDT", BaseAsset: "PRISM", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.001), MaxPrice: decimal.NewFromFloat(100), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(1000000), Status: "active"},
		{Symbol: "PULSE/USDT", BaseAsset: "PULSE", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.001), MaxPrice: decimal.NewFromFloat(100), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(1000000), Status: "active"},

		// 超低价区 (<$1)
		{Symbol: "ARCANA/USDT", BaseAsset: "ARCANA", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.00001), MaxPrice: decimal.NewFromFloat(10), MinQty: decimal.NewFromFloat(10), MaxQty: decimal.NewFromFloat(10000000), Status: "active"},
	}

	for _, pair := range pairs {
		DB.Create(&pair)
	}

	log.Printf("✅ 创建了 %d 个交易对\n", len(pairs))
}

func createVirtualUser() *models.User {
	virtualUser := models.User{
		WalletAddress: "0x0000000000000000000000000000000000000000",
		Nonce:         "virtual",
		UserLevel:     "normal",
	}
	DB.Create(&virtualUser)

	// 为虚拟用户充值
	assets := []string{
		"TITAN", "GENESIS", "LUNAR", "ORACLE", "QUANTUM", "NOVA",
		"ATLAS", "COSMOS", "NEXUS", "VERTEX", "AURORA", "ZEPHYR",
		"PRISM", "PULSE", "ARCANA", "USDT",
	}
	for _, asset := range assets {
		balance := models.Balance{
			UserID:    virtualUser.ID,
			Asset:     asset,
			Available: decimal.NewFromFloat(10000000),
			Frozen:    decimal.Zero,
		}
		DB.Create(&balance)
	}

	return &virtualUser
}

func seedTrades() {
	// 随机生成6-12个月的数据
	monthsBack := 6 + rand.Intn(7) // 6-12个月
	log.Printf("💱 生成 %d 个月交易数据（智能模式）...\n", monthsBack)

	now := time.Now()
	startTime := now.AddDate(0, -monthsBack, 0)

	symbols := []string{
		"TITAN/USDT", "GENESIS/USDT", "LUNAR/USDT",
		"ORACLE/USDT", "QUANTUM/USDT", "NOVA/USDT",
		"ATLAS/USDT", "COSMOS/USDT", "NEXUS/USDT",
		"VERTEX/USDT", "AURORA/USDT", "ZEPHYR/USDT",
		"PRISM/USDT", "PULSE/USDT", "ARCANA/USDT",
	}

	basePrices := map[string]float64{
		"TITAN/USDT":   12000,
		"GENESIS/USDT": 9500,
		"LUNAR/USDT":   8500,
		"ORACLE/USDT":  3200,
		"QUANTUM/USDT": 2800,
		"NOVA/USDT":    1250,
		"ATLAS/USDT":   450,
		"COSMOS/USDT":  380,
		"NEXUS/USDT":   125,
		"VERTEX/USDT":  85,
		"AURORA/USDT":  52,
		"ZEPHYR/USDT":  45,
		"PRISM/USDT":   8.5,
		"PULSE/USDT":   2.85,
		"ARCANA/USDT":  0.0085,
	}

	// 每个代币随机的交易活跃度
	tradeFrequency := make(map[string]int)
	for _, symbol := range symbols {
		tradeFrequency[symbol] = 1 + rand.Intn(4) // 1-4 笔/小时
	}

	log.Println("📊 各代币交易频率：")
	for symbol, freq := range tradeFrequency {
		log.Printf("   %s: %d 笔/小时", symbol, freq)
	}

	// 批量生成
	allTrades := make([]models.Trade, 0, 10000)

	for _, symbol := range symbols {
		basePrice := basePrices[symbol]
		currentPrice := basePrice
		currentTime := startTime
		maxFreq := tradeFrequency[symbol]

		// 每天的交易模式（白天更活跃）
		for currentTime.Before(now) {
			hour := currentTime.Hour()

			// 白天(8-22点)交易更频繁
			var tradesThisHour int
			if hour >= 8 && hour <= 22 {
				tradesThisHour = maxFreq
			} else {
				// 夜间交易减少
				tradesThisHour = rand.Intn(maxFreq + 1)
			}

			for i := 0; i < tradesThisHour; i++ {
				// 价格趋势波动（每月可能有大的趋势变化）
				daysSinceStart := currentTime.Sub(startTime).Hours() / 24
				trendFactor := 1.0

				// 每30天可能出现一次趋势转变
				if int(daysSinceStart)%30 == 0 && rand.Float64() > 0.5 {
					trendFactor = 0.95 + rand.Float64()*0.1 // 0.95-1.05
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
				basePrice := basePrices[symbol]
				if basePrice >= 5000 {
					quantity = 0.01 + rand.Float64()*0.5 // 高价币，小数量
				} else if basePrice >= 1000 {
					quantity = 0.1 + rand.Float64()*2
				} else if basePrice >= 100 {
					quantity = 1 + rand.Float64()*20
				} else if basePrice >= 10 {
					quantity = 5 + rand.Float64()*100
				} else if basePrice >= 1 {
					quantity = 50 + rand.Float64()*500
				} else {
					quantity = 500 + rand.Float64()*5000 // 低价币，大数量
				}

				tradeTime := currentTime.Add(time.Duration(rand.Intn(3600)) * time.Second)

				trade := models.Trade{
					Symbol:      symbol,
					BuyOrderID:  0,
					SellOrderID: 0,
					Price:       decimal.NewFromFloat(currentPrice),
					Quantity:    decimal.NewFromFloat(quantity),
				}
				trade.CreatedAt = tradeTime

				// 批量添加，稍后统一插入
				allTrades = append(allTrades, trade)
			}

			currentTime = currentTime.Add(1 * time.Hour)
		}
	}

	// 批量插入交易记录（每1000条一批）
	log.Printf("📝 批量插入 %d 条交易记录...\n", len(allTrades))
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

	log.Printf("✅ 生成了 %d 条交易记录 (%d 个月)\n", len(allTrades), monthsBack)
}

func seedKlines() {
	log.Println("📈 生成K线数据（智能模式）...")

	// 获取最早的交易时间
	var firstTrade models.Trade
	DB.Order("created_at ASC").First(&firstTrade)

	startTime := firstTrade.CreatedAt
	now := time.Now()

	symbols := []string{
		"TITAN/USDT", "GENESIS/USDT", "LUNAR/USDT",
		"ORACLE/USDT", "QUANTUM/USDT", "NOVA/USDT",
		"ATLAS/USDT", "COSMOS/USDT", "NEXUS/USDT",
		"VERTEX/USDT", "AURORA/USDT", "ZEPHYR/USDT",
		"PRISM/USDT", "PULSE/USDT", "ARCANA/USDT",
	}

	// 只生成常用周期，减少计算量
	intervals := []struct {
		name     string
		duration time.Duration
	}{
		{"1d", 24 * time.Hour},
		{"4h", 4 * time.Hour},
		{"1h", 1 * time.Hour},
		{"15m", 15 * time.Minute},
		{"5m", 5 * time.Minute},
		{"1m", 1 * time.Minute},
	}

	totalKlines := 0
	allKlines := make([]models.Kline, 0, 50000)

	for _, symbol := range symbols {
		log.Printf("   处理 %s K线...", symbol)

		for _, interval := range intervals {
			current := startTime.Truncate(interval.duration)

			for current.Before(now) {
				next := current.Add(interval.duration)

				var trades []models.Trade
				DB.Where("symbol = ? AND created_at >= ? AND created_at < ?",
					symbol, current, next).
					Order("created_at ASC").
					Find(&trades)

				if len(trades) > 0 {
					open := trades[0].Price
					close := trades[len(trades)-1].Price
					high, low := open, open
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

					kline := models.Kline{
						Symbol:   symbol,
						Interval: interval.name,
						OpenTime: current.Unix(),
						Open:     open,
						High:     high,
						Low:      low,
						Close:    close,
						Volume:   volume,
					}

					allKlines = append(allKlines, kline)
					totalKlines++
				}

				current = next
			}
		}
	}

	// 批量插入K线（每次2000条）
	log.Printf("📝 批量插入 %d 根K线...\n", len(allKlines))
	batchSize := 2000
	for i := 0; i < len(allKlines); i += batchSize {
		end := i + batchSize
		if end > len(allKlines) {
			end = len(allKlines)
		}
		DB.Create(allKlines[i:end])

		if (i/batchSize+1)%5 == 0 {
			log.Printf("   已插入 %d/%d 根...", end, len(allKlines))
		}
	}

	log.Printf("✅ 生成了 %d 根K线\n", totalKlines)
}

func seedFeeConfig() {
	// 检查是否已有配置
	var count int64
	DB.Model(&models.FeeConfig{}).Count(&count)

	if count > 0 {
		return
	}

	configs := []models.FeeConfig{
		{
			UserLevel:    "normal",
			MakerFeeRate: decimal.NewFromFloat(0.001),
			TakerFeeRate: decimal.NewFromFloat(0.002),
		},
		{
			UserLevel:    "vip1",
			MakerFeeRate: decimal.NewFromFloat(0.0008),
			TakerFeeRate: decimal.NewFromFloat(0.0015),
		},
		{
			UserLevel:    "vip2",
			MakerFeeRate: decimal.NewFromFloat(0.0005),
			TakerFeeRate: decimal.NewFromFloat(0.001),
		},
		{
			UserLevel:    "vip3",
			MakerFeeRate: decimal.NewFromFloat(0.0002),
			TakerFeeRate: decimal.NewFromFloat(0.0005),
		},
	}

	for _, config := range configs {
		DB.Create(&config)
	}
}
