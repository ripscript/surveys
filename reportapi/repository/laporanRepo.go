package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/request"
	"backend/reportapi/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func parseJSONInt64Array(raw []byte) ([]int64, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var asInts []int64
	if err := json.Unmarshal(raw, &asInts); err == nil {
		return asInts, nil
	}

	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err != nil {
		return nil, err
	}

	result := make([]int64, 0, len(asStrings))
	for _, s := range asStrings {
		var v int64
		if _, err := fmt.Sscanf(s, "%d", &v); err == nil {
			result = append(result, v)
		}
	}
	return result, nil
}

type TableDataRow struct {
	TerritoryName string
	Values        map[int]int64 // Key: FormFieldID -> Value: Angka/Jumlah Jawaban
}

type LaporanRepo interface {
	UpdateReport(laporan models.Report) (*models.Report, error)
	GetReportByID(laporanID int64) (*models.Report, error)
	GetReportCoverByReportId(laporanID int64) (*models.ReportCover, error)
	ReplaceReportCover(laporanID int64, konten models.ReportCover) error

	DeleteReportKontenByReportID(laporanID int64) error
	UpdateReportKonten(konten models.ReportCover) error

	GetListReport(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error)

	WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error
	IsNameExists(ctx context.Context, name string) (bool, error)
	CreateReport(ctx context.Context, laporan models.Report) (*models.Report, error)
	CreateReportCover(ctx context.Context, konten models.ReportCover) (*models.ReportCover, error)

	// --- FUNGSI BARU UNTUK BUILDER & RENDER LAPORAN ---
	CheckSectionTitleExists(reportID int64, title string) (bool, error)
	CheckSubSectionTitleExists(reportID int64, title string) (bool, error)
	GetMaxSectionSequence(reportID int64) (int, error)
	CreateSectionTx(section *models.ReportSection) error
	GetFormFieldLabels(fieldIDs []int) (map[int]string, error)
	CalculateNarrativeVariable(report *models.Report, v request.NarrativeVariablePayload) (string, error)
	GetTableResponseData(report *models.Report, fieldIDs []int) ([]TableDataRow, error)
}

type laporanRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewLaporanRepo(dbSlave, dbMaster *gorm.DB) LaporanRepo {
	defer utils.GeneralRecover()
	return &laporanRepo{
		dbSlave:  dbSlave,
		dbMaster: dbMaster,
	}
}

func (r *laporanRepo) UpdateReport(laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.Save(&laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

// GetReportByID menarik laporan lengkap dengan Preload Sections, SubSections, & Components terurut ASC
func (r *laporanRepo) GetReportByID(laporanID int64) (*models.Report, error) {
	defer utils.GeneralRecover()
	var report models.Report
	err := r.dbSlave.
		Preload("Cover").
		Preload("Sections", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_sections.sequence ASC")
		}).
		Preload("Sections.SubSections", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_subsections.sequence ASC")
		}).
		Preload("Sections.SubSections.Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Preload("Sections.Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Where("id = ?", laporanID).
		First(&report).Error

	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *laporanRepo) GetReportCoverByReportId(laporanID int64) (*models.ReportCover, error) {
	defer utils.GeneralRecover()
	var cover models.ReportCover
	err := r.dbSlave.Where("report_id = ?", laporanID).First(&cover).Error
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (r *laporanRepo) ReplaceReportCover(laporanID int64, konten models.ReportCover) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Where("report_id = ?", laporanID).Delete(&models.ReportCover{}).Create(&konten).Error
}

func (r *laporanRepo) DeleteReportKontenByReportID(laporanID int64) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Where("report_id = ?", laporanID).Delete(&models.ReportCover{}).Error
}

func (r *laporanRepo) UpdateReportKonten(konten models.ReportCover) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Save(&konten).Error
}

