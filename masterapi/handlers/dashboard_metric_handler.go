package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type DashboardMetricHandler interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetStatuses(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetMetricOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetPemetaanMetrikSurveyList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricHandler struct {
	dashboardMetricService service.DashboardMetricService
}

func NewDashboardMetricHandler(dashboardMetricService service.DashboardMetricService) DashboardMetricHandler {
	return &dashboardMetricHandler{
		dashboardMetricService: dashboardMetricService,
	}
}

func (handler *dashboardMetricHandler) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Create(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Update(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.GetList(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Detail(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Delete(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.GetCategoryOptions(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) GetStatuses(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.GetStatuses(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) GetMetricOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.GetMetricOptions(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) GetPemetaanMetrikSurveyList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.GetPemetaanMetrikSurveyList(ctx, req, usr, param, slug)
}
