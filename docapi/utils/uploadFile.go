package utils

import (
	"backend/docapi/enums"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/minio/minio-go"
)

func UploadService(fileName string, folderPath string, file string, module string, fileSize int) (string, error) {
	//gunakan minio
	if os.Getenv("WITH_MINIO") == "true" {
		return UploadServiceMinio(fileName, folderPath, file, module, fileSize)
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Warning: failed to load Asia/Jakarta timezone, using fixed WIB offset:", err)
		loc = time.FixedZone("WIB", 7*60*60)
	}

	// Get current date in Asia/Jakarta timezone
	currentDate := time.Now().In(loc).Format("2006-01-02")
	uploadDir := "webroot/files/" + folderPath + "/" + currentDate
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.MkdirAll(uploadDir, 0755)
		if err != nil {
			return "", err
		}
	}

	byteData, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return "", err
	}

	if !isValidSize(fileSize, module) {
		return "", errors.New("file size exceeds limit")
	}

	currentDateTime := time.Now().In(loc).Format("20060102150405000")
	fileName = currentDateTime + "-" + generateFileName(fileName)

	// Write file to the directory
	uploadPath := enums.PATH_WEBROOT_FILES + "/" + folderPath + "/" + currentDate + "/" + fileName

	dst, err := os.Create(uploadPath)
	if err != nil {
		return "", errors.New("AAAAAA")
	}
	defer dst.Close()

	if _, err := dst.Write(byteData); err != nil {
		return "", err
	}

	return currentDate + "/" + fileName, nil
}

func UploadServiceMinio(fileName string, folderPath string, file string, module string, fileSize int) (string, error) {
	// Initialize MinIO client
	var (
		endpoint        = os.Getenv("MINIO_ENDPOINT")
		accessKeyID     = os.Getenv("MINIO_ACCESS_KEY")
		secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
		useSSL          = os.Getenv("MINIO_USE_SSL")
		bucketName      = os.Getenv("MINIO_BUCKET")
	)

	// fmt.Println("endpoint", endpoint)
	// fmt.Println("accessKeyID", accessKeyID)
	// fmt.Println("secretAccessKey", secretAccessKey)
	// fmt.Println("useSSL", useSSL)

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL == "true")
	if err != nil {
		return "", fmt.Errorf("error initializing minio client: %v", err)
	}

	// Check if bucket exists and create if it doesn't
	exists, err := minioClient.BucketExists(bucketName)
	if err != nil {
		return "", fmt.Errorf("error checking bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(bucketName, "")
		if err != nil {
			return "", fmt.Errorf("error creating bucket: %v", err)
		}
		fmt.Printf("Successfully created bucket %q\n", bucketName)
	}

	// Load Jakarta timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Warning: failed to load Asia/Jakarta timezone, using fixed WIB offset:", err)
		loc = time.FixedZone("WIB", 7*60*60)
	}

	// Get current date in Asia/Jakarta timezone
	currentDate := time.Now().In(loc).Format("2006-01-02")

	// Validate file size
	if !isValidSize(fileSize, module) {
		return "", errors.New("file size exceeds limit")
	}

	// Decode base64 file
	byteData, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %v", err)
	}

	// Generate filename with instansi ID and timestamp
	currentDateTime := time.Now().In(loc).Format("20060102150405000")
	fileName = currentDateTime + "-" + fileName

	folderPath = enums.PATH_WEBROOT_FILES + "/" + folderPath

	// Construct object path (similar to directory structure)
	objectPath := fmt.Sprintf("%s/%s/%s", folderPath, currentDate, fileName)

	// Create reader from byte data
	fileReader := bytes.NewReader(byteData)

	// Get content type
	contentType := http.DetectContentType(byteData)

	// Upload to MinIO
	_, err = minioClient.PutObject(
		bucketName,
		objectPath,
		fileReader,
		int64(len(byteData)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("error uploading to minio: %v", err)
	}

	// Return the relative path (currentDate/fileName) as before
	return currentDate + "/" + fileName, nil
}

func DeleteBulkServiceMinio(paths []string, module string) error {
	if len(paths) == 0 {
		return nil
	}

	var (
		endpoint        = os.Getenv("MINIO_ENDPOINT")
		accessKeyID     = os.Getenv("MINIO_ACCESS_KEY")
		secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
		useSSL          = os.Getenv("MINIO_USE_SSL")
		bucketName      = os.Getenv("MINIO_BUCKET")
	)

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL == "true")
	if err != nil {
		return fmt.Errorf("error initializing minio client: %v", err)
	}

	objectsCh := make(chan string)

	go func() {
		defer close(objectsCh)
		for _, path := range paths {
			objectsCh <- path
		}
	}()

	errorCh := minioClient.RemoveObjects(bucketName, objectsCh)

	for rErr := range errorCh {
		if rErr.Err != nil {
			return fmt.Errorf("error deleting object %s: %v", rErr.ObjectName, rErr.Err)
		}
	}

	return nil
}

func isValidSize(size int, module string) bool {
	maxSize := 50
	switch module {
	case "video-tutorial":
		maxSize = 100
	case "respondent-survey-image":
		maxSize = 5
	}

	return size <= maxSize*1024*1024
}

func generateFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	newName := fmt.Sprintf("%s_%d%s", name, time.Now().UnixNano(), ext)
	return newName
}

func UploadServiceDataURI(fileName string, folderPath string, dataURI string, module string, fileSize int) (string, error) {
	parts := strings.SplitN(dataURI, ",", 2)
	base64String := dataURI
	if len(parts) == 2 {
		base64String = parts[1]
	}

	return UploadService(fileName, folderPath, base64String, module, fileSize)
}
