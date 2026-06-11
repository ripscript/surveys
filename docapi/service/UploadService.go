package service

import (
	"backend/docapi/enums"
	"backend/docapi/models"
	"backend/docapi/utils"
	"backend/siccore/pb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/minio/minio-go"
)

type UploadService interface {
	UploadFile(usr models.JwtCustomClaims, req map[string]interface{}, param url.Values, path string, module string) (*pb.ProxyResponse, error)
	Show(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UploadSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ShowPublicSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteBulkSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type uploadService struct {
	uploadRepo string
	// uploadRepo repository.UploadRepo
}

func NewUploadService(
	uploadRepo string,
) UploadService {
	return &uploadService{
		uploadRepo,
	}
}

func (service *uploadService) UploadFile(usr models.JwtCustomClaims, req map[string]interface{}, param url.Values, path string, module string) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	uploadFolder := path
	fileName := ""
	fileExtension := ""
	fileSize := 0
	var file string
	// is_file := ""

	if req["file_name"] != nil {
		fileName = req["file_name"].(string)
	}

	if req["file_extension"] != nil {
		fileExtension = req["file_extension"].(string)
	}

	if req["file_size"] != nil {
		fileSize = int(req["file_size"].(float64))
	}

	if req["file"] != nil {
		file = req["file"].(string)
	}

	filename, err := utils.UploadService(fileName, uploadFolder, file, module, int(fileSize))
	if err != nil {
		return utils.SendError(err, 500)
	}

	fmt.Println(module, fileExtension, filename, uploadFolder)

	res := map[string]interface{}{
		"file_name": filename,
		"file_url":  "/view/webroot/files/" + uploadFolder + "/" + filename,
	}

	return utils.SendData(res, "Berhasil")
}

func (service *uploadService) UploadSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	dataUri := req["datauri"].(string)
	datauriInfo, err := utils.ExtractBase64Info(dataUri)
	if err != nil {
		return utils.SendError(err, 500)
	}

	if datauriInfo.Extension == ".jfif" {
		datauriInfo.Extension = ".jpg"
	}

	filename := utils.GenerateUniqueFilename("", datauriInfo.Extension, false)
	folderPath := enums.PATH_RESPONDENT_SURVEY_IMAGE

	filename, err = utils.UploadServiceDataURI(filename, folderPath, dataUri, enums.MODULE_SURVEY, int(datauriInfo.SizeInMB))
	if err != nil {
		return utils.SendError(err, 500)
	}

	return utils.SendData(filename, "File berhasil di upload")
}

func (service *uploadService) ShowSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	filename, ok := slug["id"].(string)
	spew.Dump(filename)
	if !ok || filename == "" {
		return utils.SendError(errors.New("nama file tidak valid"), http.StatusBadRequest)
	}

	// Menggabungkan direktori root, direktori module survey, dan nama file
	newSlug := map[string]interface{}{
		"id": fmt.Sprintf("%s/%s/%s", enums.PATH_WEBROOT_FILES, enums.PATH_RESPONDENT_SURVEY_IMAGE, filename),
	}

	// Gunakan fungsi Show yang sudah ada (akan menangani MinIO dan local secara otomatis)
	return service.Show(newSlug)
}

func (service *uploadService) ShowPublicSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	spew.Dump(slug)
	filename, ok := slug["path"].(string)
	spew.Dump(filename)
	if !ok || filename == "" {
		return utils.SendError(errors.New("nama file tidak valid"), http.StatusBadRequest)
	}

	// Menggabungkan direktori root, direktori module survey, dan nama file
	newSlug := map[string]interface{}{
		"id": fmt.Sprintf("%s/%s/%s", enums.PATH_WEBROOT_FILES, enums.PATH_RESPONDENT_SURVEY_IMAGE, filename),
	}

	// Gunakan fungsi Show yang sudah ada (akan menangani MinIO dan local secara otomatis)
	return service.Show(newSlug)
}

func (service *uploadService) DeleteBulkSurveyImage(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	type Paths []string
	tmp, _ := json.Marshal(req["paths"])
	var paths Paths
	json.Unmarshal(tmp, &paths)

	for i := range paths {
		paths[i] = enums.PATH_WEBROOT_FILES + "/" + enums.PATH_RESPONDENT_SURVEY_IMAGE + "/" + paths[i]
	}

	var success bool
	err := utils.DeleteBulkServiceMinio(paths, enums.MODULE_SURVEY)
	if err != nil {
		return utils.SendError(err, 500)
	}

	success = true
	return utils.SendData(success, "File berhasil dihapus")
}

