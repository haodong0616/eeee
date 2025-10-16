package matching

import (
	"container/heap"
	"expchange-backend/models"
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

func (e *Engine) CancelOrder(orderID uint, side string) bool {
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

	// 聚合买单
	buyPriceMap := make(map[string]decimal.Decimal)
	for _, order := range *e.buyOrders {
		remaining := order.Quantity.Sub(order.FilledQty)
		priceKey := order.Price.String()
		buyPriceMap[priceKey] = buyPriceMap[priceKey].Add(remaining)
	}

	// 聚合卖单
	sellPriceMap := make(map[string]decimal.Decimal)
	for _, order := range *e.sellOrders {
		remaining := order.Quantity.Sub(order.FilledQty)
		priceKey := order.Price.String()
		sellPriceMap[priceKey] = sellPriceMap[priceKey].Add(remaining)
	}

	// 转换为列表（已按价格排序）
	count := 0
	for _, order := range *e.buyOrders {
		if count >= depth {
			break
		}
		priceKey := order.Price.String()
		if qty, exists := buyPriceMap[priceKey]; exists {
			orderBook.Bids = append(orderBook.Bids, models.OrderBookItem{
				Price:    order.Price,
				Quantity: qty,
			})
			delete(buyPriceMap, priceKey)
			count++
		}
	}

	count = 0
	for _, order := range *e.sellOrders {
		if count >= depth {
			break
		}
		priceKey := order.Price.String()
		if qty, exists := sellPriceMap[priceKey]; exists {
			orderBook.Asks = append(orderBook.Asks, models.OrderBookItem{
				Price:    order.Price,
				Quantity: qty,
			})
			delete(sellPriceMap, priceKey)
			count++
		}
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
