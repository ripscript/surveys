package payloads

type LoginPayload struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	Role              int    `json:"role"`
	SelectedKecamatan *int   `json:"selected_kecamatan,omitempty"`
	SelectedKelurahan *int   `json:"selected_kelurahan,omitempty"`
	SelectedRW        *int   `json:"selected_rw,omitempty"`
	SelectedRT        *int   `json:"selected_rt,omitempty"`
}
