package models

import (
	"time"

	"gorm.io/gorm"
)

type LogSurveys struct {
	No           int       `json:"no"`
	ID           int       `json:"id"`
	RespondentID int       `json:"respondent_id"`
	SurveyID     int       `json:"survey_id"`
	Keterangan   string    `json:"keterangan"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsRead       string    `json:"is_read"`
	LogLevel     string    `json:"log_level"`
	NotifFor     string    `json:"notif_for"`
}

func (u *LogSurveys) TableName() string {
	return "log__surveys"
}

type WilayahFilter struct {
	KecamatanID uint
	KelurahanID uint
	RwID        uint
}

type DataRespondentFilter struct {
	Role        string
	KecamatanID uint
	KelurahanID uint
	RwID        uint
}

type GetSurveyParams struct {
	Status        *string
	WilayahFilter WilayahFilter
}

type Survey struct {
	ID                    uint
	HashedID              string `gorm:"-"`
	Name                  string
	FlowDetailID          uint
	StartDate             string
	EndDate               string
	Type                  string
	Status                string
	CreatedAt             string
	UpdatedAt             string
	SurveyRespondents     []SurveyRespondent `gorm:"foreignKey:SurveyID"`
	SurveyDataRespondents []SurveyRespondent `gorm:"foreignKey:SurveyID"`
	ListFlowDetail        []FlowDetail       `gorm:"foreignKey:SurveyID"`
	GetFlowDetail         *FlowDetail        `gorm:"foreignKey:FlowDetailID"`
}

type SurveyRespondent struct {
	ID             uint
	SurveyID       uint
	FieldResponses []FieldResponse `gorm:"foreignKey:SurveyRespondentID"`
}

type SurveyDataRespondent struct {
	ID           uint
	SurveyID     uint
	RespondentID uint
	Respondent   *Respondent `gorm:"foreignKey:RespondentID"`
}

type FieldResponse struct {
	ID                 uint
	SurveyRespondentID uint
}

type FlowDetail struct {
	ID            uint
	SurveyID      uint
	StatusSection int
}

type FlowField struct {
	ID           uint
	FlowDetailID uint
	FormFieldID  uint
	SectionID    uint
	Breakdown    bool
}

type Respondent struct {
	ID          uint
	KecamatanID uint
	KelurahanID uint
	RwID        uint
}

type Respondentss struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name         string         `json:"name" gorm:"type:varchar(70);not null"`
	PhoneNumber  *string        `json:"phone_number" gorm:"type:varchar(191);"`
	Email        *string        `json:"email" gorm:"type:varchar(191);"`
	BLKDID       *int32         `json:"blk_id" gorm:"type:int4;"`
	NIK          *string        `json:"nik" gorm:"type:varchar(191);"`
	KecamatanID  *int64         `json:"kecamatan_id" gorm:"type:int8;"`
	KelurahanID  *int64         `json:"kelurahan_id" gorm:"type:int8;"`
	RWID         *int64         `json:"rw_id" gorm:"type:int8;"`
	RTID         *int64         `json:"rt_id" gorm:"type:int8;"`
	RoleID       *int64         `json:"role_id" gorm:"type:int8;"`
	TanggalLahir *time.Time     `json:"tanggal_lahir" gorm:"type:date;"`
	Alamat       *string        `json:"alamat" gorm:"type:text;"`
	TempatLahir  *string        `json:"tempat_lahir" gorm:"type:varchar(100);"`
	IsBlocked    bool           `json:"is_blocked" gorm:"type:boolean;default:false;"`
	Username     *string        `json:"username" gorm:"type:varchar(255);"`
	CreatedAt    time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *Respondentss) TableName() string {
	return "respondents"
}

type SurveyResult struct {
	Name              interface{} `json:"name"`
	SurveyRespondents interface{} `json:"survey_respondents"`
	StartDate         string      `json:"start_date"`
	EndDate           string      `json:"end_date"`
	GetFlowDetail     interface{} `json:"get_flow_detail"`
	Status            string      `json:"status"`
	ListFlowDetail    interface{} `json:"list_flow_detail"`
	TotalPertanyaan   int64       `json:"total_pertanyaan"`
	Breakdown         string      `json:"breakdown,omitempty"`
	ID                string      `json:"id"`
}
