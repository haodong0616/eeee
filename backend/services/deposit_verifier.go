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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// ERC20 Transfer 事件签名: Transfer(address,address,uint256)
var transferEventSignature = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// DepositVerifier 充值验证服务（支持多链）
type DepositVerifier struct {
	ctx context.Context
}

// NewDepositVerifier 创建充值验证服务
// 注意：现在充值验证通过任务队列调用，不再需要定期轮询
func NewDepositVerifier() (*DepositVerifier, error) {
	return &DepositVerifier{
		ctx: context.Background(),
	}, nil
}

// VerifyDeposit 验证单个充值记录（支持多链）
// 返回 error 表示需要重试，nil 表示验证完成（成功或失败）
func (v *DepositVerifier) VerifyDeposit(deposit *models.DepositRecord) error {
	log.Printf("🔍 验证充值: ID=%s, Chain=%s(%d), Hash=%s, Amount=%s",
		deposit.ID, deposit.Chain, deposit.ChainID, deposit.TxHash, deposit.Amount.String())

	// 1. 获取链配置
	var chainConfig models.ChainConfig
	if err := database.DB.Where("chain_id = ? AND enabled = ?", deposit.ChainID, true).First(&chainConfig).Error; err != nil {
		log.Printf("❌ 获取链配置失败 (ChainID=%d): %v", deposit.ChainID, err)
		v.MarkDepositFailed(deposit, fmt.Sprintf("Chain %d not found or disabled", deposit.ChainID))
		return nil // 不重试
	}

	// 2. 连接到对应的链
	client, err := ethclient.Dial(chainConfig.RpcURL)
	if err != nil {
		log.Printf("❌ 连接RPC失败 (%s): %v", chainConfig.ChainName, err)
		return fmt.Errorf("RETRY_LATER: RPC connection failed") // 重试
	}
	defer client.Close()

	// 3. 获取交易收据
	txHash := common.HexToHash(deposit.TxHash)
	receipt, err := client.TransactionReceipt(v.ctx, txHash)
	if err != nil {
		// 如果是交易未找到，返回特殊错误让任务队列重试
		if strings.Contains(err.Error(), "not found") {
			log.Printf("⏳ 交易还未上链，稍后重试: %s", deposit.TxHash)
			// 返回特殊错误，任务队列会识别并重试
			return fmt.Errorf("RETRY_LATER: transaction not found yet")
		}
		log.Printf("❌ 获取交易收据失败: %v", err)
		v.MarkDepositFailed(deposit, fmt.Sprintf("Failed to get receipt: %v", err))
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	// 4. 检查交易是否成功
	if receipt.Status != 1 {
		log.Printf("❌ 交易失败: %s", deposit.TxHash)
		v.MarkDepositFailed(deposit, "Transaction failed")
		return nil // 不重试
	}

	// 5. 获取交易详情
	tx, isPending, err := client.TransactionByHash(v.ctx, txHash)
	if err != nil {
		log.Printf("❌ 获取交易详情失败: %v", err)
		return fmt.Errorf("RETRY_LATER: failed to get transaction") // 重试
	}

	if isPending {
		log.Printf("⏳ 交易还在pending，稍后重试: %s", deposit.TxHash)
		return fmt.Errorf("RETRY_LATER: transaction pending") // 重试
	}

	// 6. 验证接收合约地址（必须是USDT合约）
	toAddress := strings.ToLower(tx.To().Hex())
	expectedContract := strings.ToLower(chainConfig.UsdtContractAddress)

	if toAddress != expectedContract {
		log.Printf("❌ 合约地址不匹配: got %s, want %s", toAddress, expectedContract)
		v.MarkDepositFailed(deposit, "Invalid contract address")
		return nil // 不重试
	}

	// 7. 解析 Transfer 事件验证金额和目标地址
	transferFound := false
	for _, vLog := range receipt.Logs {
		// 检查是否是 Transfer 事件
		if len(vLog.Topics) != 3 || vLog.Topics[0] != transferEventSignature {
			continue
		}

		// 解析 Transfer 事件: Transfer(from, to, value)
		// Topics[0] = event signature
		// Topics[1] = from address
		// Topics[2] = to address
		// Data = value (amount)

		toAddr := common.HexToAddress(vLog.Topics[2].Hex())
		platformAddr := common.HexToAddress(chainConfig.PlatformDepositAddress)

		// 检查接收地址是否是平台地址
		if strings.ToLower(toAddr.Hex()) != strings.ToLower(platformAddr.Hex()) {
			continue
		}

		// 解析转账金额
		amount := new(big.Int).SetBytes(vLog.Data)

		// 转换为 decimal，使用链配置的精度
		actualAmount := decimal.NewFromBigInt(amount, -int32(chainConfig.UsdtDecimals))

		log.Printf("🔍 解析到Transfer: to=%s, amount=%s USDT", toAddr.Hex(), actualAmount.String())

		// 验证金额是否匹配（允许小数点后6位的误差）
		expectedAmount := deposit.Amount
		if !actualAmount.Round(6).Equal(expectedAmount.Round(6)) {
			log.Printf("❌ 金额不匹配: got %s, want %s", actualAmount.String(), expectedAmount.String())
			v.MarkDepositFailed(deposit, fmt.Sprintf("Amount mismatch: got %s, want %s", actualAmount.String(), expectedAmount.String()))
			return nil // 不重试
		}

		transferFound = true
		break
	}

	if !transferFound {
		log.Printf("❌ 未找到有效的Transfer事件到平台地址")
		v.MarkDepositFailed(deposit, "No valid transfer event found")
		return nil // 不重试
	}

	// 8. 确认区块数（至少1个确认）
	currentBlock, err := client.BlockNumber(v.ctx)
	if err != nil {
		log.Printf("❌ 获取当前区块失败: %v", err)
		return fmt.Errorf("RETRY_LATER: failed to get block number") // 重试
	}

	confirmations := currentBlock - receipt.BlockNumber.Uint64()
	if confirmations < 1 {
		log.Printf("⏳ 等待确认: %d/1，稍后重试", confirmations)
		return fmt.Errorf("RETRY_LATER: waiting for confirmations") // 重试
	}

	// 9. 充值成功，增加用户余额
	log.Printf("✅ 充值验证成功: Chain=%s, TxHash=%s, Confirmations=%d",
		chainConfig.ChainName, deposit.TxHash, confirmations)
	v.ConfirmDeposit(deposit)
	return nil // 验证完成
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

	log.Printf("🎉 充值已到账: 用户ID=%s, 链=%s, 资产=%s, 金额=%s",
		deposit.UserID, deposit.Chain, deposit.Asset, deposit.Amount.String())
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

	log.Printf("❌ 充值已标记为失败: ID=%s, Chain=%s, 原因=%s", deposit.ID, deposit.Chain, reason)
}
