package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ArtikelService interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PublicList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailArtikelPublic(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type artikelService struct {
	fileRepo            repository.FileRepo
	artikelRepo         repository.ArtikelRepository
	artikelCategoryRepo repository.ArtikelCategoryRepository
}

func NewArtikelService(
	fileRepo repository.FileRepo,
	artikelRepo repository.ArtikelRepository,
	artikelCategoryRepo repository.ArtikelCategoryRepository,
) ArtikelService {
	return &artikelService{
		fileRepo:            fileRepo,
		artikelRepo:         artikelRepo,
		artikelCategoryRepo: artikelCategoryRepo,
	}
}

func (service *artikelService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ArtikelPayload
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

	getByName, err := service.artikelRepo.GetByName(ctx, payload.Judul)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByName != nil {
		return utils.SendError(errors.New("Artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	kategori, err := service.artikelCategoryRepo.GetByID(ctx, int(payload.Kategori))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if kategori == nil {
		return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusBadRequest)
	}

	createData := models.Artikel{
		Judul:             &payload.Judul,
		ArtikelCategoryID: utils.Int64ToPointer(int64(kategori.ID)),
		CreatedBy:         &usr.ID,
	}

	createdData, err := service.artikelRepo.Create(ctx, &createData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createdData, "Berhasil membuat artikel")
}

const artikelKontenPathMarker = "/view-artikel-konten/"

func toRelativeArtikelPath(value string) string {
	_, after, ok := strings.Cut(value, os.Getenv("API_GATEWAY_URL")+artikelKontenPathMarker)
	if !ok {
		return value
	}
	return after
}

func toFullArtikelPath(value string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, os.Getenv("API_GATEWAY_URL")+artikelKontenPathMarker) {
		return value
	}
	return os.Getenv("API_GATEWAY_URL") + artikelKontenPathMarker + value
}

func (service *artikelService) DetailArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	artikel, err := service.artikelRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if artikel.CreatedBy == nil || *artikel.CreatedBy != usr.ID {
		return utils.SendError(errors.New("Anda tidak memiliki akses untuk melihat artikel ini"), http.StatusForbidden)
	}

	var categoryID int64
	var categoryName string
	if artikel.ArtikelCategoryID != nil {
		categoryID = *artikel.ArtikelCategoryID
		kategori, err := service.artikelCategoryRepo.GetByID(ctx, int(categoryID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if kategori != nil {
			categoryName = kategori.Name
		}
	}

	var thumbnail string
	if artikel.Thumbnail != nil {
		thumbnail = toFullArtikelPath(*artikel.Thumbnail)
	}

	var storedContents []models.ArtikelContentItem
	if artikel.Contents != nil && len(*artikel.Contents) > 0 {
		if err := json.Unmarshal(*artikel.Contents, &storedContents); err != nil {
			return utils.SendError(errors.New("Gagal memproses konten artikel"), http.StatusInternalServerError)
		}
	}

	responseContents := make([]models.ArtikelContentResponse, 0, len(storedContents))
	for _, c := range storedContents {
		item := models.ArtikelContentResponse{
			Description: c.Description,
			Placeholder: c.Placeholder,
		}
		if c.Image != nil && *c.Image != "" {
			full := toFullArtikelPath(*c.Image)
			item.Image = &full
		}
		if c.File != nil && *c.File != "" {
			full := toFullArtikelPath(*c.File)
			item.File = &full
		}
		responseContents = append(responseContents, item)
	}

	var title string
	if artikel.Judul != nil {
		title = *artikel.Judul
	}

	responseData := models.ArtikelDetailResponse{
		ID:           artikel.ID,
		Title:        title,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Thumbnail:    thumbnail,
		Contents:     responseContents,
	}

	return utils.SendData(responseData, "Berhasil mengambil detail artikel")
}

type uploadCache struct {
	cache         map[string]string
	uploadedPaths []string
}

func newUploadCache() *uploadCache {
	return &uploadCache{cache: make(map[string]string)}
}

func (u *uploadCache) getOrUpload(ctx context.Context, fileRepo repository.FileRepo, datauri string) (string, error) {
	if path, ok := u.cache[datauri]; ok {
		return path, nil
	}
	path, err := fileRepo.UploadArtikelKonten(ctx, &datauri)
	if err != nil || path == nil {
		return "", errors.New("Gagal mengunggah file")
	}
	u.cache[datauri] = *path
	u.uploadedPaths = append(u.uploadedPaths, *path)
	return *path, nil
}

func (u *uploadCache) rollback(ctx context.Context, fileRepo repository.FileRepo) {
	if len(u.uploadedPaths) > 0 {
		fileRepo.DeleteArtikelKontenBulk(context.WithoutCancel(ctx), u.uploadedPaths)
	}
}

func (service *artikelService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var payload payloads.UpdateArtikelPayload
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	existingArtikel, err := service.artikelRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if existingArtikel.CreatedBy == nil || *existingArtikel.CreatedBy != usr.ID {
		return utils.SendError(errors.New("Anda tidak memiliki akses untuk mengubah artikel ini"), http.StatusForbidden)
	}

	oldPaths := make(map[string]bool)
	if existingArtikel.Thumbnail != nil && *existingArtikel.Thumbnail != "" {
		oldPaths[*existingArtikel.Thumbnail] = true
	}
	var oldContents []models.ArtikelContentItem
	if existingArtikel.Contents != nil && len(*existingArtikel.Contents) > 0 {
		_ = json.Unmarshal(*existingArtikel.Contents, &oldContents)
	}
	for _, c := range oldContents {
		if c.Image != nil && *c.Image != "" {
			oldPaths[*c.Image] = true
		}
		if c.File != nil && *c.File != "" {
			oldPaths[*c.File] = true
		}
	}

	imageMime := []string{"image/png", "image/jpg", "image/jpeg"}
	imageExt := []string{".png", ".jpg", ".jpeg"}
	maxImageSizeKB := float64(2048)

	fileMime := []string{
		"image/png", "image/jpg", "image/jpeg",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	fileExt := []string{".png", ".jpg", ".jpeg", ".pdf", ".doc", ".docx"}
	maxFileSizeKB := float64(2048)

	uploads := newUploadCache()
	defer func() {
		if r := recover(); r != nil {
			uploads.rollback(ctx, service.fileRepo)
			panic(r)
		}
	}()

	finalThumbnail := payload.Thumbnail
	if strings.HasPrefix(payload.Thumbnail, "data:") {
		if !strings.HasPrefix(payload.Thumbnail, "data:image") {
			return utils.SendError(errors.New("Format thumbnail tidak didukung. Hanya menerima JPEG, JPG, atau PNG"), http.StatusBadRequest)
		}
		base64Data, err := utils.ExtractBase64Info(payload.Thumbnail)
		if err != nil {
			return utils.SendError(errors.New("Gagal memproses thumbnail"), http.StatusBadRequest)
		}
		if !slices.Contains(imageExt, base64Data.Extension) || !slices.Contains(imageMime, base64Data.MimeType) {
			return utils.SendError(errors.New("Format thumbnail tidak didukung. Hanya menerima JPEG, JPG, atau PNG"), http.StatusBadRequest)
		}
		if base64Data.SizeInKB > maxImageSizeKB {
			return utils.SendError(errors.New("Ukuran thumbnail tidak boleh melebihi 2MB"), http.StatusBadRequest)
		}
		path, err := uploads.getOrUpload(ctx, service.fileRepo, payload.Thumbnail)
		if err != nil {
			uploads.rollback(ctx, service.fileRepo)
			return utils.SendError(errors.New("Gagal mengunggah thumbnail"), http.StatusInternalServerError)
		}
		finalThumbnail = path
	} else {
		finalThumbnail = toRelativeArtikelPath(payload.Thumbnail)
	}

	newContents := make([]models.ArtikelContentItem, 0, len(payload.Contents))

	for _, c := range payload.Contents {
		item := models.ArtikelContentItem{
			Description: c.Description,
			Placeholder: c.Placeholder,
		}

		if c.Image != nil && *c.Image != "" {
			if strings.HasPrefix(*c.Image, "data:") {
				if !strings.HasPrefix(*c.Image, "data:image") {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Format gambar konten tidak didukung. Hanya menerima JPEG, JPG, atau PNG"), http.StatusBadRequest)
				}
				base64Data, err := utils.ExtractBase64Info(*c.Image)
				if err != nil {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Gagal memproses gambar konten"), http.StatusBadRequest)
				}
				if !slices.Contains(imageExt, base64Data.Extension) || !slices.Contains(imageMime, base64Data.MimeType) {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Format gambar konten tidak didukung. Hanya menerima JPEG, JPG, atau PNG"), http.StatusBadRequest)
				}
				if base64Data.SizeInKB > maxImageSizeKB {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Ukuran gambar konten tidak boleh melebihi 2MB"), http.StatusBadRequest)
				}
				path, err := uploads.getOrUpload(ctx, service.fileRepo, *c.Image)
				if err != nil {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Gagal mengunggah gambar konten"), http.StatusInternalServerError)
				}
				item.Image = &path
			} else {
				relative := toRelativeArtikelPath(*c.Image)
				item.Image = &relative
			}
		}

		if c.File != nil && *c.File != "" {
			if strings.HasPrefix(*c.File, "data:") {
				base64Data, err := utils.ExtractBase64Info(*c.File)
				if err != nil {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Gagal memproses file konten"), http.StatusBadRequest)
				}

				if !slices.Contains(fileExt, base64Data.Extension) || !slices.Contains(fileMime, base64Data.MimeType) {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Format file konten tidak didukung. Hanya menerima JPEG, JPG, PNG, PDF, DOC, atau DOCX"), http.StatusBadRequest)
				}
				if base64Data.SizeInKB > maxFileSizeKB {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Ukuran file konten tidak boleh melebihi 2MB"), http.StatusBadRequest)
				}
				path, err := uploads.getOrUpload(ctx, service.fileRepo, *c.File)
				if err != nil {
					uploads.rollback(ctx, service.fileRepo)
					return utils.SendError(errors.New("Gagal mengunggah file konten"), http.StatusInternalServerError)
				}
				item.File = &path
			} else {
				relative := toRelativeArtikelPath(*c.File)
				item.File = &relative
			}
		}

		newContents = append(newContents, item)
	}

	contentsJSON, err := json.Marshal(newContents)
	if err != nil {
		uploads.rollback(ctx, service.fileRepo)
		return utils.SendError(errors.New("Gagal memproses konten artikel"), http.StatusInternalServerError)
	}
	newContentsJSON := datatypes.JSON(contentsJSON)

	existingArtikel.Judul = &payload.Title
	existingArtikel.ArtikelCategoryID = &payload.CategoryID
	existingArtikel.Thumbnail = &finalThumbnail
	existingArtikel.Contents = &newContentsJSON

	_, err = service.artikelRepo.Update(ctx, existingArtikel)
	if err != nil {
		uploads.rollback(ctx, service.fileRepo)
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	keepPaths := make(map[string]bool)
	keepPaths[finalThumbnail] = true
	for _, c := range newContents {
		if c.Image != nil {
			keepPaths[*c.Image] = true
		}
		if c.File != nil {
			keepPaths[*c.File] = true
		}
	}

	var pathsToDelete []string
	for oldPath := range oldPaths {
		if !keepPaths[oldPath] {
			pathsToDelete = append(pathsToDelete, oldPath)
		}
	}
	if len(pathsToDelete) > 0 {
		service.fileRepo.DeleteArtikelKontenBulk(detachedCtx, pathsToDelete)
	}

	return utils.SendData(nil, "Artikel berhasil diperbarui")
}

func (service *artikelService) DeleteArtikel(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	detachedCtx := context.WithoutCancel(ctx)

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	artikel, err := service.artikelRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if artikel.CreatedBy == nil || *artikel.CreatedBy != usr.ID {
		return utils.SendError(errors.New("Anda tidak memiliki akses untuk menghapus artikel ini"), http.StatusForbidden)
	}

	pathsToDeleteSet := make(map[string]bool)
	if artikel.Thumbnail != nil && *artikel.Thumbnail != "" {
		pathsToDeleteSet[*artikel.Thumbnail] = true
	}

	var contents []models.ArtikelContentItem
	if artikel.Contents != nil && len(*artikel.Contents) > 0 {
		if err := json.Unmarshal(*artikel.Contents, &contents); err != nil {
			return utils.SendError(errors.New("Gagal memproses konten artikel"), http.StatusInternalServerError)
		}
	}
	for _, c := range contents {
		if c.Image != nil && *c.Image != "" {
			pathsToDeleteSet[*c.Image] = true
		}
		if c.File != nil && *c.File != "" {
			pathsToDeleteSet[*c.File] = true
		}
	}

	if err := service.artikelRepo.Delete(ctx, id); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if len(pathsToDeleteSet) > 0 {
		pathsToDelete := make([]string, 0, len(pathsToDeleteSet))
		for p := range pathsToDeleteSet {
			pathsToDelete = append(pathsToDelete, p)
		}
		service.fileRepo.DeleteArtikelKontenBulk(detachedCtx, pathsToDelete)
	}

	return utils.SendData(nil, "Artikel berhasil dihapus")
}

func (service *artikelService) GetOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	var parsedIDs []int64
	for _, rawID := range rawIDs {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
			parsedIDs = append(parsedIDs, id)
		}
	}

	_req := payloads.ArtikelOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := service.artikelRepo.GetOptions(_req, usr.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi kategori artikel")
}

func (service *artikelService) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	data, totalData, err := service.artikelRepo.GetList(payload, usr.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
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

	return utils.SendData(result, "Berhasil mengambil list artikel")
}

func (service *artikelService) PublicList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	categoryId := param.Get("category_id")
	var parsedCategoryId *int64
	if categoryId != "" {
		if id, err := strconv.ParseInt(categoryId, 10, 64); err == nil {
			parsedCategoryId = &id
		} else {
			return utils.SendError(errors.New("Invalid category_id parameter"), http.StatusBadRequest)
		}
	}

	data, err := service.artikelRepo.PublicList(parsedCategoryId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for i := range data {
		if data[i].Thumbnail != nil {
			fullPath := toFullArtikelPath(*data[i].Thumbnail)
			data[i].Thumbnail = &fullPath
		}
	}

	return utils.SendData(data, "Berhasil mengambil list artikel")
}

func (service *artikelService) DetailArtikelPublic(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	artikel, err := service.artikelRepo.GetByIdPublic(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	var categoryID int64
	var categoryName string
	if artikel.ArtikelCategoryID != nil {
		categoryID = *artikel.ArtikelCategoryID
		kategori, err := service.artikelCategoryRepo.GetByID(ctx, int(categoryID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if kategori != nil {
			categoryName = kategori.Name
		}
	}

	var thumbnail string
	if artikel.Thumbnail != nil {
		thumbnail = toFullArtikelPath(*artikel.Thumbnail)
	}

	var storedContents []models.ArtikelContentItem
	if artikel.Contents != nil && len(*artikel.Contents) > 0 {
		if err := json.Unmarshal(*artikel.Contents, &storedContents); err != nil {
			return utils.SendError(errors.New("Gagal memproses konten artikel"), http.StatusInternalServerError)
		}
	}

	responseContents := make([]models.ArtikelContentResponse, 0, len(storedContents))
	for _, c := range storedContents {
		item := models.ArtikelContentResponse{
			Description: c.Description,
			Placeholder: c.Placeholder,
		}
		if c.Image != nil && *c.Image != "" {
			full := toFullArtikelPath(*c.Image)
			item.Image = &full
		}
		if c.File != nil && *c.File != "" {
			full := toFullArtikelPath(*c.File)
			item.File = &full
		}
		responseContents = append(responseContents, item)
	}

	var title string
	if artikel.Judul != nil {
		title = *artikel.Judul
	}

	responseData := models.ArtikelDetailResponse{
		ID:           artikel.ID,
		Title:        title,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Thumbnail:    thumbnail,
		Contents:     responseContents,
	}

	return utils.SendData(responseData, "Berhasil mengambil detail artikel")
}
