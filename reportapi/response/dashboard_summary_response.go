package response

type DashboardSummaryResponse struct {
	WilayahLevel int                             `json:"wilayah_level"`
	WilayahInfo  WilayahInfo                     `json:"wilayah_info"`
	WilayahCount *WilayahCount                   `json:"wilayah_count,omitempty"` // nil untuk level "rt"
	DataDiambil  string                          `json:"data_diambil"`            // timestamp query dijalankan, e.g. "25 Mei 2026 09:30 WIB"
	Metrics      map[string]DashboardMetricValue `json:"metrics"`                 // key = metric_key
}

type WilayahInfo struct {
	Nama   string `json:"nama"`
	Parent string `json:"parent,omitempty"` // e.g. "RT 04 - RW 12 - Kelurahan Dago - Kecamatan Coblong"
}

type WilayahCount struct {
	TotalRT        *int `json:"total_rt,omitempty"`
	TotalRW        *int `json:"total_rw,omitempty"`
	TotalKelurahan *int `json:"total_kelurahan,omitempty"`
	TotalKecamatan *int `json:"total_kecamatan,omitempty"`
}

type DashboardMetricValue struct {
	Label         string `json:"label"`
	CurrentValue  int64  `json:"current_value"`
	PreviousValue int64  `json:"previous_value"`
	Delta         int64  `json:"delta"`
}
