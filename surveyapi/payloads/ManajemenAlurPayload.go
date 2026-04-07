package payloads

type ManajemenAlurPayload struct {
	KodeFormulirPertanyaan string     `json:"form_code" validate:"required"`
	NamaAlur               string     `json:"name" validate:"required,max=191"`
	Pembuka                int        `json:"opening_id" validate:"required"`
	Penutup                int        `json:"closing_id" validate:"required"`
	Section                *bool      `json:"has_section" validate:"required"`
	Pertanyaan             []FlowItem `json:"flows" validate:"required,dive"`
}

type FlowItem struct {
	SectionID   *int            `json:"section_id"`
	Type        string          `json:"type"` // question atau group
	ReferenceID int             `json:"reference_id"`
	Questions   []GroupQuestion `json:"questions,omitempty"`
	Routing     RoutingRule     `json:"routing"`
}

type GroupQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Template string `json:"template"`
}

type RoutingRule struct {
	Rule        string        `json:"rule"`
	TargetType  string        `json:"target_type"` // "question", "group", "end", "none" (bila breakdown)
	TargetID    *int          `json:"target_id"`   // Pointer agar bisa null
	IsBreakdown bool          `json:"is_breakdown"`
	Options     []OptionRoute `json:"options,omitempty"`
}

type OptionRoute struct {
	OptionID   int    `json:"option_id"`
	Rule       string `json:"rule"`
	TargetType string `json:"target_type"`
	TargetID   int    `json:"target_id"`
}
