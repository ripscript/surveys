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

type PeneranganJalanItem struct {
	Label           string  `json:"label"`
	Total           int64   `json:"total"`
	Baik            int64   `json:"baik"`
	Rusak           int64   `json:"rusak"`
	PersenBerfungsi float64 `json:"persen_berfungsi"`
}

type DashboardPeneranganJalanResponse struct {
	WilayahLevel int                   `json:"wilayah_level"`
	WilayahInfo  interface{}           `json:"wilayah_info"`
	WilayahCount interface{}           `json:"wilayah_count"`
	DataDiambil  string                `json:"data_diambil"`
	Items        []PeneranganJalanItem `json:"items"`
}
