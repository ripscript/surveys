package response

type SurveyResultIndex struct {
	Respondent RespondentDataIndex `json:"respondent"`
	Sections   []SectionDataIndex  `json:"sections"`
}

type RespondentDataIndex struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Kecamatan string  `json:"kecamatan"`
	Kelurahan string  `json:"kelurahan"`
	RW        string  `json:"rw"`
	RT        string  `json:"rt"`
	Status    *string `json:"status"`
}

type SectionDataIndex struct {
	SectionCode string `json:"section_code"`
	SectionName string `json:"section_name"`
}

type SurveyResultSectionDetail struct {
	SectionName string             `json:"section_name"`
	Nodes       []SurveyResultNode `json:"nodes"`
}

type SurveyResultNode struct {
	GroupID   *int                         `json:"group_id"`
	GroupName *string                      `json:"group_name"`
	StepType  string                       `json:"step_type"` // "single" atau "group"
	Questions []SurveyResultQuestionDetail `json:"questions"`
}

// 🆕 Tambahan struct untuk menampung data Opsi
type SurveyResultOptionItem struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

type SurveyResultQuestionDetail struct {
	QuestionID int                      `json:"question_id"`
	Label      string                   `json:"label"`
	Type       string                   `json:"type"`
	IsRequired bool                     `json:"is_required"`
	Options    []SurveyResultOptionItem `json:"options"` // 🆕 Field Baru
	Answer     interface{}              `json:"answer"`
}
