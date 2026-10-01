package response

import "time"

type DashboardMetricStatusResponse struct {
	ID        int64  `json:"id"`
	StatusKey string `json:"status_key"`
	Label     string `json:"label"`
	Sequence  int    `json:"sequence"`
	IsActive  bool   `json:"is_active"`
}

type DashboardMetricResponse struct {
	ID               int64                           `json:"id"`
	MetricKey        string                          `json:"metric_key"`
	Label            string                          `json:"label"`
	ExpectedTemplate string                          `json:"expected_template"`
	Category         string                          `json:"category"`
	IsLocked         bool                            `json:"is_locked"`
	CreatedAt        time.Time                       `json:"created_at"`
	UpdatedAt        time.Time                       `json:"updated_at"`
	Statuses         []DashboardMetricStatusResponse `json:"statuses,omitempty"`
}

type PemetaanMetrikSurveyDatatableResponse struct {
	No          int64  `json:"no"`
	ID          int    `json:"-" gorm:"column:id"`
	SurveyCode  string `json:"survey_code" gorm:"column:survey_code"`
	SurveyName  string `json:"survey_name" gorm:"column:survey_name"`
	AlurName    string `json:"flow_name" gorm:"column:flow_name"`
	FlowVersion int    `json:"flow_version" gorm:"column:flow_version"`
	Status      string `json:"status" gorm:"column:status"`
	IsMapped    bool   `json:"mapping_status" gorm:"column:mapping_status"`
}
