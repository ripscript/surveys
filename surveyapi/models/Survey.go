package models

import (
	"time"
)

type Survey struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"column:name;type:text" json:"name"`
	FlowDetailID int64     `gorm:"column:flow_detail_id" json:"flow_detail_id"`
	StartDate    time.Time `gorm:"column:start_date" json:"start_date"`
	EndDate      time.Time `gorm:"column:end_date" json:"end_date"`
	Type         string    `gorm:"column:type;type:varchar(255)" json:"type"`

	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	Status         string    `gorm:"column:status;type:varchar(191)" json:"status"`
	CreatedBy      int64     `gorm:"column:created_by" json:"created_by"`
	Deskripsi      string    `gorm:"column:deskripsi;type:text" json:"deskripsi"`
	ApprovalSurvey string    `gorm:"column:approval_survey;type:varchar(255)" json:"approval_survey"`
	AlasanReject   *string   `gorm:"column:alasan_reject;type:text" json:"alasan_reject"`

	IsRepeated *bool `gorm:"column:is_repeated;type:boolean;default:false" json:"is_repeated"`
}

func (u *Survey) TableName() string {
	return "surveys"
}

type SurveyDetailResponse struct {
	SurveyCode               string  `json:"survey_code"`
	NamaSurvey               string  `json:"name"`
	TanggalPelaksanaanSurvey int32   `json:"periode_survey_id"`
	TanggalSurveyDimulai     string  `json:"start_date"`
	TanggalSurveyBerakhir    string  `json:"end_date"`
	Deskripsi                string  `json:"description"`
	Alur                     string  `json:"flow_detail_code"`
	RespondenSurvey          int     `json:"responden_survey"`
	TingkatPelaksanaan       *int    `json:"tingkat_pelaksanaan_id"`
	Kecamatan                []int64 `json:"kecamatan_ids"`
	Kelurahan                []int64 `json:"kelurahan_ids"`
	RW                       []int64 `json:"rw_ids"`
	Surveyor                 []int64 `json:"surveyor_ids"`
	IsRepeated               *bool   `json:"is_repeated"`
}

type SurveyDatatableResponse struct {
	No              int64     `json:"no"`
	ID              int       `json:"-" gorm:"column:id"`
	SurveyCode      string    `json:"survey_code" gorm:"column:survey_code"`
	SurveyName      string    `json:"survey_name" gorm:"column:survey_name"`
	SurveyDimulai   time.Time `json:"start_date" gorm:"column:start_date"`
	SurveyBerakhir  time.Time `json:"end_date" gorm:"column:end_date"`
	AlurName        string    `json:"flow_name" gorm:"column:flow_name"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy       int       `json:"-" gorm:"column:created_by"`
	CreatedByName   string    `json:"created_by_name" gorm:"column:created_by_name"`
	IsMyOwn         bool      `json:"is_my_own"`
	TotalResponden  int       `json:"total_responden" gorm:"column:total_responden"`
	Status          string    `json:"status" gorm:"column:status"`
	Approval        string    `json:"approval_survey" gorm:"column:approval_survey"`
	PosibleApproval bool      `json:"posible_approval" gorm:"column:posible_approval"`
}

type SurveyWilayahDatatableResponse struct {
	No             int64     `json:"no"`
	ID             int       `json:"id" gorm:"column:id"`
	SurveyCode     string    `json:"survey_code" gorm:"column:survey_code"`
	SurveyName     string    `json:"survey_name" gorm:"column:survey_name"`
	SurveyDimulai  time.Time `json:"start_date" gorm:"column:start_date"`
	SurveyBerakhir time.Time `json:"end_date" gorm:"column:end_date"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy      int       `json:"-" gorm:"column:created_by"`
	CreatedByName  string    `json:"created_by_name" gorm:"column:created_by_name"`
	IsMyOwn        bool      `json:"is_my_own"`
	Status         string    `json:"status" gorm:"column:status"`
	Approval       string    `json:"approval_survey" gorm:"column:approval_survey"`
	PosibleDetail  bool      `json:"posible_detail"`
	PosibleProcess bool      `json:"posible_process"`
	PosibleHistory bool      `json:"posible_history"`
	SurveyIsDone   bool      `json:"survey_is_done" gorm:"column:survey_is_done"`

	StatusKeterisian string  `json:"status_keterisian" gorm:"column:status_keterisian"`
	JumlahSoalTerisi int     `json:"jumlah_soal_terisi" gorm:"column:jumlah_soal_terisi"`
	JumlahTotalSoal  int     `json:"jumlah_total_soal" gorm:"column:jumlah_total_soal"`
	StatusRespondent *string `json:"status_respondent" gorm:"column:status_respondent"`
	CodeWilayah      *string `json:"code_wilayah" gorm:"column:code_wilayah"`
}

type SurveyKewilayahanDatatableResponse struct {
	No             int64     `json:"no"`
	ID             int       `json:"-" gorm:"column:id"`
	SurveyCode     string    `json:"survey_code" gorm:"column:survey_code"`
	SurveyName     string    `json:"survey_name" gorm:"column:survey_name"`
	SurveyDimulai  time.Time `json:"start_date" gorm:"column:start_date"`
	SurveyBerakhir time.Time `json:"end_date" gorm:"column:end_date"`
	Status         string    `json:"status" gorm:"column:status"`
}

type DetailSurveyKewilayahanDatatableResponse struct {
	ID                   int    `json:"-" gorm:"column:id"`
	WilayahId            int64  `json:"wilayah_id" gorm:"column:wilayah_id"`
	NamaWilayah          string `json:"nama_wilayah" gorm:"column:nama_wilayah"`
	StatusWilayah        string `json:"status_wilayah" gorm:"column:status_wilayah"`
	PosibleDetail        bool   `json:"posible_detail"`
	PosiblePreviewSurvey bool   `json:"posible_preview_survey"`
}

type SurveyPreviewSection struct {
	SectionId              int     `json:"-"`
	SectionCode            *string `json:"section_code"`
	SectionName            *string `json:"section_name"`
	TotalRequiredQuestions int     `json:"total_required_questions"`
	TotalOptionalQuestions int     `json:"total_optional_questions"`
}
