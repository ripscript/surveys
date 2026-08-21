package payloads

type CreateDashboardMetricRequest struct {
	MetricKey        string `json:"metric_key" validate:"required,max=100"`
	Label            string `json:"label" validate:"required,max=191"`
	ExpectedTemplate string `json:"expected_template" validate:"required,oneof=number long-answer multiple-choices image-template maps"`
	Category         string `json:"category" validate:"required,max=50"`
}

type UpdateDashboardMetricRequest struct {
	Label            string `json:"label" validate:"required,max=191"`
	ExpectedTemplate string `json:"expected_template" validate:"required,oneof=number long-answer multiple-choices image-template maps"`
	Category         string `json:"category" validate:"required,max=50"`
}
