package payloads

type GetMappingSurveyQuestionsPayload struct {
	FlowCode string `query:"flow_code"`
}

type SaveMappingSurveyOptionPayload struct {
	AnswerOptionID          int64  `json:"answer_option_id" validate:"required"`
	DashboardMetricStatusID *int64 `json:"dashboard_metric_status_id" validate:"required"`
}

type SaveMappingSurveyQuestionPayload struct {
	FormFieldID       int64                            `json:"form_field_id" validate:"required"`
	DashboardMetricID *int64                           `json:"dashboard_metric_id"`
	Options           []SaveMappingSurveyOptionPayload `json:"options" validate:"dive"`
}

type SaveMappingSurveyRequest struct {
	Questions []SaveMappingSurveyQuestionPayload `json:"questions" validate:"required,dive"`
}
