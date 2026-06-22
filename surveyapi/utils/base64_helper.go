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

func ExtractBase64Info(dataURI string) (*Base64FileInfo, error) {
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
