package payloads

type CreateKecamatanPayload struct {
	NamaKecamatan string  `json:"nama_kecamatan" validate:"required"`
	KodeWilayah   *string `json:"kode_wilayah" validate:"omitempty"`
	Lat           *string `json:"lat" validate:"omitempty"`
	Long          *string `json:"long" validate:"omitempty"`
}

type UpdateKecamatanPayload struct {
	NamaKecamatan string  `json:"nama_kecamatan"`
	KodeWilayah   *string `json:"kode_wilayah"`
	Lat           *string `json:"lat"`
	Long          *string `json:"long"`
}

type CreateKelurahanPayload struct {
	Kecamatan     int     `json:"kecamatan_id" validate:"required"`
	NamaKelurahan string  `json:"nama_kelurahan" validate:"required"`
	KodePos       *string `json:"kode_pos" validate:"omitempty"`
	KodeWilayah   *string `json:"kode_wilayah" validate:"omitempty"`
	Lat           *string `json:"lat" validate:"omitempty"`
	Long          *string `json:"long" validate:"omitempty"`
}

type UpdateKelurahanPayload struct {
	NamaKelurahan string  `json:"nama_kelurahan"`
	KodeWilayah   *string `json:"kode_wilayah"`
	Lat           *string `json:"lat"`
	Long          *string `json:"long"`
}

type UpdateRwPayload struct {
	NamaRw      string  `json:"nama_rw"`
	KodeWilayah *string `json:"kode_wilayah"`
	Lat         *string `json:"lat"`
	Long        *string `json:"long"`
}

type CreateRwPayload struct {
	KelurahanId int64   `json:"kelurahan_id"`
	NamaRw      string  `json:"nama_rw"`
	KodeWilayah *string `json:"kode_wilayah"`
	Lat         *string `json:"lat"`
	Long        *string `json:"long"`
}

type UpdateRtPayload struct {
	NamaRt      string  `json:"nama_rt"`
	KodeWilayah *string `json:"kode_wilayah"`
	Lat         *string `json:"lat"`
	Long        *string `json:"long"`
}

type CreateRtPayload struct {
	RwId        int64   `json:"rw_id"`
	NamaRt      string  `json:"nama_rt"`
	KodeWilayah *string `json:"kode_wilayah"`
	Lat         *string `json:"lat"`
	Long        *string `json:"long"`
}

type KecamatanOptionsPayload struct {
	Q     string  `form:"q" query:"q"`         // Kata kunci pencarian
	Page  int     `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit int     `form:"limit" query:"limit"` // Jumlah data per halaman
	IDs   []int64 `form:"id[]" query:"id[]"`   // Bypass untuk mengambil ID spesifik (misal saat edit data)
}

type KelurahanOptionsPayload struct {
	Q           string  `form:"q" query:"q"`                       // Kata kunci pencarian
	Page        int     `form:"page" query:"page"`                 // Halaman saat ini (untuk lazy load)
	Limit       int     `form:"limit" query:"limit"`               // Jumlah data per halaman
	IDs         []int64 `form:"id[]" query:"id[]"`                 // Bypass untuk mengambil ID spesifik (misal saat edit data)
	KecamatanId int     `form:"kecamatan_id" query:"kecamatan_id"` // Filter berdasarkan kecamatan
}

type RwOptionsPayload struct {
	Q           string  `form:"q" query:"q"`                       // Kata kunci pencarian
	Page        int     `form:"page" query:"page"`                 // Halaman saat ini (untuk lazy load)
	Limit       int     `form:"limit" query:"limit"`               // Jumlah data per halaman
	IDs         []int64 `form:"id[]" query:"id[]"`                 // Bypass untuk mengambil ID spesifik (misal saat edit data)
	KelurahanId int     `form:"kelurahan_id" query:"kelurahan_id"` // Filter berdasarkan kelurahan
}

type RtOptionsPayload struct {
	Q     string  `form:"q" query:"q"`         // Kata kunci pencarian
	Page  int     `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit int     `form:"limit" query:"limit"` // Jumlah data per halaman
	IDs   []int64 `form:"id[]" query:"id[]"`   // Bypass untuk mengambil ID spesifik (misal saat edit data)
	RwId  int     `form:"rw_id" query:"rw_id"` // Filter berdasarkan rw
}
