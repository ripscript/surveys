package service

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/reportapi/request"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LaporanService interface {
	ListLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ChangeNameReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ListSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	CreateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCalculationTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	DeleteReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetDetailReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type laporanService struct {
	laporanRepo repository.LaporanRepo
	fileRepo    repository.FileRepo
	surveyRepo  repository.SurveyRepo
	userRepo    repository.UserRepo
}

func NewLaporanService(
	laporanRepo repository.LaporanRepo,
	fileRepo repository.FileRepo,
	surveyRepo repository.SurveyRepo,
	userRepo repository.UserRepo,
) LaporanService {
	return &laporanService{
		laporanRepo: laporanRepo,
		fileRepo:    fileRepo,
		surveyRepo:  surveyRepo,
		userRepo:    userRepo,
	}
}

func (service *laporanService) ListLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := request.LaporanDatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, err := service.laporanRepo.GetListReport(usr.RespondentID, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1

	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries": totalData,
			"current_page":  payload.Page,
			"per_page":      payload.Limit,
			"showing_from":  showingFrom,
			"showing_to":    showingTo,
		},
	}

	return utils.SendData(result, "Berhasil membuat laporan")
}

func (service *laporanService) ChangeNameReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	var payload request.ChangeNameLaporanPayload
	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	if laporan.Name != payload.Name {
		isNameExists, err := service.laporanRepo.IsNameExists(nil, payload.Name)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if isNameExists {
			return utils.SendError(errors.New("Nama laporan sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
		}
	}

	laporan.Name = payload.Name

	_, err = service.laporanRepo.UpdateReport(*laporan)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(nil, "Laporan berhasil diperbarui")
}

func (service *laporanService) CreateReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil || respondentLogin == nil || respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	var _req request.CreateLaporanPayload
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if err := json.Unmarshal(reqBytes, &_req); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(_req)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	exists, err := service.laporanRepo.IsNameExists(ctx, _req.NamaLaporan)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if exists {
		return utils.SendError(errors.New("nama laporan sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
	}

	var (
		tingkatWilayah int
		kecamatanIDs   []int64
		kelurahanIDs   []int64
		rwIDs          []int64
	)

	roleLogin := *respondentLogin.RoleId

	switch roleLogin {
	case int64(enums.ROLE_ADMIN):
		switch _req.TingkatWilayah {
		case 6:
			tingkatWilayah = 6
			// breakdown otomatis = semua kecamatan aktif, tidak perlu kecamatan_ids
		case 5:
			if len(_req.KecamatanIDs) == 0 {
				return utils.SendError(errors.New("kecamatan_ids wajib diisi untuk tingkat wilayah kecamatan"), http.StatusBadRequest)
			}
			valid, verr := service.laporanRepo.ValidateKecamatanExist(_req.KecamatanIDs)
			if verr != nil {
				return utils.SendError(verr, http.StatusInternalServerError)
			}
			if !valid {
				return utils.SendError(errors.New("kecamatan_ids tidak valid"), http.StatusBadRequest)
			}
			tingkatWilayah = 5
			kecamatanIDs = _req.KecamatanIDs

			for _, kecID := range kecamatanIDs {
				children, cerr := service.laporanRepo.GetKelurahanIDsByKecamatanID(kecID)
				if cerr != nil {
					return utils.SendError(cerr, http.StatusInternalServerError)
				}
				kelurahanIDs = append(kelurahanIDs, children...)
			}
		case 4:
			if len(_req.KelurahanIDs) == 0 {
				return utils.SendError(errors.New("kelurahan_ids wajib diisi untuk tingkat wilayah kelurahan"), http.StatusBadRequest)
			}
			valid, verr := service.laporanRepo.ValidateKelurahanExist(_req.KelurahanIDs)
			if verr != nil {
				return utils.SendError(verr, http.StatusInternalServerError)
			}
			if !valid {
				return utils.SendError(errors.New("kelurahan_ids tidak valid"), http.StatusBadRequest)
			}
			tingkatWilayah = 4
			kelurahanIDs = _req.KelurahanIDs

			for _, kelID := range kelurahanIDs {
				children, cerr := service.laporanRepo.GetRWIDsByKelurahanID(kelID)
				if cerr != nil {
					return utils.SendError(cerr, http.StatusInternalServerError)
				}
				rwIDs = append(rwIDs, children...)
			}
		default:
			return utils.SendError(errors.New("tingkat_wilayah tidak valid"), http.StatusBadRequest)
		}

	case int64(enums.ROLE_KECAMATAN):
		if respondentLogin.KecamatanId == nil || *respondentLogin.KecamatanId == 0 {
			return utils.SendError(errors.New("akun kecamatan tidak memiliki data wilayah"), http.StatusForbidden)
		}
		tingkatWilayah = 5
		kecamatanID := *respondentLogin.KecamatanId
		kecamatanIDs = []int64{kecamatanID}

		kelurahanIDs, err = service.laporanRepo.GetKelurahanIDsByKecamatanID(kecamatanID)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

	case int64(enums.ROLE_KELURAHAN):
		if respondentLogin.KelurahanId == nil || *respondentLogin.KelurahanId == 0 {
			return utils.SendError(errors.New("akun kelurahan tidak memiliki data wilayah"), http.StatusForbidden)
		}
		tingkatWilayah = 4
		kelurahanID := *respondentLogin.KelurahanId
		kelurahanIDs = []int64{kelurahanID}

		rwIDs, err = service.laporanRepo.GetRWIDsByKelurahanID(kelurahanID)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

	default:
		return utils.SendError(errors.New("role Anda tidak diizinkan membuat laporan"), http.StatusForbidden)
	}

	kecamatanJSON, err := json.Marshal(kecamatanIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	kelurahanJSON, err := json.Marshal(kelurahanIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	rwJSON, err := json.Marshal(rwIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	surveyJSON, err := json.Marshal(_req.SurveyIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if string(kecamatanJSON) == "null" {
		kecamatanJSON = []byte("[]")
	}
	if string(kelurahanJSON) == "null" {
		kelurahanJSON = []byte("[]")
	}
	if string(rwJSON) == "null" {
		rwJSON = []byte("[]")
	}

	laporan := models.Report{
		Name:           _req.NamaLaporan,
		RespondentID:   usr.RespondentID,
		TingkatWilayah: strconv.Itoa(tingkatWilayah),
		KecamatanID:    datatypes.JSON(kecamatanJSON),
		KelurahanID:    datatypes.JSON(kelurahanJSON),
		RWID:           datatypes.JSON(rwJSON),
		SurveyID:       datatypes.JSON(surveyJSON),
	}

	var result *models.Report

	// Bungkus create laporan + create cover dalam satu transaction
	err = service.laporanRepo.WithTransaction(ctx, func(txRepo repository.LaporanRepo) error {
		var txErr error
		result, txErr = txRepo.CreateReport(ctx, laporan)
		if txErr != nil {
			return txErr
		}

		laporanCover := models.ReportCover{
			ReportID: int64(result.ID),
		}

		_, txErr = txRepo.CreateReportCover(ctx, laporanCover)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		// Tangani duplicate key dari DB (race condition safety net)
		if utils.IsDuplicateKeyError(err) {
			return utils.SendError(errors.New("nama laporan sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	data := map[string]interface{}{
		"id": result.ID,
	}

	return utils.SendData(data, "Berhasil membuat laporan")
}

func (service *laporanService) GetCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	getLaporanKonten, err := service.laporanRepo.GetReportCoverByReportId(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getLaporanKonten == nil {
		return utils.SendError(errors.New("Konten laporan tidak ditemukan"), http.StatusNotFound)
	}

	var pathDepan *string
	if getLaporanKonten.ImgDepan != nil {
		pathDepan = utils.StringToPointer(os.Getenv("API_GATEWAY_URL") + "/view-laporan-konten-image/" + *getLaporanKonten.ImgDepan)
	}

	var pathBelakang *string
	if getLaporanKonten.ImgBelakang != nil {
		pathBelakang = utils.StringToPointer(os.Getenv("API_GATEWAY_URL") + "/view-laporan-konten-image/" + *getLaporanKonten.ImgBelakang)
	}

	var response = response.CoverLaporanResponse{
		NamaLaporan:               laporan.Name,
		DeskripsiHalamanDepan:     getLaporanKonten.TextDepan,
		BackgroundHalamanDepan:    pathDepan,
		DeskripsiHalamanBelakang:  getLaporanKonten.TextBelakang,
		BackgroundHalamanBelakang: pathBelakang,
		KataPengantar:             getLaporanKonten.KataPengantar,
	}

	return utils.SendData(response, "Berhasil mengambil data cover laporan")
}

func (service *laporanService) GetDetailReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	if laporan.Cover != nil {
		if laporan.Cover.ImgDepan != nil {
			laporan.Cover.LinkImgDepan = utils.StringToPointer(os.Getenv("API_GATEWAY_URL") + "/view-laporan-konten-image/" + *laporan.Cover.ImgDepan)
		}

		if laporan.Cover.ImgBelakang != nil {
			laporan.Cover.LinkImgBelakang = utils.StringToPointer(os.Getenv("API_GATEWAY_URL") + "/view-laporan-konten-image/" + *laporan.Cover.ImgBelakang)
		}
	}

	return utils.SendData(laporan, "Berhasil mengambil data laporan")
}

func (service *laporanService) ListSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := request.SectionLaporanDatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, err := service.laporanRepo.GetListSectionReport(laporan.ID, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1

	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries": totalData,
			"current_page":  payload.Page,
			"per_page":      payload.Limit,
			"showing_from":  showingFrom,
			"showing_to":    showingTo,
		},
	}

	return utils.SendData(result, "Berhasil membuat laporan")
}

func isHTTPURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func isManagedStorageURL(s string, storageBaseURL string) bool {
	return storageBaseURL != "" && strings.HasPrefix(s, storageBaseURL)
}

const cmsImagePathMarker = "/view-laporan-konten-image/"

func toRelativeImagePath(image string) string {
	_, after, ok := strings.Cut(image, cmsImagePathMarker)
	if !ok {
		return image
	}
	return after
}

func (service *laporanService) UpdateCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	var payload request.UpdateLaporanPayload
	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	_, err = service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	// konten lama (bisa nil kalau belum pernah ada)
	existingKonten, err := service.laporanRepo.GetReportCoverByReportId(laporanID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	maxSizeInKB := float64(5120)

	var uploadedFiles []string
	rollbackUploadedFiles := func() {
		if len(uploadedFiles) > 0 {
			service.fileRepo.DeleteLaporanKontenImageBulk(detachedCtx, uploadedFiles)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			rollbackUploadedFiles()
			panic(r)
		}
	}()

	// path lama dari DB, dipakai sebagai fallback saat payload berupa "link" (keep existing)
	var existingImgDepan, existingImgBelakang *string
	if existingKonten != nil {
		existingImgDepan = existingKonten.ImgDepan
		existingImgBelakang = existingKonten.ImgBelakang
	}

	finalImgDepan, err := service.resolveBackgroundImage(
		ctx, payload.BackgroundHalamanDepan, existingImgDepan,
		availableMime, availablesExt, maxSizeInKB,
		&uploadedFiles, rollbackUploadedFiles,
	)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	finalImgBelakang, err := service.resolveBackgroundImage(
		ctx, payload.BackgroundHalamanBelakang, existingImgBelakang,
		availableMime, availablesExt, maxSizeInKB,
		&uploadedFiles, rollbackUploadedFiles,
	)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// tentukan file lama mana yang perlu dihapus (yang tidak lagi dipakai di hasil akhir)
	// karena mode "keep" sekarang selalu memakai existingImgDepan/Belakang secara langsung,
	// perbandingan pointer/isi string di sini sudah aman dan tidak akan salah hapus.
	var pathsToDelete []string
	if existingKonten != nil {
		if existingKonten.ImgDepan != nil && !isSamePath(existingKonten.ImgDepan, finalImgDepan) {
			pathsToDelete = append(pathsToDelete, *existingKonten.ImgDepan)
		}
		if existingKonten.ImgBelakang != nil && !isSamePath(existingKonten.ImgBelakang, finalImgBelakang) {
			pathsToDelete = append(pathsToDelete, *existingKonten.ImgBelakang)
		}
	}

	konten := models.ReportCover{
		ReportID:      laporanID,
		TextDepan:     payload.DeskripsiHalamanDepan,
		ImgDepan:      finalImgDepan,
		TextBelakang:  payload.DeskripsiHalamanBelakang,
		ImgBelakang:   finalImgBelakang,
		KataPengantar: payload.KataPengantar,
	}

	// hapus konten lama, ganti dengan yang baru (sesuai keinginan kamu: hapus & timpa)
	err = service.laporanRepo.ReplaceReportCover(laporanID, konten)
	if err != nil {
		rollbackUploadedFiles()
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if len(pathsToDelete) > 0 {
		service.fileRepo.DeleteLaporanKontenImageBulk(detachedCtx, pathsToDelete)
	}

	return utils.SendData(nil, "Laporan berhasil diperbarui")
}

func (service *laporanService) resolveBackgroundImage(
	ctx context.Context,
	newValue *string,
	existingPath *string,
	availableMime []string,
	availableExt []string,
	maxSizeInKB float64,
	uploadedFiles *[]string,
	rollbackUploadedFiles func(),
) (*string, error) {

	// null / kosong -> hapus background
	if newValue == nil || strings.TrimSpace(*newValue) == "" {
		return nil, nil
	}
	val := *newValue

	// bukan data URI -> mode "keep existing", abaikan isi link, pakai path lama dari DB
	if !strings.HasPrefix(val, "data:") {
		return existingPath, nil
	}

	// mode "upload baru"
	if !strings.HasPrefix(val, "data:image") {
		rollbackUploadedFiles()
		return nil, errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)")
	}
	base64Data, err := utils.ExtractBase64Info(val)
	if err != nil {
		rollbackUploadedFiles()
		return nil, errors.New("Gagal memproses gambar")
	}
	if !slices.Contains(availableExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
		rollbackUploadedFiles()
		return nil, errors.New("Format file gambar tidak didukung")
	}
	if base64Data.SizeInKB > maxSizeInKB {
		rollbackUploadedFiles()
		return nil, errors.New("Ukuran gambar tidak boleh melebihi 5MB")
	}

	path, err := service.fileRepo.UploadLaporanKontenImage(ctx, &val)
	if err != nil || path == nil {
		rollbackUploadedFiles()
		return nil, errors.New("Gagal mengunggah gambar: " + err.Error())
	}
	*uploadedFiles = append(*uploadedFiles, *path)
	return path, nil
}

func isSamePath(old *string, new *string) bool {
	if old == nil || new == nil {
		return old == new
	}
	return *old == *new
}

// Helper untuk mengekstrak ekstensi dari mime type yang aman
func getExtFromMime(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg"):
		return ".jpg"
	case strings.Contains(mimeType, "png"):
		return ".png"
	case strings.Contains(mimeType, "webp"):
		return ".webp"
	default:
		return ".jpg" // Fallback aman
	}
}

func (service *laporanService) renderNarrativeComponent(m core.Maroto, laporan *models.Report, comp models.ReportComponent) {
	if comp.NarrativeTemplate == nil || *comp.NarrativeTemplate == "" {
		return
	}

	finalText := *comp.NarrativeTemplate

	// Unmarshal NarrativeLogic dari datatypes.JSONB
	if len(comp.NarrativeLogic) > 0 {
		var logic request.NarrativeLogicPayload
		if err := json.Unmarshal(comp.NarrativeLogic, &logic); err == nil {
			for _, v := range logic.Variables {
				// Hitung nilai variabel via SQL Repo
				calcVal, err := service.laporanRepo.CalculateNarrativeVariable(laporan, v)
				if err == nil {
					finalText = strings.ReplaceAll(finalText, v.Placeholder, calcVal)
				}
			}
		}
	}

	// Cetak teks hasil kalkulasi
	paragraphs := strings.Split(finalText, "\n\n")
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m.AddAutoRow(
			text.NewCol(12, p, props.Text{
				Family: "Tinos",
				Size:   11,
				Align:  align.Justify,
			}),
		)
		m.AddRow(4, text.NewCol(12, " "))
	}
}

func sumWidthsBefore(groupSpans []int, uptoGroupIdx int) int {
	count := 0
	for i := 0; i < uptoGroupIdx; i++ {
		count += groupSpans[i]
	}
	return count
}

func validateTableConfig(cfg *request.TableConfigPayload) error {
	if cfg == nil {
		return nil
	}
	switch cfg.TableStyle {
	case "grouped_header":
		if len(cfg.Groups) == 0 {
			return errors.New("groups wajib diisi jika table_style = grouped_header")
		}

	case "respondent_text_grouped", "respondent_media_grouped":
		if len(cfg.Groups) == 0 {
			return fmt.Errorf("groups wajib diisi jika table_style = %s", cfg.TableStyle)
		}

	case "simple":
		if len(cfg.Columns) == 0 {
			return errors.New("columns wajib diisi jika table_style = simple")
		}

	case "respondent_text", "respondent_media":
		if len(cfg.Columns) == 0 {
			return fmt.Errorf("columns wajib diisi jika table_style = %s", cfg.TableStyle)
		}
	}
	return nil
}

func validateChartConfig(cfg *request.ChartConfigPayload) error {
	if cfg == nil {
		return errors.New("chart_config wajib diisi jika component_type = chart")
	}

	switch cfg.ChartType {
	case "pie":
		if cfg.ChartDirection != nil {
			return errors.New("chart_direction tidak boleh diisi untuk chart_type = pie")
		}
		if !cfg.IsMultipleData {
			return errors.New("is_multiple_data wajib bernilai true untuk chart_type = pie")
		}
		if len(cfg.FormFieldIDs) < 1 {
			return errors.New("form_field_ids wajib diisi untuk chart_type = pie")
		}
		// Aturan jumlah field yang detail (min 2 untuk number, tepat 1 untuk
		// multiple-choices) dicek belakangan di validateFieldTypesForComponent,
		// karena butuh tahu tipe field dulu dari database.
	case "bar", "line":
		if cfg.ChartDirection == nil {
			return errors.New("chart_direction wajib diisi untuk chart_type = bar/line")
		}
		if cfg.IsMultipleData {
			if len(cfg.FormFieldIDs) < 2 {
				return errors.New("form_field_ids minimal 2 jika is_multiple_data bernilai true")
			}
		} else {
			if len(cfg.FormFieldIDs) != 1 {
				return errors.New("form_field_ids harus tepat 1 jika is_multiple_data bernilai false")
			}
		}

	case "map":
		if cfg.ChartDirection != nil {
			return errors.New("chart_direction tidak boleh diisi untuk chart_type = map")
		}
		if cfg.IsMultipleData {
			return errors.New("is_multiple_data harus bernilai false untuk chart_type = map")
		}
		if len(cfg.FormFieldIDs) != 1 {
			return errors.New("form_field_ids harus tepat 1 untuk chart_type = map")
		}
		if cfg.MapType == nil {
			return errors.New("map_type wajib diisi untuk chart_type = map")
		}
	}
	return nil
}

func extractFieldIDsFromConfig(componentType string, tableCfg *request.TableConfigPayload, chartCfg *request.ChartConfigPayload) []int {
	var ids []int

	switch componentType {
	case "table":
		if tableCfg == nil {
			return ids
		}
		switch tableCfg.TableStyle {
		case "grouped_header", "respondent_text_grouped", "respondent_media_grouped":
			for _, g := range tableCfg.Groups {
				for _, col := range g.Columns {
					ids = append(ids, col.FormFieldID)
				}
			}
		case "simple", "respondent_text", "respondent_media":
			for _, col := range tableCfg.Columns {
				ids = append(ids, col.FormFieldID)
			}
		}
	case "chart":
		if chartCfg == nil {
			return ids
		}
		ids = append(ids, chartCfg.FormFieldIDs...)
	}

	return ids
}

func allowedFieldTypesForComponent(componentType string, tableCfg *request.TableConfigPayload, chartCfg *request.ChartConfigPayload) map[string]bool {
	switch componentType {
	case "table":
		if tableCfg == nil {
			return map[string]bool{"number": true}
		}
		switch tableCfg.TableStyle {
		case "respondent_text", "respondent_text_grouped":
			return map[string]bool{"long-answer": true}
		case "respondent_media", "respondent_media_grouped":
			return map[string]bool{"image-template": true}
		default: // simple, grouped_header
			return map[string]bool{"number": true}
		}
	case "chart":
		if chartCfg != nil {
			switch chartCfg.ChartType {
			case "pie":
				return map[string]bool{"number": true, "multiple-choices": true}
			case "map":
				return map[string]bool{"maps": true}
			}
		}
		return map[string]bool{"number": true}
	default:
		return map[string]bool{}
	}
}

func (service *laporanService) validateFieldTypesForComponent(componentType string, tableCfg *request.TableConfigPayload, chartCfg *request.ChartConfigPayload) error {
	fieldIDs := extractFieldIDsFromConfig(componentType, tableCfg, chartCfg)
	if len(fieldIDs) == 0 {
		return nil
	}

	fieldTypes, err := service.laporanRepo.GetFormFieldTypes(fieldIDs)
	if err != nil {
		return err
	}

	allowed := allowedFieldTypesForComponent(componentType, tableCfg, chartCfg)

	for _, id := range fieldIDs {
		fieldType, exists := fieldTypes[id]
		if !exists {
			return fmt.Errorf("form_field_id %d tidak ditemukan", id)
		}
		if !allowed[fieldType] {
			if componentType == "table" {
				return fmt.Errorf("form_field_id %d bertipe '%s' tidak sesuai dengan table_style '%s'", id, fieldType, tableCfg.TableStyle)
			}
			return fmt.Errorf("form_field_id %d bertipe '%s' tidak bisa dipakai di chart_type '%s'", id, fieldType, chartCfg.ChartType)
		}
	}

	// Aturan khusus pie: tidak boleh campur number + multiple-choices,
	// dan jumlah field harus sesuai mode-nya.
	if componentType == "chart" && chartCfg != nil && chartCfg.ChartType == "pie" {
		hasMultipleChoice := false
		hasNumber := false
		for _, id := range fieldIDs {
			switch fieldTypes[id] {
			case "multiple-choices":
				hasMultipleChoice = true
			case "number":
				hasNumber = true
			}
		}

		if hasMultipleChoice && hasNumber {
			return errors.New("chart pie tidak boleh mencampur form_field bertipe 'number' dan 'multiple-choices' dalam satu chart")
		}
		if hasMultipleChoice {
			if len(fieldIDs) != 1 {
				return errors.New("chart pie dengan tipe multiple-choices hanya boleh menggunakan 1 form_field_id (slice diambil otomatis dari opsi jawaban field tersebut)")
			}
		} else {
			if len(fieldIDs) < 2 {
				return errors.New("chart pie dengan tipe number minimal harus 2 form_field_ids")
			}
		}
	}

	return nil
}

func (service *laporanService) validateComponentConfig(componentType string, tableCfg *request.TableConfigPayload, chartCfg *request.ChartConfigPayload) error {
	switch componentType {
	case "table":
		if err := validateTableConfig(tableCfg); err != nil {
			return err
		}
	case "chart":
		if err := validateChartConfig(chartCfg); err != nil {
			return err
		}
	}

	return service.validateFieldTypesForComponent(componentType, tableCfg, chartCfg)
}

func parseComponentConfig(componentType string, raw json.RawMessage) (
	tableCfg *request.TableConfigPayload,
	chartCfg *request.ChartConfigPayload,
	err error,
) {
	if len(raw) == 0 {
		return nil, nil, nil
	}

	switch componentType {
	case "table":
		var cfg request.TableConfigPayload
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, nil, fmt.Errorf("component_config tidak valid untuk component_type = table: %w", err)
		}
		return &cfg, nil, nil

	case "chart":
		var cfg request.ChartConfigPayload
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, nil, fmt.Errorf("component_config tidak valid untuk component_type = chart: %w", err)
		}
		return nil, &cfg, nil

	default:
		// component_type tak dikenal sudah ditolak oleh validator (oneof=table chart)
		// sebelum fungsi ini dipanggil, jadi baris ini hanya jaring pengaman.
		return nil, nil, fmt.Errorf("component_type tidak dikenal: %s", componentType)
	}
}

func (service *laporanService) CreateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// 1. Validasi ID Laporan dari URL Slug
	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	// 2. Parsing & Unmarshal JSON ke DTO
	var payloads request.UpsertSectionPayload
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if err := json.Unmarshal(reqBytes, &payloads); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// 3. Validasi Struct Payload dasar
	var validate = validator.New()
	if err := validate.Struct(payloads); err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	if !*payloads.HasSubSection {
		tableCfg, chartCfg, err := parseComponentConfig(payloads.ComponentType, payloads.ComponentConfig)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		if err := service.validateComponentConfig(payloads.ComponentType, tableCfg, chartCfg); err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	} else {
		for _, pSub := range payloads.SubSections {
			tableCfg, chartCfg, err := parseComponentConfig(pSub.ComponentType, pSub.ComponentConfig)
			if err != nil {
				return utils.SendError(err, http.StatusBadRequest)
			}
			if err := service.validateComponentConfig(pSub.ComponentType, tableCfg, chartCfg); err != nil {
				return utils.SendError(err, http.StatusBadRequest)
			}
		}
	}

	// 3.5 VALIDASI BISNIS: konsistensi has_sub_section vs isi payload
	if *payloads.HasSubSection {
		if len(payloads.SubSections) == 0 {
			return utils.SendError(errors.New("sub_sections wajib diisi jika has_sub_section bernilai true"), http.StatusBadRequest)
		}
	} else {
		if len(payloads.SubSections) > 0 {
			return utils.SendError(errors.New("sub_sections tidak boleh diisi jika has_sub_section bernilai false"), http.StatusBadRequest)
		}
	}

	// 4. Pastikan Laporan Induk ada di Database
	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	// 4.5 VALIDASI BISNIS: choropleth hanya untuk tingkat kecamatan/kota
	if !*payloads.HasSubSection {
		_, chartCfg, _ := parseComponentConfig(payloads.ComponentType, payloads.ComponentConfig)
		if err := validateMapTypeAgainstTingkatWilayah(chartCfg, laporan.TingkatWilayah); err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	} else {
		for _, pSub := range payloads.SubSections {
			_, chartCfg, _ := parseComponentConfig(pSub.ComponentType, pSub.ComponentConfig)
			if err := validateMapTypeAgainstTingkatWilayah(chartCfg, laporan.TingkatWilayah); err != nil {
				return utils.SendError(err, http.StatusBadRequest)
			}
		}
	}

	// ========================================================================
	// 5. VALIDASI UNIQUE TITLE (BUSINESS LOGIC)
	// ========================================================================
	isSectionExists, err := service.laporanRepo.CheckSectionTitleExists(laporanID, payloads.Title)
	if err != nil {
		return utils.SendError(errors.New("Gagal memvalidasi nama section"), http.StatusInternalServerError)
	}
	if isSectionExists {
		return utils.SendError(errors.New("Nama Section sudah digunakan di laporan ini. Silakan gunakan nama lain."), http.StatusConflict)
	}

	if *payloads.HasSubSection {
		subMapInternal := make(map[string]bool)

		for _, pSub := range payloads.SubSections {
			subTitleLower := strings.ToLower(pSub.Title)

			if subMapInternal[subTitleLower] {
				return utils.SendError(fmt.Errorf("Terdapat duplikasi nama Sub-Section ('%s') di dalam form Anda", pSub.Title), http.StatusBadRequest)
			}
			subMapInternal[subTitleLower] = true

			isSubExists, err := service.laporanRepo.CheckSubSectionTitleExists(laporanID, pSub.Title)
			if err != nil {
				return utils.SendError(errors.New("Gagal memvalidasi nama sub-section"), http.StatusInternalServerError)
			}
			if isSubExists {
				return utils.SendError(fmt.Errorf("Nama Sub-Section '%s' sudah ada di laporan ini. Silakan gunakan nama lain.", pSub.Title), http.StatusConflict)
			}
		}
	}

	// ========================================================================
	// 5.5 AUTO-GENERATE SEQUENCE SECTION
	// ========================================================================
	maxSeq, err := service.laporanRepo.GetMaxSectionSequence(laporanID)
	if err != nil {
		return utils.SendError(errors.New("Gagal menghitung urutan section terbaru"), http.StatusInternalServerError)
	}
	nextSequence := maxSeq + 1

	// ========================================================================
	// 6. DATA MAPPING: Payload Frontend -> GORM Model
	// ========================================================================
	sectionModel := models.ReportSection{
		ReportID:      laporanID,
		Title:         payloads.Title,
		Sequence:      nextSequence,
		HasSubSection: payloads.HasSubSection,
	}

	if *payloads.HasSubSection {
		// --- Pola lama: section punya sub_sections, masing-masing punya component sendiri ---
		var subSections []models.ReportSubSection

		for _, pSub := range payloads.SubSections {
			subModel := models.ReportSubSection{
				Title:    pSub.Title,
				Sequence: pSub.Sequence,
			}

			// error diabaikan aman: payload ini sudah lolos parseComponentConfig
			// + validateComponentConfig di langkah validasi sebelumnya.
			subTableCfg, subChartCfg, _ := parseComponentConfig(pSub.ComponentType, pSub.ComponentConfig)
			subModel.Components = buildVisualAndNarrativeComponents(
				pSub.ComponentType,
				subTableCfg,
				subChartCfg,
				pSub.NarrativePosition,
				pSub.NarrativeTemplate,
				pSub.NarrativeLogic,
			)

			subSections = append(subSections, subModel)
		}
		sectionModel.SubSections = subSections

	} else {
		// --- Pola baru: section flat, component nempel langsung ke section ---
		tableCfg, chartCfg, _ := parseComponentConfig(payloads.ComponentType, payloads.ComponentConfig)
		sectionModel.Components = buildVisualAndNarrativeComponents(
			payloads.ComponentType,
			tableCfg,
			chartCfg,
			payloads.NarrativePosition,
			payloads.NarrativeTemplate,
			payloads.NarrativeLogic,
		)
	}

	// ========================================================================
	// 7. EKSEKUSI DATABASE DENGAN TRANSACTION (PURE CREATE)
	// ========================================================================
	errTx := service.laporanRepo.WithTransaction(ctx, func(txRepo repository.LaporanRepo) error {
		errCreate := txRepo.CreateSectionTx(&sectionModel)
		if errCreate != nil {
			return errCreate
		}
		return nil
	})

	if errTx != nil {
		return utils.SendError(errors.New("Gagal menyimpan struktur laporan: "+errTx.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil membuat section laporan baru")
}

func buildVisualAndNarrativeComponents(
	componentType string,
	componentConfig *request.TableConfigPayload,
	chartConfig *request.ChartConfigPayload,
	narrativePosition *string,
	narrativeTemplate *string,
	narrativeLogic *request.NarrativeLogicPayload,
) []models.ReportComponent {

	visualComp := models.ReportComponent{
		Type: componentType,
	}

	switch componentType {
	case "table":
		if componentConfig != nil {
			visualComp.TableStyle = &componentConfig.TableStyle
			configBytes, _ := json.Marshal(componentConfig)
			visualComp.TableConfig = datatypes.JSON(configBytes)
		}
	case "chart":
		if chartConfig != nil {
			chartType := chartConfig.ChartType
			visualComp.ChartType = &chartType
			visualComp.ChartDirection = chartConfig.ChartDirection
			visualComp.IsMultipleData = chartConfig.IsMultipleData
			visualComp.MapType = chartConfig.MapType

			fieldIDsBytes, _ := json.Marshal(chartConfig.FormFieldIDs)
			visualComp.FormFieldIDs = datatypes.JSON(fieldIDsBytes)
		}
	}

	var narrativeComp *models.ReportComponent
	if narrativeTemplate != nil && narrativeLogic != nil {
		logicBytes, _ := json.Marshal(narrativeLogic)
		narrativeComp = &models.ReportComponent{
			Type:              "narrative",
			NarrativeTemplate: narrativeTemplate,
			NarrativeLogic:    datatypes.JSON(logicBytes),
		}
	}

	pos := "bottom"
	if narrativePosition != nil {
		pos = *narrativePosition
	}

	var components []models.ReportComponent
	if narrativeComp != nil {
		if pos == "top" {
			narrativeComp.Sequence = 1
			visualComp.Sequence = 2
			components = append(components, *narrativeComp, visualComp)
		} else {
			visualComp.Sequence = 1
			narrativeComp.Sequence = 2
			components = append(components, visualComp, *narrativeComp)
		}
	} else {
		visualComp.Sequence = 1
		components = append(components, visualComp)
	}

	return components
}

func (service *laporanService) GetCalculationTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	formFieldIdStr := param.Get("form_id")
	if formFieldIdStr == "" {
		return utils.SendError(errors.New("form_id wajib diisi"), http.StatusBadRequest)
	}

	formFieldId, err := utils.StringToInt64(formFieldIdStr)
	if err != nil {
		return utils.SendError(errors.New("form_id tidak valid"), http.StatusBadRequest)
	}

	template, err := service.surveyRepo.GetFormFieldTemplateByID(formFieldId)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	questionType := enums.QuestionType(template)
	if !questionType.IsQuestionTypeValid() {
		return utils.SendError(errors.New("Tipe pertanyaan tidak valid"), http.StatusBadRequest)
	}

	calculationTypes := enums.GetCalculationTypesByQuestionType(questionType)

	allOptions := make([]response.EnumOption, 0, len(calculationTypes))
	for _, ct := range calculationTypes {
		allOptions = append(allOptions, response.EnumOption{
			ID:   string(ct),
			Name: enums.CalculationTypeLabel[ct],
		})
	}

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	var filteredOptions []response.EnumOption
	if len(rawIDs) > 0 {
		idMap := make(map[string]bool)
		for _, id := range rawIDs {
			idMap[id] = true
		}
		for _, opt := range allOptions {
			if idMap[opt.ID] {
				filteredOptions = append(filteredOptions, opt)
			}
		}
	} else {
		filteredOptions = allOptions
	}

	var excludeIDs []string
	if len(param["exclude_id[]"]) > 0 {
		excludeIDs = param["exclude_id[]"]
	} else if len(param["exclude_id"]) > 0 {
		excludeIDs = param["exclude_id"]
	}

	if len(excludeIDs) > 0 {
		excludeMap := make(map[string]bool)
		for _, id := range excludeIDs {
			excludeMap[id] = true
		}
		var afterExclude []response.EnumOption
		for _, opt := range filteredOptions {
			if !excludeMap[opt.ID] {
				afterExclude = append(afterExclude, opt)
			}
		}
		filteredOptions = afterExclude
	}

	q := strings.ToLower(param.Get("q"))
	var searchedOptions []response.EnumOption
	if q != "" {
		for _, opt := range filteredOptions {
			if strings.Contains(strings.ToLower(opt.Name), q) || strings.Contains(strings.ToLower(opt.ID), q) {
				searchedOptions = append(searchedOptions, opt)
			}
		}
	} else {
		searchedOptions = filteredOptions
	}

	totalData := int64(len(searchedOptions))
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	var paginatedOptions []response.EnumOption
	if startIndex < len(searchedOptions) {
		if endIndex > len(searchedOptions) {
			endIndex = len(searchedOptions)
		}
		paginatedOptions = searchedOptions[startIndex:endIndex]
	} else {
		paginatedOptions = []response.EnumOption{}
	}

	currentTotalLoaded := startIndex + len(paginatedOptions)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsStringIdResponse{
		Options: paginatedOptions,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi tipe kalkulasi")
}

func extractComponentFields(components []models.ReportComponent) (
	string,
	*request.TableConfigPayload,
	*request.ChartConfigPayload,
	*string,
	*string,
	*request.NarrativeLogicPayload,
) {
	var visual *models.ReportComponent
	var narrative *models.ReportComponent

	for i := range components {
		c := &components[i]
		if c.Type == "narrative" {
			narrative = c
		} else {
			visual = c
		}
	}

	var componentType string
	var tableConfig *request.TableConfigPayload
	var chartConfig *request.ChartConfigPayload

	if visual != nil {
		componentType = visual.Type

		switch visual.Type {
		case "table":
			if len(visual.TableConfig) > 0 {
				var cfg request.TableConfigPayload
				if err := json.Unmarshal(visual.TableConfig, &cfg); err == nil {
					tableConfig = &cfg
				}
			}
		case "chart":
			cfg := request.ChartConfigPayload{
				IsMultipleData: visual.IsMultipleData,
				ChartDirection: visual.ChartDirection,
			}
			if visual.ChartType != nil {
				cfg.ChartType = *visual.ChartType
			}
			if len(visual.FormFieldIDs) > 0 {
				var ids []int
				if err := json.Unmarshal(visual.FormFieldIDs, &ids); err == nil {
					cfg.FormFieldIDs = ids
				}
			}
			chartConfig = &cfg
		}
	}

	var narrativePosition *string
	var narrativeTemplate *string
	var narrativeLogic *request.NarrativeLogicPayload

	if narrative != nil {
		pos := "bottom"
		if visual != nil && narrative.Sequence < visual.Sequence {
			pos = "top"
		}
		narrativePosition = &pos
		narrativeTemplate = narrative.NarrativeTemplate

		if len(narrative.NarrativeLogic) > 0 {
			var logic request.NarrativeLogicPayload
			if err := json.Unmarshal(narrative.NarrativeLogic, &logic); err == nil {
				narrativeLogic = &logic
			}
		}
	}

	return componentType, tableConfig, chartConfig, narrativePosition, narrativeTemplate, narrativeLogic
}

func (service *laporanService) GetSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	sectionIdStr, ok := slug["section_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}
	sectionID, err := utils.StringToInt64(sectionIdStr)
	if err != nil {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}

	section, err := service.laporanRepo.GetSectionByID(sectionID)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	resp := mapSectionToDetailResponse(section)
	return utils.SendData(resp, "Berhasil mengambil detail section")
}

func mapSectionToDetailResponse(section *models.ReportSection) response.SectionDetailResponse {
	resp := response.SectionDetailResponse{
		ID:            section.ID,
		Title:         section.Title,
		HasSubSection: section.HasSubSection,
		Sequence:      section.Sequence,
	}

	if section.HasSubSection != nil && *section.HasSubSection {
		for _, sub := range section.SubSections {
			ctype, tcfg, ccfg, pos, tmpl, logic := extractComponentFields(sub.Components)
			resp.SubSections = append(resp.SubSections, response.SubSectionDetailResponse{
				ID:                sub.ID,
				Title:             sub.Title,
				Sequence:          sub.Sequence,
				ComponentType:     ctype,
				ComponentConfig:   tcfg,
				ChartConfig:       ccfg,
				NarrativePosition: pos,
				NarrativeTemplate: tmpl,
				NarrativeLogic:    logic,
			})
		}
	} else {
		ctype, tcfg, ccfg, pos, tmpl, logic := extractComponentFields(section.Components)
		resp.ComponentType = ctype
		resp.ComponentConfig = tcfg
		resp.ChartConfig = ccfg
		resp.NarrativePosition = pos
		resp.NarrativeTemplate = tmpl
		resp.NarrativeLogic = logic
	}

	return resp
}

func (service *laporanService) UpdateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// 1. Validasi ID Laporan & Section dari URL Slug
	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	sectionIdStr, ok := slug["section_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}
	sectionID, err := utils.StringToInt64(sectionIdStr)
	if err != nil {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}

	// 2. Parsing & Unmarshal JSON ke DTO
	var payloads request.UpdateSectionPayload
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if err := json.Unmarshal(reqBytes, &payloads); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// 3. Validasi Struct Payload dasar
	var validate = validator.New()
	if err := validate.Struct(payloads); err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	// 3.5 Validasi table_config sesuai style
	if !*payloads.HasSubSection {
		tableCfg, chartCfg, err := parseComponentConfig(payloads.ComponentType, payloads.ComponentConfig)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		if err := service.validateComponentConfig(payloads.ComponentType, tableCfg, chartCfg); err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	} else {
		for _, pSub := range payloads.SubSections {
			tableCfg, chartCfg, err := parseComponentConfig(pSub.ComponentType, pSub.ComponentConfig)
			if err != nil {
				return utils.SendError(err, http.StatusBadRequest)
			}
			if err := service.validateComponentConfig(pSub.ComponentType, tableCfg, chartCfg); err != nil {
				return utils.SendError(err, http.StatusBadRequest)
			}
		}
	}

	// 3.6 Validasi bisnis: konsistensi has_sub_section vs isi payload
	if *payloads.HasSubSection {
		if len(payloads.SubSections) == 0 {
			return utils.SendError(errors.New("sub_sections wajib diisi jika has_sub_section bernilai true"), http.StatusBadRequest)
		}
	} else {
		if len(payloads.SubSections) > 0 {
			return utils.SendError(errors.New("sub_sections tidak boleh diisi jika has_sub_section bernilai false"), http.StatusBadRequest)
		}
	}

	// 4. Pastikan Laporan Induk ada
	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if laporan == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	// 5. Pastikan Section yang mau di-update ada & benar-benar milik Laporan ini
	existingSection, err := service.laporanRepo.GetSectionByID(sectionID)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if existingSection.ReportID != laporanID {
		return utils.SendError(errors.New("Section tidak ditemukan pada laporan ini"), http.StatusNotFound)
	}

	// ========================================================================
	// 6. VALIDASI UNIQUE TITLE — EXCLUDE section ini sendiri
	// ========================================================================
	if !strings.EqualFold(existingSection.Title, payloads.Title) {
		isSectionExists, err := service.laporanRepo.CheckSectionTitleExistsExcludingID(laporanID, payloads.Title, sectionID)
		if err != nil {
			return utils.SendError(errors.New("Gagal memvalidasi nama section"), http.StatusInternalServerError)
		}
		if isSectionExists {
			return utils.SendError(errors.New("Nama Section sudah digunakan di laporan ini. Silakan gunakan nama lain."), http.StatusConflict)
		}
	}

	if *payloads.HasSubSection {
		subMapInternal := make(map[string]bool)
		for _, pSub := range payloads.SubSections {
			subTitleLower := strings.ToLower(pSub.Title)
			if subMapInternal[subTitleLower] {
				return utils.SendError(fmt.Errorf("Terdapat duplikasi nama Sub-Section ('%s') di dalam form Anda", pSub.Title), http.StatusBadRequest)
			}
			subMapInternal[subTitleLower] = true

			// Cek ke DB, exclude sub-section milik section ini sendiri
			// (karena section ini akan di-replace total, semua sub-section lama section ini boleh "bentrok" dengan judul baru)
			isSubExists, err := service.laporanRepo.CheckSubSectionTitleExistsExcludingSection(laporanID, pSub.Title, sectionID)
			if err != nil {
				return utils.SendError(errors.New("Gagal memvalidasi nama sub-section"), http.StatusInternalServerError)
			}
			if isSubExists {
				return utils.SendError(fmt.Errorf("Nama Sub-Section '%s' sudah ada di laporan ini. Silakan gunakan nama lain.", pSub.Title), http.StatusConflict)
			}
		}
	}

	// ========================================================================
	// 7. DATA MAPPING: Payload -> GORM Model (sama seperti create)
	// ========================================================================
	var newSubSections []models.ReportSubSection
	var newFlatComponents []models.ReportComponent

	if *payloads.HasSubSection {
		for _, pSub := range payloads.SubSections {
			subModel := models.ReportSubSection{
				Title:    pSub.Title,
				Sequence: pSub.Sequence,
			}
			subTableCfg, subChartCfg, _ := parseComponentConfig(pSub.ComponentType, pSub.ComponentConfig)
			subModel.Components = buildVisualAndNarrativeComponents(
				pSub.ComponentType,
				subTableCfg,
				subChartCfg,
				pSub.NarrativePosition,
				pSub.NarrativeTemplate,
				pSub.NarrativeLogic,
			)
			newSubSections = append(newSubSections, subModel)
		}
	} else {
		tableCfg, chartCfg, _ := parseComponentConfig(payloads.ComponentType, payloads.ComponentConfig)
		newFlatComponents = buildVisualAndNarrativeComponents(
			payloads.ComponentType,
			tableCfg,
			chartCfg,
			payloads.NarrativePosition,
			payloads.NarrativeTemplate,
			payloads.NarrativeLogic,
		)
	}

	// ========================================================================
	// 8. EKSEKUSI DALAM TRANSAKSI: update meta + replace total sub_sections/components
	// ========================================================================
	errTx := service.laporanRepo.WithTransaction(ctx, func(txRepo repository.LaporanRepo) error {
		if err := txRepo.UpdateSectionMetaTx(sectionID, payloads.Title, payloads.HasSubSection); err != nil {
			return err
		}

		// Hapus SEMUA sub-section & component lama milik section ini (cascade rapi via FK)
		if err := txRepo.DeleteSubSectionsBySectionIDTx(sectionID); err != nil {
			return err
		}
		if err := txRepo.DeleteComponentsBySectionIDTx(sectionID); err != nil {
			return err
		}

		// Insert ulang sesuai payload terbaru
		if *payloads.HasSubSection {
			if err := txRepo.CreateSubSectionsTxWrapped(sectionID, newSubSections); err != nil {
				return err
			}
		} else {
			if err := txRepo.CreateComponentsForSectionTxWrapped(sectionID, newFlatComponents); err != nil {
				return err
			}
		}

		return nil
	})

	if errTx != nil {
		return utils.SendError(errors.New("Gagal memperbarui struktur laporan: "+errTx.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil memperbarui section laporan")
}

func (service *laporanService) DeleteReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	report, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if report == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	var kontenImage []string

	if report.Cover != nil {
		if report.Cover.ImgDepan != nil {
			kontenImage = append(kontenImage, *report.Cover.ImgDepan)
		}
		if report.Cover.ImgBelakang != nil {
			kontenImage = append(kontenImage, *report.Cover.ImgBelakang)
		}
	}

	_, err = service.fileRepo.DeleteLaporanKontenImageBulk(detachedCtx, kontenImage)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan saat menghapus section laporan: "+err.Error()), http.StatusInternalServerError)
	}

	err = service.laporanRepo.DeleteReportByID(laporanID)

	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan saat menghapus laporan: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus laporan")
}

func (service *laporanService) DeleteSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	report, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if report == nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	sectionIdStr, ok := slug["section_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}
	sectionID, err := utils.StringToInt64(sectionIdStr)
	if err != nil {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}

	section, err := service.laporanRepo.GetSectionByID(sectionID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if section == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	err = service.laporanRepo.DeleteSectionReportBySectionID(sectionID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan saat menghapus section laporan: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus section laporan")
}

func validateMapTypeAgainstTingkatWilayah(chartCfg *request.ChartConfigPayload, tingkatWilayah string) error {
	if chartCfg == nil || chartCfg.ChartType != "map" || chartCfg.MapType == nil {
		return nil
	}
	if *chartCfg.MapType == "choropleth" && tingkatWilayah == "4" {
		return errors.New("map_type 'choropleth' tidak bisa digunakan untuk laporan tingkat kelurahan")
	}
	return nil
}
