package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type ManajemenAlurHandler interface {
	CreateManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenAlurHandler struct {
	manajemenAlurService service.ManajemenAlurService
}

func NewManajemenAlurHandler(
	manajemenAlurService service.ManajemenAlurService,
) ManajemenAlurHandler {
	return &manajemenAlurHandler{
		manajemenAlurService: manajemenAlurService,
	}
}

func (handler *manajemenAlurHandler) CreateManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.CreateManajemenAlur(usr, req)
}
