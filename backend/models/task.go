package models

import (
	"expchange-backend/utils"
	"time"

	"gorm.io/gorm"
)

// Task 任务记录
type Task struct {
	ID         string     `gorm:"primaryKey;size:24" json:"id"`
	Type       string     `gorm:"size:50;not null;index" json:"type"`       // generate_trades, generate_klines, verify_deposit, process_withdraw
	Status     string     `gorm:"size:20;not null;index" json:"status"`     // pending, running, completed, failed
	Symbol     string     `gorm:"size:20;index" json:"symbol,omitempty"`    // 交易对符号（用于数据生成任务）
	RecordID   string     `gorm:"size:24;index" json:"record_id,omitempty"` // 关联记录ID（用于充值/提现任务）
	RecordType string     `gorm:"size:20" json:"record_type,omitempty"`     // deposit, withdraw
	StartTime  *time.Time `json:"start_time,omitempty"`                     // 数据生成的开始时间
	EndTime    *time.Time `json:"end_time,omitempty"`                       // 数据生成的结束时间
	Message    string     `gorm:"type:varchar(1000)" json:"message"`
	Error      string     `gorm:"type:text" json:"error,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"` // 任务开始执行时间
	EndedAt    *time.Time `json:"ended_at,omitempty"`   // 任务结束时间
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.GenerateObjectID()
	}
	return nil
}
