package handlers

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/service"
	"context"
	"net/url"
)

type SurveyHandler interface {
	CreateSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SurveyBundlingSubmit(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHistoryDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	// RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	// VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type surveyHandler struct {
	manajemenAlurService service.ManajemenAlurService
	surveyService        service.SurveyService
}

func NewSurveyHandler(
	manajemenAlurService service.ManajemenAlurService,
	surveyService service.SurveyService,
) SurveyHandler {
	return &surveyHandler{
		manajemenAlurService: manajemenAlurService,
		surveyService:        surveyService,
	}
}

func (handler *surveyHandler) CreateSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.CreateSurvey(ctx, usr, req)
}

func (handler *surveyHandler) OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.OptionsPeriodeSurvey(usr, param)
}

func (handler *surveyHandler) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetListSurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) ApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ApprovalSurvey(ctx, usr, req, slug)
}

func (handler *surveyHandler) AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.AvailableSurveyWilayah(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.PreviewSurveyIndex(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.PreviewSurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) SurveyBundlingSubmit(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.SubmitSurveyAnswers(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.UpdateSurveyRespondentStatus(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetApprovalHistorySurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetHistoryDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetHistoryDetail(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetSurveyKewilayahan(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetDetailSurveyKewilayahan(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) SurveyResultIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.SurveyResultIndex(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) SurveyResultSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.SurveyResultSectionDetail(ctx, req, usr, param, slug)
}

// func (handler *surveyHandler) RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
// 	return handler.surveyService.RejectSurveyAnswers(ctx, req, usr, param, slug)
// }

// func (handler *surveyHandler) VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
// 	return handler.surveyService.ApproveSurveyAnswers(ctx, req, usr, param, slug)
// }
