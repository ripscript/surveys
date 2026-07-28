package response

type DatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

type StatistikKewilayahanDatatableResponse struct {
	No                int64  `json:"no"`
	ID                int64  `json:"-" gorm:"column:id"`
	Wilayah           string `json:"wilayah" gorm:"column:wilayah"`
	TotalRT           int64  `json:"total_rt" gorm:"column:total_rt"`
	TotalSudahMengisi int64  `json:"total_sudah_mengisi" gorm:"column:total_sudah_mengisi"`
	TotalBelumMengisi int64  `json:"total_belum_mengisi"`
	TypeWilayah       int    `json:"type_wilayah"`
	IsLeaf            bool   `json:"is_leaf"`
	WilayahCode       string `json:"code_wilayah"`
}

type StatistikKewilayahanAggregate struct {
	TotalWilayah      int64 `json:"total_wilayah"`
	TotalSudahMengisi int64 `json:"total_sudah_mengisi"`
	TotalBelumMengisi int64 `json:"total_belum_mengisi"`
}
