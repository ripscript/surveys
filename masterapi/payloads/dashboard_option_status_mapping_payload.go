package payloads

type CreateDashboardOptionStatusMappingRequest struct {
	DashboardMetricMappingID int64  `json:"dashboard_metric_mapping_id" validate:"required"`
	AnswerOptionID           int64  `json:"answer_option_id" validate:"required"`
	StatusKey                string `json:"status_key" validate:"required,max=100"`
}

type UpdateDashboardOptionStatusMappingRequest struct {
	StatusKey string `json:"status_key" validate:"required,max=100"`
}

type BulkAssignDashboardOptionStatusMappingRequest struct {
	DashboardMetricMappingID int64                        `json:"dashboard_metric_mapping_id" validate:"required"`
	Assignments              []OptionStatusAssignmentItem `json:"assignments" validate:"required,min=1,dive"`
}

type OptionStatusAssignmentItem struct {
	AnswerOptionID int64  `json:"answer_option_id" validate:"required"`
	StatusKey      string `json:"status_key" validate:"required,max=100"`
}

type SyncDashboardOptionStatusMappingRequest struct {
	DashboardMetricMappingID int64 `json:"dashboard_metric_mapping_id" validate:"required"`
	Assignments              []struct {
		AnswerOptionID int64  `json:"answer_option_id" validate:"required"`
		StatusKey      string `json:"status_key" validate:"required"`
	} `json:"assignments" validate:"required,min=1,dive"`
}
