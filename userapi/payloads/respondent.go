package payloads

type UpdateRespondent struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	RoleID      int    `json:"role"`
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
