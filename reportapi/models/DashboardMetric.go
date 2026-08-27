package models

import "time"

type DashboardMetric struct {
	ID               int64     `gorm:"column:id;primaryKey" json:"id"`
	MetricKey        string    `gorm:"column:metric_key;type:varchar(100);unique;not null" json:"metric_key"`
	Label            string    `gorm:"column:label;type:varchar(191);not null" json:"label"`
	ExpectedTemplate string    `gorm:"column:expected_template;type:varchar(20);not null" json:"expected_template"` // number | multiple-choices | long-answer | image-template | maps
	Category         string    `gorm:"column:category;type:varchar(50);not null" json:"category"`                   // summary | sampah | infrastruktur
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`

	Mappings []DashboardMetricMapping `gorm:"foreignKey:DashboardMetricID" json:"mappings,omitempty"`
}

func (DashboardMetric) TableName() string {
	return "dashboard_metrics"
}
