package response

type OptionItem struct {
	ID    *int64  `form:"id"`
	Label *string `form:"label"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	HasMore     bool  `json:"has_more"`
}

type OptionsResponse struct {
	Options []OptionItem   `json:"options"`
	Meta    PaginationMeta `json:"meta"`
}
