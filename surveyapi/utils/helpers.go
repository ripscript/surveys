package utils

import (
	"backend/surveyapi/models"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
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

func FormatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			snakeCaseField := toSnakeCase(e.Field())
			field := formatToTitleCase(snakeCaseField)

			switch e.Tag() {
			case "required":
				return field + " wajib diisi"
			case "len":
				return field + " harus tepat " + e.Param() + " karakter"
			case "numeric":
				return field + " hanya boleh berisi angka"
			case "email":
				return "Format " + field + " tidak valid"
			default:
				return field + " tidak valid"
			}
		}
	}

	// 2. Penanganan error dari proses binding (menggunakan Regex agar lebih kebal)
	errMsg := err.Error()
	if strings.Contains(errMsg, "error binding field") || strings.Contains(errMsg, "strconv") {

		// Mengambil nama field dari pesan error menggunakan Regex
		// Pola ini akan mencari teks "error binding field " diikuti oleh nama field
		re := regexp.MustCompile(`error binding field ([a-zA-Z0-9_]+)`)
		matches := re.FindStringSubmatch(errMsg)

		// Set default nama field jika gagal diekstrak
		field := "Input"
		if len(matches) > 1 {
			rawField := matches[1]
			snakeCaseField := toSnakeCase(rawField)
			field = formatToTitleCase(snakeCaseField)
		}

		// Menentukan pesan error berdasarkan tipe kegagalan strconv
		if strings.Contains(errMsg, "strconv.ParseBool") {
			return field + " harus berupa nilai true atau false"
		} else if strings.Contains(errMsg, "strconv.ParseInt") || strings.Contains(errMsg, "strconv.ParseUint") {
			return field + " harus berupa angka bulat"
		} else if strings.Contains(errMsg, "strconv.ParseFloat") {
			return field + " harus berupa angka desimal"
		}

		return "Format " + field + " tidak sesuai tipe data yang diharapkan"
	}

	return err.Error()
}

func GenerateUniqueString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}

		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

func GenerateShortHash() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func StringToBool(data string) bool {
	lowerData := strings.ToLower(data)
	return lowerData == "true" || lowerData == "1" || lowerData == "yes"
}

func IntToBool(data int) bool {
	return data != 0
}

func BoolToInt(data bool) int {
	if data {
		return 1
	}
	return 0
}

func GetTimeGreeting() string {
	var currentTime time.Time

	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		currentTime = time.Now()
	} else {
		currentTime = time.Now().In(location)
	}

	hour := currentTime.Hour()

	switch {
	case hour >= 5 && hour < 11:
		return "pagi"
	case hour >= 11 && hour < 15:
		return "siang"
	case hour >= 15 && hour < 18:
		return "sore"
	default:
		return "malam"
	}
}

func ReplaceStringRespondentVariable(input string, respondent *models.DetailRespondent) string {

	if respondent == nil || respondent.ID == 0 {
		return input
	}

	makeStr := input
	if strings.Contains(makeStr, "(Nama Responden)") {
		makeStr = strings.ReplaceAll(makeStr, "(Nama Responden)", respondent.Name)
	}

	if strings.Contains(makeStr, "(Kontak Responden)") {
		makeStr = strings.ReplaceAll(makeStr, "(Kontak Responden)", respondent.PhoneNumber)
	}

	if strings.Contains(makeStr, "(Email Responden)") {
		makeStr = strings.ReplaceAll(makeStr, "(Email Responden)", respondent.Email)
	}

	if strings.Contains(makeStr, "(Waktu Pembuka)") {
		makeStr = strings.ReplaceAll(makeStr, "(Waktu Pembuka)", GetTimeGreeting())
	}
	if strings.Contains(makeStr, "(Wilayah Responden)") {
		makeStr = strings.ReplaceAll(makeStr, "(Wilayah Responden)", "Kecamatan "+respondent.Kecamatan+" - Kelurahan "+respondent.Kelurahan+" - RW "+respondent.RW+" - RT "+respondent.RT)
	}

	return makeStr
}
