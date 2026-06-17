package models

import "time"

// Survey status constants
const (
	SurveyStatusUpcoming = "upcoming"
	SurveyStatusOngoing  = "ongoing"
	SurveyStatusFinished = "finished"
)

// func (u *DataRt) TableName() string {
// 	return "data__rts"
// }

func (u *SurveyWilayah) TableName() string {
	return "survey_wilayahs"
}

func (u *Surveyss) TableName() string {
	return "surveys"
}

func (u *DataRw) TableName() string {
	return "data__rws"
}

func (u *Respondentsss) TableName() string {
	return "respondents"
}

func (u *Role) TableName() string {
	return "role"
}

type Surveyss struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	FlowDetailID   uint      `json:"flow_detail_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	CreatedBy      uint      `json:"created_by"`
	Deskripsi      string    `json:"deskripsi"`
	TingkatWilayah int       `json:"tingkat_wilayah"`
	ApprovalSurvey *string   `json:"approval_survey"`
	AlasanReject   *string   `json:"alasan_reject"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	FlowDetail        *FlowDetails        `gorm:"foreignKey:FlowDetailID" json:"flow_detail,omitempty"`
	SurveyRespondents []SurveyRespondents `gorm:"foreignKey:SurveyID" json:"survey_respondents,omitempty"`
	SurveyWilayah     []SurveyWilayah     `gorm:"foreignKey:SurveyID" json:"survey_wilayah,omitempty"`
	SurveySurveyor    []SurveySurveyor    `gorm:"foreignKey:SurveyID" json:"survey_surveyor,omitempty"`

	// Virtual field (not stored in DB)
	HashedID string `gorm:"-" json:"id,omitempty"`
}

type SurveyRespondents struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	SurveyID       uint   `json:"survey_id"`
	RespondentID   uint   `json:"respondent_id"`
	Status         string `json:"status"`
	StatusApproval string `json:"status_approval"`

	// Relations
	Respondent     *Respondent     `gorm:"foreignKey:RespondentID" json:"respondent,omitempty"`
	FieldResponses []FieldResponse `gorm:"foreignKey:SurveyRespondentID" json:"field_responses,omitempty"`
}

type SurveyWilayah struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	SurveyID       uint `json:"survey_id"`
	TingkatWilayah int  `json:"tingkat_wilayah"`
	KecamatanID    uint `json:"kecamatan_id"`
	KelurahanID    uint `json:"kelurahan_id"`
	RwID           uint `json:"rw_id"`
}

type SurveySurveyor struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	SurveyID     uint `json:"survey_id"`
	RespondentID uint `json:"respondent_id"`
}

type FlowDetails struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	StatusSection int          `json:"status_section"`
	FlowFields    []FlowFields `gorm:"foreignKey:FlowDetailID" json:"flow_fields,omitempty"`
}

type FlowFields struct {
	ID           uint  `gorm:"primaryKey" json:"id"`
	FormID       uint  `json:"form_id"`
	Sequence     int   `json:"sequence"`
	ChildID      *uint `json:"child_id"`
	FlowDetailID uint  `json:"flow_detail_id"`
	FormFieldID  uint  `json:"form_field_id"`
	SectionID    *uint `json:"section_id"`
	GroupChildID *uint `json:"group_child_id"`
	GroupID      *uint `json:"group_id"`
	Breakdown    bool  `json:"breakdown"`

	// Relations
	FormField *FormField   `gorm:"foreignKey:FormFieldID" json:"form_field,omitempty"`
	Section   *FlowSection `gorm:"foreignKey:SectionID" json:"section,omitempty"`
}

type FormField struct {
	ID uint `gorm:"primaryKey" json:"id"`
}

type FlowSection struct {
	ID uint `gorm:"primaryKey" json:"id"`
}

type FieldResponses struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	FormFieldID uint `json:"form_field_id"`
}

type Respondentsss struct {
	ID          uint  `gorm:"primaryKey" json:"id"`
	RoleID      uint  `json:"role_id"`
	KecamatanID uint  `json:"kecamatan_id"`
	KelurahanID uint  `json:"kelurahan_id"`
	RwID        uint  `json:"rw_id"`
	Role        *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Kecamatan struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	SubDistrictName string `json:"sub_district_name"`
}

type Kelurahan struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	VillageName string `json:"village_name"`
	KecamatanID uint   `json:"kecamatan_id"`
}

type DataRw struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	KelurahanID uint `json:"kelurahan_id"`
}

type SurveyResults struct {
	Name            string       `json:"name"`
	StartDate       time.Time    `json:"start_date"`
	EndDate         time.Time    `json:"end_date"`
	GetFlowDetail   *FlowDetails `json:"get_flow_detail"`
	Status          string       `json:"status"`
	ListFlowDetail  *FlowDetails `json:"list_flow_detail"`
	TotalPertanyaan int          `json:"total_pertanyaan"`
	Breakdown       string       `json:"breakdown,omitempty"`
	ID              string       `json:"id"`
}
