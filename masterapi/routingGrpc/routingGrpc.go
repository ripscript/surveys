package routingGrpc

import (
	"backend/masterapi/configs"
	"backend/masterapi/handlers"
	"backend/masterapi/models"
	"backend/masterapi/repository"
	"backend/masterapi/service"
	"backend/masterapi/utils"
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
	manajemenWilayahRepo  repository.ManajemenWilayahRepo  = repository.NewManajemenWilayahRepo(dbSlave, dbMaster)
	manajemenPenggunaRepo repository.ManajemenPenggunaRepo = repository.NewManajemenPenggunaRepo(dbSlave, dbMaster)
	manajemenArtikelRepo  repository.ManajemenArtikelRepo  = repository.NewManajemenArtikelRepo(dbSlave, dbMaster)
	manajemenPejabatRepo  repository.ManajemenPejabatRepo  = repository.NewManajemenPejabatRepo(dbSlave, dbMaster)
)

var (
	manajemenWilayahService service.ManajemenWilayahService = service.NewManajemenWilayahService(
		manajemenWilayahRepo,
	)
	manajemenPenggunaService service.ManajemenPenggunaService = service.NewManajemenPenggunaService(
		manajemenPenggunaRepo,
	)
	manajemenArtikelService service.ManajemenArtikelService = service.NewManajemenArtikelService(
		manajemenArtikelRepo,
	)
	manajemenPejabatService service.ManajemenPejabatService = service.NewManajemenPejabatService(
		manajemenPejabatRepo,
	)
)

