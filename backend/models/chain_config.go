package models

import (
	"expchange-backend/utils"
	"time"

	"gorm.io/gorm"
)

// ChainConfig 链配置
// 注意：ChainName 和 ChainID 创建后不可修改
type ChainConfig struct {
	ID                         string    `gorm:"primaryKey;size:24" json:"id"`
	ChainName                  string    `gorm:"uniqueIndex;size:100;not null" json:"chain_name"`        // 链名称（创建后不可修改）
	ChainID                    int       `gorm:"uniqueIndex;not null" json:"chain_id"`                   // 链ID（创建后不可修改）
	RpcURL                     string    `gorm:"type:varchar(500);not null" json:"rpc_url"`              // RPC地址
	BlockExplorerURL           string    `gorm:"type:varchar(500)" json:"block_explorer_url"`            // 区块浏览器地址
	UsdtContractAddress        string    `gorm:"size:42;not null" json:"usdt_contract_address"`          // USDT合约地址
	UsdtDecimals               int       `gorm:"not null;default:18" json:"usdt_decimals"`               // USDT精度（6或18）
	PlatformDepositAddress     string    `gorm:"size:42;not null" json:"platform_deposit_address"`       // 平台充值收款地址
	PlatformWithdrawPrivateKey string    `gorm:"type:varchar(500)" json:"platform_withdraw_private_key"` // 平台提现转账私钥
	Enabled                    bool      `gorm:"default:true" json:"enabled"`                            // 是否启用
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

func (c *ChainConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = utils.GenerateObjectID()
	}
	return nil
}
