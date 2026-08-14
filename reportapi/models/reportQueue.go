package models

import (
	"time"
)

type ReportQueue struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	UserID       uint64     `gorm:"not null;column:user_id"`
	LaporanID    uint64     `gorm:"not null;column:laporan_id"`
	Status       string     `gorm:"type:varchar(20);not null;default:'pending';column:status"`
	DocumentID   *string    `gorm:"type:varchar(255);column:document_id"`
	ErrorMessage *string    `gorm:"type:text;column:error_message"`
	CreatedAt    time.Time  `gorm:"not null;column:created_at"`
	StartedAt    *time.Time `gorm:"column:started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
}

func (ReportQueue) TableName() string {
	return "report_queue"
}
