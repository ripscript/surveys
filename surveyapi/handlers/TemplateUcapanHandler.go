package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type TemplateUcapanHandler interface {
	GetTemplateUcapanDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	OptionsVariableTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type templateUcapanHandler struct {
	templateUcapanService service.TemplateUcapanService
}

func NewTemplateUcapanHandler(
	templateUcapanService service.TemplateUcapanService,
) TemplateUcapanHandler {
	return &templateUcapanHandler{
		templateUcapanService: templateUcapanService,
	}
}

func (handler *templateUcapanHandler) GetTemplateUcapanDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.GetGeneralTemplate(slug)
}

func (handler *templateUcapanHandler) OptionsVariableTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.GetTemplateUcapanOptions(param)
}

func (handler *templateUcapanHandler) CreateTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.CreateTemplateUcapan(usr, req)
}

func (handler *templateUcapanHandler) UpdateTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.UpdateTemplateUcapan(usr, req, slug)
}

func (handler *templateUcapanHandler) DeleteTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.DeleteTemplateUcapan(usr, slug)
}

func (handler *templateUcapanHandler) GetListTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateUcapanService.GetListTemplateUcapan(usr, req, slug)
}
