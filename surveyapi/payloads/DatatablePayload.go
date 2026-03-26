package payloads

type DatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}
