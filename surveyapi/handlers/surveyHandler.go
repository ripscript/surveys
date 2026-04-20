package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type SurveyHandler interface {
	CreateSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type surveyHandler struct {
	manajemenAlurService service.ManajemenAlurService
	surveyService        service.SurveyService
}

func NewSurveyHandler(
	manajemenAlurService service.ManajemenAlurService,
	surveyService service.SurveyService,
) SurveyHandler {
	return &surveyHandler{
		manajemenAlurService: manajemenAlurService,
		surveyService:        surveyService,
	}
}

func (handler *surveyHandler) CreateSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.CreateSurvey(ctx, usr, req)
}

func (handler *surveyHandler) OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.OptionsPeriodeSurvey(usr, param)
}

func (handler *surveyHandler) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetListSurvey(ctx, usr, req, slug)
}

func (handler *surveyHandler) ApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ApprovalSurvey(ctx, usr, req, slug)
}

func (handler *surveyHandler) AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.AvailableSurveyWilayah(ctx, req, usr, param, slug)
}
