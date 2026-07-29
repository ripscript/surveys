package repository

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/request"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type StatistikRepo interface {
	GetListAvailableSurvey(userLogin models.JwtCustomClaims, respondentLogin *models.RespondentModel_1, req request.SurveyWilayahDatatablePayload) ([]models.SurveyWilayahDatatableResponseV1, int64, error)
	GetStatistikKewilayahan(req request.StatistikKewilayahanPayload) ([]response.StatistikKewilayahanDatatableResponse, int64, *response.StatistikKewilayahanAggregate, error)
	GetWilayahAncestry(typeWilayah int, id int64) (int64, int64, int64, error)
	GetExportStatistikExcel(surveyId int, typeWilayah int, parentId int64) ([]response.ExportStatistikRow, error)
}

func NewStatistikRepo(dbSlave, dbMaster *gorm.DB) *statistikRepo {
	defer utils.GeneralRecover()
	return &statistikRepo{dbSlave, dbMaster}
}

type statistikRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func (repository *statistikRepo) GetListAvailableSurvey(userLogin models.JwtCustomClaims, respondentLogin *models.RespondentModel_1, req request.SurveyWilayahDatatablePayload) ([]models.SurveyWilayahDatatableResponseV1, int64, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyWilayahDatatableResponseV1
	var totalData int64

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN users ON users.id = surveys.created_by").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id").
		Where("surveys.approval_survey = 'approved' OR surveys.approval_survey = 'non_approval'")

	if req.StatusSurvey == string(enums.STATUS_SURVEY_UPCOMING) {
		db = db.Where("surveys.start_date > NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_ONGOING) {
		db = db.Where("surveys.start_date <= NOW() AND surveys.end_date >= NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_FINISHED) {
		db = db.Where("surveys.end_date < NOW()")
	}

	if respondentLogin.RoleId != nil {
		if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) {

			var kecId, kelId, rwId int64
			if respondentLogin.KecamatanId != nil {
				kecId = *respondentLogin.KecamatanId
			}
			if respondentLogin.KelurahanId != nil {
				kelId = *respondentLogin.KelurahanId
			}
			if respondentLogin.RWId != nil {
				rwId = *respondentLogin.RWId
			}

			db = db.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id 
				AND survey_wilayahs.kecamatan_id = ?
				AND (survey_wilayahs.kelurahan_id IS NULL OR survey_wilayahs.kelurahan_id = ?)
				AND (survey_wilayahs.rw_id IS NULL OR survey_wilayahs.rw_id = ?)
			) 
			OR NOT EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id
			)`, kecId, kelId, rwId)
		}
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ? OR
				DATE(surveys.start_date) = ? OR
				DATE(surveys.end_date) = ? OR
				DATE(flow_details.created_at) = ? OR
				DATE(flow_details.updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM surveys.start_date)::TEXT = ? OR
				EXTRACT(YEAR FROM surveys.end_date)::TEXT = ? OR
				EXTRACT(YEAR FROM flow_details.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM flow_details.updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchStr, searchStr, searchStr, searchStr)
		} else {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ?
			`, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		surveys.id,
		surveys.name AS survey_name,
		surveys.start_date,
		surveys.end_date,
		surveys.created_at,
		surveys.updated_at,
		surveys.created_by,
		users.first_name AS created_by_name,
		surveys.status,
		surveys.approval_survey
	`)

	if req.OrderBy != "" {
		finalOrderBy := "surveys.created_at"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":              "surveys.id",
			"survey_name":     "surveys.name",
			"start_date":      "surveys.start_date",
			"end_date":        "surveys.end_date",
			"created_at":      "surveys.created_at",
			"updated_at":      "surveys.updated_at",
			"created_by":      "surveys.created_by",
			"created_by_name": "users.first_name",
			"status":          "surveys.status",
			"approval_survey": "surveys.approval_survey",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("surveys.created_at desc")
	}

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

func (repository *statistikRepo) GetStatistikKewilayahan(req request.StatistikKewilayahanPayload) ([]response.StatistikKewilayahanDatatableResponse, int64, *response.StatistikKewilayahanAggregate, error) {
	defer utils.GeneralRecover()

	var data []response.StatistikKewilayahanDatatableResponse
	var totalData int64

	var baseQuery string
	baseArgs := []interface{}{}

	switch req.TypeWilayah {
	case 0:
		baseQuery = `
			SELECT
				k.id AS id,
				k.sub_district_name AS wilayah,
				(SELECT COUNT(*) FROM data__rts rt
					JOIN data__rws rw ON rw.id = rt.rw_id
					JOIN kelurahans kel ON kel.id = rw.kelurahan_id
					WHERE kel.sub_district_id = k.id) AS total_rt,
				(SELECT COUNT(DISTINCT r.rt_id) FROM respondents r
					JOIN survey_respondents sr ON sr.respondent_id = r.id
					WHERE sr.survey_id = ? AND sr.status = 2 AND r.kecamatan_id = k.id) AS total_sudah_mengisi
			FROM kecamatans k
			WHERE k.deleted_at IS NULL
		`
		baseArgs = []interface{}{req.SurveyId}

	case int(enums.ROLE_KECAMATAN):
		baseQuery = `
			SELECT
				kel.id AS id,
				kel.village_name AS wilayah,
				(SELECT COUNT(*) FROM data__rts rt
					JOIN data__rws rw ON rw.id = rt.rw_id
					WHERE rw.kelurahan_id = kel.id) AS total_rt,
				(SELECT COUNT(DISTINCT r.rt_id) FROM respondents r
					JOIN survey_respondents sr ON sr.respondent_id = r.id
					WHERE sr.survey_id = ? AND sr.status = 2 AND r.kelurahan_id = kel.id) AS total_sudah_mengisi
			FROM kelurahans kel
			WHERE kel.sub_district_id = ? AND kel.deleted_at IS NULL
		`
		baseArgs = []interface{}{req.SurveyId, req.ParentId}

	case int(enums.ROLE_KELURAHAN):
		baseQuery = `
			SELECT
				rw.id AS id,
				rw.nama_rw AS wilayah,
				(SELECT COUNT(*) FROM data__rts rt WHERE rt.rw_id = rw.id) AS total_rt,
				(SELECT COUNT(DISTINCT r.rt_id) FROM respondents r
					JOIN survey_respondents sr ON sr.respondent_id = r.id
					WHERE sr.survey_id = ? AND sr.status = 2 AND r.rw_id = rw.id) AS total_sudah_mengisi
			FROM data__rws rw
			WHERE rw.kelurahan_id = ? AND rw.deleted_at IS NULL
		`
		baseArgs = []interface{}{req.SurveyId, req.ParentId}

	case int(enums.ROLE_RW):
		baseQuery = `
			SELECT
				rt.id AS id,
				rt.nama_rt AS wilayah,
				1 AS total_rt,
				LEAST(
					(SELECT COUNT(DISTINCT r.id) FROM respondents r
						JOIN survey_respondents sr ON sr.respondent_id = r.id
						WHERE sr.survey_id = ? AND sr.status = 2 AND r.rt_id = rt.id),
					1
				) AS total_sudah_mengisi
			FROM data__rts rt
			WHERE rt.rw_id = ? AND rt.deleted_at IS NULL
		`
		baseArgs = []interface{}{req.SurveyId, req.ParentId}

	default:
		return nil, 0, nil, errors.New("type_wilayah tidak valid")
	}

	cteQuery := "WITH base AS (" + baseQuery + ")"

	whereClause := ""
	whereArgs := []interface{}{}
	if req.Search != "" {
		whereClause = " WHERE base.wilayah ILIKE ?"
		whereArgs = append(whereArgs, "%"+req.Search+"%")
	}

	countQuery := cteQuery + " SELECT COUNT(*) FROM base" + whereClause
	countArgs := append(append([]interface{}{}, baseArgs...), whereArgs...)

	err := repository.dbSlave.Raw(countQuery, countArgs...).Scan(&totalData).Error
	if err != nil {
		return nil, 0, nil, err
	}

	type aggregateResult struct {
		TotalWilayah      int64
		TotalSudahMengisi int64
		TotalRTSum        int64
	}
	var agg aggregateResult

	aggQuery := cteQuery + `
		SELECT
			COUNT(*) AS total_wilayah,
			COALESCE(SUM(base.total_sudah_mengisi), 0) AS total_sudah_mengisi,
			COALESCE(SUM(base.total_rt), 0) AS total_rt_sum
		FROM base` + whereClause

	aggArgs := append(append([]interface{}{}, baseArgs...), whereArgs...)

	err = repository.dbSlave.Raw(aggQuery, aggArgs...).Scan(&agg).Error
	if err != nil {
		return nil, 0, nil, err
	}

	aggregate := &response.StatistikKewilayahanAggregate{
		TotalWilayah:      agg.TotalWilayah,
		TotalSudahMengisi: agg.TotalSudahMengisi,
		TotalBelumMengisi: agg.TotalRTSum - agg.TotalSudahMengisi,
	}

	allowedOrderCols := map[string]string{
		"id":                  "base.id",
		"wilayah":             "base.wilayah",
		"total_rt":            "base.total_rt",
		"total_sudah_mengisi": "base.total_sudah_mengisi",
	}
	finalOrderBy := "base.wilayah"
	finalOrderDir := "asc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
		finalOrderBy = mappedCol
	}
	if strings.ToLower(req.OrderDir) == "desc" {
		finalOrderDir = "desc"
	}

	offset := (req.Page - 1) * req.Limit

	dataQuery := cteQuery + " SELECT base.* FROM base" + whereClause +
		" ORDER BY " + finalOrderBy + " " + finalOrderDir +
		" LIMIT ? OFFSET ?"

	dataArgs := append(append([]interface{}{}, baseArgs...), whereArgs...)
	dataArgs = append(dataArgs, req.Limit, offset)

	err = repository.dbSlave.Raw(dataQuery, dataArgs...).Scan(&data).Error
	if err != nil {
		return nil, 0, nil, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
		data[i].TotalBelumMengisi = data[i].TotalRT - data[i].TotalSudahMengisi
	}

	return data, totalData, aggregate, nil
}

func (repository *statistikRepo) GetWilayahAncestry(typeWilayah int, id int64) (int64, int64, int64, error) {
	defer utils.GeneralRecover()

	var row struct {
		KecamatanId int64
		KelurahanId int64
		RwId        int64
	}

	var query string

	switch typeWilayah {
	case int(enums.ROLE_KECAMATAN):
		return id, 0, 0, nil

	case int(enums.ROLE_KELURAHAN):
		query = `SELECT sub_district_id AS kecamatan_id, id AS kelurahan_id, 0 AS rw_id FROM kelurahans WHERE id = ? AND deleted_at IS NULL`

	case int(enums.ROLE_RW):
		query = `
			SELECT kel.sub_district_id AS kecamatan_id, rw.kelurahan_id AS kelurahan_id, rw.id AS rw_id
			FROM data__rws rw
			JOIN kelurahans kel ON kel.id = rw.kelurahan_id
			WHERE rw.id = ? AND rw.deleted_at IS NULL
		`

	default:
		return 0, 0, 0, errors.New("type_wilayah tidak valid untuk validasi ancestry")
	}

	err := repository.dbSlave.Raw(query, id).Scan(&row).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if row.KecamatanId == 0 && row.KelurahanId == 0 && row.RwId == 0 {
		return 0, 0, 0, errors.New("wilayah tidak ditemukan")
	}

	return row.KecamatanId, row.KelurahanId, row.RwId, nil
}

func (repository *statistikRepo) GetExportStatistikExcel(surveyId int, typeWilayah int, parentId int64) ([]response.ExportStatistikRow, error) {
	defer utils.GeneralRecover()

	query := `
		SELECT
			kec.sub_district_name AS kecamatan,
			kel.village_name AS kelurahan,
			rw.nama_rw AS rw,
			(SELECT COUNT(*) FROM data__rts rt WHERE rt.rw_id = rw.id AND rt.deleted_at IS NULL) AS jumlah_rt,
			(
				SELECT COUNT(DISTINCT r.rt_id) 
				FROM respondents r
				JOIN survey_respondents sr ON sr.respondent_id = r.id
				WHERE sr.survey_id = ? AND sr.status = 2 AND r.rw_id = rw.id
			) AS rt_sudah_mengisi
		FROM data__rws rw
		JOIN kelurahans kel ON rw.kelurahan_id = kel.id
		JOIN kecamatans kec ON kel.sub_district_id = kec.id
		WHERE rw.deleted_at IS NULL AND kel.deleted_at IS NULL AND kec.deleted_at IS NULL
	`
	args := []interface{}{surveyId}

	// Filter berdasarkan jenjang kewilayahan yang di-request / hak akses user
	switch typeWilayah {
	case int(enums.ROLE_KECAMATAN):
		query += " AND kec.id = ?"
		args = append(args, parentId)
	case int(enums.ROLE_KELURAHAN):
		query += " AND kel.id = ?"
		args = append(args, parentId)
	case int(enums.ROLE_RW):
		query += " AND rw.id = ?"
		args = append(args, parentId)
	}

	query += " ORDER BY kec.sub_district_name ASC, kel.village_name ASC, rw.nama_rw ASC"

	var data []response.ExportStatistikRow
	err := repository.dbSlave.Raw(query, args...).Scan(&data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}
