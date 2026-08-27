package response

type DashboardOptionStatusMappingResponse struct {
	ID                       int64  `json:"id"`
	DashboardMetricMappingID int64  `json:"dashboard_metric_mapping_id"`
	AnswerOptionID           int64  `json:"answer_option_id"`
	AnswerOptionLabel        string `json:"answer_option_label,omitempty"` // enrichment dari form_answer_fields.option
	StatusKey                string `json:"status_key"`
}