var (
	manajemenWilayahHandler handlers.ManajemenWilayahHandler = handlers.NewManajemenWilayahHandler(
		manajemenWilayahService,
	)
	manajemenPenggunaHandler handlers.ManajemenPenggunaHandler = handlers.NewManajemenPenggunaHandler(
		manajemenPenggunaService,
	)
	manajemenArtikelHandler handlers.ManajemenArtikelHandler = handlers.NewManajemenArtikelHandler(
		manajemenArtikelService,
	)
	manajemenPejabatHandler handlers.ManajemenPejabatHandler = handlers.NewManajemenPejabatHandler(
		manajemenPejabatService,
	)
)

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error){
	"/masterapi/healthy": {"GET": handlers.Healthy},

	"/manajemen-wilayah/kecamatan/create":               {"POST": manajemenWilayahHandler.CreateKecamatan},
	"/manajemen-wilayah/kecamatan/list":                 {"GET": manajemenWilayahHandler.GetListKecamatan},
	"/manajemen-wilayah/kecamatan/detail/:kecamatan_id": {"GET": manajemenWilayahHandler.GetKecamatanDetail},
	"/manajemen-wilayah/kecamatan/update/:kecamatan_id": {"PUT": manajemenWilayahHandler.UpdateKecamatan},
	"/manajemen-wilayah/kecamatan/options":              {"GET": manajemenWilayahHandler.OptionsKecamatan},
	"/manajemen-wilayah/kecamatan/delete/:kecamatan_id": {"DELETE": manajemenWilayahHandler.DeleteKecamatan},

	"/manajemen-wilayah/kelurahan/create":               {"POST": manajemenWilayahHandler.CreateKelurahan},
	"/manajemen-wilayah/kelurahan/detail/:kelurahan_id": {"GET": manajemenWilayahHandler.GetKelurahanDetail},
	"/manajemen-wilayah/kelurahan/update/:kelurahan_id": {"PUT": manajemenWilayahHandler.UpdateKelurahan},
	"/manajemen-wilayah/kelurahan/list":                 {"GET": manajemenWilayahHandler.GetListKelurahan},
	"/manajemen-wilayah/kelurahan/list/:kecamatan_id":   {"GET": manajemenWilayahHandler.GetListKelurahan},
	"/manajemen-wilayah/kelurahan/options":              {"GET": manajemenWilayahHandler.OptionsKelurahan},
	"/manajemen-wilayah/kelurahan/delete/:kelurahan_id": {"DELETE": manajemenWilayahHandler.DeleteKelurahan},

	"/manajemen-wilayah/rw/detail/:rw_id":      {"GET": manajemenWilayahHandler.GetRwDetail},
	"/manajemen-wilayah/rw/update/:rw_id":      {"PUT": manajemenWilayahHandler.UpdateRw},
	"/manajemen-wilayah/rw/list":               {"GET": manajemenWilayahHandler.GetListRw},
	"/manajemen-wilayah/rw/list/:kelurahan_id": {"GET": manajemenWilayahHandler.GetListRwByKelurahan},
	"/manajemen-wilayah/rw/create":             {"POST": manajemenWilayahHandler.CreateRw},
	"/manajemen-wilayah/rw/options":            {"GET": manajemenWilayahHandler.OptionsRw},
	"/manajemen-wilayah/rw/delete/:rw_id":      {"DELETE": manajemenWilayahHandler.DeleteRw},

	"/manajemen-wilayah/rt/detail/:rt_id": {"GET": manajemenWilayahHandler.GetRtDetail},
	"/manajemen-wilayah/rt/update/:rt_id": {"PUT": manajemenWilayahHandler.UpdateRt},
	"/manajemen-wilayah/rt/list":          {"GET": manajemenWilayahHandler.GetListRt},
	"/manajemen-wilayah/rt/list/:rw_id":   {"GET": manajemenWilayahHandler.GetListRtByRw},
	"/manajemen-wilayah/rt/create":        {"POST": manajemenWilayahHandler.CreateRt},
	"/manajemen-wilayah/rt/options":       {"GET": manajemenWilayahHandler.OptionsRt},
	"/manajemen-wilayah/rt/delete/:rt_id": {"DELETE": manajemenWilayahHandler.DeleteRT},

	"/manajemen-pejabat/create":     {"POST": manajemenPejabatHandler.CreatePejabat},
	"/manajemen-pejabat/detail/:id": {"GET": manajemenPejabatHandler.DetailPejabat},
	"/manajemen-pejabat/update/:id": {"PUT": manajemenPejabatHandler.UpdatePejabat},
	"/manajemen-pejabat/delete/:id": {"DELETE": manajemenPejabatHandler.DeletePejabat},
	"/manajemen-pejabat/list":       {"GET": manajemenPejabatHandler.GetListPejabat},

	"/manajemen-pengguna/role/options":     {"GET": manajemenPenggunaHandler.OptionsRole},
	"/manajemen-pengguna/responden/create": {"POST": manajemenPenggunaHandler.CreateResponden},

	"/manajemen-artikel/kategori/create":     {"POST": manajemenArtikelHandler.CreateKategoriArtikel},
	"/manajemen-artikel/kategori/update/:id": {"PUT": manajemenArtikelHandler.UpdateKategoriArtikel},
	"/manajemen-artikel/kategori/delete/:id": {"DELETE": manajemenArtikelHandler.DeleteKategoriArtikel},
	"/manajemen-artikel/kategori/detail/:id": {"GET": manajemenArtikelHandler.GetKategoriArtikel},
	"/manajemen-artikel/kategori/list":       {"GET": manajemenArtikelHandler.GetListKategoriArtikel},
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
	// if req.GetIsSecure() {
	// 	normalizedPath := NormalizePath(path)
	// 	allowed := CheckPermission(int(userLogin.Role), normalizedPath, method)
	// 	if !allowed {
	// 		message := "Anda tidak memiliki hak akses"
	// 		utils.LogErrors(message)
	// 		return utils.SetResponseData([]byte{}, false, message, http.StatusForbidden, nil, ""), nil
	// 	}
	// }

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

func NormalizePath(path string) string {
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllString(path, "/:id")
}

func CheckPermission(roleID int, path string, method string) bool {
	var count int64

	dbSlave.Table("menu_permissions mp").
		Joins("JOIN menus m ON m.id = mp.menu_id").
		Where("mp.role_id = ?", roleID).
		Where("m.endpoint = ?", path).
		Where("m.method = ?", method).
		Count(&count)

	return count > 0
}
