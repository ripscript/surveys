package payloads

import (
	"backend/surveyapi/enums"
	"encoding/json"
)

type SurveyRequest struct {
	NamaSurvey               string                  `json:"name" validate:"required"`
	TanggalPelaksanaanSurvey enums.PeriodeSurveyType `json:"periode_survey_id" validate:"required,gt=0"`
	TanggalSurveyDimulai     string                  `json:"start_date" validate:"required,datetime=2006-01-02 15:04:05"`
	TanggalSurveyBerakhir    string                  `json:"end_date" validate:"required,datetime=2006-01-02 15:04:05"`
	Deskripsi                string                  `json:"description" validate:"required"`
	Alur                     string                  `json:"flow_detail_code" validate:"required"`
	RespondenSurvey          int                     `json:"responden_survey" validate:"required,oneof=1 2"`
	TingkatPelaksanaan       int                     `json:"tingkat_pelaksanaan_id" validate:"required,gt=0"`
	Kecamatan                []int64                 `json:"kecamatan_ids"`
	Kelurahan                []int64                 `json:"kelurahan_ids"`
	RW                       []int64                 `json:"rw_ids"`
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

type ApprovalSurveyRequest struct {
	Notes  *string `json:"notes"`
	Action string  `json:"action" validate:"required,oneof=approved rejected"`
}

type SurveyWilayahDatatablePayload struct {
	Search       string `json:"search"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	OrderBy      string `json:"order_by"`
	OrderDir     string `json:"order_dir"`
	StatusSurvey string `json:"status_survey" validate:"omitempty,oneof=upcoming ongoing finished"`
}

// KEBUTUH SUBMIT HASIL SURVEY =================================================>
type SubmitSurveyRequest struct {
	Answers []SurveyAnswer `json:"answers"`
}

type SurveyAnswer struct {
	QuestionID int    `json:"question_id"`
	Type       string `json:"type"`

	Value json.RawMessage `json:"value"`
}

type MapValue struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type ImageValue struct {
	Slot     int    `json:"slot"`
	FileName string `json:"file_name"`
	Base64   string `json:"base64"`
}

// KEBUTUH SUBMIT HASIL SURVEY =================================================>
