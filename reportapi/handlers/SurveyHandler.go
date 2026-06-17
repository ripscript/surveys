package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	"backend/reportapi/utils"
	pb "backend/siccore/pb"
	"context"
	"net/http"
	"net/url"
)

type SurveyHandler interface {
	LogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	InsertLogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyActvitiesExport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type surveyHandler struct {
	surveyService service.SurveyService
}

func NewSurveyHandler(
	surveyService service.SurveyService,
) SurveyHandler {
	return &surveyHandler{
		surveyService,
	}
}

func (handler *surveyHandler) LogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.LogSurveys(param)
}

func (handler *surveyHandler) InsertLogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.InsertLogSurveys(req)
}

func (handler *surveyHandler) SurveyActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	var status *string
	if s := param.Get("status"); s != "" {
		status = &s
	}

	return handler.surveyService.GetDataSurvey(usr, status, param)
}

func (handler *surveyHandler) SurveyActvitiesExport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	surveyIDStr := param.Get("survey_id")
	wilayah := param.Get("wilayah")
	wilayahIDStr := param.Get("wilayah_id")

	exportReq := service.ExportRequest{
		SurveyID:  surveyIDStr,
		Wilayah:   wilayah,
		WilayahID: wilayahIDStr,
	}

	result, err := handler.surveyService.SurveyActivitiesExport(exportReq)
	if err != nil {
		return utils.SetResponseData(
			nil,
			false,
			"Gagal membuat export: "+err.Error(),
			http.StatusInternalServerError,
			err,
			"",
		), nil
	}

	return utils.SetResponseData(
		result.Data,
		true,
		"Data File,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		http.StatusOK,
		nil,
		"",
	), nil
}
func decodeID(encoded string) (int, error) {
	id, err := utils.DecryptInt(encoded)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
