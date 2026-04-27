package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/utils"
	"strings"

	"gorm.io/gorm"
)

type ManajemenAlurRepo interface {
	RunInTransaction(fn func(txRepo ManajemenAlurRepo) error) error

	CreateFlowDetail(detail *models.FlowDetail) error
	CreateSection(section *models.FlowSection) error
	CreateGroup(group *models.FlowGroup) error
	CreateFlowField(field *models.FlowField) error
	CreateAdvancedOption(logic *models.AdvancedOptionFlow) error
	ManajemenAlurCodeIsExist(code string) (bool, error)
	GetAnswerOptionsByQuestionID(questionID int) ([]int, error)

	GetFlowDetailByCode(code string) (*models.FlowDetail, error)
	GetFlowFieldsByDetailID(detailID int) ([]models.FlowField, error)
	GetAdvancedOptionsByFieldIDs(fieldIDs []int) ([]models.AdvancedOptionFlow, error)
	GetSectionsByIDs(sectionIDs []int) ([]models.FlowSection, error)
	GetGroupsByIDs(groupIDs []int) ([]models.FlowGroup, error)
	GetFlowDetailByNameCaseInsensitive(name string) (*models.FlowDetail, error)

	GetFlowDetailByID(id int) (*models.FlowDetail, error)
	UpdateFlowDetail(detail *models.FlowDetail) error
	SaveSection(section *models.FlowSection) error
	SaveGroup(group *models.FlowGroup) error
	DeleteRoutingByDetailID(detailID int) error

	DeleteAlurByID(detailID int) error
	GetListAlur(req payloads.DatatablePayload) ([]models.FlowDetailDatatableResponse, int64, error)
	GetPreviewSectionByFlowDetailId(detailID int, statusSection string) ([]models.FlowPreviewSection, error)

	GetRawNodesForPreview(detailID int, sectionID int) ([]models.RawNodeData, error)
	GetAnswerOptionsByQuestionIDList(questionIDs []int) ([]models.FormAnswerField, error)
}

type manajemenAlurRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenAlurRepo(dbSlave, dbMaster *gorm.DB) ManajemenAlurRepo {
	defer utils.GeneralRecover()
	return &manajemenAlurRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenAlurRepo) RunInTransaction(fn func(txRepo ManajemenAlurRepo) error) error {
	// Memulai transaksi WAJIB di dbMaster
	tx := repository.dbMaster.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txRepo := &manajemenAlurRepo{
		dbSlave:  tx,
		dbMaster: tx,
	}

	err := fn(txRepo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repository *manajemenAlurRepo) CreateFlowDetail(detail *models.FlowDetail) error {
	return repository.dbMaster.Create(detail).Error
}

func (repository *manajemenAlurRepo) CreateSection(section *models.FlowSection) error {
	return repository.dbMaster.Create(section).Error
}

func (repository *manajemenAlurRepo) CreateGroup(group *models.FlowGroup) error {
	return repository.dbMaster.Create(group).Error
}

func (repository *manajemenAlurRepo) CreateFlowField(field *models.FlowField) error {
	return repository.dbMaster.Create(field).Error
}

func (repository *manajemenAlurRepo) CreateAdvancedOption(logic *models.AdvancedOptionFlow) error {
	return repository.dbMaster.Create(logic).Error
}

func (repository *manajemenAlurRepo) ManajemenAlurCodeIsExist(code string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FlowDetail{}).
		Where("flow_details.code = ?", code).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *manajemenAlurRepo) GetAnswerOptionsByQuestionID(questionID int) ([]int, error) {
	var optionIDs []int

	// Gunakan dbSlave (Saat di dalam RunInTransaction, dbSlave ini sudah otomatis menjadi 'tx' object)
	err := repository.dbSlave.Model(&models.FormAnswerField{}).
		Where("form_field_id = ?", questionID).
		Pluck("id", &optionIDs).Error

	if err != nil {
		return nil, err
	}

	return optionIDs, nil
}

func (repository *manajemenAlurRepo) GetFlowDetailByCode(code string) (*models.FlowDetail, error) {
	var detail models.FlowDetail
	err := repository.dbSlave.Where("code = ?", code).First(&detail).Error
	return &detail, err
}

