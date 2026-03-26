package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"fmt"
	"net/url"
)

type TemplateUcapanHandler interface {
	GetTemplateUcapanDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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
	fmt.Println("===================")
	return handler.templateUcapanService.GetGeneralTemplate(slug)
}
