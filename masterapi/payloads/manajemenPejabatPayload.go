package payloads

type PejabatPayload struct {
	NamaPejabat  int64   `json:"responden_id" validate:"required"`
	PeriodeAwal  *string `json:"periode_awal" validate:"required,datetime=2006-01-02"`
	PeriodeAkhir *string `json:"periode_akhir" validate:"required,datetime=2006-01-02"`
	NoSK         *string `json:"no_sk" validate:"omitempty"`
	StatusJabat  *bool   `json:"status_jabat" validate:"required"`
}
