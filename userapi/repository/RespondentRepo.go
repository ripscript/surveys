package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

type RespondentRepo interface {
	GetKecamatanByName(kecamatan string) (models.KecamatanOptions, error)
	GetKelurahanByName(kelurahan string) (models.KelurahanOptions, error)
	GetRespondent(offset int, limit int, param url.Values) ([]models.Respondents, int64, error)
	GetDetailRespondent(id int) (models.Respondents, error)
	DeleteUsers(deletedUsers models.DeleteRespondent) error
	UpdateUsers(id int, updateData models.UpdateRespondents) error
	BeginTx() *gorm.DB
	StoreUsers(tx *gorm.DB, data models.CreateRespondents) error
}

type respondentRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewRespondentRepo(dbSlave, dbMaster *gorm.DB) *respondentRepo {
	defer utils.GeneralRecover()
	return &respondentRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *respondentRepo) GetKecamatanByName(kecamatan string) (models.KecamatanOptions, error) {
	defer utils.GeneralRecover()
	var data models.KecamatanOptions
	db := r.dbSlave
	kecamatan = strings.TrimSpace(kecamatan)
	kecamatan = strings.ToLower(kecamatan)
	err := db.Where("LOWER(sub_district_name) LIKE ?", "%"+kecamatan+"%").First(&data).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return data, fmt.Errorf("Data Kecamatan Tidak Di Temukan")
		} else {
			return data, err
		}
	}

	return data, nil
}

func (r *respondentRepo) GetKelurahanByName(kelurahan string) (models.KelurahanOptions, error) {
	defer utils.GeneralRecover()
	var data models.KelurahanOptions
	db := r.dbSlave
	kelurahan = strings.TrimSpace(kelurahan)
	kelurahan = strings.ToLower(kelurahan)
	err := db.Where("LOWER(village_name) LIKE ?", "%"+kelurahan+"%").First(&data).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return data, fmt.Errorf("Data Kelurahan Tidak Di Temukan")
		} else {
			return data, err
		}
	}

	return data, nil
}

func (r *respondentRepo) GetRespondent(offset int, limit int, param url.Values) ([]models.Respondents, int64, error) {
	var respondentList []models.Respondents
	var total int64

	question := param.Get("question")
	answer := param.Get("answer")

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL")

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
		if err := query.Offset(offset).Limit(limit).Find(&respondentList).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Offset(offset).Find(&respondentList).Error; err != nil {
			return nil, 0, err
		}
	}

	for i := range respondentList {
		respondentList[i].No = offset + i + 1
		respondentList[i].Kecamatan = respondentList[i].KecamatanJoin.SubDistrictName
		respondentList[i].Kelurahan = respondentList[i].KelurahanJoin.VillageName
		respondentList[i].Rw = respondentList[i].RwJoin.NamaRw
		respondentList[i].Rt = respondentList[i].RtJoin.NamaRt
	}

	return respondentList, total, nil
}

func (r *respondentRepo) GetDetailRespondent(id int) (models.Respondents, error) {
	var respondentList models.Respondents

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL").Where("id = ?", id)
	if err := query.Find(&respondentList).Error; err != nil {
		return respondentList, err
	}

	respondentList.Kecamatan = respondentList.KecamatanJoin.SubDistrictName
	respondentList.Kelurahan = respondentList.KelurahanJoin.VillageName
	respondentList.Rw = respondentList.RwJoin.NamaRw
	respondentList.Rt = respondentList.RtJoin.NamaRt

	return respondentList, nil
}

func (r *respondentRepo) DeleteUsers(deletedUsers models.DeleteRespondent) error {
	defer utils.GeneralRecover()
	var modelsDelete models.DeleteRespondent
	db := r.dbMaster
	err := db.Model(modelsDelete).Where("id = ?", deletedUsers.Id).Updates(deletedUsers).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *respondentRepo) UpdateUsers(id int, updateData models.UpdateRespondents) error {
	defer utils.GeneralRecover()
	var modelsUpdate models.UpdateRespondents
	db := r.dbMaster
	err := db.Model(modelsUpdate).Where("id = ?", id).Updates(updateData).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *respondentRepo) BeginTx() *gorm.DB {
	defer utils.GeneralRecover()
	return r.dbMaster.Begin()
}

func (r *respondentRepo) StoreUsers(tx *gorm.DB, data models.CreateRespondents) error {
	defer utils.GeneralRecover()

	err := tx.Create(&data).Error
	if err != nil {
		return err
	}

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		return err
	}

	randomBytes := make([]byte, 48)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}

	var dataUsers models.StoreUsers
	dataUsers.FirstName = data.Name
	dataUsers.LastName = strings.ReplaceAll(strings.ToLower(data.Name), " ", "-")
	dataUsers.Email = data.Email
	dataUsers.Password = hashedPassword
	dataUsers.EmailToken = base64.URLEncoding.EncodeToString(randomBytes)
	dataUsers.RespondentId = data.Id
	dataUsers.Nik = data.NIK
	dataUsers.CreatedAt = utils.TimeNow()
	dataUsers.UpdatedAt = utils.TimeNow()

	err = tx.Create(&dataUsers).Error
	if err != nil {
		return err
	}

	return nil
}
