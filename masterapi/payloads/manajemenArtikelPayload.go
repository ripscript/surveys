package payloads

type KategoriArtikelPayload struct {
	Nama string `json:"nama" validate:"required,max=191"`
}
