package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type DashboardMetricMappingSurveyHandler interface {
	GetMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SaveMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricMappingSurveyHandler struct {
	dashboardMetricMappingSurveyService service.DashboardMetricMappingSurveyService
}

func NewDashboardMetricMappingSurveyHandler(dashboardMetricMappingSurveyService service.DashboardMetricMappingSurveyService) DashboardMetricMappingSurveyHandler {
	return &dashboardMetricMappingSurveyHandler{
		dashboardMetricMappingSurveyService: dashboardMetricMappingSurveyService,
	}
}

func (handler *dashboardMetricMappingSurveyHandler) GetMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricMappingSurveyService.GetMappingSurvey(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricMappingSurveyHandler) SaveMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricMappingSurveyService.SaveMappingSurvey(ctx, req, usr, param, slug)
}
