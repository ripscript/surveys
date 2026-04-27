package routingGrpc

import (
	pb "backend/siccore/pb"
	"backend/userapi/configs"
	"backend/userapi/handlers"
	"backend/userapi/models"
	"backend/userapi/repository"
	"backend/userapi/service"
	"backend/userapi/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type GRPCServer struct {
	pb.UnimplementedProxyServer
}

var (
	dbMaster *gorm.DB = configs.SetupDatabaseMasterConnection()
	dbSlave  *gorm.DB = configs.SetupDatabaseSlaveConnection()
)

var (
	penggunaRepo    repository.PenggunaRepo    = repository.NewPenggunaRepo(dbSlave, dbMaster)
	regionRepo      repository.RegionRepo      = repository.NewRegionRepo(dbSlave, dbMaster)
	respondentRepo  repository.RespondentRepo  = repository.NewRespondentRepo(dbSlave, dbMaster)
	usersRepo       repository.UsersRepo       = repository.NewUsersRepo(dbSlave, dbMaster)
	usersBlokirRepo repository.UsersBlokirRepo = repository.NewUsersBlokirRepo(dbSlave, dbMaster)
)

var (
	penggunaService service.PenggunaService = service.NewPenggunaService(
		penggunaRepo,
	)
	regionService service.RegionService = service.NewRegionService(
		regionRepo,
	)
	respondentService service.RespondentService = service.NewRespondentService(
		respondentRepo,
		usersRepo,
	)
	usersService service.UsersService = service.NewUsersService(
		usersRepo,
	)
	usersBlokirService service.UsersBlokirService = service.NewUsersBlokirService(
		usersBlokirRepo,
	)
)

var (
	penggunaHandler handlers.PenggunaHandler = handlers.NewPenggunaHandler(
		penggunaService,
	)
	regionHandler handlers.RegionHandler = handlers.NewRegionHandler(
		regionService,
	)
	respondentHandler handlers.RespondentHandler = handlers.NewRespondentHandler(
		respondentService,
	)
	usersHandler handlers.UsersHandler = handlers.NewUsersHandler(
		usersService,
	)
	usersBlokirHandler handlers.UsersBlokirHandler = handlers.NewUsersBlokirHandler(
		usersBlokirService,
	)
)

// && Key Menu && \\
// template
// manajemen-alur
// survey
// dashboard
// pengaturan
// laporan
// monitoring
// admin
// formulir-pertanyaan
// ucapan
// list-survey
// hasil
// manajemen-pengguna
// manajemen-wilayah
// manajemen-cms
// manajemen-artikel
// rating
// statistik
// aktifitas-survey
// profil-saya
// keluar
// management-responden
// user
// manajemen-blokir
// manajemen-wilayah-child
// manajemen-pejabat
// artikel
// promote
// kategori

// ROUTING GRPC
type RouteConfig struct {
	Handler func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error)
	MenuKey string
}

// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]RouteConfig{
	// Public / No Permission
	"/userapi/healthy":  {"GET": {Handler: handlers.Healthy, MenuKey: ""}},
	"/login":            {"POST": {Handler: penggunaHandler.Login, MenuKey: ""}},
	"/logout":           {"POST": {Handler: penggunaHandler.Logout, MenuKey: ""}},
	"/menus/permission": {"GET": {Handler: penggunaHandler.Menus, MenuKey: ""}},
	// Region
	"/kecamatan/options": {"GET": {Handler: regionHandler.KecamatanOptions, MenuKey: ""}},
	"/kelurahan/options": {"GET": {Handler: regionHandler.KelurahansOptions, MenuKey: ""}},
	"/rw/options":        {"GET": {Handler: regionHandler.RwOptions, MenuKey: ""}},
	"/rt/options":        {"GET": {Handler: regionHandler.RtOptions, MenuKey: ""}},
	// Respondent
	"/respondent":         {"GET": {Handler: respondentHandler.GetRespondent, MenuKey: "management-responden"}, "POST": {Handler: respondentHandler.CreateRespondent, MenuKey: "management-responden"}},
	"/respondent/import":  {"GET": {Handler: respondentHandler.GetExampleImport, MenuKey: "management-responden"}, "POST": {Handler: respondentHandler.ImportRespondent, MenuKey: "management-responden"}},
	"/respondent/:id":     {"GET": {Handler: respondentHandler.GetDetailRespondent, MenuKey: "management-responden"}, "PUT": {Handler: respondentHandler.UpdateRespondent, MenuKey: "management-responden"}, "DELETE": {Handler: respondentHandler.DeleteRespondent, MenuKey: "respondent"}},
	"/respondent/raw/:id": {"GET": {Handler: respondentHandler.GetRawDetailRespondent, MenuKey: "management-responden"}},
	// Users
	"/users":              {"GET": {Handler: usersHandler.GetUsers, MenuKey: "user"}, "POST": {Handler: usersHandler.CreateUsers, MenuKey: "user"}},
	"/users/export":       {"GET": {Handler: usersHandler.UserExport, MenuKey: "user"}},
	"/users/:id":          {"GET": {Handler: usersHandler.GetDetailUsers, MenuKey: "user"}, "PUT": {Handler: usersHandler.UpdateUsers, MenuKey: "user"}, "DELETE": {Handler: usersHandler.DeleteUsers, MenuKey: "user"}},
	"/reset/password/:id": {"PUT": {Handler: usersHandler.ResetPassword, MenuKey: "user"}},
	// Users Blokir
	"/users/blokir":     {"GET": {Handler: usersBlokirHandler.GetListdata, MenuKey: "user"}},
	"/users/blokir/:id": {"PUT": {Handler: usersBlokirHandler.OpenBlokir, MenuKey: "user"}},
	// Surveyor
	"/surveyor/options": {"GET": {Handler: respondentHandler.SurveyorOption, MenuKey: ""}},
}

