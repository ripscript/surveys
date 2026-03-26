package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UsersRepo interface {
	GetUsers(offset int, limit int, param url.Values) ([]models.Respondents, int64, error)
	GetDetailUsers(id int) (models.Respondents, error)
	UpdateUsers(id int, updateData models.UpdateRespondent) error
	DeleteUsers(deletedUsers models.DeleteRespondent) error
	ResetPasswordUsers(id int) error
}

type usersRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewUsersRepo(dbSlave, dbMaster *gorm.DB) *usersRepo {
	defer utils.GeneralRecover()
	return &usersRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *usersRepo) GetUsers(offset int, limit int, param url.Values) ([]models.Respondents, int64, error) {
	var usersList []models.Respondents
	var total int64

	question := param.Get("question")
	answer := param.Get("answer")

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL").Where("role_id IN (?)", []int{7, 8})

	if question != "" {
		query = query.Where("LOWER(question) LIKE ?", "%"+strings.ToLower(question)+"%")
	}
	if answer != "" {
		query = query.Where("LOWER(answer) LIKE ?", "%"+strings.ToLower(answer)+"%")
	}

	if err := query.Model(&models.Respondents{}).Count(&total).Error; err != nil {
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
		var role string

		if usersList[i].RoleId == 7 {
			role = "Admin"
		} else {
			role = "Surveyor"
		}
		usersList[i].No = offset + i + 1
		usersList[i].Kecamatan = usersList[i].KecamatanJoin.SubDistrictName
		usersList[i].Kelurahan = usersList[i].KelurahanJoin.VillageName
		usersList[i].Rw = usersList[i].RwJoin.NamaRw
		usersList[i].Rt = usersList[i].RtJoin.NamaRt
		usersList[i].Role = role
	}

	return usersList, total, nil
}

func (r *usersRepo) GetDetailUsers(id int) (models.Respondents, error) {
	defer utils.GeneralRecover()
	var usersList models.Respondents

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL").Where("id = ?", id)
	if err := query.Find(&usersList).Error; err != nil {
		return usersList, err
	}

	var role string

	if usersList.RoleId == 7 {
		role = "Admin"
	} else {
		role = "Surveyor"
	}

	usersList.Kecamatan = usersList.KecamatanJoin.SubDistrictName
	usersList.Kelurahan = usersList.KelurahanJoin.VillageName
	usersList.Rw = usersList.RwJoin.NamaRw
	usersList.Rt = usersList.RtJoin.NamaRt
	usersList.Role = role

	return usersList, nil
}

func (r *usersRepo) UpdateUsers(id int, updateData models.UpdateRespondent) error {
	defer utils.GeneralRecover()
	var modelsUpdate models.UpdateRespondent
	db := r.dbMaster
	err := db.Model(modelsUpdate).Where("id = ?", id).Updates(updateData).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *usersRepo) DeleteUsers(deletedUsers models.DeleteRespondent) error {
	defer utils.GeneralRecover()
	var modelsDelete models.DeleteRespondent
	db := r.dbMaster
	err := db.Model(modelsDelete).Where("id = ?", deletedUsers.Id).Updates(deletedUsers).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *usersRepo) ResetPasswordUsers(id int) error {
	defer utils.GeneralRecover()

	var users models.Users

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		return err
	}

	err = r.dbMaster.Where("respondent_id = ?", id).First(&users).Error
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}

	err = r.dbMaster.Model(&models.Users{}).
		Where("respondent_id = ?", id).
		Updates(updates).Error
	if err != nil {
		return err
	}

	err = r.dbMaster.Where("user_id = ?", users.ID).
		Delete(&models.LogLogin{}).Error
	if err != nil {
		return err
	}

	return nil
}
