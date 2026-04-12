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
	templateUcapanRepo             repository.TemplateUcapanRepo             = repository.NewTemplateUcapanRepo(dbSlave, dbMaster)
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo = repository.NewTemplateFormulirPertanyaanRepo(dbSlave, dbMaster)
	manajemenAlurRepo              repository.ManajemenAlurRepo              = repository.NewManajemenAlurRepo(dbSlave, dbMaster)
)

var (
	templateUcapanService service.TemplateUcapanService = service.NewTemplateUcapanService(
		templateUcapanRepo,
	)
	templateFormulirPertanyaanService service.TemplateFormulirPertanyaanService = service.NewTemplateFormulirPertanyaanService(
		templateFormulirPertanyaanRepo,
	)
	manajemenAlurService service.ManajemenAlurService = service.NewManajemenAlurService(
		manajemenAlurRepo,
		templateFormulirPertanyaanRepo,
		templateUcapanRepo,
	)
)

var (
	templateUcapanHandler handlers.TemplateUcapanHandler = handlers.NewTemplateUcapanHandler(
		templateUcapanService,
	)
	templateFormulirPertanyaanHandler handlers.TemplateFormulirPertanyaanHandler = handlers.NewTemplateFormulirPertanyaanHandler(
		templateFormulirPertanyaanService,
	)
	manajemenAlurHandler handlers.ManajemenAlurHandler = handlers.NewManajemenAlurHandler(
		manajemenAlurService,
	)
)

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error){
	"/surveyapi/healthy": {"GET": handlers.Healthy},

	"/template/ucapan/detail/:template_ucapan_id": {"GET": templateUcapanHandler.GetTemplateUcapanDetail},
	"/template/ucapan/variable-options":           {"GET": templateUcapanHandler.OptionsVariableTemplateUcapan},
	"/template/ucapan/create":                     {"POST": templateUcapanHandler.CreateTemplateUcapan},
	"/template/ucapan/update/:template_ucapan_id": {"PUT": templateUcapanHandler.UpdateTemplateUcapan},
	"/template/ucapan/delete/:template_ucapan_id": {"DELETE": templateUcapanHandler.DeleteTemplateUcapan},
	"/template/ucapan/list":                       {"GET": templateUcapanHandler.GetListTemplateUcapan},
	"/template/ucapan/options":                    {"GET": templateUcapanHandler.GetUcapanOptions},

	"/template/formulir-pertanyaan/create":                                       {"POST": templateFormulirPertanyaanHandler.CreateTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/detail/:template_formulir_pertanyaan_code":    {"GET": templateFormulirPertanyaanHandler.GetDetailTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/update/:template_formulir_pertanyaan_code":    {"PUT": templateFormulirPertanyaanHandler.UpdateTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/duplicate/:template_formulir_pertanyaan_code": {"POST": templateFormulirPertanyaanHandler.DuplicateTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/delete/:template_formulir_pertanyaan_code":    {"DELETE": templateFormulirPertanyaanHandler.DeleteTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/list":                                         {"GET": templateFormulirPertanyaanHandler.GetListTemplateFormulirPertanyaan},
	"/template/formulir-pertanyaan/question-type-options":                        {"GET": templateFormulirPertanyaanHandler.GetQuestionTypeOptions},
	"/template/formulir-pertanyaan/options":                                      {"GET": templateFormulirPertanyaanHandler.GetFormulirPertanyaanOptions},
	"/template/formulir-pertanyaan/pertanyaan-options/:form_code":                {"GET": templateFormulirPertanyaanHandler.GetPertanyaanOptions},
	"/template/formulir-pertanyaan/detail-pertanyaan/:id":                        {"GET": templateFormulirPertanyaanHandler.DetailPertanyaan},
	"/template/formulir-pertanyaan/multiple-choice-options/:form_field_id":       {"GET": templateFormulirPertanyaanHandler.GetMultipleChoiceOptionByFormFieldId},

	"/manajemen-alur/create":                                {"POST": manajemenAlurHandler.CreateManajemenAlur},
	"/manajemen-alur/detail/:code":                          {"GET": manajemenAlurHandler.GetDetailManajemenAlur},
	"/manajemen-alur/update/:code":                          {"PUT": manajemenAlurHandler.UpdateManajemenAlur},
	"/manajemen-alur/delete/:code":                          {"DELETE": manajemenAlurHandler.DeleteManajemenAlur},
	"/manajemen-alur/list":                                  {"GET": manajemenAlurHandler.GetListManajemenAlur},
	"/manajemen-alur/preview-index/:code":                   {"GET": manajemenAlurHandler.FlowPreviewIndex},
	"/manajemen-alur/preview-alur/:flow_code/:section_code": {"GET": manajemenAlurHandler.PreviewAlurSurvey},
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
