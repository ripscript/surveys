package service

import (
	"backend/siccore/pb"
	"backend/wsapi/models"
	"backend/wsapi/repository"
	"backend/wsapi/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ChannelService interface {
	GetTicket(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type channelService struct {
	channelRepo repository.ChannelRepo
}

func NewChannelService(
	channelRepo repository.ChannelRepo,
) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
	}
}

const maxTicketGenerationAttempts = 5

func (service *channelService) GetTicket(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var userId int64
	if usr.ID != 0 {
		userId = usr.ID
	}

	getTicket, err := service.channelRepo.GetTicketByUserId(userId)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	if getTicket == nil {
		createdTicket, err := service.createTicketWithRetry(userId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		return utils.SendData(createdTicket, "Ticket websocket berhasil dibuat")
	}

	return utils.SendData(getTicket, "Ticket websocket sudah ada")
}

func (service *channelService) createTicketWithRetry(userId int64) (*models.WsTicket, error) {
	var lastErr error

	for attempt := 0; attempt < maxTicketGenerationAttempts; attempt++ {
		ticket, err := utils.GenerateTicket()
		if err != nil {
			return nil, err
		}

		newTicket := &models.WsTicket{
			Ticket:    ticket,
			UserID:    int32(userId),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		createdTicket, err := service.channelRepo.CreateTicket(newTicket)
		if err == nil {
			return createdTicket, nil
		}

		if !isUniqueViolation(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed to generate unique ticket after %d attempts: %w", maxTicketGenerationAttempts, lastErr)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
