package services

import (
	"context"
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/pkg/noncemanager"
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
	ctx          context.Context
	nonceManager *noncemanager.NonceManager
}

// NewWithdrawProcessor 创建提现处理服务（支持多链）
// 注意：现在提现处理通过任务队列调用，不再需要定期轮询
// 提现Worker是单独进程，串行处理，确保Nonce管理器的线程安全
func NewWithdrawProcessor() (*WithdrawProcessor, error) {
	// 注意：不再需要固定的client和privateKey，每次提现时动态创建
	return &WithdrawProcessor{
		ctx:          context.Background(),
		nonceManager: noncemanager.NewNonceManager(database.DB),
	}, nil
}

// ProcessWithdrawal 处理单个提现记录
func (p *WithdrawProcessor) ProcessWithdrawal(withdrawal *models.WithdrawRecord) {
	log.Printf("💸 处理提现: ID=%s, Amount=%s %s, Chain=%s(%d), Address=%s",
		withdrawal.ID, withdrawal.Amount.String(), withdrawal.Asset, withdrawal.Chain, withdrawal.ChainID, withdrawal.Address)

	// 1. 获取链配置
	var chainConfig models.ChainConfig
	if err := database.DB.Where("chain_id = ? AND enabled = ?", withdrawal.ChainID, true).First(&chainConfig).Error; err != nil {
		log.Printf("❌ 获取链配置失败: %v", err)
		p.MarkWithdrawalFailed(withdrawal, fmt.Sprintf("Chain %d not found or disabled", withdrawal.ChainID))
		return
	}

	// 2. 验证私钥已配置
	if chainConfig.PlatformWithdrawPrivateKey == "" {
		log.Printf("❌ 链 %s 未配置提现私钥", chainConfig.ChainName)
		p.MarkWithdrawalFailed(withdrawal, "Private key not configured for this chain")
		return
	}

	// 3. 标记为处理中
	if err := database.DB.Model(withdrawal).Update("status", "processing").Error; err != nil {
		log.Printf("❌ 更新提现状态失败: %v", err)
		return
	}

	// 4. 执行链上转账
	txHash, err := p.TransferUSDT(
		chainConfig.RpcURL,
		chainConfig.UsdtContractAddress,
		chainConfig.PlatformWithdrawPrivateKey,
		withdrawal.Address,
		withdrawal.Amount,
		withdrawal.ChainID,
		chainConfig.UsdtDecimals,
	)
	if err != nil {
		log.Printf("❌ 转账失败: %v", err)
		p.MarkWithdrawalFailed(withdrawal, err.Error())
		return
	}

	log.Printf("✅ 转账成功: Chain=%s, TxHash=%s", chainConfig.ChainName, txHash)

	// 5. 更新提现记录
	p.ConfirmWithdrawal(withdrawal, txHash)
}

// TransferUSDT 执行 USDT 转账（支持多链，线程安全的nonce管理）
func (p *WithdrawProcessor) TransferUSDT(
	rpcURL string,
	usdtContract string,
	privateKeyHex string,
	toAddress string,
	amount decimal.Decimal,
	chainID int,
	usdtDecimals int,
) (string, error) {
	// 1. 连接到链
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RPC %s: %w", rpcURL, err)
	}
	defer client.Close()

	// 2. 加载私钥
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	fromAddressStr := fromAddress.Hex()

	// 3. 使用 NonceManager 获取 nonce（线程安全）
	nonce, releaseNonce, err := p.nonceManager.AcquireNonce(rpcURL, fromAddressStr, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to acquire nonce: %w", err)
	}
	defer releaseNonce() // 确保释放锁

	log.Printf("📝 使用 Nonce: %d (Address: %s, ChainID: %d)", nonce, fromAddressStr, chainID)

	// 4. 获取 gas price
	gasPrice, err := client.SuggestGasPrice(p.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// 5. 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(transferABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// 6. 根据链的USDT精度转换金额
	value := new(big.Int)
	value.SetString(amount.Shift(int32(usdtDecimals)).StringFixed(0), 10)

	// 7. 打包 transfer 函数调用
	data, err := parsedABI.Pack("transfer", common.HexToAddress(toAddress), value)
	if err != nil {
		return "", fmt.Errorf("failed to pack data: %w", err)
	}

	// 8. 创建交易
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(usdtContract),
		big.NewInt(0),  // value = 0 (ERC20 transfer)
		uint64(100000), // gas limit
		gasPrice,
		data,
	)

	// 9. 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(int64(chainID))), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 10. 发送交易
	err = client.SendTransaction(p.ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	log.Printf("✅ 交易已发送: TxHash=%s, Nonce=%d", txHash, nonce)

	return txHash, nil
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

	log.Printf("🎉 提现已完成: 用户ID=%s, 资产=%s, 金额=%s, TxHash=%s",
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

	log.Printf("❌ 提现已标记为失败并解冻资金: ID=%s, 原因=%s", withdrawal.ID, reason)
}
