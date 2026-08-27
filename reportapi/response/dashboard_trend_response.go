package response

type DashboardTrendResponse struct {
	WilayahLevel int               `json:"wilayah_level"`
	WilayahInfo  WilayahInfo       `json:"wilayah_info"`
	DataDiambil  string            `json:"data_diambil"`
	MetricKey    string            `json:"metric_key"`
	Label        string            `json:"label"`
	Months       []TrendMonthPoint `json:"months"`
}

type TrendMonthPoint struct {
	MonthLabel string `json:"month_label"`
	TahunLalu  int64  `json:"tahun_lalu"`
	TahunIni   int64  `json:"tahun_ini"`
}
