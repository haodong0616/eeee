package database

import (
	"expchange-backend/models"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// K线生成互斥锁，防止并发写入导致死锁
var klineGenerationMutex sync.Mutex

// 自动检测并初始化数据（仅初始化基础配置）
func AutoSeed() {
	// 检查是否已有交易对
	var count int64
	DB.Model(&models.TradingPair{}).Count(&count)

	if count > 0 {
		log.Printf("✅ 数据库中已有 %d 个交易对，跳过初始化", count)
		log.Println("💡 如需重新初始化，请删除 expchange.db 文件后重启")
		return
	}

	log.Println("🚀 首次启动，初始化基础数据...")
	startTime := time.Now()

	// 1. 创建交易对
	seedTradingPairs()

	// 2. 初始化手续费配置
	seedFeeConfig()

	// 3. 初始化系统配置
	seedSystemConfig()

	// 4. 初始化链配置
	seedChainConfig()

	elapsed := time.Since(startTime)
	log.Printf("🎉 基础数据初始化完成！耗时: %.2f秒\n", elapsed.Seconds())
	log.Println("💡 访问 http://localhost:3000 查看前端")
	log.Println("💡 访问 http://localhost:3001 查看管理后台")
	log.Println("💡 提示：请在管理后台为交易对生成初始化数据")
	log.Println("💡 提示：请在管理后台配置链的提现私钥")
}

func seedTradingPairs() {
	log.Println("📊 创建默认交易对...")

	pairs := []models.TradingPair{
		// 高价区 (>$5,000)
		{Symbol: "TITAN/USDT", BaseAsset: "TITAN", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},
		{Symbol: "GENESIS/USDT", BaseAsset: "GENESIS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},
		{Symbol: "LUNAR/USDT", BaseAsset: "LUNAR", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(1), MaxPrice: decimal.NewFromFloat(50000), MinQty: decimal.NewFromFloat(0.001), MaxQty: decimal.NewFromFloat(50000), Status: "active"},

		// 中高价区 ($1,000-$5,000)
		{Symbol: "ORACLE/USDT", BaseAsset: "ORACLE", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.1), MaxPrice: decimal.NewFromFloat(10000), MinQty: decimal.NewFromFloat(0.01), MaxQty: decimal.NewFromFloat(100000), Status: "active"},
		{Symbol: "QUANTUM/USDT", BaseAsset: "QUANTUM", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.1), MaxPrice: decimal.NewFromFloat(10000), MinQty: decimal.NewFromFloat(0.01), MaxQty: decimal.NewFromFloat(100000), Status: "active"},

		// 中价区 ($100-$1,000)
		{Symbol: "ATLAS/USDT", BaseAsset: "ATLAS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(5000), MinQty: decimal.NewFromFloat(0.1), MaxQty: decimal.NewFromFloat(200000), Status: "active"},
		{Symbol: "NEXUS/USDT", BaseAsset: "NEXUS", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(5000), MinQty: decimal.NewFromFloat(0.5), MaxQty: decimal.NewFromFloat(200000), Status: "active"},

		// 中低价区 ($10-$100)
		{Symbol: "AURORA/USDT", BaseAsset: "AURORA", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(1000), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(500000), Status: "active"},
		{Symbol: "ZEPHYR/USDT", BaseAsset: "ZEPHYR", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.01), MaxPrice: decimal.NewFromFloat(1000), MinQty: decimal.NewFromFloat(0.1), MaxQty: decimal.NewFromFloat(500000), Status: "active"},

		// 低价区 ($1-$10)
		{Symbol: "PULSE/USDT", BaseAsset: "PULSE", QuoteAsset: "USDT", MinPrice: decimal.NewFromFloat(0.001), MaxPrice: decimal.NewFromFloat(100), MinQty: decimal.NewFromFloat(1), MaxQty: decimal.NewFromFloat(1000000), Status: "active"},
	}

	for _, pair := range pairs {
		DB.Create(&pair)
	}

	log.Printf("✅ 创建了 %d 个默认交易对\n", len(pairs))
}

