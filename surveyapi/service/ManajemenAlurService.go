package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/utils"
	"errors"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
)

type ManajemenAlurService interface {
	CreateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenAlurService struct {
	manajemenAlurRepo repository.ManajemenAlurRepo
}

func NewManajemenAlurService(
	manajemenAlurRepo repository.ManajemenAlurRepo,
) ManajemenAlurService {
	return &manajemenAlurService{
		manajemenAlurRepo: manajemenAlurRepo,
	}
}

func (service *manajemenAlurService) CreateManajemenAlur(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ManajemenAlurPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	if len(payload.Pertanyaan) == 0 {
		return utils.SendError(errors.New("Pertanyaan wajib diisi"), http.StatusBadRequest)
	}

	spew.Dump(payload)

	return utils.SendData(nil, "-")
}
