package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type TemplateFormulirRepo interface {
	GetDetailPertanyaan(ctx context.Context, id int64) (*models.FormFieldWithOption, error)
}

type templateFormulirRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewTemplateFormulirRepo(dbSlave, dbMaster *gorm.DB) TemplateFormulirRepo {
	defer utils.GeneralRecover()
	return &templateFormulirRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *templateFormulirRepo) GetDetailPertanyaan(ctx context.Context, id int64) (*models.FormFieldWithOption, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/template/formulir-pertanyaan/detail-pertanyaan/:id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data detail pertanyaan dari SurveyAPI")
	}

	var data models.FormFieldWithOption
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data detail pertanyaan dari SurveyAPI")
	}
	return &data, nil
}
