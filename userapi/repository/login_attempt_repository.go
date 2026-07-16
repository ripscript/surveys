package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"strconv"

	"gorm.io/gorm"
)

type LoginAttemptRepository interface {
	CountFailed(userID int) (int64, error)
	RecordFailed(userID int) error
	ResetFailed(userID int) error
}

type loginAttemptRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewloginAttemptRepository(dbSlave, dbMaster *gorm.DB) *loginAttemptRepository {
	defer utils.GeneralRecover()
	return &loginAttemptRepository{
		dbSlave,
		dbMaster,
	}
}

func (r *loginAttemptRepository) CountFailed(userID int) (int64, error) {
	var count int64
	err := r.dbSlave.Model(&models.LogBlockLogin{}).
		Where("user_credential = ?", strconv.Itoa(userID)).
		Count(&count).Error
	return count, err
}

func (r *loginAttemptRepository) RecordFailed(userID int) error {
	return r.dbMaster.Create(&models.LogBlockLogin{
		UserCredential: strconv.Itoa(userID),
	}).Error
}

func (r *loginAttemptRepository) ResetFailed(userID int) error {
	return r.dbMaster.
		Where("user_credential = ?", strconv.Itoa(userID)).
		Delete(&models.LogBlockLogin{}).Error
}
