package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ManajemenCMSService interface {
	ListSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateStatusSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateNameSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSectionContent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateOrderSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenCMSService struct {
	manajemenCMSRepo repository.ManajemenCMSRepo
	fileRepo         repository.FileRepo
}

func NewManajemenCMSService(
	manajemenCMSRepo repository.ManajemenCMSRepo,
	fileRepo repository.FileRepo,
) ManajemenCMSService {
	return &manajemenCMSService{
		manajemenCMSRepo,
		fileRepo,
	}
}

func (service *manajemenCMSService) ListSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := payloads.DatatablePayload{
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

	data, totalData, err := service.manajemenCMSRepo.GetListSection(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	noEditableContent := map[models.SectionType]bool{
		models.SectionTypeMap:   true,
		models.SectionTypeTable: true,
	}

	for i := range data {
		isSystem := data[i].IsSystem
		hasEditableContent := !noEditableContent[data[i].Type]

		data[i].PosibleUpdate = hasEditableContent
		data[i].PosibleDelete = !isSystem
		data[i].PosibleChangeName = !isSystem

		if isSystem {
			data[i].PosibleChangeOrderUp = false
			data[i].PosibleChangeOrderDown = false
		} else {
			data[i].PosibleChangeOrderUp = (i > 0) && !data[i-1].IsSystem

			data[i].PosibleChangeOrderDown = (i < len(data)-1) && !data[i+1].IsSystem
		}
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total":      totalData,
			"page":       payload.Page,
			"limit":      payload.Limit,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "berhasil mendapatkan list section")
}

func (service *manajemenCMSService) UpdateStatusSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateStatusSectionPayload

	err := utils.DynamicBind(req, &payload)
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

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getSectionById, err := service.manajemenCMSRepo.GetSectionById(id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getSectionById == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	getSectionById.Status = *payload.Status

	_, err = service.manajemenCMSRepo.UpdateSection(getSectionById)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Status section berhasil diperbarui")
}

func (service *manajemenCMSService) UpdateNameSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateNameSectionPayload

	err := utils.DynamicBind(req, &payload)
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

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getSectionById, err := service.manajemenCMSRepo.GetSectionById(id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getSectionById == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	if getSectionById.IsSystem {
		return utils.SendError(errors.New("Section ini tidak dapat diubah"), http.StatusBadRequest)
	}

	getSectionById.Name = payload.Name

	_, err = service.manajemenCMSRepo.UpdateSection(getSectionById)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Status section berhasil diperbarui")
}

func (service *manajemenCMSService) DeleteSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getSectionById, err := service.manajemenCMSRepo.GetSectionById(id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getSectionById == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	if getSectionById.IsSystem {
		return utils.SendError(errors.New("Section ini tidak dapat dihapus"), http.StatusBadRequest)
	}

	err = service.manajemenCMSRepo.DeleteSection(getSectionById)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Section berhasil dihapus")
}

func (service *manajemenCMSService) CreateSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ManajemenCMSPayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	exist, err := service.manajemenCMSRepo.IsNameSectionExist(payload.NamaSection)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if exist {
		return utils.SendError(errors.New("Nama section sudah digunakan"), http.StatusBadRequest)
	}

	baseSlug := utils.StringToSlug(payload.NamaSection, "-")
	finalSectionSlug := baseSlug
	counter := 1

	for {
		exist, err := service.manajemenCMSRepo.IsSlugSectionExist(finalSectionSlug)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}

		if !exist {
			break
		}

		finalSectionSlug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	lastOrder, err := service.manajemenCMSRepo.GetLastSectionOrder()
	if err != nil {
		return utils.SendError(errors.New("Gagal menghitung urutan section"), http.StatusInternalServerError)
	}

	newOrder := lastOrder + 10

	newSection := models.CMSSection{
		Slug:         finalSectionSlug,
		Name:         payload.NamaSection,
		Type:         models.SectionType(payload.TipeSection),
		SectionOrder: newOrder,
		IsRepeatable: true,
		IsSystem:     false,
		IsEditable:   utils.BoolToPointer(true),
		Status:       true,
	}

	err = service.manajemenCMSRepo.CreateSection(&newSection)
	if err != nil {
		return utils.SendError(errors.New("Gagal membuat section baru"), http.StatusInternalServerError)
	}

	return utils.SendData(newSection, "Section berhasil dibuat")
}

func (service *manajemenCMSService) GetSectionBySlug(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	strSlug := slug["slug"].(string)

	getSectionBySlug, err := service.manajemenCMSRepo.GetSectionBySlug(strSlug)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getSectionBySlug == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	for i := range getSectionBySlug.Items {
		var filePath string

		if getSectionBySlug.Items[i].Image != nil {
			_filePath := *getSectionBySlug.Items[i].Image
			filePath = os.Getenv("API_GATEWAY_URL") + "/view-cms-image/" + _filePath
		}
		getSectionBySlug.Items[i].Image = &filePath
	}

	return utils.SendData(getSectionBySlug, "Berhasil mendapatkan section")
}

func (service *manajemenCMSService) UpdateSectionContent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	sectionSlug, ok := slug["slug"].(string)
	if !ok || sectionSlug == "" {
		return utils.SendError(errors.New("Slug section tidak valid"), http.StatusBadRequest)
	}

	getSectionBySlug, err := service.manajemenCMSRepo.GetSectionBySlug(sectionSlug)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getSectionBySlug == nil {
		return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
	}

	if getSectionBySlug.IsEditable != nil && !*getSectionBySlug.IsEditable {
		return utils.SendError(errors.New("Section ini tidak dapat diubah kontennya"), http.StatusForbidden)
	}

	switch getSectionBySlug.Type {
	case models.SectionTypeContent, models.SectionTypeText:
		return service.updateContentSection(ctx, req, getSectionBySlug)
	case models.SectionTypeItems:
		return service.updateItemsSection(ctx, req, getSectionBySlug)
	case models.SectionTypePicture:
		return service.updateMediaSection(ctx, req, getSectionBySlug)
	default:
		return utils.SendError(errors.New("Section ini tidak mendukung pembaruan konten"), http.StatusBadRequest)
	}
}

