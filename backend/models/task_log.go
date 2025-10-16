package models

import (
	"expchange-backend/utils"
	"time"

	"gorm.io/gorm"
)

// TaskLog 任务执行日志
type TaskLog struct {
	ID        string    `gorm:"primaryKey;size:24" json:"id"`
	TaskID    string    `gorm:"size:24;not null;index" json:"task_id"`      // 关联的任务ID
	Level     string    `gorm:"size:20;not null" json:"level"`              // info, warning, error
	Stage     string    `gorm:"size:50;not null" json:"stage"`              // 执行阶段：start, processing, completed, failed 等
	Message   string    `gorm:"type:varchar(2000);not null" json:"message"` // 日志消息
	Details   string    `gorm:"type:text" json:"details,omitempty"`         // 详细信息（可选）
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (l *TaskLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = utils.GenerateObjectID()
	}
	return nil
}

