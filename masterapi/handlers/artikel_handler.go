package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ArtikelHandler interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type artikelHandler struct {
	artikelService service.ArtikelService
}

func NewArtikelHandler(
	artikelService service.ArtikelService,
) ArtikelHandler {
	return &artikelHandler{
		artikelService,
	}
}

func (handler *artikelHandler) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelService.Create(ctx, req, usr, param, slug)
}

func (handler *artikelHandler) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.artikelService.Update(ctx, req, usr, param, slug)
}
