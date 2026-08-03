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
	DetailSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SurveyBundlingSubmit(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetHistoryApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHistoryDetailPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	RejectValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetPublicImageSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelSurveyResultsPerRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelSurveyResultsMassal(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	ResetStatusToVerifySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetAllRejectedQuestions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRejectedQuestionsBySurveyCode(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (handler *surveyHandler) DetailSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetDetailSurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) OptionsPeriodeSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.OptionsPeriodeSurvey(usr, param)
}

func (handler *surveyHandler) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	go func() {
		bgCtx := context.Background()
		handler.surveyService.SyncExpiredSurveysStatus(bgCtx)
	}()

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

func (handler *surveyHandler) GetHistoryApprovalSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetApprovalHistorySurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetApprovalHistorySurveyPerWilayah(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetHistoryDetailPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetHistoryDetailPerWilayah(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	go func() {
		bgCtx := context.Background()
		handler.surveyService.SyncExpiredSurveysStatus(bgCtx)
	}()
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

func (handler *surveyHandler) RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.RejectSurveyAnswers(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.VerifySurveyAnswers(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) RejectValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.RejectValidateSurveyAnswers(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) ValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ValidateSurveyAnswers(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetPublicImageSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetPublicImageSurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) SurveyOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.SurveyOptions(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) ExportExcelSurveyResultsPerRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ExportExcelSurveyResultsPerRT(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) ExportExcelSurveyResultsMassal(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ExportExcelSurveyResultsMassal(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) ResetStatusToVerifySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.ResetStatusToVerifySurvey(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetAllRejectedQuestions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetAllRejectedQuestions(ctx, req, usr, param, slug)
}

func (handler *surveyHandler) GetRejectedQuestionsBySurveyCode(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.surveyService.GetRejectedQuestionsBySurveyCode(ctx, req, usr, param, slug)
}
