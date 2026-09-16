package response

import "time"

type DashboardMetricResponse struct {
	ID               int64                            `json:"id"`
	MetricKey        string                           `json:"metric_key"`
	Label            string                           `json:"label"`
	ExpectedTemplate string                           `json:"expected_template"`
	Category         string                           `json:"category"`
	CreatedAt        time.Time                        `json:"created_at"`
	UpdatedAt        time.Time                        `json:"updated_at"`
	Mappings         []DashboardMetricMappingResponse `json:"mappings,omitempty"`
	IsDashboard      *bool                            `json:"is_dashboard"`
}

type DashboardMetricMappingResponse struct {
	ID                int64  `json:"id"`
	DashboardMetricID int64  `json:"dashboard_metric_id"`
	FormID            int64  `json:"form_id"`
	FormFieldID       int64  `json:"form_field_id"`
	FormFieldQuestion string `json:"form_field_question,omitempty"` // enrichment, diisi service kalau di-preload
}
