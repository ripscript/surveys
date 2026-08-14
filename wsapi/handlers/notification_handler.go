package handlers

import (
	"backend/siccore/pb"
	"backend/wsapi/models"
	"backend/wsapi/service"
	"context"
	"net/url"
)

type NotificationHandler interface {
	NotifyReportStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type notificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(
	notificationService service.NotificationService,
) NotificationHandler {
	return &notificationHandler{
		notificationService: notificationService,
	}
}

func (handler *notificationHandler) NotifyReportStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	return handler.notificationService.NotifyReportStatus(ctx, req, usr, param, slug)
}
