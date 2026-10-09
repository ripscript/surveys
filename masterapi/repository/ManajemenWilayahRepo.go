package repository

import (
	"backend/masterapi/enums"
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type ManajemenWilayahRepo interface {
	CreateKecamatan(kecamatan models.Kecamatan) (*models.Kecamatan, error)
	GetKecamatanByID(id int) (*models.Kecamatan, error)
	UpdateKecamatan(kecamatan models.Kecamatan) (models.Kecamatan, error)
	GetKecamatanByName(name string) (*models.Kecamatan, error)
	GetKecamatanByNameToLower(name string) (*models.Kecamatan, error)
	GetKecamatanBySlug(slug string) (*models.Kecamatan, error)
	GetListKecamatan(req payloads.DatatablePayload, respondent *models.Respondent) ([]models.KecamatanDatatableResponse, int64, error)
	GetKecamatanOptions(req payloads.KecamatanOptionsPayload) ([]response.OptionItem, int64, error)
	IsKecamatanUsed(id int64) (bool, error)
	DeleteKecamatanById(id int64) error

	CreateKelurahan(kelurahan models.Kelurahan) (*models.Kelurahan, error)
	GetKelurahanByNameToLower(name string) (*models.Kelurahan, error)
	GetKelurahanByID(id int) (*models.KelurahanDetail, error)
	GetKelurahanByName(name string) (*models.KelurahanDetail, error)
	GetKelurahanBySlug(slug string) (*models.KelurahanDetail, error)
	UpdateKelurahan(kelurahan models.Kelurahan) (*models.Kelurahan, error)
	GetListKelurahan(req payloads.DatatablePayload, kecamatanId *int64, respondent *models.Respondent) ([]models.KelurahanDatatableResponse, int64, error)
	GetKelurahanOptions(req payloads.KelurahanOptionsPayload) ([]response.OptionItem, int64, error)
	IsKelurahanUsed(id int64) (bool, error)
	DeleteKelurahanById(id int64) error

	GetRwByID(id int) (*models.DataRwDetail, error)
	GetRwByName(name string, kelurahanId *int64) (*models.DataRwDetail, error)
	UpdateRw(rw models.DataRw) (*models.DataRw, error)
	GetListRw(req payloads.DatatablePayload, kelurahanId *int64, respondent *models.Respondent) ([]models.RwDatatableResponse, int64, error)
	CreateRw(rw models.DataRw) (*models.DataRw, error)
	GetRwOptions(req payloads.RwOptionsPayload) ([]response.OptionItem, int64, error)
	IsRwUsed(id int64) (bool, error)
	DeleteRwById(id int64) error

	GetRtByID(id int64) (*models.DataRtDetail, error)
	GetRtByName(name string, rwId *int64) (*models.DataRtDetail, error)
	UpdateRt(rw models.DataRt) (*models.DataRt, error)
	GetListRt(req payloads.DatatablePayload, rwId *int64) ([]models.RtDatatableResponse, int64, error)
	CreateRt(rt models.DataRt) (*models.DataRt, error)
	GetRtOptions(req payloads.RtOptionsPayload) ([]response.OptionItem, int64, error)
	IsRtUsed(id int64) (bool, error)
	DeleteRtById(id int64) error

	GetListKecamatanV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultKecamatan, int64, error)
	GetListKelurahanV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultKelurahan, int64, error)
	GetListRWV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultRW, int64, error)
	GetListRTV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultRT, int64, error)

	GetWilayahSummary(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload) (response.WilayahSummary, error)
}

type manajemenWilayahRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenWilayahRepo(dbSlave, dbMaster *gorm.DB) *manajemenWilayahRepo {
	defer utils.GeneralRecover()
	return &manajemenWilayahRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenWilayahRepo) CreateKecamatan(kecamatan models.Kecamatan) (*models.Kecamatan, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&kecamatan).Error
	if err != nil {
		return nil, err
	}

	return &kecamatan, nil
}

