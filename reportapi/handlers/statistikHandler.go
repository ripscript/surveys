package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type StatistikHandler interface {
	GetStatistikIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type statistikHandler struct {
	statistikService service.StatistikService
}

func NewStatistikHandler(statistikService service.StatistikService) StatistikHandler {
	return &statistikHandler{
		statistikService: statistikService,
	}
}

func (handler *statistikHandler) GetStatistikIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.statistikService.GetStatistikIndex(ctx, req, usr, param, slug)
}

func (handler *statistikHandler) GetStatistik(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.statistikService.GetStatistik(ctx, req, usr, param, slug)
}
