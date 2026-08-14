package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	pb "backend/siccore/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var DOCAPI string = "DOCAPI"
var REPORTAPI string = "REPORTAPI"

func TrxData(service string, path string, method string, slugs map[string]string, param string, payload interface{}) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	// Mengkonversi payload JSON ke byte data
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Marshaling map ke JSON
	slugBytes, err := json.Marshal(slugs)
	if err != nil {
		return nil, err
	}

	// Membuat ProxyRequest dengan payload byte data dan handlerName
	req := &pb.ProxyRequest{
		Data:     payloadBytes,
		Path:     path,
		Method:   method,
		IsSecure: false,
		Param:    []byte(param),
		Slug:     slugBytes,
	}

	// Mendapatkan nilai header "Authorization" dari environment atau diatur secara manual
	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")

	// Menambahkan header "Authorization" ke dalam metadata gRPC
	md := metadata.New(map[string]string{
		"Authorization": authorizationHeader,
	})

	// Mengirim permintaan ke server gRPC dengan metadata yang telah ditambahkan
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	host := os.Getenv(service + "_HOST")
	port := os.Getenv(service + "_PORT")

	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewProxyClient(conn)
	res, err := client.SendData(ctx, req)
	if err != nil {
		return nil, err
	}

	// Mengambil data dari res.GetData()
	dataBytes := res.GetData()

	return dataBytes, nil
}

func GrpcRelasi(service string, path string, method string, payload map[string]interface{}, slugs map[string]string, params string) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	peg, err := TrxData(service, path, method, slugs, params, payload)
	if err != nil {
		return nil, err
	}
	// pemohon dari grpc
	if peg != nil {
		var data interface{}
		err := json.Unmarshal(peg, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, nil
}

func SaveLogActivities(serviceName string, modul string, action string, userId int, userName string, entityID string, description string) error {
	GeneralRecover()

	service := REPORTAPI
	path := "/save/log/activities"
	method := "POST"
	slugs := map[string]string{}
	params := ""
	payload := map[string]interface{}{
		"serviceName": serviceName,
		"module":      modul,
		"action":      action,
		"userId":      userId,
		"userName":    userName,
		"entityID":    entityID,
		"description": description,
	}

	res, err := TrxData(service, path, method, slugs, params, payload)
	if err != nil {
		return err
	}

	if res != nil {
		var dt interface{}
		err := json.Unmarshal(res, &dt)
		if err != nil {
			return err
		}
	}
	return nil
}
