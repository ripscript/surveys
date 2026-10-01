package models

import "time"

type DashboardMetricStatus struct {
	ID                int64     `gorm:"column:id;primaryKey" json:"id"`
	DashboardMetricID int64     `gorm:"column:dashboard_metric_id;not null" json:"dashboard_metric_id"`
	StatusKey         string    `gorm:"column:status_key;type:varchar(100);not null" json:"status_key"`
	Label             string    `gorm:"column:label;type:varchar(191);not null" json:"label"`
	Sequence          int       `gorm:"column:sequence;not null;default:0" json:"sequence"`
	IsActive          bool      `gorm:"column:is_active;not null" json:"is_active"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`

	DashboardMetric *DashboardMetric `gorm:"foreignKey:DashboardMetricID" json:"dashboard_metric,omitempty"`
}

func (DashboardMetricStatus) TableName() string {
	return "dashboard_metric_statuses"
}
