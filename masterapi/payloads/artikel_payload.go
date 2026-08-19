package payloads

type ArtikelPayload struct {
	Judul    string `json:"title" validate:"required,min=3,max=191"`
	Kategori int64  `json:"category_id" validate:"required"`
}

type ArtikelContentPayload struct {
	Description *string `json:"description"`
	Image       *string `json:"image"`
	File        *string `json:"file"`
	Placeholder *string `json:"placeholder"`
}

type UpdateArtikelPayload struct {
	Title      string                  `json:"title" validate:"required,min=3,max=191"`
	CategoryID int64                   `json:"category_id" validate:"required"`
	Thumbnail  string                  `json:"thumbnail" validate:"required"`
	Contents   []ArtikelContentPayload `json:"contents" validate:"required,min=1,dive"`
}

type ArtikelOptionsPayload struct {
	Q     string  `form:"q" query:"q"`
	Page  int     `form:"page" query:"page"`
	Limit int     `form:"limit" query:"limit"`
	IDs   []int64 `form:"id[]" query:"id[]"`
}
