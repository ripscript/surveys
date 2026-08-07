package service

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/reportapi/request"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-playground/validator/v10"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	fontrepo "github.com/johnfercher/maroto/v2/pkg/repository"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LaporanService interface {
	ListLaporan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ChangeNameReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCoverReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PrintReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	CreateSectionReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCalculationTypeOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type laporanService struct {
	laporanRepo repository.LaporanRepo
	fileRepo    repository.FileRepo
	surveyRepo  repository.SurveyRepo
}

func NewLaporanService(
	laporanRepo repository.LaporanRepo,
	fileRepo repository.FileRepo,
	surveyRepo repository.SurveyRepo,
) LaporanService {
	return &laporanService{
		laporanRepo: laporanRepo,
		fileRepo:    fileRepo,
		surveyRepo:  surveyRepo,
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

	// Validasi tingkat wilayah — whitelist eksplisit, tolak nilai tak dikenal
	switch _req.TingkatWilayah {
	case 6: // kota — tidak butuh kecamatan/kelurahan
	case 5: // kecamatan
		if len(_req.KecamatanIDs) == 0 {
			return utils.SendError(errors.New("kecamatan_ids wajib diisi untuk tingkat wilayah kecamatan"), http.StatusBadRequest)
		}
	case 4: // kelurahan
		if len(_req.KelurahanIDs) == 0 {
			return utils.SendError(errors.New("kelurahan_ids wajib diisi untuk tingkat wilayah kelurahan"), http.StatusBadRequest)
		}
	default:
		return utils.SendError(errors.New("tingkat_wilayah tidak valid"), http.StatusBadRequest)
	}

	kecamatanJSON, err := json.Marshal(_req.KecamatanIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	kelurahanJSON, err := json.Marshal(_req.KelurahanIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	surveyJSON, err := json.Marshal(_req.SurveyIDs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	laporan := models.Report{
		Name:           _req.NamaLaporan,
		RespondentID:   usr.ID,
		TingkatWilayah: strconv.Itoa(_req.TingkatWilayah),
		KecamatanID:    datatypes.JSON(kecamatanJSON),
		KelurahanID:    datatypes.JSON(kelurahanJSON),
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
		return nil, errors.New("Gagal mengunggah gambar")
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

func (service *laporanService) PrintReport(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	laporanIdStr, ok := slug["laporan_id"].(string)
	if !ok {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}
	laporanID, err := utils.StringToInt64(laporanIdStr)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak valid"), http.StatusBadRequest)
	}

	// 1. Tarik Data Laporan Beserta Nested Sections, SubSections, & Components
	laporan, err := service.laporanRepo.GetReportByID(laporanID)
	if err != nil {
		return utils.SendError(errors.New("Laporan tidak ditemukan"), http.StatusNotFound)
	}

	konten, err := service.laporanRepo.GetReportCoverByReportId(laporanID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if konten == nil {
		return utils.SendError(errors.New("Konten laporan tidak ditemukan"), http.StatusNotFound)
	}

	var bgDepanBytes, bgBelakangBytes []byte
	var bgDepanExt, bgBelakangExt string

	// Ambil background Cover Depan jika ada
	if konten.ImgDepan != nil && *konten.ImgDepan != "" {
		bgBytes, mimeType, err := service.fileRepo.GetLaporanKontenImageBytes(ctx, *konten.ImgDepan)
		if err == nil {
			bgDepanBytes = bgBytes
			bgDepanExt = getExtFromMime(mimeType)
		}
	}

	// Ambil background Cover Belakang jika ada
	if konten.ImgBelakang != nil && *konten.ImgBelakang != "" {
		bgBytes, mimeType, err := service.fileRepo.GetLaporanKontenImageBytes(ctx, *konten.ImgBelakang)
		if err == nil {
			bgBelakangBytes = bgBytes
			bgBelakangExt = getExtFromMime(mimeType)
		}
	}

	// 2. Build PDF berdasarkan urutan halaman yang baku
	pdfBytes, err := service.buildLaporanPDF(laporan, konten, bgDepanBytes, bgDepanExt, bgBelakangBytes, bgBelakangExt)
	if err != nil {
		return utils.SendError(errors.New("Gagal membuat PDF: "+err.Error()), http.StatusInternalServerError)
	}

	message := fmt.Sprintf("Data File,%s,%s", "application/pdf", "Laporan.pdf")
	return utils.SetResponseData(pdfBytes, true, message, http.StatusOK, nil, ""), nil
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

func resizeCoverToA4(imgBytes []byte, ext string) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal decode gambar: %w", err)
	}

	// Target rasio A4 dalam pixel (pakai skala tinggi biar hasil tidak pecah, misal 150 DPI)
	const targetW = 1240 // ~210mm @150dpi
	const targetH = 1754 // ~297mm @150dpi

	// Fill: resize + crop supaya penuh menutupi target, tanpa distorsi
	resized := imaging.Fill(src, targetW, targetH, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	switch ext {
	case ".png":
		err = png.Encode(&buf, resized)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	default:
		return nil, fmt.Errorf("format tidak didukung: %s", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal encode gambar: %w", err)
	}

	return buf.Bytes(), nil
}

func (service *laporanService) buildLaporanPDF(laporan *models.Report, konten *models.ReportCover, bgDepanBytes []byte, bgDepanExt string, bgBelakangBytes []byte, bgBelakangExt string) ([]byte, error) {
	var pdfParts [][]byte

	// ========================================================================
	// 1. COVER DEPAN
	// ========================================================================
	var deskripsiDepanHTML string
	if konten.TextDepan != nil {
		deskripsiDepanHTML = *konten.TextDepan
	}

	deskripsiDepan, err := utils.ConvertCkeditorToPlainText(deskripsiDepanHTML)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses teks deskripsi depan: %w", err)
	}

	pageDepanBytes, err := service.buildPageWithBackground(deskripsiDepan, bgDepanBytes, bgDepanExt)
	if err != nil {
		return nil, fmt.Errorf("gagal buat halaman depan: %w", err)
	}
	pdfParts = append(pdfParts, pageDepanBytes)

	// ========================================================================
	// 2. KATA PENGANTAR
	// ========================================================================
	if konten.KataPengantar != nil && strings.TrimSpace(*konten.KataPengantar) != "" {
		deskripsiKataPengantar, err := utils.ConvertCkeditorToPlainText(*konten.KataPengantar)
		if err == nil && strings.TrimSpace(deskripsiKataPengantar) != "" {
			pageKataPengantarBytes, err := service.buildKataPengantar(deskripsiKataPengantar)
			if err == nil {
				pdfParts = append(pdfParts, pageKataPengantarBytes)
			}
		}
	}

	// ========================================================================
	// 3. SECTION CONTENT (Dinamis: Bab -> SubBab -> Tabel & Narasi)
	// ========================================================================
	if len(laporan.Sections) > 0 {
		pageSectionsBytes, err := service.buildSectionsPDF(laporan)
		if err == nil && len(pageSectionsBytes) > 0 {
			pdfParts = append(pdfParts, pageSectionsBytes)
		}
	}

	// ========================================================================
	// 4. COVER BELAKANG (Penutup)
	// ========================================================================
	var deskripsiBelakangHTML string
	if konten.TextBelakang != nil {
		deskripsiBelakangHTML = *konten.TextBelakang
	}

	deskripsiBelakang, err := utils.ConvertCkeditorToPlainText(deskripsiBelakangHTML)
	if err == nil && (strings.TrimSpace(deskripsiBelakang) != "" || len(bgBelakangBytes) > 0) {
		pageBelakangBytes, err := service.buildPageWithBackground(deskripsiBelakang, bgBelakangBytes, bgBelakangExt)
		if err == nil {
			pdfParts = append(pdfParts, pageBelakangBytes)
		}
	}

	// Merge seluruh bagian PDF
	merged, err := mergePDFs(pdfParts)
	if err != nil {
		return nil, fmt.Errorf("gagal merge pdf: %w", err)
	}

	return merged, nil
}

func (service *laporanService) buildPageWithBackground(textContent string, bgImageBytes []byte, bgExt string) ([]byte, error) {

	// Jika tidak ada teks dan tidak ada background, buat 1 halaman kosong sebagai placeholder
	if textContent == "" && len(bgImageBytes) == 0 {
		return service.buildPlainPage(" ")
	}

	cfgBuilder := buildConfig() // Gunakan builder dasar

	// Jika background tersedia, inject ke builder
	if len(bgImageBytes) > 0 {
		var ext extension.Type
		switch bgExt {
		case ".png":
			ext = extension.Png
		case ".jpg", ".jpeg":
			ext = extension.Jpg // Butuh konversi ke JPEG biasanya jika Maroto tidak support natif WEBP
		default:
			ext = extension.Jpg
		}

		processedBg, err := resizeCoverToA4(bgImageBytes, bgExt)
		if err != nil {
			return nil, fmt.Errorf("gagal resize background: %w", err)
		}

		cfgBuilder = cfgBuilder.WithBackgroundImage(processedBg, ext)
	}

	m := maroto.New(cfgBuilder.Build())

	// Jika teks kosong tapi ada gambar, paksa cetak 1 spasi agar halaman ter-render
	if textContent == "" {
		m.AddRow(10, text.NewCol(12, " "))
	} else {
		// Looping per paragraf
		paragraphs := strings.Split(textContent, "\n\n")
		for _, p := range paragraphs {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			m.AddAutoRow(
				text.NewCol(12, p, props.Text{
					Family: "Tinos",
					Size:   12,
					Align:  align.Left,
				}),
			)
			// Jarak antar paragraf
			m.AddRow(5, text.NewCol(12, " "))
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return doc.GetBytes(), nil
}

func buildConfig() config.Builder {
	fr := fontrepo.New()
	fr.AddUTF8Font("Tinos", fontstyle.Normal, "assets/fonts/Tinos-Regular.ttf")
	fr.AddUTF8Font("Tinos", fontstyle.Bold, "assets/fonts/Tinos-Bold.ttf")
	fr.AddUTF8Font("Tinos", fontstyle.Italic, "assets/fonts/Tinos-Italic.ttf")
	fr.AddUTF8Font("Tinos", fontstyle.BoldItalic, "assets/fonts/Tinos-BoldItalic.ttf")

	customFonts, err := fr.Load()
	if err != nil {
		log.Fatal(err)
	}

	return config.NewBuilder().
		WithTopMargin(15).
		WithLeftMargin(25).
		WithRightMargin(25).
		WithCustomFonts(customFonts)
}

func (service *laporanService) buildKataPengantar(content string) ([]byte, error) {
	cfg := buildConfig().
		Build()

	m := maroto.New(cfg)

	m.AddRow(15,
		text.NewCol(12, "Kata Pengantar", props.Text{
			Size:   18,
			Family: "Tinos",
			Align:  align.Center,
			Style:  fontstyle.Bold,
		}),
	)

	paragraphs := strings.Split(content, "\n\n")
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m.AddRow(5,
			text.NewCol(12, " ", props.Text{
				Family: "Tinos",
				Size:   12,
				Align:  align.Justify,
			}),
		)
		m.AddAutoRow(
			text.NewCol(12, p, props.Text{
				Family: "Tinos",
				Size:   12,
				Align:  align.Justify,
			}),
		)
	}

	m.AddRow(60,
		text.NewCol(12, " "),
	)

	m.AddRow(5,
		text.NewCol(12, "Kota Bandung, 07 November 2026", props.Text{
			Size:   12,
			Family: "Tinos",
			Align:  align.Right,
		}),
	)
	m.AddRow(35,
		text.NewCol(12, " "),
	)
	m.AddRow(5,
		text.NewCol(12, "Walikota Bandung", props.Text{
			Size:   12,
			Family: "Tinos",
			Align:  align.Right,
		}),
	)
	m.AddRow(5,
		text.NewCol(12, "Muhammad Farhan, S.E.", props.Text{
			Size:   12,
			Family: "Tinos",
			Align:  align.Right,
		}),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("gagal generate pdf halaman kedua: %w", err)
	}

	return doc.GetBytes(), nil
}

func (service *laporanService) buildSectionsPDF(laporan *models.Report) ([]byte, error) {
	cfg := buildConfig().Build()
	m := maroto.New(cfg)

	for sIdx, sec := range laporan.Sections {
		// 1. Render Judul Section
		secTitle := fmt.Sprintf("%d. %s", sIdx+1, strings.ToUpper(sec.Title))
		m.AddAutoRow(
			text.NewCol(12, secTitle, props.Text{
				Family: "Tinos",
				Size:   14,
				Style:  fontstyle.Bold,
				Align:  align.Left,
			}),
		)
		m.AddRow(4, text.NewCol(12, " "))

		// 2. Jika Section Langsung Berisi Komponen (Tanpa SubSection)
		if !sec.HasSubSection && len(sec.Components) > 0 {
			for _, comp := range sec.Components {
				service.renderComponent(m, laporan, comp)
			}
		}

		// 3. Jika Section Punya SubSection
		if sec.HasSubSection && len(sec.SubSections) > 0 {
			for subIdx, subSec := range sec.SubSections {
				subTitle := fmt.Sprintf("%d.%d %s", sIdx+1, subIdx+1, subSec.Title)
				m.AddAutoRow(
					text.NewCol(12, subTitle, props.Text{
						Family: "Tinos",
						Size:   12,
						Style:  fontstyle.Bold,
						Align:  align.Left,
					}),
				)
				m.AddRow(3, text.NewCol(12, " "))

				for _, comp := range subSec.Components {
					service.renderComponent(m, laporan, comp)
				}
			}
		}

		m.AddRow(10, text.NewCol(12, " "))
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("gagal generate pdf sections: %w", err)
	}

	return doc.GetBytes(), nil
}

func (service *laporanService) renderComponent(m core.Maroto, laporan *models.Report, comp models.ReportComponent) {
	switch comp.Type {
	case "narrative":
		service.renderNarrativeComponent(m, laporan, comp)
	case "table":
		service.renderTableComponent(m, laporan, comp)
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

func (service *laporanService) renderTableComponent(m core.Maroto, laporan *models.Report, comp models.ReportComponent) {
	if len(comp.TableConfig) == 0 {
		return
	}

	var configPayload request.TableConfigPayload
	if err := json.Unmarshal(comp.TableConfig, &configPayload); err != nil {
		return
	}

	// ---- Kumpulkan FormFieldID ----
	var fieldIDs []int
	var groupSpans []int

	if configPayload.TableStyle == "grouped_header" {
		for _, g := range configPayload.Groups {
			fieldIDs = append(fieldIDs, g.FormFieldIDs...)
			groupSpans = append(groupSpans, len(g.FormFieldIDs))
		}
	} else {
		for _, c := range configPayload.Columns {
			fieldIDs = append(fieldIDs, c.FormFieldID)
		}
	}

	if len(fieldIDs) == 0 {
		return
	}

	fieldLabels, _ := service.laporanRepo.GetFormFieldLabels(fieldIDs)
	tableRows, _ := service.laporanRepo.GetTableResponseData(laporan, fieldIDs)

	// ---- Palet Warna ----
	headerBg := &props.Color{Red: 0, Green: 112, Blue: 192}    // biru header
	whiteText := &props.Color{Red: 255, Green: 255, Blue: 255} // teks putih
	bandBg := &props.Color{Red: 222, Green: 235, Blue: 250}    // biru muda selang-seling
	whiteBg := &props.Color{Red: 255, Green: 255, Blue: 255}
	borderColor := &props.Color{Red: 170, Green: 170, Blue: 170}

	// Helper cell style dengan padding proporsional (margin di sekeliling teks dalam sel)
	cellStyle := func(bg *props.Color) *props.Cell {
		return &props.Cell{
			BackgroundColor: bg,
			BorderType:      border.Full,
			BorderColor:     borderColor,
			BorderThickness: 0.2,
		}
	}

	cellStyleWithBorder := func(bg *props.Color, bType border.Type) *props.Cell {
		return &props.Cell{
			BackgroundColor: bg,
			BorderType:      bType,
			BorderColor:     borderColor,
			BorderThickness: 0.2,
		}
	}

	// PERBAIKAN: TextStyle dibersihkan dari paksaan posisi (Top, Right, Padding).
	// Biarkan Maroto v2 menangani perataannya secara native di dalam Cell.
	textStyle := func(bold bool, color *props.Color, size float64) props.Text {
		st := fontstyle.Normal
		if bold {
			st = fontstyle.Bold
		}
		return props.Text{
			Family: "Tinos",
			Size:   size,
			Style:  st,
			Align:  align.Center,
			Color:  color,
			// Hapus Top dan VerticalPadding. Maroto akan auto-center berdasarkan Cell-nya
		}
	}

	// ---- Hitung Lebar Grid (Total 12 unit) ----
	dataColCount := len(fieldIDs)
	territoryWidth := 3
	if dataColCount > 6 {
		territoryWidth = 2
	}
	if !configPayload.ShowTerritoryCol {
		territoryWidth = 0
	}

	remainingWidth := 12 - territoryWidth
	if remainingWidth < dataColCount {
		remainingWidth = dataColCount
	}

	baseColWidth := remainingWidth / dataColCount
	remainder := remainingWidth % dataColCount

	widths := make([]int, dataColCount)
	for i := 0; i < dataColCount; i++ {
		widths[i] = baseColWidth
	}
	for i := 0; i < remainder; i++ {
		widths[i]++
	}

	sumWidthsBefore := func(arr []int, idx int) int {
		sum := 0
		for i := 0; i < idx; i++ {
			sum += arr[i]
		}
		return sum
	}

	// ============================================================
	// HEADER BARIS 1
	// ============================================================
	var headerRow1 []core.Col
	if configPayload.ShowTerritoryCol {
		if configPayload.TableStyle == "grouped_header" {
			// Karena ini header ganda, kita gunakan trik spasi kosong (\n) agar visualnya pas di tengah
			headerRow1 = append(headerRow1,
				col.New(territoryWidth).
					Add(text.New("\nWilayah", textStyle(true, whiteText, 10))).
					WithStyle(cellStyleWithBorder(headerBg, border.Top)),
			)
		} else {
			headerRow1 = append(headerRow1,
				col.New(territoryWidth).
					Add(text.New("Wilayah", textStyle(true, whiteText, 10))).
					WithStyle(cellStyle(headerBg)),
			)
		}
	}

	if configPayload.TableStyle == "grouped_header" {
		for gi, g := range configPayload.Groups {
			spanWidth := 0
			start := sumWidthsBefore(groupSpans, gi)
			for k := 0; k < groupSpans[gi]; k++ {
				spanWidth += widths[start+k]
			}
			headerRow1 = append(headerRow1,
				col.New(spanWidth).
					Add(text.New(g.GroupName, textStyle(true, whiteText, 10))).
					WithStyle(cellStyle(headerBg)),
			)
		}
	} else {
		for i, c := range configPayload.Columns {
			lbl := fieldLabels[c.FormFieldID]
			if lbl == "" {
				lbl = fmt.Sprintf("#%d", c.FormFieldID)
			}
			headerRow1 = append(headerRow1,
				col.New(widths[i]).
					Add(text.New(lbl, textStyle(true, whiteText, 10))).
					WithStyle(cellStyle(headerBg)),
			)
		}
	}
	m.AddAutoRow(headerRow1...)

	// ============================================================
	// HEADER BARIS 2 — hanya untuk grouped_header
	// ============================================================
	if configPayload.TableStyle == "grouped_header" {
		var headerRow2 []core.Col
		if configPayload.ShowTerritoryCol {
			headerRow2 = append(headerRow2,
				col.New(territoryWidth).
					Add(text.New("", textStyle(true, whiteText, 9))).
					WithStyle(cellStyleWithBorder(headerBg, border.Bottom)),
			)
		}
		for i, fID := range fieldIDs {
			lbl := fieldLabels[fID]
			if lbl == "" {
				lbl = fmt.Sprintf("#%d", fID)
			}
			headerRow2 = append(headerRow2,
				col.New(widths[i]).
					Add(text.New(lbl, textStyle(true, whiteText, 9))).
					WithStyle(cellStyle(headerBg)),
			)
		}
		m.AddAutoRow(headerRow2...)
	}

	// ============================================================
	// BARIS DATA (Body)
	// ============================================================
	for rIdx, row := range tableRows {
		bg := whiteBg
		if rIdx%2 == 1 {
			bg = bandBg
		}

		var dataCols []core.Col
		if configPayload.ShowTerritoryCol {
			dataCols = append(dataCols,
				col.New(territoryWidth).
					// Tambahkan sedikit spasi di awal string wilayah agar tidak mepet ke garis kiri jika panjang
					Add(text.New(" "+row.TerritoryName+" ", textStyle(false, nil, 9))).
					WithStyle(cellStyle(bg)),
			)
		}
		for i, fID := range fieldIDs {
			val := row.Values[fID]
			dataCols = append(dataCols,
				col.New(widths[i]).
					Add(text.New(fmt.Sprintf("%d", val), textStyle(false, nil, 9))).
					WithStyle(cellStyle(bg)),
			)
		}

		// Gunakan m.AddRow dengan perhitungan height cerdas berdasarkan text panjang, bukan AddAutoRow murni
		// Ini trik untuk Maroto v2 agar space rapi tanpa error mepet border.
		calculatedHeight := float64(10)  // Height default
		if len(row.TerritoryName) > 20 { // Sesuaikan batasan karakter
			calculatedHeight = 16 // Tambah height jika kemungkinan wrap
		}
		m.AddRow(calculatedHeight, dataCols...)
	}

	m.AddRow(8, text.NewCol(12, " "))
}

// buildPlainPage — dokumen maroto TERPISAH tanpa background sama sekali
func (service *laporanService) buildPlainPage(content string) ([]byte, error) {
	m := maroto.New() // tanpa config background

	m.AddRow(10,
		text.NewCol(12, content, props.Text{
			Size:  12,
			Align: align.Center,
		}),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("gagal generate pdf halaman kedua: %w", err)
	}

	return doc.GetBytes(), nil
}

// mergePDFs menggabungkan beberapa PDF (dalam bentuk bytes) jadi satu file PDF
func mergePDFs(pdfs [][]byte) ([]byte, error) {
	readers := make([]io.ReadSeeker, len(pdfs))
	for i, p := range pdfs {
		readers[i] = bytes.NewReader(p)
	}

	var buf bytes.Buffer
	if err := api.MergeRaw(readers, &buf, false, nil); err != nil {
		return nil, fmt.Errorf("gagal merge pdf: %w", err)
	}

	return buf.Bytes(), nil
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

	// ========================================================================
	// 5. VALIDASI UNIQUE TITLE (BUSINESS LOGIC)
	// ========================================================================

	// A. Cek apakah Judul Section sudah ada di Laporan ini
	isSectionExists, err := service.laporanRepo.CheckSectionTitleExists(laporanID, payloads.Title)
	if err != nil {
		return utils.SendError(errors.New("Gagal memvalidasi nama section"), http.StatusInternalServerError)
	}
	if isSectionExists {
		return utils.SendError(errors.New("Nama Section sudah digunakan di laporan ini. Silakan gunakan nama lain."), http.StatusConflict)
	}

	// B. Cek duplikasi Judul Sub-Section (Jika ada)
	if payloads.HasSubSection && len(payloads.SubSections) > 0 {

		// Map untuk mengecek duplikasi di dalam payload request itu sendiri
		subMapInternal := make(map[string]bool)

		for _, pSub := range payloads.SubSections {
			subTitleLower := strings.ToLower(pSub.Title)

			// Validasi duplikasi internal (dalam 1 request payload)
			if subMapInternal[subTitleLower] {
				return utils.SendError(fmt.Errorf("Terdapat duplikasi nama Sub-Section ('%s') di dalam form Anda", pSub.Title), http.StatusBadRequest)
			}
			subMapInternal[subTitleLower] = true

			// Validasi duplikasi eksternal (ke Database)
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

	if payloads.HasSubSection && len(payloads.SubSections) > 0 {
		var subSections []models.ReportSubSection

		for _, pSub := range payloads.SubSections {
			subModel := models.ReportSubSection{
				Title:    pSub.Title,
				Sequence: pSub.Sequence,
			}

			var components []models.ReportComponent

			// -- Siapkan Objek Visual (Table / Chart) --
			visualComp := models.ReportComponent{
				Type: pSub.ComponentType,
			}
			if pSub.ComponentConfig != nil {
				visualComp.TableStyle = &pSub.ComponentConfig.TableStyle
				configBytes, _ := json.Marshal(pSub.ComponentConfig)
				visualComp.TableConfig = datatypes.JSON(configBytes)
			}

			// -- Siapkan Objek Narasi --
			var narrativeComp *models.ReportComponent
			if pSub.NarrativeTemplate != nil && pSub.NarrativeLogic != nil {
				logicBytes, _ := json.Marshal(pSub.NarrativeLogic)
				narrativeComp = &models.ReportComponent{
					Type:              "narrative",
					NarrativeTemplate: pSub.NarrativeTemplate,
					NarrativeLogic:    datatypes.JSON(logicBytes),
				}
			}

			// -- Susun Sequence Berdasarkan Posisi Narasi --
			pos := "bottom"
			if pSub.NarrativePosition != nil {
				pos = *pSub.NarrativePosition
			}

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

			subModel.Components = components
			subSections = append(subSections, subModel)
		}
		sectionModel.SubSections = subSections
	}

	// ========================================================================
	// 7. EKSEKUSI DATABASE DENGAN TRANSACTION (PURE CREATE)
	// ========================================================================

	errTx := service.laporanRepo.WithTransaction(ctx, func(txRepo repository.LaporanRepo) error {
		// HANYA CREATE (Tidak ada penghapusan data lama)
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
