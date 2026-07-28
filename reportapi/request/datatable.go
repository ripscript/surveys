package request

type SurveyWilayahDatatablePayload struct {
	Search       string `json:"search"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	OrderBy      string `json:"order_by"`
	OrderDir     string `json:"order_dir"`
	StatusSurvey string `json:"status_survey" validate:"omitempty,oneof=upcoming ongoing finished"`
}
