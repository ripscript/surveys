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
	GetDetailManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	FlowPreviewIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewAlurSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	AlurOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (handler *manajemenAlurHandler) GetDetailManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.GetDetailManajemenAlur(usr, slug)
}

func (handler *manajemenAlurHandler) UpdateManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.UpdateManajemenAlur(usr, req, slug)
}

func (handler *manajemenAlurHandler) DeleteManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.DeleteManajemenAlur(usr, slug)
}

func (handler *manajemenAlurHandler) GetListManajemenAlur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.GetListManajemenAlur(ctx, req, usr, param, slug)
}

func (handler *manajemenAlurHandler) FlowPreviewIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.FlowPreviewIndex(usr, slug)
}

func (handler *manajemenAlurHandler) PreviewAlurSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.PreviewAlurSurvey(ctx, usr, slug)
}

func (handler *manajemenAlurHandler) AlurOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.manajemenAlurService.AlurOptions(ctx, req, usr, param, slug)
}