func (service *manajemenCMSService) updateContentSection(ctx context.Context, req map[string]interface{}, section *models.CMSSection) (*pb.ProxyResponse, error) {
	var payload payloads.UpdateContentSectionPayload

	err := utils.DynamicBind(req, &payload)
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

	contents := make([]models.CMSContent, 0, len(payload.Contents))
	for _, c := range payload.Contents {
		contents = append(contents, models.CMSContent{
			SectionID: section.ID,
			Key:       c.Key,
			ValueText: c.ValueText,
			Status:    true,
		})
	}

	err = service.manajemenCMSRepo.UpsertContents(section.ID, contents)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Konten section berhasil diperbarui")
}

func (service *manajemenCMSService) updateItemsSection(ctx context.Context, req map[string]interface{}, section *models.CMSSection) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

	var payload payloads.UpdateItemsSectionPayload

	err := utils.DynamicBind(req, &payload)
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

	if !section.IsRepeatable && len(payload.Items) > 1 {
		return utils.SendError(errors.New("Section ini hanya boleh memiliki satu item"), http.StatusBadRequest)
	}

	items := make([]models.CMSItem, 0, len(payload.Items))

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	maxSizeInKB := float64(5120)
	var uploadedPaths []string
	var successfullyUploadedFiles []string
	defer func() {
		if r := recover(); r != nil {
			service.fileRepo.DeleteSurveyImageBulk(detachedCtx, successfullyUploadedFiles)
			panic(r)
		}
	}()
	// gatewayURL := os.Getenv("API_GATEWAY_URL") + "/view-cms-image/"
	for _, it := range payload.Items {
		var _uploadedPath *string
		if strings.HasPrefix(*it.Image, "data:") {
			if !strings.HasPrefix(*it.Image, "data:image") {
				return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)"), http.StatusBadRequest)
			}

			base64Data, err := utils.ExtractBase64Info(*it.Image)
			if err != nil {
				return utils.SendError(errors.New("Gagal memproses gambar"), http.StatusBadRequest)
			}

			if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
				return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
			}

			if base64Data.SizeInKB > maxSizeInKB {
				return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
			}

			path, err := service.fileRepo.UploadCMSImage(ctx, it.Image)
			if err != nil {
				return utils.SendError(errors.New("Gagal mengunggah gambar: "+err.Error()), http.StatusInternalServerError)
			}

			if path != nil {
				_uploadedPath = path
				uploadedPaths = append(uploadedPaths, *path)
				successfullyUploadedFiles = append(successfullyUploadedFiles, *path)
			}

		}

		var itTitle string
		if it.Title != nil {
			itTitle = *it.Title
		}
		items = append(items, models.CMSItem{
			SectionID:   section.ID,
			ItemOrder:   it.ItemOrder,
			Title:       itTitle,
			Description: it.Description,
			Category:    it.Category,
			Image:       _uploadedPath,
			Status:      true,
		})
	}

	err = service.manajemenCMSRepo.UpsertItems(section.ID, items)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Konten section berhasil diperbarui")
}

func (service *manajemenCMSService) updateMediaSection(ctx context.Context, req map[string]interface{}, section *models.CMSSection) (*pb.ProxyResponse, error) {
	var payload payloads.UpdateMediaSectionPayload

	err := utils.DynamicBind(req, &payload)
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

	media := make([]models.CMSMedia, 0, len(payload.Media))
	for _, m := range payload.Media {
		media = append(media, models.CMSMedia{
			SectionID: section.ID,
			ItemOrder: m.ItemOrder,
			ImageURL:  m.ImageURL,
			Caption:   m.Caption,
			Status:    true,
		})
	}

	err = service.manajemenCMSRepo.UpsertMedia(section.ID, media)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Konten section berhasil diperbarui")
}

func (service *manajemenCMSService) UpdateOrderSection(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// 1. Ambil Slug section dari URL parameter
	sectionSlug, ok := slug["slug"].(string)
	if !ok || sectionSlug == "" {
		return utils.SendError(errors.New("Slug section tidak valid"), http.StatusBadRequest)
	}

	// 2. Bind payload request
	var payload payloads.UpdateOrderSectionPayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// 3. Validasi payload (direction wajib: "up" atau "down")
	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	// 4. Cari Section yang bersangkutan
	existingSection, err := service.manajemenCMSRepo.GetSectionBySlug(sectionSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Section tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// 5. PROTEKSI KETAT: Section Starter (IsSystem == true) DILARANG di-reorder
	if existingSection.IsSystem {
		return utils.SendError(errors.New("Section starter tidak dapat diubah urutannya"), http.StatusBadRequest)
	}

	// 6. Cari Section Tetangga (khusus non-system) untuk tukar posisi
	adjacentSection, err := service.manajemenCMSRepo.GetAdjacentSectionForReorder(existingSection.SectionOrder, payload.Direction)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Tidak dapat mengubah urutan lagi pada arah tersebut"), http.StatusBadRequest)
		}
		return utils.SendError(errors.New("Terjadi kesalahan saat menentukan posisi section"), http.StatusInternalServerError)
	}

	// 7. Eksekusi swap posisi di DB Transaction
	err = service.manajemenCMSRepo.SwapSectionOrder(existingSection, adjacentSection)
	if err != nil {
		return utils.SendError(errors.New("Gagal memperbarui urutan section"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Urutan section berhasil diperbarui")
}
