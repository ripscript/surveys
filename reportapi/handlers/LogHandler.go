package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type LogHandler interface {
	SaveLogActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetLogActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type logHandler struct {
	logService service.LogService
}

func NewLogHandler(
	logService service.LogService,
) LogHandler {
	return &logHandler{
		logService,
	}
}

func (handler *logHandler) SaveLogActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.logService.SaveLogActivities(req)
}

func (handler *logHandler) GetLogActivities(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.logService.GetLogActivities(param)
}
