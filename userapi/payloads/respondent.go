package payloads

type UpdateRespondent struct {
	Name        string `json:"name" validate:"required,max=70"`
	PhoneNumber string `json:"phone_number" validate:"required,max=15"`
	Email       string `json:"email" validate:"required,email,max=191"`
	RoleID      int    `json:"role" validate:"required"`
}

type UpdateRespondents struct {
	NIK          string `json:"nik"`
	Name         string `json:"name"`
	TempatLahir  string `json:"place_of_birth"`
	TanggalLahir string `json:"date_of_birth"`
	Alamat       string `json:"address"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	RoleID       int    `json:"role"`
	Kecamatan    string `json:"kecamatan"`
	Kelurahan    string `json:"kelurahan"`
	RW           string `json:"rw"`
	RT           string `json:"rt"`
}

type ImportRespondents struct {
	NIK          string `json:"nik"`
	Name         string `json:"name"`
	TempatLahir  string `json:"place_of_birth"`
	TanggalLahir string `json:"date_of_birth"`
	Alamat       string `json:"address"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Role         string `json:"roleString"`
	Kecamatan    string `json:"kecamatan"`
	Kelurahan    string `json:"kelurahan"`
	RW           string `json:"rw"`
	RT           string `json:"rt"`
}

type CreateRespondent struct {
	Respondent []UpdateRespondents `json:"respondent"`
}

type SurveyorOptionsPayload struct {
	Q     string  `form:"q" query:"q"`
	Page  int     `form:"page" query:"page"`
	Limit int     `form:"limit" query:"limit"`
	IDs   []int64 `form:"id[]" query:"id[]"`
}

type UpdatePasswordRespondent struct {
	PasswordBaru           string `json:"password" validate:"required,min=8,max=191,password_rule"`
	KonfirmasiPasswordBaru string `json:"confirm_password" validate:"required,min=8,max=191,eqfield=PasswordBaru"`
}

type RespondentOptionsPayload struct {
	Q           string   `form:"q" query:"q"`
	Page        int      `form:"page" query:"page"`
	Limit       int      `form:"limit" query:"limit"`
	IDs         []string `form:"id[]" query:"id[]"`
	TipeWilayah string   `form:"tipe_wilayah" query:"tipe_wilayah"`
	KecamatanId string   `form:"kecamatan_id" query:"kecamatan_id"`
	KelurahanId string   `form:"kelurahan_id" query:"kelurahan_id"`
	RWId        string   `form:"rw_id" query:"rw_id"`
	RTId        string   `form:"rt_id" query:"rt_id"`
	Status      string   `form:"status" query:"status"`
	HasJabatan  string   `form:"has_jabatan" query:"has_jabatan"`
}
