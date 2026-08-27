package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type DashboardOptionStatusMappingHandler interface {
	BulkAssign(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardOptionStatusMappingHandler struct {
	dashboardOptionStatusMappingService service.DashboardOptionStatusMappingService
}

func NewDashboardOptionStatusMappingHandler(
	dashboardOptionStatusMappingService service.DashboardOptionStatusMappingService,
) DashboardOptionStatusMappingHandler {
	return &dashboardOptionStatusMappingHandler{
		dashboardOptionStatusMappingService,
	}
}

func (handler *dashboardOptionStatusMappingHandler) BulkAssign(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.dashboardOptionStatusMappingService.BulkAssign(ctx, req, usr, param, slug)
}
