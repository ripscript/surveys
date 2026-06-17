package payloads

type Surveys struct {
	RespondentID int `json:"respondent_id"`
	SurveyID     int `json:"survey_id"`
}

type DetailSurveyKewilayahanDatatablePayload struct {
	Search      string `json:"search"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	OrderBy     string `json:"order_by"`
	OrderDir    string `json:"order_dir"`
	TipeWilayah *int   `json:"tipe_wilayah" validate:"required"`
	KecamatanId *int64 `json:"kecamatan_id,omitempty"`
	KelurahanId *int64 `json:"kelurahan_id,omitempty"`
	RWId        *int64 `json:"rw_id,omitempty"`
}

type DatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}
