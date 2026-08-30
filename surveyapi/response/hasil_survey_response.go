package response

type SurveyResultSummaryResponse struct {
	SurveyCode     string                        `json:"survey_code"`
	SurveyName     string                        `json:"survey_name"`
	TotalResponden int64                         `json:"total_responden"`
	Questions      []SurveyResultQuestionSummary `json:"questions"`
}

type SurveyResultQuestionSummary struct {
	QuestionID    int                     `json:"question_id"`
	Question      string                  `json:"question"`
	Sequence      int                     `json:"sequence"`
	Template      string                  `json:"template"`
	TemplateLabel string                  `json:"template_label"`
	TotalAnswered int64                   `json:"total_answered"`
	HasChart      bool                    `json:"has_chart"`
	ChartData     []SurveyResultChartItem `json:"chart_data,omitempty"`
}

type SurveyResultChartItem struct {
	OptionID   *int    `json:"option_id,omitempty"`
	Label      string  `json:"label"`
	Value      int64   `json:"value"`
	Percentage float64 `json:"percentage"`
}
