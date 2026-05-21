package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"strings"

	"gorm.io/gorm"
)

type TemplateFormulirPertanyaanRepo interface {
	GetFormById(id int64) (*models.Form, error)
	GetFormByCode(code string) (*models.Form, error)
	GetFormByTitle(title string) (*models.Form, error)
	TemplateFormCodeIsExist(code string) (bool, error)
	TemplateFormAttributeIsExists(attribute string) (bool, error)
	GetFormFieldByFormId(formId int) ([]models.FormField, error)
	GetFormFieldOptionsByFormFieldId(formFieldId int) ([]models.FormAnswerField, error)
	CreateFormWithFields(form models.Form, formFields []models.FormFieldWithOption) (*models.Form, error)
	GetFormFieldByAttribute(attribute string) (*models.FormField, error)
	AttributeIsExistInFormField(attribute string) (bool, error)
	UpdateFormTransactionTx(form models.Form, fields []models.UpdateFormFieldWithOption, attributeToKeep []string, isQuestionUpdate bool) error
	GetAllFieldsByFormID(id int) ([]models.FormField, error)
	GetFullFormByCode(code string) (*models.FullForm, error)
	CreateFormTransactionTx(form models.Form, fields []models.FormFieldWithOption) error
	DeleteFormTransactionByCode(code string) error
	GetListTemplateFormulirPertanyaan(req payloads.DatatablePayload) ([]models.FormDatatableResponse, int64, error)
	GetFormulirPertanyaanOptions(req payloads.FormulirPertanyaanOptionsPayload) ([]response.StringOptionItem, int64, error)
	GetPertanyaanOptionsByFormCode(req payloads.PertanyaanOptionsPayload, formCode string) ([]response.FormFieldOptionItem, int64, error)
	GetPertanyaanById(id int) (*models.FormFieldWithOption, error)
	IsQuestionExistsById(id int) (bool, error)
	IsQuestionOptionsExistById(id int) (bool, error)
	GetMultipleChoiceOptions(req payloads.MultipleChoiceOptionsPayload, form_field_id int64) ([]response.OptionItem, int64, error)

	GetQuestionTemplateById(id int) (string, error)
	GetFormFieldByID(id int) (*models.FormField, error)
	IsTemplateFormulirPertanyaanUsed(id int64) (bool, error)
}

type templateFormulirPertanyaanRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewTemplateFormulirPertanyaanRepo(dbSlave, dbMaster *gorm.DB) *templateFormulirPertanyaanRepo {
	defer utils.GeneralRecover()
	return &templateFormulirPertanyaanRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *templateFormulirPertanyaanRepo) GetFormById(id int64) (*models.Form, error) {
	defer utils.GeneralRecover()

	var data models.Form
	db := repository.dbSlave

	err := db.Select("forms.*").
		Where("forms.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormByTitle(title string) (*models.Form, error) {
	defer utils.GeneralRecover()

	var data models.Form
	db := repository.dbSlave

	err := db.Select("forms.*").
		Where("forms.title = ?", title).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormByCode(code string) (*models.Form, error) {
	defer utils.GeneralRecover()

	var data models.Form
	db := repository.dbSlave

	err := db.Select("forms.*").
		Where("forms.code = ?", code).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateFormulirPertanyaanRepo) TemplateFormCodeIsExist(code string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.Form{}).
		Where("forms.code = ?", code).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *templateFormulirPertanyaanRepo) TemplateFormAttributeIsExists(attribute string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FormField{}).
		Where("attribute = ?", attribute).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormFieldByFormId(formId int) ([]models.FormField, error) {
	defer utils.GeneralRecover()

	var data []models.FormField
	db := repository.dbSlave

	err := db.Select("form_fields.*").
		Where("form_fields.form_id = ?", formId).
		Order("form_fields.sequence ASC").
		Find(&data).Error

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormFieldOptionsByFormFieldId(formFieldId int) ([]models.FormAnswerField, error) {
	defer utils.GeneralRecover()

	var data []models.FormAnswerField
	db := repository.dbSlave

	err := db.Select("form_answer_fields.*").
		Where("form_answer_fields.form_field_id = ?", formFieldId).
		Order("form_answer_fields.sequence ASC").
		Find(&data).Error

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (repository *templateFormulirPertanyaanRepo) CreateFormWithFields(form models.Form, formFields []models.FormFieldWithOption) (*models.Form, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&form).Error; err != nil {
			return err
		}

		for _, v := range formFields {
			v.FormId = form.ID
			if err := tx.Create(&v.FormField).Error; err != nil {
				return err
			}

			for _, option := range v.T_Options {
				option.FormId = form.ID
				option.FormFieldId = v.ID
				if err := tx.Create(&option).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &form, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormFieldByAttribute(attribute string) (*models.FormField, error) {
	defer utils.GeneralRecover()

	var data models.FormField
	db := repository.dbSlave

	err := db.Select("form_fields.*").
		Where("form_fields.attribute = ?", attribute).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateFormulirPertanyaanRepo) GetAllFieldsByFormID(id int) ([]models.FormField, error) {
	defer utils.GeneralRecover()

	var data []models.FormField
	db := repository.dbSlave

	err := db.Select("form_fields.*").
		Where("form_fields.form_id = ?", id).
		Find(&data).Error

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (repository *templateFormulirPertanyaanRepo) AttributeIsExistInFormField(attribute string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FormField{}).
		Where("form_fields.attribute = ?", attribute).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *templateFormulirPertanyaanRepo) UpdateFormTransactionTx(form models.Form, fields []models.UpdateFormFieldWithOption, attributeToKeep []string, isQuestionUpdate bool) error {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&form).Error; err != nil {
			return err
		}

		if isQuestionUpdate == true {
			if len(attributeToKeep) > 0 {
				if err := tx.Where("form_id = ? AND attribute NOT IN ?", form.ID, attributeToKeep).Delete(&models.FormField{}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Where("form_id = ?", form.ID).Delete(&models.FormField{}).Error; err != nil {
					return err
				}
			}

			for _, fieldWithOpt := range fields {
				field := fieldWithOpt.FormField

				if err := tx.Save(&field).Error; err != nil {
					return err
				}

				if field.Template == "multiple-choices" {
					if len(fieldWithOpt.OptionsToKeep) > 0 {
						if err := tx.Where("form_field_id = ? AND id NOT IN ?", field.ID, fieldWithOpt.OptionsToKeep).Delete(&models.FormAnswerField{}).Error; err != nil {
							return err
						}
					} else {
						if err := tx.Where("form_field_id = ?", field.ID).Delete(&models.FormAnswerField{}).Error; err != nil {
							return err
						}
					}

					for _, opt := range fieldWithOpt.Options {
						opt.FormId = form.ID
						opt.FormFieldId = field.ID
						if err := tx.Save(&opt).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repository *templateFormulirPertanyaanRepo) GetFullFormByCode(code string) (*models.FullForm, error) {
	var form models.FullForm

	err := repository.dbMaster.
		Preload("FormFields").
		Preload("FormFields.Options").
		Where("code = ?", code).
		First(&form).Error

	if err != nil {
		return nil, err
	}
	return &form, nil
}

func (repository *templateFormulirPertanyaanRepo) CreateFormTransactionTx(form models.Form, fields []models.FormFieldWithOption) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&form).Error; err != nil {
			return err
		}

		for _, fieldWithOpt := range fields {
			field := fieldWithOpt.FormField
			field.FormId = form.ID

			if err := tx.Create(&field).Error; err != nil {
				return err
			}

			if field.Template == "multiple-choices" {
				for _, opt := range fieldWithOpt.T_Options {
					opt.FormId = form.ID
					opt.FormFieldId = field.ID
					if err := tx.Create(&opt).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func (repository *templateFormulirPertanyaanRepo) DeleteFormTransactionByCode(code string) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		var form models.Form
		if err := tx.Where("code = ?", code).First(&form).Error; err != nil {
			return err
		}

		if err := tx.Delete(&form).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repository *templateFormulirPertanyaanRepo) GetListTemplateFormulirPertanyaan(req payloads.DatatablePayload) ([]models.FormDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.FormDatatableResponse
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	db := repository.dbSlave.Table("forms").
		Select(`
			forms.id AS id,
			forms.code AS code,
			forms.title AS title,
			forms.created_at AS created_at,
			forms.updated_at AS updated_at,
			forms.deleted_at AS deleted_at,
			forms.flag_tematik AS flag_tematik,
			users.first_name AS created_by_name
		`).
		Joins("LEFT JOIN users ON users.id = forms.user_id").
		Where("forms.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("forms").
		Joins("LEFT JOIN users ON users.id = forms.user_id").
		Where("forms.deleted_at IS NULL")

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				forms.title ILIKE ? OR 
				users.first_name ILIKE ? OR
				DATE(forms.created_at) = ? OR
				DATE(forms.updated_at) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				forms.title ILIKE ? OR 
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM forms.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM forms.updated_at)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				forms.title ILIKE ? OR 
				users.first_name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("forms.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":              "forms.id",
		"code":            "forms.code",
		"name":            "forms.title",
		"created_at":      "forms.created_at",
		"updated_at":      "forms.updated_at",
		"flag_tematik":    "forms.flag_tematik",
		"created_by_name": "users.first_name",
	}

	finalOrderBy := "forms.id"
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

func (repository *templateFormulirPertanyaanRepo) GetFormulirPertanyaanOptions(req payloads.FormulirPertanyaanOptionsPayload) ([]response.StringOptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.StringOptionItem
	var totalData int64

	db := repository.dbSlave.Table("forms").
		Select(`
			forms.code AS id, 
			forms.title AS label
		`).
		Where("forms.deleted_at IS NULL")

	if len(req.IDs) > 0 {
		db = db.Where("forms.code IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("forms.title ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("forms.id asc")

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

func (repository *templateFormulirPertanyaanRepo) GetPertanyaanOptionsByFormCode(req payloads.PertanyaanOptionsPayload, formCode string) ([]response.FormFieldOptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.FormFieldOptionItem
	var totalData int64

	db := repository.dbSlave.Table("form_fields").
		Joins("JOIN forms ON forms.id = form_fields.form_id").
		Where("forms.code = ?", formCode).
		Select(`
			form_fields.id AS id, 
			form_fields.question AS label,
			form_fields.template AS type
		`).
		Where("form_fields.deleted_at IS NULL")

	if req.Type != nil && *req.Type != "" {
		db = db.Where("form_fields.template = ?", *req.Type)
	}

	if len(req.ExcludeIDs) > 0 {
		db = db.Where("form_fields.id NOT IN ?", req.ExcludeIDs)
	}

	if len(req.IDs) > 0 {
		db = db.Where("form_fields.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("form_fields.question ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("form_fields.sequence asc")

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

func (repository *templateFormulirPertanyaanRepo) GetPertanyaanById(id int) (*models.FormFieldWithOption, error) {
	defer utils.GeneralRecover()

	var data models.FormFieldWithOption
	db := repository.dbSlave

	err := db.Select("form_fields.*").
		Where("form_fields.id = ?", id).
		First(&data.FormField).Error

	if err != nil {
		return nil, err
	}

	if data.Template == "multiple-choices" {
		err = db.Select("form_answer_fields.*").
			Where("form_answer_fields.form_field_id = ?", data.ID).
			Order("form_answer_fields.sequence ASC").
			Find(&data.T_Options).Error

		if err != nil {
			return nil, err
		}
	} else {
		data.T_Options = make([]models.FormAnswerField, 0)
	}

	return &data, nil
}

func (repository *templateFormulirPertanyaanRepo) IsQuestionExistsById(id int) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FormField{}).
		Where("form_fields.id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *templateFormulirPertanyaanRepo) IsQuestionOptionsExistById(id int) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FormAnswerField{}).
		Where("form_answer_fields.id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *templateFormulirPertanyaanRepo) GetMultipleChoiceOptions(req payloads.MultipleChoiceOptionsPayload, form_field_id int64) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("form_answer_fields").
		Joins("JOIN form_fields ON form_fields.id = form_answer_fields.form_field_id").
		Select(`
			form_answer_fields.id AS id, 
			form_answer_fields.option AS label
		`).
		Where("form_answer_fields.form_field_id = ?", form_field_id)

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("form_answer_fields.option ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("form_answer_fields.sequence asc")

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

func (repository *templateFormulirPertanyaanRepo) GetQuestionTemplateById(id int) (string, error) {
	defer utils.GeneralRecover()

	var field models.FormField
	err := repository.dbSlave.Select("template").Where("id = ?", id).First(&field).Error
	if err != nil {
		return "", err
	}

	return field.Template, nil
}

func (repository *templateFormulirPertanyaanRepo) GetFormFieldByID(id int) (*models.FormField, error) {
	defer utils.GeneralRecover()

	var data models.FormField
	db := repository.dbSlave

	err := db.Select("form_fields.*").
		Where("form_fields.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil

}

func (repository *templateFormulirPertanyaanRepo) IsTemplateFormulirPertanyaanUsed(id int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FlowDetail{}).
		Where("flow_details.form_id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
