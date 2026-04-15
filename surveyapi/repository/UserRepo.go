package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"
)

type UserRepo interface {
	GetRespondentById(ctx context.Context, id int64) (*models.Respondent, error)
}

type userRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewUserRepo(dbSlave, dbMaster *gorm.DB) UserRepo {
	defer utils.GeneralRecover()
	return &userRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *userRepo) GetRespondentById(ctx context.Context, id int64) (*models.Respondent, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/raw/:id", newSlug, nil)
	if err != nil {
		spew.Dump(err)
		return nil, errors.New("Gagal mendapatkan data surveyor dari UserAPI")
	}

	var data models.Respondent
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data surveyor dari UserAPI")
	}
	return &data, nil
}
