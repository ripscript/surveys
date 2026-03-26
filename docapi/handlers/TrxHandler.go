package handlers

import (
	"backend/docapi/models"
	"backend/docapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type TrxHandler interface {
	GenerateReceipt(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadBanner(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type trxHandler struct {
	trxService service.TrxService
}

func NewTrxHandler(
	trxService service.TrxService,
) TrxHandler {
	return &trxHandler{
		trxService,
	}
}

func (handler *trxHandler) GenerateReceipt(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.trxService.GenerateReceipt(req)
}

func (handler *trxHandler) UploadBanner(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.trxService.UploadBanner(usr, req, param, "banner", "")
}
