package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func EncryptInt(data int) (string, error) {
	key := os.Getenv("ENCRYPT_KEY")
	if len(key) != 32 {
		return "", errors.New("ENCRYPT_KEY harus 32 karakter")
	}

	plaintext := []byte(strconv.Itoa(data))

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	result := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(result), nil
}
func DecryptInt(encoded string) (int, error) {
	key := os.Getenv("ENCRYPT_KEY")
	if len(key) != 32 {
		return 0, errors.New("ENCRYPT_KEY harus 32 karakter")
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return 0, err
	}

	nonceSize := 12
	if len(data) < nonceSize {
		return 0, errors.New("data tidak valid")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(plaintext))
}

func ContainsInt64(slice []int64, value int64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func TryParseIndonesianDate(input string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))

	// Map bulan bahasa Indonesia ke angka
	bulanMap := map[string]string{
		"januari": "01", "februari": "02", "maret": "03", "april": "04",
		"mei": "05", "juni": "06", "juli": "07", "agustus": "08",
		"september": "09", "oktober": "10", "november": "11", "desember": "12",
	}

	// Ganti kata bulan dengan angka
	for id, num := range bulanMap {
		if strings.Contains(input, id) {
			input = strings.Replace(input, id, num, 1)
			break
		}
	}

	// Hapus spasi ekstra jika ada
	input = strings.Join(strings.Fields(input), " ")

	// Coba parsing ke format time.Time (format referensi Go: "02 01 2006")
	t, err := time.Parse("02 01 2006", input)
	if err == nil {
		// Jika berhasil, kembalikan dalam format standar SQL (YYYY-MM-DD)
		return t.Format("2006-01-02"), true
	}

	return "", false
}
