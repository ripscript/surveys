package service

import (
	"backend/docapi/models"
	"backend/docapi/utils"
	"backend/siccore/pb"
	"fmt"
	"net/url"
)

type TrxService interface {
	GenerateReceipt(req map[string]interface{}) (*pb.ProxyResponse, error)
	UploadBanner(usr models.JwtCustomClaims, req map[string]interface{}, param url.Values, path string, module string) (*pb.ProxyResponse, error)
}

type trxService struct {
	trxRepo string
}

func NewTrxService(
	trxRepo string,
) TrxService {
	return &trxService{
		trxRepo,
	}
}

func (service *trxService) GenerateReceipt(req map[string]interface{}) (*pb.ProxyResponse, error) {
	fmt.Println(req)
	return utils.SendData(req)
}

func (service *trxService) UploadBanner(usr models.JwtCustomClaims, req map[string]interface{}, param url.Values, path string, module string) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	uploadFolder := path
	fileName := ""
	fileExtension := ""
	fileSize := 0
	var file string

	if req["file_name"] != nil {
		fileName = req["file_name"].(string)
	}

	if req["file_extension"] != nil {
		fileExtension = req["file_extension"].(string)
	}

	if req["file_size"] != nil {
		fileSize = int(req["file_size"].(float64))
	}

	if req["file"] != nil {
		file = req["file"].(string)
	}

	filename, err := utils.UploadService(fileName, uploadFolder, file, module, int(fileSize))
	if err != nil {
		return utils.SendError(err, 500)
	}

	fmt.Println(module, fileExtension, filename, uploadFolder)

	res := map[string]interface{}{
		"file_name": filename,
		"file_url":  "/view/webroot/files/" + uploadFolder + "/" + filename,
	}

	return utils.SendData(res, "Berhasil")
}
