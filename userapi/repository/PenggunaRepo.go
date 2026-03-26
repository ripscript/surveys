package repository

import (
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/utils"
	"strconv"

	"gorm.io/gorm"
)

type PenggunaRepo interface {
	FindRespondentByRole(payload payloads.LoginPayload) (*models.Respondent, string, error)
	FindUserByRespondentID(id int) (*models.User, error)
	CountFailedLogin(userID int) (int64, error)
	InsertFailedLogin(userID int) error
	BlockRespondent(id int) error
	CheckActiveJabatan(id int) (bool, error)
	FindUserByID(id int) (*models.User, error)
}

type penggunaRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewPenggunaRepo(dbSlave, dbMaster *gorm.DB) *penggunaRepo {
	defer utils.GeneralRecover()
	return &penggunaRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *penggunaRepo) FindRespondentByRole(payload payloads.LoginPayload) (*models.Respondent, string, error) {
	var respondent models.Respondent
	var requestEmail string

	query := r.dbSlave.Where("role_id = ?", payload.Role).Where("deleted_at IS NULL")

	if payload.Role == 8 || payload.Role == 7 || payload.Role == 9 {
		query = query.Where("email = ?", payload.Email)
		requestEmail = payload.Email
	} else {
		if payload.SelectedKecamatan != nil {
			query = query.Where("kecamatan_id = ?", *payload.SelectedKecamatan)
		}
		if payload.SelectedKelurahan != nil {
			query = query.Where("kelurahan_id = ?", *payload.SelectedKelurahan)
		}
		if payload.SelectedRW != nil {
			query = query.Where("rw_id = ?", *payload.SelectedRW)
		}
		if payload.SelectedRT != nil {
			query = query.Where("rt_id = ?", *payload.SelectedRT)
		}
	}

	err := query.First(&respondent).Error
	if err != nil {
		return nil, "", err
	}

	if respondent.Email != "" {
		requestEmail = respondent.Email
	} else if respondent.Username != "" {
		requestEmail = respondent.Username
	} else {
		requestEmail = respondent.NIK
	}

	return &respondent, requestEmail, nil
}

func (r *penggunaRepo) FindUserByRespondentID(id int) (*models.User, error) {
	var user models.User
	err := r.dbSlave.Where("respondent_id = ?", id).First(&user).Error
	return &user, err
}

func (r *penggunaRepo) CountFailedLogin(userID int) (int64, error) {
	var count int64

	err := r.dbSlave.Model(&models.LogBlockLogin{}).
		Where("user_credential = ?", strconv.Itoa(userID)).
		Count(&count).Error

	return count, err
}

func (r *penggunaRepo) InsertFailedLogin(userID int) error {
	return r.dbMaster.Create(&models.LogBlockLogin{
		UserCredential: userID,
	}).Error
}

func (r *penggunaRepo) BlockRespondent(id int) error {
	return r.dbMaster.Model(&models.Respondent{}).
		Where("id = ?", id).
		Update("is_blocked", true).Error
}

func (r *penggunaRepo) CheckActiveJabatan(id int) (bool, error) {
	var data models.PejabatWilayah

	err := r.dbSlave.Where("id_responden = ?", id).
		Where("status_jabat = ?", 1).
		First(&data).Error

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *penggunaRepo) FindUserByID(id int) (*models.User, error) {
	var user models.User
	err := r.dbSlave.Where("id = ?", id).First(&user).Error
	return &user, err
}
