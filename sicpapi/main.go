package main

import (
	"backend/sicpapi/routes"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan")
		}
	}()
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// mengatur cors
	e.Use(middleware.CORS())

	routes.SetupRoutes(e)

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		e.Logger.Fatal("Error loading .env file : ", err.Error())
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
