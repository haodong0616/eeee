package simulator

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/websocket"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

type MarketSimulator struct {
	hub       *websocket.Hub
	running   bool
	symbols   []string
	basePrice map[string]float64
}

func NewMarketSimulator(hub *websocket.Hub) *MarketSimulator {
	return &MarketSimulator{
		hub:     hub,
		running: false,
		symbols: []string{"BTC/USDT", "ETH/USDT", "BNB/USDT", "SOL/USDT", "XRP/USDT"},
		basePrice: map[string]float64{
			"BTC/USDT": 65000,
			"ETH/USDT": 3500,
			"BNB/USDT": 580,
			"SOL/USDT": 145,
			"XRP/USDT": 0.52,
		},
	}
}

func (s *MarketSimulator) Start() {
	if s.running {
		return
	}

	s.running = true
	log.Println("🎮 市场模拟器已启动")

	// 为每个交易对启动独立的模拟器
	for _, symbol := range s.symbols {
		go s.simulateMarket(symbol)
	}
}

func (s *MarketSimulator) Stop() {
	s.running = false
	log.Println("🛑 市场模拟器已停止")
}

func (s *MarketSimulator) simulateMarket(symbol string) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒生成一次交易
	defer ticker.Stop()

	currentPrice := s.basePrice[symbol]

	for range ticker.C {
		if !s.running {
			return
		}

		// 获取当前价格（从最近的成交记录）
		var lastTrade models.Trade
		result := database.DB.Where("symbol = ?", symbol).
			Order("created_at DESC").
			First(&lastTrade)

		if result.Error == nil {
			currentPrice, _ = lastTrade.Price.Float64()
		}

		// 生成随机价格波动 (-0.5% 到 +0.5%)
		priceChange := (rand.Float64() - 0.5) * 0.01 * currentPrice
		newPrice := currentPrice + priceChange

		// 确保价格不会太离谱
		basePrice := s.basePrice[symbol]
		if newPrice < basePrice*0.8 {
			newPrice = basePrice * 0.8
		} else if newPrice > basePrice*1.2 {
			newPrice = basePrice * 1.2
		}

		// 生成随机交易量
		var quantity float64
		switch symbol {
		case "BTC/USDT":
			quantity = 0.01 + rand.Float64()*0.2
		case "ETH/USDT":
			quantity = 0.1 + rand.Float64()*2
		case "BNB/USDT":
			quantity = 1 + rand.Float64()*20
		case "SOL/USDT":
			quantity = 5 + rand.Float64()*50
		case "XRP/USDT":
			quantity = 100 + rand.Float64()*1000
		}

		// 创建虚拟成交记录
		trade := models.Trade{
			Symbol:      symbol,
			BuyOrderID:  0, // 虚拟交易
			SellOrderID: 0,
			Price:       decimal.NewFromFloat(newPrice),
			Quantity:    decimal.NewFromFloat(quantity),
		}

		if err := database.DB.Create(&trade).Error; err != nil {
			log.Printf("❌ 创建虚拟交易失败 (%s): %v", symbol, err)
			continue
		}

		log.Printf("💹 模拟交易: %s @ $%.2f x %.4f", symbol, newPrice, quantity)

		// 通过WebSocket推送交易数据
		if s.hub != nil {
			s.hub.BroadcastTrade(map[string]interface{}{
				"symbol":   symbol,
				"price":    newPrice,
				"quantity": quantity,
				"time":     time.Now().Unix(),
			})
		}
	}
}

// 市场趋势模拟器 - 更复杂的价格模型
type TrendSimulator struct {
	hub       *websocket.Hub
	running   bool
	symbols   []string
	basePrice map[string]float64
	trend     map[string]float64 // 当前趋势 (-1 到 1)
}

func NewTrendSimulator(hub *websocket.Hub) *TrendSimulator {
	return &TrendSimulator{
		hub:     hub,
		running: false,
		symbols: []string{"LUNAR/USDT", "NOVA/USDT", "ZEPHYR/USDT", "PULSE/USDT", "ARCANA/USDT", "NEXUS/USDT"},
		basePrice: map[string]float64{
			"LUNAR/USDT":  8500,
			"NOVA/USDT":   1250,
			"ZEPHYR/USDT": 45.5,
			"PULSE/USDT":  2.85,
			"ARCANA/USDT": 0.0085,
			"NEXUS/USDT":  125,
		},
		trend: make(map[string]float64),
	}
}

