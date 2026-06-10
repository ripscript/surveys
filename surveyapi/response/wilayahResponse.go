package response

import "backend/surveyapi/models"

type Meta struct {
	Limit      int `json:"limit"`
	Page       int `json:"page"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type RTDatatableResponse struct {
	Data []models.RtDatatableResponse `json:"data"`
	Meta Meta                         `json:"meta"`
}

type DetailSurveyKewilayahanDatatableResponse struct {
	Data []DetailSurveyKewilayahanResponse `json:"data"`
	Meta Meta                              `json:"meta"`
}

type DetailSurveyKewilayahanResponse struct {
	No                     int64  `json:"no"`
	ID                     int64  `json:"-"`
	Code                   string `json:"code"`
	NamaWilayah            string `json:"nama_wilayah"`
	Status                 string `json:"status"`
	IsPosibleDetail        bool   `json:"posible_detail"`
	IsPosiblePreviewSurvey bool   `json:"posible_preview_survey"`
	IsPosibleBackAccess    bool   `json:"posible_back_access"`
}

type RWDatatableResponse struct {
	Data []models.RwDatatableResponse `json:"data"`
	Meta Meta                         `json:"meta"`
}

type KelurahanDatatableResponse struct {
	Data []models.KelurahanDatatableResponse `json:"data"`
	Meta Meta                                `json:"meta"`
}

type KecamatanDatatableResponse struct {
	Data []models.KecamatanDatatableResponse `json:"data"`
	Meta Meta                                `json:"meta"`
}
