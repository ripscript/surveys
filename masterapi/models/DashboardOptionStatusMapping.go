package models

import "time"

type DashboardOptionStatusMapping struct {
	ID                       int64     `gorm:"column:id;primaryKey" json:"id"`
	DashboardMetricMappingID int64     `gorm:"column:dashboard_metric_mapping_id;not null" json:"dashboard_metric_mapping_id"`
	AnswerOptionID           int64     `gorm:"column:answer_option_id;not null" json:"answer_option_id"`
	StatusKey                string    `gorm:"column:status_key;type:varchar(100);not null" json:"status_key"`
	CreatedAt                time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at" json:"updated_at"`

	DashboardMetricMapping *DashboardMetricMapping `gorm:"foreignKey:DashboardMetricMappingID" json:"dashboard_metric_mapping,omitempty"`
}

func (DashboardOptionStatusMapping) TableName() string {
	return "dashboard_option_status_mappings"
}
