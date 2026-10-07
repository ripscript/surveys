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
	GetRawDetailRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyorOption(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	BlockRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetOptionsRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetRespondentByKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByRW(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	UpdatePasswordRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GeneratePassword(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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
	return handler.respondentService.DeleteRespondent(usr, slug)
}

func (handler *respondentHandler) UpdateRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.UpdateRespondent(usr, slug, req)
}

func (handler *respondentHandler) GetExampleImport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetExampleImport(ctx)
}

func (handler *respondentHandler) ImportRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.ImportRespondentV2(usr, req)
}

func (handler *respondentHandler) GetRawDetailRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRawDetailRespondent(slug)
}

func (handler *respondentHandler) SurveyorOption(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.SurveyorOption(param)
}

func (handler *respondentHandler) BlockRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.BlockRespondent(usr, param)
}

func (handler *respondentHandler) GetOptionsRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.RespondentOptions(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) GetRespondentByKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRespondentByKecamatan(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) GetRespondentByKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRespondentByKelurahan(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) GetRespondentByRW(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRespondentByRW(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) GetRespondentByRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GetRespondentByRT(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) UpdatePasswordRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.UpdatePasswordRespondent(ctx, req, usr, param, slug)
}

func (handler *respondentHandler) GeneratePassword(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.respondentService.GeneratePassword(ctx, req, usr, param, slug)
}
