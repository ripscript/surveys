package handlers

import (
	pb "backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/service"
	"context"
	"net/url"
)

type UsersBlokirHandler interface {
	GetListdata(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	OpenBlokir(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type usersBlokirHandler struct {
	usersBlokirService service.UsersBlokirService
}

func NewUsersBlokirHandler(
	usersBlokirService service.UsersBlokirService,
) UsersBlokirHandler {
	return &usersBlokirHandler{
		usersBlokirService,
	}
}

func (handler *usersBlokirHandler) GetListdata(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.usersBlokirService.GetListdata(param)
}

func (handler *usersBlokirHandler) OpenBlokir(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.usersBlokirService.OpenBlokir(usr, slug)
}
