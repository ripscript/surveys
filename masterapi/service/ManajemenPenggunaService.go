package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ManajemenPenggunaService interface {
	CreateResponden(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenPenggunaService struct {
	manajemenPenggunaRepo repository.ManajemenPenggunaRepo
}

func NewManajemenPenggunaService(
	manajemenPenggunaRepo repository.ManajemenPenggunaRepo,
) ManajemenPenggunaService {
	return &manajemenPenggunaService{
		manajemenPenggunaRepo,
	}
}

func (service *manajemenPenggunaService) CreateResponden(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateRespondenPayload

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

	role, err := service.manajemenPenggunaRepo.GetRoleByID(payload.RoleId)
	if err != nil {
		return utils.SendError(errors.New("Role tidak ditemukan"), http.StatusBadRequest)
	}

	if role.Name != "" {
		role.Name = strings.ToLower(role.Name)
	}

	var validationErrors []string

	isRoleRT := role.Name == "rt"
	isRoleRW := role.Name == "rw" || isRoleRT
	isRoleLurah := role.Name == "lurah" || isRoleRW
	isRoleCamat := role.Name == "camat" || isRoleLurah

	if isRoleCamat && payload.KecamatanId == nil {
		validationErrors = append(validationErrors, "Kecamatan tidak boleh kosong untuk role ini")
	}
	if isRoleLurah && payload.KelurahanId == nil {
		validationErrors = append(validationErrors, "Kelurahan tidak boleh kosong untuk role ini")
	}
	if isRoleRW && payload.RwId == nil {
		validationErrors = append(validationErrors, "RW tidak boleh kosong untuk role ini")
	}
	if isRoleRT && payload.RtId == nil {
		validationErrors = append(validationErrors, "RT tidak boleh kosong untuk role ini")
	}

	if len(validationErrors) > 0 {
		errMsg := strings.Join(validationErrors, " | ")
		return utils.SendError(errors.New(errMsg), http.StatusBadRequest)
	}

	return utils.SendData(nil, "Berhasil create data")
}
