package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ManajemenArtikelHandler interface {
	CreateKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenArtikelHandler struct {
	manajemenArtikelService service.ManajemenArtikelService
}

func NewManajemenArtikelHandler(
	manajemenArtikelService service.ManajemenArtikelService,
) ManajemenArtikelHandler {
	return &manajemenArtikelHandler{
		manajemenArtikelService,
	}
}

func (handler *manajemenArtikelHandler) CreateKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenArtikelService.CreateKategoriArtikel(usr, req)
}

func (handler *manajemenArtikelHandler) UpdateKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenArtikelService.UpdateKategoriArtikel(usr, req, slug)
}

func (handler *manajemenArtikelHandler) DeleteKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenArtikelService.DeleteKategoriArtikel(usr, slug)
}

func (handler *manajemenArtikelHandler) GetKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenArtikelService.GetKategoriArtikel(usr, slug)
}

func (handler *manajemenArtikelHandler) GetListKategoriArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenArtikelService.GetListKategoriArtikel(req, slug)
}