func (repository *manajemenAlurRepo) GetFlowFieldsByDetailID(detailID int) ([]models.FlowField, error) {
	var fields []models.FlowField
	// Kita order berdasarkan sequence dan ID agar rapi saat dirakit
	err := repository.dbSlave.Where("flow_detail_id = ?", detailID).Order("sequence ASC, id ASC").Find(&fields).Error
	return fields, err
}

func (repository *manajemenAlurRepo) GetAdvancedOptionsByFieldIDs(fieldIDs []int) ([]models.AdvancedOptionFlow, error) {
	var logics []models.AdvancedOptionFlow
	if len(fieldIDs) == 0 {
		return logics, nil
	}
	err := repository.dbSlave.Where("flow_field_id IN ?", fieldIDs).Find(&logics).Error
	return logics, err
}

func (repository *manajemenAlurRepo) GetSectionsByIDs(sectionIDs []int) ([]models.FlowSection, error) {
	var sections []models.FlowSection
	if len(sectionIDs) == 0 {
		return sections, nil
	}
	err := repository.dbSlave.Where("id IN ?", sectionIDs).Find(&sections).Error
	return sections, err
}

func (repository *manajemenAlurRepo) GetGroupsByIDs(groupIDs []int) ([]models.FlowGroup, error) {
	var groups []models.FlowGroup
	if len(groupIDs) == 0 {
		return groups, nil
	}
	err := repository.dbSlave.Where("id IN ?", groupIDs).Find(&groups).Error
	return groups, err
}

func (repository *manajemenAlurRepo) GetFlowDetailByID(id int) (*models.FlowDetail, error) {
	var detail models.FlowDetail
	err := repository.dbSlave.Where("id = ?", id).First(&detail).Error
	return &detail, err
}

func (repository *manajemenAlurRepo) UpdateFlowDetail(detail *models.FlowDetail) error {
	return repository.dbMaster.Save(detail).Error
}

func (repository *manajemenAlurRepo) SaveSection(section *models.FlowSection) error {
	return repository.dbMaster.Save(section).Error
}

func (repository *manajemenAlurRepo) SaveGroup(group *models.FlowGroup) error {
	return repository.dbMaster.Save(group).Error
}

func (repository *manajemenAlurRepo) DeleteRoutingByDetailID(detailID int) error {
	queryLogic := `DELETE FROM advanced_option_flows WHERE flow_field_id IN (SELECT id FROM flow_fields WHERE flow_detail_id = ?)`
	if err := repository.dbMaster.Exec(queryLogic, detailID).Error; err != nil {
		return err
	}

	if err := repository.dbMaster.Where("flow_detail_id = ?", detailID).Delete(&models.FlowField{}).Error; err != nil {
		return err
	}
	return nil
}

func (repository *manajemenAlurRepo) DeleteAlurByID(detailID int) error {
	if err := repository.dbMaster.Where("id = ?", detailID).Delete(&models.FlowDetail{}).Error; err != nil {
		return err
	}
	return nil
}

