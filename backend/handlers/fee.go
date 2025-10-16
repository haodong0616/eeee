package handlers

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeeHandler struct {
	feeService *services.FeeService
}

func NewFeeHandler() *FeeHandler {
	return &FeeHandler{
		feeService: services.NewFeeService(),
	}
}

// 获取手续费配置列表
func (h *FeeHandler) GetFeeConfigs(c *gin.Context) {
	var configs []models.FeeConfig
	database.DB.Order("user_level").Find(&configs)
	c.JSON(http.StatusOK, configs)
}

// 获取用户手续费统计
func (h *FeeHandler) GetUserFeeStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	stats, err := h.feeService.GetUserFeeStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fee stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// 获取用户手续费记录
func (h *FeeHandler) GetUserFeeRecords(c *gin.Context) {
	userID := c.GetUint("user_id")

	var records []models.FeeRecord
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&records)

	c.JSON(http.StatusOK, records)
}

// 管理员：获取所有手续费记录
func (h *FeeHandler) GetAllFeeRecords(c *gin.Context) {
	var records []models.FeeRecord
	database.DB.Order("created_at DESC").Limit(500).Find(&records)

	c.JSON(http.StatusOK, records)
}

// 管理员：更新用户等级
func (h *FeeHandler) UpdateUserLevel(c *gin.Context) {
	userID := c.Param("id")
	level := c.PostForm("level")

	if level != "normal" && level != "vip1" && level != "vip2" && level != "vip3" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user level"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.UserLevel = level
	database.DB.Save(&user)

	c.JSON(http.StatusOK, user)
}

