package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type UserRepo interface {
	GetRespondentById(ctx context.Context, id int64) (*models.Respondent, error)
	GetRespondentDetailById(ctx context.Context, id int64) (*models.DetailRespondent, error)
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
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data surveyor dari UserAPI")
	}

	var data models.Respondent
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data surveyor dari UserAPI")
	}
	return &data, nil
}

func (repository *userRepo) GetRespondentDetailById(ctx context.Context, id int64) (*models.DetailRespondent, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/:id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data surveyor dari UserAPI")
	}

	var data models.DetailRespondent
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data surveyor dari UserAPI")
	}
	return &data, nil
}
