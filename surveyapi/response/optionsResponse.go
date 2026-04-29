package response

type OptionItem struct {
	ID    *int64  `form:"id" json:"value"`
	Label *string `form:"label" json:"label"`
}

type StringOptionItem struct {
	ID    *string `form:"id" json:"value"`
	Label *string `form:"label" json:"label"`
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

type StringOptionsResponse struct {
	Options []StringOptionItem `json:"options"`
	Meta    PaginationMeta     `json:"meta"`
}

type OptionsStringIdResponse struct {
	Options []EnumOption   `json:"options"`
	Meta    PaginationMeta `json:"meta"`
}

type EnumOption struct {
	ID   string `json:"value"`
	Name string `json:"label"`
}

type FormFieldOptionsResponse struct {
	Options []FormFieldOptionItem `json:"options"`
	Meta    PaginationMeta        `json:"meta"`
}

type FormFieldOptionItem struct {
	ID    *int64  `form:"id" json:"value"`
	Label *string `form:"label" json:"label"`
	Type  *string `form:"type" json:"type"`
}
