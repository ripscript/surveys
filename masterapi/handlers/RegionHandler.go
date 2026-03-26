package handlers

import (
	"backend/masterapi/models"
	"backend/masterapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type RegionHandler interface {
	KecamatanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	KelurahansOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RwOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RtOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type regionHandler struct {
	regionService service.RegionService
}

func NewRegionHandler(
	regionService service.RegionService,
) RegionHandler {
	return &regionHandler{
		regionService,
	}
}

func (handler *regionHandler) KecamatanOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.regionService.KecamatanOptions(param)
}

func (handler *regionHandler) KelurahansOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.regionService.KelurahansOptions(param)
}

func (handler *regionHandler) RwOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.regionService.RwOptions(param)
}

func (handler *regionHandler) RtOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.regionService.RtOptions(param)
}
