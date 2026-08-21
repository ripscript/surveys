package payloads

type CreateDashboardMetricMappingRequest struct {
	DashboardMetricID int64  `json:"dashboard_metric_id" validate:"required"`
	FormID            int64  `json:"form_id" validate:"required"`
	FormFieldID       int64  `json:"form_field_id" validate:"required"`
	AnswerOptionID    *int64 `json:"answer_option_id"` // wajib diisi kalau metric.expected_template = multiple-choices, divalidasi di service
}

type UpdateDashboardMetricMappingRequest struct {
	FormFieldID    int64  `json:"form_field_id" validate:"required"`
	AnswerOptionID *int64 `json:"answer_option_id"`
}
