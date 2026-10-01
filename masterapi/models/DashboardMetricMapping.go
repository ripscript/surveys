package models

import "time"

type DashboardMetricMapping struct {
	ID                int64     `gorm:"column:id;primaryKey" json:"id"`
	DashboardMetricID int64     `gorm:"column:dashboard_metric_id;not null" json:"dashboard_metric_id"`
	FormID            int64     `gorm:"column:form_id;not null" json:"form_id"`
	FormFieldID       int64     `gorm:"column:form_field_id;not null" json:"form_field_id"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`

	// foreignKey/references diperbaiki eksplisit sesuai catatan KT
	DashboardMetric *DashboardMetric `gorm:"foreignKey:DashboardMetricID;references:ID" json:"dashboard_metric,omitempty"`
	Form            *Form            `gorm:"foreignKey:FormID;references:ID" json:"form,omitempty"`
	FormField       *FormField       `gorm:"foreignKey:FormFieldID;references:ID" json:"form_field,omitempty"`
}

func (DashboardMetricMapping) TableName() string {
	return "dashboard_metric_mappings"
}
