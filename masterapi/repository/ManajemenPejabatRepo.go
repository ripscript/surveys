package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"strings"

	"gorm.io/gorm"
)

type ManajemenPejabatRepo interface {
	GetFirstPejabatanWilayahByWilayahIdDanTipeWilayah(id int, tipeWilayah int) (*models.PejabatWilayah, error)
	GetRespondenById(id int64) (*models.Respondent, error)
	CreatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error)
	IsPejabatExistByRespondenId(respondenId int64, tipeWilayah int64) (bool, error)
	GetPejabatById(id int64) (*models.DetailPejabatWilayah, error)
	UpdatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error)
	DeletePejabatById(id int64) error
	GetListPejabat(req payloads.DatatablePejabatPayload, respondent *models.Respondent) ([]models.PejabatWilayahList, int64, error)
	IsRespondentHaveActivePejabat(respondenId int64) (bool, error)
}

type manajemenPejabatRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenPejabatRepo(dbSlave, dbMaster *gorm.DB) *manajemenPejabatRepo {
	defer utils.GeneralRecover()
	return &manajemenPejabatRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenPejabatRepo) GetFirstPejabatanWilayahByWilayahIdDanTipeWilayah(id int, tipeWilayah int) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	var data models.PejabatWilayah
	db := repository.dbSlave

	err := db.Preload("Respondent").Where("tipe_wilayah = ? AND id_wilayah = ?", tipeWilayah, id).Order("id DESC").First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenPejabatRepo) GetRespondenById(id int64) (*models.Respondent, error) {
	defer utils.GeneralRecover()

	var data models.Respondent
	db := repository.dbSlave

	err := db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenPejabatRepo) IsPejabatExistByRespondenId(respondenId int64, tipeWilayah int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.PejabatWilayah{}).Where("id_responden = ? AND tipe_wilayah = ? AND status_jabat = 1", respondenId, tipeWilayah).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *manajemenPejabatRepo) CreatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.PejabatWilayah{}).
			Where("id_responden = ? AND status_jabat = 1", pejabat.IdResponden).
			Update("status_jabat", 0).Error
		if err != nil {
			return err
		}

		return tx.Create(pejabat).Error
	})

	if err != nil {
		return nil, err
	}

	return pejabat, nil
}

func (repository *manajemenPejabatRepo) UpdatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Transaction(func(tx *gorm.DB) error {
		if pejabat.StatusJabat != nil && *pejabat.StatusJabat == 1 {
			err := tx.Model(&models.PejabatWilayah{}).
				Where("id_responden = ? AND status_jabat = 1", pejabat.IdResponden).
				Update("status_jabat", 0).Error
			if err != nil {
				return err
			}
		}

		return tx.Save(pejabat).Error
	})

	if err != nil {
		return nil, err
	}

	return pejabat, nil
}

