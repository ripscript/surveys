package payloads

import "backend/surveyapi/enums"

type TemplateUcapanOptionsPayload struct {
	Q     string  `form:"q" query:"q"`         // Kata kunci pencarian
	Page  int     `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit int     `form:"limit" query:"limit"` // Jumlah data per halaman
	IDs   []int64 `form:"id[]" query:"id[]"`   // Bypass untuk mengambil ID spesifik (misal saat edit data)
}

type TemplateUcapanPayload struct {
	NamaTemplate string                   `json:"nama_template" label:"Nama Template" validate:"required,max=191"`
	Tipe         enums.TypeTemplateUcapan `json:"tipe" label:"Tipe Ucapan" validate:"required,is_type_template"`
	Konten       string                   `json:"konten" label:"Konten" validate:"required"`
}
