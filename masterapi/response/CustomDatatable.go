package response

type TabelDataKotaBandungResponse struct {
	No              int64  `json:"no"`
	ID              int64  `json:"-"`
	Code            string `json:"code"`
	NamaKecamatan   string `json:"nama_kecamatan"`
	NamaKelurahan   string `json:"nama_kelurahan"`
	NamaRw          string `json:"nama_rw"`
	NamaRt          string `json:"nama_rt"`
	TotalKelurahan  int64  `json:"total_kelurahan"`
	TotalRw         int64  `json:"total_rw"`
	TotalRt         int64  `json:"total_rt"`
	IsPosibleDetail bool   `json:"posible_detail"`
}

type WilayahSummary struct {
	TotalKecamatan int64 `json:"total_kecamatan"`
	TotalKelurahan int64 `json:"total_kelurahan"`
	TotalRw        int64 `json:"total_rw"`
	TotalRt        int64 `json:"total_rt"`
}
