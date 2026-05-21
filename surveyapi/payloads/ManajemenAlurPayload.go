package payloads

type ManajemenAlurPayload struct {
	ID                         int        `json:"id,omitempty"`
	FlowCode                   *string    `json:"flow_code,omitempty"`
	TemplateFormulirPertanyaan *string    `json:"form_code" validate:"omitempty,gt=0"`
	NamaAlur                   string     `json:"name" validate:"required"`
	Pembuka                    int        `json:"opening_id" validate:"required,gt=0"`
	Penutup                    int        `json:"closing_id" validate:"required,gt=0"`
	HasSection                 bool       `json:"has_section"`
	Flows                      []FlowItem `json:"flows" validate:"required,min=1,dive"`
}

type FlowItem struct {
	FieldIDs     []int       `json:"field_ids"` // ID database flow_fields (untuk response)
	Sequence     int         `json:"sequence" validate:"required,gt=0"`
	SectionID    *int        `json:"section_id"` // <-- INI YANG TERLEWAT, ditambahkan untuk response detail
	SectionIndex *int        `json:"section_index"`
	SectionName  *string     `json:"section_name"`
	IsGroup      bool        `json:"is_group"`
	GroupID      *int        `json:"group_id"` // ID database group (untuk response)
	GroupName    *string     `json:"group_name"`
	QuestionIDs  []int       `json:"question_ids" validate:"required,min=1"`
	Routing      RoutingRule `json:"routing" validate:"required"`
}

type RoutingRule struct {
	Rule             *string       `json:"rule" validate:"omitempty,oneof=jump-to logic"`
	IsBreakdown      bool          `json:"is_breakdown"`
	TargetQuestionID *int          `json:"target_question_id"`
	IsEnd            bool          `json:"is_end"`
	Options          []OptionRoute `json:"options" validate:"dive"`
	Logics           []LogicRoute  `json:"logics" validate:"dive"`
}

type OptionRoute struct {
	FieldID          int          `json:"field_id,omitempty"` // ID database flow_fields (untuk response)
	OptionID         int          `json:"option_id" validate:"required,gt=0"`
	Rule             *string      `json:"rule" validate:"omitempty,oneof=jump-to logic"`
	TargetQuestionID *int         `json:"target_question_id"`
	IsEnd            bool         `json:"is_end"`
	Logics           []LogicRoute `json:"logics" validate:"dive"`
}

type LogicRoute struct {
	LogicID          int  `json:"logic_id,omitempty"` // ID database advanced_option_flows (untuk response)
	IfQuestionID     int  `json:"if_question_id" validate:"required,gt=0"`
	IfOptionID       int  `json:"if_option_id" validate:"required,gt=0"`
	TargetQuestionID *int `json:"target_question_id"`
	IsEnd            bool `json:"is_end"`
}

// UNTUK KEPERLUAN DETAIL

type ManajemenAlurDetailResponse struct {
	ID         int              `json:"id"`
	FormCode   string           `json:"form_code"`
	Name       string           `json:"name"`
	OpeningID  int              `json:"opening_id"`
	ClosingID  int              `json:"closing_id"`
	HasSection bool             `json:"has_section"`
	Flows      []FlowDetailItem `json:"flows"`
}

type FlowDetailItem struct {
	Sequence     int               `json:"sequence"`
	SectionID    *int              `json:"section_id,omitempty"`
	SectionIndex *int              `json:"section_index,omitempty"`
	SectionName  *string           `json:"section_name,omitempty"`
	IsGroup      bool              `json:"is_group"`
	GroupID      *int              `json:"group_id,omitempty"`
	GroupName    *string           `json:"group_name,omitempty"`
	QuestionIDs  []int             `json:"question_ids"`
	FieldIDs     []int             `json:"field_ids"`
	Routing      RoutingDetailRule `json:"routing"`
}

type RoutingDetailRule struct {
	Rule             *string               `json:"rule"` // Pointer agar bisa null
	IsBreakdown      bool                  `json:"is_breakdown"`
	TargetQuestionID *int                  `json:"target_question_id"`
	IsEnd            bool                  `json:"is_end"`
	Options          []RoutingOptionDetail `json:"options"`
	Logics           []RoutingLogicDetail  `json:"logics"`
}

type RoutingOptionDetail struct {
	FieldID          int                  `json:"field_id"`
	OptionID         int                  `json:"option_id"`
	Rule             *string              `json:"rule"` // PERBAIKAN: Harus ada untuk logic di dalam opsi
	TargetQuestionID *int                 `json:"target_question_id"`
	IsEnd            bool                 `json:"is_end"`
	Logics           []RoutingLogicDetail `json:"logics"` // PERBAIKAN: Harus ada untuk menampung logic
}

type RoutingLogicDetail struct {
	LogicID          int  `json:"logic_id"`
	IfQuestionID     int  `json:"if_question_id"`
	IfOptionID       int  `json:"if_option_id"`
	TargetQuestionID *int `json:"target_question_id"`
	IsEnd            bool `json:"is_end"`
}

// KEPERLUAN UPDATE

type ManajemenAlurUpdatePayload struct {
	NamaAlur   string           `json:"name" validate:"required"`
	Pembuka    int              `json:"opening_id" validate:"required,gt=0"`
	Penutup    int              `json:"closing_id" validate:"required,gt=0"`
	HasSection bool             `json:"has_section"`
	Flows      []FlowUpdateItem `json:"flows" validate:"required,min=1,dive"`
}

type FlowUpdateItem struct {
	Sequence     int               `json:"sequence" validate:"required,gt=0"`
	SectionID    *int              `json:"section_id"`
	SectionIndex *int              `json:"section_index"`
	SectionName  *string           `json:"section_name,omitempty"`
	IsGroup      bool              `json:"is_group"`
	GroupID      *int              `json:"group_id"`
	GroupName    *string           `json:"group_name,omitempty"`
	QuestionIDs  []int             `json:"question_ids" validate:"required,min=1"`
	FieldIDs     []int             `json:"field_ids"`
	Routing      RoutingUpdateRule `json:"routing" validate:"required"`
}

type RoutingUpdateRule struct {
	Rule             string                `json:"rule" validate:"required,oneof=jump-to logic"`
	IsBreakdown      bool                  `json:"is_breakdown"`
	TargetQuestionID *int                  `json:"target_question_id"`
	IsEnd            bool                  `json:"is_end"`
	Options          []RoutingOptionUpdate `json:"options"`
	Logics           []RoutingLogicUpdate  `json:"logics"`
}

type RoutingOptionUpdate struct {
	FieldID          *int `json:"field_id"`
	OptionID         int  `json:"option_id"`
	TargetQuestionID *int `json:"target_question_id"`
	IsEnd            bool `json:"is_end"`
}

type RoutingLogicUpdate struct {
	LogicID          *int `json:"logic_id"`
	IfQuestionID     int  `json:"if_question_id"`
	IfOptionID       int  `json:"if_option_id"`
	TargetQuestionID *int `json:"target_question_id"`
	IsEnd            bool `json:"is_end"`
}

type ManajemenAlurOptionsPayload struct {
	Q          string   `form:"q" query:"q"`         // Kata kunci pencarian
	Page       int      `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit      int      `form:"limit" query:"limit"` // Jumlah data per halaman
	IDs        []string `form:"id[]" query:"id[]"`   // Bypass untuk mengambil ID spesifik (misal saat edit data)
	ExcludeIDs []string `json:"exclude_ids"`
}
