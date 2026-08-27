package response

type DashboardInfrastrukturResponse struct {
	WilayahLevel int                 `json:"wilayah_level"`
	WilayahInfo  WilayahInfo         `json:"wilayah_info"`
	WilayahCount *WilayahCount       `json:"wilayah_count,omitempty"`
	DataDiambil  string              `json:"data_diambil"`
	Items        []InfrastrukturItem `json:"items"`
}

type InfrastrukturItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}
