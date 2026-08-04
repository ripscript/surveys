package response

type PreviewAlurSurveyBlueprintResponse struct {
	SurveyInfo PreviewAlurSurveyInfo         `json:"survey_info"`
	Opening    *string                       `json:"opening"`
	Closing    *string                       `json:"closing"`
	Nodes      map[int]PreviewAlurSurveyStep `json:"nodes"`
}

type PreviewAlurSurveyInfo struct {
	Name              string  `json:"name"`
	HasSection        bool    `json:"has_section"`
	ActiveSectionCode *string `json:"active_section_code"`
	ActiveSectionName *string `json:"active_section_name"`
	TotalRequired     int     `json:"total_required"`
	TotalOptional     int     `json:"total_optional"`
	EntryNodeId       int     `json:"entry_node_id"`
}

type PreviewAlurSurveyStep struct {
	StepType  string                            `json:"step_type"` // single atau group
	GroupId   *int                              `json:"group_id"`
	GroupName *string                           `json:"group_name"`
	Questions []PreviewAlurSurveyQuestionDetail `json:"questions"`
	Routing   PreviewAlurSurveyRoutingDetail    `json:"routing"`
}

type PreviewAlurSurveyQuestionDetail struct {
	QuestionId         int                           `json:"question_id"`
	Type               string                        `json:"type"` // multiple-choices, long-answer, number, image-template, maps
	Label              string                        `json:"label"`
	Description        *string                       `json:"description"`
	IsRequired         bool                          `json:"is_required"`
	ExpectedImageCount *int                          `json:"expected_image_count"`
	Options            []PreviewAlurSurveyOptionItem `json:"options"`
	Answer             interface{}                   `json:"answer"`
}

type PreviewAlurSurveyOptionItem struct {
	ID               int                                  `json:"id"`
	Label            string                               `json:"label"`
	Rule             *string                              `json:"rule"`
	TargetQuestionId *int                                 `json:"target_question_id"`
	IsEnd            bool                                 `json:"is_end"`
	Logics           []PreviewAlurSurveyAdvancedLogicItem `json:"logics"` // Array bersarang untuk logic per-opsi
}

type PreviewAlurSurveyRoutingDetail struct {
	Rule             *string                              `json:"rule"`
	IsBreakdown      bool                                 `json:"is_breakdown"`
	IsEnd            bool                                 `json:"is_end"`
	TargetQuestionId *int                                 `json:"target_question_id"`
	Logics           []PreviewAlurSurveyAdvancedLogicItem `json:"logics"` // Logic utama (non-breakdown)
}

type PreviewAlurSurveyAdvancedLogicItem struct {
	IfQuestionId     int  `json:"if_question_id"`
	IfOptionId       int  `json:"if_option_id"`
	TargetQuestionId *int `json:"target_question_id"`
	IsEnd            bool `json:"is_end"`
}