func (s *TrendSimulator) Start() {
	if s.running {
		return
	}

	s.running = true
	log.Println("📈 趋势模拟器已启动")

	// 初始化趋势
	for _, symbol := range s.symbols {
		s.trend[symbol] = 0
	}

	// 为每个交易对启动独立的模拟器
	for _, symbol := range s.symbols {
		go s.simulateWithTrend(symbol)
	}

	// 定期改变趋势
	go s.changeTrends()
}

func (s *TrendSimulator) Stop() {
	s.running = false
	log.Println("🛑 趋势模拟器已停止")
}

func (s *TrendSimulator) changeTrends() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒可能改变趋势
	defer ticker.Stop()

	for range ticker.C {
		if !s.running {
			return
		}

		for _, symbol := range s.symbols {
			// 30% 概率改变趋势
			if rand.Float64() < 0.3 {
				// 趋势在 -1 到 1 之间
				s.trend[symbol] = (rand.Float64() - 0.5) * 2

				trendStr := "震荡"
				if s.trend[symbol] > 0.3 {
					trendStr = "上涨"
				} else if s.trend[symbol] < -0.3 {
					trendStr = "下跌"
				}

				log.Printf("🔄 %s 趋势变化: %s (%.2f)", symbol, trendStr, s.trend[symbol])
			}
		}
	}
}

func (s *TrendSimulator) simulateWithTrend(symbol string) {
	ticker := time.NewTicker(3 * time.Second) // 每3秒生成一次交易
	defer ticker.Stop()

	currentPrice := s.basePrice[symbol]

	for range ticker.C {
		if !s.running {
			return
		}

		// 获取当前价格
		var lastTrade models.Trade
		result := database.DB.Where("symbol = ?", symbol).
			Order("created_at DESC").
			First(&lastTrade)

		if result.Error == nil {
			currentPrice, _ = lastTrade.Price.Float64()
		}

		// 基于趋势生成价格变化
		trend := s.trend[symbol]

		// 随机波动 + 趋势影响
		randomChange := (rand.Float64() - 0.5) * 0.005 // ±0.5%
		trendChange := trend * 0.003                   // 趋势影响 ±0.3%

		priceChangePercent := randomChange + trendChange
		newPrice := currentPrice * (1 + priceChangePercent)

		// 价格边界
		basePrice := s.basePrice[symbol]
		minPrice := basePrice * 0.85
		maxPrice := basePrice * 1.15

		if newPrice < minPrice {
			newPrice = minPrice
			s.trend[symbol] = 0.5 // 反弹
		} else if newPrice > maxPrice {
			newPrice = maxPrice
			s.trend[symbol] = -0.5 // 回调
		}

		// 生成交易量（交易量随波动性增加）
		volatility := math.Abs(priceChangePercent) * 100
		var baseQty float64

		switch symbol {
		case "LUNAR/USDT":
			baseQty = 0.2 + volatility*0.5
		case "NOVA/USDT":
			baseQty = 2.0 + volatility*5
		case "ZEPHYR/USDT":
			baseQty = 20 + volatility*50
		case "PULSE/USDT":
			baseQty = 200 + volatility*500
		case "ARCANA/USDT":
			baseQty = 2000 + volatility*5000
		case "NEXUS/USDT":
			baseQty = 10 + volatility*30
		}

		quantity := baseQty * (0.5 + rand.Float64())

		// 创建虚拟成交记录
		trade := models.Trade{
			Symbol:      symbol,
			BuyOrderID:  0,
			SellOrderID: 0,
			Price:       decimal.NewFromFloat(newPrice),
			Quantity:    decimal.NewFromFloat(quantity),
		}

		if err := database.DB.Create(&trade).Error; err != nil {
			continue
		}

		// 推送数据
		if s.hub != nil {
			s.hub.BroadcastTrade(map[string]interface{}{
				"symbol":   symbol,
				"price":    newPrice,
				"quantity": quantity,
				"change":   priceChangePercent * 100,
				"time":     time.Now().Unix(),
			})
		}
	}
}
