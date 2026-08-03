package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type LaporanHandler interface {
	ListLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ChangeNameLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateCoverLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCoverLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CetakLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type laporanHandler struct {
	laporanService service.LaporanService
}

func NewLaporanHandler(laporanService service.LaporanService) LaporanHandler {
	return &laporanHandler{
		laporanService: laporanService,
	}
}

func (handler *laporanHandler) ListLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.ListLaporan(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) ChangeNameLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.ChangeNameLaporan(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) CreateLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.CreateLaporan(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) UpdateCoverLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.UpdateCoverLaporan(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) GetCoverLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.GetCoverLaporan(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) CetakLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.CetakLaporan(ctx, req, usr, param, slug)
}
