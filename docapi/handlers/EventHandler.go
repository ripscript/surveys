package handlers

import (
	"backend/docapi/models"
	"backend/docapi/service"
	pb "backend/siccore/pb"
	"context"
	"net/url"
)

type EventHandler interface {
	UploadProduct(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ViewProduct(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type eventHandler struct {
	eventService service.UploadService
}

func NewEventHandler(
	eventService service.UploadService,
) EventHandler {
	return &eventHandler{
		eventService,
	}
}

func (handler *eventHandler) UploadProduct(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.eventService.UploadFile(usr, req, param, "product", "")
}

func (handler *eventHandler) ViewProduct(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.eventService.Show(slug)
}
