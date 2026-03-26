package routingGrpc

import (
	pb "backend/siccore/pb"
	"backend/surveyapi/configs"
	"backend/surveyapi/handlers"
	"backend/surveyapi/models"
	"backend/surveyapi/repository"
	"backend/surveyapi/service"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
	templateUcapanRepo repository.TemplateUcapanRepo = repository.NewTemplateUcapanRepo(dbSlave, dbMaster)
)

var (
	templateUcapanService service.TemplateUcapanService = service.NewTemplateUcapanService(
		templateUcapanRepo,
	)
)

var (
	templateUcapanHandler handlers.TemplateUcapanHandler = handlers.NewTemplateUcapanHandler(
		templateUcapanService,
	)
)

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error){
	"/surveyapi/healthy": {"GET": handlers.Healthy},

	"/template/ucapan/detail/:template_ucapan_id": {"GET": templateUcapanHandler.GetTemplateUcapanDetail},
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

	success, mess, code, userLogin := ValidasiToken(ctx, req)
	if !success {
		utils.LogErrors(mess)
		return utils.SetResponseData([]byte{}, success, mess, code, nil), nil
	}

	// Cek apakah handler sesuai dengan path dan method HTTP
	handler, ok := grpcMap[path][method]
	if !ok {
		message := "Terjadi kesalahan saat mencari grpc path method"
		utils.LogErrors(message)
		return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil), nil
	}

	// reqs
	var reqs map[string]interface{}
	if err := json.Unmarshal(req.Data, &reqs); err != nil {
		message := "Terjadi kesalahan saat unmarshal request data : " + err.Error()
		utils.LogErrors(message)
		return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil), nil
	}

	// slug
	var slug map[string]interface{}
	if err := json.Unmarshal(req.Slug, &slug); err != nil {
		message := "Terjadi kesalahan saat unmarshal slug : " + err.Error()
		utils.LogErrors(message)
		return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil), nil
	}

	// mengambil param dan meneruskan
	paramString := string(req.Param)
	queryValues, _ := url.ParseQuery(paramString)

	// Panggil handler yang sesuai
	return handler(ctx, reqs, userLogin, queryValues, slug)
}

func ValidasiToken(ctx context.Context, req *pb.ProxyRequest) (bool, string, int, models.JwtCustomClaims) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()
	var withToken bool = true
	var userData models.JwtCustomClaims

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		message := "Terjadi kesalahan saat unmarshal mengambil metadata"
		utils.LogErrors(message)
		return false, message, int(http.StatusInternalServerError), userData
	}

	// Ambil token dari metadata
	token := ""
	if val, ok := md["authorization"]; ok {
		// Harap header berbentuk "Bearer <token>"
		parts := strings.Split(val[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			withToken = false
			if req.GetIsSecure() {
				return false, "Format header token tidak valid", int(http.StatusUnauthorized), userData
			}
		}
		if withToken {
			token = parts[1]
		}
	} else {
		withToken = false
		if req.GetIsSecure() {
			return false, "Header token tidak ditemukan", int(http.StatusUnauthorized), userData
		}
	}

	// Parse token
	if withToken {
		claims := &models.JwtCustomClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
					utils.LogErrors(message)
				}
			}()
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil {
			withToken = false
			if req.GetIsSecure() {
				return false, "Token Tidak Valid", int(http.StatusUnauthorized), userData
			} else {
				return true, "Tervalidasi", int(http.StatusOK), userData
			}
		}

		// Dapatkan klaim dari token
		userData := models.JwtCustomClaims{
			ID:    int64(claims.ID),
			Name:  claims.Name,
			Email: claims.Email,
			Role:  claims.Role,
		}

		if err != nil {
			if req.GetIsSecure() {
				message := "Terjadi kesalahan : " + err.Error()
				utils.LogErrors(message)
				return false, message, int(http.StatusInternalServerError), userData
			}
		}

		return true, "Tervalidasi", int(http.StatusOK), userData

	}

	return true, "Tervalidasi", int(http.StatusOK), userData
}