func (service *uploadService) Show(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// Jika menggunakan Minio
	if os.Getenv("WITH_MINIO") == "true" {
		return service.ShowMinio(slug)
	}

	// Validasi path
	path := slug["id"].(string)
	if path == "" {
		return utils.SendError(errors.New("path tidak valid"), http.StatusBadRequest)
	}

	// Cek keberadaan file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return utils.SendError(errors.New("NOT FOUND"), http.StatusNotFound)
	}

	// Baca file
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return utils.SendError(errors.New("NOT FOUND"), http.StatusNotFound)
	}

	// Deteksi MIME type berdasarkan ekstensi
	ext := strings.ToLower(filepath.Ext(path))
	var mimeType string

	switch ext {
	case ".xlsx":
		mimeType = "application/vnd.openxmluploadats-officedocument.spreadsheetml.sheet"
	case ".xls":
		mimeType = "application/vnd.ms-excel"
	case ".pdf":
		mimeType = "application/pdf"
	case ".doc":
		mimeType = "application/msword"
	case ".docx":
		mimeType = "application/vnd.openxmluploadats-officedocument.wordprocessingml.document"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".txt":
		mimeType = "text/plain"
	case ".csv":
		mimeType = "text/csv"
	case ".zip":
		mimeType = "application/zip"
	case ".rar":
		mimeType = "application/x-rar-compressed"
	case ".7z":
		mimeType = "application/x-7z-compressed"
	default:
		mimeType = http.DetectContentType(fileBytes)
	}

	fmt.Println("mimeType:", mimeType)

	return utils.SetResponseData(fileBytes, true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *uploadService) ShowMinio(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	path := slug["id"].(string)
	if path == "" {
		return utils.SendError(errors.New("path tidak valid"), http.StatusBadRequest)
	}

	// Initialize MinIO client
	var (
		endpoint        = os.Getenv("MINIO_ENDPOINT")
		accessKeyID     = os.Getenv("MINIO_ACCESS_KEY")
		secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
		useSSL          = os.Getenv("MINIO_USE_SSL")
		bucketName      = os.Getenv("MINIO_BUCKET")
	)

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL == "true")
	if err != nil {
		return utils.SendError(fmt.Errorf("error initializing minio client: %v", err), http.StatusInternalServerError)
	}

	// Check if object exists
	_, err = minioClient.StatObject(bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return utils.SendError(errors.New("NOT FOUND"), http.StatusNotFound)
		}
		return utils.SendError(fmt.Errorf("error checking object: %v", err), http.StatusInternalServerError)
	}

	// Get object
	object, err := minioClient.GetObject(bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return utils.SendError(fmt.Errorf("error getting object: %v", err), http.StatusInternalServerError)
	}
	defer object.Close()

	// Read all data
	fileBytes, err := io.ReadAll(object)
	if err != nil {
		return utils.SendError(fmt.Errorf("error reading object: %v", err), http.StatusInternalServerError)
	}

	// Get file extension and determine MIME type
	ext := strings.ToLower(filepath.Ext(path))
	var mimeType string

	switch ext {
	case ".xlsx":
		mimeType = "application/vnd.openxmluploadats-officedocument.spreadsheetml.sheet"
	case ".xls":
		mimeType = "application/vnd.ms-excel"
	case ".pdf":
		mimeType = "application/pdf"
	case ".doc":
		mimeType = "application/msword"
	case ".docx":
		mimeType = "application/vnd.openxmluploadats-officedocument.wordprocessingml.document"
	default:
		mimeType = http.DetectContentType(fileBytes)
	}

	fmt.Println("mimeType:", mimeType)

	return utils.SetResponseData(fileBytes, true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *uploadService) DownloadFilePath(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	if os.Getenv("WITH_MINIO") == "true" {
		return service.DownloadFilePathMinio(param)
	}

	var filename, foldername, force string

	if param.Get("filename") != "" {
		filename = param.Get("filename")
	} else {
		err := errors.New("filename tidak boleh kosong")
		utils.SendError(err, http.StatusBadRequest)
	}

	if param.Get("foldername") != "" {
		foldername = param.Get("foldername")
	} else {
		foldername = "upload"
		// err := errors.New("foldername tidak boleh kosong")
		// return utils.SendError(err, http.StatusBadRequest)
	}

	if param.Get("force") != "" {
		force = param.Get("force")
	}

	fmt.Println(filename, foldername, force)

	filePath := "webroot/files/" + foldername + "/" + filename

	// Memeriksa apakah file ada atau tidak
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		foldername = "dokumen_pemohon"
		filePath = "webroot/files/" + foldername + "/" + filename

		// return utils.SendError(err, http.StatusBadRequest)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return utils.SendError(err, http.StatusBadRequest)
	}

	rs := map[string]interface{}{
		"path": filePath,
		"url":  "http://202.10.51.67/api/show?filename=" + filename,
	}

	// Convert map to JSON byte slice
	jsonData, err := json.Marshal(rs)
	if err != nil {
		return nil, err
	}

	return utils.SetResponseData(jsonData, true, "Berhasil", 200, nil, ""), nil
}

func (service *uploadService) DownloadFilePathMinio(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// Initialize MinIO client
	var (
		endpoint        = os.Getenv("MINIO_ENDPOINT")
		accessKeyID     = os.Getenv("MINIO_ACCESS_KEY")
		secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
		useSSL          = os.Getenv("MINIO_USE_SSL")
		bucketName      = os.Getenv("MINIO_BUCKET")
	)

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL == "true")
	if err != nil {
		return utils.SendError(fmt.Errorf("error initializing minio client: %v", err), http.StatusInternalServerError)
	}

	var filename, foldername, force string

	if param.Get("filename") != "" {
		filename = param.Get("filename")
	} else {
		err := errors.New("filename tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	if param.Get("foldername") != "" {
		foldername = param.Get("foldername")
	} else {
		foldername = "upload"
	}

	if param.Get("force") != "" {
		force = param.Get("force")
	}

	fmt.Println(filename, foldername, force)

	// Construct file path dengan tetap mempertahankan struktur webroot/files
	filePath := "webroot/files/" + foldername + "/" + filename

	// Check if file exists in MinIO
	_, err = minioClient.StatObject(bucketName, filePath, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return utils.SendError(errors.New("file tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	// Optional: Generate presigned URL jika diperlukan
	presignedURL, err := minioClient.PresignedGetObject(bucketName, filePath, time.Hour*24, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	rs := map[string]interface{}{
		"path": filePath,
		"url":  "http://202.10.51.67/api/show?filename=" + filename,
		// Optional: tambahkan presigned URL jika diperlukan
		"presigned_url": presignedURL.String(),
	}

	// Convert map to JSON byte slice
	jsonData, err := json.Marshal(rs)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SetResponseData(jsonData, true, "Berhasil", 200, nil, ""), nil
}
