package payloads

type UpdateProfileUmum struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	NoTlp string `json:"phone_number"`
}

type UpdateProfileJastiper struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	NoTlp         string `json:"phone_number"`
	NoRek         string `json:"account_number"`
	Avatar        string `json:"avatar"`
	IdentityPhoto string `json:"identity_photo"`
}

type UpdateProfileBundlePayload struct {
	NIK          *string `json:"nik"`
	Name         string  `json:"nama_responden" validate:"required"`
	Alamat       *string `json:"alamat"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"`
	PhoneNumber  *string `json:"nomor_telepon"`
	Email        string  `json:"email" validate:"required,email"`

	NoSK string `json:"no_sk" validate:"required"`

	CurrentPassword string `json:"password_saat_ini,omitempty" validate:"omitempty,required_with=NewPassword"`

	NewPassword string `json:"password_baru,omitempty" validate:"omitempty,password_rule,required_with=CurrentPassword"`

	ConfirmNewPassword string `json:"konfirmasi_password_baru,omitempty" validate:"omitempty,eqfield=NewPassword,required_with=NewPassword"`
}
