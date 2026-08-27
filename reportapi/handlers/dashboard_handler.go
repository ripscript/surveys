package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type DashboardHandler interface {
	GetDashboardSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDashboardSaranaPengelolaanSampah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDashboardCardInfrastruktur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDashboardStuntingVsRTLH(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDashboardTrend(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDashboardComparison(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHeatMap(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardHandler struct {
	dashboardService service.DashboardService
}

func NewDashboardHandler(dashboardService service.DashboardService) DashboardHandler {
	return &dashboardHandler{
		dashboardService: dashboardService,
	}
}

func (handler *dashboardHandler) GetDashboardSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetSummary(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetDashboardSaranaPengelolaanSampah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetSampah(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetDashboardCardInfrastruktur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetInfrastruktur(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetDashboardStuntingVsRTLH(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetStuntingVsRTLH(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetDashboardTrend(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetTrend(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetDashboardComparison(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetComparison(ctx, req, usr, param, slug)
}

func (handler *dashboardHandler) GetHeatMap(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardService.GetHeatmap(ctx, req, usr, param, slug)
}
