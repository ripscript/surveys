package utils

import (
	pb "backend/siccore/pb"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func SendData(js interface{}, messages ...string) (*pb.ProxyResponse, error) {
	data, errs := json.Marshal(js)
	if errs != nil {
		message := "Terjadi Kesalahan : " + errs.Error()
		LogErrors(message)
		return SetResponseData(data, false, message, 500, errs, ""), errs
	}
	var message string
	if len(messages) > 0 {
		message = messages[0]
	} else {
		message = "Berhasil"
	}
	return SetResponseData(data, true, message, 200, nil, ""), nil
}

func SendError(err error, code int) (*pb.ProxyResponse, error) {
	if err.Error() != "" {
		message := err.Error()
		return SetResponseData(nil, false, message, code, nil, ""), nil
	}
	message := "Terjadi kesalahan "
	return SetResponseData(nil, false, message, code, nil, ""), nil
}

func GeneralRecover() {
	if r := recover(); r != nil {
		// Buffer to store stack trace information
		buf := make([]byte, 1<<16) // 64KB
		runtime.Stack(buf, false)

		// Get details of where the panic occurred
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "unknown"
			line = 0
		}

		// Log the error with detailed information
		message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v\nFile: %s\nLineS: %d\nStack Trace: %s", r, file, line, buf)
		LogErrors(message)
	}
}

func SetPagination(param url.Values) (int, int, int, string, error) {
	pageStr := "1"
	limitStr := "5"
	offsetStr := "0"
	keyword := ""

	if param.Get("page") != "" {
		pageStr = param.Get("page")
	}

	if param.Get("limit") != "" {
		limitStr = param.Get("limit")
	}

	if param.Get("offset") != "" {
		offsetStr = param.Get("offset")
	}

	if param.Get("q") != "" {
		if param.Get("q") == "null" || param.Get("q") == "nill" || param.Get("q") == "undefined" {
			keyword = ""
		} else {
			keyword = strings.ToLower(param.Get("q"))
		}
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		message := "Terjadi kesalahan" + err.Error()
		LogErrors(message)
		return 0, 0, 0, "", err
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		fmt.Println("Error:", err)
		message := "Terjadi kesalahan" + err.Error()
		LogErrors(message)
		return 0, 0, 0, "", err

	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		fmt.Println("Error:", err)
		message := "Terjadi kesalahan" + err.Error()
		LogErrors(message)
		return 0, 0, 0, "", err
	}

	if offset == 0 && offsetStr == "0" {
		offset = (page - 1) * limit
	}

	return page, limit, offset, keyword, nil

}

func ToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return int64(i), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

func ToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int8, int16, int32, int64:
		return strconv.FormatInt(v.(int64), 10), nil
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(v.(uint64), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

func GenerateUniqueFilename(prefix string, extension string, useTime bool) string {
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	var parts []string

	if prefix != "" {
		cleanPrefix := strings.ReplaceAll(strings.ToLower(prefix), " ", "_")
		parts = append(parts, cleanPrefix)
	}

	if useTime {
		parts = append(parts, time.Now().Format("20060102_150405"))
	}

	parts = append(parts, randomHex)

	return strings.Join(parts, "_") + extension
}

func TranslateError(err validator.FieldError) string {
	snakeCaseField := toSnakeCase(err.Field())
	field := formatToTitleCase(snakeCaseField)
	switch err.Tag() {
	// Error dari Standard Tags
	case "required":
		return fmt.Sprintf("%s wajib diisi.", field)
	case "min":
		return fmt.Sprintf("%s tidak memenuhi batas minimal atau kosong.", field)
	case "gt":
		return fmt.Sprintf("%s harus lebih besar dari 0.", field)
	case "oneof":
		return fmt.Sprintf("%s nilainya tidak valid. Harus salah satu dari: %s.", field, strings.ReplaceAll(err.Param(), " ", ", "))

	// Error dari Custom Business Rules (Section & Group Structure)
	case "required_with_has_section":
		return fmt.Sprintf("%s wajib diisi angka karena form ini diatur menggunakan sistem Section (has_section: true).", field)
	case "must_be_null_if_no_section":
		return fmt.Sprintf("%s harus dikirim sebagai null karena form ini tidak menggunakan sistem Section (has_section: false).", field)
	case "group_cannot_breakdown":
		return fmt.Sprintf("Terjadi kesalahan pada %s. Sebuah Grup Pertanyaan tidak boleh memiliki status breakdown.", field)
	case "group_must_have_multiple_questions":
		return fmt.Sprintf("%s tidak valid. Sebuah Grup minimal harus berisi 2 buah ID pertanyaan.", field)
	case "group_cannot_use_logic":
		return fmt.Sprintf("Terjadi kesalahan pada %s. Sebuah Grup Pertanyaan hanya dapat menggunakan rule 'jump-to', tidak boleh menggunakan rule 'logic'.", field)

	// Error Penamaan (Section & Group Name)
	case "group_must_have_name":
		return fmt.Sprintf("Grup pada %s wajib diberikan nama (group_name).", field)
	case "must_be_null_if_not_group":
		return fmt.Sprintf("Atribut %s tidak diperlukan karena item ini bukan merupakan grup pertanyaan.", field)
	case "duplicate_section_name":
		return fmt.Sprintf("Nama section tidak valid pada %s. Nama ini sudah digunakan oleh Section Index ke-%s. Nama section harus unik.", field, err.Param())
	case "section_must_have_name":
		return fmt.Sprintf("Data tidak lengkap pada %s. Section Index ke-%s belum memiliki nama. Harap isi atribut section_name minimal satu kali pada anggota section ini.", field, err.Param())
	case "must_be_null_if_not_using_section":
		return fmt.Sprintf("Atribut %s harus dikosongkan (null) karena form ini tidak menggunakan sistem Section (has_section: false).", field)

	// Error Routing / Logic
	case "required_if_breakdown_true":
		return fmt.Sprintf("Array %s wajib diisi detail opsinya karena is_breakdown bernilai true.", field)
	case "must_be_null_if_breakdown":
		return fmt.Sprintf("%s harus dikirim sebagai null karena pergerakan alur diatur oleh opsi jawaban (breakdown).", field)
	case "required_if_rule_logic":
		return fmt.Sprintf("Array %s wajib diisi karena rule diset sebagai 'logic'.", field)
	case "must_be_null_if_end":
		return fmt.Sprintf("Field %s harus dikirim sebagai null karena alur diset berakhir (is_end: true).", field)
	case "must_have_target_or_end":
		return fmt.Sprintf("Field %s tidak valid. Anda harus menyertakan target_question_id atau mengubah is_end menjadi true.", field)

	// Error Duplikasi Data
	case "duplicate_question_id":
		return fmt.Sprintf("Ditemukan ID pertanyaan duplikat pada %s. ID ini sudah digunakan sebelumnya pada urutan (sequence) ke-%s.", field, err.Param())

	// Default fallback
	default:
		return fmt.Sprintf("%s tidak valid pada validasi '%s'.", field, err.Tag())
	}
}
