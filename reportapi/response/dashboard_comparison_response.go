package response

type DashboardComparisonResponse struct {
	WilayahLevel int                       `json:"wilayah_level"`
	WilayahInfo  WilayahInfo               `json:"wilayah_info"`
	DataDiambil  string                    `json:"data_diambil"`
	SeriesA      DashboardComparisonSeries `json:"series_a"`
	SeriesB      DashboardComparisonSeries `json:"series_b"`
	Items        []DashboardComparisonItem `json:"items"`
}

type DashboardComparisonSeries struct {
	MetricKey string `json:"metric_key"`
	Label     string `json:"label"`
}

type DashboardComparisonItem struct {
	ChildLabel string `json:"child_label"`
	ValueA     int64  `json:"value_a"`
	ValueB     int64  `json:"value_b"`
}
