package repository

import (
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/response"
	"backend/userapi/utils"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
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
	GetRespondenById(id int) (*models.RawRespondents, error)
	GetSurveyorOptions(req payloads.SurveyorOptionsPayload) ([]response.OptionItem, int64, error)
	SurveyorOption() ([]models.SurveyorOptions, error)
	CheckRespondent(idRespondent int64) (models.BlockRespondent, error)
	RespondentBlock(data models.BlockRespondent) error
	GetOptionsRespondent(param url.Values) ([]models.RespondentOptions, error)

	GetRespondentByKecamatanId(ctx context.Context, kecamatanID int64) (*models.RawRespondents, error)
	GetRespondentByKelurahanId(ctx context.Context, kelurahanID int64) (*models.RawRespondents, error)
	GetRespondentByRWId(ctx context.Context, rwID int64) (*models.RawRespondents, error)
	GetRespondentByRTId(ctx context.Context, rtID int64) (*models.RawRespondents, error)
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

func (r *respondentRepo) GetOptionsRespondent(param url.Values) ([]models.RespondentOptions, error) {
	defer utils.GeneralRecover()
	var data []models.RespondentOptions
	db := r.dbSlave

	kecamatan := param.Get("kecamatan_id")
	kelurahan_id := param.Get("kelurahan_id")
	rw := param.Get("rw")
	rt := param.Get("rt")
	status := param.Get("status")

	query := db.Model(data)

	if kecamatan != "" {
		kecamatanId, err := utils.ToInt64(kecamatan)
		if err != nil {
			return nil, err
		}
		query.Where("kecamatan_id = ?", kecamatanId)
	}
	if kelurahan_id != "" {
		kelurahanId, err := utils.ToInt64(kelurahan_id)
		if err != nil {
			return nil, err
		}
		query.Where("kelurahan_id = ?", kelurahanId)
	}
	if rw != "" {
		rwId, err := utils.ToInt64(rw)
		if err != nil {
			return nil, err
		}
		query.Where("rw_id = ?", rwId)
	}
	if rt != "" {
		rtId, err := utils.ToInt64(rt)
		if err != nil {
			return nil, err
		}
		query.Where("rt_id = ?", rtId)
	}
	if status != "" {
		if status == "active" {
			query.Where("deleted_at IS NULL AND is_blocked = ?", "false")
		} else if status == "blocked" {
			query.Where("deleted_at IS NULL AND is_blocked = ?", "true")
		} else if status == "inactive" {
			query.Where("deleted_at IS NOT NULL")
		}
	} else {
		query.Where("deleted_at IS NULL")
	}

	err := query.Find(&data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *respondentRepo) RespondentBlock(data models.BlockRespondent) error {
	defer utils.GeneralRecover()
	db := r.dbMaster

	err := db.Where("id = ?", data.ID).Updates(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *respondentRepo) CheckRespondent(idRespondent int64) (models.BlockRespondent, error) {
	defer utils.GeneralRecover()
	var data models.BlockRespondent
	db := r.dbSlave

	err := db.Where("id = ?", idRespondent).First(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
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

	search := param.Get("search")
	name := param.Get("name")
	email := param.Get("email")
	nik := param.Get("nik")
	phoneNumber := param.Get("phoneNumber")
	kecamatan := param.Get("kecamatan")
	kelurahan := param.Get("kelurahan")
	rw := param.Get("rw")
	rt := param.Get("rt")
	status := param.Get("status")

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin")

	if search != "" {
		query = query.Where("LOWER(email) LIKE ? OR LOWER(name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(phone_number) LIKE ?", "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%")
	}

	if name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}
	if email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(email)+"%")
	}
	if nik != "" {
		query = query.Where("LOWER(nik) LIKE ?", "%"+strings.ToLower(nik)+"%")
	}
	if phoneNumber != "" {
		query = query.Where("LOWER(phone_number) LIKE ?", "%"+strings.ToLower(phoneNumber)+"%")
	}

	if kecamatan != "" {
		query = query.Where("kecamatan_id = ?", kecamatan)
	}

	if kelurahan != "" {
		query = query.Where("kelurahan_id = ?", kelurahan)
	}

	if rw != "" {
		query = query.Where("rw_id = ?", rw)
	}

	if rt != "" {
		query = query.Where("rt_id = ?", rt)
	}

	if status != "" {
		if status == "active" {
			query.Where("deleted_at IS NULL AND is_blocked = ?", "false")
		} else if status == "blocked" {
			query.Where("deleted_at IS NULL AND is_blocked = ?", "true")
		} else if status == "inactive" {
			query.Where("deleted_at IS NOT NULL")
		}
	} else {
		query.Where("deleted_at IS NULL")
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
		if !respondentList[i].DeletedAt.IsZero() {
			respondentList[i].Status = "inactive"
		} else if respondentList[i].IsBlocked == "true" {
			respondentList[i].Status = "blocked"
		} else if respondentList[i].IsBlocked == "false" {
			respondentList[i].Status = "active"
		}
	}

	// err := utils.SaveLogActivities()
	// if err != nil {

	// }

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

func (r *respondentRepo) SurveyorOption() ([]models.SurveyorOptions, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyorOptions
	db := r.dbSlave

	err := db.Where("role_id = ?", 8).Find(&data).Error
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return nil, err
		}
	}

	return data, nil
}

func (repository *respondentRepo) GetSurveyorOptions(req payloads.SurveyorOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("respondents").
		Select(`
			respondents.id, 
			respondents.name as label
		`).
		Where("respondents.role_id = ? AND respondents.deleted_at IS NULL", 8)

	if len(req.IDs) > 0 {
		db = db.Where("respondents.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("respondents.name ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("respondents.name asc")

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	err = db.Limit(limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (repository *respondentRepo) GetRespondentByKecamatanId(ctx context.Context, kecamatanID int64) (*models.RawRespondents, error) {
	defer utils.GeneralRecover()

	var respondent models.RawRespondents
	err := repository.dbSlave.WithContext(ctx).
		Where("kecamatan_id = ? AND role_id = ? AND deleted_at IS NULL", kecamatanID, 5).
		First(&respondent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &respondent, nil
}

func (repository *respondentRepo) GetRespondentByKelurahanId(ctx context.Context, kelurahanID int64) (*models.RawRespondents, error) {
	defer utils.GeneralRecover()

	var respondent models.RawRespondents
	err := repository.dbSlave.WithContext(ctx).
		Where("kelurahan_id = ? AND role_id = ? AND deleted_at IS NULL", kelurahanID, 4).
		First(&respondent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &respondent, nil
}

func (repository *respondentRepo) GetRespondentByRWId(ctx context.Context, rwID int64) (*models.RawRespondents, error) {
	defer utils.GeneralRecover()

	var respondent models.RawRespondents
	err := repository.dbSlave.WithContext(ctx).
		Where("rw_id = ? AND role_id = ? AND deleted_at IS NULL", rwID, 3).
		First(&respondent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &respondent, nil
}

func (repository *respondentRepo) GetRespondentByRTId(ctx context.Context, rtID int64) (*models.RawRespondents, error) {
	defer utils.GeneralRecover()

	var respondent models.RawRespondents
	err := repository.dbSlave.WithContext(ctx).
		Where("rt_id = ? AND role_id = ? AND deleted_at IS NULL", rtID, 2).
		First(&respondent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &respondent, nil
}
