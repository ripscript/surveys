package handlers

import (
	"backend/reportapi/models"
	"backend/reportapi/utils"
	pb "backend/siccore/pb"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func Healthy(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	var data []byte
	var success bool = true
	var message string = "Service ReportApi Is Healthy"
	var code int = int(http.StatusOK)
	return utils.SetResponseData(data, success, message, code, nil, ""), nil
}
