package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type PenggunaService interface {
	Login(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	Logout(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
}

type penggunaService struct {
	penggunaRepo repository.PenggunaRepo
}

func NewPenggunaService(
	penggunaRepo repository.PenggunaRepo,

) PenggunaService {
	return &penggunaService{
		penggunaRepo,
	}
}

func (service *penggunaService) Login(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.LoginPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Password == "" {
		return utils.SendError(errors.New("password wajib diisi"), http.StatusBadRequest)
	}

	respondent, requestEmail, err := service.penggunaRepo.FindRespondentByRole(payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if respondent == nil {
		return utils.SendError(errors.New("data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	storedUser, err := service.penggunaRepo.FindUserByRespondentID(respondent.ID)
	if err != nil {
		return utils.SendError(errors.New("user tidak ditemukan"), http.StatusUnauthorized)
	}

	loginCount, err := service.penggunaRepo.CountFailedLogin(int(storedUser.ID))
	if err == nil && loginCount >= 3 {
		service.penggunaRepo.BlockRespondent(respondent.ID)
		return utils.SendError(errors.New("akun diblokir"), http.StatusUnauthorized)
	}
	isPasswordValid := utils.VerifyPassword(storedUser.Password, payload.Password)
	if !isPasswordValid {
		if storedUser.Email != "admin@gmail.com" {
			service.penggunaRepo.InsertFailedLogin(storedUser.ID)
		}
		return utils.SendError(errors.New("password salah"), http.StatusUnauthorized)
	}

	if !utils.IsDirectRole(respondent.RoleID) {
		validJabatan, err := service.penggunaRepo.CheckActiveJabatan(respondent.ID)
		if err != nil || !validJabatan {
			return utils.SendError(errors.New("jabatan tidak aktif"), http.StatusUnauthorized)
		}
	}

	claims := &models.JwtCustomClaims{
		ID:    int64(storedUser.ID),
		Name:  respondent.Name,
		Email: requestEmail,
		Role:  respondent.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return utils.SendError(errors.New("gagal membuat token"), http.StatusBadRequest)
	}

	storedUser.LastLogin = utils.TimeNow()
	// service.penggunaRepo.EditLastLog(storedUser)

	return utils.SendData(encryptedToken, "Login berhasil")
}

func (service *penggunaService) Logout(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	userID := uint(usr.ID)

	_, err := service.penggunaRepo.FindUserByID(int(userID))
	if err != nil {
		return utils.SendError(errors.New("user tidak ditemukan"), http.StatusUnauthorized)
	}

	return utils.SendData(nil, "Logout berhasil")
}
