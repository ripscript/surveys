package payloads

type ArtikelPromotePayload struct {
	ArtikelID int64  `json:"artikel_id" validate:"required"`
	Status    string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateArtikelPromotePayload struct {
	Status string `json:"status" validate:"required,oneof=active inactive"`
}

type ArtikelPromoteOptionsPayload struct {
	Q      string  `form:"q" query:"q"`
	Page   int     `form:"page" query:"page"`
	Limit  int     `form:"limit" query:"limit"`
	Status string  `form:"status" query:"status"`
	IDs    []int64 `form:"id[]" query:"id[]"`
}
