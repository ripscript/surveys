package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type HasilSurveyHandler interface {
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultQuestionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type hasilSurveyHandler struct {
	manajemenAlurService service.ManajemenAlurService
	hasilSurveyService   service.HasilSurveyService
	surveyService        service.SurveyService
}

func NewHasilSurveyHandler(
	manajemenAlurService service.ManajemenAlurService,
	hasilSurveyService service.HasilSurveyService,
	surveyService service.SurveyService,
) HasilSurveyHandler {
	return &hasilSurveyHandler{
		manajemenAlurService: manajemenAlurService,
		hasilSurveyService:   hasilSurveyService,
		surveyService:        surveyService,
	}
}

func (handler *hasilSurveyHandler) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	go func() {
		bgCtx := context.Background()
		handler.surveyService.SyncExpiredSurveysStatus(bgCtx)
	}()

	return handler.hasilSurveyService.GetListSurvey(ctx, req, usr, param, slug)
}

func (handler *hasilSurveyHandler) SurveyResultSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.hasilSurveyService.SurveyResultSummary(ctx, req, usr, param, slug)
}

func (handler *hasilSurveyHandler) SurveyResultQuestionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.hasilSurveyService.SurveyResultQuestionDetail(ctx, req, usr, param, slug)
}
