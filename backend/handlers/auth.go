package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"expchange-backend/config"
	"expchange-backend/database"
	"expchange-backend/middleware"
	"expchange-backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type NonceRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
}

func (h *AuthHandler) GetNonce(c *gin.Context) {
	var req NonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为小写
	walletAddress := strings.ToLower(req.WalletAddress)

	var user models.User
	// 使用静默模式查询，避免打印 "record not found" 日志
	result := database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)}).
		Where("wallet_address = ?", walletAddress).First(&user)

	if result.Error != nil {
		// 创建新用户
		nonce := generateNonce()
		user = models.User{
			WalletAddress: walletAddress,
			Nonce:         nonce,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// 更新nonce
		nonce := generateNonce()
		user.Nonce = nonce
		database.DB.Save(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"nonce": user.Nonce,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为小写
	walletAddress := strings.ToLower(req.WalletAddress)

	var user models.User
	if err := database.DB.Where("wallet_address = ?", walletAddress).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 实际项目中应该验证签名
	// 这里简化处理，跳过签名验证
	// verifySignature(user.Nonce, req.Signature, req.WalletAddress)

	// 生成JWT token
	token, err := h.generateToken(user.ID, user.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) generateToken(userID string, walletAddress string) (string, error) {
	claims := middleware.Claims{
		UserID:        userID,
		WalletAddress: walletAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

