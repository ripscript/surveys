package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"strings"

	"gorm.io/gorm"
)

type ManajemenWilayahRepo interface {
	GetKecamatanByID(id int) (*models.Kecamatan, error)
	UpdateKecamatan(kecamatan models.Kecamatan) (models.Kecamatan, error)
	GetKecamatanByName(name string) (*models.Kecamatan, error)
	GetKecamatanBySlug(slug string) (*models.Kecamatan, error)
	GetListKecamatan(req payloads.DatatablePayload) ([]models.KecamatanDatatableResponse, int64, error)
	GetKecamatanOptions(req payloads.KecamatanOptionsPayload) ([]response.OptionItem, int64, error)

	GetKelurahanByID(id int) (*models.KelurahanDetail, error)
	GetKelurahanByName(name string) (*models.KelurahanDetail, error)
	GetKelurahanBySlug(slug string) (*models.KelurahanDetail, error)
	UpdateKelurahan(kelurahan models.Kelurahan) (*models.Kelurahan, error)
	GetListKelurahan(req payloads.DatatablePayload, kecamatanId *int64) ([]models.KelurahanDatatableResponse, int64, error)
	GetKelurahanOptions(req payloads.KelurahanOptionsPayload) ([]response.OptionItem, int64, error)

	GetRwByID(id int) (*models.DataRwDetail, error)
	GetRwByName(name string, kelurahanId *int64) (*models.DataRwDetail, error)
	UpdateRw(rw models.DataRw) (*models.DataRw, error)
	GetListRw(req payloads.DatatablePayload, kelurahanId *int64) ([]models.RwDatatableResponse, int64, error)
	CreateRw(rw models.DataRw) (*models.DataRw, error)
	GetRwOptions(req payloads.RwOptionsPayload) ([]response.OptionItem, int64, error)

	GetRtByID(id int) (*models.DataRtDetail, error)
	GetRtByName(name string, rwId *int64) (*models.DataRtDetail, error)
	UpdateRt(rw models.DataRt) (*models.DataRt, error)
	GetListRt(req payloads.DatatablePayload, rwId *int64) ([]models.RtDatatableResponse, int64, error)
	CreateRt(rt models.DataRt) (*models.DataRt, error)
	GetRtOptions(req payloads.RtOptionsPayload) ([]response.OptionItem, int64, error)
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

func (repository *manajemenWilayahRepo) GetListKecamatan(req payloads.DatatablePayload) ([]models.KecamatanDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.KecamatanDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("kecamatans").
		Select(`
			kecamatans.id, kecamatans.sub_district_name, kecamatans.sub_district_slug, 
			kecamatans.kode_wilayah, kecamatans.lat, kecamatans.long, kecamatans.created_at, kecamatans.updated_at,
			respondents.name as nama_pejabat, 
			pejabat__wilayahs.periode_awal, pejabat__wilayahs.periode_akhir
		`).
		Joins(`LEFT JOIN pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = kecamatans.id AND pejabat__wilayahs.tipe_wilayah = 5`).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR respondents.name ILIKE ? 
				OR DATE(pejabat__wilayahs.periode_awal) = ? OR DATE(pejabat__wilayahs.periode_akhir) = ?
			`, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR respondents.name ILIKE ? 
				OR EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? 
				OR EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where("kecamatans.sub_district_name ILIKE ? OR respondents.name ILIKE ?", searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "kecamatans.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
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

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("kecamatans.id desc")
	}

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
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
        `)

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

func (repository *manajemenWilayahRepo) GetListKelurahan(req payloads.DatatablePayload, kecamatanId *int64) ([]models.KelurahanDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.KelurahanDatatableResponse
	var totalData int64

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
		Joins(`LEFT JOIN pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = kelurahans.id AND pejabat__wilayahs.tipe_wilayah = 4`).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kecamatans ON kecamatans.id = kelurahans.sub_district_id`)

	if kecamatanId != nil {
		db = db.Where("kelurahans.sub_district_id = ?", kecamatanId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR kelurahans.village_name ILIKE ? OR respondents.name ILIKE ? 
				OR DATE(pejabat__wilayahs.periode_awal) = ? OR DATE(pejabat__wilayahs.periode_akhir) = ?
			`, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR respondents.name ILIKE ? 
				OR EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? 
				OR EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where("kecamatans.sub_district_name ILIKE ? OR kelurahans.village_name ILIKE ? OR respondents.name ILIKE ?", searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "kelurahans.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
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

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("kelurahans.id desc")
	}

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
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
        `)

	if req.KecamatanId != 0 {
		db = db.Where("kelurahans.sub_district_id = ?", req.KecamatanId)
	}

	if len(req.IDs) > 0 {
		db = db.Where("kelurahans.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
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

func (repository *manajemenWilayahRepo) GetListRw(req payloads.DatatablePayload, kelurahanId *int64) ([]models.RwDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.RwDatatableResponse
	var totalData int64

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
		Joins(`LEFT JOIN pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = data__rws.id AND pejabat__wilayahs.tipe_wilayah = 3`).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id`)

	if kelurahanId != nil {
		db = db.Where("data__rws.kelurahan_id = ?", kelurahanId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR 
				respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR
				respondents.name ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
			kecamatans.sub_district_name ILIKE ? OR 
			kelurahans.village_name ILIKE ? OR 
			data__rws.nama_rw ILIKE ? OR 
			respondents.name ILIKE ?
			`, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "data__rws.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
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

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("data__rws.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
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

	db := repository.dbSlave.Table("data__rws").
		Select(`
            data__rws.id AS id, 
            data__rws.nama_rw AS label
        `)

	if req.KelurahanId != 0 {
		db = db.Where("data__rws.kelurahan_id = ?", req.KelurahanId)
	}

	if len(req.IDs) > 0 {
		db = db.Where("data__rws.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		db = db.Where("data__rws.nama_rw ILIKE ?", searchTerm)
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

func (repository *manajemenWilayahRepo) GetRtByID(id int) (*models.DataRtDetail, error) {
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
		Joins(`LEFT JOIN pejabat__wilayahs ON pejabat__wilayahs.id_wilayah = data__rts.id AND pejabat__wilayahs.tipe_wilayah = 2`).
		Joins(`LEFT JOIN respondents ON respondents.id = pejabat__wilayahs.id_responden AND respondents.deleted_at IS NULL`).
		Joins(`JOIN data__rws ON data__rws.id = data__rts.rw_id`).
		Joins(`JOIN kelurahans ON data__rws.kelurahan_id = kelurahans.id`).
		Joins(`JOIN kecamatans ON kelurahans.sub_district_id = kecamatans.id`)

	if rwId != nil {
		db = db.Where("data__rts.rw_id = ?", rwId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR 
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR 
				data__rts.nama_rt ILIKE ? OR
				respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR 
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
				respondents.name ILIKE ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
			kecamatans.sub_district_name ILIKE ? OR 
			kelurahans.village_name ILIKE ? OR 
			data__rws.nama_rw ILIKE ? OR 
			data__rts.nama_rt ILIKE ? OR
			respondents.name ILIKE ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "data__rts.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
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

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("data__rts.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
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

	db := repository.dbSlave.Table("data__rts").
		Select(`
            data__rts.id AS id, 
            data__rts.nama_rt AS label
        `)

	if req.RwId != 0 {
		db = db.Where("data__rts.rw_id = ?", req.RwId)
	}

	if len(req.IDs) > 0 {
		db = db.Where("data__rts.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		db = db.Where("data__rts.nama_rt ILIKE ?", searchTerm)
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
