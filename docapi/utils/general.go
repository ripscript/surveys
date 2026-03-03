package utils

import (
	pb "backend/siccore/pb"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"
)

func SendData(js interface{}, messages ...string) (*pb.ProxyResponse, error) {
	data, errs := json.Marshal(js)
	if errs != nil {
		message := "Terjadi Kesalahan : " + errs.Error()
		LogErrors(message)
		return SetResponseData(data, false, message, 500, errs), errs
	}
	var message string
	if len(messages) > 0 {
		message = messages[0]
	} else {
		message = "Berhasil"
	}
	return SetResponseData(data, true, message, 200, nil), nil
}

func SendError(err error, code int) (*pb.ProxyResponse, error) {
	if err.Error() != "" {
		message := err.Error()
		return SetResponseData(nil, false, message, code, nil), nil
	}
	message := "Terjadi kesalahan "
	return SetResponseData(nil, false, message, code, nil), nil
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
