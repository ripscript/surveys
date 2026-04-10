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

	GetFlowDetailByID(id int) (*models.FlowDetail, error)
	UpdateFlowDetail(detail *models.FlowDetail) error
	SaveSection(section *models.FlowSection) error
	SaveGroup(group *models.FlowGroup) error
	DeleteRoutingByDetailID(detailID int) error

	DeleteAlurByID(detailID int) error
	GetListAlur(req payloads.DatatablePayload) ([]models.FlowDetailDatatableResponse, int64, error)
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

// func (repositry *manajemenAlurRepo) GetSectionByFlowId(detailID int) ([]models.FlowSection, error) {
// 	var flowDetail models.FlowDetail
// 	err := repositry.dbSlave.Where("id = ?", detailID).First(&flowDetail).Error
// 	if err != nil {
// 		return nil, err
// 	}

// }
