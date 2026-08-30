package response

type SurveyResultDetailResponse struct {
	QuestionCode string                   `json:"question_code"`
	Question     string                   `json:"question"`
	Template     string                   `json:"template"`
	Data         []SurveyResultDetailItem `json:"data"`
}

type SurveyResultDetailItem struct {
	No     int         `json:"no"`
	Name   string      `json:"name"`
	Answer interface{} `json:"answer"`
}

type SurveyResultMapCoordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
