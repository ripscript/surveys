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
	case "simple":
		if len(cfg.Columns) == 0 {
			return errors.New("columns wajib diisi jika table_style = simple")
		}
	}
	return nil
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
		if err := validateTableConfig(payloads.ComponentConfig); err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	} else {
		for _, pSub := range payloads.SubSections {
			if err := validateTableConfig(pSub.ComponentConfig); err != nil {
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
		// component_type sudah divalidasi wajib oleh tag required_if di atas
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

			subModel.Components = buildVisualAndNarrativeComponents(
				pSub.ComponentType,
				pSub.ComponentConfig,
				pSub.NarrativePosition,
				pSub.NarrativeTemplate,
				pSub.NarrativeLogic,
			)

			subSections = append(subSections, subModel)
		}
		sectionModel.SubSections = subSections

	} else {
		// --- Pola baru: section flat, component nempel langsung ke section ---
		sectionModel.Components = buildVisualAndNarrativeComponents(
			payloads.ComponentType,
			payloads.ComponentConfig,
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
	narrativePosition *string,
	narrativeTemplate *string,
	narrativeLogic *request.NarrativeLogicPayload,
) []models.ReportComponent {

	visualComp := models.ReportComponent{
		Type: componentType,
	}
	if componentConfig != nil {
		visualComp.TableStyle = &componentConfig.TableStyle
		configBytes, _ := json.Marshal(componentConfig)
		visualComp.TableConfig = datatypes.JSON(configBytes)
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
