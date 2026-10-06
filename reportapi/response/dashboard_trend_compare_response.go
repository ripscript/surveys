package response

type DashboardTrendCompareResponse struct {
	WilayahLevel int                 `json:"wilayah_level"`
	WilayahInfo  WilayahInfo         `json:"wilayah_info"`
	DataDiambil  string              `json:"data_diambil"`
	MetricKey    string              `json:"metric_key"`
	Label        string              `json:"label"`
	MonthLabels  []string            `json:"month_labels"` // sumbu-X bersama, mis. ["Jul 2026", ... "Des 2026"]
	Series       []TrendSurveySeries `json:"series"`       // 1 entry = 1 garis = 1 survey
}

type TrendSurveySeries struct {
	SurveyID   int64               `json:"survey_id"`
	SurveyName string              `json:"survey_name"`
	IsLatest   bool                `json:"is_latest"` // true untuk garis survey terbaru (fixed)
	Points     []TrendComparePoint `json:"points"`
}

type TrendComparePoint struct {
	MonthLabel string `json:"month_label"`
	Value      int64  `json:"value"`
}

type TrendSurveyOption struct {
	SurveyID   int64  `json:"survey_id"`
	SurveyName string `json:"survey_name"`
}
