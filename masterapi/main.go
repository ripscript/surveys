package main

import (
	"backend/masterapi/configs"
	"backend/masterapi/databases/migrations"
	"backend/masterapi/databases/seeders"
	"backend/masterapi/routingGrpc"
	"backend/masterapi/utils"
	pb "backend/siccore/pb"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct{}

func main() {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file :" + err.Error())
	}

	logFile, err := os.Create("logs/app.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	migrateFlag := flag.Bool("migrate", false, "Jalankan migrasi database")
	seederFlag := flag.Bool("seeder", false, "Jalankan seeder database")
	flag.Parse()
	if *migrateFlag {
		fmt.Println("Masuk Create Database")
		db := configs.SetupDatabaseMasterConnection()
		if err := migrations.Migrate(db); err != nil {
			log.Fatal("Gagal melakukan migrasi:", err)
		}

		if err := migrations.AddDashboardPerformanceIndexes(db); err != nil {
			log.Fatal("gagal menambahkan index:", err)
		}

		log.Println("Migrasi berhasil")
	} else if *seederFlag {
		db := configs.SetupDatabaseMasterConnection()
		if err := seeders.Seed(db); err != nil {
			log.Fatal("Gagal melakukan seeder:", err)
		}
		log.Println("Seeder berhasil")
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}

		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		s := grpc.NewServer()

		// LIST GRPC HANDLER
		grpc_health_v1.RegisterHealthServer(s, &HealthServer{})
		pb.RegisterProxyServer(s, &routingGrpc.GRPCServer{})

		log.Println("gRPC server is running on port " + port)
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}
}

func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()
	// Implementation for Watch method (can be left empty)
	return nil
}

func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
