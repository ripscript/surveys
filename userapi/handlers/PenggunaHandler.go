package handlers

import (
	pb "backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/service"
	"context"
	"net/url"
)

type PenggunaHandler interface {
	Login(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Logout(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Menus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type penggunaHandler struct {
	penggunaService service.PenggunaService
}

func NewPenggunaHandler(
	penggunaService service.PenggunaService,
) PenggunaHandler {
	return &penggunaHandler{
		penggunaService,
	}
}
func (handler *penggunaHandler) Login(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	// return handler.penggunaService.Login(usr, req)
	return handler.penggunaService.LoginV2(usr, req)
}

func (handler *penggunaHandler) Logout(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.penggunaService.Logout(usr)
}

func (handler *penggunaHandler) Menus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.penggunaService.Menus(usr)
}
