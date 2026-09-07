package request

type CreateLaporanPayload struct {
	NamaLaporan    string  `json:"nama_laporan" validate:"required"`
	TingkatWilayah int     `json:"tingkat_wilayah" validate:"omitempty,oneof=4 5 6"`
	KecamatanIDs   []int64 `json:"kecamatan_ids"`
	KelurahanIDs   []int64 `json:"kelurahan_ids"`
	RWIDs          []int64 `json:"rw_ids"`
	SurveyIDs      []int64 `json:"survey_ids" validate:"required,min=1"`
}

type ChangeNameLaporanPayload struct {
	Name string `json:"name" validate:"required"`
}

type UpdateLaporanPayload struct {
	DeskripsiHalamanDepan  *string `json:"deskripsi_halaman_depan"`
	BackgroundHalamanDepan *string `json:"background_halaman_depan"`

	DeskripsiHalamanBelakang  *string `json:"deskripsi_halaman_belakang"`
	BackgroundHalamanBelakang *string `json:"background_halaman_belakang"`

	KataPengantar *string `json:"kata_pengantar"`
}
