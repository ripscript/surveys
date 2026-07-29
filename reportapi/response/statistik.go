package response

type ExportStatistikRow struct {
	Kecamatan      string `json:"kecamatan"`
	Kelurahan      string `json:"kelurahan"`
	Rw             string `json:"rw"`
	JumlahRt       int    `json:"jumlah_rt"`
	RtSudahMengisi int    `json:"rt_sudah_mengisi"`
}
