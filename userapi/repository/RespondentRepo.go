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
	"strconv"
	"strings"
	"time"

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
	UpdateRespondent(respondent *models.RespondentRaw) (*models.RespondentRaw, error)
	WithTx(tx *gorm.DB) *respondentRepo
	GetRespondentById(respondentId int) (*models.RespondentRaw, error)
	CheckIsWilayahAvailable(roleID, kecamatanID, kelurahanID, rwID, rtID *int) (bool, error)

	RespondentOptions(req payloads.RespondentOptionsPayload) ([]response.OptionItem, int64, error)

	FindByRole(payload payloads.LoginPayload) (*models.Respondent, error)
	IsJabatanActive(respondentID int) (bool, error)
	Block(respondentID int) error
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
	hasJabatan := param.Get("has_jabatan")

	query := db.Model(data)

	if kecamatan != "" {
		kecamatanId, err := utils.ToInt64(kecamatan)
		if err != nil {
			return nil, err
		}
		query = query.Where("kecamatan_id = ?", kecamatanId)
	}
	if kelurahan_id != "" {
		kelurahanId, err := utils.ToInt64(kelurahan_id)
		if err != nil {
			return nil, err
		}
		query = query.Where("kelurahan_id = ?", kelurahanId)
	}
	if rw != "" {
		rwId, err := utils.ToInt64(rw)
		if err != nil {
			return nil, err
		}
		query = query.Where("rw_id = ?", rwId)
	}
	if rt != "" {
		rtId, err := utils.ToInt64(rt)
		if err != nil {
			return nil, err
		}
		query = query.Where("rt_id = ?", rtId)
	}
	if status != "" {
		switch status {
		case "active":
			query = query.Where("deleted_at IS NULL AND is_blocked = ?", "false")
		case "blocked":
			query = query.Where("deleted_at IS NULL AND is_blocked = ?", "true")
		case "inactive":
			query = query.Where("deleted_at IS NOT NULL")
		}
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	if hasJabatan != "" {
		hasJabatanBool, err := strconv.ParseBool(hasJabatan)
		if err != nil {
			return nil, err
		}

		if hasJabatanBool {
			// Ada relasi ke pejabat__wilayahs DAN status_jabat = 1
			query = query.Where(
				`EXISTS (
					SELECT 1 FROM pejabat__wilayahs pw
					WHERE pw.id_responden = respondents.id
					AND pw.status_jabat = 1
				)`,
			)
		} else {
			// Tidak ada relasi SAMA SEKALI, ATAU ada relasi tapi status_jabat = 0
			query = query.Where(
				`NOT EXISTS (
					SELECT 1 FROM pejabat__wilayahs pw
					WHERE pw.id_responden = respondents.id
					AND pw.status_jabat = 1
				)`,
			)
		}
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

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").
		Where("role_id NOT IN ?", []int{1, 6, 7, 8, 9}).
		Order("created_at desc")

	if search != "" {
		searchstr := "%" + strings.ToLower(search) + "%"
		query = query.Where(`
			LOWER(email) LIKE ? OR 
			LOWER(name) LIKE ? OR 
			LOWER(username) LIKE ? OR 
			LOWER(phone_number) LIKE ? OR
			nik ILIKE ?
		`, searchstr, searchstr, searchstr, searchstr, searchstr)
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
	dataUsers.MustChangePassword = true

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

func (repository *respondentRepo) UpdateRespondent(respondent *models.RespondentRaw) (*models.RespondentRaw, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Save(respondent).Error
	if err != nil {
		return nil, err
	}

	return respondent, nil
}

func (repository *respondentRepo) WithTx(tx *gorm.DB) *respondentRepo {
	return &respondentRepo{dbMaster: tx}
}

func (r *respondentRepo) GetRespondentById(respondentId int) (*models.RespondentRaw, error) {
	defer utils.GeneralRecover()
	var respondent models.RespondentRaw
	err := r.dbSlave.Where("id = ?", respondentId).First(&respondent).Error
	if err != nil {
		return nil, err
	}

	return &respondent, nil
}

func (repository *respondentRepo) CheckIsWilayahAvailable(roleID, kecamatanID, kelurahanID, rwID, rtID *int) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	query := repository.dbSlave.Model(&models.RespondentRaw{}).Where("role_id = ? AND deleted_at IS NULL", *roleID)

	if kecamatanID != nil {
		query = query.Where("kecamatan_id = ?", *kecamatanID)
	}
	if kelurahanID != nil {
		query = query.Where("kelurahan_id = ?", *kelurahanID)
	}
	if rwID != nil {
		query = query.Where("rw_id = ?", *rwID)
	}
	if rtID != nil {
		query = query.Where("rt_id = ?", *rtID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *respondentRepo) RespondentOptions(req payloads.RespondentOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()

	var data []response.OptionItem
	var totalData int64

	db := r.dbSlave.Table("respondents").
		Select(`
			respondents.id, 
			respondents.name AS label
		`)

	filterCond := r.dbSlave.Session(&gorm.Session{})

	if req.Q != "" {
		filterCond = filterCond.Where("respondents.name ILIKE ?", "%"+req.Q+"%")
	}

	if req.KecamatanId != "" {
		filterCond = filterCond.Where("respondents.kecamatan_id = ?", req.KecamatanId)
	}
	if req.KelurahanId != "" {
		filterCond = filterCond.Where("respondents.kelurahan_id = ?", req.KelurahanId)
	}
	if req.RWId != "" {
		filterCond = filterCond.Where("respondents.rw_id = ?", req.RWId)
	}
	if req.RTId != "" {
		filterCond = filterCond.Where("respondents.rt_id = ?", req.RTId)
	}

	if req.Status != "" {
		switch req.Status {
		case "active":
			filterCond = filterCond.Where("respondents.deleted_at IS NULL AND respondents.is_blocked = ?", false)
		case "blocked":
			filterCond = filterCond.Where("respondents.deleted_at IS NULL AND respondents.is_blocked = ?", true)
		case "inactive":
			filterCond = filterCond.Where("respondents.deleted_at IS NOT NULL")
		}
	} else {
		filterCond = filterCond.Where("respondents.deleted_at IS NULL")
	}

	if req.HasJabatan != "" {
		hasJabatanBool, err := strconv.ParseBool(req.HasJabatan)
		if err == nil {
			if hasJabatanBool {
				filterCond = filterCond.Where(
					`EXISTS (
						SELECT 1 FROM pejabat__wilayahs pw
						WHERE pw.id_responden = respondents.id
						AND pw.status_jabat = 1
					)`,
				)
			} else {
				filterCond = filterCond.Where(
					`NOT EXISTS (
						SELECT 1 FROM pejabat__wilayahs pw
						WHERE pw.id_responden = respondents.id
						AND pw.status_jabat = 1
					)`,
				)
			}
		}
	}

	if len(req.IDs) > 0 {
		db = db.Where(filterCond).Or("respondents.id IN ?", req.IDs)
	} else {
		db = db.Where(filterCond)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("respondents.name ASC")

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

var directLoginRoles = map[int]bool{
	1: true,
	6: true,
	7: true,
	8: true,
	9: true,
}

func isDirectLoginRole(role int) bool {
	return directLoginRoles[role]
}

func (r *respondentRepo) FindByRole(payload payloads.LoginPayload) (*models.Respondent, error) {
	var respondent models.Respondent

	query := r.buildCriteriaQuery(payload)

	if !isDirectLoginRole(payload.Role) {
		query = query.
			Joins("JOIN pejabat__wilayahs pw ON pw.id_responden = respondents.id").
			Where("pw.status_jabat = ?", 1).
			Where("pw.periode_akhir IS NULL OR pw.periode_akhir >= ?", time.Now())
	}

	err := query.First(&respondent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRespondentNotFound
		}
		return nil, err
	}

	return &respondent, nil
}

func (r *respondentRepo) buildCriteriaQuery(payload payloads.LoginPayload) *gorm.DB {
	query := r.dbSlave.
		Where("respondents.role_id = ?", payload.Role).
		Where("respondents.deleted_at IS NULL")

	if isDirectLoginRole(payload.Role) {
		query = query.Where("respondents.email = ?", payload.Email)
		return query
	}

	if payload.SelectedKecamatan != nil {
		query = query.Where("respondents.kecamatan_id = ?", *payload.SelectedKecamatan)
	}
	if payload.SelectedKelurahan != nil {
		query = query.Where("respondents.kelurahan_id = ?", *payload.SelectedKelurahan)
	}
	if payload.SelectedRW != nil {
		query = query.Where("respondents.rw_id = ?", *payload.SelectedRW)
	}
	if payload.SelectedRT != nil {
		query = query.Where("respondents.rt_id = ?", *payload.SelectedRT)
	}

	return query
}

func (r *respondentRepo) IsJabatanActive(respondentID int) (bool, error) {
	var data models.PejabatWilayah

	err := r.dbSlave.
		Where("id_responden = ?", respondentID).
		Where("status_jabat = ?", 1).
		Where("periode_akhir IS NULL OR periode_akhir >= ?", time.Now()).
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *respondentRepo) Block(respondentID int) error {
	return r.dbMaster.Model(&models.Respondent{}).
		Where("id = ?", respondentID).
		Update("is_blocked", true).Error
}