func (repository *manajemenWilayahRepo) GetKecamatanByID(id int) (*models.Kecamatan, error) {
	defer utils.GeneralRecover()
	var data models.Kecamatan
	db := repository.dbSlave

	err := db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetKecamatanByName(name string) (*models.Kecamatan, error) {
	defer utils.GeneralRecover()
	var data models.Kecamatan
	db := repository.dbSlave

	err := db.Where("sub_district_name = ?", name).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetKecamatanByNameToLower(name string) (*models.Kecamatan, error) {
	defer utils.GeneralRecover()
	var data models.Kecamatan
	db := repository.dbSlave

	err := db.Where("LOWER(sub_district_name) = ?", strings.ToLower(name)).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetKecamatanBySlug(slug string) (*models.Kecamatan, error) {
	defer utils.GeneralRecover()
	var data models.Kecamatan
	db := repository.dbSlave

	err := db.Where("sub_district_slug = ?", slug).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) UpdateKecamatan(kecamatan models.Kecamatan) (models.Kecamatan, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Save(&kecamatan).Error
	if err != nil {
		return kecamatan, err
	}

	return kecamatan, nil
}

func (repository *manajemenWilayahRepo) GetListKecamatan(req payloads.DatatablePayload, respondent *models.Respondent) ([]models.KecamatanDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.KecamatanDatatableResponse
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	pejabatJoinQuery := `LEFT JOIN (
		SELECT DISTINCT ON (id_wilayah) id_wilayah, id_responden, periode_awal, periode_akhir
		FROM pejabat__wilayahs
		WHERE tipe_wilayah = 5
		ORDER BY id_wilayah, periode_akhir DESC NULLS FIRST
	) pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = kecamatans.id`

	db := repository.dbSlave.Table("kecamatans").
		Select(`
			kecamatans.id, 
			kecamatans.sub_district_name, 
			kecamatans.sub_district_slug, 
			kecamatans.kode_wilayah, 
			kecamatans.lat, 
			kecamatans.long, 
			kecamatans.created_at, 
			kecamatans.updated_at,
			respondents.name as nama_pejabat, 
			pejabat__wilayahs.periode_awal, 
			pejabat__wilayahs.periode_akhir
		`).
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Where("kecamatans.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("kecamatans").
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Where("kecamatans.deleted_at IS NULL")

	if respondent != nil {
		if *respondent.RoleId != int64(enums.ROLE_WALIKOTA) && *respondent.RoleId != int64(enums.ROLE_ADMIN) && *respondent.RoleId != int64(enums.ROLE_KECAMATAN) && *respondent.RoleId != int64(enums.ROLE_SURVEYOR) {
			db = db.Where("1 = 0")
			countDB = countDB.Where("1 = 0")
		} else {
			if *respondent.RoleId == int64(enums.ROLE_KECAMATAN) {
				db = db.Where("kecamatans.id = ?", *respondent.KecamatanId)
				countDB = countDB.Where("kecamatans.id = ?", *respondent.KecamatanId)
			}
		}
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				respondents.name ILIKE ? OR 
				kecamatans.kode_wilayah ILIKE ? OR 
				kecamatans.lat ILIKE ? OR 
				kecamatans.long ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				respondents.name ILIKE ? OR 
				kecamatans.kode_wilayah ILIKE ? OR 
				kecamatans.lat ILIKE ? OR 
				kecamatans.long ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				respondents.name ILIKE ? OR
				kecamatans.kode_wilayah ILIKE ? OR 
				kecamatans.lat ILIKE ? OR 
				kecamatans.long ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("kecamatans.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":                "kecamatans.id",
		"sub_district_name": "kecamatans.sub_district_name",
		"sub_district_slug": "kecamatans.sub_district_slug",
		"kode_wilayah":      "kecamatans.kode_wilayah",
		"nama_pejabat":      "respondents.name",
		"periode_awal":      "pejabat__wilayahs.periode_awal",
		"periode_akhir":     "pejabat__wilayahs.periode_akhir",
		"lat":               "kecamatans.lat",
		"long":              "kecamatans.long",
		"created_at":        "kecamatans.created_at",
	}

	finalOrderBy := "kecamatans.id"
	finalOrderDir := "desc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
		finalOrderBy = mappedCol
	}

	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	}

	db = db.Order(finalOrderBy + " " + finalOrderDir)

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *manajemenWilayahRepo) GetKecamatanOptions(req payloads.KecamatanOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("kecamatans").
		Select(`
			kecamatans.id AS id, 
			kecamatans.sub_district_name AS label
		`).
		Where("kecamatans.deleted_at IS NULL")

	if len(req.IDs) > 0 {
		db = db.Where("kecamatans.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("kecamatans.sub_district_name ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("kecamatans.sub_district_name asc")

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

func (repository *manajemenWilayahRepo) IsKecamatanUsed(id int64) (bool, error) {
	defer utils.GeneralRecover()

	defer utils.GeneralRecover()

	db := repository.dbSlave
	var surveyWilayahCount int64
	err := db.Model(&models.SurveyWilayah{}).Where("kecamatan_id = ?", id).Count(&surveyWilayahCount).Error
	if err != nil {
		return false, err
	}

	if surveyWilayahCount > 0 {
		fmt.Println("Data digunakan Survey Wilayah")
		return true, nil
	}

	var pejabatWilayahCount int64
	err = db.Model(&models.PejabatWilayah{}).Where("id_wilayah = ? AND tipe_wilayah = 5", id).Count(&pejabatWilayahCount).Error
	if err != nil {
		return false, err
	}

	if pejabatWilayahCount > 0 {
		fmt.Println("Data digunakan Pejabat Wilayah")
		return true, nil
	}

	var respCount int64
	err = db.Model(&models.Respondent{}).Where("kecamatan_id = ?", id).Count(&respCount).Error
	if err != nil {
		return false, err
	}

	if respCount > 0 {
		fmt.Println("Data digunakan Respondent")
		return true, nil
	}

	var countKelurahan int64
	err = db.Model(&models.Kelurahan{}).Where("sub_district_id = ?", id).Count(&countKelurahan).Error
	if err != nil {
		return false, err
	}

	if countKelurahan > 0 {
		fmt.Println("Data digunakan kelurahan")
		return true, nil
	}

	return false, nil
}

func (repository *manajemenWilayahRepo) DeleteKecamatanById(id int64) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Delete(&models.Kecamatan{}, id).Error
	return err
}

func (repository *manajemenWilayahRepo) GetKelurahanByNameToLower(name string) (*models.Kelurahan, error) {
	defer utils.GeneralRecover()
	var data models.Kelurahan
	db := repository.dbSlave

	err := db.Where("LOWER(village_name) = ?", strings.ToLower(name)).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) CreateKelurahan(kelurahan models.Kelurahan) (*models.Kelurahan, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&kelurahan).Error
	if err != nil {
		return nil, err
	}

	return &kelurahan, nil
}

func (repository *manajemenWilayahRepo) GetKelurahanByID(id int) (*models.KelurahanDetail, error) {
	defer utils.GeneralRecover()
	var data models.KelurahanDetail
	db := repository.dbSlave

	err := db.Select("kelurahans.*, kecamatans.sub_district_name").
		Joins("JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").
		Where("kelurahans.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetKelurahanByName(name string) (*models.KelurahanDetail, error) {
	defer utils.GeneralRecover()
	var data models.KelurahanDetail
	db := repository.dbSlave

	err := db.Select("kelurahans.*, kecamatans.sub_district_name").
		Joins("JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").
		Where("kelurahans.village_name = ?", name).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetKelurahanBySlug(slug string) (*models.KelurahanDetail, error) {
	defer utils.GeneralRecover()
	var data models.KelurahanDetail
	db := repository.dbSlave

	err := db.Select("kelurahans.*, kecamatans.sub_district_name").
		Joins("JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").
		Where("kelurahans.village_name_slug = ?", slug).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) UpdateKelurahan(kelurahan models.Kelurahan) (*models.Kelurahan, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Save(&kelurahan).Error
	if err != nil {
		return nil, err
	}

	return &kelurahan, nil
}

func (repository *manajemenWilayahRepo) GetListKelurahan(req payloads.DatatablePayload, kecamatanId *int64, respondent *models.Respondent) ([]models.KelurahanDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.KelurahanDatatableResponse
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	pejabatJoinQuery := `LEFT JOIN (
		SELECT DISTINCT ON (id_wilayah) id_wilayah, id_responden, periode_awal, periode_akhir
		FROM pejabat__wilayahs
		WHERE tipe_wilayah = 4
		ORDER BY id_wilayah, periode_akhir DESC NULLS FIRST
	) pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = kelurahans.id`

	db := repository.dbSlave.Table("kelurahans").
		Select(`
			kelurahans.id, 
			kecamatans.sub_district_name,
			kecamatans.id as sub_district_id,
			kelurahans.kode_wilayah, 
			kelurahans.village_name,
			respondents.name as nama_pejabat, 
			pejabat__wilayahs.periode_awal,
			pejabat__wilayahs.periode_akhir,
			kelurahans.lat,
			kelurahans.long,
			kelurahans.created_at,
			kelurahans.updated_at
		`).
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id AND kecamatans.deleted_at IS NULL`).
		Where("kelurahans.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("kelurahans").
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id AND kecamatans.deleted_at IS NULL`).
		Where("kelurahans.deleted_at IS NULL")

	if respondent != nil && respondent.RoleId != nil {
		roleId := *respondent.RoleId

		switch roleId {
		case 5:
			if respondent.KecamatanId != nil {
				db = db.Where("kelurahans.sub_district_id = ?", *respondent.KecamatanId)
				countDB = countDB.Where("kelurahans.sub_district_id = ?", *respondent.KecamatanId)
			} else {
				db = db.Where("1 = 0")
				countDB = countDB.Where("1 = 0")
			}
		case 4:
			db = db.Where("kelurahans.id = ?", *respondent.KelurahanId)
			countDB = countDB.Where("kelurahans.id = ?", *respondent.KelurahanId)
		case 3, 2:
			db = db.Where("1 = 0")
			countDB = countDB.Where("1 = 0")
		default:
			if kecamatanId != nil {
				db = db.Where("kelurahans.sub_district_id = ?", *kecamatanId)
				countDB = countDB.Where("kelurahans.sub_district_id = ?", *kecamatanId)
			}
		}
	} else {
		db = db.Where("1 = 0")
		countDB = countDB.Where("1 = 0")
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				respondents.name ILIKE ? OR 
				kelurahans.village_postal_code ILIKE ? OR 
				kelurahans.kode_wilayah ILIKE ? OR 
				kelurahans.lat ILIKE ? OR 
				kelurahans.long ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR 
				respondents.name ILIKE ? OR 
				kelurahans.village_postal_code ILIKE ? OR 
				kelurahans.kode_wilayah ILIKE ? OR 
				kelurahans.lat ILIKE ? OR 
				kelurahans.long ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? 
				OR EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				respondents.name ILIKE ? OR
				kelurahans.village_postal_code ILIKE ? OR 
				kelurahans.kode_wilayah ILIKE ? OR 
				kelurahans.lat ILIKE ? OR 
				kelurahans.long ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("kelurahans.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":                "kelurahans.id",
		"kode_wilayah":      "kelurahans.kode_wilayah",
		"sub_district_name": "kecamatans.sub_district_name",
		"village_name":      "kelurahans.village_name",
		"nama_pejabat":      "respondents.name",
		"periode_awal":      "pejabat__wilayahs.periode_awal",
		"periode_akhir":     "pejabat__wilayahs.periode_akhir",
		"lat":               "kelurahans.lat",
		"long":              "kelurahans.long",
		"created_at":        "kecamatans.created_at",
	}

	finalOrderBy := "kelurahans.id"
	finalOrderDir := "desc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
		finalOrderBy = mappedCol
	}

	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	}

	db = db.Order(finalOrderBy + " " + finalOrderDir)

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *manajemenWilayahRepo) GetKelurahanOptions(req payloads.KelurahanOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("kelurahans").
		Select(`
            kelurahans.id AS id, 
            kelurahans.village_name AS label
        `).
		Where("kelurahans.deleted_at IS NULL")

	if len(req.KecamatanIds) > 0 {
		db = db.Where("kelurahans.sub_district_id IN ?", req.KecamatanIds)
	}

	if len(req.IDs) > 0 {
		db = db.Where("kelurahans.id IN ?", req.IDs)
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		db = db.Where("kelurahans.village_name ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("kelurahans.village_name asc")

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

func (repository *manajemenWilayahRepo) IsKelurahanUsed(id int64) (bool, error) {
	defer utils.GeneralRecover()

	db := repository.dbSlave
	var surveyWilayahCount int64
	err := db.Model(&models.SurveyWilayah{}).Where("kelurahan_id = ?", id).Count(&surveyWilayahCount).Error
	if err != nil {
		return false, err
	}

	if surveyWilayahCount > 0 {
		fmt.Println("Data digunakan Survey Wilayah")
		return true, nil
	}

	var pejabatWilayahCount int64
	err = db.Model(&models.PejabatWilayah{}).Where("id_wilayah = ? AND tipe_wilayah = 4", id).Count(&pejabatWilayahCount).Error
	if err != nil {
		return false, err
	}

	if pejabatWilayahCount > 0 {
		fmt.Println("Data digunakan Pejabat Wilayah")
		return true, nil
	}

	var respCount int64
	err = db.Model(&models.Respondent{}).Where("kelurahan_id = ?", id).Count(&respCount).Error
	if err != nil {
		return false, err
	}

	if respCount > 0 {
		fmt.Println("Data digunakan Respondent")
		return true, nil
	}

	var countRW int64
	err = db.Model(&models.DataRw{}).Where("kelurahan_id = ?", id).Count(&countRW).Error
	if err != nil {
		return false, err
	}

	if countRW > 0 {
		fmt.Println("Data digunakan RW")
		return true, nil
	}

	return false, nil
}

func (repository *manajemenWilayahRepo) DeleteKelurahanById(id int64) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Delete(&models.Kelurahan{}, id).Error
	return err
}

func (repository *manajemenWilayahRepo) GetRwByID(id int) (*models.DataRwDetail, error) {
	defer utils.GeneralRecover()
	var data models.DataRwDetail
	db := repository.dbSlave

	err := db.Select("data__rws.*, kelurahans.village_name, kelurahans.id as village_id, kecamatans.id as sub_district_id, kecamatans.sub_district_name").
		Joins("JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id").
		Where("data__rws.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetRwByName(name string, kelurahanId *int64) (*models.DataRwDetail, error) {
	defer utils.GeneralRecover()
	var data models.DataRwDetail
	db := repository.dbSlave

	err := db.Select("data__rws.*, kelurahans.village_name, kelurahans.id as village_id, kecamatans.id as sub_district_id, kecamatans.sub_district_name").
		Joins("JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id").
		Where("data__rws.nama_rw = ? AND data__rws.kelurahan_id = ?", name, kelurahanId).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) UpdateRw(rw models.DataRw) (*models.DataRw, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Save(&rw).Error
	if err != nil {
		return nil, err
	}

	return &rw, nil
}

func (repository *manajemenWilayahRepo) GetListRw(req payloads.DatatablePayload, kelurahanId *int64, respondent *models.Respondent) ([]models.RwDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.RwDatatableResponse
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	pejabatJoinQuery := `LEFT JOIN (
		SELECT DISTINCT ON (id_wilayah) id_wilayah, id_responden, periode_awal, periode_akhir
		FROM pejabat__wilayahs
		WHERE tipe_wilayah = 3
		ORDER BY id_wilayah, periode_akhir DESC NULLS FIRST
	) pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = data__rws.id`

	db := repository.dbSlave.Table("data__rws").
		Select(`
			data__rws.id,
			data__rws.kode_wilayah, 
			kecamatans.sub_district_name,
			kecamatans.id as sub_district_id,
			kelurahans.village_name,
			kelurahans.id as village_id,
			data__rws.nama_rw,
			respondents.name as nama_pejabat, 
			pejabat__wilayahs.periode_awal,
			pejabat__wilayahs.periode_akhir,
			data__rws.lat,
			data__rws.long,
			data__rws.created_at,
			data__rws.updated_at
		`).
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id AND kelurahans.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id AND kecamatans.deleted_at IS NULL`).
		Where("data__rws.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("data__rws").
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id AND kelurahans.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id AND kecamatans.deleted_at IS NULL`).
		Where("data__rws.deleted_at IS NULL")

	if respondent != nil && respondent.RoleId != nil {
		roleId := *respondent.RoleId

		switch roleId {
		case 5:
			if kelurahanId != nil {
				db = db.Where("data__rws.kelurahan_id = ?", *kelurahanId)
				countDB = countDB.Where("data__rws.kelurahan_id = ?", *kelurahanId)
			} else if respondent.KecamatanId != nil {
				db = db.Where("kecamatans.id = ?", *respondent.KecamatanId)
				countDB = countDB.Where("kecamatans.id = ?", *respondent.KecamatanId)
			} else {
				db = db.Where("1 = 0")
				countDB = countDB.Where("1 = 0")
			}
		case 4:
			// Lurah: selalu dibatasi ke kelurahannya sendiri.
			db = db.Where("kelurahans.id = ?", *respondent.KelurahanId)
			countDB = countDB.Where("kelurahans.id = ?", *respondent.KelurahanId)
		case 3:
			db = db.Where("data__rws.id = ?", *respondent.RWId)
			countDB = countDB.Where("data__rws.id = ?", *respondent.RWId)
		case 2:
			db = db.Where("1 = 0")
			countDB = countDB.Where("1 = 0")
		default:
			if kelurahanId != nil {
				db = db.Where("data__rws.kelurahan_id = ?", *kelurahanId)
				countDB = countDB.Where("data__rws.kelurahan_id = ?", *kelurahanId)
			}
		}
	} else {
		db = db.Where("1 = 0")
		countDB = countDB.Where("1 = 0")
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR 
				data__rws.kode_wilayah ILIKE ? OR 
				data__rws.lat ILIKE ? OR 
				data__rws.long ILIKE ? OR 
				respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR
				data__rws.kode_wilayah ILIKE ? OR 
				data__rws.lat ILIKE ? OR 
				data__rws.long ILIKE ? OR 
				respondents.name ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
			kecamatans.sub_district_name ILIKE ? OR 
			kelurahans.village_name ILIKE ? OR 
			data__rws.nama_rw ILIKE ? OR
			data__rws.kode_wilayah ILIKE ? OR 
			data__rws.lat ILIKE ? OR 
			data__rws.long ILIKE ? OR  
			respondents.name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("data__rws.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":             "data__rws.id",
		"kode_wilayah":   "kelurahans.kode_wilayah",
		"nama_kecamatan": "kecamatans.sub_district_name",
		"nama_kelurahan": "kelurahans.village_name",
		"nama_rw":        "data__rws.nama_rw",
		"nama_pejabat":   "respondents.name",
		"periode_awal":   "pejabat__wilayahs.periode_awal",
		"periode_akhir":  "pejabat__wilayahs.periode_akhir",
		"lat":            "kelurahans.lat",
		"long":           "kelurahans.long",
		"created_at":     "kecamatans.created_at",
	}

	finalOrderBy := "CAST(data__rws.nama_rw AS INTEGER)"
	finalOrderDir := "asc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
		finalOrderBy = mappedCol
	}

	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	} else if strings.ToLower(req.OrderDir) == "desc" {
		finalOrderDir = "desc"
	}

	db = db.Order(finalOrderBy + " " + finalOrderDir)

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *manajemenWilayahRepo) CreateRw(rw models.DataRw) (*models.DataRw, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&rw).Error
	if err != nil {
		return nil, err
	}

	return &rw, nil
}

func (repository *manajemenWilayahRepo) GetRwOptions(req payloads.RwOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	selectClause := `
		data__rws.id AS id, 
		data__rws.nama_rw AS label
	`

	if len(req.KelurahanIds) > 1 {
		selectClause = `
			data__rws.id AS id, 
			CONCAT(data__rws.nama_rw, ' - ', kelurahans.village_name) AS label
		`
	}

	db := repository.dbSlave.Table("data__rws").
		Select(selectClause).
		Where("data__rws.deleted_at IS NULL")

	if len(req.KelurahanIds) > 1 {
		db = db.Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id AND kelurahans.deleted_at IS NULL")
	}

	if len(req.KelurahanIds) > 0 {
		db = db.Where("data__rws.kelurahan_id IN ?", req.KelurahanIds)
	}

	if len(req.IDs) > 0 {
		db = db.Where("data__rws.id IN ?", req.IDs)
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		if len(req.KelurahanIds) > 1 {
			db = db.Where("data__rws.nama_rw ILIKE ? OR kelurahans.village_name ILIKE ?", searchTerm, searchTerm)
		} else {
			db = db.Where("data__rws.nama_rw ILIKE ?", searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("CAST(nama_rw AS INTEGER) ASC")

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

func (repository *manajemenWilayahRepo) IsRwUsed(id int64) (bool, error) {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var surveyWilayahCount int64
	err := db.Model(&models.SurveyWilayah{}).Where("rw_id = ?", id).Count(&surveyWilayahCount).Error
	if err != nil {
		return false, err
	}

	if surveyWilayahCount > 0 {
		fmt.Println("Data digunakan Survey Wilayah")
		return true, nil
	}

	var pejabatWilayahCount int64
	err = db.Model(&models.PejabatWilayah{}).Where("id_wilayah = ? AND tipe_wilayah = 3", id).Count(&pejabatWilayahCount).Error
	if err != nil {
		return false, err
	}

	if pejabatWilayahCount > 0 {
		fmt.Println("Data digunakan Pejabat Wilayah")
		return true, nil
	}

	var respCount int64
	err = db.Model(&models.Respondent{}).Where("rw_id = ?", id).Count(&respCount).Error
	if err != nil {
		return false, err
	}

	if respCount > 0 {
		fmt.Println("Data digunakan Respondent")
		return true, nil
	}

	var countRT int64
	err = db.Model(&models.DataRt{}).Where("rw_id = ?", id).Count(&countRT).Error
	if err != nil {
		return false, err
	}

	if countRT > 0 {
		fmt.Println("Data digunakan RT")
		return true, nil
	}

	return false, nil
}

func (repository *manajemenWilayahRepo) DeleteRwById(id int64) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Delete(&models.DataRw{}, id).Error
	return err
}

func (repository *manajemenWilayahRepo) GetRtByID(id int64) (*models.DataRtDetail, error) {
	defer utils.GeneralRecover()
	var data models.DataRtDetail
	db := repository.dbSlave

	err := db.Select(`
		data__rts.*, 
		data__rws.nama_rw,
		data__rws.id as rw_id,
		kelurahans.village_name, 
		kelurahans.id as village_id, 
		kecamatans.id as sub_district_id, 
		kecamatans.sub_district_name
	`).
		Joins("JOIN data__rws ON data__rws.id = data__rts.rw_id").
		Joins("JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id").
		Where("data__rts.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) GetRtByName(name string, rwId *int64) (*models.DataRtDetail, error) {
	defer utils.GeneralRecover()
	var data models.DataRtDetail
	db := repository.dbSlave

	err := db.Select(`
		data__rts.*, 
		data__rws.nama_rw,
		data__rws.id as rw_id,
		kelurahans.village_name, 
		kelurahans.id as village_id, 
		kecamatans.id as sub_district_id, 
		kecamatans.sub_district_name
	`).
		Joins("JOIN data__rws ON data__rws.id = data__rts.rw_id").
		Joins("JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id").
		Where("data__rts.nama_rt = ? AND data__rts.rw_id = ?", name, rwId).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenWilayahRepo) UpdateRt(rt models.DataRt) (*models.DataRt, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Save(&rt).Error
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

func (repository *manajemenWilayahRepo) GetListRt(req payloads.DatatablePayload, rwId *int64) ([]models.RtDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.RtDatatableResponse
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	pejabatJoinQuery := `LEFT JOIN (
		SELECT DISTINCT ON (id_wilayah) id_wilayah, id_responden, periode_awal, periode_akhir
		FROM pejabat__wilayahs
		WHERE tipe_wilayah = 2
		ORDER BY id_wilayah, periode_akhir DESC NULLS FIRST
	) pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = data__rts.id`

	db := repository.dbSlave.Table("data__rts").
		Select(`
			data__rts.id,
			data__rts.kode_wilayah, 
			kecamatans.sub_district_name,
			kecamatans.id as sub_district_id,
			kelurahans.village_name,
			kelurahans.id as village_id,
			data__rws.nama_rw,
			data__rws.id as rw_id,
			data__rts.nama_rt,
			respondents.name as nama_pejabat, 
			pejabat__wilayahs.periode_awal,
			pejabat__wilayahs.periode_akhir,
			data__rts.lat,
			data__rts.long,
			data__rts.created_at,
			data__rts.updated_at
		`).
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN data__rws ON data__rws.id = data__rts.rw_id AND data__rws.deleted_at IS NULL`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id AND kelurahans.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id AND kecamatans.deleted_at IS NULL`).
		Where("data__rts.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("data__rts").
		Joins(pejabatJoinQuery).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN data__rws ON data__rws.id = data__rts.rw_id AND data__rws.deleted_at IS NULL`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id AND kelurahans.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id AND kecamatans.deleted_at IS NULL`).
		Where("data__rts.deleted_at IS NULL")

	if rwId != nil {
		db = db.Where("data__rts.rw_id = ?", *rwId)
		countDB = countDB.Where("data__rts.rw_id = ?", *rwId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR 
				data__rts.nama_rt ILIKE ? OR
				data__rts.kode_wilayah ILIKE ? OR
				data__rts.lat ILIKE ? OR
				data__rts.long ILIKE ? OR
				respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
				data__rts.kode_wilayah ILIKE ? OR
				data__rts.lat ILIKE ? OR
				data__rts.long ILIKE ? OR
				respondents.name ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR 
				data__rts.nama_rt ILIKE ? OR
				data__rts.kode_wilayah ILIKE ? OR
				data__rts.lat ILIKE ? OR
				data__rts.long ILIKE ? OR
				respondents.name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("data__rts.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":             "data__rts.id",
		"kode_wilayah":   "kelurahans.kode_wilayah",
		"nama_kecamatan": "kecamatans.sub_district_name",
		"nama_kelurahan": "kelurahans.village_name",
		"nama_rw":        "data__rws.nama_rw",
		"nama_rt":        "data__rts.nama_rt",
		"nama_pejabat":   "respondents.name",
		"periode_awal":   "pejabat__wilayahs.periode_awal",
		"periode_akhir":  "pejabat__wilayahs.periode_akhir",
		"lat":            "kelurahans.lat",
		"long":           "kelurahans.long",
		"created_at":     "kecamatans.created_at",
	}

	finalOrderBy := "CAST(data__rts.nama_rt AS INTEGER)"
	finalOrderDir := "asc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
		finalOrderBy = mappedCol
	}

	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	} else if strings.ToLower(req.OrderDir) == "desc" {
		finalOrderDir = "desc"
	}

	db = db.Order(finalOrderBy + " " + finalOrderDir)

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *manajemenWilayahRepo) CreateRt(rt models.DataRt) (*models.DataRt, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&rt).Error
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

func (repository *manajemenWilayahRepo) GetRtOptions(req payloads.RtOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	selectClause := `
		data__rts.id AS id, 
		data__rts.nama_rt AS label
	`

	if len(req.RwIds) > 1 {
		selectClause = `
			data__rts.id AS id, 
			CONCAT(data__rts.nama_rt, ' - ', data__rws.nama_rw, ' - ', kelurahans.village_name) AS label
		`
	}

	db := repository.dbSlave.Table("data__rts").
		Select(selectClause).
		Where("data__rts.deleted_at IS NULL")

	if len(req.RwIds) > 1 {
		db = db.Joins("LEFT JOIN data__rws ON data__rts.rw_id = data__rws.id AND data__rws.deleted_at IS NULL").
			Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id AND kelurahans.deleted_at IS NULL")
	}

	if len(req.RwIds) > 0 {
		db = db.Where("data__rts.rw_id IN ?", req.RwIds)
	}

	if len(req.IDs) > 0 {
		db = db.Where("data__rts.id IN ?", req.IDs)
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		if len(req.RwIds) > 1 {
			db = db.Where("data__rts.nama_rt ILIKE ? OR data__rws.nama_rw ILIKE ? OR kelurahans.village_name ILIKE ?", searchTerm, searchTerm, searchTerm)
		} else {
			db = db.Where("data__rts.nama_rt ILIKE ?", searchTerm)
		}

	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("CAST(nama_rt AS INTEGER) ASC")

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

func (repository *manajemenWilayahRepo) IsRtUsed(id int64) (bool, error) {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var pejabatWilayahCount int64
	err := db.Model(&models.PejabatWilayah{}).Where("id_wilayah = ? AND tipe_wilayah = 2", id).Count(&pejabatWilayahCount).Error
	if err != nil {
		return false, err
	}

	if pejabatWilayahCount > 0 {
		fmt.Println("Data digunakan Pejabat Wilayah")
		return true, nil
	}

	var respCount int64
	err = db.Model(&models.Respondent{}).Where("rt_id = ?", id).Count(&respCount).Error
	if err != nil {
		return false, err
	}

	if respCount > 0 {
		fmt.Println("Data digunakan Respondent")
		return true, nil
	}

	return false, nil
}

func (repository *manajemenWilayahRepo) DeleteRtById(id int64) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Delete(&models.DataRt{}, id).Error
	return err
}

func (repository *manajemenWilayahRepo) GetListKecamatanV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultKecamatan, int64, error) {
	var results []models.ResultKecamatan
	var metaTotal int64

	query := repository.dbSlave.WithContext(ctx).Table("kecamatans").Select(`
		id, 
		sub_district_name,
		(SELECT COUNT(*) FROM kelurahans WHERE sub_district_id = kecamatans.id) AS total_kelurahan,
		(SELECT COUNT(*) FROM data__rws JOIN kelurahans k ON k.id = data__rws.kelurahan_id WHERE k.sub_district_id = kecamatans.id) AS total_rw,
		(SELECT COUNT(*) FROM data__rts JOIN data__rws rw ON rw.id = data__rts.rw_id JOIN kelurahans k ON k.id = rw.kelurahan_id WHERE k.sub_district_id = kecamatans.id) AS total_rt
	`).Where("deleted_at IS NULL")

	if payload.Search != "" {
		query = query.Where("sub_district_name ILIKE ?", "%"+payload.Search+"%")
	}

	query.Count(&metaTotal)
	err := query.Order(payload.OrderBy + " " + payload.OrderDir).Limit(payload.Limit).Offset(offset).Find(&results).Error

	return results, metaTotal, err
}

func (repository *manajemenWilayahRepo) GetListKelurahanV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultKelurahan, int64, error) {
	var results []models.ResultKelurahan
	var metaTotal int64

	query := repository.dbSlave.WithContext(ctx).Table("kelurahans").Select(`
		kelurahans.id, 
		kelurahans.village_name, 
		kecamatans.sub_district_name,
		(SELECT COUNT(*) FROM data__rws WHERE kelurahan_id = kelurahans.id) AS total_rw,
		(SELECT COUNT(*) FROM data__rts JOIN data__rws rw ON rw.id = data__rts.rw_id WHERE rw.kelurahan_id = kelurahans.id) AS total_rt
	`).
		Joins("LEFT JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").Where("kelurahans.deleted_at IS NULL AND kecamatans.deleted_at IS NULL")

	if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		query = query.Where("kelurahans.sub_district_id = ?", *payload.KecamatanId)
	}
	if payload.Search != "" {
		query = query.Where("kelurahans.village_name ILIKE ?", "%"+payload.Search+"%")
	}

	orderQuery := payload.OrderBy
	if payload.OrderBy == "id" {
		orderQuery = "kelurahans.id"
	}

	query.Count(&metaTotal)
	err := query.Order(orderQuery + " " + payload.OrderDir).Limit(payload.Limit).Offset(offset).Find(&results).Error

	return results, metaTotal, err
}

func (repository *manajemenWilayahRepo) GetListRWV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultRW, int64, error) {
	var results []models.ResultRW
	var metaTotal int64

	query := repository.dbSlave.WithContext(ctx).Table("data__rws").Select(`
		data__rws.id, 
		data__rws.nama_rw, 
		kelurahans.village_name, 
		kecamatans.sub_district_name,
		(SELECT COUNT(*) FROM data__rts WHERE rw_id = data__rws.id) AS total_rt
	`).
		Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("LEFT JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").
		Where("data__rws.deleted_at IS NULL AND kelurahans.deleted_at IS NULL AND kecamatans.deleted_at IS NULL")

	if payload.KelurahanId != nil && *payload.KelurahanId > 0 {
		query = query.Where("data__rws.kelurahan_id = ?", *payload.KelurahanId)
	} else if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		query = query.Where("kelurahans.sub_district_id = ?", *payload.KecamatanId)
	}

	if payload.Search != "" {
		query = query.Where("data__rws.nama_rw ILIKE ?", "%"+payload.Search+"%")
	}

	orderQuery := payload.OrderBy
	if payload.OrderBy == "id" {
		orderQuery = "data__rws.id"
	}

	query.Count(&metaTotal)
	err := query.Order(orderQuery + " " + payload.OrderDir).Limit(payload.Limit).Offset(offset).Find(&results).Error

	return results, metaTotal, err
}

func (repository *manajemenWilayahRepo) GetListRTV2(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload, offset int) ([]models.ResultRT, int64, error) {
	var results []models.ResultRT
	var metaTotal int64

	query := repository.dbSlave.WithContext(ctx).Table("data__rts").
		Select("data__rts.id, data__rts.nama_rt, data__rws.nama_rw, kelurahans.village_name, kecamatans.sub_district_name").
		Joins("LEFT JOIN data__rws ON data__rws.id = data__rts.rw_id").
		Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
		Joins("LEFT JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id").
		Where("data__rts.deleted_at IS NULL AND data__rws.deleted_at IS NULL AND kelurahans.deleted_at IS NULL AND kecamatans.deleted_at IS NULL")

	if payload.RWId != nil && *payload.RWId > 0 {
		query = query.Where("data__rts.rw_id = ?", *payload.RWId)
	} else if payload.KelurahanId != nil && *payload.KelurahanId > 0 {
		query = query.Where("data__rws.kelurahan_id = ?", *payload.KelurahanId)
	} else if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		query = query.Where("kelurahans.sub_district_id = ?", *payload.KecamatanId)
	}

	if payload.Search != "" {
		query = query.Where("data__rts.nama_rt ILIKE ?", "%"+payload.Search+"%")
	}

	orderQuery := payload.OrderBy
	if payload.OrderBy == "id" {
		orderQuery = "data__rts.id"
	}

	query.Count(&metaTotal)

	err := query.Order(orderQuery + " " + payload.OrderDir).
		Limit(payload.Limit).
		Offset(offset).
		Find(&results).Error

	return results, metaTotal, err
}

func (repository *manajemenWilayahRepo) GetWilayahSummary(ctx context.Context, payload payloads.DatatableDataWilayahKotaBandungPayload) (response.WilayahSummary, error) {
	var summary response.WilayahSummary

	qKec := repository.dbSlave.WithContext(ctx).Table("kecamatans")
	if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		qKec = qKec.Where("id = ?", *payload.KecamatanId)
	}
	qKec.Count(&summary.TotalKecamatan)

	qKel := repository.dbSlave.WithContext(ctx).Table("kelurahans")
	if payload.KelurahanId != nil && *payload.KelurahanId > 0 {
		qKel = qKel.Where("id = ?", *payload.KelurahanId)
	} else if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		qKel = qKel.Where("sub_district_id = ?", *payload.KecamatanId)
	}
	qKel.Count(&summary.TotalKelurahan)

	qRW := repository.dbSlave.WithContext(ctx).Table("data__rws")
	if payload.RWId != nil && *payload.RWId > 0 {
		qRW = qRW.Where("id = ?", *payload.RWId)
	} else if payload.KelurahanId != nil && *payload.KelurahanId > 0 {
		qRW = qRW.Where("kelurahan_id = ?", *payload.KelurahanId)
	} else if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		qRW = qRW.Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
			Where("kelurahans.sub_district_id = ?", *payload.KecamatanId)
	}
	qRW.Count(&summary.TotalRw)

	qRT := repository.dbSlave.WithContext(ctx).Table("data__rts")
	if payload.RWId != nil && *payload.RWId > 0 {
		qRT = qRT.Where("rw_id = ?", *payload.RWId)
	} else if payload.KelurahanId != nil && *payload.KelurahanId > 0 {
		qRT = qRT.Joins("LEFT JOIN data__rws ON data__rws.id = data__rts.rw_id").
			Where("data__rws.kelurahan_id = ?", *payload.KelurahanId)
	} else if payload.KecamatanId != nil && *payload.KecamatanId > 0 {
		qRT = qRT.Joins("LEFT JOIN data__rws ON data__rws.id = data__rts.rw_id").
			Joins("LEFT JOIN kelurahans ON kelurahans.id = data__rws.kelurahan_id").
			Where("kelurahans.sub_district_id = ?", *payload.KecamatanId)
	}
	qRT.Count(&summary.TotalRt)

	return summary, nil
}
