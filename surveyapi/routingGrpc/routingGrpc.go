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
	surveyRepo                     repository.SurveyRepo                     = repository.NewSurveyRepo(dbSlave, dbMaster)
	wilayahRepo                    repository.WilayahRepo                    = repository.NewWilayahRepo(dbSlave, dbMaster)
	userRepo                       repository.UserRepo                       = repository.NewUserRepo(dbSlave, dbMaster)
	fileRepo                       repository.FileRepo                       = repository.NewFileRepo(dbSlave, dbMaster)
)

var (
	templateUcapanService service.TemplateUcapanService = service.NewTemplateUcapanService(
		templateUcapanRepo,
		userRepo,
	)
	templateFormulirPertanyaanService service.TemplateFormulirPertanyaanService = service.NewTemplateFormulirPertanyaanService(
		templateFormulirPertanyaanRepo,
	)
	manajemenAlurService service.ManajemenAlurService = service.NewManajemenAlurService(
		manajemenAlurRepo,
		templateFormulirPertanyaanRepo,
		templateUcapanRepo,
	)
	surveyService service.SurveyService = service.NewSurveyService(
		manajemenAlurRepo,
		templateFormulirPertanyaanRepo,
		templateUcapanRepo,
		surveyRepo,
		wilayahRepo,
		userRepo,
		fileRepo,
	)
	hasilSurveyService service.HasilSurveyService = service.NewHasilSurveyService(
		manajemenAlurRepo,
		templateFormulirPertanyaanRepo,
		templateUcapanRepo,
		surveyRepo,
		wilayahRepo,
		userRepo,
		fileRepo,
	)
	seedAnswerService service.SeedAnswerService = service.NewSeedAnswerService(
		surveyRepo,
		manajemenAlurRepo,
		surveyService,
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
	surveyHandler handlers.SurveyHandler = handlers.NewSurveyHandler(
		manajemenAlurService,
		surveyService,
	)
	hasilSurveyHandler handlers.HasilSurveyHandler = handlers.NewHasilSurveyHandler(
		manajemenAlurService,
		hasilSurveyService,
		surveyService,
	)
	seedAnswerHandler handlers.SeedAnswerHandler = handlers.NewSeedAnswerHandler(
		seedAnswerService,
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
	"/surveyapi/healthy": {"GET": {
		Handler: handlers.Healthy,
		MenuKey: "",
	}},

	"/template/ucapan/detail/:template_ucapan_id": {"GET": {
		Handler: templateUcapanHandler.GetTemplateUcapanDetail,
		MenuKey: "template-ucapan",
	}},

	"/template/ucapan/variable-options": {"GET": {
		Handler: templateUcapanHandler.OptionsVariableTemplateUcapan,
		MenuKey: "template-ucapan",
	}},
	"/template/ucapan/create": {"POST": {
		Handler: templateUcapanHandler.CreateTemplateUcapan,
		MenuKey: "template-ucapan",
	}},
	"/template/ucapan/update/:template_ucapan_id": {"PUT": {
		Handler: templateUcapanHandler.UpdateTemplateUcapan,
		MenuKey: "template-ucapan",
	}},
	"/template/ucapan/delete/:template_ucapan_id": {"DELETE": {
		Handler: templateUcapanHandler.DeleteTemplateUcapan,
		MenuKey: "template-ucapan",
	}},
	"/template/ucapan/list": {"GET": {
		Handler: templateUcapanHandler.GetListTemplateUcapan,
		MenuKey: "template-ucapan",
	}},
	"/template/ucapan/options": {"GET": {
		Handler: templateUcapanHandler.GetUcapanOptions,
		MenuKey: "template-ucapan",
	}},

	"/template/formulir-pertanyaan/create": {"POST": {
		Handler: templateFormulirPertanyaanHandler.CreateTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/detail/:template_formulir_pertanyaan_code": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetDetailTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/update/:template_formulir_pertanyaan_code": {"PUT": {
		Handler: templateFormulirPertanyaanHandler.UpdateTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/duplicate/:template_formulir_pertanyaan_code": {"POST": {
		Handler: templateFormulirPertanyaanHandler.DuplicateTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/delete/:template_formulir_pertanyaan_code": {"DELETE": {
		Handler: templateFormulirPertanyaanHandler.DeleteTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/list": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetListTemplateFormulirPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/question-type-options": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetQuestionTypeOptions,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/options": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetFormulirPertanyaanOptions,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/pertanyaan-options/:form_code": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetPertanyaanOptions,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/detail-pertanyaan/:id": {"GET": {
		Handler: templateFormulirPertanyaanHandler.DetailPertanyaan,
		MenuKey: "template-pertanyaan",
	}},
	"/template/formulir-pertanyaan/multiple-choice-options/:form_field_id": {"GET": {
		Handler: templateFormulirPertanyaanHandler.GetMultipleChoiceOptionByFormFieldId,
		MenuKey: "template-pertanyaan",
	}},

	"/manajemen-alur/create": {"POST": {
		Handler: manajemenAlurHandler.CreateManajemenAlur,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/detail/:code": {"GET": {
		Handler: manajemenAlurHandler.GetDetailManajemenAlur,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/update/:code": {"PUT": {
		Handler: manajemenAlurHandler.UpdateManajemenAlur,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/delete/:code": {"DELETE": {
		Handler: manajemenAlurHandler.DeleteManajemenAlur,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/list": {"GET": {
		Handler: manajemenAlurHandler.GetListManajemenAlur,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/preview-index/:code": {"GET": {
		Handler: manajemenAlurHandler.FlowPreviewIndex,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/preview-alur/:flow_code/:section_code": {"GET": {
		Handler: manajemenAlurHandler.PreviewAlurSurvey,
		MenuKey: "management-alur",
	}},
	"/manajemen-alur/options": {"GET": {
		Handler: manajemenAlurHandler.AlurOptions,
		MenuKey: "management-alur",
	}},

	// SURVEY ===========================================
	"/survey/create": {"POST": {
		Handler: surveyHandler.CreateSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/detail/:survey_code": {"GET": {
		Handler: surveyHandler.DetailSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/periode-options": {"GET": {
		Handler: surveyHandler.OptionsPeriodeSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/list": {"GET": {
		Handler: surveyHandler.GetListSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/approval/:code": {"POST": {
		Handler: surveyHandler.ApprovalSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/history-approval/survey/:survey_code": {"GET": {
		Handler: surveyHandler.GetHistoryApprovalSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/show-image/:survey_code/:code_wilayah/:path": {"GET": {
		Handler: surveyHandler.GetPublicImageSurvey,
		MenuKey: "pengelola-survey",
	}},
	"/survey/options": {"GET": {
		Handler: surveyHandler.SurveyOptions,
		MenuKey: "pengelola-survey",
	}},
	"/survey/question-options/:survey_id": {"GET": {
		Handler: surveyHandler.SurveyQuestionOptions,
		MenuKey: "pengelola-survey",
	}},
	"/survey/public-options": {"GET": {
		Handler: surveyHandler.PublicSurveyOptions,
		MenuKey: "",
	}},
	"/survey/public-question-options": {"GET": {
		Handler: surveyHandler.PublicSurveyQuestionOptions,
		MenuKey: "",
	}},
	"/survey/heat-point-wilayah-kota-bandung": {"POST": {
		Handler: surveyHandler.HeatPointWilayahKotaBandung,
		MenuKey: "",
	}},

	// SURVEY WILAYAH ===========================================
	"/survey-wilayah/preview-index/:code": {"GET": {
		Handler: surveyHandler.PreviewSurveyIndex,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-wilayah/preview/:survey_code/:section_code": {"GET": {
		Handler: surveyHandler.PreviewSurvey,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-wilayah/:survey_code/section/:section_code/submit": {"POST": {
		Handler: surveyHandler.SurveyBundlingSubmit,
		MenuKey: "survey-kewilayahan",
	}},

	"/survey-wilayah/list": {"GET": {
		Handler: surveyHandler.AvailableSurveyWilayah,
		MenuKey: "survey-kewilayahan",
	}},

	"/survey-wilayah/update-status/:survey_code": {"PUT": {
		Handler: surveyHandler.UpdateSurveyRespondentStatus,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-wilayah/action-required/count": {"GET": {
		Handler: surveyHandler.ActionRequiredCount,
		MenuKey: "survey-kewilayahan",
	}},

	"/survey-kewilayahan/list": {"GET": {
		Handler: surveyHandler.GetSurveyKewilayahan,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/detail/:survey_code": {"GET": {
		Handler: surveyHandler.GetDetailSurveyKewilayahan,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan-aktifitas/list": {"GET": {
		Handler: surveyHandler.GetSurveyKewilayahan,
		MenuKey: "aktifitas-survey",
	}},
	"/survey-kewilayahan-aktifitas/detail/:survey_code": {"GET": {
		Handler: surveyHandler.GetDetailSurveyKewilayahan,
		MenuKey: "aktifitas-survey",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code": {"GET": {
		Handler: surveyHandler.SurveyResultIndex,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code/section/:section_code": {"GET": {
		Handler: surveyHandler.SurveyResultSectionDetail,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code/approve-verify": {"PUT": {
		Handler: surveyHandler.VerifySurveyAnswers,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code/reject-verify": {"PUT": {
		Handler: surveyHandler.RejectSurveyAnswers,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code/reject-validate": {"PUT": {
		Handler: surveyHandler.RejectValidateSurveyAnswers,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/survey/:survey_code/result-index/:code/approve-validate": {"PUT": {
		Handler: surveyHandler.ValidateSurveyAnswers,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/approval-history/survey/:survey_code/wilayah/:code_wilayah": {"GET": {
		Handler: surveyHandler.GetApprovalHistorySurvey,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/history-detail/survey/:survey_code/wilayah/:code_wilayah": {"GET": {
		Handler: surveyHandler.GetHistoryDetailPerWilayah,
		MenuKey: "survey-kewilayahan",
	}},

	"/survey-kewilayahan/export-excel/survey/:survey_code/wilayah/:code_wilayah": {"GET": {
		Handler: surveyHandler.ExportExcelSurveyResultsPerRT,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/export-excel/survey/:survey_code": {"GET": {
		Handler: surveyHandler.ExportExcelSurveyResultsMassal,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/reset-status-survey/:survey_code/wilayah/:code_wilayah": {"PUT": {
		Handler: surveyHandler.ResetStatusToVerifySurvey,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/respondent-rejected-all": {"GET": {
		Handler: surveyHandler.GetAllRejectedQuestions,
		MenuKey: "survey-kewilayahan",
	}},
	"/survey-kewilayahan/respondent-rejected-survey/:survey_code": {"GET": {
		Handler: surveyHandler.GetRejectedQuestionsBySurveyCode,
		MenuKey: "survey-kewilayahan",
	}},

	"/hasil-survey/list": {"GET": {
		Handler: hasilSurveyHandler.GetListSurvey,
		MenuKey: "hasil-survey",
	}},
	"/hasil-survey/result-summary/survey/:survey_code": {"GET": {
		Handler: hasilSurveyHandler.SurveyResultSummary,
		MenuKey: "hasil-survey",
	}},

	"/hasil-survey/result-summary/survey/:survey_code/question/:question_id": {"GET": {
		Handler: hasilSurveyHandler.SurveyResultQuestionDetail,
		MenuKey: "hasil-survey",
	}},
	"/hasil-survey/result-summary/survey/:survey_code/respondents": {"GET": {
		Handler: hasilSurveyHandler.SurveyResultRespondentList,
		MenuKey: "hasil-survey",
	}},
	"/hasil-survey/result-summary/survey/:survey_code/survey-respondent/:survey_respondent_id": {"GET": {
		Handler: hasilSurveyHandler.SurveyResultRespondentDetail,
		MenuKey: "hasil-survey",
	}},
	"/hasil-survey/result-summary/export/survey/:survey_code/survey-respondent/:survey_respondent_id": {"GET": {
		Handler: hasilSurveyHandler.ExportExcelSurveyResultPerRespondent,
		MenuKey: "hasil-survey",
	}},
	"/hasil-survey/result-summary/export/survey/:survey_code": {"GET": {
		Handler: hasilSurveyHandler.ExportExcelSurveyResultAll,
		MenuKey: "hasil-survey",
	}},

	"/seed-answer/run": {"POST": {
		Handler: seedAnswerHandler.SeedSurveyAnswer,
		MenuKey: "",
	}},
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
