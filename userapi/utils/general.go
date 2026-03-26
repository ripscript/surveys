package utils

import (
	pb "backend/siccore/pb"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func IsDirectRole(role int) bool {
	return role == 1 || role == 6 || role == 7 || role == 8 || role == 9
}

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

	if param.Get("search") != "" {
		if param.Get("search") == "null" || param.Get("search") == "nill" || param.Get("search") == "undefined" {
			keyword = ""
		} else {
			keyword = strings.ToLower(param.Get("search"))
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

// ToString converts an interface{} to a string
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

func DynamicBind(payload interface{}, target interface{}) error {
	if target == nil {
		return fmt.Errorf("target tidak boleh nil")
	}

	// Convert payload ke map
	var payloadMap map[string]interface{}
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payloadData, &payloadMap); err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target harus berupa pointer")
	}

	return recursiveBind(payloadMap, targetValue.Elem())
}

// recursiveBind fungsi rekursif untuk binding
func recursiveBind(payload interface{}, target reflect.Value) error {
	if !target.CanSet() {
		return nil
	}

	switch target.Kind() {
	case reflect.Ptr:
		// Handle pointer types
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return recursiveBind(payload, target.Elem())

	case reflect.Struct:
		return bindStruct(payload, target)

	case reflect.Slice:
		return bindSlice(payload, target)

	case reflect.String:
		return bindString(payload, target)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return bindInt(payload, target)

	case reflect.Float32, reflect.Float64:
		return bindFloat(payload, target)

	case reflect.Bool:
		return bindBool(payload, target)

	default:
		return fmt.Errorf("tipe data tidak didukung: %v", target.Kind())
	}
}

// bindStruct menangani binding untuk struct
func bindStruct(payload interface{}, target reflect.Value) error {
	if target.Type() == reflect.TypeOf(time.Time{}) {
		return bindTime(payload, target)
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("payload bukan map[string]interface{}")
	}

	for i := 0; i < target.NumField(); i++ {
		field := target.Type().Field(i)
		fieldValue := target.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		payloadValue, exists := payloadMap[jsonName]
		if !exists {
			continue
		}

		if err := recursiveBind(payloadValue, fieldValue); err != nil {
			return fmt.Errorf("error binding field %s: %v", field.Name, err)
		}
	}

	return nil
}

// bindSlice menangani binding untuk slice
func bindSlice(payload interface{}, target reflect.Value) error {
	payloadSlice, ok := payload.([]interface{})
	if !ok {
		return fmt.Errorf("payload bukan slice")
	}

	sliceType := target.Type()
	newSlice := reflect.MakeSlice(sliceType, len(payloadSlice), len(payloadSlice))

	for i, item := range payloadSlice {
		if err := recursiveBind(item, newSlice.Index(i)); err != nil {
			return err
		}
	}

	target.Set(newSlice)
	return nil
}

// bindString menangani binding untuk string
func bindString(payload interface{}, target reflect.Value) error {
	var strValue string
	switch v := payload.(type) {
	case string:
		strValue = v
	case float64:
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		strValue = strconv.FormatBool(v)
	case nil:
		strValue = ""
	default:
		return fmt.Errorf("tidak dapat mengkonversi ke string: %v", payload)
	}
	target.SetString(strValue)
	return nil
}

// bindInt menangani binding untuk integer
func bindInt(payload interface{}, target reflect.Value) error {
	var intValue int64
	switch v := payload.(type) {
	case float64:
		intValue = int64(v)
	case string:
		var err error
		intValue, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	case int:
		intValue = int64(v)
	case int64:
		intValue = v
	case nil:
		intValue = 0
	default:
		return fmt.Errorf("tidak dapat mengkonversi ke int: %v", payload)
	}
	target.SetInt(intValue)
	return nil
}

// bindFloat menangani binding untuk float
func bindFloat(payload interface{}, target reflect.Value) error {
	var floatValue float64
	switch v := payload.(type) {
	case float64:
		floatValue = v
	case string:
		var err error
		floatValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
	case int:
		floatValue = float64(v)
	case nil:
		floatValue = 0
	default:
		return fmt.Errorf("tidak dapat mengkonversi ke float: %v", payload)
	}
	target.SetFloat(floatValue)
	return nil
}

// bindBool menangani binding untuk boolean
func bindBool(payload interface{}, target reflect.Value) error {
	var boolValue bool
	switch v := payload.(type) {
	case bool:
		boolValue = v
	case string:
		var err error
		boolValue, err = strconv.ParseBool(v)
		if err != nil {
			return err
		}
	case nil:
		boolValue = false
	default:
		return fmt.Errorf("tidak dapat mengkonversi ke bool: %v", payload)
	}
	target.SetBool(boolValue)
	return nil
}

// bindTime menangani binding untuk time.Time
func bindTime(payload interface{}, target reflect.Value) error {
	switch v := payload.(type) {
	case string:
		// Untuk string kosong, set nilai nil
		if v == "" {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}

		// Coba beberapa format waktu umum
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02-01-2006", // Format DD-MM-YYYY
		}

		var parsedTime time.Time
		var err error

		// Coba parse dengan format standar
		for _, format := range formats {
			if parsedTime, err = time.Parse(format, v); err == nil {
				break
			}
		}

		// Jika gagal dengan format standar, coba format custom
		if err != nil {
			if parsedTime, err = parseCustomDate(v); err != nil {
				return fmt.Errorf("format waktu tidak valid: %v", v)
			}
		}

		// Cek apakah hasil parsing adalah zero time
		if parsedTime.Equal(time.Time{}) || parsedTime.Year() == 1 {
			target.Set(reflect.Zero(target.Type()))
			return nil
		}

		target.Set(reflect.ValueOf(parsedTime))
		return nil

	case nil:
		target.Set(reflect.Zero(target.Type()))
		return nil
	default:
		return fmt.Errorf("tidak dapat mengkonversi ke time.Time: %v", payload)
	}
}

