package configs

import (
	"backend/reportapi/utils"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
