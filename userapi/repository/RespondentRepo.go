package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

type RespondentRepo interface {
	GetRespondent(offset int, limit int, param url.Values) ([]models.Respondents, int64, error)
	GetDetailRespondent(id int) (models.Respondents, error)
	DeleteUsers(deletedUsers models.DeleteRespondent) error
	UpdateUsers(id int, updateData models.UpdateRespondents) error
	GetRespondenById(id int) (*models.RawRespondents, error)
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

func (r *respondentRepo) GetRespondenById(id int) (*models.RawRespondents, error) {
	defer utils.GeneralRecover()

	db := r.dbSlave

	var responden models.RawRespondents
	err := db.Where("id = ?", id).Where("deleted_at IS NULL").First(&responden).Error
	if err != nil {
		return nil, err
	}

	return &responden, nil
}
