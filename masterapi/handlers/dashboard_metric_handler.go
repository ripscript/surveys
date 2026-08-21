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
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricHandler struct {
	dashboardMetricService service.DashboardMetricService
}

func NewDashboardMetricHandler(
	dashboardMetricService service.DashboardMetricService,
) DashboardMetricHandler {
	return &dashboardMetricHandler{
		dashboardMetricService,
	}
}

func (handler *dashboardMetricHandler) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Create(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Detail(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Update(ctx, req, usr, param, slug)
}

func (handler *dashboardMetricHandler) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardMetricService.Delete(ctx, req, usr, param, slug)
}
