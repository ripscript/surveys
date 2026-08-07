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
	ChangeNameReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ListSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetCalculationTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SectionReportDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (handler *laporanHandler) ChangeNameReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.ChangeNameReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) CreateReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.CreateReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) UpdateCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.UpdateCoverReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) GetCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.GetCoverReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) ListSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.ListSectionReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) CreateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.CreateSectionReport(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) GetCalculationTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.GetCalculationTypeOptions(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) SectionReportDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.GetSectionDetail(ctx, req, usr, param, slug)
}

func (handler *laporanHandler) UpdateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.laporanService.UpdateSectionReport(ctx, req, usr, param, slug)
}
