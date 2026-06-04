package repository

import (
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"os"

	"gorm.io/gorm"
)

type FileRepo interface {
	UploadSurveyImage(ctx context.Context, datauri *string) (*string, error)
	DeleteSurveyImageBulk(ctx context.Context, paths []string) (*bool, error)
	GetPublicImageSurvey(ctx context.Context, path *string) ([]byte, error)
}

type fileRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewFileRepo(dbSlave, dbMaster *gorm.DB) FileRepo {
	defer utils.GeneralRecover()
	return &fileRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *fileRepo) UploadSurveyImage(ctx context.Context, datauri *string) (*string, error) {
	defer utils.GeneralRecover()

	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

	dataBytes, err := utils.HitBackend(ctx, host, "POST", "/upload-survey-image", nil, map[string]interface{}{"datauri": datauri})
	if err != nil {
		return nil, errors.New("Upload file gagal error: " + err.Error())
	}

	var data *string
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data dari DOCAPI")
	}

	return data, nil
}

func (repository *fileRepo) DeleteSurveyImageBulk(ctx context.Context, paths []string) (*bool, error) {
	defer utils.GeneralRecover()

	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

	dataBytes, err := utils.HitBackend(ctx, host, "POST", "/delete-bulk-survey-image", nil, map[string]interface{}{"paths": paths})
	if err != nil {
		return nil, errors.New("Delete file gagal error: " + err.Error())
	}

	var data *bool
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data dari DOCAPI")
	}

	return data, nil
}

func (repository *fileRepo) GetPublicImageSurvey(ctx context.Context, path *string) ([]byte, error) {
	defer utils.GeneralRecover()

	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

	// Pastikan URL dan parameternya sesuai dengan endpoint di DOCAPI Anda
	dataBytes, err := utils.HitBackendNotSecure(ctx, host, "POST", "/view-public-survey-image/:path", nil, map[string]interface{}{"path": path})
	if err != nil {
		return nil, errors.New("Gagal mengambil file: " + err.Error())
	}

	return dataBytes, nil
}
