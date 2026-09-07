package payloads

type SeedSurveyAnswerRequest struct {
	SurveyID     int64 `json:"survey_id" validate:"required"`
	RespondentID int64 `json:"respondent_id" validate:"required"`
}

type SeedSurveyAnswerResponse struct {
	TotalFieldTerisi int `json:"total_field_terisi"`
}
