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
	penggunaRepo   repository.PenggunaRepo   = repository.NewPenggunaRepo(dbSlave, dbMaster)
	regionRepo     repository.RegionRepo     = repository.NewRegionRepo(dbSlave, dbMaster)
	respondentRepo repository.RespondentRepo = repository.NewRespondentRepo(dbSlave, dbMaster)
	usersRepo      repository.UsersRepo      = repository.NewUsersRepo(dbSlave, dbMaster)
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
	)
	usersService service.UsersService = service.NewUsersService(
		usersRepo,
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
)

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error){
	"/userapi/healthy":   {"GET": handlers.Healthy},
	"/login":             {"POST": penggunaHandler.Login},
	"/logout":            {"POST": penggunaHandler.Logout},
	"/kecamatan/options": {"GET": regionHandler.KecamatanOptions},
	"/kelurahan/options": {"GET": regionHandler.KelurahansOptions},
	"/rw/options":        {"GET": regionHandler.RwOptions},
	"/rt/options":        {"GET": regionHandler.RtOptions},

	// Respondent Management
	"/respondent":         {"GET": respondentHandler.GetRespondent},
	"/respondent/:id":     {"GET": respondentHandler.GetDetailRespondent, "DELETE": respondentHandler.DeleteRespondent, "PUT": respondentHandler.UpdateRespondent},
	"/respondent/raw/:id": {"GET": respondentHandler.GetRawDetailRespondent},

	// Users Management
	"/users":              {"GET": usersHandler.GetUsers},
	"/reset/password/:id": {"PUT": usersHandler.ResetPassword},
	"/users/:id":          {"GET": usersHandler.GetDetailUsers, "PUT": usersHandler.UpdateUsers, "DELETE": usersHandler.DeleteUsers},
}

// Metode untuk menangani permintaan yang masuk
func (s *GRPCServer) SendData(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	path := req.GetPath()
	method := req.GetMethod()

	// Validasi token
	success, mess, code, userLogin, newToken := ValidasiToken(ctx, req)
	if !success {
		utils.LogErrors(mess)
		return utils.SetResponseData([]byte{}, success, mess, code, nil, ""), nil
	}

	// Cek path tersedia
	methodMap, ok := grpcMap[path]
	if !ok {
		message := "Path grpc tidak ditemukan"
		utils.LogErrors(message)
		return utils.SetResponseData([]byte{}, false, message, http.StatusNotFound, nil, ""), nil
	}

	// Cek method tersedia
	handler, ok := methodMap[method]
	if !ok {
		message := "Method grpc tidak ditemukan"
		utils.LogErrors(message)
		return utils.SetResponseData([]byte{}, false, message, http.StatusMethodNotAllowed, nil, ""), nil
	}

	// Parsing request body
	var reqs map[string]interface{}
	if len(req.Data) > 0 {
		if err := json.Unmarshal(req.Data, &reqs); err != nil {
			message := "Terjadi kesalahan saat unmarshal request data : " + err.Error()
			utils.LogErrors(message)
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}

	// Parsing slug
	var slug map[string]interface{}
	if len(req.Slug) > 0 {
		if err := json.Unmarshal(req.Slug, &slug); err != nil {
			message := "Terjadi kesalahan saat unmarshal slug : " + err.Error()
			utils.LogErrors(message)
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}

	// Parsing query param
	paramString := string(req.Param)
	queryValues, err := url.ParseQuery(paramString)
	if err != nil {
		message := "Terjadi kesalahan saat parsing query param : " + err.Error()
		utils.LogErrors(message)
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
			ID:    int64(claims.ID),
			Name:  claims.Name,
			Email: claims.Email,
			Role:  claims.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: claims.ExpiresAt,
			},
		}

		if claims.ExpiresAt != nil {
			remaining := claims.ExpiresAt.Time.Sub(time.Now())

			if remaining > 0 && remaining < 30*time.Minute {
				refreshedToken, err := GenerateJWTToken(userData)
				if err != nil {
					utils.LogErrors("Gagal generate token baru: " + err.Error())
				} else {
					newToken = refreshedToken
				}
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
		ID:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiredAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return encryptedToken, nil
}
