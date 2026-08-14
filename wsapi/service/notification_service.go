package service

import (
	"backend/siccore/pb"
	"backend/wsapi/hub"
	"backend/wsapi/models"
	"backend/wsapi/request"
	"backend/wsapi/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
)

type NotificationService interface {
	NotifyReportStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type notificationService struct {
	hub *hub.Hub
}

func NewNotificationService(hub *hub.Hub) NotificationService {
	return &notificationService{hub: hub}
}

func (service *notificationService) NotifyReportStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload request.ReportStatusNotification

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	client, ok := service.hub.Get(int32(payload.UserID))
	if !ok {
		return utils.SendData(nil, "User tidak sedang terhubung ke websocket")
	}

	message, err := json.Marshal(map[string]interface{}{
		"event":         "report_status",
		"laporan_id":    payload.LaporanID,
		"success":       payload.Success,
		"document_id":   payload.DocumentID,
		"error_message": payload.ErrorMessage,
	})
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	select {
	case client.Send <- message:
	default:
		return utils.SendError(errors.New("gagal mengirim: buffer client penuh"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Notifikasi terkirim")
}
