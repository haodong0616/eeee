package models

import (
	"expchange-backend/utils"
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置
type SystemConfig struct {
	ID          string    `gorm:"primaryKey;size:24" json:"id"`
	Key         string    `gorm:"uniqueIndex;size:100;not null" json:"key"` // 配置键
	Value       string    `gorm:"size:500;not null" json:"value"`           // 配置值
	Description string    `gorm:"size:200" json:"description"`              // 配置说明
	Category    string    `gorm:"size:50;index" json:"category"`            // 配置分类
	ValueType   string    `gorm:"size:20" json:"value_type"`                // 值类型: string, number, boolean
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *SystemConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = utils.GenerateObjectID()
	}
	return nil
}



