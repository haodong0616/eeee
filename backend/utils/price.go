package utils

import (
	"github.com/shopspring/decimal"
)

// GetPricePrecision 获取价格精度（小数位数）
// 根据价格区间决定小数位，与前端 formatPrice 保持完全一致
func GetPricePrecision(price decimal.Decimal) int32 {
	priceFloat, _ := price.Float64()
	
	if priceFloat >= 1000 {
		return 2 // >= $1,000: 2位小数 (如 8,500.00)
	} else if priceFloat >= 100 {
		return 2 // >= $100: 2位小数 (如 125.50)
	} else if priceFloat >= 1 {
		return 3 // >= $1: 3位小数 (如 45.500)
	} else if priceFloat >= 0.01 {
		return 4 // >= $0.01: 4位小数 (如 0.0850)
	} else if priceFloat >= 0.0001 {
		return 6 // >= $0.0001: 6位小数 (如 0.008500)
	} else {
		return 8 // < $0.0001: 8位小数 (如 0.00008500)
	}
}

// RoundPrice 按价格精度舍入
func RoundPrice(price decimal.Decimal) decimal.Decimal {
	precision := GetPricePrecision(price)
	return price.Round(precision)
}

// GetQuantityByPrice 根据价格区间获取合理的数量范围
// 用于虚拟交易、做市商等场景
func GetQuantityByPrice(price decimal.Decimal) (minQty, maxQty decimal.Decimal) {
	priceFloat, _ := price.Float64()
	
	if priceFloat >= 5000 {
		// 高价币 (>$5000): 0.01-0.06
		minQty = decimal.NewFromFloat(0.01)
		maxQty = decimal.NewFromFloat(0.06)
	} else if priceFloat >= 1000 {
		// 中高价币 ($1000-$5000): 0.1-0.4
		minQty = decimal.NewFromFloat(0.1)
		maxQty = decimal.NewFromFloat(0.4)
	} else if priceFloat >= 100 {
		// 中价币 ($100-$1000): 1-6
		minQty = decimal.NewFromFloat(1)
		maxQty = decimal.NewFromFloat(6)
	} else if priceFloat >= 10 {
		// 中低价币 ($10-$100): 10-40
		minQty = decimal.NewFromFloat(10)
		maxQty = decimal.NewFromFloat(40)
	} else {
		// 低价币 (<$10): 50-200
		minQty = decimal.NewFromFloat(50)
		maxQty = decimal.NewFromFloat(200)
	}
	
	return minQty, maxQty
}

// GetQuantityPrecision 获取数量精度（小数位数）
// 根据价格区间决定数量的小数位
func GetQuantityPrecision(price decimal.Decimal) int32 {
	priceFloat, _ := price.Float64()
	
	if priceFloat >= 1000 {
		return 4 // 高价币: 4位小数
	} else if priceFloat >= 100 {
		return 3 // 中价币: 3位小数
	} else if priceFloat >= 10 {
		return 2 // 中低价币: 2位小数
	} else if priceFloat >= 1 {
		return 2 // 低价币: 2位小数
	} else {
		return 0 // 超低价币: 整数
	}
}

// RoundQuantity 按数量精度舍入
func RoundQuantity(quantity decimal.Decimal, price decimal.Decimal) decimal.Decimal {
	precision := GetQuantityPrecision(price)
	return quantity.Round(precision)
}

// FormatPriceString 格式化价格为字符串（用于日志输出）
func FormatPriceString(price decimal.Decimal) string {
	precision := GetPricePrecision(price)
	return price.StringFixed(precision)
}

// FormatQuantityString 格式化数量为字符串（用于日志输出）
func FormatQuantityString(quantity decimal.Decimal, price decimal.Decimal) string {
	precision := GetQuantityPrecision(price)
	return quantity.StringFixed(precision)
}

