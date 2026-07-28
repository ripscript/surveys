package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type UserRepo interface {
	GetRespondentById(ctx context.Context, id int64) (*models.RespondentModel_1, error)
	GetRespondentDetailById(ctx context.Context, id int64) (*models.DetailRespondent, error)

	GetRespondentByKecamatanId(ctx context.Context, kecamatan_id int64) (*models.RespondentModel_1, error)
	GetRespondentByKelurahanId(ctx context.Context, kelurahan_id int64) (*models.RespondentModel_1, error)
	GetRespondentByRWId(ctx context.Context, rw_id int64) (*models.RespondentModel_1, error)
	GetRespondentByRTId(ctx context.Context, rt_id int64) (*models.RespondentModel_1, error)
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

func (repository *userRepo) GetRespondentById(ctx context.Context, id int64) (*models.RespondentModel_1, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/raw/:id", newSlug, nil)
	if err != nil {
		return nil, errors.New("Gagal mendapatkan data respondent dari UserAPI")
	}

	var data models.RespondentModel_1
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal memparsing data respondent dari UserAPI")
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

func (repository *userRepo) GetRespondentByKecamatanId(ctx context.Context, kecamatan_id int64) (*models.RespondentModel_1, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"kecamatan_id": strconv.FormatInt(kecamatan_id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/kecamatan/:kecamatan_id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data respondent dari UserAPI")
	}

	var data models.RespondentModel_1
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data respondent dari UserAPI")
	}
	return &data, nil
}

func (repository *userRepo) GetRespondentByKelurahanId(ctx context.Context, kelurahan_id int64) (*models.RespondentModel_1, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"kelurahan_id": strconv.FormatInt(kelurahan_id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/kelurahan/:kelurahan_id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data respondent dari UserAPI")
	}

	var data models.RespondentModel_1
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data respondent dari UserAPI")
	}
	return &data, nil
}

func (repository *userRepo) GetRespondentByRWId(ctx context.Context, rw_id int64) (*models.RespondentModel_1, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"rw_id": strconv.FormatInt(rw_id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/rw/:rw_id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data respondent dari UserAPI")
	}

	var data models.RespondentModel_1
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data respondent dari UserAPI")
	}
	return &data, nil
}

func (repository *userRepo) GetRespondentByRTId(ctx context.Context, rt_id int64) (*models.RespondentModel_1, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"rt_id": strconv.FormatInt(rt_id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/rt/:rt_id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data respondent dari UserAPI")
	}

	var data models.RespondentModel_1
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data respondent dari UserAPI")
	}
	return &data, nil
}
