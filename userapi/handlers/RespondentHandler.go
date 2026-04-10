package handlers

import (
	pb "backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/service"
	"context"
	"net/url"
)

type RespondentHandler interface {
	CreateRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetExampleImport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ImportRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type respondentHandler struct {
	respondentService service.RespondentService
}

func NewRespondentHandler(
	respondentService service.RespondentService,
) RespondentHandler {
	return &respondentHandler{
		respondentService,
	}
}
func (handler *respondentHandler) CreateRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.CreateRespondent(req, usr)
}
func (handler *respondentHandler) GetRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRespondent(usr, param)
}

func (handler *respondentHandler) GetDetailRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetDetailRespondent(slug)
}

func (handler *respondentHandler) DeleteRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.DeleteRespondent(slug)
}

func (handler *respondentHandler) UpdateRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.UpdateRespondent(slug, req)
}

func (handler *respondentHandler) GetExampleImport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetExampleImport()
}

func (handler *respondentHandler) ImportRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.ImportRespondent(req)
}
