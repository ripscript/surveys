package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ManajemenPejabatHandler interface {
	CreatePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdatePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, reqSlug map[string]interface{}) (*pb.ProxyResponse, error)
	DeletePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenPejabatHandler struct {
	manajemenPejabatService service.ManajemenPejabatService
}

func NewManajemenPejabatHandler(
	manajemenPejabatService service.ManajemenPejabatService,
) ManajemenPejabatHandler {
	return &manajemenPejabatHandler{
		manajemenPejabatService,
	}
}

func (handler *manajemenPejabatHandler) CreatePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPejabatService.CreatePejabat(ctx, usr, req)
}

func (handler *manajemenPejabatHandler) DetailPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPejabatService.DetailPejabat(ctx, usr, req, slug)
}

func (handler *manajemenPejabatHandler) UpdatePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, reqSlug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPejabatService.UpdatePejabat(ctx, usr, req, reqSlug)
}

func (handler *manajemenPejabatHandler) DeletePejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPejabatService.DeletePejabat(ctx, usr, req, slug)
}

func (handler *manajemenPejabatHandler) GetListPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPejabatService.GetListPejabat(ctx, req, usr, param, slug)
}
