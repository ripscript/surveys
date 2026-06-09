package models

import "time"

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
