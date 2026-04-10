package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

type UsersBlokirRepo interface {
	GetListData(offset int, limit int, param url.Values) ([]models.RespondentBlock, int64, error)
	CheckUser(id int) (models.RespondentBlock, error)
	OpenBlokir(id int) error
}

type usersBlokirRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewUsersBlokirRepo(dbSlave, dbMaster *gorm.DB) *usersBlokirRepo {
	defer utils.GeneralRecover()
	return &usersBlokirRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *usersBlokirRepo) GetListData(offset int, limit int, param url.Values) ([]models.RespondentBlock, int64, error) {
	var usersList []models.RespondentBlock
	var total int64

	question := param.Get("question")
	answer := param.Get("answer")

	query := r.dbSlave.Where("deleted_at IS NULL").Where("is_blocked = ?", "true")

	if question != "" {
		query = query.Where("LOWER(question) LIKE ?", "%"+strings.ToLower(question)+"%")
	}
	if answer != "" {
		query = query.Where("LOWER(answer) LIKE ?", "%"+strings.ToLower(answer)+"%")
	}

	if err := query.Model(&models.RespondentBlock{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit != 1 {
		if err := query.Offset(offset).Limit(limit).Find(&usersList).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Offset(offset).Find(&usersList).Error; err != nil {
			return nil, 0, err
		}
	}

	for i := range usersList {
		usersList[i].No = offset + i + 1
	}

	return usersList, total, nil
}

func (r *usersBlokirRepo) CheckUser(id int) (models.RespondentBlock, error) {
	defer utils.GeneralRecover()
	var blockedRespondent models.RespondentBlock
	db := r.dbSlave
	err := db.Where("id = ?", id).First(&blockedRespondent).Error
	if err != nil {
		return blockedRespondent, err
	}

	return blockedRespondent, nil
}

func (r *usersBlokirRepo) OpenBlokir(id int) error {
	defer utils.GeneralRecover()
	var respondent models.Respondent

	var users models.Users
	db := r.dbMaster
	tx := db.Begin()

	err := tx.Model(respondent).Where("id = ?", id).Update("is_blocked", "false").Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Where("respondent_id = ?", id).First(&users).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	userId, err := utils.ToString(users.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Where("user_credential = ?", userId).Delete(&models.LogBlockLogin{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
