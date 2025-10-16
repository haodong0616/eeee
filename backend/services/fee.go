package services

import (
	"expchange-backend/database"
	"expchange-backend/models"

	"github.com/shopspring/decimal"
)

type FeeService struct{}

func NewFeeService() *FeeService {
	return &FeeService{}
}

// 初始化默认手续费配置
func (s *FeeService) InitDefaultFeeConfig() error {
	var count int64
	database.DB.Model(&models.FeeConfig{}).Count(&count)
	
	if count > 0 {
		return nil // 已经初始化过
	}

	// 创建默认手续费配置
	configs := []models.FeeConfig{
		{
			UserLevel:    "normal",
			MakerFeeRate: decimal.NewFromFloat(0.001), // 0.1%
			TakerFeeRate: decimal.NewFromFloat(0.002), // 0.2%
		},
		{
			UserLevel:    "vip1",
			MakerFeeRate: decimal.NewFromFloat(0.0008), // 0.08%
			TakerFeeRate: decimal.NewFromFloat(0.0015), // 0.15%
		},
		{
			UserLevel:    "vip2",
			MakerFeeRate: decimal.NewFromFloat(0.0005), // 0.05%
			TakerFeeRate: decimal.NewFromFloat(0.001),  // 0.1%
		},
		{
			UserLevel:    "vip3",
			MakerFeeRate: decimal.NewFromFloat(0.0002), // 0.02%
			TakerFeeRate: decimal.NewFromFloat(0.0005), // 0.05%
		},
	}

	for _, config := range configs {
		if err := database.DB.Create(&config).Error; err != nil {
			return err
		}
	}

	return nil
}

// 获取用户手续费率
func (s *FeeService) GetUserFeeRate(userLevel string) (*models.FeeConfig, error) {
	var config models.FeeConfig
	err := database.DB.Where("user_level = ?", userLevel).First(&config).Error
	if err != nil {
		// 如果找不到，返回默认配置
		database.DB.Where("user_level = ?", "normal").First(&config)
	}
	return &config, nil
}

// 计算交易手续费
func (s *FeeService) CalculateFee(userLevel string, isMaker bool, tradeAmount decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	config, err := s.GetUserFeeRate(userLevel)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	var feeRate decimal.Decimal
	if isMaker {
		feeRate = config.MakerFeeRate
	} else {
		feeRate = config.TakerFeeRate
	}

	fee := tradeAmount.Mul(feeRate)
	return fee, feeRate, nil
}

// 记录手续费
func (s *FeeService) RecordFee(userID, orderID, tradeID uint, asset string, amount, feeRate decimal.Decimal, orderSide string) error {
	record := models.FeeRecord{
		UserID:    userID,
		OrderID:   orderID,
		TradeID:   tradeID,
		Asset:     asset,
		Amount:    amount,
		FeeRate:   feeRate,
		OrderSide: orderSide,
	}
	return database.DB.Create(&record).Error
}

// 获取用户手续费统计
func (s *FeeService) GetUserFeeStats(userID uint) (map[string]decimal.Decimal, error) {
	var records []models.FeeRecord
	database.DB.Where("user_id = ?", userID).Find(&records)

	stats := make(map[string]decimal.Decimal)
	for _, record := range records {
		if _, exists := stats[record.Asset]; !exists {
			stats[record.Asset] = decimal.Zero
		}
		stats[record.Asset] = stats[record.Asset].Add(record.Amount)
	}

	return stats, nil
}

