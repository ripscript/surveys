package payloads

import "backend/surveyapi/enums"

type SurveyRequest struct {
	NamaSurvey               string                  `json:"name" validate:"required"`
	TanggalPelaksanaanSurvey enums.PeriodeSurveyType `json:"periode_survey_id" validate:"required,gt=0"`
	TanggalSurveyDimulai     string                  `json:"start_date" validate:"required,datetime=2006-01-02 15:04:05"`
	TanggalSurveyBerakhir    string                  `json:"end_date" validate:"required,datetime=2006-01-02 15:04:05"`
	Deskripsi                string                  `json:"description" validate:"required"`
	Alur                     string                  `json:"flow_detail_code" validate:"required"`
	RespondenSurvey          int                     `json:"responden_survey" validate:"required,oneof=1 2"`
	TingkatPelaksanaan       int                     `json:"tingkat_pelaksanaan_id" validate:"required,gt=0"`
	Kecamatan                int64                   `json:"kecamatan_id"`
	Kelurahan                int64                   `json:"kelurahan_id"`
	RW                       int64                   `json:"rw_id"`
	Surveyor                 []int64                 `json:"surveyor_ids"`
}

type SurveyDatatablePayload struct {
	Search        string `json:"search"`
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	OrderBy       string `json:"order_by"`
	OrderDir      string `json:"order_dir"`
	SurveyDiikuti bool   `json:"survey_diikuti"`
}
