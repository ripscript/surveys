package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ManajemenCMSHandler interface {
	ListSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateStatusSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateNameSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateOrderSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetLandingPage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GeoJsonKotaBandungLevelKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenCMSHandler struct {
	manajemenCMSService service.ManajemenCMSService
}

func NewManajemenCMSHandler(
	manajemenCMSService service.ManajemenCMSService,
) ManajemenCMSHandler {
	return &manajemenCMSHandler{
		manajemenCMSService,
	}
}

func (handler *manajemenCMSHandler) ListSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.ListSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) UpdateStatusSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.UpdateStatusSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) UpdateNameSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.UpdateNameSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) DeleteSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.DeleteSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) CreateSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.CreateSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) GetSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.GetSectionBySlug(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) UpdateSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.UpdateSectionContent(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) UpdateOrderSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.UpdateOrderSection(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) GetLandingPage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.GetLandingPage(ctx, req, usr, param, slug)
}

func (handler *manajemenCMSHandler) GeoJsonKotaBandungLevelKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenCMSService.GeoJsonKotaBandungLevelKecamatan(ctx, req, usr, param, slug)
}
