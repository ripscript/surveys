package request

type StatistikKewilayahanPayload struct {
	SurveyId    int   `json:"survey_id"`
	TypeWilayah int   `json:"type_wilayah"`
	ParentId    int64 `json:"parent_id"`

	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}
