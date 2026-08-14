package main

import (
	pb "backend/siccore/pb"
	"backend/wsapi/configs"
	"backend/wsapi/databases/migrations"
	"backend/wsapi/handlers"
	"backend/wsapi/hub"
	"backend/wsapi/repository"
	"backend/wsapi/routingGrpc"
	"backend/wsapi/utils"
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct{}

func main() {
	defer func() {
		if r := recover(); r != nil {
			utils.LogErrors(fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r))
		}
	}()

	configs.LoadConfig()

	migrateFlag := flag.Bool("migrate", false, "Jalankan migrasi database")
	flag.Parse()
	if *migrateFlag {
		db := configs.SetupDatabaseMasterConnection()
		if err := migrations.Migrate(db); err != nil {
			utils.LogErrors("Gagal melakukan migrasi: " + err.Error())
		}
		utils.Logger.Info("Migrasi berhasil")
	} else {
		h := hub.NewHub()
		go startGrpcServer(h)
		startHttpServer(h)
	}
}

func startHttpServer(h *hub.Hub) {
	e := echo.New()
	e.Use(utils.ErrorLogger())
	e.Use(middleware.CORS())

	dbMaster := configs.SetupDatabaseMasterConnection()
	dbSlave := configs.SetupDatabaseSlaveConnection()

	channelRepo := repository.NewChannelRepo(dbSlave, dbMaster)
	wsHandler := handlers.NewWsHandler(h, channelRepo)

	e.GET("/ws", wsHandler.HandleWs)

	e.Logger.Fatal(e.Start(":" + configs.App.Port))
}

func startGrpcServer(h *hub.Hub) {
	listener, err := net.Listen("tcp", ":"+configs.App.GrpcPort)
	if err != nil {
		utils.Logger.WithField("error", err).Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(s, &HealthServer{})
	pb.RegisterProxyServer(s, routingGrpc.NewGRPCServer(h))

	utils.Logger.Info("gRPC server is running on port " + configs.App.GrpcPort)
	utils.Logger.Info("HTTP/WS server berjalan di port " + configs.App.Port)
	if err := s.Serve(listener); err != nil {
		utils.Logger.WithField("error", err).Fatal("Failed to serve")
	}
}

func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	return nil
}

func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
