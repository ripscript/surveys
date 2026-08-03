package handlers

import (
	"backend/docapi/models"
	"backend/docapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type UploadHandler interface {
	UploadAvatar(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadIdentity(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadFile(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ViewFIle(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadBank(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadEvent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SendFileWebchat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SendFileWithdraw(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowPublicSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteBulkSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	UploadCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteBulkCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	UploadLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteBulkLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetLaporanKontenImageBytes(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type uploadHandler struct {
	uploadService service.UploadService
}

func NewUploadHandler(
	uploadService service.UploadService,
) UploadHandler {
	return &uploadHandler{
		uploadService,
	}
}

func (handler *uploadHandler) SendFileWithdraw(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "withdraw", "")
}

func (handler *uploadHandler) SendFileWebchat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "webchat", "")
}

func (handler *uploadHandler) UploadBank(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "bank", "")
}

func (handler *uploadHandler) UploadAvatar(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "avatar", "")
}

func (handler *uploadHandler) UploadIdentity(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "identity", "")
}

func (handler *uploadHandler) UploadFile(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "upload", "")
}

func (handler *uploadHandler) UploadEvent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadFile(usr, req, param, "event", "")
}

func (handler *uploadHandler) ViewFIle(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.Show(slug)
}

func (handler *uploadHandler) UploadSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadSurveyImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) ShowSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.ShowSurveyImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) ShowPublicSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.ShowPublicSurveyImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) DeleteBulkSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.DeleteBulkSurveyImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) UploadCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadCMSImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) ShowCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.ShowCMSImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) DeleteBulkCMSImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.DeleteBulkCMSImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) UploadLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.UploadLaporanKontenImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) ShowLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.ShowLaporanKontenImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) DeleteBulkLaporanKontenImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.DeleteBulkLaporanKontenImage(ctx, req, usr, param, slug)
}

func (handler *uploadHandler) GetLaporanKontenImageBytes(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.uploadService.GetLaporanKontenImageBytes(ctx, req, usr, param, slug)
}
