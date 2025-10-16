package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	WalletAddress string    `gorm:"uniqueIndex;size:42;not null" json:"wallet_address"`
	Nonce         string    `json:"-"`
	UserLevel     string    `gorm:"size:20;default:'normal'" json:"user_level"` // normal, vip1, vip2, vip3
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TradingPair struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	Symbol     string          `gorm:"uniqueIndex;size:20;not null" json:"symbol"` // e.g., BTC/USDT
	BaseAsset  string          `gorm:"size:10;not null" json:"base_asset"`         // e.g., BTC
	QuoteAsset string          `gorm:"size:10;not null" json:"quote_asset"`        // e.g., USDT
	MinPrice   decimal.Decimal `gorm:"type:decimal(20,8)" json:"min_price"`
	MaxPrice   decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_price"`
	MinQty     decimal.Decimal `gorm:"type:decimal(20,8)" json:"min_qty"`
	MaxQty     decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_qty"`
	Status     string          `gorm:"size:20;default:'active'" json:"status"` // active, inactive
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Balance struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Available decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"available"`
	Frozen    decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"frozen"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Order struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"`
	Symbol    string          `gorm:"size:20;not null;index" json:"symbol"`
	OrderType string          `gorm:"size:20;not null" json:"order_type"` // limit, market
	Side      string          `gorm:"size:10;not null" json:"side"`       // buy, sell
	Price     decimal.Decimal `gorm:"type:decimal(20,8)" json:"price"`
	Quantity  decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"quantity"`
	FilledQty decimal.Decimal `gorm:"type:decimal(20,8);default:0" json:"filled_qty"`
	Status    string          `gorm:"size:20;not null;index" json:"status"` // pending, filled, partial, cancelled
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Trade struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Symbol      string          `gorm:"size:20;not null;index" json:"symbol"`
	BuyOrderID  uint            `gorm:"not null" json:"buy_order_id"`
	SellOrderID uint            `gorm:"not null" json:"sell_order_id"`
	Price       decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"price"`
	Quantity    decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"quantity"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Ticker struct {
	Symbol    string          `json:"symbol"`
	LastPrice decimal.Decimal `json:"last_price"`
	Change24h decimal.Decimal `json:"change_24h"`
	High24h   decimal.Decimal `json:"high_24h"`
	Low24h    decimal.Decimal `json:"low_24h"`
	Volume24h decimal.Decimal `json:"volume_24h"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Kline struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Symbol    string          `gorm:"size:20;not null;index" json:"symbol"`
	Interval  string          `gorm:"size:10;not null" json:"interval"` // 1m, 5m, 15m, 1h, 4h, 1d
	OpenTime  int64           `gorm:"not null;index" json:"open_time"`
	Open      decimal.Decimal `gorm:"type:decimal(20,8)" json:"open"`
	High      decimal.Decimal `gorm:"type:decimal(20,8)" json:"high"`
	Low       decimal.Decimal `gorm:"type:decimal(20,8)" json:"low"`
	Close     decimal.Decimal `gorm:"type:decimal(20,8)" json:"close"`
	Volume    decimal.Decimal `gorm:"type:decimal(20,8)" json:"volume"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type OrderBook struct {
	Symbol string          `json:"symbol"`
	Bids   []OrderBookItem `json:"bids"` // 买单（从高到低）
	Asks   []OrderBookItem `json:"asks"` // 卖单（从低到高）
}

type OrderBookItem struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

// 充值记录
type DepositRecord struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Amount    decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`
	TxHash    string          `gorm:"size:66;uniqueIndex;not null" json:"tx_hash"` // 交易hash
	Status    string          `gorm:"size:20;not null;index" json:"status"`        // pending, confirmed, failed
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 提现记录
type WithdrawRecord struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Amount    decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`
	Address   string          `gorm:"size:42;not null" json:"address"`      // 提现地址
	TxHash    string          `gorm:"size:66;index" json:"tx_hash"`         // 转账hash（成功后填充）
	Status    string          `gorm:"size:20;not null;index" json:"status"` // pending, processing, completed, failed
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
