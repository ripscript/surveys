package utils

import (
	pb "backend/siccore/pb"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	// Cache koneksi agar tidak perlu Dial berulang kali (Connection Pooling)
	grpcConnections = make(map[string]*grpc.ClientConn)
	connMutex       sync.RWMutex
)

// getClient mengambil koneksi yang sudah ada, atau membuat baru jika belum ada
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

func HitBackendNotSecure(ctx context.Context, targetHost, method, path string, slugData map[string]interface{}, reqBody map[string]interface{}) ([]byte, error) {
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
		IsSecure: false,
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
