package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DashboardMetricMappingRepository interface {
	Create(ctx context.Context, m *models.DashboardMetricMapping) error
	Update(ctx context.Context, m *models.DashboardMetricMapping) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error)
	FindByMetricAndForm(ctx context.Context, metricID, formID int64) (*models.DashboardMetricMapping, error)
	ListByFormAndCategory(ctx context.Context, formID int64, category string) ([]models.DashboardMetricMapping, error)

	GetList(ctx context.Context, req payloads.DatatablePayload, formCode string, dashboardMetricID int64) ([]MappingListRow, int64, error)
	FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error)
	CountOptionStatusByMappingID(ctx context.Context, mappingID int64) (int64, error)
	CountByMetricID(ctx context.Context, metricID int64) (int64, error)

	GetFlowDetailById(ctx context.Context, flowId int64) (*FlowDetailLite, error)
	GetSurveyQuestionsForMapping(ctx context.Context, flowDetailID int64) ([]MappingSurveyRawRow, error)
	SaveSurveyMapping(ctx context.Context, formID int64, questions []payloads.SaveMappingSurveyQuestionPayload) error
}

type FlowDetailLite struct {
	ID       int64
	Name     string
	Code     string
	FormID   int64
	FormCode string
}

type MappingSurveyRawRow struct {
	FormFieldID       int64
	Question          string
	Description       string
	Template          string
	Sequence          int
	SectionID         *int64
	SectionName       *string
	DashboardMetricID *int64
	MetricKey         *string
	AnswerOptionID    *int64
	AnswerLabel       *string
	StatusID          *int64
}

type dashboardMetricMappingRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricMappingRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricMappingRepository {
	return &dashboardMetricMappingRepository{dbMaster: dbMaster, dbSlave: dbSlave}
}

