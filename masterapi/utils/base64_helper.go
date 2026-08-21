package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"net/http"
	"strings"
)

type Base64FileInfo struct {
	MimeType    string
	Extension   string
	SizeInBytes int
	SizeInKB    float64
	SizeInMB    float64
	IsImage     bool
	Width       int
	Height      int
	RawData     []byte
}

func ExtractBase64InfoOld(dataURI string) (*Base64FileInfo, error) {
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("format base64 tidak valid, kehilangan pemisah koma")
	}

	b64Data := parts[1]

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("gagal mendecode data base64: %v", err)
	}

	trueMimeType := http.DetectContentType(decodedBytes)

	var extension string
	switch trueMimeType {
	case "image/jpeg", "image/jpg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	case "image/gif":
		extension = ".gif"
	default:
		extension = ""
	}

	sizeBytes := len(decodedBytes)

	fileInfo := &Base64FileInfo{
		MimeType:    trueMimeType,
		Extension:   extension,
		SizeInBytes: sizeBytes,
		SizeInKB:    float64(sizeBytes) / 1024.0,
		SizeInMB:    float64(sizeBytes) / (1024.0 * 1024.0),
		RawData:     decodedBytes,
	}

	if strings.HasPrefix(trueMimeType, "image/") {
		fileInfo.IsImage = true

		imgConfig, _, err := image.DecodeConfig(bytes.NewReader(decodedBytes))
		if err == nil {
			fileInfo.Width = imgConfig.Width
			fileInfo.Height = imgConfig.Height
		}
	}

	return fileInfo, nil
}

func ExtractBase64Info(dataURI string) (*Base64FileInfo, error) {
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("format base64 tidak valid, kehilangan pemisah koma")
	}

	headerPart := parts[0]
	b64Data := parts[1]

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("gagal mendecode data base64: %v", err)
	}

	// 1. Dapatkan MIME type asli dari bytes (bagus untuk mendeteksi gambar secara akurat)
	trueMimeType := http.DetectContentType(decodedBytes)

	// 2. Ekstrak MIME type dari header text base64
	headerMimeType := ""
	if strings.HasPrefix(headerPart, "data:") {
		headerMimeType = strings.TrimPrefix(headerPart, "data:")
		headerMimeType = strings.Split(headerMimeType, ";")[0]
	}

	// 3. Normalisasi MIME type dokumen (Menangani format aneh seperti @file/... dari frontend)
	finalMimeType := trueMimeType
	headerMimeLower := strings.ToLower(headerMimeType)

	if strings.Contains(headerMimeLower, "vnd.openxmlformats-officedocument.wordprocessingml.document") {
		finalMimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document" // DOCX
	} else if strings.Contains(headerMimeLower, "msword") {
		finalMimeType = "application/msword" // DOC
	} else if strings.Contains(headerMimeLower, "pdf") {
		finalMimeType = "application/pdf" // PDF
	}

	// 4. Set Ekstensi berdasarkan finalMimeType yang sudah dinormalisasi
	var extension string
	switch finalMimeType {
	case "image/jpeg", "image/jpg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	case "image/gif":
		extension = ".gif"
	case "application/pdf":
		extension = ".pdf"
	case "application/msword":
		extension = ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		extension = ".docx"
	default:
		extension = ""
	}

	sizeBytes := len(decodedBytes)

	fileInfo := &Base64FileInfo{
		MimeType:    finalMimeType,
		Extension:   extension,
		SizeInBytes: sizeBytes,
		SizeInKB:    float64(sizeBytes) / 1024.0,
		SizeInMB:    float64(sizeBytes) / (1024.0 * 1024.0),
		RawData:     decodedBytes,
	}

	// 5. Validasi khusus jika itu adalah gambar
	if strings.HasPrefix(finalMimeType, "image/") {
		fileInfo.IsImage = true

		imgConfig, _, err := image.DecodeConfig(bytes.NewReader(decodedBytes))
		if err == nil {
			fileInfo.Width = imgConfig.Width
			fileInfo.Height = imgConfig.Height
		}
	}

	return fileInfo, nil
}
