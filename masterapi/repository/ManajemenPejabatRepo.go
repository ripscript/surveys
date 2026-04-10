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
	IsPejabatExistByRespondenId(respondenId int64) (bool, error)
	GetPejabatById(id int64) (*models.DetailPejabatWilayah, error)
	UpdatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error)
	DeletePejabatById(id int64) error
	GetListPejabat(req payloads.DatatablePejabatPayload) ([]models.PejabatWilayahList, int64, error)
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

func (repository *manajemenPejabatRepo) IsPejabatExistByRespondenId(respondenId int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.PejabatWilayah{}).Where("id_responden = ?", respondenId).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *manajemenPejabatRepo) CreatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Create(pejabat).Error
	if err != nil {
		return nil, err
	}

	return pejabat, nil
}

func (repository *manajemenPejabatRepo) UpdatePejabat(pejabat *models.PejabatWilayah) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Save(pejabat).Error
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

func (repository *manajemenPejabatRepo) GetListPejabat(req payloads.DatatablePejabatPayload) ([]models.PejabatWilayahList, int64, error) {
	defer utils.GeneralRecover()
	var data []models.PejabatWilayahList
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}

	db := repository.dbSlave.Table("pejabat__wilayahs").
		Joins("LEFT JOIN respondents ON pejabat__wilayahs.id_responden = respondents.id").
		Joins("LEFT JOIN kecamatans ON respondents.kecamatan_id = kecamatans.id").
		Joins("LEFT JOIN kelurahans ON respondents.kelurahan_id = kelurahans.id").
		Joins("LEFT JOIN data__rws ON respondents.rw_id = data__rws.id").
		Joins("LEFT JOIN data__rts ON respondents.rt_id = data__rts.id").
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
        `)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
                respondents.name ILIKE ? OR 
				DATE(pejabat__wilayahs.periode_awal) = ? OR 
				DATE(pejabat__wilayahs.periode_akhir) = ? OR 
                DATE(pejabat__wilayahs.created_at) = ? OR 
                DATE(pejabat__wilayahs.updated_at) = ?
            `, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
		} else if len(req.Search) == 4 {
			db = db.Where(`
				kecamatans.sub_district_name ILIKE ? OR
				kelurahans.village_name ILIKE ? OR
				data__rws.nama_rw ILIKE ? OR
				data__rts.nama_rt ILIKE ? OR
                respondents.name ILIKE ? OR  
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_awal)::TEXT = ? OR 
				EXTRACT(YEAR FROM pejabat__wilayahs.periode_akhir)::TEXT = ? OR 
                EXTRACT(YEAR FROM pejabat__wilayahs.created_at)::TEXT = ? OR 
                EXTRACT(YEAR FROM pejabat__wilayahs.updated_at)::TEXT = ?
            `, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
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

	if req.FNama != nil {
		db = db.Where("respondents.name ILIKE ?", "%"+*req.FNama+"%")
	}

	if req.FPeriodeAwal != nil && len(*req.FPeriodeAwal) >= 10 {
		periodeAwal := *req.FPeriodeAwal
		if len(periodeAwal) >= 10 {
			periodeAwal = periodeAwal[:10]
		}
		db = db.Where("DATE(pejabat__wilayahs.periode_awal) >= ?::DATE", periodeAwal)
	}

	if req.FPeriodeAkhir != nil && len(*req.FPeriodeAkhir) >= 10 {
		periodeAkhir := *req.FPeriodeAkhir
		if len(periodeAkhir) >= 10 {
			periodeAkhir = periodeAkhir[:10]
		}
		db = db.Where("DATE(pejabat__wilayahs.periode_akhir) <= ?::DATE", periodeAkhir)
	}

	if req.FStatus != nil && (*req.FStatus == 0 || *req.FStatus == 1) {
		db = db.Where("pejabat__wilayahs.status_jabat = ?", *req.FStatus)
	}

	if req.FKecamatan != nil {
		db = db.Where("kecamatans.id = ?", *req.FKecamatan)
	}

	if req.FKelurahan != nil {
		db = db.Where("kelurahans.id = ?", *req.FKelurahan)
	}

	if req.FRw != nil {
		db = db.Where("data__rws.id = ?", *req.FRw)
	}

	if req.FRt != nil {
		db = db.Where("data__rts.id = ?", *req.FRt)
	}

	// HITUNG TOTAL DATA
	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	// 3. FITUR ORDER BY
	if req.OrderBy != "" {
		finalOrderBy := "pejabat__wilayahs.id"
		finalOrderDir := "desc"

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

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("pejabat__wilayahs.id desc")
	}

	// 4. FITUR PAGINATION
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}
