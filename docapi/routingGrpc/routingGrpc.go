package routingGrpc

import (
	"backend/docapi/handlers"
	"backend/docapi/models"
	"backend/docapi/service"
	"backend/docapi/utils"
	pb "backend/siccore/pb"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type GRPCServer struct {
	pb.UnimplementedProxyServer
}

var (
	uploadService service.UploadService = service.NewUploadService("up")
	trxService    service.TrxService    = service.NewTrxService("up")
)

var (
	uploadHandler handlers.UploadHandler = handlers.NewUploadHandler(
		uploadService,
	)
	eventHandler handlers.EventHandler = handlers.NewEventHandler(
		uploadService,
	)
	trxHandler handlers.TrxHandler = handlers.NewTrxHandler(
		trxService,
	)
)

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error){
	"/view/:id":          {"GET": uploadHandler.ViewFIle},
	"/upload":            {"POST": uploadHandler.UploadFile},
	"/identity/upload":   {"POST": uploadHandler.UploadIdentity},
	"/avatar/upload":     {"POST": uploadHandler.UploadAvatar},
	"/bank/upload":       {"POST": uploadHandler.UploadBank},
	"/webchat/sendfile":  {"POST": uploadHandler.SendFileWebchat},
	"/withdraw/sendfile": {"POST": uploadHandler.SendFileWebchat},

	"/product/upload":  {"POST": eventHandler.UploadProduct},
	"/GenerateReceipt": {"POST": trxHandler.GenerateReceipt},
	"/banner/upload":   {"POST": trxHandler.UploadBanner},
	"/event/upload":    {"POST": trxHandler.UploadBanner},

	"/upload-foto-profil":      {"POST": uploadHandler.UploadFotoProfile},
	"/delete-bulk-foto-profil": {"POST": uploadHandler.DeleteBulkFotoProfil},
	"/view-foto-profil/:path":  {"GET": uploadHandler.ShowFotoProfil},

	"/upload-survey-image":            {"POST": uploadHandler.UploadSurveyImage},
	"/view-survey-image/:id":          {"GET": uploadHandler.ShowSurveyImage},
	"/view-public-survey-image/:path": {"POST": uploadHandler.ShowPublicSurveyImage},
	"/delete-bulk-survey-image":       {"POST": uploadHandler.DeleteBulkSurveyImage},

	"/upload-cms-image":      {"POST": uploadHandler.UploadCMSImage},
	"/view-cms-image/:path":  {"GET": uploadHandler.ShowCMSImage},
	"/delete-bulk-cms-image": {"POST": uploadHandler.DeleteBulkCMSImage},

	"/upload-laporan-konten-image":                   {"POST": uploadHandler.UploadLaporanKontenImage},
	"/view-laporan-konten-image/:path":               {"GET": uploadHandler.ShowLaporanKontenImage},
	"/delete-bulk-laporan-konten-image":              {"POST": uploadHandler.DeleteBulkLaporanKontenImage},
	"/internal/get-laporan-konten-image-bytes/:path": {"GET": uploadHandler.GetLaporanKontenImageBytes},
}

// Metode untuk menangani permintaan yang masuk
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
		message := "Terjadi kesalahan saat unmarshal mengambil metadata"
		utils.LogErrors(message)
		return false, message, int(http.StatusInternalServerError), userData, newToken
	}

	// Ambil token dari metadata
	token := ""
	if val, ok := md["authorization"]; ok {
		// Harap header berbentuk "Bearer <token>"
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
				return false, "Token Tidak Valid", int(http.StatusUnauthorized), userData, newToken
			} else {
				return true, "Tervalidasi", int(http.StatusOK), userData, newToken
			}
		}

		// Dapatkan klaim dari token
		userData = models.JwtCustomClaims{
			ID:           int64(claims.ID),
			RespondentID: claims.RespondentID,
			Name:         claims.Name,
			Email:        claims.Email,
			Role:         claims.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: claims.ExpiresAt,
			},
		}

		if err != nil {
			if req.GetIsSecure() {
				message := "Terjadi kesalahan : " + err.Error()
				utils.LogErrors(message)
				return false, message, int(http.StatusInternalServerError), userData, newToken
			}
		}
		refreshedToken, err := GenerateJWTToken(userData)
		if err != nil {
			utils.LogErrors("Gagal generate token baru: " + err.Error())
		} else {
			newToken = refreshedToken
		}

		return true, "Tervalidasi", int(http.StatusOK), userData, newToken

	}

	return true, "Tervalidasi", int(http.StatusOK), userData, newToken
}

func GenerateJWTToken(user models.JwtCustomClaims) (string, error) {
	defer utils.GeneralRecover()

	claims := models.JwtCustomClaims{
		ID:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return encryptedToken, nil
}
