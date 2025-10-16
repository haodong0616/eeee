package services

import (
	"context"
	"crypto/ecdsa"
	"expchange-backend/database"
	"expchange-backend/models"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// ERC20 Transfer ABI
const transferABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`

// WithdrawProcessor 提现处理服务
type WithdrawProcessor struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	ctx        context.Context
}

// NewWithdrawProcessor 创建提现处理服务
func NewWithdrawProcessor(privateKeyHex string) (*WithdrawProcessor, error) {
	client, err := ethclient.Dial(BSC_RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BSC RPC: %w", err)
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &WithdrawProcessor{
		client:     client,
		privateKey: privateKey,
		ctx:        context.Background(),
	}, nil
}

// Start 启动提现处理队列
func (p *WithdrawProcessor) Start() {
	log.Println("🚀 提现处理队列已启动")

	// 每1分钟检查一次待处理的提现
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.ProcessPendingWithdrawals()
		}
	}
}

// ProcessPendingWithdrawals 处理待处理的提现记录
func (p *WithdrawProcessor) ProcessPendingWithdrawals() {
	var withdrawals []models.WithdrawRecord

	// 查询待处理的提现
	err := database.DB.Where("status = ?", "pending").Find(&withdrawals).Error
	if err != nil {
		log.Printf("❌ 查询待处理提现失败: %v", err)
		return
	}

	if len(withdrawals) == 0 {
		return
	}

	log.Printf("📋 发现 %d 条待处理提现", len(withdrawals))

	for _, withdrawal := range withdrawals {
		p.ProcessWithdrawal(&withdrawal)
	}
}

// ProcessWithdrawal 处理单个提现记录
func (p *WithdrawProcessor) ProcessWithdrawal(withdrawal *models.WithdrawRecord) {
	log.Printf("💸 处理提现: ID=%d, Amount=%s %s, Address=%s",
		withdrawal.ID, withdrawal.Amount.String(), withdrawal.Asset, withdrawal.Address)

	// 1. 标记为处理中
	if err := database.DB.Model(withdrawal).Update("status", "processing").Error; err != nil {
		log.Printf("❌ 更新提现状态失败: %v", err)
		return
	}

	// 2. 执行链上转账
	txHash, err := p.TransferUSDT(withdrawal.Address, withdrawal.Amount)
	if err != nil {
		log.Printf("❌ 转账失败: %v", err)
		p.MarkWithdrawalFailed(withdrawal, err.Error())
		return
	}

	log.Printf("✅ 转账成功: %s", txHash)

	// 3. 更新提现记录
	p.ConfirmWithdrawal(withdrawal, txHash)
}

// TransferUSDT 执行 USDT 转账
func (p *WithdrawProcessor) TransferUSDT(toAddress string, amount decimal.Decimal) (string, error) {
	// 获取 nonce
	fromAddress := crypto.PubkeyToAddress(p.privateKey.PublicKey)
	nonce, err := p.client.PendingNonceAt(p.ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// 获取 gas price
	gasPrice, err := p.client.SuggestGasPrice(p.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(transferABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// USDT 有 18 位小数
	value := new(big.Int)
	value.SetString(amount.Shift(18).StringFixed(0), 10)

	// 打包 transfer 函数调用
	data, err := parsedABI.Pack("transfer", common.HexToAddress(toAddress), value)
	if err != nil {
		return "", fmt.Errorf("failed to pack data: %w", err)
	}

	// 获取 chain ID
	chainID, err := p.client.NetworkID(p.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}

	// 创建交易
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(USDT_CONTRACT),
		big.NewInt(0),  // value = 0 (ERC20 transfer)
		uint64(100000), // gas limit
		gasPrice,
		data,
	)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), p.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	err = p.client.SendTransaction(p.ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// ConfirmWithdrawal 确认提现完成
func (p *WithdrawProcessor) ConfirmWithdrawal(withdrawal *models.WithdrawRecord, txHash string) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新提现记录
	if err := tx.Model(withdrawal).Updates(map[string]interface{}{
		"status":     "completed",
		"tx_hash":    txHash,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新提现记录失败: %v", err)
		return
	}

	// 2. 减少冻结余额
	var balance models.Balance
	if err := tx.Where("user_id = ? AND asset = ?",
		withdrawal.UserID, withdrawal.Asset).First(&balance).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 查询余额失败: %v", err)
		return
	}

	balance.Frozen = balance.Frozen.Sub(withdrawal.Amount)
	if err := tx.Save(&balance).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新余额失败: %v", err)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事务失败: %v", err)
		return
	}

	log.Printf("🎉 提现已完成: 用户ID=%d, 资产=%s, 金额=%s, TxHash=%s",
		withdrawal.UserID, withdrawal.Asset, withdrawal.Amount.String(), txHash)
}

// MarkWithdrawalFailed 标记提现失败
func (p *WithdrawProcessor) MarkWithdrawalFailed(withdrawal *models.WithdrawRecord, reason string) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新提现记录状态
	if err := tx.Model(withdrawal).Updates(map[string]interface{}{
		"status":     "failed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新提现记录失败: %v", err)
		return
	}

	// 2. 解冻资金
	var balance models.Balance
	if err := tx.Where("user_id = ? AND asset = ?",
		withdrawal.UserID, withdrawal.Asset).First(&balance).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 查询余额失败: %v", err)
		return
	}

	balance.Frozen = balance.Frozen.Sub(withdrawal.Amount)
	balance.Available = balance.Available.Add(withdrawal.Amount)
	if err := tx.Save(&balance).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新余额失败: %v", err)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事务失败: %v", err)
		return
	}

	log.Printf("❌ 提现已标记为失败并解冻资金: ID=%d, 原因=%s", withdrawal.ID, reason)
}
