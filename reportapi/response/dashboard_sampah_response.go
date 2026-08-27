package response

type DashboardSampahResponse struct {
	WilayahLevel int           `json:"wilayah_level"`
	WilayahInfo  WilayahInfo   `json:"wilayah_info"`
	WilayahCount *WilayahCount `json:"wilayah_count,omitempty"`
	DataDiambil  string        `json:"data_diambil"`
	TotalUnit    int64         `json:"total_unit"`
	Items        []SampahItem  `json:"items"`
}

type SampahItem struct {
	MetricKey  string  `json:"metric_key"`
	Label      string  `json:"label"`
	Value      int64   `json:"value"`
	Percentage float64 `json:"percentage"`
}
