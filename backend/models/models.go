package models

import (
	"expchange-backend/utils"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type User struct {
	ID            string    `gorm:"primaryKey;size:24" json:"id"`
	WalletAddress string    `gorm:"uniqueIndex;size:42;not null" json:"wallet_address"`
	Nonce         string    `gorm:"size:100" json:"-"`
	UserLevel     string    `gorm:"size:20;default:'normal'" json:"user_level"` // normal, vip1, vip2, vip3
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = utils.GenerateObjectID()
	}
	// 强制转换为小写
	u.WalletAddress = strings.ToLower(u.WalletAddress)
	return nil
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	// 强制转换为小写
	u.WalletAddress = strings.ToLower(u.WalletAddress)
	return nil
}

type TradingPair struct {
	ID               string          `gorm:"primaryKey;size:24" json:"id"`
	Symbol           string          `gorm:"uniqueIndex;size:20;not null" json:"symbol"` // e.g., BTC/USDT
	BaseAsset        string          `gorm:"size:10;not null" json:"base_asset"`         // e.g., BTC
	QuoteAsset       string          `gorm:"size:10;not null" json:"quote_asset"`        // e.g., USDT
	MinPrice         decimal.Decimal `gorm:"type:decimal(20,8)" json:"min_price"`
	MaxPrice         decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_price"`
	MinQty           decimal.Decimal `gorm:"type:decimal(20,8)" json:"min_qty"`
	MaxQty           decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_qty"`
	Status           string          `gorm:"size:20;default:'active'" json:"status"` // active, inactive
	SimulatorEnabled bool            `gorm:"default:false" json:"simulator_enabled"` // 是否启用市场模拟器
	// 活跃度配置（1-10，默认5）
	ActivityLevel      int             `gorm:"default:5" json:"activity_level"`                          // 1=低活跃度, 5=中等, 10=高活跃度
	OrderbookDepth     int             `gorm:"default:15" json:"orderbook_depth"`                        // 订单簿档位数（5-30）
	TradeFrequency     int             `gorm:"default:20" json:"trade_frequency"`                        // 成交间隔秒数（5-60）
	PriceVolatility    decimal.Decimal `gorm:"type:decimal(10,4);default:0.01" json:"price_volatility"`  // 价格波动率（0.001-0.05）
	VirtualTradePer10s int             `gorm:"default:10" json:"virtual_trade_per_10s"`                  // 虚拟成交频率（每10秒N笔，1-30）
	PriceSpreadRatio   decimal.Decimal `gorm:"type:decimal(10,2);default:1.0" json:"price_spread_ratio"` // 盘口价格分布范围倍数（0.5-5.0）
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

func (t *TradingPair) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.GenerateObjectID()
	}
	return nil
}

type Balance struct {
	ID        string          `gorm:"primaryKey;size:24" json:"id"`
	UserID    string          `gorm:"size:24;index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Available decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"available"`
	Frozen    decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"frozen"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (b *Balance) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = utils.GenerateObjectID()
	}
	return nil
}

type Order struct {
	ID        string          `gorm:"primaryKey;size:24" json:"id"`
	UserID    string          `gorm:"size:24;index;not null" json:"user_id"`
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

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = utils.GenerateObjectID()
	}
	return nil
}

type Trade struct {
	ID          string          `gorm:"primaryKey;size:24" json:"id"`
	Symbol      string          `gorm:"size:20;not null;index" json:"symbol"`
	BuyOrderID  string          `gorm:"size:24" json:"buy_order_id"`
	SellOrderID string          `gorm:"size:24" json:"sell_order_id"`
	Price       decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"price"`
	Quantity    decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"quantity"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (t *Trade) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.GenerateObjectID()
	}
	return nil
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
	ID        string          `gorm:"primaryKey;size:24" json:"id"`
	Symbol    string          `gorm:"size:20;not null;index:idx_kline_unique,unique" json:"symbol"`
	Interval  string          `gorm:"size:10;not null;index:idx_kline_unique,unique" json:"interval"` // 15s, 30s, 1m, 5m, 15m, 1h, 4h, 1d
	OpenTime  int64           `gorm:"not null;index;index:idx_kline_unique,unique" json:"open_time"`
	Open      decimal.Decimal `gorm:"type:decimal(20,8)" json:"open"`
	High      decimal.Decimal `gorm:"type:decimal(20,8)" json:"high"`
	Low       decimal.Decimal `gorm:"type:decimal(20,8)" json:"low"`
	Close     decimal.Decimal `gorm:"type:decimal(20,8)" json:"close"`
	Volume    decimal.Decimal `gorm:"type:decimal(20,8)" json:"volume"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (k *Kline) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = utils.GenerateObjectID()
	}
	return nil
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
	ID        string          `gorm:"primaryKey;size:24" json:"id"`
	UserID    string          `gorm:"size:24;index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Amount    decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`
	TxHash    string          `gorm:"size:66;uniqueIndex;not null" json:"tx_hash"` // 交易hash
	Chain     string          `gorm:"size:20;not null;default:'bsc'" json:"chain"` // bsc, sepolia
	ChainID   int             `gorm:"not null;default:56" json:"chain_id"`         // 链ID
	Status    string          `gorm:"size:20;not null;index" json:"status"`        // pending, confirmed, failed
	TaskID    string          `gorm:"size:24;index" json:"task_id,omitempty"`      // 关联的验证任务ID
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (d *DepositRecord) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = utils.GenerateObjectID()
	}
	return nil
}

// 提现记录
type WithdrawRecord struct {
	ID        string          `gorm:"primaryKey;size:24" json:"id"`
	UserID    string          `gorm:"size:24;index;not null" json:"user_id"`
	Asset     string          `gorm:"size:10;not null" json:"asset"`
	Amount    decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`
	Address   string          `gorm:"size:42;not null" json:"address"`             // 提现地址
	TxHash    string          `gorm:"size:66;index" json:"tx_hash"`                // 转账hash（成功后填充）
	Chain     string          `gorm:"size:20;not null;default:'bsc'" json:"chain"` // bsc, sepolia
	ChainID   int             `gorm:"not null;default:56" json:"chain_id"`         // 链ID
	Status    string          `gorm:"size:20;not null;index" json:"status"`        // pending, processing, completed, failed
	TaskID    string          `gorm:"size:24;index" json:"task_id,omitempty"`      // 关联的处理任务ID
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (w *WithdrawRecord) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = utils.GenerateObjectID()
	}
	return nil
}

// 做市商盈亏记录
type MarketMakerPnL struct {
	ID            string          `gorm:"primaryKey;size:24" json:"id"`
	Symbol        string          `gorm:"size:20;not null;index" json:"symbol"`
	TradeID       string          `gorm:"size:24;index" json:"trade_id"`                    // 关联的交易ID
	Side          string          `gorm:"size:10;not null" json:"side"`                     // buy, sell
	ExecutePrice  decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"execute_price"` // 执行价格
	MarketPrice   decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"market_price"`  // 当时市价
	Quantity      decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"quantity"`
	ProfitLoss    decimal.Decimal `gorm:"type:decimal(20,8)" json:"profit_loss"`    // 盈亏（USDT）
	ProfitPercent decimal.Decimal `gorm:"type:decimal(10,4)" json:"profit_percent"` // 盈亏百分比
	CreatedAt     time.Time       `json:"created_at"`
}

func (m *MarketMakerPnL) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = utils.GenerateObjectID()
	}
	return nil
}
