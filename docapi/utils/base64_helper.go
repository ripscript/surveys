package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"mime"
	"strings"
)

type Base64FileInfo struct {
	MimeType    string  // Contoh: "image/png"
	Extension   string  // Contoh: ".png"
	SizeInBytes int     // Ukuran asli file dalam Bytes
	SizeInKB    float64 // Ukuran dalam Kilobytes
	SizeInMB    float64 // Ukuran dalam Megabytes
	IsImage     bool    // True jika file terdeteksi sebagai gambar
	Width       int     // Lebar gambar dalam pixel (jika berupa gambar)
	Height      int     // Tinggi gambar dalam pixel (jika berupa gambar)
	RawData     []byte  // Data binary asli yang siap disimpan ke Storage/S3
}

func ExtractBase64Info(dataURI string) (*Base64FileInfo, error) {
	// 1. Guard Clause: Pastikan format memisahkan metadata dan data dengan koma
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("format base64 tidak valid, kehilangan pemisah koma")
	}

	header := parts[0]  // Bagian header, e.g., "data:image/png;base64"
	b64Data := parts[1] // Bagian data base64 murni

	// 2. Ekstrak Mime Type dari header
	if !strings.HasPrefix(header, "data:") {
		return nil, errors.New("format base64 tidak valid, harus diawali 'data:'")
	}

	mimeAndEncoding := strings.TrimPrefix(header, "data:")
	mimeParts := strings.Split(mimeAndEncoding, ";")
	mimeType := mimeParts[0]

	// 3. Dapatkan Ekstensi File berdasarkan Mime Type
	var extension string
	exts, err := mime.ExtensionsByType(mimeType)
	if err == nil && len(exts) > 0 {
		extension = exts[0] // Ambil ekstensi pertama yang paling relevan
	} else {
		// Fallback manual jika package mime OS tidak mengenalinya
		splitMime := strings.Split(mimeType, "/")
		if len(splitMime) == 2 {
			extension = "." + splitMime[1]
		}
	}

	// 4. Decode string base64 menjadi bentuk byte asli
	decodedBytes, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("gagal mendecode data base64: %v", err)
	}

	// 5. Kalkulasi Ukuran File dari data aslinya (bukan dari string base64)
	sizeBytes := len(decodedBytes)

	// 6. Bentuk response dasar
	fileInfo := &Base64FileInfo{
		MimeType:    mimeType,
		Extension:   extension,
		SizeInBytes: sizeBytes,
		SizeInKB:    float64(sizeBytes) / 1024.0,
		SizeInMB:    float64(sizeBytes) / (1024.0 * 1024.0),
		RawData:     decodedBytes,
	}

	// 7. Deteksi khusus jika tipe file adalah gambar untuk mendapatkan resolusi
	if strings.HasPrefix(mimeType, "image/") {
		fileInfo.IsImage = true

		// DecodeConfig sangat ringan karena hanya membaca Header gambar
		imgConfig, _, err := image.DecodeConfig(bytes.NewReader(decodedBytes))
		if err == nil {
			fileInfo.Width = imgConfig.Width
			fileInfo.Height = imgConfig.Height
		}
	}

	return fileInfo, nil
}
