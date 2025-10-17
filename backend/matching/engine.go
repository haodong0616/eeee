package matching

import (
	"container/heap"
	"expchange-backend/models"
	"expchange-backend/utils"
	"log"
	"sync"

	"github.com/shopspring/decimal"
)

type Engine struct {
	symbol     string
	buyOrders  *BuyOrderQueue
	sellOrders *SellOrderQueue
	mu         sync.RWMutex
	tradeChan  chan *models.Trade
}

func NewEngine(symbol string, tradeChan chan *models.Trade) *Engine {
	buyQueue := &BuyOrderQueue{}
	sellQueue := &SellOrderQueue{}
	heap.Init(buyQueue)
	heap.Init(sellQueue)

	return &Engine{
		symbol:     symbol,
		buyOrders:  buyQueue,
		sellOrders: sellQueue,
		tradeChan:  tradeChan,
	}
}

func (e *Engine) AddOrder(order *models.Order) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if order.Side == "buy" {
		heap.Push(e.buyOrders, order)
	} else {
		heap.Push(e.sellOrders, order)
	}

	e.match()
}

func (e *Engine) CancelOrder(orderID string, side string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if side == "buy" {
		for i, order := range *e.buyOrders {
			if order.ID == orderID {
				heap.Remove(e.buyOrders, i)
				return true
			}
		}
	} else {
		for i, order := range *e.sellOrders {
			if order.ID == orderID {
				heap.Remove(e.sellOrders, i)
				return true
			}
		}
	}
	return false
}

func (e *Engine) match() {
	for e.buyOrders.Len() > 0 && e.sellOrders.Len() > 0 {
		buyOrder := (*e.buyOrders)[0]
		sellOrder := (*e.sellOrders)[0]

		// 市价单或限价单满足条件
		if buyOrder.OrderType == "market" || sellOrder.OrderType == "market" ||
			buyOrder.Price.GreaterThanOrEqual(sellOrder.Price) {

			// 确定成交价格（取卖单价格）
			tradePrice := sellOrder.Price
			if sellOrder.OrderType == "market" {
				tradePrice = buyOrder.Price
			}

			// 计算成交数量
			buyRemaining := buyOrder.Quantity.Sub(buyOrder.FilledQty)
			sellRemaining := sellOrder.Quantity.Sub(sellOrder.FilledQty)
			tradeQty := buyRemaining
			if sellRemaining.LessThan(buyRemaining) {
				tradeQty = sellRemaining
			}

			// 更新订单状态
			buyOrder.FilledQty = buyOrder.FilledQty.Add(tradeQty)
			sellOrder.FilledQty = sellOrder.FilledQty.Add(tradeQty)

			// 生成成交记录
			trade := &models.Trade{
				Symbol:      e.symbol,
				BuyOrderID:  buyOrder.ID,
				SellOrderID: sellOrder.ID,
				Price:       tradePrice,
				Quantity:    tradeQty,
			}

			// 发送成交记录到通道
			select {
			case e.tradeChan <- trade:
			default:
			}

			// 检查订单是否完全成交
			if buyOrder.FilledQty.Equal(buyOrder.Quantity) {
				heap.Pop(e.buyOrders)
			}
			if sellOrder.FilledQty.Equal(sellOrder.Quantity) {
				heap.Pop(e.sellOrders)
			}

			// 如果有任何一方没有完全成交，继续匹配
			if buyOrder.FilledQty.LessThan(buyOrder.Quantity) ||
				sellOrder.FilledQty.LessThan(sellOrder.Quantity) {
				continue
			}
		}
		break
	}
}

