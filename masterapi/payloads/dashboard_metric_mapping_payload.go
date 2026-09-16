package payloads

type CreateDashboardMetricMappingRequest struct {
	DashboardMetricID int64  `json:"dashboard_metric_id" validate:"required"`
	FormCode          string `json:"form_code" validate:"required"`
	FormFieldID       int64  `json:"form_field_id" validate:"required"`
}

type UpdateDashboardMetricMappingRequest struct {
	FormFieldID int64 `json:"form_field_id" validate:"required"`
}