func (repository *manajemenAlurRepo) GetListAlur(req payloads.DatatablePayload) ([]models.FlowDetailDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.FlowDetailDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("flow_details").
		Select(`
			flow_details.id,
			flow_details.code AS flow_code,
			flow_details.name AS flow_name,
			forms.title AS form_name,
			flow_details.created_at,
			flow_details.updated_at,
			flow_details.created_by AS created_by,
			users.first_name AS created_by_name
		`).
		Joins("LEFT JOIN users ON users.id = flow_details.created_by").
		Joins("LEFT JOIN forms ON forms.id = flow_details.form_id")

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				flow_details.name ILIKE ? OR
				forms.title ILIKE ? OR 
				users.first_name ILIKE ? OR
				DATE(flow_details.created_at) = ? OR
				DATE(flow_details.updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				flow_details.name ILIKE ? OR
				forms.title ILIKE ? OR 
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM flow_details.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM flow_details.updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
			flow_details.name ILIKE ? OR
			forms.title ILIKE ? OR 
			users.first_name ILIKE ?
			`, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "flow_details.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":              "flow_details.id",
			"flow_name":       "flow_details.name",
			"form_name":       "forms.title",
			"created_at":      "flow_details.created_at",
			"updated_at":      "flow_details.updated_at",
			"created_by_name": "users.first_name",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("flow_details.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (repositry *manajemenAlurRepo) GetPreviewSectionByFlowDetailId(detailID int, statusSection string) ([]models.FlowPreviewSection, error) {
	defer utils.GeneralRecover()
	var sections []models.FlowPreviewSection

	// Base query yang sama-sama digunakan
	db := repositry.dbSlave.Table("flow_fields").
		Where("flow_fields.flow_detail_id = ?", detailID).
		Joins("JOIN form_fields ON form_fields.id = flow_fields.form_field_id")

	if statusSection == "1" {
		err := db.Select(`
				flow__sections.id AS section_id,
                flow__sections.name AS section_name,
                COUNT(CASE WHEN form_fields.required = true THEN 1 END) AS total_required_questions,
                COUNT(CASE WHEN form_fields.required = false THEN 1 END) AS total_optional_questions
            `).
			Joins("LEFT JOIN flow__sections ON flow__sections.id = flow_fields.section_id").
			Group("flow__sections.id, flow__sections.name").
			Order("flow__sections.id ASC").
			Find(&sections).Error

		if err != nil {
			return nil, err
		}
	} else {
		err := db.Select(`
				 0 AS section_id,
                NULL AS section_name,
                COUNT(CASE WHEN form_fields.required = true THEN 1 END) AS total_required_questions,
                COUNT(CASE WHEN form_fields.required = false THEN 1 END) AS total_optional_questions
            `).
			Find(&sections).Error

		if err != nil {
			return nil, err
		}
	}

	return sections, nil
}

func (repository *manajemenAlurRepo) GetRawNodesForPreview(detailID int, sectionID int) ([]models.RawNodeData, error) {
	defer utils.GeneralRecover()
	var data []models.RawNodeData

	query := repository.dbSlave.Table("flow_fields").
		Select(`
			flow_fields.id AS flow_field_id,
			flow_fields.sequence,
			flow_fields.form_field_id,
			form_fields.template,
			form_fields.question AS label,
			form_fields.required AS is_required,
			form_fields.image_quantity,
			flow_fields.section_id,
			flow__sections.name AS section_name,
			flow_fields.group_id,
			flow_groups.name AS group_name,
			flow_fields.child_id,
			flow_fields.group_child_id,
			flow_fields.breakdown,
			flow_fields.is_advanced_option,
			flow_fields.form_answer_field_id
		`).
		Joins("JOIN form_fields ON form_fields.id = flow_fields.form_field_id").
		Joins("LEFT JOIN flow__sections ON flow__sections.id = flow_fields.section_id").
		Joins("LEFT JOIN flow_groups ON flow_groups.id = flow_fields.group_id").
		Where("flow_fields.flow_detail_id = ?", detailID)

	// Jika sectionID == 0, query ini akan diabaikan (artinya menarik semua data tanpa memandang section)
	if sectionID != 0 {
		query = query.Where("flow_fields.section_id = ?", sectionID)
	}

	// Urutkan berdasarkan sequence untuk merakit array soal secara berurutan
	err := query.Order("flow_fields.sequence ASC, flow_fields.id ASC").Find(&data).Error
	return data, err
}

func (repository *manajemenAlurRepo) GetAnswerOptionsByQuestionIDList(questionIDs []int) ([]models.FormAnswerField, error) {
	var options []models.FormAnswerField
	if len(questionIDs) == 0 {
		return options, nil
	}
	// 'sequence' pada form_answer_fields digunakan untuk mengurutkan A, B, C, dst.
	err := repository.dbSlave.Where("form_field_id IN ?", questionIDs).Order("sequence ASC").Find(&options).Error
	return options, err
}

func (repository *manajemenAlurRepo) GetFlowDetailByNameCaseInsensitive(name string) (*models.FlowDetail, error) {
	var detail models.FlowDetail
	err := repository.dbSlave.Where("LOWER(name) = LOWER(?)", name).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, err
}