// Fungsi helper untuk parsing format DD-MM-YYYY
func parseCustomDate(dateStr string) (time.Time, error) {
	// Cek apakah string sesuai pattern DD-MM-YYYY
	if matched, _ := regexp.MatchString(`^\d{2}-\d{2}-\d{4}$`, dateStr); matched {
		// Split string berdasarkan "-"
		parts := strings.Split(dateStr, "-")
		if len(parts) != 3 {
			return time.Time{}, fmt.Errorf("format tanggal tidak valid")
		}

		// Susun ulang menjadi format YYYY-MM-DD
		rearranged := fmt.Sprintf("%s-%s-%s", parts[2], parts[1], parts[0])

		// Parse dengan format standar
		return time.Parse("2006-01-02", rearranged)
	}
	return time.Time{}, fmt.Errorf("format tanggal tidak sesuai DD-MM-YYYY")
}

func StructToMap(obj interface{}) ([]map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		// Jika unmarshal ke slice gagal, coba unmarshal ke single map
		var singleResult map[string]interface{}
		err = json.Unmarshal(data, &singleResult)
		if err != nil {
			return nil, err
		}
		// Jika berhasil, wrap single map dalam slice
		result = []map[string]interface{}{singleResult}
	}

	return result, nil
}

func Mask(val string, pattren string) string {
	return ""
}

func ConvertZeroToNil(v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Cek jika field adalah pointer
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			// Ambil nilai yang ditunjuk pointer
			elem := field.Elem()

			// Cek tipe numerik
			switch elem.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				// Jika nilai = 0, set pointer menjadi nil
				if elem.Int() == 0 {
					field.Set(reflect.Zero(field.Type()))
				}
			case reflect.Float32, reflect.Float64:
				if elem.Float() == 0 {
					field.Set(reflect.Zero(field.Type()))
				}
			}
		}
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
