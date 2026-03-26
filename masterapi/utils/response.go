package utils

import (
	pb "backend/siccore/pb"
	"fmt"
	"log"
	"os"
	"strings"
)

func SetResponseData(data []byte, success bool, message string, code int, err error, token string) *pb.ProxyResponse {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	// Jika ada kesalahan, simpan pesan kesalahan ke dalam file error.log
	if err != nil {
		if strings.Contains(err.Error(), "rpc error") {
			message = "Terjadi masalah pada service yang sedang anda akses"
		}
	}

	response := &pb.ProxyResponse{
		Success: success,
		Message: message,
		Code:    int32(code),
		Data:    data,
		Token:   token,
	}

	return response
}

func logError(err error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	// Buka file error.log dengan mode menambahkan (append)
	file, err := os.OpenFile("errors/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Buat logger baru untuk menulis ke file
	logger := log.New(file, "", log.LstdFlags)

	// Tulis pesan kesalahan ke file error.log
	logger.Printf("Error: %v\n", err)
}
