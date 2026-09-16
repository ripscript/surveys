package response

import "time"

type DashboardMetricMappingListItem struct {
	ID                int64     `json:"id"`
	DashboardMetricID int64     `json:"dashboard_metric_id"`
	MetricKey         string    `json:"metric_key"`
	MetricLabel       string    `json:"metric_label"`
	ExpectedTemplate  string    `json:"expected_template"`
	Category          string    `json:"category"`
	FormID            int64     `json:"form_id"`
	FormTitle         string    `json:"form_title"`
	FormCode          string    `json:"form_code"`
	FormFieldID       int64     `json:"form_field_id"`
	FormFieldQuestion string    `json:"form_field_question"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type DashboardMetricMappingDetailResponse struct {
	ID                int64                  `json:"id"`
	DashboardMetric   DashboardMetricSummary `json:"dashboard_metric"`
	Form              FormSummary            `json:"form"`
	FormField         FormFieldSummary       `json:"form_field"`
	OptionStatusCount int                    `json:"option_status_count"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type DashboardMetricSummary struct {
	ID               int64  `json:"id"`
	MetricKey        string `json:"metric_key"`
	Label            string `json:"label"`
	ExpectedTemplate string `json:"expected_template"`
	Category         string `json:"category"`
}

type FormSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

type FormFieldSummary struct {
	ID       int64  `json:"id"`
	Question string `json:"question"`
	Template string `json:"template"`
}
