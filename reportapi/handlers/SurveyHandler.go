package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type SurveyHandler interface {
	SurveyActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	LogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (handler *surveyHandler) SurveyActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.SurveyActivities(slug)
}

func (handler *surveyHandler) LogSurveys(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.LogSurveys(param)
}
