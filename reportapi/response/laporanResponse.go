package response

type LaporanCoverResponse struct {
	DeskripsiHalamanDepan  string  `json:"deskripsi_halaman_depan" validate:"required"`
	BackgroundHalamanDepan *string `json:"background_halaman_depan"`

	DeskripsiHalamanBelakang  string  `json:"deskripsi_halaman_belakang" validate:"required"`
	BackgroundHalamanBelakang *string `json:"background_halaman_belakang"`

	KataPengantar string `json:"kata_pengantar" validate:"required"`
}

type CoverLaporanResponse struct {
	NamaLaporan            string  `json:"nama_laporan"`
	DeskripsiHalamanDepan  *string `json:"deskripsi_halaman_depan" validate:"required"`
	BackgroundHalamanDepan *string `json:"background_halaman_depan"`

	DeskripsiHalamanBelakang  *string `json:"deskripsi_halaman_belakang" validate:"required"`
	BackgroundHalamanBelakang *string `json:"background_halaman_belakang"`

	KataPengantar *string `json:"kata_pengantar" validate:"required"`
}