func (r *laporanRepo) GetListReport(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LaporanDatatable
	var totalData int64

	db := r.dbSlave.Model(&models.Report{}).Where("respondent_id = ?", respondentId)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				reports.name ILIKE ? OR
				DATE(reports.created_at) = ? OR
				DATE(reports.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				reports.name ILIKE ? OR
				EXTRACT(YEAR FROM reports.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM reports.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				reports.name ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		reports.id,
		reports.name,
		reports.updated_at,
		reports.created_at
	`)

	if req.OrderBy != "" {
		finalOrderBy := "reports.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "reports.id",
			"name":       "reports.name",
			"created_at": "reports.created_at",
			"updated_at": "reports.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(fmt.Sprintf("%s %s", finalOrderBy, finalOrderDir))
	} else {
		db = db.Order("reports.id desc")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err = db.Limit(limit).Offset(offset).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (r *laporanRepo) WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error {
	defer utils.GeneralRecover()
	return r.dbMaster.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &laporanRepo{
			dbSlave:  tx,
			dbMaster: tx,
		}
		return fn(txRepo)
	})
}

func (r *laporanRepo) IsNameExists(ctx context.Context, name string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.WithContext(ctx).Model(&models.Report{}).
		Where("LOWER(name) = LOWER(?)", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *laporanRepo) CreateReport(ctx context.Context, laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.WithContext(ctx).Create(&laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func (r *laporanRepo) CreateReportCover(ctx context.Context, konten models.ReportCover) (*models.ReportCover, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.WithContext(ctx).Create(&konten).Error
	if err != nil {
		return nil, err
	}
	return &konten, nil
}

// CheckSectionTitleExists mengecek ketersediaan Judul Section di laporan tertentu
func (r *laporanRepo) CheckSectionTitleExists(reportID int64, title string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Model(&models.ReportSection{}).
		Where("report_id = ? AND LOWER(title) = LOWER(?)", reportID, title).
		Count(&count).Error

	return count > 0, err
}

// CheckSubSectionTitleExists mengecek ketersediaan Judul SubSection di laporan tertentu
func (r *laporanRepo) CheckSubSectionTitleExists(reportID int64, title string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Table("report_subsections").
		Joins("JOIN report_sections ON report_sections.id = report_subsections.report_section_id").
		Where("report_sections.report_id = ? AND LOWER(report_subsections.title) = LOWER(?)", reportID, title).
		Count(&count).Error

	return count > 0, err
}

// GetMaxSectionSequence mengambil urutan section tertinggi
func (r *laporanRepo) GetMaxSectionSequence(reportID int64) (int, error) {
	defer utils.GeneralRecover()
	var maxSeq int
	err := r.dbSlave.Model(&models.ReportSection{}).
		Where("report_id = ?", reportID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&maxSeq).Error

	if err != nil {
		return 0, err
	}

	return maxSeq, nil
}

// CreateSectionTx menyimpan section berserta nested anak-anaknya ke DB
func (r *laporanRepo) CreateSectionTx(section *models.ReportSection) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Create(section).Error
}

// GetFormFieldLabels mengambil nama soal dari form_fields
func (r *laporanRepo) GetFormFieldLabels(fieldIDs []int) (map[int]string, error) {
	defer utils.GeneralRecover()
	result := make(map[int]string)
	if len(fieldIDs) == 0 {
		return result, nil
	}

	type FieldLabel struct {
		ID       int    `gorm:"column:id"`
		Question string `gorm:"column:question"` // Ubah dari Title menjadi Question
	}
	var fields []FieldLabel
	err := r.dbSlave.Table("form_fields").
		Select("id, question"). // Ambil kolom question
		Where("id IN ?", fieldIDs).
		Scan(&fields).Error

	if err != nil {
		return result, nil
	}

	for _, f := range fields {
		result[f.ID] = f.Question
	}

	return result, nil
}

func (r *laporanRepo) CalculateNarrativeVariable(report *models.Report, v request.NarrativeVariablePayload) (string, error) {
	defer utils.GeneralRecover()

	var surveyIDs, kecIDs, kelIDs []int64
	_ = json.Unmarshal(report.SurveyID, &surveyIDs)
	_ = json.Unmarshal(report.KecamatanID, &kecIDs)
	_ = json.Unmarshal(report.KelurahanID, &kelIDs)

	groupByCol := "kelurahans.village_name"
	if report.TingkatWilayah == "6" {
		groupByCol = "kecamatans.sub_district_name"
	} else if report.TingkatWilayah == "4" {
		groupByCol = "data__rws.nama_rw"
	}

	baseQuery := r.dbSlave.Table("field_responses").
		Joins("JOIN survey_respondents ON survey_respondents.id = field_responses.form_response_id").
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Joins("LEFT JOIN kelurahans ON kelurahans.id = respondents.kelurahan_id AND kelurahans.deleted_at IS NULL").
		Joins("LEFT JOIN kecamatans ON kecamatans.id = respondents.kecamatan_id AND kecamatans.deleted_at IS NULL").
		Joins("LEFT JOIN data__rws ON data__rws.id = respondents.rw_id AND data__rws.deleted_at IS NULL").
		Where("field_responses.form_field_id = ?", v.SourceFormFieldID)

	if len(surveyIDs) > 0 {
		baseQuery = baseQuery.Where("survey_respondents.survey_id IN ?", surveyIDs)
	}
	if report.TingkatWilayah == "5" && len(kecIDs) > 0 {
		baseQuery = baseQuery.Where("respondents.kecamatan_id IN ?", kecIDs)
	} else if report.TingkatWilayah == "4" && len(kelIDs) > 0 {
		baseQuery = baseQuery.Where("respondents.kelurahan_id IN ?", kelIDs)
	}

	type CalcResult struct {
		Name  string  `gorm:"column:name"`
		Val   float64 `gorm:"column:val"`
		Text  string  `gorm:"column:text"`
		Count int64   `gorm:"column:cnt"`
	}

	var res CalcResult

	switch v.CalculationType {
	case "max_row_name":
		err := baseQuery.Session(&gorm.Session{}).
			Select(fmt.Sprintf("%s as name, SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val", groupByCol)).
			Where(fmt.Sprintf("%s IS NOT NULL", groupByCol)).
			Group(groupByCol).Order("val DESC").Limit(1).Scan(&res).Error
		if err != nil || res.Name == "" {
			return "-", nil
		}
		return res.Name, nil

	case "max_row_value":
		err := baseQuery.Session(&gorm.Session{}).
			Select(fmt.Sprintf("%s as name, SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val", groupByCol)).
			Where(fmt.Sprintf("%s IS NOT NULL", groupByCol)).
			Group(groupByCol).Order("val DESC").Limit(1).Scan(&res).Error
		if err != nil {
			return "0", nil
		}
		return fmt.Sprintf("%.0f", res.Val), nil

	case "min_row_name":
		err := baseQuery.Session(&gorm.Session{}).
			Select(fmt.Sprintf("%s as name, SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val", groupByCol)).
			Where(fmt.Sprintf("%s IS NOT NULL", groupByCol)).
			Group(groupByCol).Order("val ASC").Limit(1).Scan(&res).Error
		if err != nil || res.Name == "" {
			return "-", nil
		}
		return res.Name, nil

	case "min_row_value":
		err := baseQuery.Session(&gorm.Session{}).
			Select(fmt.Sprintf("%s as name, SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val", groupByCol)).
			Where(fmt.Sprintf("%s IS NOT NULL", groupByCol)).
			Group(groupByCol).Order("val ASC").Limit(1).Scan(&res).Error
		if err != nil {
			return "0", nil
		}
		return fmt.Sprintf("%.0f", res.Val), nil

	case "total_sum":
		err := baseQuery.Session(&gorm.Session{}).
			Select("SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val").
			Scan(&res).Error
		if err != nil {
			return "0", nil
		}
		return fmt.Sprintf("%.0f", res.Val), nil

	case "average":
		err := baseQuery.Session(&gorm.Session{}).
			Select("AVG(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as val").
			Scan(&res).Error
		if err != nil {
			return "0", nil
		}
		return fmt.Sprintf("%.1f", res.Val), nil

	case "most_frequent_option":
		err := baseQuery.Session(&gorm.Session{}).
			Select("field_responses.answer as text, COUNT(field_responses.id) as cnt").
			Where("field_responses.answer != ''").
			Group("field_responses.answer").Order("cnt DESC").Limit(1).Scan(&res).Error
		if err != nil || res.Text == "" {
			return "-", nil
		}
		return res.Text, nil

	case "most_frequent_count":
		err := baseQuery.Session(&gorm.Session{}).
			Select("COUNT(field_responses.id) as cnt").
			Where("field_responses.answer != ''").
			Group("field_responses.answer").Order("cnt DESC").Limit(1).Scan(&res).Error
		if err != nil {
			return "0", nil
		}
		return fmt.Sprintf("%d", res.Count), nil

	default:
		return "-", nil
	}
}

func (r *laporanRepo) GetTableResponseData(report *models.Report, fieldIDs []int) ([]TableDataRow, error) {
	defer utils.GeneralRecover()
	var results []TableDataRow
	if len(fieldIDs) == 0 {
		return results, nil
	}

	// Parsing JSON IDs
	var surveyIDs, kecIDs, kelIDs []int64
	_ = json.Unmarshal(report.SurveyID, &surveyIDs)
	_ = json.Unmarshal(report.KecamatanID, &kecIDs)
	_ = json.Unmarshal(report.KelurahanID, &kelIDs)

	selectTerritory := ""
	groupByCol := ""

	// Siapkan map dan urutan wilayah agar wilayah yang datanya 0 tetap tampil di tabel
	rowMap := make(map[string]map[int]int64)
	var territoryOrder []string

	// Logika wilayah: 6 (Kota), 5 (Kecamatan), 4 (Kelurahan)
	switch report.TingkatWilayah {
	case "6":
		// Tingkat Kota -> Breakdown per Kecamatan
		selectTerritory = "kecamatans.sub_district_name as territory_name"
		groupByCol = "kecamatans.sub_district_name"

		// Ambil SEMUA kecamatan aktif (deleted_at IS NULL)
		var kecNames []string
		r.dbSlave.Table("kecamatans").Where("deleted_at IS NULL").Order("sub_district_name ASC").Pluck("sub_district_name", &kecNames)
		for _, name := range kecNames {
			territoryOrder = append(territoryOrder, name)
			rowMap[name] = make(map[int]int64)
		}

	case "5":
		// Tingkat Kecamatan -> Breakdown per Kelurahan
		selectTerritory = "kelurahans.village_name as territory_name"
		groupByCol = "kelurahans.village_name"

		// Ambil SEMUA kelurahan aktif di bawah kecamatan terkait
		var kelNames []string
		kelQuery := r.dbSlave.Table("kelurahans").Where("deleted_at IS NULL")
		if len(kecIDs) > 0 {
			kelQuery = kelQuery.Where("sub_district_id IN ?", kecIDs)
		}
		kelQuery.Order("village_name ASC").Pluck("village_name", &kelNames)
		for _, name := range kelNames {
			territoryOrder = append(territoryOrder, name)
			rowMap[name] = make(map[int]int64)
		}

	case "4":
		// Tingkat Kelurahan -> Breakdown per RW
		selectTerritory = "data__rws.nama_rw as territory_name"
		groupByCol = "data__rws.nama_rw"

		// Ambil SEMUA RW aktif di bawah kelurahan terkait
		var rwNames []string
		rwQuery := r.dbSlave.Table("data__rws").Where("deleted_at IS NULL")
		if len(kelIDs) > 0 {
			rwQuery = rwQuery.Where("kelurahan_id IN ?", kelIDs)
		}
		rwQuery.Order("nama_rw ASC").Pluck("nama_rw", &rwNames)
		for _, name := range rwNames {
			territoryOrder = append(territoryOrder, name)
			rowMap[name] = make(map[int]int64)
		}

	default:
		selectTerritory = "'Tidak Diketahui' as territory_name"
		groupByCol = "'Tidak Diketahui'"
	}

	type RawAggData struct {
		TerritoryName string `gorm:"column:territory_name"`
		FormFieldID   int    `gorm:"column:form_field_id"`
		TotalValue    int64  `gorm:"column:total_value"`
	}

	// Join menggunakan survey_respondents dan memfilter deleted_at IS NULL pada wilayah
	dbQuery := r.dbSlave.Table("field_responses").
		Select(fmt.Sprintf("%s, field_responses.form_field_id, SUM(CAST(NULLIF(field_responses.answer, '') AS NUMERIC)) as total_value", selectTerritory)).
		Joins("JOIN survey_respondents ON survey_respondents.id = field_responses.form_response_id").
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Joins("LEFT JOIN kecamatans ON kecamatans.id = respondents.kecamatan_id AND kecamatans.deleted_at IS NULL").
		Joins("LEFT JOIN kelurahans ON kelurahans.id = respondents.kelurahan_id AND kelurahans.deleted_at IS NULL").
		Joins("LEFT JOIN data__rws ON data__rws.id = respondents.rw_id AND data__rws.deleted_at IS NULL").
		Where("field_responses.form_field_id IN ?", fieldIDs)

	if len(surveyIDs) > 0 {
		dbQuery = dbQuery.Where("survey_respondents.survey_id IN ?", surveyIDs)
	}

	// Filter berdasarkan cakupan laporan
	if report.TingkatWilayah == "5" && len(kecIDs) > 0 {
		dbQuery = dbQuery.Where("respondents.kecamatan_id IN ?", kecIDs)
	} else if report.TingkatWilayah == "4" && len(kelIDs) > 0 {
		dbQuery = dbQuery.Where("respondents.kelurahan_id IN ?", kelIDs)
	}

	var rawData []RawAggData
	err := dbQuery.Group(fmt.Sprintf("%s, field_responses.form_field_id", groupByCol)).
		Scan(&rawData).Error

	if err != nil {
		return results, nil
	}

	// Rekap hasil aggregasi ke dalam map master yang sudah kita buat
	for _, d := range rawData {
		tName := d.TerritoryName
		if tName == "" {
			tName = "Lainnya"
		}

		// Jika tName belum terdaftar di map (misal "Lainnya"), tambahkan
		if _, exists := rowMap[tName]; !exists {
			rowMap[tName] = make(map[int]int64)
			territoryOrder = append(territoryOrder, tName)
		}

		rowMap[tName][d.FormFieldID] = d.TotalValue
	}

	for _, tName := range territoryOrder {
		results = append(results, TableDataRow{
			TerritoryName: tName,
			Values:        rowMap[tName], // Otomatis mengisi 0 jika soal tidak ada yang jawab
		})
	}

	return results, nil
}
