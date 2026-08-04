package service

import (
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
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
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
}

type laporanService struct {
	laporanRepo repository.LaporanRepo
	fileRepo    repository.FileRepo
}

func NewLaporanService(
	laporanRepo repository.LaporanRepo,
	fileRepo repository.FileRepo,
) LaporanService {
	return &laporanService{
		laporanRepo: laporanRepo,
		fileRepo:    fileRepo,
	}
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

	bgBytes, mimeType, err := service.fileRepo.GetLaporanKontenImageBytes(ctx, *konten.ImgDepan)
	if err != nil {
		return utils.SendError(errors.New("Gagal mengambil gambar background"), http.StatusInternalServerError)
	}

	ext := ".png"
	if strings.Contains(mimeType, "jpeg") {
		ext = ".jpg"
	}

	pdfBytes, err := service.buildLaporanPDF(laporan, konten, bgBytes, ext)
	if err != nil {
		return utils.SendError(errors.New("Gagal membuat PDF"), http.StatusInternalServerError)
	}

	message := fmt.Sprintf("Data File,%s,%s", mimeType, "Laporan.pdf")
	return utils.SetResponseData(pdfBytes, true, message, http.StatusOK, nil, ""), nil
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

func (service *laporanService) buildLaporanPDF(laporan *models.Report, konten *models.ReportCover, bgDepanBytes []byte, bgDepanExt string) ([]byte, error) {
	// BEGIN: HALAMAN DEPAN ===========================================
	var deskripsiDepanHTML string
	if konten.TextDepan != nil {
		deskripsiDepanHTML = *konten.TextDepan
	}

	deskripsiDepan, err := utils.ConvertCkeditorToPlainText(deskripsiDepanHTML)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses teks deskripsi: %w", err)
	}

	pageDepanBytes, err := service.buildPageWithBackground(deskripsiDepan, bgDepanBytes, bgDepanExt)
	if err != nil {
		return nil, fmt.Errorf("gagal buat halaman depan: %w", err)
	}
	// END: HALAMAN DEPAN ===========================================

	// BEGIN: KATA PENGANTAR ===========================================
	var deskripsiKataPengantarHTML string
	if konten.KataPengantar != nil {
		deskripsiKataPengantarHTML = *konten.KataPengantar
	}

	deskripsiKataPengantar, err := utils.ConvertCkeditorToPlainText(deskripsiKataPengantarHTML)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses teks deskripsi: %w", err)
	}
	pageKataPengantarBytes, err := service.buildKataPengantar(deskripsiKataPengantar)
	if err != nil {
		return nil, fmt.Errorf("gagal buat kata pengantar: %w", err)
	}
	// END: KATA PENGANTAR ===========================================

	// ==== HALAMAN 2: tanpa background, isi "test" ====
	pageKeduaBytes, err := service.buildPlainPage("test")
	if err != nil {
		return nil, fmt.Errorf("gagal buat halaman kedua: %w", err)
	}

	// ==== MERGE semua halaman jadi 1 PDF ====
	merged, err := mergePDFs([][]byte{pageDepanBytes, pageKataPengantarBytes, pageKeduaBytes})
	if err != nil {
		return nil, fmt.Errorf("gagal merge pdf: %w", err)
	}

	return merged, nil
}

func (service *laporanService) buildPageWithBackground(textContent string, bgImageBytes []byte, bgExt string) ([]byte, error) {
	var ext extension.Type
	switch bgExt {
	case ".png":
		ext = extension.Png
	case ".jpg", ".jpeg":
		ext = extension.Jpg
	default:
		return nil, fmt.Errorf("format background tidak didukung: %s", bgExt)
	}

	processedBg, err := resizeCoverToA4(bgImageBytes, bgExt)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses gambar background: %w", err)
	}

	cfg := config.NewBuilder().
		WithBackgroundImage(processedBg, ext).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	paragraphs := strings.Split(textContent, "\n\n")
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m.AddAutoRow(
			text.NewCol(12, "\n", props.Text{
				Size:  12,
				Align: align.Left,
			}),
		)
		m.AddAutoRow(
			text.NewCol(12, p, props.Text{
				Size:  12,
				Align: align.Left,
			}),
		)
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("gagal generate pdf halaman depan: %w", err)
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
