package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func formatToTitleCase(snakeStr string) string {
	spacedStr := strings.ReplaceAll(snakeStr, "_", " ")

	words := strings.Fields(spacedStr)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func ToSnakeCase(s string) string {
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func SplitFullName(fullName string) (firstName string, lastName string) {
	trimmed := strings.TrimSpace(fullName)
	if trimmed == "" {
		return "", ""
	}

	// Split berdasarkan whitespace, otomatis handle multiple spaces
	parts := strings.Fields(trimmed)

	if len(parts) == 1 {
		return parts[0], ""
	}

	firstName = parts[0]
	lastName = strings.Join(parts[1:], " ")

	return firstName, lastName
}

func BoolToPointer(data bool) *bool {
	return &data
}
