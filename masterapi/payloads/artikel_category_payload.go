package payloads

type ArtikelCategoryPayload struct {
	Nama string `json:"name" validate:"required"`
}

type ArtikelCategoryOptionsPayload struct {
	Q     string  `form:"q" query:"q"`
	Page  int     `form:"page" query:"page"`
	Limit int     `form:"limit" query:"limit"`
	IDs   []int64 `form:"id[]" query:"id[]"`
}
