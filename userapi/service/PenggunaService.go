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
	"gorm.io/gorm"
)

type PenggunaService interface {
	Login(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	Logout(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	Menus(usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	LoginV2(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
}

type penggunaService struct {
	penggunaRepo repository.PenggunaRepo
	usersRepo    repository.UsersRepo

	respondentRepo repository.RespondentRepo
	attemptRepo    repository.LoginAttemptRepository
	permissionRepo repository.PermissionRepository
}

func NewPenggunaService(
	penggunaRepo repository.PenggunaRepo,
	usersRepo repository.UsersRepo,
	respondentRepo repository.RespondentRepo,
	attemptRepo repository.LoginAttemptRepository,
	permissionRepo repository.PermissionRepository,

) PenggunaService {
	return &penggunaService{
		penggunaRepo:   penggunaRepo,
		usersRepo:      usersRepo,
		respondentRepo: respondentRepo,
		attemptRepo:    attemptRepo,
		permissionRepo: permissionRepo,
	}
}

const maxFailedLoginAttempts = 3

func (service *penggunaService) LoginV2(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.LoginPayload
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Password == "" {
		return utils.SendError(errors.New("password wajib diisi"), http.StatusBadRequest)
	}

	respondent, err := service.respondentRepo.FindByRole(payload)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRespondentNotFound):
			return utils.SendError(errors.New("NIK, email atau password salah"), http.StatusUnauthorized)
		default:
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	storedUser, err := service.usersRepo.FindByRespondentID(respondent.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}
	if storedUser == nil {
		return utils.SendError(errors.New("NIK, email atau password salah"), http.StatusUnauthorized)
	}

	if err := service.checkNotBlocked(storedUser.ID, respondent.ID); err != nil {
		return utils.SendError(err, http.StatusUnauthorized)
	}

	if !utils.VerifyPassword(*storedUser.Password, payload.Password) {
		if *storedUser.Email != "admin@gmail.com" {
			if err := service.attemptRepo.RecordFailed(int(storedUser.ID)); err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
		}
		return utils.SendError(errors.New("NIK, email atau password salah"), http.StatusUnauthorized)
	}

	if err := service.attemptRepo.ResetFailed(int(storedUser.ID)); err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	requestEmail := resolveIdentity(respondent)

	permMap, err := service.buildPermissionMap(respondent.RoleID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	encryptedToken, err := generateToken(storedUser.ID, respondent, requestEmail, permMap)
	if err != nil {
		return utils.SendError(errors.New("gagal membuat token"), http.StatusBadRequest)
	}

	if err := service.usersRepo.UpdateLastLogin(int(storedUser.ID), time.Now()); err != nil {
		return utils.SendError(errors.New("gagal memperbarui data user"), http.StatusInternalServerError)
	}

	return utils.SendData(encryptedToken, "Login berhasil")
}

// checkNotBlocked mengembalikan error jika user sudah melebihi batas percobaan gagal.
// Jika baru melewati batas pada pemanggilan ini, respondent otomatis di-block.
func (service *penggunaService) checkNotBlocked(userID int, respondentID int) error {
	failedCount, err := service.attemptRepo.CountFailed(int(userID))
	if err != nil {
		return err
	}
	if failedCount >= maxFailedLoginAttempts {
		_ = service.respondentRepo.Block(respondentID)
		return errors.New("akun diblokir, terlalu banyak percobaan gagal")
	}
	return nil
}

func resolveIdentity(respondent *models.Respondent) string {
	switch {
	case respondent.Email != "":
		return respondent.Email
	case respondent.Username != "":
		return respondent.Username
	default:
		return respondent.NIK
	}
}

func (service *penggunaService) buildPermissionMap(roleID int) (map[string]models.PermissionAction, error) {
	menuPermission, err := service.permissionRepo.GetMenuPermission(roleID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[string]models.PermissionAction, len(menuPermission))
	for _, p := range menuPermission {
		if !p.ViewAction {
			continue // tidak punya akses view -> tidak perlu dikirim ke response
		}
		permMap[p.Menu.Key] = models.PermissionAction{
			V: p.ViewAction,
			C: p.CreateAction,
			U: p.UpdateAction,
			D: p.DeleteAction,
		}
	}
	return permMap, nil
}

func generateToken(userID int, respondent *models.Respondent, email string, permMap map[string]models.PermissionAction) (string, error) {
	claims := &models.JwtCustomClaims{
		ID:           int64(userID),
		RespondentID: int64(respondent.ID),
		Name:         respondent.Name,
		Email:        email,
		Role:         respondent.RoleID,
		Permissions:  permMap,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
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

	// FindRespondentByRole
	// GetRespondentHasJabatanActive
	respondent, requestEmail, err := service.penggunaRepo.FindRespondentByRole(payload)
	if err != nil {
		if errors.Is(err, repository.ErrRespondentNotFound) {
			return utils.SendError(errors.New("data responden tidak ditemukan"), http.StatusUnauthorized)
		} else if errors.Is(err, repository.ErrJabatanNotActive) {
			return utils.SendError(errors.New("jabatan tidak aktif"), http.StatusUnauthorized)
		} else {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	if respondent == nil {
		return utils.SendError(errors.New("data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	storedUser, err := service.usersRepo.GetUserByRespondentId(respondent.ID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	if storedUser == nil {
		return utils.SendError(errors.New("user tidak ditemukan"), http.StatusUnauthorized)
	}

	// storedUser, err := service.penggunaRepo.FindUserByRespondentID(respondent.ID)

	loginCount, err := service.penggunaRepo.CountFailedLogin(int(storedUser.ID))

	if err == nil && loginCount >= 3 {
		service.penggunaRepo.BlockRespondent(respondent.ID)
		return utils.SendError(errors.New("akun diblokir"), http.StatusUnauthorized)
	}

	isPasswordValid := utils.VerifyPassword(*storedUser.Password, payload.Password)
	if !isPasswordValid {
		if *storedUser.Email != "admin@gmail.com" {
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

	menuPermission, err := service.penggunaRepo.GetMenuPermission(respondent.RoleID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	permMap := make(map[string]models.PermissionAction)

	for _, p := range menuPermission {
		permMap[p.Menu.Key] = models.PermissionAction{
			V: p.ViewAction,
			C: p.CreateAction,
			U: p.UpdateAction,
			D: p.DeleteAction,
		}
	}

	claims := &models.JwtCustomClaims{
		ID:           int64(storedUser.ID),
		RespondentID: int64(respondent.ID),
		Name:         respondent.Name,
		Email:        requestEmail,
		Role:         respondent.RoleID,
		Permissions:  permMap,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return utils.SendError(errors.New("gagal membuat token"), http.StatusBadRequest)
	}

	storedUser.LastLogin = utils.TimeNowPointer()
	_, err = service.usersRepo.UpdateUser(storedUser)
	if err != nil {
		return utils.SendError(errors.New("gagal memperbarui data user"), http.StatusInternalServerError)
	}
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

func (service *penggunaService) Menus(usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	roleIdString := usr.Role

	roleId, err := utils.ToInt64(roleIdString)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	listMenus, err := service.penggunaRepo.ListMenus(roleId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(listMenus)
}
