package repository

import (
	"backend/masterapi/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type FormFieldRepository interface {
	FindByID(ctx context.Context, id int64) (*models.FormField, error)
	AnswerOptionBelongsToField(ctx context.Context, answerOptionID, formFieldID int64) (bool, error)
}

type formFieldRepository struct {
	dbSlave *gorm.DB
}

func NewFormFieldRepository(dbSlave *gorm.DB) FormFieldRepository {
	return &formFieldRepository{dbSlave: dbSlave}
}

func (r *formFieldRepository) FindByID(ctx context.Context, id int64) (*models.FormField, error) {
	var f models.FormField
	err := r.dbSlave.WithContext(ctx).First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// AnswerOptionBelongsToField memastikan answer_option_id yang dikirim admin
// benar-benar salah satu opsi jawaban milik form_field_id tersebut
// (form_answer_fields.form_field_id harus cocok) -- bukan opsi punya pertanyaan lain.
func (r *formFieldRepository) AnswerOptionBelongsToField(ctx context.Context, answerOptionID, formFieldID int64) (bool, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Table("form_answer_fields").
		Where("id = ? AND form_field_id = ?", answerOptionID, formFieldID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
