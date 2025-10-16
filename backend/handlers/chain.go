package handlers

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChainHandler struct{}

func NewChainHandler() *ChainHandler {
	return &ChainHandler{}
}

// GetChains 获取所有链配置（包括禁用的）
func (h *ChainHandler) GetChains(c *gin.Context) {
	var chains []models.ChainConfig
	database.DB.Order("chain_id ASC").Find(&chains)
	c.JSON(http.StatusOK, chains)
}

// GetEnabledChains 获取启用的链配置
func (h *ChainHandler) GetEnabledChains(c *gin.Context) {
	var chains []models.ChainConfig
	database.DB.Where("enabled = ?", true).
		Order("chain_id ASC").
		Find(&chains)
	c.JSON(http.StatusOK, chains)
}

// GetChain 获取单个链配置
func (h *ChainHandler) GetChain(c *gin.Context) {
	chainID := c.Param("id")

	var chain models.ChainConfig
	if err := database.DB.Where("id = ?", chainID).First(&chain).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}

	c.JSON(http.StatusOK, chain)
}

// CreateChain 创建链配置（管理员）- 注意：链名称和Chain ID创建后不可修改
func (h *ChainHandler) CreateChain(c *gin.Context) {
	var chain models.ChainConfig
	if err := c.ShouldBindJSON(&chain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必填字段
	if chain.ChainName == "" || chain.ChainID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chain name and Chain ID are required"})
		return
	}

	// 检查链ID是否已存在
	var existing models.ChainConfig
	if err := database.DB.Where("chain_id = ?", chain.ChainID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chain ID already exists"})
		return
	}

	// 检查链名称是否已存在
	if err := database.DB.Where("chain_name = ?", chain.ChainName).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chain name already exists"})
		return
	}

	if err := database.DB.Create(&chain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chain"})
		return
	}

	c.JSON(http.StatusOK, chain)
}

// UpdateChain 更新链配置（管理员）
func (h *ChainHandler) UpdateChain(c *gin.Context) {
	chainID := c.Param("id")

	var chain models.ChainConfig
	if err := database.DB.Where("id = ?", chainID).First(&chain).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}

	var req models.ChainConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段（ChainName和ChainID不可修改）
	chain.RpcURL = req.RpcURL
	chain.BlockExplorerURL = req.BlockExplorerURL
	chain.UsdtContractAddress = req.UsdtContractAddress
	chain.UsdtDecimals = req.UsdtDecimals
	chain.PlatformDepositAddress = req.PlatformDepositAddress
	
	// 只在提供了私钥且不为空时更新
	if req.PlatformWithdrawPrivateKey != "" {
		chain.PlatformWithdrawPrivateKey = req.PlatformWithdrawPrivateKey
	}

	if err := database.DB.Save(&chain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chain"})
		return
	}

	c.JSON(http.StatusOK, chain)
}

// UpdateChainStatus 更新链启用状态（管理员）
func (h *ChainHandler) UpdateChainStatus(c *gin.Context) {
	chainID := c.Param("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var chain models.ChainConfig
	if err := database.DB.Where("id = ?", chainID).First(&chain).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}

	// 检查是否至少保留一个启用的链
	if !req.Enabled {
		var enabledCount int64
		database.DB.Model(&models.ChainConfig{}).Where("enabled = ? AND id != ?", true, chainID).Count(&enabledCount)
		if enabledCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one chain must be enabled"})
			return
		}
	}

	chain.Enabled = req.Enabled
	database.DB.Save(&chain)

	c.JSON(http.StatusOK, chain)
}

// DeleteChain 删除链配置（管理员）- 建议：不要删除系统预设的链，只禁用即可
func (h *ChainHandler) DeleteChain(c *gin.Context) {
	chainID := c.Param("id")

	var chain models.ChainConfig
	if err := database.DB.Where("id = ?", chainID).First(&chain).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chain not found"})
		return
	}

	// 检查是否有相关的充值/提现记录
	var depositCount int64
	database.DB.Model(&models.DepositRecord{}).Where("chain_id = ?", chain.ChainID).Count(&depositCount)
	if depositCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete chain with existing deposit records. Please disable it instead."})
		return
	}

	var withdrawCount int64
	database.DB.Model(&models.WithdrawRecord{}).Where("chain_id = ?", chain.ChainID).Count(&withdrawCount)
	if withdrawCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete chain with existing withdrawal records. Please disable it instead."})
		return
	}

	database.DB.Delete(&chain)
	c.JSON(http.StatusOK, gin.H{"message": "Chain deleted successfully"})
}
