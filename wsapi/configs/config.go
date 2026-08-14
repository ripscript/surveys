package configs

import (
	"backend/wsapi/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Port             string
	GrpcPort         string
	JwtSecretKey     string
	WsAllowedOrigins []string
}

var App Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file tidak ditemukan, menggunakan env system")
	}

	App = Config{
		Port:         getEnv("PORT", "8090"),
		GrpcPort:     getEnv("GRPC_PORT", "8091"),
		JwtSecretKey: getEnv("JWT_SECRET_KEY", ""),
	}

	if App.JwtSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY wajib diisi")
	}

	origins := os.Getenv("WS_ALLOWED_ORIGINS")
	App.WsAllowedOrigins = strings.Split(origins, ",")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func SetupDatabaseMasterConnection() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		utils.LogErrors("Terjadi kesalahan : " + err.Error())
		return nil
	}

	var (
		dbUser     = os.Getenv("DB_USER_MASTER")
		dbPassword = os.Getenv("DB_PASSWORD_MASTER")
		dbHost     = os.Getenv("DB_HOST_MASTER")
		dbPort     = os.Getenv("DB_PORT_MASTER")
		dbSslMode  = os.Getenv("DB_SSL_MASTER")
		dbTimezone = os.Getenv("DB_TIMEZONE_MASTER")
		dbName     = os.Getenv("DB_NAME_MASTER")
	)

	psqlInfo := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v", dbHost, dbUser, dbPassword, dbName, dbPort, dbSslMode, dbTimezone)
	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatal("Error Connection Database DB_USER_MASTER "+dbHost, err.Error())
	}

	// * Setup Database pooling
	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetMaxOpenConns(5)
	sqlDb.SetConnMaxIdleTime(2 * time.Minute)
	sqlDb.SetConnMaxLifetime(60 * time.Hour)

	return db
}

func SetupDatabaseSlaveConnection() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		utils.LogErrors("Terjadi kesalahan saat memuat .env: " + err.Error())
		return nil
	}

	var (
		dbUser     = os.Getenv("DB_USER_SLAVE")
		dbPassword = os.Getenv("DB_PASSWORD_SLAVE")
		dbHost     = os.Getenv("DB_HOST_SLAVE")
		dbPort     = os.Getenv("DB_PORT_SLAVE")
		dbSslMode  = os.Getenv("DB_SSL_SLAVE")
		dbTimezone = os.Getenv("DB_TIMEZONE_SLAVE")
		dbName     = os.Getenv("DB_NAME_SLAVE")
	)

	psqlInfo := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSslMode, dbTimezone)

	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable SQL logging
	})
	if err != nil {
		log.Fatal("Error Connection Database: DB_USER_SLAVE "+dbHost, err.Error())
	}

	// Setup Database pooling
	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("Error getting database instance: "+dbHost, err.Error())
	}

	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetMaxOpenConns(5)
	sqlDb.SetConnMaxIdleTime(2 * time.Minute)
	sqlDb.SetConnMaxLifetime(60 * time.Minute) // Changed from 60 hours to 60 minutes

	// Verify connection
	err = sqlDb.Ping()
	if err != nil {
		log.Fatal("Error pinging database: "+dbHost, err.Error())
	}

	// Log connection info
	var dbNameCheck string
	err = sqlDb.QueryRow("SELECT current_database()").Scan(&dbNameCheck)
	if err != nil {
		log.Printf("Error getting database name: %v", err)
	} else {
		log.Printf("Connected to database: %s", dbNameCheck)
	}

	return db
}

var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, allowed := range App.WsAllowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},

	HandshakeTimeout: 10 * time.Second,
}
