package payloads

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterPayload struct {
	Email              string `json:"email"`
	Name               string `json:"name"`
	Password           string `json:"password"`
	KonfirmasiPassword string `json:"konfirmasi_password"`
}

type CreatePengguna struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Role            string `json:"role"`
}

type UpdatePengguna struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type CompleteProfile struct {
	ID            int64  `json:"id"`
	Jastiper      bool   `json:"jastiper"`
	Name          string `json:"name"`
	PhoneNumber   string `json:"phoneNumber"`
	IdentityPhoto string `json:"identityPhoto"`
	Avatar        string `json:"avatar"`
	PostalCode    string `json:"postalCode"`
}

type UserBank struct {
	UserId        int64  `json:"user_id"`
	AccountNumber string `json:"accountNumber"`
	BankId        int64  `json:"bank_id"`
}
