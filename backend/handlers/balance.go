package handlers

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type BalanceHandler struct{}

func NewBalanceHandler() *BalanceHandler {
	return &BalanceHandler{}
}

func (h *BalanceHandler) GetBalances(c *gin.Context) {
	userID := c.GetUint("user_id")

	var balances []models.Balance
	database.DB.Where("user_id = ?", userID).Find(&balances)

	c.JSON(http.StatusOK, balances)
}

func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID := c.GetUint("user_id")
	asset := c.Param("asset")

	var balance models.Balance
	if err := database.DB.Where("user_id = ? AND asset = ?", userID, asset).First(&balance).Error; err != nil {
		c.JSON(http.StatusOK, models.Balance{
			UserID:    userID,
			Asset:     asset,
			Available: decimal.Zero,
			Frozen:    decimal.Zero,
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}

type DepositRequest struct {
	Asset  string `json:"asset" binding:"required"`
	Amount string `json:"amount" binding:"required"`
	TxHash string `json:"txHash" binding:"required"`
}

// Deposit 创建充值记录并等待验证
func (h *BalanceHandler) Deposit(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// 验证交易hash格式
	if !strings.HasPrefix(req.TxHash, "0x") || len(req.TxHash) != 66 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction hash"})
		return
	}

	// 检查交易hash是否已存在
	var existing models.DepositRecord
	if err := database.DB.Where("tx_hash = ?", strings.ToLower(req.TxHash)).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash already exists"})
		return
	}

	// 创建充值记录（待验证状态）
	deposit := models.DepositRecord{
		UserID: userID,
		Asset:  req.Asset,
		Amount: amount,
		TxHash: strings.ToLower(req.TxHash),
		Status: "pending",
	}

	if err := database.DB.Create(&deposit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deposit record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deposit submitted for verification",
		"deposit": deposit,
	})
}

type WithdrawRequest struct {
	Asset   string `json:"asset" binding:"required"`
	Amount  string `json:"amount" binding:"required"`
	Address string `json:"address" binding:"required"`
}

// Withdraw 创建提现申请并冻结资金
func (h *BalanceHandler) Withdraw(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// 验证地址格式
	if !strings.HasPrefix(req.Address, "0x") || len(req.Address) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet address"})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询用户余额
	var balance models.Balance
	if err := tx.Where("user_id = ? AND asset = ?", userID, req.Asset).First(&balance).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// 检查可用余额是否足够
	if balance.Available.LessThan(amount) {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient available balance"})
		return
	}

	// 冻结资金
	balance.Available = balance.Available.Sub(amount)
	balance.Frozen = balance.Frozen.Add(amount)
	if err := tx.Save(&balance).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to freeze balance"})
		return
	}

	// 创建提现记录（待处理状态）
	withdrawal := models.WithdrawRecord{
		UserID:  userID,
		Asset:   req.Asset,
		Amount:  amount,
		Address: strings.ToLower(req.Address),
		Status:  "pending",
	}

	if err := tx.Create(&withdrawal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal record"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Withdrawal request submitted",
		"withdrawal": withdrawal,
	})
}

// GetDepositRecords 获取充值记录
func (h *BalanceHandler) GetDepositRecords(c *gin.Context) {
	userID := c.GetUint("user_id")

	var deposits []models.DepositRecord
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&deposits)

	c.JSON(http.StatusOK, deposits)
}

// GetWithdrawRecords 获取提现记录
func (h *BalanceHandler) GetWithdrawRecords(c *gin.Context) {
	userID := c.GetUint("user_id")

	var withdrawals []models.WithdrawRecord
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&withdrawals)

	c.JSON(http.StatusOK, withdrawals)
}
