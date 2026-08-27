package models

import "time"

type DashboardMetricMapping struct {
	ID                int64     `gorm:"column:id;primaryKey" json:"id"`
	DashboardMetricID int64     `gorm:"column:dashboard_metric_id;not null" json:"dashboard_metric_id"`
	FormID            int64     `gorm:"column:form_id;not null" json:"form_id"`
	FormFieldID       int64     `gorm:"column:form_field_id;not null" json:"form_field_id"`
	AnswerOptionID    *int64    `gorm:"column:answer_option_id" json:"answer_option_id"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`

	DashboardMetric *DashboardMetric `gorm:"foreignKey:DashboardMetricID" json:"dashboard_metric,omitempty"`
}

func (DashboardMetricMapping) TableName() string {
	return "dashboard_metric_mappings"
}
