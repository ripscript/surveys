package service

import (
	"backend/masterapi/assets"
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	_ "embed"
	"encoding/json"
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
	GetLandingPage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GeoJsonKotaBandungLevelKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

	if getSectionById.Name != payload.Name {
		exist, err := service.manajemenCMSRepo.IsNameSectionExist(payload.Name)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}

		if exist {
			return utils.SendError(errors.New("Nama section sudah digunakan"), http.StatusBadRequest)
		}
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

	filesToDelete, err := service.manajemenCMSRepo.DeleteSection(getSectionById)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server saat menghapus section, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if len(filesToDelete) > 0 {
		detachedCtx := context.WithoutCancel(ctx)
		service.fileRepo.DeleteCMSImageBulk(detachedCtx, filesToDelete)
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

	createdSection, err := service.manajemenCMSRepo.CreateSection(&newSection)
	if err != nil {
		return utils.SendError(errors.New("Gagal membuat section baru"), http.StatusInternalServerError)
	}

	if createdSection.Type == models.SectionTypeText {
		content := models.CMSContent{
			SectionID: createdSection.ID,
			Key:       "konten",
			Status:    true,
		}
		_, err = service.manajemenCMSRepo.CreateCMSContent(&content)
		if err != nil {
			return utils.SendError(errors.New("Gagal membuat konten untuk section baru"), http.StatusInternalServerError)
		}
	}

	return utils.SendData(nil, "Section berhasil dibuat")
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

	for i := range getSectionBySlug.Contents {
		var filePath string

		if getSectionBySlug.Contents[i].Key == "button_link" {
			if getSectionBySlug.Contents[i].ValueText != nil && *getSectionBySlug.Contents[i].ValueText != "" {
				_filePath := *getSectionBySlug.Contents[i].ValueText
				filePath = os.Getenv("API_GATEWAY_URL") + "/view-cms-image/" + _filePath
			}
			getSectionBySlug.Contents[i].ValueText = &filePath
		}
	}

	for i := range getSectionBySlug.Items {
		var filePath string

		if getSectionBySlug.Items[i].Image != nil {
			_filePath := *getSectionBySlug.Items[i].Image
			filePath = os.Getenv("API_GATEWAY_URL") + "/view-cms-image/" + _filePath
		}
		getSectionBySlug.Items[i].Image = &filePath
	}

	for i := range getSectionBySlug.Media {
		var filePath string

		_filePath := getSectionBySlug.Media[i].ImageURL
		filePath = os.Getenv("API_GATEWAY_URL") + "/view-cms-image/" + _filePath
		getSectionBySlug.Media[i].ImageURL = filePath
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
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

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

	// Ambil content lama dari DB untuk keperluan cleanup file
	existingContents, err := service.manajemenCMSRepo.GetContentsBySectionID(section.ID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	existingByKey := make(map[string]*models.CMSContent, len(existingContents))
	for i := range existingContents {
		existingByKey[existingContents[i].Key] = &existingContents[i]
	}

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp", "application/pdf"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif", ".pdf"}
	maxSizeInKB := float64(5120)

	var uploadedFiles []string
	rollbackUploadedFiles := func() {
		if len(uploadedFiles) > 0 {
			service.fileRepo.DeleteCMSImageBulk(detachedCtx, uploadedFiles)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			rollbackUploadedFiles()
			panic(r)
		}
	}()

	var oldPathsToDelete []string

	contents := make([]models.CMSContent, 0, len(payload.Contents))
	for _, c := range payload.Contents {
		if c.Key == "button_link" {
			old, hasOld := existingByKey[c.Key]
			var oldValue *string
			if hasOld {
				oldValue = old.ValueText
			}

			switch {
			case c.ValueText == nil || *c.ValueText == "":
				if oldValue != nil && *oldValue != "" {
					oldPathsToDelete = append(oldPathsToDelete, *oldValue)
				}

			case strings.HasPrefix(*c.ValueText, "data:"):
				if !strings.HasPrefix(*c.ValueText, "data:image") && !strings.HasPrefix(*c.ValueText, "data:application/pdf") {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP, PDF)"), http.StatusBadRequest)
				}

				base64Data, err := utils.ExtractBase64Info(*c.ValueText)
				if err != nil {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Gagal memproses gambar"), http.StatusBadRequest)
				}
				if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
				}
				if base64Data.SizeInKB > maxSizeInKB {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
				}

				path, err := service.fileRepo.UploadCMSImage(ctx, c.ValueText)
				if err != nil || path == nil {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Gagal mengunggah gambar"), http.StatusInternalServerError)
				}

				if oldValue != nil && *oldValue != "" {
					oldPathsToDelete = append(oldPathsToDelete, *oldValue)
				}

				c.ValueText = path
				uploadedFiles = append(uploadedFiles, *path)

			default:
				// bukan data URI -> dianggap link lama, TIDAK diupload ulang.
				relativePath := toRelativeImagePath(*c.ValueText)
				c.ValueText = &relativePath
			}

		}
		contents = append(contents, models.CMSContent{
			SectionID: section.ID,
			Key:       c.Key,
			ValueText: c.ValueText,
			Status:    true,
		})
	}

	err = service.manajemenCMSRepo.UpsertContents(section.ID, contents)
	if err != nil {
		rollbackUploadedFiles()
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if len(oldPathsToDelete) > 0 {
		service.fileRepo.DeleteCMSImageBulk(detachedCtx, oldPathsToDelete)
	}

	return utils.SendData(nil, "Konten section berhasil diperbarui")
}

const cmsImagePathMarker = "/view-cms-image/"

func toRelativeImagePath(image string) string {
	_, after, ok := strings.Cut(image, cmsImagePathMarker)
	if !ok {
		return image
	}
	return after
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

	// gambar lama per item_order -> dipakai untuk tahu file mana yang harus
	// dihapus kalau item tersebut diganti gambar barunya
	oldImageByOrder := make(map[int]string, len(section.Items))
	for _, it := range section.Items {
		if it.Image != nil && *it.Image != "" {
			oldImageByOrder[it.ItemOrder] = *it.Image
		}
	}

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	maxSizeInKB := float64(5120)

	var uploadedFiles []string
	rollbackUploadedFiles := func() {
		if len(uploadedFiles) > 0 {
			service.fileRepo.DeleteCMSImageBulk(detachedCtx, uploadedFiles)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			rollbackUploadedFiles()
			panic(r)
		}
	}()

	items := make([]models.CMSItem, 0, len(payload.Items))
	var oldPathsToDelete []string

	for _, it := range payload.Items {
		old, hadOldImage := oldImageByOrder[it.ItemOrder]

		item := models.CMSItem{
			SectionID:   section.ID,
			ItemOrder:   it.ItemOrder,
			Description: it.Description,
			Category:    it.Category,
			Status:      true,
		}
		if it.Title != nil {
			item.Title = *it.Title
		}

		switch {
		case it.Image == nil || *it.Image == "":
			// tidak ada gambar sama sekali untuk item ini.
			// kalau sebelumnya ada gambar, berarti sengaja dihapus user -> hapus filenya juga.
			if hadOldImage {
				oldPathsToDelete = append(oldPathsToDelete, old)
			}

		case strings.HasPrefix(*it.Image, "data:"):
			// selalu berarti gambar BARU -> upload, replace yang lama
			if !strings.HasPrefix(*it.Image, "data:image") {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)"), http.StatusBadRequest)
			}

			base64Data, err := utils.ExtractBase64Info(*it.Image)
			if err != nil {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Gagal memproses gambar"), http.StatusBadRequest)
			}
			if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
			}
			if base64Data.SizeInKB > maxSizeInKB {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
			}

			path, err := service.fileRepo.UploadCMSImage(ctx, it.Image)
			if err != nil || path == nil {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Gagal mengunggah gambar"), http.StatusInternalServerError)
			}

			item.Image = path
			uploadedFiles = append(uploadedFiles, *path)

			if hadOldImage && old != *path {
				oldPathsToDelete = append(oldPathsToDelete, old)
			}

		default:
			// bukan data URI -> dianggap link lama, TIDAK diupload ulang.
			// normalisasi balik ke path relatif supaya konsisten dengan yang tersimpan.
			relativePath := toRelativeImagePath(*it.Image)
			item.Image = &relativePath
		}

		items = append(items, item)
	}

	if err := service.manajemenCMSRepo.UpsertItems(section.ID, items); err != nil {
		rollbackUploadedFiles()
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if len(oldPathsToDelete) > 0 {
		service.fileRepo.DeleteCMSImageBulk(detachedCtx, oldPathsToDelete)
	}

	return utils.SendData(nil, "Konten section berhasil diperbarui")
}