func (e *Engine) GetOrderBook(depth int) *models.OrderBook {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orderBook := &models.OrderBook{
		Symbol: e.symbol,
		Bids:   []models.OrderBookItem{},
		Asks:   []models.OrderBookItem{},
	}

	// 确定价格精度（使用统一的工具函数）
	var pricePrecision int32 = 3 // 默认3位小数
	if len(*e.buyOrders) > 0 {
		samplePrice := (*e.buyOrders)[0].Price
		pricePrecision = utils.GetPricePrecision(samplePrice)
		log.Printf("🔢 %s 价格精度设置: %s → %d位小数", e.symbol, utils.FormatPriceString(samplePrice), pricePrecision)
	}

	// 聚合买单（按显示精度舍入后合并）
	buyPriceMap := make(map[string]decimal.Decimal)
	for _, order := range *e.buyOrders {
		remaining := order.Quantity.Sub(order.FilledQty)
		// 按显示精度舍入，确保1.258和1.2580001会被合并为1.258
		roundedPrice := order.Price.Round(pricePrecision)
		priceKey := roundedPrice.String()
		buyPriceMap[priceKey] = buyPriceMap[priceKey].Add(remaining)
	}

	// 聚合卖单（按显示精度舍入后合并）
	sellPriceMap := make(map[string]decimal.Decimal)
	for _, order := range *e.sellOrders {
		remaining := order.Quantity.Sub(order.FilledQty)
		// 按显示精度舍入
		roundedPrice := order.Price.Round(pricePrecision)
		priceKey := roundedPrice.String()
		sellPriceMap[priceKey] = sellPriceMap[priceKey].Add(remaining)
	}

	// 从map中提取并排序输出（确保正确顺序）
	// 买单：价格从高到低
	type priceQty struct {
		price decimal.Decimal
		qty   decimal.Decimal
	}

	buyList := make([]priceQty, 0, len(buyPriceMap))
	for priceKey, qty := range buyPriceMap {
		price, _ := decimal.NewFromString(priceKey)
		buyList = append(buyList, priceQty{price: price, qty: qty})
	}

	// 买单排序：价格从高到低
	for i := 0; i < len(buyList); i++ {
		for j := i + 1; j < len(buyList); j++ {
			if buyList[j].price.GreaterThan(buyList[i].price) {
				buyList[i], buyList[j] = buyList[j], buyList[i]
			}
		}
	}

	// 输出买单（前depth档）
	for i := 0; i < len(buyList) && i < depth; i++ {
		orderBook.Bids = append(orderBook.Bids, models.OrderBookItem{
			Price:    buyList[i].price,
			Quantity: buyList[i].qty,
		})
	}

	// 卖单：价格从低到高
	sellList := make([]priceQty, 0, len(sellPriceMap))
	for priceKey, qty := range sellPriceMap {
		price, _ := decimal.NewFromString(priceKey)
		sellList = append(sellList, priceQty{price: price, qty: qty})
	}

	// 卖单排序：价格从低到高
	for i := 0; i < len(sellList); i++ {
		for j := i + 1; j < len(sellList); j++ {
			if sellList[j].price.LessThan(sellList[i].price) {
				sellList[i], sellList[j] = sellList[j], sellList[i]
			}
		}
	}

	// 输出卖单（前depth档）
	for i := 0; i < len(sellList) && i < depth; i++ {
		orderBook.Asks = append(orderBook.Asks, models.OrderBookItem{
			Price:    sellList[i].price,
			Quantity: sellList[i].qty,
		})
	}

	// 调试日志
	if len(orderBook.Bids) >= 3 && len(orderBook.Asks) >= 3 {
		log.Printf("🔢 %s 盘口排序 - 买[%.3f>%.3f>%.3f] 卖[%.3f<%.3f<%.3f]",
			e.symbol,
			orderBook.Bids[0].Price.InexactFloat64(),
			orderBook.Bids[1].Price.InexactFloat64(),
			orderBook.Bids[2].Price.InexactFloat64(),
			orderBook.Asks[0].Price.InexactFloat64(),
			orderBook.Asks[1].Price.InexactFloat64(),
			orderBook.Asks[2].Price.InexactFloat64())
	}

	return orderBook
}

// BuyOrderQueue 买单优先队列（价格从高到低）
type BuyOrderQueue []*models.Order

func (q BuyOrderQueue) Len() int { return len(q) }

func (q BuyOrderQueue) Less(i, j int) bool {
	// 价格高的优先，价格相同时间早的优先
	if q[i].Price.Equal(q[j].Price) {
		return q[i].CreatedAt.Before(q[j].CreatedAt)
	}
	return q[i].Price.GreaterThan(q[j].Price)
}

func (q BuyOrderQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *BuyOrderQueue) Push(x interface{}) {
	*q = append(*q, x.(*models.Order))
}

func (q *BuyOrderQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}

// SellOrderQueue 卖单优先队列（价格从低到高）
type SellOrderQueue []*models.Order

func (q SellOrderQueue) Len() int { return len(q) }

func (q SellOrderQueue) Less(i, j int) bool {
	// 价格低的优先，价格相同时间早的优先
	if q[i].Price.Equal(q[j].Price) {
		return q[i].CreatedAt.Before(q[j].CreatedAt)
	}
	return q[i].Price.LessThan(q[j].Price)
}

func (q SellOrderQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *SellOrderQueue) Push(x interface{}) {
	*q = append(*q, x.(*models.Order))
}

func (q *SellOrderQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}
