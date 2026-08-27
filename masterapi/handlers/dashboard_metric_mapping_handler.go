package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type DashboardMetricMappingHandler interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricMappingHandler struct {
	dashboardMetricMappingService service.DashboardMetricMappingService
}

func NewDashboardMetricMappingHandler(
	dashboardMetricMappingService service.DashboardMetricMappingService,
) DashboardMetricMappingHandler {
	return &dashboardMetricMappingHandler{
		dashboardMetricMappingService,
	}
}

func (handler *dashboardMetricMappingHandler) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricMappingService.Create(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricMappingHandler) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricMappingService.Delete(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricMappingHandler) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricMappingService.Update(ctx, req, usr, param, slug)
}
