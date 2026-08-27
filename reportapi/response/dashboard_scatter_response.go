package response

type DashboardScatterResponse struct {
	WilayahLevel int            `json:"wilayah_level"`
	WilayahInfo  WilayahInfo    `json:"wilayah_info"`
	DataDiambil  string         `json:"data_diambil"`
	XAxisLabel   string         `json:"x_axis_label"`
	YAxisLabel   string         `json:"y_axis_label"`
	ThresholdX   float64        `json:"threshold_x"`
	ThresholdY   float64        `json:"threshold_y"`
	Points       []ScatterPoint `json:"points"`
}

type ScatterPoint struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}
