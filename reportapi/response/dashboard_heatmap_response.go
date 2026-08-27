package response

type DashboardHeatmapResponse struct {
	WilayahLevel int                    `json:"wilayah_level"`
	WilayahInfo  WilayahInfo            `json:"wilayah_info"`
	DataDiambil  string                 `json:"data_diambil"`
	Items        []DashboardHeatmapItem `json:"items"`
}

type DashboardHeatmapItem struct {
	Label   string  `json:"label"`
	GeoName *string `json:"geo_name"`
	Value   int64   `json:"value"`
}
