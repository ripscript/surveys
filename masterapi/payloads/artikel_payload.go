package payloads

type ArtikelPayload struct {
	Judul    string `json:"title" validate:"required,min=3,max=191"`
	Kategori int64  `json:"category_id" validate:"required"`
}

type ArtikelUpdatePayload struct {
	Judul     string `json:"title" validate:"required,min=3,max=191"`
	Kategori  int64  `json:"category_id" validate:"required"`
	Thumbnail string `json:"thumbnail" validate:"omitempty"`
}
