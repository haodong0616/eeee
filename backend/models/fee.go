package models

import (
	"expchange-backend/utils"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 交易手续费配置
type FeeConfig struct {
	ID              string          `gorm:"primaryKey;size:24" json:"id"`
	UserLevel       string          `gorm:"size:20;not null" json:"user_level"` // normal, vip1, vip2, vip3
	MakerFeeRate    decimal.Decimal `gorm:"type:decimal(10,6);not null" json:"maker_fee_rate"` // Maker 手续费率
	TakerFeeRate    decimal.Decimal `gorm:"type:decimal(10,6);not null" json:"taker_fee_rate"` // Taker 手续费率
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (f *FeeConfig) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = utils.GenerateObjectID()
	}
	return nil
}

// 手续费记录
type FeeRecord struct {
	ID          string          `gorm:"primaryKey;size:24" json:"id"`
	UserID      string          `gorm:"size:24;index;not null" json:"user_id"`
	OrderID     string          `gorm:"size:24;index;not null" json:"order_id"`
	TradeID     string          `gorm:"size:24;index;not null" json:"trade_id"`
	Asset       string          `gorm:"size:10;not null" json:"asset"`
	Amount      decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`
	FeeRate     decimal.Decimal `gorm:"type:decimal(10,6);not null" json:"fee_rate"`
	OrderSide   string          `gorm:"size:10;not null" json:"order_side"` // maker, taker
	CreatedAt   time.Time       `json:"created_at"`
}

func (f *FeeRecord) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = utils.GenerateObjectID()
	}
	return nil
}

