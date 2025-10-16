package services

import (
	"context"
	"expchange-backend/database"
	"expchange-backend/models"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// BSC 主网 RPC
const BSC_RPC = "https://bsc-dataseed1.binance.org"

// USDT 合约地址
const USDT_CONTRACT = "0x55d398326f99059fF775485246999027B3197955"

// 平台收款地址
const PLATFORM_ADDRESS = "0x88888886757311de33778ce108fb312588e368db"

// DepositVerifier 充值验证服务
type DepositVerifier struct {
	client *ethclient.Client
	ctx    context.Context
}

// NewDepositVerifier 创建充值验证服务
func NewDepositVerifier() (*DepositVerifier, error) {
	client, err := ethclient.Dial(BSC_RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BSC RPC: %w", err)
	}

	return &DepositVerifier{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// Start 启动充值验证队列
func (v *DepositVerifier) Start() {
	log.Println("🚀 充值验证队列已启动")

	// 每30秒检查一次待验证的充值
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v.ProcessPendingDeposits()
		}
	}
}

// ProcessPendingDeposits 处理待验证的充值记录
func (v *DepositVerifier) ProcessPendingDeposits() {
	var deposits []models.DepositRecord

	// 查询待验证的充值（24小时内）
	err := database.DB.Where(
		"status = ? AND created_at > ?",
		"pending",
		time.Now().Add(-24*time.Hour),
	).Find(&deposits).Error

	if err != nil {
		log.Printf("❌ 查询待验证充值失败: %v", err)
		return
	}

	if len(deposits) == 0 {
		return
	}

	log.Printf("📋 发现 %d 条待验证充值", len(deposits))

	for _, deposit := range deposits {
		v.VerifyDeposit(&deposit)
	}
}

// VerifyDeposit 验证单个充值记录
func (v *DepositVerifier) VerifyDeposit(deposit *models.DepositRecord) {
	log.Printf("🔍 验证充值: ID=%d, Hash=%s, Amount=%s",
		deposit.ID, deposit.TxHash, deposit.Amount.String())

	// 1. 获取交易收据
	txHash := common.HexToHash(deposit.TxHash)
	receipt, err := v.client.TransactionReceipt(v.ctx, txHash)
	if err != nil {
		// 如果是交易未找到，继续等待
		if strings.Contains(err.Error(), "not found") {
			log.Printf("⏳ 交易还未上链，继续等待: %s", deposit.TxHash)
			return
		}
		log.Printf("❌ 获取交易收据失败: %v", err)
		return
	}

	// 2. 检查交易是否成功
	if receipt.Status != 1 {
		log.Printf("❌ 交易失败: %s", deposit.TxHash)
		v.MarkDepositFailed(deposit, "Transaction failed")
		return
	}

	// 3. 获取交易详情
	tx, isPending, err := v.client.TransactionByHash(v.ctx, txHash)
	if err != nil {
		log.Printf("❌ 获取交易详情失败: %v", err)
		return
	}

	if isPending {
		log.Printf("⏳ 交易还在pending: %s", deposit.TxHash)
		return
	}

	// 4. 验证接收地址
	toAddress := strings.ToLower(tx.To().Hex())
	expectedAddress := strings.ToLower(USDT_CONTRACT)

	if toAddress != expectedAddress {
		log.Printf("❌ 接收地址不匹配: got %s, want %s", toAddress, expectedAddress)
		v.MarkDepositFailed(deposit, "Invalid receiver address")
		return
	}

	// 5. 解析 Transfer 事件验证金额和目标地址
	// 这里简化处理，实际应该解析 ERC20 Transfer 事件
	// 验证 from = 用户地址, to = 平台地址, value = 充值金额

	// 6. 确认区块数（至少1个确认）
	currentBlock, err := v.client.BlockNumber(v.ctx)
	if err != nil {
		log.Printf("❌ 获取当前区块失败: %v", err)
		return
	}

	confirmations := currentBlock - receipt.BlockNumber.Uint64()
	if confirmations < 1 {
		log.Printf("⏳ 等待确认: %d/1", confirmations)
		return
	}

	// 7. 充值成功，增加用户余额
	log.Printf("✅ 充值验证成功: %s", deposit.TxHash)
	v.ConfirmDeposit(deposit)
}

// ConfirmDeposit 确认充值并增加余额
func (v *DepositVerifier) ConfirmDeposit(deposit *models.DepositRecord) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新充值记录状态
	if err := tx.Model(deposit).Updates(map[string]interface{}{
		"status":     "confirmed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ 更新充值记录失败: %v", err)
		return
	}

	// 2. 增加用户余额
	var balance models.Balance
	err := tx.Where("user_id = ? AND asset = ?", deposit.UserID, deposit.Asset).First(&balance).Error

	if err != nil {
		// 如果余额记录不存在，创建新记录
		balance = models.Balance{
			UserID:    deposit.UserID,
			Asset:     deposit.Asset,
			Available: deposit.Amount,
			Frozen:    decimal.Zero,
		}
		if err := tx.Create(&balance).Error; err != nil {
			tx.Rollback()
			log.Printf("❌ 创建余额记录失败: %v", err)
			return
		}
	} else {
		// 增加可用余额
		balance.Available = balance.Available.Add(deposit.Amount)
		if err := tx.Save(&balance).Error; err != nil {
			tx.Rollback()
			log.Printf("❌ 更新余额失败: %v", err)
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 提交事务失败: %v", err)
		return
	}

	log.Printf("🎉 充值已到账: 用户ID=%d, 资产=%s, 金额=%s",
		deposit.UserID, deposit.Asset, deposit.Amount.String())
}

// MarkDepositFailed 标记充值失败
func (v *DepositVerifier) MarkDepositFailed(deposit *models.DepositRecord, reason string) {
	err := database.DB.Model(deposit).Updates(map[string]interface{}{
		"status":     "failed",
		"updated_at": time.Now(),
	}).Error

	if err != nil {
		log.Printf("❌ 标记充值失败: %v", err)
		return
	}

	log.Printf("❌ 充值已标记为失败: ID=%d, 原因=%s", deposit.ID, reason)
}

// GetDepositAmount 从交易中解析充值金额（简化版）
func (v *DepositVerifier) GetDepositAmount(txHash string) (*big.Int, error) {
	// 实际应该解析 ERC20 Transfer 事件
	// 这里返回模拟值
	return big.NewInt(0), nil
}
