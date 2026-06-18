package models

import "time"

type SurveyExport struct {
	ID                    uint                         `gorm:"primaryKey" json:"id"`
	Name                  string                       `json:"name"`
	StartDate             time.Time                    `json:"start_date"`
	EndDate               time.Time                    `json:"end_date"`
	FlowDetailID          uint                         `json:"flow_detail_id"`
	FlowDetail            FlowDetailExport             `gorm:"foreignKey:FlowDetailID" json:"flow_detail"`
	SurveyDataRespondents []SurveyDataRespondentExport `gorm:"foreignKey:SurveyID" json:"survey_data_respondents"`
}

func (u *SurveyExport) TableName() string {
	return "surveys"
}

type FlowDetailExport struct {
	ID     uint       `gorm:"primaryKey" json:"id"`
	FormID uint       `json:"form_id"`
	Form   FormExport `gorm:"foreignKey:FormID" json:"form"`
}

func (u *FlowDetailExport) TableName() string {
	return "flow_details"
}

type FormExport struct {
	ID     uint              `gorm:"primaryKey" json:"id"`
	Fields []FormFieldExport `gorm:"foreignKey:FormID" json:"fields"`
}

func (u *FormExport) TableName() string {
	return "forms"
}

type FormFieldExport struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FormID   uint   `json:"form_id"`
	Sequence int    `json:"sequence"`
	Question string `json:"question"`
	Template string `json:"template"`
}

func (u *FormFieldExport) TableName() string {
	return "form_fields"
}

type FormAnswerFieldExport struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	FormFieldID uint   `json:"form_field_id"`
	Option      string `json:"option"`
}

func (u *FormAnswerFieldExport) TableName() string {
	return "form_answer_fields"
}

type SurveyDataRespondentExport struct {
	ID                 uint                      `gorm:"primaryKey" json:"id"`
	SurveyID           uint                      `json:"survey_id"`
	RespondentID       uint                      `json:"respondent_id"`
	Respondent         RespondentExport          `gorm:"foreignKey:RespondentID" json:"respondent"`
	FieldDataResponses []FieldDataResponseExport `gorm:"foreignKey:FormResponseID" json:"field_data_responses"`
}

func (u *SurveyDataRespondentExport) TableName() string {
	return "survey_respondents"
}

type RespondentExport struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	Name        string           `json:"name"`
	KecamatanID *uint            `json:"kecamatan_id"`
	KelurahanID *uint            `json:"kelurahan_id"`
	RWID        *uint            `json:"rw_id"`
	RTID        *uint            `json:"rt_id"`
	Kecamatan   *KecamatanExport `gorm:"foreignKey:KecamatanID" json:"kecamatan"`
	Kelurahan   *KelurahanExport `gorm:"foreignKey:KelurahanID" json:"kelurahan"`
	RW          *DataRwExport    `gorm:"foreignKey:RWID" json:"rw"`
	RT          *DataRtExport    `gorm:"foreignKey:RTID" json:"rt"`
}

func (u *RespondentExport) TableName() string {
	return "respondents"
}

type KecamatanExport struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	SubDistrictName string `json:"sub_district_name"`
}

func (u *KecamatanExport) TableName() string {
	return "kecamatans"
}

type KelurahanExport struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	VillageName string          `json:"village_name"`
	KecamatanID uint            `json:"sub_district_id"`
	Kecamatan   KecamatanExport `gorm:"foreignKey:KecamatanID" json:"kecamatan"`
}

func (u *KelurahanExport) TableName() string {
	return "kelurahans"
}

type DataRwExport struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	NamaRw      string          `json:"nama_rw"`
	KelurahanID uint            `json:"kelurahan_id"`
	Kelurahan   KelurahanExport `gorm:"foreignKey:KelurahanID" json:"kelurahan"`
}

func (u *DataRwExport) TableName() string {
	return "data__rws"
}

type DataRtExport struct {
	ID     uint         `gorm:"primaryKey" json:"id"`
	NamaRt string       `json:"nama_rt"`
	RwID   uint         `json:"rw_id"`
	Rw     DataRwExport `gorm:"foreignKey:RwID" json:"rw"`
}

func (u *DataRtExport) TableName() string {
	return "data__rts"
}

type FieldDataResponseExport struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	FormResponseID  uint   `json:"form_response_id"`
	FormFieldID     uint   `json:"form_field_id"`
	Answer          string `json:"answer"`
	LocationAddress string `json:"location_address"`
}

func (u *FieldDataResponseExport) TableName() string {
	return "field_responses"
}
