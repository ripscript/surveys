package response

type DashboardTopWilayahResponse struct {
	DataDiambil string                    `json:"data_diambil"`
	Items       []DashboardTopWilayahItem `json:"items"`
}

type DashboardTopWilayahItem struct {
	KecamatanLabel string `json:"kecamatan_label"`
	TotalRT        int64  `json:"total_rt"`
	JumlahData     int64  `json:"jumlah_data"`
}
