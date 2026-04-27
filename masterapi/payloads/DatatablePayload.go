package payloads

type DatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

type DatatablePejabatPayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`

	FNama         *string `json:"f_nama"`
	FPeriodeAwal  *string `json:"f_periode_awal"`
	FPeriodeAkhir *string `json:"f_periode_akhir"`
	FStatus       *int    `json:"f_status"`
	FKecamatan    *int64  `json:"f_kecamatan"`
	FKelurahan    *int64  `json:"f_kelurahan"`
	FRw           *int64  `json:"f_rw"`
	FRt           *int64  `json:"f_rt"`
}
