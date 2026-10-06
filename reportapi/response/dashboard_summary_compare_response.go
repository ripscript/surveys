package response

type DashboardSummaryCompareResponse struct {
	WilayahLevel int                    `json:"wilayah_level"`
	WilayahInfo  WilayahInfo            `json:"wilayah_info"`
	WilayahCount *WilayahCount          `json:"wilayah_count,omitempty"`
	DataDiambil  string                 `json:"data_diambil"`
	Series       []SummaryCompareSeries `json:"series"` // 1 entry = 1 survey (umumnya 2)
	Items        []SummaryCompareItem   `json:"items"`  // 1 entry per metrik (bar), berisi nilai tiap survey
}

type SummaryCompareSurveyOpt struct {
	SurveyID   int64  `json:"survey_id"`
	SurveyName string `json:"survey_name"`
}

type SummaryCompareSeries struct {
	SurveyID   int64  `json:"survey_id"`
	SurveyName string `json:"survey_name"`
	IsLatest   bool   `json:"is_latest"`
}

type SummaryCompareItem struct {
	MetricKey string                      `json:"metric_key"`
	Label     string                      `json:"label"`
	Values    []SummaryCompareMetricValue `json:"values"` // selaras urutan dengan Series
}

type SummaryCompareMetricValue struct {
	SurveyID int64 `json:"survey_id"`
	Value    int64 `json:"value"`
}
