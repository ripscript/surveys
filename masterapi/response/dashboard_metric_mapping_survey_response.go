package response

type MappingSurveyOptionItem struct {
	AnswerOptionID          int64  `json:"answer_option_id"`
	AnswerLabel             string `json:"answer_label"`
	DashboardMetricStatusID *int64 `json:"dashboard_metric_status_id"`
}

type MappingSurveyQuestionItem struct {
	FormFieldID       int64                     `json:"form_field_id"`
	Question          string                    `json:"question"`
	Description       string                    `json:"description"`
	Template          string                    `json:"template"`
	Sequence          int                       `json:"sequence"`
	DashboardMetricID *int64                    `json:"dashboard_metric_id"`
	MetricKey         *string                   `json:"metric_key"`
	Options           []MappingSurveyOptionItem `json:"options"`
}

type MappingSurveySectionGroup struct {
	SectionID   int                         `json:"section_id"`
	SectionName *string                     `json:"section_name"`
	Questions   []MappingSurveyQuestionItem `json:"questions"`
}

type MappingSurveyResponse struct {
	SurveyID   int64                       `json:"survey_id"`
	SurveyCode string                      `json:"survey_code"`
	SurveyName string                      `json:"survey_name"`
	FlowCode   string                      `json:"flow_code"`
	FlowName   string                      `json:"flow_name"`
	FormID     int64                       `json:"form_id"`
	FormCode   string                      `json:"form_code"`
	Sections   []MappingSurveySectionGroup `json:"sections"`
}