func (service *manajemenCMSService) updateMediaSection(ctx context.Context, req map[string]interface{}, section *models.CMSSection) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

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

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	maxSizeInKB := float64(5120)

	var uploadedFiles []string
	rollbackUploadedFiles := func() {
		if len(uploadedFiles) > 0 {
			service.fileRepo.DeleteCMSImageBulk(detachedCtx, uploadedFiles)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			rollbackUploadedFiles()
			panic(r)
		}
	}()

	media := make([]models.CMSMedia, 0, len(payload.Media))

	for _, m := range payload.Media {
		var finalPath string

		switch {
		case strings.HasPrefix(m.ImageURL, "data:"):
			if !strings.HasPrefix(m.ImageURL, "data:image") {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)"), http.StatusBadRequest)
			}

			base64Data, err := utils.ExtractBase64Info(m.ImageURL)
			if err != nil {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Gagal memproses gambar"), http.StatusBadRequest)
			}
			if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
			}
			if base64Data.SizeInKB > maxSizeInKB {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
			}

			path, err := service.fileRepo.UploadCMSImage(ctx, &m.ImageURL)
			if err != nil || path == nil {
				rollbackUploadedFiles()
				return utils.SendError(errors.New("Gagal mengunggah gambar"), http.StatusInternalServerError)
			}

			finalPath = *path
			uploadedFiles = append(uploadedFiles, *path)

		default:
			finalPath = toRelativeImagePath(m.ImageURL)
		}

		media = append(media, models.CMSMedia{
			SectionID: section.ID,
			ItemOrder: m.ItemOrder,
			ImageURL:  finalPath,
			Caption:   m.Caption,
			Status:    true,
		})
	}

	removedImagePaths, err := service.manajemenCMSRepo.SyncMedia(section.ID, media)
	if err != nil {
		rollbackUploadedFiles()
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	var pathsToDelete []string
	for _, p := range removedImagePaths {
		if p != "" && !slices.Contains(uploadedFiles, p) {
			pathsToDelete = append(pathsToDelete, p)
		}
	}
	if len(pathsToDelete) > 0 {
		service.fileRepo.DeleteCMSImageBulk(detachedCtx, pathsToDelete)
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

func (service *manajemenCMSService) GetLandingPage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// 1. Ambil data mentah dari repository
	sections, err := service.manajemenCMSRepo.GetLandingPageSections()
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan saat mengambil data landing page"), http.StatusInternalServerError)
	}

	apiGatewayURL := os.Getenv("API_GATEWAY_URL")
	baseImageURL := apiGatewayURL + "/view-cms-image/"

	// 2. Buat array baru untuk menampung data yang sudah diformat secara dinamis
	var responseData []map[string]interface{}

	for i := range sections {
		// Buat base object untuk setiap section (hanya ambil field yang relevan untuk Frontend)
		secData := map[string]interface{}{
			"id":            sections[i].ID,
			"slug":          sections[i].Slug,
			"name":          sections[i].Name,
			"description":   sections[i].Description,
			"type":          sections[i].Type,
			"section_order": sections[i].SectionOrder,
		}

		// Sisipkan data relasi sesuai dengan Tipenya (Grouped)
		switch sections[i].Type {

		case models.SectionTypeContent, models.SectionTypeText:
			// Ubah array of database rows menjadi satu Object Key-Value mentah
			contentMap := make(map[string]interface{})

			for _, c := range sections[i].Contents {
				// Jika tipe datanya gambar, gabungkan dengan base URL
				if c.ValueImage != nil && *c.ValueImage != "" {
					contentMap[c.Key] = baseImageURL + *c.ValueImage
				} else if c.ValueText != nil {
					contentMap[c.Key] = *c.ValueText
				} else {
					contentMap[c.Key] = ""
				}
			}
			secData["content"] = contentMap

		case models.SectionTypeItems:
			// Perbaiki path gambar pada array Items
			for j := range sections[i].Items {
				if sections[i].Items[j].Image != nil && *sections[i].Items[j].Image != "" {
					fullPath := baseImageURL + *sections[i].Items[j].Image
					sections[i].Items[j].Image = &fullPath
				}
				if sections[i].Items[j].Icon != nil && *sections[i].Items[j].Icon != "" {
					fullPath := baseImageURL + *sections[i].Items[j].Icon
					sections[i].Items[j].Icon = &fullPath
				}
			}
			secData["items"] = sections[i].Items

		case models.SectionTypePicture:
			// Perbaiki path gambar pada array Media
			for k := range sections[i].Media {
				if sections[i].Media[k].ImageURL != "" {
					fullPath := baseImageURL + sections[i].Media[k].ImageURL
					sections[i].Media[k].ImageURL = fullPath
				}
			}
			secData["media"] = sections[i].Media

		case models.SectionTypeMap, models.SectionTypeTable:
			// Tipe statis: Jangan mengirimkan key items, media, atau content
			// Frontend cukup mengecek tipe atau slug-nya saja
		}

		responseData = append(responseData, secData)
	}

	return utils.SendData(responseData, "Berhasil mendapatkan data landing page")
}

func (service *manajemenCMSService) GeoJsonKotaBandungLevelKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var geoData map[string]interface{}

	// Panggil data byte-nya dari package assets yang tadi kita buat
	err := json.Unmarshal(assets.GeoJsonKotaBandungLevelKecamatan, &geoData)
	if err != nil {
		return utils.SendError(errors.New("Gagal memproses data peta wilayah"), http.StatusInternalServerError)
	}

	return utils.SendData(geoData, "Berhasil mendapatkan GeoJSON Kota Bandung Level Kecamatan")
}
