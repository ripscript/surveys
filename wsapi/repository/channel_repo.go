package repository

import (
	"backend/wsapi/models"
	"backend/wsapi/utils"
	"time"

	"gorm.io/gorm"
)

type ChannelRepo interface {
	GetTicketByUserId(userId int64) (*models.WsTicket, error)
	CreateTicket(ticket *models.WsTicket) (*models.WsTicket, error)
	GetTicketByTicket(ticket string) (*models.WsTicket, error)
	DeleteTicket(ticketID int64) error
}

type channelRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewChannelRepo(dbSlave, dbMaster *gorm.DB) ChannelRepo {
	defer utils.GeneralRecover()
	return &channelRepo{
		dbSlave,
		dbMaster,
	}
}

func (repo *channelRepo) GetTicketByUserId(userId int64) (*models.WsTicket, error) {
	defer utils.GeneralRecover()

	var ticket models.WsTicket
	err := repo.dbSlave.Where("user_id = ? AND expires_at > ?", userId, time.Now()).First(&ticket).Error
	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (repo *channelRepo) GetTicketByTicket(ticket string) (*models.WsTicket, error) {
	defer utils.GeneralRecover()

	var wsTicket models.WsTicket
	err := repo.dbSlave.Where("ticket = ?", ticket).First(&wsTicket).Error
	if err != nil {
		return nil, err
	}

	return &wsTicket, nil
}

func (repo *channelRepo) CreateTicket(ticket *models.WsTicket) (*models.WsTicket, error) {
	defer utils.GeneralRecover()

	err := repo.dbMaster.Create(ticket).Error
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (repo *channelRepo) DeleteTicket(ticketID int64) error {
	defer utils.GeneralRecover()

	err := repo.dbMaster.Delete(&models.WsTicket{}, ticketID).Error
	if err != nil {
		return err
	}

	return nil
}
