package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ArtikelPromoteHandler interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	List(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type artikelPromoteHandler struct {
	artikelPromoteService service.ArtikelPromoteService
}

func NewArtikelPromoteHandler(
	artikelPromoteService service.ArtikelPromoteService,
) ArtikelPromoteHandler {
	return &artikelPromoteHandler{
		artikelPromoteService,
	}
}

func (handler *artikelPromoteHandler) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelPromoteService.Create(ctx, req, usr, param, slug)
}

func (handler *artikelPromoteHandler) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelPromoteService.Delete(ctx, req, usr, param, slug)
}

func (handler *artikelPromoteHandler) UpdateStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelPromoteService.UpdateStatus(ctx, req, usr, param, slug)
}

func (handler *artikelPromoteHandler) List(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelPromoteService.GetList(ctx, req, usr, param, slug)
}
