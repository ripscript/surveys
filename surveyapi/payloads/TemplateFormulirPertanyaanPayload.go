package payloads

type TemplateFormPayload struct {
	Title             string `json:"title" validate:"required,max=191"`
	Description       string `json:"description" validate:"required"`
	IsTematikQuestion *bool  `json:"is_tematik_question" validate:"required"`
}

type TemplateFormulirPertanyaanPayload struct {
	Title             string     `json:"title" validate:"required,max=191"`
	Description       string     `json:"description" validate:"required"`
	IsTematikQuestion *bool      `json:"is_tematik_question" validate:"required"`
	Questions         []Question `json:"questions" validate:"dive"`
}

type Question struct {
	InputType     string   `json:"input_type" validate:"required,max=191"`
	Question      string   `json:"question" validate:"required,max=191"`
	Description   *string  `json:"description"`
	Required      bool     `json:"required" validate:"required"`
	Options       []Option `json:"options,omitempty" validate:"omitempty,dive"`
	ImageQuantity *int     `json:"image_quantity,omitempty"`
}

type Option struct {
	ID   *int   `json:"id" validate:"omitempty"`
	Name string `json:"name" validate:"required"`
}

type UpdateTemplateFormulirPertanyaanPayload struct {
	Title              string           `json:"title" validate:"required,max=191"`
	Description        string           `json:"description" validate:"required"`
	IsTematikQuestion  *bool            `json:"is_tematik_question" validate:"required"`
	IsQuestionsUpdated *bool            `json:"is_questions_updated,omitempty"`
	Questions          []UpdateQuestion `json:"questions" validate:"dive"`
}

type UpdateQuestion struct {
	Attribute     string   `json:"attribute" validate:"required"`
	InputType     string   `json:"input_type" validate:"required,max=191"`
	Question      string   `json:"question" validate:"required,max=191"`
	Description   *string  `json:"description"`
	Required      bool     `json:"required" validate:"required"`
	Options       []Option `json:"options,omitempty" validate:"omitempty,dive"`
	ImageQuantity *int     `json:"image_quantity,omitempty"`
}

type TypeQuestionOptionsPayload struct {
	Q     string `form:"q" query:"q"`         // Kata kunci pencarian
	Page  int    `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit int    `form:"limit" query:"limit"` // Jumlah data per halaman
}

type FormulirPertanyaanOptionsPayload struct {
	Q     string   `form:"q" query:"q"`         // Kata kunci pencarian
	Page  int      `form:"page" query:"page"`   // Halaman saat ini (untuk lazy load)
	Limit int      `form:"limit" query:"limit"` // Jumlah data per halaman
	IDs   []string `form:"id[]" query:"id[]"`   // Bypass untuk mengambil ID spesifik (misal saat edit data)
}
