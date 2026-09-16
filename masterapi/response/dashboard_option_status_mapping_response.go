package response

type DashboardOptionStatusMappingResponse struct {
	ID                       int64  `json:"id"`
	DashboardMetricMappingID int64  `json:"dashboard_metric_mapping_id"`
	AnswerOptionID           int64  `json:"answer_option_id"`
	AnswerOptionLabel        string `json:"answer_option_label,omitempty"` // enrichment dari form_answer_fields.option
	StatusKey                string `json:"status_key"`
}

type OptionStatusItem struct {
	DashboardOptionStatusID *int64  `json:"dashboard_option_status_id"`
	AnswerOptionID          int64   `json:"answer_option_id"`
	OptionLabel             string  `json:"option_label"`
	StatusKey               *string `json:"status_key"`
}

type DashboardOptionStatusByMappingResponse struct {
	DashboardMetricMappingID int64              `json:"dashboard_metric_mapping_id"`
	FormFieldID              int64              `json:"form_field_id"`
	FormFieldQuestion        string             `json:"form_field_question"`
	Options                  []OptionStatusItem `json:"options"`
}
