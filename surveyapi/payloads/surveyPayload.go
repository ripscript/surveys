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

// SUBMIT SURVEY RESPONDEN ========================================>
type SubmitSurveyPayload struct {
	Answers []AnswerPayload `json:"answers" validate:"required,dive"`
}

type AnswerPayload struct {
	QuestionID int    `json:"question_id" validate:"required"`
	Type       string `json:"type" validate:"required,oneof=long-answer number multiple-choices image-template maps"`

	GroupID *int `json:"group_id,omitempty"`

	// Gunakan pointer agar kita bisa membedakan antara nilai kosong/nol dan tidak diisi
	ValueString   *string                `json:"value_string,omitempty"`    // Untuk long-answer
	ValueNumber   *int64                 `json:"value_number,omitempty"`    // Untuk number
	ValueOptionID *int                   `json:"value_option_id,omitempty"` // Untuk multiple-choices
	ValueImages   []string               `json:"value_images,omitempty"`    // Untuk image-template (Array of Base64)
	ValueMaps     []MapCoordinatePayload `json:"value_maps,omitempty"`      // Untuk maps
}

type MapCoordinatePayload struct {
	Lat float64 `json:"lat" validate:"required"`
	Lng float64 `json:"lng" validate:"required"`
}

// ========================================>

type UpdateSurveyStatusPayload struct {
	Status *int `json:"status" validate:"required,oneof=0 1 2"`
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

// KEBUTUHAN VERIFIKASI ATAU VALIDASI SURVEY ===============================>
type VerifyAllSurveyAnswersPayload struct {
	Answers []VerifySurveyAnswerItem `json:"answers" validate:"required,dive"`
}

type VerifySurveyAnswerItem struct {
	QuestionID    int           `json:"question_id" validate:"required"`
	Type          string        `json:"type" validate:"required"`
	GroupID       *int          `json:"group_id"`
	ValueString   *string       `json:"value_string"`
	ValueNumber   *int          `json:"value_number"`
	ValueOptionID *int          `json:"value_option_id"`
	ValueMaps     []interface{} `json:"value_maps"`
	ValueImages   []string      `json:"value_images"`
}

type RejectAllSurveyAnswersPayload struct {
	FlaggedQuestionIDs []int `json:"flagged_question_ids" validate:"required,min=1"`
}