func (s *GRPCServer) SendData(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	path := req.GetPath()
	method := req.GetMethod()

	success, mess, code, userLogin, newToken := ValidasiToken(ctx, req)
	if !success {
		utils.LogErrors(mess)
		return utils.SetResponseData([]byte{}, success, mess, code, nil, ""), nil
	}

	methodMap, ok := grpcMap[path]
	if !ok {
		message := "Path grpc tidak ditemukan"
		return utils.SetResponseData([]byte{}, false, message, http.StatusNotFound, nil, ""), nil
	}

	routeConfig, ok := methodMap[method]
	if !ok {
		message := "Method grpc tidak ditemukan"
		return utils.SetResponseData([]byte{}, false, message, http.StatusMethodNotAllowed, nil, ""), nil
	}

	handler := routeConfig.Handler
	menuKey := routeConfig.MenuKey

	if req.GetIsSecure() {
		allowed := CheckPermission(userLogin, menuKey, method)
		if !allowed {
			message := "Anda tidak memiliki hak akses"
			return utils.SetResponseData([]byte{}, false, message, http.StatusForbidden, nil, ""), nil
		}
	}

	var reqs map[string]interface{}
	if len(req.Data) > 0 {
		if err := json.Unmarshal(req.Data, &reqs); err != nil {
			message := "Terjadi kesalahan saat unmarshal request data : " + err.Error()
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}
	var slug map[string]interface{}
	if len(req.Slug) > 0 {
		if err := json.Unmarshal(req.Slug, &slug); err != nil {
			message := "Terjadi kesalahan saat unmarshal slug : " + err.Error()
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}

	paramString := string(req.Param)
	queryValues, err := url.ParseQuery(paramString)
	if err != nil {
		message := "Terjadi kesalahan saat parsing query param : " + err.Error()
		return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
	}
	response, err := handler(ctx, reqs, userLogin, queryValues, slug)
	if err != nil {
		return nil, err
	}
	if response != nil && newToken != "" {
		response.Token = newToken
	}

	return response, nil
}

func ValidasiToken(ctx context.Context, req *pb.ProxyRequest) (bool, string, int, models.JwtCustomClaims, string) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	var withToken bool = true
	var userData models.JwtCustomClaims
	var newToken string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		message := "Terjadi kesalahan saat mengambil metadata"
		utils.LogErrors(message)
		return false, message, int(http.StatusInternalServerError), userData, newToken
	}

	token := ""
	if val, ok := md["authorization"]; ok {
		parts := strings.Split(val[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			withToken = false
			if req.GetIsSecure() {
				return false, "Format header token tidak valid", int(http.StatusUnauthorized), userData, newToken
			}
		}

		if withToken {
			token = parts[1]
		}
	} else {
		withToken = false
		if req.GetIsSecure() {
			return false, "Header token tidak ditemukan", int(http.StatusUnauthorized), userData, newToken
		}
	}

	if withToken {
		claims := &models.JwtCustomClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil {
			if req.GetIsSecure() {
				return false, "Token Tidak Valid", int(http.StatusUnauthorized), userData, newToken
			}
			return true, "Tervalidasi", int(http.StatusOK), userData, newToken
		}

		userData = models.JwtCustomClaims{
			ID:           int64(claims.ID),
			RespondentID: claims.RespondentID,
			Name:         claims.Name,
			Email:        claims.Email,
			Role:         claims.Role,
			Permissions:  claims.Permissions,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: claims.ExpiresAt,
			},
		}

		if claims.ExpiresAt != nil {
			refreshedToken, err := GenerateJWTToken(userData)
			if err != nil {
				utils.LogErrors("Gagal generate token baru: " + err.Error())
			} else {
				newToken = refreshedToken
			}
		}

		return true, "Tervalidasi", int(http.StatusOK), userData, newToken
	}

	return true, "Tervalidasi", int(http.StatusOK), userData, newToken
}

func GenerateJWTToken(user models.JwtCustomClaims) (string, error) {
	defer utils.GeneralRecover()

	expiredAt := user.ExpiresAt

	if user.ExpiresAt != nil {
		remaining := user.ExpiresAt.Time.Sub(time.Now())
		if remaining < 30*time.Minute {
			expiredAt = jwt.NewNumericDate(time.Now().Add(3 * time.Hour))
		}
	} else {
		expiredAt = jwt.NewNumericDate(time.Now().Add(3 * time.Hour))
	}

	claims := models.JwtCustomClaims{
		RespondentID: user.RespondentID,
		ID:           int64(user.ID),
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Permissions:  user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiredAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", err
	}
	return encryptedToken, nil
}

func NormalizePath(path string) string {
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllString(path, "/:id")
}

func methodToAction(method string) string {
	switch method {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}
func CheckPermission(claims models.JwtCustomClaims, menuKey string, method string) bool {
	action := methodToAction(method)
	if menuKey == "" {
		return true
	}
	perm, ok := claims.Permissions[menuKey]
	if !ok {
		return false
	}

	switch action {
	case "view":
		return perm.V
	case "create":
		return perm.C
	case "update":
		return perm.U
	case "delete":
		return perm.D
	}

	return false
}