// SeedKlinesForSymbol 为单个交易对生成K线数据（优化版：并发 + 一次查询）
func SeedKlinesForSymbol(symbol string) {
	// 使用互斥锁，确保K线生成任务串行执行，避免死锁
	klineGenerationMutex.Lock()
	defer klineGenerationMutex.Unlock()

	log.Printf("📈 开始为 %s 生成K线数据...\n", symbol)

	// 获取交易对信息
	var pair models.TradingPair
	if err := DB.Where("symbol = ?", symbol).First(&pair).Error; err != nil {
		log.Printf("❌ 交易对不存在: %s", symbol)
		return
	}

	// 一次性查询所有交易数据（按时间排序）
	var allTrades []models.Trade
	start := time.Now()
	result := DB.Where("symbol = ?", symbol).
		Order("created_at ASC").
		Find(&allTrades)

	if result.Error != nil || len(allTrades) == 0 {
		log.Printf("❌ 该交易对没有交易数据: %s", symbol)
		return
	}

	log.Printf("📊 查询到 %d 条交易记录，耗时: %.2fs", len(allTrades), time.Since(start).Seconds())

	startTime := allTrades[0].CreatedAt
	now := time.Now()
	log.Printf("📊 时间范围: %s ~ %s", startTime.Format("2006-01-02 15:04"), now.Format("2006-01-02 15:04"))

	// 支持的K线周期（从小到大排序）
	intervals := []struct {
		name     string
		duration time.Duration
	}{
		{"15s", 15 * time.Second},
		{"30s", 30 * time.Second},
		{"1m", 1 * time.Minute},
		{"5m", 5 * time.Minute},
		{"15m", 15 * time.Minute},
		{"1h", 1 * time.Hour},
		{"4h", 4 * time.Hour},
		{"1d", 24 * time.Hour},
	}

	// 使用并发生成不同周期的K线
	type klineResult struct {
		interval string
		klines   []models.Kline
	}

	resultChan := make(chan klineResult, len(intervals))

	for _, interval := range intervals {
		go func(name string, duration time.Duration) {
			klines := generateKlinesForInterval(symbol, name, duration, allTrades, startTime, now)
			resultChan <- klineResult{interval: name, klines: klines}
		}(interval.name, interval.duration)
	}

	// 收集所有结果（优化内存分配：估算更准确的容量）
	estimatedSize := len(allTrades) / 2 // 保守估计
	allKlines := make([]models.Kline, 0, estimatedSize)
	totalKlines := 0

	for i := 0; i < len(intervals); i++ {
		result := <-resultChan
		allKlines = append(allKlines, result.klines...)
		log.Printf("   %s: %d 根", result.interval, len(result.klines))
		totalKlines += len(result.klines)
	}

	// 批量插入K线（使用UPSERT，避免重复和死锁）
	if len(allKlines) > 0 {
		log.Printf("📝 批量插入/更新 %d 根K线...", len(allKlines))
		insertStart := time.Now()

		// 使用事务批量插入/更新
		err := DB.Transaction(func(tx *gorm.DB) error {
			// 批量插入：每批500条
			batchSize := 500

			for i := 0; i < len(allKlines); i += batchSize {
				end := i + batchSize
				if end > len(allKlines) {
					end = len(allKlines)
				}

				batch := allKlines[i:end]

				// 使用 Clauses 实现 UPSERT
				// ON DUPLICATE KEY UPDATE: 如果唯一键冲突，则更新
				if err := tx.Clauses(clause.OnConflict{
					UpdateAll: true, // 更新所有字段
				}).CreateInBatches(batch, batchSize).Error; err != nil {
					return err
				}

				// 每插入5000条显示一次进度
				if end%5000 == 0 || end == len(allKlines) {
					log.Printf("   已插入 %d/%d 根...", end, len(allKlines))
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("❌ 插入K线失败: %v", err)
			return
		}

		log.Printf("✅ %s 生成了 %d 根K线，总耗时: %.2fs (插入耗时: %.2fs)\n",
			symbol, totalKlines, time.Since(start).Seconds(), time.Since(insertStart).Seconds())
	}
}

// generateKlinesForInterval 为单个周期生成K线（内存计算，不查数据库）
func generateKlinesForInterval(symbol string, interval string, duration time.Duration,
	allTrades []models.Trade, startTime, endTime time.Time) []models.Kline {

	klines := make([]models.Kline, 0, 1000)
	current := startTime.Truncate(duration)
	tradeIndex := 0

	for current.Before(endTime) {
		next := current.Add(duration)

		// 找到该时间段内的交易
		periodTrades := make([]models.Trade, 0)

		// 从上次结束的位置开始查找
		for tradeIndex < len(allTrades) && allTrades[tradeIndex].CreatedAt.Before(current) {
			tradeIndex++
		}

		tempIndex := tradeIndex
		for tempIndex < len(allTrades) && allTrades[tempIndex].CreatedAt.Before(next) {
			periodTrades = append(periodTrades, allTrades[tempIndex])
			tempIndex++
		}

		if len(periodTrades) > 0 {
			open := periodTrades[0].Price
			close := periodTrades[len(periodTrades)-1].Price
			high, low := open, open
			volume := decimal.Zero

			for _, trade := range periodTrades {
				if trade.Price.GreaterThan(high) {
					high = trade.Price
				}
				if trade.Price.LessThan(low) {
					low = trade.Price
				}
				volume = volume.Add(trade.Quantity)
			}

			klines = append(klines, models.Kline{
				Symbol:   symbol,
				Interval: interval,
				OpenTime: current.Unix(),
				Open:     open,
				High:     high,
				Low:      low,
				Close:    close,
				Volume:   volume,
			})
		}

		current = next
	}

	return klines
}

func seedFeeConfig() {
	// 检查是否已有配置
	var count int64
	DB.Model(&models.FeeConfig{}).Count(&count)

	if count > 0 {
		return
	}

	log.Println("📊 创建手续费配置...")

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

	log.Printf("✅ 创建了 %d 个手续费配置\n", len(configs))
}

func seedSystemConfig() {
	// 检查是否已有配置
	var count int64
	DB.Model(&models.SystemConfig{}).Count(&count)

	if count > 0 {
		return
	}

	log.Println("📊 创建系统配置...")

	configs := []models.SystemConfig{
		// 任务队列配置
		{Key: "task.queue.workers", Value: "10", Description: "任务队列并发Worker数量", Category: "task", ValueType: "number"},

		// 数据刷新配置
		{Key: "deposit.check.interval", Value: "30", Description: "充值检查间隔(秒)", Category: "blockchain", ValueType: "number"},
		{Key: "withdraw.check.interval", Value: "30", Description: "提现检查间隔(秒)", Category: "blockchain", ValueType: "number"},

		// WebSocket配置
		{Key: "websocket.reconnect.interval", Value: "5", Description: "WebSocket重连间隔(秒)", Category: "websocket", ValueType: "number"},

		// 市场模拟器配置
		{Key: "simulator.enabled", Value: "false", Description: "是否启用市场模拟器", Category: "simulator", ValueType: "boolean"},
		{Key: "simulator.trade.interval", Value: "3", Description: "模拟交易生成间隔(秒)", Category: "simulator", ValueType: "number"},
		{Key: "simulator.order.interval", Value: "10", Description: "模拟订单生成间隔(秒)", Category: "simulator", ValueType: "number"},

		// K线配置
		{Key: "kline.intervals", Value: "15s,30s,1m,5m,15m,30m,1h,4h,1d", Description: "K线周期（逗号分隔）", Category: "kline", ValueType: "string"},

		// 手续费配置
		{Key: "fee.normal.maker", Value: "0.001", Description: "普通用户Maker手续费率(0.1%)", Category: "fee", ValueType: "number"},
		{Key: "fee.normal.taker", Value: "0.002", Description: "普通用户Taker手续费率(0.2%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip1.maker", Value: "0.0008", Description: "VIP1用户Maker手续费率(0.08%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip1.taker", Value: "0.0015", Description: "VIP1用户Taker手续费率(0.15%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip2.maker", Value: "0.0005", Description: "VIP2用户Maker手续费率(0.05%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip2.taker", Value: "0.001", Description: "VIP2用户Taker手续费率(0.1%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip3.maker", Value: "0.0002", Description: "VIP3用户Maker手续费率(0.02%)", Category: "fee", ValueType: "number"},
		{Key: "fee.vip3.taker", Value: "0.0005", Description: "VIP3用户Taker手续费率(0.05%)", Category: "fee", ValueType: "number"},

		// 平台配置
		{Key: "platform.name", Value: "Velocity Exchange", Description: "平台名称", Category: "platform", ValueType: "string"},
		{Key: "platform.deposit.address", Value: "0x88888886757311de33778ce108fb312588e368db", Description: "平台充值收款地址", Category: "platform", ValueType: "string"},
		{Key: "platform.withdraw.address", Value: "0x88888886757311de33778ce108fb312588e368db", Description: "平台提现转账地址", Category: "platform", ValueType: "string"},
		{Key: "platform.private.key", Value: "", Description: "平台转账私钥（敏感信息）", Category: "platform", ValueType: "string"},
	}

	for _, config := range configs {
		DB.Create(&config)
	}

	log.Printf("✅ 创建了 %d 个系统配置\n", len(configs))
}

func seedChainConfig() {
	var count int64
	DB.Model(&models.ChainConfig{}).Count(&count)

	if count > 0 {
		return
	}

	log.Println("🔗 创建链配置...")

	chains := []models.ChainConfig{
		{
			ChainName:                  "Ethereum Mainnet",
			ChainID:                    1,
			RpcURL:                     "https://eth.llamarpc.com",
			BlockExplorerURL:           "https://etherscan.io",
			UsdtContractAddress:        "0xdac17f958d2ee523a2206206994597c13d831ec7", // USDT on Ethereum
			UsdtDecimals:               6,                                            // Ethereum USDT使用6位精度
			PlatformDepositAddress:     "0x88888886757311de33778ce108fb312588e368db",
			PlatformWithdrawPrivateKey: "",    // 需要在管理后台配置
			Enabled:                    false, // 默认禁用，管理员可手动启用
		},
		{
			ChainName:                  "BSC Mainnet",
			ChainID:                    56,
			RpcURL:                     "https://bsc-dataseed.binance.org",
			BlockExplorerURL:           "https://bscscan.com",
			UsdtContractAddress:        "0x55d398326f99059fF775485246999027B3197955",
			UsdtDecimals:               18, // BSC USDT使用18位精度
			PlatformDepositAddress:     "0x88888886757311de33778ce108fb312588e368db",
			PlatformWithdrawPrivateKey: "",   // 需要在管理后台配置
			Enabled:                    true, // 默认启用
		},
		{
			ChainName:                  "Polygon Mainnet",
			ChainID:                    137,
			RpcURL:                     "https://polygon-rpc.com",
			BlockExplorerURL:           "https://polygonscan.com",
			UsdtContractAddress:        "0xc2132d05d31c914a87c6611c10748aeb04b58e8f", // USDT on Polygon
			UsdtDecimals:               6,                                            // Polygon USDT使用6位精度
			PlatformDepositAddress:     "0x88888886757311de33778ce108fb312588e368db",
			PlatformWithdrawPrivateKey: "",    // 需要在管理后台配置
			Enabled:                    false, // 默认禁用
		},
		{
			ChainName:                  "Arbitrum One",
			ChainID:                    42161,
			RpcURL:                     "https://arb1.arbitrum.io/rpc",
			BlockExplorerURL:           "https://arbiscan.io",
			UsdtContractAddress:        "0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9", // USDT on Arbitrum
			UsdtDecimals:               6,                                            // Arbitrum USDT使用6位精度
			PlatformDepositAddress:     "0x88888886757311de33778ce108fb312588e368db",
			PlatformWithdrawPrivateKey: "",    // 需要在管理后台配置
			Enabled:                    false, // 默认禁用
		},
		{
			ChainName:                  "Sepolia Testnet",
			ChainID:                    11155111,
			RpcURL:                     "https://ethereum-sepolia.publicnode.com",
			BlockExplorerURL:           "https://sepolia.etherscan.io",
			UsdtContractAddress:        "0x49433da9Bb68917A4dc35eB7565629289aA1BDf8", // 测试USDT
			UsdtDecimals:               6,                                            // Sepolia测试USDT使用6位精度
			PlatformDepositAddress:     "0x88888886757311de33778ce108fb312588e368db",
			PlatformWithdrawPrivateKey: "",   // 需要在管理后台配置
			Enabled:                    true, // 测试网默认启用
		},
	}

	for _, chain := range chains {
		DB.Create(&chain)
	}

	log.Printf("✅ 创建了 %d 个链配置\n", len(chains))
}