func (repository *manajemenPejabatRepo) GetPejabatById(id int64) (*models.DetailPejabatWilayah, error) {
	defer utils.GeneralRecover()

	var data models.DetailPejabatWilayah
	db := repository.dbSlave

	err := db.Preload("Respondent").Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenPejabatRepo) DeletePejabatById(id int64) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Delete(&models.PejabatWilayah{}, id).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *manajemenPejabatRepo) GetListPejabat(req payloads.DatatablePejabatPayload, respondent *models.Respondent) ([]models.PejabatWilayahList, int64, error) {
	defer utils.GeneralRecover()
	var data []models.PejabatWilayahList
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	db := repository.dbSlave.Table("pejabat__wilayahs").
		Select(`
			pejabat__wilayahs.id AS id,
			CONCAT_WS(' - ', kecamatans.sub_district_name, kelurahans.village_name, data__rws.nama_rw, data__rts.nama_rt) AS daftar_wilayah,
			CASE pejabat__wilayahs.tipe_wilayah
				WHEN 5 THEN 'kecamatan'
				WHEN 4 THEN 'kelurahan'
				WHEN 3 THEN 'rw'
				WHEN 2 THEN 'rt'
				ELSE 'lainnya'
			END AS tipe_wilayah,
			respondents.name AS nama_pejabat,
			pejabat__wilayahs.periode_awal AS periode_awal,
			pejabat__wilayahs.periode_akhir AS periode_akhir,
			CASE pejabat__wilayahs.status_jabat
				WHEN 1 THEN true
				ELSE false
			END AS status_jabat,
			pejabat__wilayahs.created_at AS created_at,
			pejabat__wilayahs.updated_at AS updated_at
		`).
		Joins("LEFT JOIN respondents ON pejabat__wilayahs.id_responden = respondents.id").
		Joins("LEFT JOIN kecamatans ON respondents.kecamatan_id = kecamatans.id").
		Joins("LEFT JOIN kelurahans ON respondents.kelurahan_id = kelurahans.id").
		Joins("LEFT JOIN data__rws ON respondents.rw_id = data__rws.id").
		Joins("LEFT JOIN data__rts ON respondents.rt_id = data__rts.id")

	countDB := repository.dbSlave.Table("pejabat__wilayahs").
		Joins("LEFT JOIN respondents ON pejabat__wilayahs.id_responden = respondents.id").
		Joins("LEFT JOIN kecamatans ON respondents.kecamatan_id = kecamatans.id").
		Joins("LEFT JOIN kelurahans ON respondents.kelurahan_id = kelurahans.id").
		Joins("LEFT JOIN data__rws ON respondents.rw_id = data__rws.id").
		Joins("LEFT JOIN data__rts ON respondents.rt_id = data__rts.id")

	switch *respondent.RoleId {
	case 7:
	case 4:
		db = db.Where("data__rws.kelurahan_id = ?", *respondent.KelurahanId)
		countDB = countDB.Where("data__rws.kelurahan_id = ?", *respondent.KelurahanId)
	default:
		db = db.Where("1 = 0")
		countDB = countDB.Where("1 = 0")
	}

	if req.FNama != nil {
		db = db.Where("respondents.name ILIKE ?", "%"+*req.FNama+"%")
		countDB = countDB.Where("respondents.name ILIKE ?", "%"+*req.FNama+"%")
	}

	if req.FPeriodeAwal != nil && len(*req.FPeriodeAwal) >= 10 {
		periodeAwal := *req.FPeriodeAwal
		if len(periodeAwal) >= 10 {
			periodeAwal = periodeAwal[:10]
		}
		db = db.Where("DATE(pejabat__wilayahs.periode_awal) >= ?::DATE", periodeAwal)
		countDB = countDB.Where("DATE(pejabat__wilayahs.periode_awal) >= ?::DATE", periodeAwal)
	}

	if req.FPeriodeAkhir != nil && len(*req.FPeriodeAkhir) >= 10 {
		periodeAkhir := *req.FPeriodeAkhir
		if len(periodeAkhir) >= 10 {
			periodeAkhir = periodeAkhir[:10]
		}
		db = db.Where("DATE(pejabat__wilayahs.periode_akhir) <= ?::DATE", periodeAkhir)
		countDB = countDB.Where("DATE(pejabat__wilayahs.periode_akhir) <= ?::DATE", periodeAkhir)
	}

	if req.FStatus != nil && (*req.FStatus == 0 || *req.FStatus == 1) {
		db = db.Where("pejabat__wilayahs.status_jabat = ?", *req.FStatus)
		countDB = countDB.Where("pejabat__wilayahs.status_jabat = ?", *req.FStatus)
	}

	if req.FKecamatan != nil {
		db = db.Where("kecamatans.id = ?", *req.FKecamatan)
		countDB = countDB.Where("kecamatans.id = ?", *req.FKecamatan)
	}

	if req.FKelurahan != nil {
		db = db.Where("kelurahans.id = ?", *req.FKelurahan)
		countDB = countDB.Where("kelurahans.id = ?", *req.FKelurahan)
	}

	if req.FRw != nil {
		db = db.Where("data__rws.id = ?", *req.FRw)
		countDB = countDB.Where("data__rws.id = ?", *req.FRw)
	}

	if req.FRt != nil {
		db = db.Where("data__rts.id = ?", *req.FRt)
		countDB = countDB.Where("data__rts.id = ?", *req.FRt)
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
                respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ? OR 
                DATE(pejabat__wilayahs.created_at) = ? OR 
                DATE(pejabat__wilayahs.updated_at) = ?
            `
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
		} else if len(req.Search) == 4 {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
                respondents.name ILIKE ? OR  
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ? OR 
                EXTRACT(YEAR FROM pejabat__wilayahs.created_at)::TEXT = ? OR 
                EXTRACT(YEAR FROM pejabat__wilayahs.updated_at)::TEXT = ?
            `
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr, searchStr, searchStr)
		} else {
			condition := `
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
                respondents.name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("pejabat__wilayahs.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":             "pejabat__wilayahs.id",
		"nama_pejabat":   "respondents.name",
		"created_at":     "pejabat__wilayahs.created_at",
		"updated_at":     "pejabat__wilayahs.updated_at",
		"periode_awal":   "pejabat__wilayahs.periode_awal",
		"periode_akhir":  "pejabat__wilayahs.periode_akhir",
		"status_jabat":   "pejabat__wilayahs.status_jabat",
		"daftar_wilayah": "daftar_wilayah",
		"tipe_wilayah":   "tipe_wilayah",
	}

	finalOrderBy := "pejabat__wilayahs.id"
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

func (repository *manajemenPejabatRepo) IsRespondentHaveActivePejabat(respondenId int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.PejabatWilayah{}).Where("id_responden = ? AND status_jabat = 1", respondenId).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
