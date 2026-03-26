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
