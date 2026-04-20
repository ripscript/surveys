package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type ManajemenPenggunaHandler interface {
	CreateResponden(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenPenggunaHandler struct {
	manajemenPenggunaService service.ManajemenPenggunaService
}

func NewManajemenPenggunaHandler(
	manajemenPenggunaService service.ManajemenPenggunaService,
) ManajemenPenggunaHandler {
	return &manajemenPenggunaHandler{
		manajemenPenggunaService,
	}
}

func (handler *manajemenPenggunaHandler) CreateResponden(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenPenggunaService.CreateResponden(usr, req)
}
