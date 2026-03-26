package payloads

type CreateRespondenPayload struct {
	NIK           *string `json:"nik" validate:"omitempty,len=16,numeric"`
	NamaResponden string  `json:"nama_responden" validate:"required,max=70"`
	TempatLahir   *string `json:"tempat_lahir" validate:"omitempty,max=100"`
	TanggalLahir  *string `json:"tanggal_lahir" validate:"omitempty,datetime=2006-01-02"`
	Alamat        *string `json:"alamat" validate:"omitempty,max=500"`
	NoTelepon     *string `json:"no_telepon" validate:"omitempty,max=191"`
	Email         *string `json:"email" validate:"omitempty,email,max=191"`
	RoleId        int64   `json:"role_id" validate:"required"` // Pasti angka karena tipe int64
	KecamatanId   *int64  `json:"kecamatan_id"`
	KelurahanId   *int64  `json:"kelurahan_id"`
	RwId          *int64  `json:"rw_id"`
	RtId          *int64  `json:"rt_id"`
}
