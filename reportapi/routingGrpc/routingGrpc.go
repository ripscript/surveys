package routingGrpc

import (
	"backend/reportapi/configs"
	"backend/reportapi/handlers"
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/reportapi/service"
	"backend/reportapi/utils"
	pb "backend/siccore/pb"
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
	homeRepo         repository.HomeRepo         = repository.NewHomeRepo(dbSlave, dbMaster)
	logRepo          repository.LogRepo          = repository.NewLogRepo(dbSlave, dbMaster)
	surveyRepo       repository.SurveyRepo       = repository.NewSurveyRepo(dbSlave, dbMaster)
	surveyExportRepo repository.SurveyExportRepo = repository.NewSurveyExportRepo(dbSlave, dbMaster)
	statistikRepo    repository.StatistikRepo    = repository.NewStatistikRepo(dbSlave, dbMaster)
	userRepo         repository.UserRepo         = repository.NewUserRepo(dbSlave, dbMaster)
	laporanRepo      repository.LaporanRepo      = repository.NewLaporanRepo(dbSlave, dbMaster)
	fileRepo         repository.FileRepo         = repository.NewFileRepo(dbSlave, dbMaster)
)

var (
	homeService service.HomeService = service.NewHomeService(
		homeRepo,
	)
	logService service.LogService = service.NewLogService(
		logRepo,
	)
	surveyService service.SurveyService = service.NewSurveyService(
		surveyRepo,
		surveyExportRepo,
	)
	statistikService service.StatistikService = service.NewStatistikService(
		statistikRepo,
		userRepo,
	)
	laporanService service.LaporanService = service.NewLaporanService(
		laporanRepo,
		fileRepo,
	)
)

var (
	homeHandler handlers.HomeHandler = handlers.NewHomeHandler(
		homeService,
	)
	logHandler handlers.LogHandler = handlers.NewLogHandler(
		logService,
	)
	surveyHandler handlers.SurveyHandler = handlers.NewSurveyHandler(
		surveyService,
	)
	statistikHandler handlers.StatistikHandler = handlers.NewStatistikHandler(
		statistikService,
	)
	laporanHandler handlers.LaporanHandler = handlers.NewLaporanHandler(
		laporanService,
	)
)

// && Key Menu && \\
// template
// management-alur
// survey
// dashboard
// pengaturan
// laporan
// monitoring
// admin
// formulir-pertanyaan
// ucapan
// master-data
// hasil
// management-pengguna
// management-wilayah
// management-cms
// management-artikel
// rating
// statistik
// aktifitas-survey
// profil-saya
// keluar
// management-responden
// user
// management-blokir
// management-wilayah-child
// management-pejabat
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
	"/reportapi/healthy":        {"GET": {Handler: handlers.Healthy, MenuKey: ""}},
	"/count/kecamatan":          {"GET": {Handler: homeHandler.CountKecamatan, MenuKey: ""}},
	"/count/kelurahan":          {"GET": {Handler: homeHandler.CountKelurahan, MenuKey: ""}},
	"/count/rw":                 {"GET": {Handler: homeHandler.CountRw, MenuKey: ""}},
	"/count/rt":                 {"GET": {Handler: homeHandler.CountRt, MenuKey: ""}},
	"/count/survey/ongoing":     {"GET": {Handler: homeHandler.CountSurveyOngoing, MenuKey: ""}},
	"/count/survey/upcoming":    {"GET": {Handler: homeHandler.CountSurveyUpcoming, MenuKey: ""}},
	"/count/survey/finished":    {"GET": {Handler: homeHandler.CountSurveyFinished, MenuKey: ""}},
	"/save/log/activities":      {"POST": {Handler: logHandler.SaveLogActivities, MenuKey: ""}},
	"/log/activities":           {"GET": {Handler: logHandler.GetLogActivities, MenuKey: ""}},
	"/survey/activities":        {"GET": {Handler: surveyHandler.SurveyActivities, MenuKey: ""}},
	"/survey/activities/export": {"GET": {Handler: surveyHandler.SurveyActvitiesExport, MenuKey: ""}},
	"/log/surveys":              {"GET": {Handler: surveyHandler.LogSurveys, MenuKey: ""}, "POST": {Handler: surveyHandler.InsertLogSurveys, MenuKey: ""}},

	// BEGIN:MONITORING & LAPORAN ==================================
	// BEGIN:STATISTIK ==================================
	"/monitoring-dan-laporan/statistik": {
		"GET": {
			Handler: statistikHandler.GetStatistikIndex,
			MenuKey: "statistik",
		},
	},
	"/monitoring-dan-laporan/statistik/survey/:survey_code": {
		"GET": {
			Handler: statistikHandler.GetStatistik,
			MenuKey: "statistik",
		},
	},

	"/monitoring-dan-laporan/statistik/survey/:survey_code/export-excel": {
		"GET": {
			Handler: statistikHandler.ExportExcelStatistik,
			MenuKey: "statistik",
		},
	},
	// END:STATISTIK ==================================

	// BEGIN:LAPORAN ==================================
	"/monitoring-dan-laporan/laporan/list": {
		"GET": {
			Handler: laporanHandler.ListLaporan,
			MenuKey: "laporan",
		},
	},
	"/monitoring-dan-laporan/laporan/create": {
		"POST": {
			Handler: laporanHandler.CreateReport,
			MenuKey: "laporan",
		},
	},
	"/monitoring-dan-laporan/laporan/change-name/:laporan_id": {
		"PATCH": {
			Handler: laporanHandler.ChangeNameReport,
			MenuKey: "laporan",
		},
	},
	"/monitoring-dan-laporan/laporan/update-cover/:laporan_id": {
		"PUT": {
			Handler: laporanHandler.UpdateCoverReport,
			MenuKey: "laporan",
		},
	},
	"/monitoring-dan-laporan/laporan/get-cover/:laporan_id": {
		"GET": {
			Handler: laporanHandler.GetCoverReport,
			MenuKey: "laporan",
		},
	},
	"/monitoring-dan-laporan/laporan/cetak/:laporan_id": {
		"GET": {
			Handler: laporanHandler.PrintReport,
			MenuKey: "laporan",
		},
	},
	// END:LAPORAN ==================================
	// END:MONITORING & LAPORAN ==================================
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
