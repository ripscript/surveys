package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"crypto/rand"
	"encoding/base64"
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
	UserExport() ([]models.ExportUsers, error)
	StoreUsers(data models.CreateRespondent) error
	CheckEmail(email string) (int64, int64, error)
	CheckPhoneNumber(phoneNumber string) (int64, error)
	CheckNik(nik string) (int64, int64, error)
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

func (r *usersRepo) UserExport() ([]models.ExportUsers, error) {
	defer utils.GeneralRecover()
	var users []models.ExportUsers
	db := r.dbSlave

	err := db.Where("role_id = ? AND deleted_at IS NUlL", 8).Find(&users).Error
	if err != nil {
		return users, err
	}

	return users, nil
}

func (r *usersRepo) CheckEmail(email string) (int64, int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	var countUsers int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("email = ?", email).Count(&countRespondent).Error
	if err != nil {
		return 0, 0, err
	}
	err = db.Model(models.Users{}).Where("email = ?", email).Count(&countUsers).Error
	if err != nil {
		return 0, 0, err
	}

	return countRespondent, countUsers, nil
}

func (r *usersRepo) CheckPhoneNumber(phoneNumber string) (int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("phone_number = ?", phoneNumber).Count(&countRespondent).Error
	if err != nil {
		return 0, err
	}

	return countRespondent, nil
}

func (r *usersRepo) CheckNik(nik string) (int64, int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	var countUsers int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("nik = ?", nik).Count(&countRespondent).Error
	if err != nil {
		return 0, 0, err
	}
	err = db.Model(models.Users{}).Where("nik = ?", nik).Count(&countUsers).Error
	if err != nil {
		return 0, 0, err
	}

	return countRespondent, countUsers, nil
}

func (r *usersRepo) StoreUsers(data models.CreateRespondent) error {
	defer utils.GeneralRecover()
	db := r.dbMaster
	tx := db.Begin()

	err := tx.Create(&data).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		tx.Rollback()
		return err
	}

	randomBytes := make([]byte, 48)
	if _, err := rand.Read(randomBytes); err != nil {
		tx.Rollback()
		return err
	}

	var dataUsers models.StoreUsers
	dataUsers.FirstName = data.Name
	dataUsers.LastName = strings.ReplaceAll(strings.ToLower(data.Name), " ", "-")
	dataUsers.Email = data.Email
	dataUsers.Password = hashedPassword
	dataUsers.EmailToken = base64.URLEncoding.EncodeToString(randomBytes)
	dataUsers.RespondentId = data.Id
	dataUsers.CreatedAt = utils.TimeNow()
	dataUsers.UpdatedAt = utils.TimeNow()

	err = tx.Create(&dataUsers).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
