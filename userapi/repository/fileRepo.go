package repository

import (
	"backend/userapi/utils"
	"context"
	"encoding/json"
	"errors"
	"os"

	"gorm.io/gorm"
)

type FileRepo interface {
	UploadFotoProfil(ctx context.Context, datauri *string) (*string, error)
	DeleteFotoProfilBulk(ctx context.Context, paths []string) (*bool, error)
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

func (repository *fileRepo) UploadFotoProfil(ctx context.Context, datauri *string) (*string, error) {
	defer utils.GeneralRecover()

	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

	dataBytes, err := utils.HitBackend(ctx, host, "POST", "/upload-foto-profil", nil, map[string]interface{}{"datauri": datauri})
	if err != nil {
		return nil, errors.New("Upload file gagal error: " + err.Error())
	}

	var data *string
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data dari DOCAPI")
	}

	return data, nil
}

func (repository *fileRepo) DeleteFotoProfilBulk(ctx context.Context, paths []string) (*bool, error) {
	defer utils.GeneralRecover()

	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

	dataBytes, err := utils.HitBackend(ctx, host, "POST", "/delete-bulk-foto-profil", nil, map[string]interface{}{"paths": paths})
	if err != nil {
		return nil, errors.New("Delete file gagal error: " + err.Error())
	}

	var data *bool
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data dari DOCAPI")
	}

	return data, nil
}

// func (repository *fileRepo) GetLaporanKontenImageBytes(ctx context.Context, path string) ([]byte, string, error) {
// 	defer utils.GeneralRecover()

// 	host := os.Getenv("DOCAPI_HOST") + ":" + os.Getenv("DOCAPI_PORT")

// 	// HitBackend
// 	dataBytes, err := utils.HitBackendNotSecure(ctx, host, "GET", "/internal/get-laporan-konten-image-bytes/:path", map[string]interface{}{"path": path}, nil)
// 	if err != nil {
// 		spew.Dump(err)
// 		return nil, "", errors.New("Gagal mengambil gambar dari DOCAPI: " + err.Error())
// 	}

// 	mimeType := http.DetectContentType(dataBytes)

// 	return dataBytes, mimeType, nil
// }
