package payloads

type HasilSurveyDatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
	Status   string `json:"status" validate:"omitempty,oneof=upcoming ongoing finished"`
}
