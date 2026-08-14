package handlers

import (
	"backend/siccore/pb"
	"backend/wsapi/models"
	"backend/wsapi/utils"
	"context"
	"net/url"
)

type HealthyHandler interface {
	Healthy(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type healthyHandler struct {
}

func NewHealthyHandler() HealthyHandler {
	return &healthyHandler{}
}

func (handler *healthyHandler) Healthy(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return utils.SendData(nil, "wsapi is healthy")
}
