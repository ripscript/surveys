package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	pb "backend/siccore/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var DOCAPI string = "DOCAPI"

var (
	// Cache koneksi agar tidak perlu Dial berulang kali (Connection Pooling)
	grpcConnections = make(map[string]*grpc.ClientConn)
	connMutex       sync.RWMutex
)

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

func getClient(host string) (pb.ProxyClient, error) {
	connMutex.RLock()
	conn, exists := grpcConnections[host]
	connMutex.RUnlock()

	if !exists {
		connMutex.Lock()
		defer connMutex.Unlock()

		// Double check setelah lock
		if conn, exists = grpcConnections[host]; !exists {
			var err error
			conn, err = grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			grpcConnections[host] = conn
		}
	}
	return pb.NewProxyClient(conn), nil
}

func HitBackend(ctx context.Context, targetHost, method, path string, slugData map[string]interface{}, reqBody map[string]interface{}) ([]byte, error) {
	client, err := getClient(targetHost)
	if err != nil {
		return nil, errors.New("gagal terhubung ke service target: " + err.Error())
	}

	var slugBytes, bodyBytes []byte
	if slugData != nil {
		slugBytes, _ = json.Marshal(slugData)
	}
	if reqBody != nil {
		bodyBytes, _ = json.Marshal(reqBody)
	}

	request := &pb.ProxyRequest{
		Method:   method,
		Path:     path,
		Slug:     slugBytes,
		Data:     bodyBytes,
		IsSecure: true,
	}

	// Forward Token
	outCtx := context.Background()
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authTokens, exists := md["authorization"]; exists && len(authTokens) > 0 {
			outMD := metadata.Pairs("authorization", authTokens[0])
			outCtx = metadata.NewOutgoingContext(context.Background(), outMD)
		}
	}

	res, err := client.SendData(outCtx, request)
	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, errors.New(res.Message)
	}

	return res.Data, nil
}
