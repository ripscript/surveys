package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type TemplateFormulirPertanyaanHandler interface {
	CreateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DuplicateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetQuestionTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetFormulirPertanyaanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetPertanyaanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetMultipleChoiceOptionByFormFieldId(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type templateFormulirPertanyaanHandler struct {
	templateFormulirPertanyaanService service.TemplateFormulirPertanyaanService
}

func NewTemplateFormulirPertanyaanHandler(
	templateFormulirPertanyaanService service.TemplateFormulirPertanyaanService,
) TemplateFormulirPertanyaanHandler {
	return &templateFormulirPertanyaanHandler{
		templateFormulirPertanyaanService: templateFormulirPertanyaanService,
	}
}

func (handler *templateFormulirPertanyaanHandler) CreateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.CreateTemplateFormulirPertanyaan(usr, req)
}

func (handler *templateFormulirPertanyaanHandler) GetDetailTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetDetailTemplateFormulirPertanyaan(usr, slug)
}

func (handler *templateFormulirPertanyaanHandler) UpdateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.UpdateTemplateFormulirPertanyaan(usr, req, slug)
}

func (handler *templateFormulirPertanyaanHandler) DuplicateTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.DuplicateTemplateFormulirPertanyaan(usr, req, slug)
}

func (handler *templateFormulirPertanyaanHandler) DeleteTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.DeleteTemplateFormulirPertanyaan(usr, req, slug)
}

func (handler *templateFormulirPertanyaanHandler) GetListTemplateFormulirPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetListTemplateFormulirPertanyaan(usr, req, slug)
}

func (handler *templateFormulirPertanyaanHandler) GetQuestionTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetQuestionTypeOptions(usr, param)
}

func (handler *templateFormulirPertanyaanHandler) GetFormulirPertanyaanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetFormulirPertanyaanOptions(param)
}

func (handler *templateFormulirPertanyaanHandler) GetPertanyaanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetPertanyaanOptions(param, slug)
}

func (handler *templateFormulirPertanyaanHandler) DetailPertanyaan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.DetailPertanyaan(usr, slug)
}

func (handler *templateFormulirPertanyaanHandler) GetMultipleChoiceOptionByFormFieldId(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.templateFormulirPertanyaanService.GetMultipleChoiceOptionByFormFieldId(param, slug)
}