type MappingListRow struct {
	ID                int64
	DashboardMetricID int64
	MetricKey         string
	MetricLabel       string
	ExpectedTemplate  string
	Category          string
	FormID            int64
	FormTitle         string
	FormCode          string
	FormFieldID       int64
	FormFieldQuestion string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (r *dashboardMetricMappingRepository) Create(ctx context.Context, m *models.DashboardMetricMapping) error {
	return r.dbMaster.WithContext(ctx).Create(m).Error
}

func (r *dashboardMetricMappingRepository) Update(ctx context.Context, m *models.DashboardMetricMapping) error {
	return r.dbMaster.WithContext(ctx).Save(m).Error
}

func (r *dashboardMetricMappingRepository) Delete(ctx context.Context, id int64) error {
	return r.dbMaster.WithContext(ctx).Delete(&models.DashboardMetricMapping{}, id).Error
}

func (r *dashboardMetricMappingRepository) FindByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).Preload("DashboardMetric").First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) FindByMetricAndForm(ctx context.Context, metricID, formID int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Where("dashboard_metric_id = ? AND form_id = ?", metricID, formID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) ListByFormAndCategory(ctx context.Context, formID int64, category string) ([]models.DashboardMetricMapping, error) {
	var list []models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Joins("JOIN dashboard_metrics dm ON dm.id = dashboard_metric_mappings.dashboard_metric_id").
		Where("dashboard_metric_mappings.form_id = ? AND dm.category = ?", formID, category).
		Preload("DashboardMetric").
		Find(&list).Error
	return list, err
}

func (r *dashboardMetricMappingRepository) GetList(ctx context.Context, req payloads.DatatablePayload, formCode string, dashboardMetricID int64) ([]MappingListRow, int64, error) {
	defer utils.GeneralRecover()

	var data []MappingListRow
	var totalData int64

	db := r.dbSlave.WithContext(ctx).
		Table("dashboard_metric_mappings dmm").
		Select(`
			dmm.id,
			dmm.dashboard_metric_id,
			dm.metric_key,
			dm.label as metric_label,
			dm.expected_template,
			dm.category,
			dmm.form_id,
			f.title as form_title,
			f.code as form_code,
			dmm.form_field_id,
			ff.question as form_field_question,
			dmm.created_at,
			dmm.updated_at
		`).
		Joins("JOIN dashboard_metrics dm ON dm.id = dmm.dashboard_metric_id").
		Joins("JOIN forms f ON f.id = dmm.form_id").
		Joins("JOIN form_fields ff ON ff.id = dmm.form_field_id")

	if formCode != "" {
		db = db.Where("f.code = ?", formCode)
	}
	if dashboardMetricID > 0 {
		db = db.Where("dmm.dashboard_metric_id = ?", dashboardMetricID)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		db = db.Where(`
			dm.metric_key ILIKE ? OR
			dm.label ILIKE ? OR
			f.title ILIKE ? OR
			ff.question ILIKE ?
		`, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := db.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	finalOrderBy := "dmm.id"
	finalOrderDir := "desc"
	allowedOrderCols := map[string]string{
		"id":         "dmm.id",
		"metric_key": "dm.metric_key",
		"form_title": "f.title",
		"created_at": "dmm.created_at",
		"updated_at": "dmm.updated_at",
	}
	if mappedCol, ok := allowedOrderCols[req.OrderBy]; ok {
		finalOrderBy = mappedCol
	}
	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	}
	db = db.Order(finalOrderBy + " " + finalOrderDir)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 25
	}
	offset := (req.Page - 1) * req.Limit

	if err := db.Limit(req.Limit).Offset(offset).Scan(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (r *dashboardMetricMappingRepository) FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Preload("DashboardMetric").
		Preload("Form").
		Preload("FormField").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) CountOptionStatusByMappingID(ctx context.Context, mappingID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Table("dashboard_option_status_mappings").
		Where("dashboard_metric_mapping_id = ?", mappingID).
		Count(&count).Error
	return count, err
}

func (r *dashboardMetricMappingRepository) CountByMetricID(ctx context.Context, metricID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Model(&models.DashboardMetricMapping{}).
		Where("dashboard_metric_id = ?", metricID).
		Count(&count).Error
	return count, err
}

func (r *dashboardMetricMappingRepository) GetFlowDetailById(ctx context.Context, flowId int64) (*FlowDetailLite, error) {
	var row FlowDetailLite
	err := r.dbSlave.WithContext(ctx).
		Table("flow_details").
		Select(`
			flow_details.id,
			flow_details.name,
			flow_details.code,
			flow_details.form_id,
			forms.code AS form_code
		`).
		Joins("JOIN forms ON forms.id = flow_details.form_id").
		Where("flow_details.id = ?", flowId).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *dashboardMetricMappingRepository) GetSurveyQuestionsForMapping(ctx context.Context, flowDetailID int64) ([]MappingSurveyRawRow, error) {
	var rows []MappingSurveyRawRow

	err := r.dbSlave.WithContext(ctx).
		Table("(?) AS uniq_ff", r.dbSlave.
			Table("flow_fields").
			Select("MIN(flow_fields.id) AS flow_field_id").
			Where("flow_fields.flow_detail_id = ?", flowDetailID).
			Group("flow_fields.form_field_id"),
		).
		Joins("JOIN flow_fields ON flow_fields.id = uniq_ff.flow_field_id").
		Select(`
			flow_fields.form_field_id,
			form_fields.question,
			form_fields.deskripsi AS description,
			form_fields.template,
			flow_fields.sequence,
			flow_fields.section_id,
			flow__sections.name AS section_name,
			dmm.dashboard_metric_id,
			dm.metric_key,
			faf.id AS answer_option_id,
			faf.option AS answer_label,
			dosm.dashboard_metric_status_id AS status_id
		`).
		Joins("JOIN form_fields ON form_fields.id = flow_fields.form_field_id").
		Joins("LEFT JOIN flow__sections ON flow__sections.id = flow_fields.section_id").
		Joins("LEFT JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = form_fields.id").
		Joins("LEFT JOIN dashboard_metrics dm ON dm.id = dmm.dashboard_metric_id").
		Joins("LEFT JOIN form_answer_fields faf ON faf.form_field_id = form_fields.id").
		Joins(`LEFT JOIN dashboard_option_status_mappings dosm
			ON dosm.answer_option_id = faf.id
			AND dosm.dashboard_metric_mapping_id = dmm.id`).
		Order("flow_fields.sequence ASC, flow_fields.form_field_id ASC, faf.sequence ASC, faf.id ASC").
		Find(&rows).Error

	return rows, err
}

func (r *dashboardMetricMappingRepository) SaveSurveyMapping(ctx context.Context, formID int64, questions []payloads.SaveMappingSurveyQuestionPayload) error {
	return r.dbMaster.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("form_id = ?", formID).
			Delete(&models.DashboardMetricMapping{}).Error; err != nil {
			return err
		}

		for _, q := range questions {
			mapping := models.DashboardMetricMapping{
				DashboardMetricID: *q.DashboardMetricID,
				FormID:            formID,
				FormFieldID:       q.FormFieldID,
			}
			if err := tx.Create(&mapping).Error; err != nil {
				return err
			}

			if len(q.Options) == 0 {
				continue
			}

			optionMappings := make([]models.DashboardOptionStatusMapping, 0, len(q.Options))
			for _, opt := range q.Options {
				optionMappings = append(optionMappings, models.DashboardOptionStatusMapping{
					DashboardMetricMappingID: mapping.ID,
					AnswerOptionID:           opt.AnswerOptionID,
					DashboardMetricStatusID:  *opt.DashboardMetricStatusID,
				})
			}
			if err := tx.Create(&optionMappings).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
